package auth

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	chatfeature "example.com/api/internal/features/chat"
	usersfeature "example.com/api/internal/features/users"
	"example.com/api/internal/platform/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAuthServiceRegisterLoginMe(t *testing.T) {
	pool := setupFeaturesTestDB(t)
	defer pool.Close()

	authRepo := NewPGRepository(pool)
	svc := NewService(authRepo, []byte("services-test-secret"))
	email := uniqueEmail("authsvc")
	password := "Password123"
	defer cleanupUsers(t, pool, []string{email})

	token, user, err := svc.Register(context.Background(), "  "+strings.ToUpper(email)+"  ", password)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
	if user.Email != email {
		t.Fatalf("expected normalized email %q, got %q", email, user.Email)
	}

	_, _, err = svc.Register(context.Background(), email, password)
	if !errors.Is(err, ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}

	_, _, err = svc.Login(context.Background(), email, "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}

	loginToken, loggedIn, err := svc.Login(context.Background(), email, password)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if loginToken == "" {
		t.Fatalf("expected non-empty login token")
	}
	if loggedIn.ID != user.ID {
		t.Fatalf("expected login user id %s, got %s", user.ID, loggedIn.ID)
	}

	meUser, err := svc.Me(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("me existing user: %v", err)
	}
	if meUser.Email != email {
		t.Fatalf("expected me email %q, got %q", email, meUser.Email)
	}

	_, err = svc.Me(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUsersServiceListAndGetByID(t *testing.T) {
	pool := setupFeaturesTestDB(t)
	defer pool.Close()

	authRepo := NewPGRepository(pool)
	authSvc := NewService(authRepo, []byte("services-test-secret"))
	usersRepo := usersfeature.NewPGRepository(pool)
	usersSvc := usersfeature.NewService(usersRepo)

	email1 := uniqueEmail("users1")
	email2 := uniqueEmail("users2")
	email3 := uniqueEmail("users3")
	password := "Password123"
	defer cleanupUsers(t, pool, []string{email1, email2, email3})

	_, user1, err := authSvc.Register(context.Background(), email1, password)
	if err != nil {
		t.Fatalf("register user1: %v", err)
	}
	_, user2, err := authSvc.Register(context.Background(), email2, password)
	if err != nil {
		t.Fatalf("register user2: %v", err)
	}
	_, user3, err := authSvc.Register(context.Background(), email3, password)
	if err != nil {
		t.Fatalf("register user3: %v", err)
	}

	listed, err := usersSvc.ListWithPagination(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if !containsUser(listed, user1.ID) || !containsUser(listed, user2.ID) || !containsUser(listed, user3.ID) {
		t.Fatalf("expected user list to contain all registered users")
	}

	gotUser, err := usersSvc.GetByID(context.Background(), user2.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}
	if gotUser.Email != email2 {
		t.Fatalf("expected user email %q, got %q", email2, gotUser.Email)
	}

	_, err = usersSvc.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, usersfeature.ErrUserNotFound) {
		t.Fatalf("expected users ErrUserNotFound, got %v", err)
	}
}

func TestChatServiceSendAndListFlow(t *testing.T) {
	pool := setupFeaturesTestDB(t)
	defer pool.Close()

	authRepo := NewPGRepository(pool)
	authSvc := NewService(authRepo, []byte("services-test-secret"))
	chatRepo := chatfeature.NewPGRepository(pool)
	chatSvc := chatfeature.NewService(chatRepo)

	email1 := uniqueEmail("chat1")
	email2 := uniqueEmail("chat2")
	email3 := uniqueEmail("chat3")
	password := "Password123"
	defer cleanupUsers(t, pool, []string{email1, email2, email3})

	_, user1, err := authSvc.Register(context.Background(), email1, password)
	if err != nil {
		t.Fatalf("register user1: %v", err)
	}
	_, user2, err := authSvc.Register(context.Background(), email2, password)
	if err != nil {
		t.Fatalf("register user2: %v", err)
	}
	_, user3, err := authSvc.Register(context.Background(), email3, password)
	if err != nil {
		t.Fatalf("register user3: %v", err)
	}

	_, err = chatSvc.SendMessage(context.Background(), user1.ID, "00000000-0000-0000-0000-000000000000", "should fail")
	if !errors.Is(err, chatfeature.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound for missing recipient, got %v", err)
	}

	_, err = chatSvc.SendMessage(context.Background(), user1.ID, user2.ID, "   ")
	if !errors.Is(err, chatfeature.ErrMessageContentNeeded) {
		t.Fatalf("expected ErrMessageContentNeeded, got %v", err)
	}

	sentToUser2, err := chatSvc.SendMessage(context.Background(), user1.ID, user2.ID, "hello service chat")
	if err != nil {
		t.Fatalf("send message to user2: %v", err)
	}
	if sentToUser2.Content != "hello service chat" {
		t.Fatalf("unexpected sent message content: %q", sentToUser2.Content)
	}

	if _, err := chatSvc.SendMessage(context.Background(), user1.ID, user3.ID, "hello third user"); err != nil {
		t.Fatalf("send message to user3: %v", err)
	}

	messages, err := chatSvc.ListMessages(context.Background(), user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("list messages: %v", err)
	}
	if len(messages) == 0 {
		t.Fatalf("expected at least one message")
	}

	chats, err := chatSvc.ListChats(context.Background(), user1.ID)
	if err != nil {
		t.Fatalf("list chats: %v", err)
	}
	if !containsChatSummary(chats, user2.ID) || !containsChatSummary(chats, user3.ID) {
		t.Fatalf("expected chats list to include both conversation users")
	}
}

func setupFeaturesTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL_TEST")
	if databaseURL == "" {
		databaseURL = "postgresql://app:app@localhost:5432/app?sslmode=disable"
	}

	pool, err := db.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("skip features integration tests: postgres unavailable (%v)", err)
	}

	if err := db.RunMigrations(context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("run migrations: %v", err)
	}

	return pool
}

func uniqueEmail(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000") + "@boilerplate.test"
}

func cleanupUsers(t *testing.T, pool *pgxpool.Pool, emails []string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM users WHERE email = ANY($1)`, emails)
	if err != nil {
		t.Fatalf("cleanup users: %v", err)
	}
}

func containsUser(users []usersfeature.User, userID string) bool {
	for _, user := range users {
		if user.ID == userID {
			return true
		}
	}
	return false
}

func containsChatSummary(chats []chatfeature.ChatSummary, userID string) bool {
	for _, chat := range chats {
		if chat.UserID == userID {
			return true
		}
	}
	return false
}
