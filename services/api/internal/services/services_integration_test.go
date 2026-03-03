package services

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"example.com/api/internal/db"
	"example.com/api/internal/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAuthServiceRegisterLoginMe(t *testing.T) {
	pool := setupServicesTestDB(t)
	defer pool.Close()

	authRepo := repositories.NewPGAuthRepository(pool)
	svc := NewAuthService(authRepo, []byte("services-test-secret"))
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

func TestSocialServiceLikeMatchBlockFlow(t *testing.T) {
	pool := setupServicesTestDB(t)
	defer pool.Close()

	authRepo := repositories.NewPGAuthRepository(pool)
	authSvc := NewAuthService(authRepo, []byte("services-test-secret"))
	socialRepo := repositories.NewPGSocialRepository(pool)
	socialSvc := NewSocialService(socialRepo)

	email1 := uniqueEmail("social1")
	email2 := uniqueEmail("social2")
	email3 := uniqueEmail("social3")
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
	_, _, err = authSvc.Register(context.Background(), email3, password)
	if err != nil {
		t.Fatalf("register user3: %v", err)
	}

	discoverBefore, err := socialSvc.Discover(context.Background(), user1.ID)
	if err != nil {
		t.Fatalf("discover before: %v", err)
	}
	if !containsDiscoverUser(discoverBefore, user2.ID) {
		t.Fatalf("expected discover to include user2 before interactions")
	}

	matched, err := socialSvc.Like(context.Background(), user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("like user1->user2: %v", err)
	}
	if matched {
		t.Fatalf("expected first like to be unmatched")
	}

	matched, err = socialSvc.Like(context.Background(), user2.ID, user1.ID)
	if err != nil {
		t.Fatalf("like user2->user1: %v", err)
	}
	if !matched {
		t.Fatalf("expected reciprocal like to match")
	}

	matches, err := socialSvc.Matches(context.Background(), user1.ID)
	if err != nil {
		t.Fatalf("matches before block: %v", err)
	}
	if !containsDiscoverUser(matches, user2.ID) {
		t.Fatalf("expected matches to include user2 before block")
	}

	if err := socialSvc.Block(context.Background(), user1.ID, user2.ID); err != nil {
		t.Fatalf("block user2: %v", err)
	}

	matchesAfter, err := socialSvc.Matches(context.Background(), user1.ID)
	if err != nil {
		t.Fatalf("matches after block: %v", err)
	}
	if containsDiscoverUser(matchesAfter, user2.ID) {
		t.Fatalf("expected blocked user to be removed from matches")
	}

	_, err = socialSvc.Like(context.Background(), user2.ID, user1.ID)
	if !errors.Is(err, ErrInteractionBlock) {
		t.Fatalf("expected ErrInteractionBlock after block, got %v", err)
	}

	discoverAfter, err := socialSvc.Discover(context.Background(), user1.ID)
	if err != nil {
		t.Fatalf("discover after block: %v", err)
	}
	if containsDiscoverUser(discoverAfter, user2.ID) {
		t.Fatalf("expected blocked user to be hidden from discover")
	}
}

func TestChatServiceFlowAndGuards(t *testing.T) {
	pool := setupServicesTestDB(t)
	defer pool.Close()

	authRepo := repositories.NewPGAuthRepository(pool)
	authSvc := NewAuthService(authRepo, []byte("services-test-secret"))
	socialRepo := repositories.NewPGSocialRepository(pool)
	socialSvc := NewSocialService(socialRepo)
	chatRepo := repositories.NewPGChatRepository(pool)
	chatSvc := NewChatService(chatRepo)

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

	_, err = chatSvc.SendMessage(context.Background(), user1.ID, user3.ID, "should fail")
	if !errors.Is(err, ErrChatRequiresMatch) {
		t.Fatalf("expected ErrChatRequiresMatch without match, got %v", err)
	}

	if _, err := socialSvc.Like(context.Background(), user1.ID, user2.ID); err != nil {
		t.Fatalf("like user1->user2: %v", err)
	}
	if _, err := socialSvc.Like(context.Background(), user2.ID, user1.ID); err != nil {
		t.Fatalf("like user2->user1: %v", err)
	}

	_, err = chatSvc.SendMessage(context.Background(), user1.ID, user2.ID, "   ")
	if !errors.Is(err, ErrMessageContentNeeded) {
		t.Fatalf("expected ErrMessageContentNeeded, got %v", err)
	}

	sent, err := chatSvc.SendMessage(context.Background(), user1.ID, user2.ID, "hello service chat")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if sent.Content != "hello service chat" {
		t.Fatalf("unexpected sent message content: %q", sent.Content)
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
	if !containsChatSummary(chats, user2.ID) {
		t.Fatalf("expected chats list to include matched user")
	}

	if err := socialSvc.Block(context.Background(), user1.ID, user2.ID); err != nil {
		t.Fatalf("block user2: %v", err)
	}

	_, err = chatSvc.SendMessage(context.Background(), user1.ID, user2.ID, "blocked now")
	if !errors.Is(err, ErrInteractionBlock) {
		t.Fatalf("expected ErrInteractionBlock after block, got %v", err)
	}

	_, err = chatSvc.ListMessages(context.Background(), user1.ID, user2.ID)
	if !errors.Is(err, ErrInteractionBlock) {
		t.Fatalf("expected ErrInteractionBlock on list after block, got %v", err)
	}
}

func setupServicesTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL_TEST")
	if databaseURL == "" {
		databaseURL = "postgresql://app:app@localhost:5432/app?sslmode=disable"
	}

	pool, err := db.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("skip services integration tests: postgres unavailable (%v)", err)
	}

	if err := db.RunMigrations(context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("run migrations: %v", err)
	}

	return pool
}

func uniqueEmail(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000") + "@lovr.test"
}

func cleanupUsers(t *testing.T, pool *pgxpool.Pool, emails []string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `DELETE FROM users WHERE email = ANY($1)`, emails)
	if err != nil {
		t.Fatalf("cleanup users: %v", err)
	}
}

func containsDiscoverUser(users []DiscoverUser, userID string) bool {
	for _, user := range users {
		if user.ID == userID {
			return true
		}
	}
	return false
}

func containsChatSummary(chats []ChatSummary, userID string) bool {
	for _, chat := range chats {
		if chat.UserID == userID {
			return true
		}
	}
	return false
}
