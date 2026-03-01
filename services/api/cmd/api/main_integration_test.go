package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"example.com/api/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type testResponse struct {
	Status int
	Body   []byte
}

func TestHealthEndpoint(t *testing.T) {
	router, pool := setupTestRouter(t)
	defer pool.Close()

	resp := doRequest(t, router, http.MethodGet, "/health", nil, "")
	if resp.Status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", resp.Status, string(resp.Body))
	}
	if !strings.Contains(string(resp.Body), `"status":"ok"`) {
		t.Fatalf("unexpected health body: %s", string(resp.Body))
	}
}

func TestAuthAndSocialFlow(t *testing.T) {
	router, pool := setupTestRouter(t)
	defer pool.Close()

	email1 := uniqueEmail("flow1")
	email2 := uniqueEmail("flow2")
	password := "Password123"
	defer cleanupUsers(t, pool, []string{email1, email2})

	register1 := doRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email1,
		"password": password,
	}, "")
	if register1.Status != http.StatusCreated {
		t.Fatalf("register1 expected 201, got %d body=%s", register1.Status, string(register1.Body))
	}

	register2 := doRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email2,
		"password": password,
	}, "")
	if register2.Status != http.StatusCreated {
		t.Fatalf("register2 expected 201, got %d body=%s", register2.Status, string(register2.Body))
	}

	var auth1 authResponse
	if err := json.Unmarshal(register1.Body, &auth1); err != nil {
		t.Fatalf("unmarshal register1: %v", err)
	}
	var auth2 authResponse
	if err := json.Unmarshal(register2.Body, &auth2); err != nil {
		t.Fatalf("unmarshal register2: %v", err)
	}
	if auth1.Token == "" || auth2.Token == "" {
		t.Fatalf("expected non-empty JWT tokens")
	}

	meResp := doRequest(t, router, http.MethodGet, "/me", nil, auth1.Token)
	if meResp.Status != http.StatusOK {
		t.Fatalf("/me expected 200, got %d body=%s", meResp.Status, string(meResp.Body))
	}
	var meUser meResponse
	if err := json.Unmarshal(meResp.Body, &meUser); err != nil {
		t.Fatalf("unmarshal /me: %v", err)
	}
	if meUser.Email != email1 {
		t.Fatalf("expected /me email %s, got %s", email1, meUser.Email)
	}

	discoverBefore := doRequest(t, router, http.MethodGet, "/discover", nil, auth1.Token)
	if discoverBefore.Status != http.StatusOK {
		t.Fatalf("/discover before like expected 200, got %d body=%s", discoverBefore.Status, string(discoverBefore.Body))
	}
	var discoverUsers discoverResponse
	if err := json.Unmarshal(discoverBefore.Body, &discoverUsers); err != nil {
		t.Fatalf("unmarshal /discover before: %v", err)
	}
	if !containsUser(discoverUsers.Users, auth2.User.ID) {
		t.Fatalf("expected discover list to contain second user before interactions")
	}

	like1 := doRequest(t, router, http.MethodPost, "/likes", map[string]string{
		"toUserId": auth2.User.ID,
	}, auth1.Token)
	if like1.Status != http.StatusOK {
		t.Fatalf("first like expected 200, got %d body=%s", like1.Status, string(like1.Body))
	}
	var like1Payload likeResponse
	if err := json.Unmarshal(like1.Body, &like1Payload); err != nil {
		t.Fatalf("unmarshal first like: %v", err)
	}
	if like1Payload.Matched {
		t.Fatalf("expected first like to be unmatched")
	}

	like2 := doRequest(t, router, http.MethodPost, "/likes", map[string]string{
		"toUserId": auth1.User.ID,
	}, auth2.Token)
	if like2.Status != http.StatusOK {
		t.Fatalf("second like expected 200, got %d body=%s", like2.Status, string(like2.Body))
	}
	var like2Payload likeResponse
	if err := json.Unmarshal(like2.Body, &like2Payload); err != nil {
		t.Fatalf("unmarshal second like: %v", err)
	}
	if !like2Payload.Matched {
		t.Fatalf("expected reciprocal like to match")
	}

	matchesBeforeBlock := doRequest(t, router, http.MethodGet, "/matches", nil, auth1.Token)
	if matchesBeforeBlock.Status != http.StatusOK {
		t.Fatalf("/matches before block expected 200, got %d body=%s", matchesBeforeBlock.Status, string(matchesBeforeBlock.Body))
	}
	var matchesPayload matchesResponse
	if err := json.Unmarshal(matchesBeforeBlock.Body, &matchesPayload); err != nil {
		t.Fatalf("unmarshal /matches before block: %v", err)
	}
	if !containsUser(matchesPayload.Matches, auth2.User.ID) {
		t.Fatalf("expected matches list to contain second user before block")
	}

	blockResp := doRequest(t, router, http.MethodPost, "/block", map[string]string{
		"toUserId": auth2.User.ID,
	}, auth1.Token)
	if blockResp.Status != http.StatusOK {
		t.Fatalf("/block expected 200, got %d body=%s", blockResp.Status, string(blockResp.Body))
	}

	matchesAfterBlock1 := doRequest(t, router, http.MethodGet, "/matches", nil, auth1.Token)
	if matchesAfterBlock1.Status != http.StatusOK {
		t.Fatalf("/matches after block (u1) expected 200, got %d body=%s", matchesAfterBlock1.Status, string(matchesAfterBlock1.Body))
	}
	var matchesAfter1 matchesResponse
	if err := json.Unmarshal(matchesAfterBlock1.Body, &matchesAfter1); err != nil {
		t.Fatalf("unmarshal /matches after block (u1): %v", err)
	}
	if containsUser(matchesAfter1.Matches, auth2.User.ID) {
		t.Fatalf("expected blocked user to be removed from matches (u1)")
	}

	matchesAfterBlock2 := doRequest(t, router, http.MethodGet, "/matches", nil, auth2.Token)
	if matchesAfterBlock2.Status != http.StatusOK {
		t.Fatalf("/matches after block (u2) expected 200, got %d body=%s", matchesAfterBlock2.Status, string(matchesAfterBlock2.Body))
	}
	var matchesAfter2 matchesResponse
	if err := json.Unmarshal(matchesAfterBlock2.Body, &matchesAfter2); err != nil {
		t.Fatalf("unmarshal /matches after block (u2): %v", err)
	}
	if containsUser(matchesAfter2.Matches, auth1.User.ID) {
		t.Fatalf("expected blocker user to be removed from matches (u2)")
	}

	likeAfterBlock12 := doRequest(t, router, http.MethodPost, "/likes", map[string]string{
		"toUserId": auth2.User.ID,
	}, auth1.Token)
	if likeAfterBlock12.Status != http.StatusForbidden {
		t.Fatalf("expected like blocker->blocked to return 403, got %d body=%s", likeAfterBlock12.Status, string(likeAfterBlock12.Body))
	}

	likeAfterBlock21 := doRequest(t, router, http.MethodPost, "/likes", map[string]string{
		"toUserId": auth1.User.ID,
	}, auth2.Token)
	if likeAfterBlock21.Status != http.StatusForbidden {
		t.Fatalf("expected like blocked->blocker to return 403, got %d body=%s", likeAfterBlock21.Status, string(likeAfterBlock21.Body))
	}

	discoverAfterBlock1 := doRequest(t, router, http.MethodGet, "/discover", nil, auth1.Token)
	if discoverAfterBlock1.Status != http.StatusOK {
		t.Fatalf("/discover after block (u1) expected 200, got %d body=%s", discoverAfterBlock1.Status, string(discoverAfterBlock1.Body))
	}
	var discoverAfter1 discoverResponse
	if err := json.Unmarshal(discoverAfterBlock1.Body, &discoverAfter1); err != nil {
		t.Fatalf("unmarshal discover after block (u1): %v", err)
	}
	if containsUser(discoverAfter1.Users, auth2.User.ID) {
		t.Fatalf("expected blocked user to be hidden from discover (u1)")
	}

	discoverAfterBlock2 := doRequest(t, router, http.MethodGet, "/discover", nil, auth2.Token)
	if discoverAfterBlock2.Status != http.StatusOK {
		t.Fatalf("/discover after block (u2) expected 200, got %d body=%s", discoverAfterBlock2.Status, string(discoverAfterBlock2.Body))
	}
	var discoverAfter2 discoverResponse
	if err := json.Unmarshal(discoverAfterBlock2.Body, &discoverAfter2); err != nil {
		t.Fatalf("unmarshal discover after block (u2): %v", err)
	}
	if containsUser(discoverAfter2.Users, auth1.User.ID) {
		t.Fatalf("expected blocker user to be hidden from discover (u2)")
	}
}

func TestUnauthorizedEndpoints(t *testing.T) {
	router, pool := setupTestRouter(t)
	defer pool.Close()

	for _, endpoint := range []string{"/me", "/discover", "/matches", "/chats", "/chats/11111111-1111-1111-1111-111111111111/messages"} {
		resp := doRequest(t, router, http.MethodGet, endpoint, nil, "")
		if resp.Status != http.StatusUnauthorized {
			t.Fatalf("%s expected 401 without token, got %d body=%s", endpoint, resp.Status, string(resp.Body))
		}
	}
}

func TestChatFlowRequiresMatchAndRespectsBlock(t *testing.T) {
	router, pool := setupTestRouter(t)
	defer pool.Close()

	email1 := uniqueEmail("chat1")
	email2 := uniqueEmail("chat2")
	email3 := uniqueEmail("chat3")
	password := "Password123"
	defer cleanupUsers(t, pool, []string{email1, email2, email3})

	register1 := doRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email1,
		"password": password,
	}, "")
	register2 := doRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email2,
		"password": password,
	}, "")
	register3 := doRequest(t, router, http.MethodPost, "/auth/register", map[string]string{
		"email":    email3,
		"password": password,
	}, "")

	if register1.Status != http.StatusCreated || register2.Status != http.StatusCreated || register3.Status != http.StatusCreated {
		t.Fatalf("registers expected 201, got [%d, %d, %d]", register1.Status, register2.Status, register3.Status)
	}

	var auth1 authResponse
	var auth2 authResponse
	var auth3 authResponse
	if err := json.Unmarshal(register1.Body, &auth1); err != nil {
		t.Fatalf("unmarshal register1: %v", err)
	}
	if err := json.Unmarshal(register2.Body, &auth2); err != nil {
		t.Fatalf("unmarshal register2: %v", err)
	}
	if err := json.Unmarshal(register3.Body, &auth3); err != nil {
		t.Fatalf("unmarshal register3: %v", err)
	}

	like1 := doRequest(t, router, http.MethodPost, "/likes", map[string]string{"toUserId": auth2.User.ID}, auth1.Token)
	like2 := doRequest(t, router, http.MethodPost, "/likes", map[string]string{"toUserId": auth1.User.ID}, auth2.Token)
	if like1.Status != http.StatusOK || like2.Status != http.StatusOK {
		t.Fatalf("expected reciprocal likes to succeed, got [%d, %d]", like1.Status, like2.Status)
	}

	send := doRequest(t, router, http.MethodPost, "/chats/"+auth2.User.ID+"/messages", map[string]string{
		"content": "hello there",
	}, auth1.Token)
	if send.Status != http.StatusCreated {
		t.Fatalf("send message expected 201, got %d body=%s", send.Status, string(send.Body))
	}
	var sent chatMessageResponse
	if err := json.Unmarshal(send.Body, &sent); err != nil {
		t.Fatalf("unmarshal sent message: %v", err)
	}
	if sent.Content != "hello there" {
		t.Fatalf("unexpected message content: %s", sent.Content)
	}

	getMessages := doRequest(t, router, http.MethodGet, "/chats/"+auth2.User.ID+"/messages", nil, auth1.Token)
	if getMessages.Status != http.StatusOK {
		t.Fatalf("get messages expected 200, got %d body=%s", getMessages.Status, string(getMessages.Body))
	}
	var messagesPayload chatMessagesResponse
	if err := json.Unmarshal(getMessages.Body, &messagesPayload); err != nil {
		t.Fatalf("unmarshal get messages: %v", err)
	}
	if len(messagesPayload.Messages) == 0 {
		t.Fatalf("expected at least one chat message")
	}

	getChats := doRequest(t, router, http.MethodGet, "/chats", nil, auth1.Token)
	if getChats.Status != http.StatusOK {
		t.Fatalf("get chats expected 200, got %d body=%s", getChats.Status, string(getChats.Body))
	}
	var chatsPayload chatsResponse
	if err := json.Unmarshal(getChats.Body, &chatsPayload); err != nil {
		t.Fatalf("unmarshal get chats: %v", err)
	}
	if !containsChat(chatsPayload.Chats, auth2.User.ID) {
		t.Fatalf("expected chats list to include matched user")
	}

	sendWithoutMatch := doRequest(t, router, http.MethodPost, "/chats/"+auth3.User.ID+"/messages", map[string]string{
		"content": "should fail",
	}, auth1.Token)
	if sendWithoutMatch.Status != http.StatusForbidden {
		t.Fatalf("send without match expected 403, got %d body=%s", sendWithoutMatch.Status, string(sendWithoutMatch.Body))
	}

	blockResp := doRequest(t, router, http.MethodPost, "/block", map[string]string{"toUserId": auth2.User.ID}, auth1.Token)
	if blockResp.Status != http.StatusOK {
		t.Fatalf("block expected 200, got %d body=%s", blockResp.Status, string(blockResp.Body))
	}

	sendAfterBlock := doRequest(t, router, http.MethodPost, "/chats/"+auth2.User.ID+"/messages", map[string]string{
		"content": "blocked now",
	}, auth1.Token)
	if sendAfterBlock.Status != http.StatusForbidden {
		t.Fatalf("send after block expected 403, got %d body=%s", sendAfterBlock.Status, string(sendAfterBlock.Body))
	}

	getAfterBlock := doRequest(t, router, http.MethodGet, "/chats/"+auth2.User.ID+"/messages", nil, auth1.Token)
	if getAfterBlock.Status != http.StatusForbidden {
		t.Fatalf("get messages after block expected 403, got %d body=%s", getAfterBlock.Status, string(getAfterBlock.Body))
	}
}

func setupTestRouter(t *testing.T) (*gin.Engine, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL_TEST")
	if databaseURL == "" {
		databaseURL = "postgresql://app:app@localhost:5432/app?sslmode=disable"
	}

	pool, err := db.Connect(context.Background(), databaseURL)
	if err != nil {
		t.Skipf("skip integration tests: postgres unavailable (%v)", err)
	}

	if err := db.RunMigrations(context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("run migrations: %v", err)
	}

	a := &app{
		dbPool:    pool,
		jwtSecret: []byte("integration-test-secret"),
	}

	return setupRouter(a), pool
}

func doRequest(t *testing.T, handler http.Handler, method, path string, payload any, token string) testResponse {
	t.Helper()

	var bodyBytes []byte
	var err error
	if payload != nil {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return testResponse{Status: rec.Code, Body: rec.Body.Bytes()}
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

func containsUser(users []discoverUserResponse, userID string) bool {
	for _, user := range users {
		if user.ID == userID {
			return true
		}
	}
	return false
}

func containsChat(chats []chatSummaryResponse, userID string) bool {
	for _, chat := range chats {
		if chat.User.ID == userID {
			return true
		}
	}
	return false
}
