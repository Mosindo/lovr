package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"example.com/api/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type app struct {
	dbPool    *pgxpool.Pool
	jwtSecret []byte
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token string     `json:"token"`
	User  meResponse `json:"user"`
}

type discoverUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type discoverResponse struct {
	Users []discoverUserResponse `json:"users"`
}

type likeRequest struct {
	ToUserID string `json:"toUserId" binding:"required,uuid"`
}

type likeResponse struct {
	Matched bool `json:"matched"`
}

type blockRequest struct {
	ToUserID string `json:"toUserId" binding:"required,uuid"`
}

type blockResponse struct {
	Blocked bool `json:"blocked"`
}

type matchesResponse struct {
	Matches []discoverUserResponse `json:"matches"`
}

type chatMessagePreview struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type chatSummaryResponse struct {
	User        discoverUserResponse `json:"user"`
	LastMessage *chatMessagePreview  `json:"lastMessage,omitempty"`
}

type chatsResponse struct {
	Chats []chatSummaryResponse `json:"chats"`
}

type sendMessageRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type chatMessageResponse struct {
	ID              string    `json:"id"`
	SenderUserID    string    `json:"senderUserId"`
	RecipientUserID string    `json:"recipientUserId"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
}

type chatMessagesResponse struct {
	Messages []chatMessageResponse `json:"messages"`
}

type meResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type userClaims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

func main() {
	port := getenv("PORT", "8080")
	databaseURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatal(err)
	}

	a := &app{dbPool: pool, jwtSecret: []byte(jwtSecret)}

	r := setupRouter(a)

	addr := ":" + port
	log.Printf("api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func setupRouter(a *app) *gin.Engine {
	r := gin.Default()
	r.GET("/health", a.health)
	r.POST("/auth/register", a.register)
	r.POST("/auth/login", a.login)
	r.GET("/me", a.requireUser(), a.me)
	r.GET("/discover", a.requireUser(), a.discover)
	r.POST("/likes", a.requireUser(), a.like)
	r.POST("/block", a.requireUser(), a.block)
	r.GET("/matches", a.requireUser(), a.matches)
	r.GET("/chats", a.requireUser(), a.chats)
	r.GET("/chats/:userId/messages", a.requireUser(), a.chatMessages)
	r.POST("/chats/:userId/messages", a.requireUser(), a.sendChatMessage)
	return r
}

func (a *app) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (a *app) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	var user meResponse
	err = a.dbPool.QueryRow(c.Request.Context(), `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at
	`, email, string(hash)).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	token, err := a.signToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (a *app) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	var user meResponse
	var passwordHash string
	err := a.dbPool.QueryRow(c.Request.Context(), `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not login"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := a.signToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{Token: token, User: user})
}

func (a *app) me(c *gin.Context) {
	userID := c.GetString("userID")
	var user meResponse
	err := a.dbPool.QueryRow(c.Request.Context(), `
		SELECT id, email, created_at
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (a *app) discover(c *gin.Context) {
	userID := c.GetString("userID")
	rows, err := a.dbPool.Query(c.Request.Context(), `
		SELECT u.id, u.email, u.created_at
		FROM users u
		WHERE u.id <> $1
		  AND NOT EXISTS (
			SELECT 1
			FROM likes l
			WHERE l.from_user_id = $1
			  AND l.to_user_id = u.id
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = u.id)
			   OR (b.blocker_user_id = u.id AND b.blocked_user_id = $1)
		  )
		ORDER BY u.created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch discover users"})
		return
	}
	defer rows.Close()

	users := make([]discoverUserResponse, 0)
	for rows.Next() {
		var user discoverUserResponse
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch discover users"})
			return
		}
		users = append(users, user)
	}

	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch discover users"})
		return
	}

	c.JSON(http.StatusOK, discoverResponse{Users: users})
}

func (a *app) like(c *gin.Context) {
	userID := c.GetString("userID")
	var req likeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	toUserID := strings.TrimSpace(req.ToUserID)
	if toUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot like yourself"})
		return
	}

	var exists bool
	err := a.dbPool.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`, toUserID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save like"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var blocked bool
	err = a.dbPool.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM blocks
			WHERE (blocker_user_id = $1 AND blocked_user_id = $2)
			   OR (blocker_user_id = $2 AND blocked_user_id = $1)
		)
	`, userID, toUserID).Scan(&blocked)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save like"})
		return
	}
	if blocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "interaction blocked"})
		return
	}

	if _, err := a.dbPool.Exec(c.Request.Context(), `
		INSERT INTO likes (from_user_id, to_user_id)
		VALUES ($1, $2)
		ON CONFLICT (from_user_id, to_user_id) DO NOTHING
	`, userID, toUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save like"})
		return
	}

	var matched bool
	err = a.dbPool.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM likes
			WHERE from_user_id = $1
			  AND to_user_id = $2
		)
	`, toUserID, userID).Scan(&matched)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute match"})
		return
	}

	c.JSON(http.StatusOK, likeResponse{Matched: matched})
}

func (a *app) block(c *gin.Context) {
	userID := c.GetString("userID")
	var req blockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	toUserID := strings.TrimSpace(req.ToUserID)
	if toUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
		return
	}

	var exists bool
	err := a.dbPool.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`, toUserID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not block user"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	tx, err := a.dbPool.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not block user"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	if _, err := tx.Exec(c.Request.Context(), `
		INSERT INTO blocks (blocker_user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT (blocker_user_id, blocked_user_id) DO NOTHING
	`, userID, toUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not block user"})
		return
	}

	if _, err := tx.Exec(c.Request.Context(), `
		DELETE FROM likes
		WHERE (from_user_id = $1 AND to_user_id = $2)
		   OR (from_user_id = $2 AND to_user_id = $1)
	`, userID, toUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not block user"})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not block user"})
		return
	}

	c.JSON(http.StatusOK, blockResponse{Blocked: true})
}

func (a *app) matches(c *gin.Context) {
	userID := c.GetString("userID")
	rows, err := a.dbPool.Query(c.Request.Context(), `
		SELECT u.id, u.email, u.created_at
		FROM likes sent
		JOIN likes received
		  ON received.from_user_id = sent.to_user_id
		 AND received.to_user_id = sent.from_user_id
		JOIN users u
		  ON u.id = sent.to_user_id
		WHERE sent.from_user_id = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = sent.to_user_id)
			   OR (b.blocker_user_id = sent.to_user_id AND b.blocked_user_id = $1)
		  )
		ORDER BY sent.created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch matches"})
		return
	}
	defer rows.Close()

	matches := make([]discoverUserResponse, 0)
	for rows.Next() {
		var user discoverUserResponse
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch matches"})
			return
		}
		matches = append(matches, user)
	}

	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch matches"})
		return
	}

	c.JSON(http.StatusOK, matchesResponse{Matches: matches})
}

func (a *app) chats(c *gin.Context) {
	userID := c.GetString("userID")
	rows, err := a.dbPool.Query(c.Request.Context(), `
		SELECT
			u.id,
			u.email,
			u.created_at,
			lm.content,
			lm.created_at
		FROM likes sent
		JOIN likes received
		  ON received.from_user_id = sent.to_user_id
		 AND received.to_user_id = sent.from_user_id
		JOIN users u
		  ON u.id = sent.to_user_id
		LEFT JOIN LATERAL (
			SELECT m.content, m.created_at
			FROM messages m
			WHERE (m.sender_user_id = $1 AND m.recipient_user_id = sent.to_user_id)
			   OR (m.sender_user_id = sent.to_user_id AND m.recipient_user_id = $1)
			ORDER BY m.created_at DESC
			LIMIT 1
		) lm ON true
		WHERE sent.from_user_id = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM blocks b
			WHERE (b.blocker_user_id = $1 AND b.blocked_user_id = sent.to_user_id)
			   OR (b.blocker_user_id = sent.to_user_id AND b.blocked_user_id = $1)
		  )
		ORDER BY COALESCE(lm.created_at, u.created_at) DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch chats"})
		return
	}
	defer rows.Close()

	chats := make([]chatSummaryResponse, 0)
	for rows.Next() {
		var user discoverUserResponse
		var lastContent sql.NullString
		var lastCreatedAt sql.NullTime
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt, &lastContent, &lastCreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch chats"})
			return
		}

		chat := chatSummaryResponse{User: user}
		if lastContent.Valid && lastCreatedAt.Valid {
			chat.LastMessage = &chatMessagePreview{
				Content:   lastContent.String,
				CreatedAt: lastCreatedAt.Time,
			}
		}
		chats = append(chats, chat)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch chats"})
		return
	}

	c.JSON(http.StatusOK, chatsResponse{Chats: chats})
}

func (a *app) chatMessages(c *gin.Context) {
	userID := c.GetString("userID")
	otherUserID := strings.TrimSpace(c.Param("userId"))
	if !isUUIDLike(otherUserID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if otherUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open self chat"})
		return
	}

	if allowed, status, msg := a.ensureCanChat(c.Request.Context(), userID, otherUserID); !allowed {
		c.JSON(status, gin.H{"error": msg})
		return
	}

	rows, err := a.dbPool.Query(c.Request.Context(), `
		SELECT id, sender_user_id, recipient_user_id, content, created_at
		FROM messages
		WHERE (sender_user_id = $1 AND recipient_user_id = $2)
		   OR (sender_user_id = $2 AND recipient_user_id = $1)
		ORDER BY created_at ASC
		LIMIT 200
	`, userID, otherUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch messages"})
		return
	}
	defer rows.Close()

	messages := make([]chatMessageResponse, 0)
	for rows.Next() {
		var message chatMessageResponse
		if err := rows.Scan(&message.ID, &message.SenderUserID, &message.RecipientUserID, &message.Content, &message.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch messages"})
			return
		}
		messages = append(messages, message)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch messages"})
		return
	}

	c.JSON(http.StatusOK, chatMessagesResponse{Messages: messages})
}

func (a *app) sendChatMessage(c *gin.Context) {
	userID := c.GetString("userID")
	otherUserID := strings.TrimSpace(c.Param("userId"))
	if !isUUIDLike(otherUserID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if otherUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot message yourself"})
		return
	}

	if allowed, status, msg := a.ensureCanChat(c.Request.Context(), userID, otherUserID); !allowed {
		c.JSON(status, gin.H{"error": msg})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message content required"})
		return
	}

	var message chatMessageResponse
	err := a.dbPool.QueryRow(c.Request.Context(), `
		INSERT INTO messages (sender_user_id, recipient_user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, sender_user_id, recipient_user_id, content, created_at
	`, userID, otherUserID, content).Scan(
		&message.ID,
		&message.SenderUserID,
		&message.RecipientUserID,
		&message.Content,
		&message.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not send message"})
		return
	}

	c.JSON(http.StatusCreated, message)
}

func (a *app) ensureCanChat(ctx context.Context, userID, otherUserID string) (bool, int, string) {
	exists, err := a.userExists(ctx, otherUserID)
	if err != nil {
		return false, http.StatusInternalServerError, "could not validate chat target"
	}
	if !exists {
		return false, http.StatusNotFound, "user not found"
	}

	blocked, err := a.usersBlocked(ctx, userID, otherUserID)
	if err != nil {
		return false, http.StatusInternalServerError, "could not validate chat access"
	}
	if blocked {
		return false, http.StatusForbidden, "interaction blocked"
	}

	matched, err := a.usersMatched(ctx, userID, otherUserID)
	if err != nil {
		return false, http.StatusInternalServerError, "could not validate chat access"
	}
	if !matched {
		return false, http.StatusForbidden, "chat allowed only after match"
	}

	return true, 0, ""
}

func (a *app) userExists(ctx context.Context, userID string) (bool, error) {
	var exists bool
	err := a.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)
	`, userID).Scan(&exists)
	return exists, err
}

func (a *app) usersBlocked(ctx context.Context, userID, otherUserID string) (bool, error) {
	var blocked bool
	err := a.dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM blocks
			WHERE (blocker_user_id = $1 AND blocked_user_id = $2)
			   OR (blocker_user_id = $2 AND blocked_user_id = $1)
		)
	`, userID, otherUserID).Scan(&blocked)
	return blocked, err
}

func (a *app) usersMatched(ctx context.Context, userID, otherUserID string) (bool, error) {
	var matched bool
	err := a.dbPool.QueryRow(ctx, `
		SELECT
			EXISTS (
				SELECT 1
				FROM likes
				WHERE from_user_id = $1
				  AND to_user_id = $2
			)
			AND EXISTS (
				SELECT 1
				FROM likes
				WHERE from_user_id = $2
				  AND to_user_id = $1
			)
	`, userID, otherUserID).Scan(&matched)
	return matched, err
}

func (a *app) requireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := bearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &userClaims{}, func(token *jwt.Token) (interface{}, error) {
			if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected signing method")
			}
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method type")
			}
			return a.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(*userClaims)
		if !ok || claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func (a *app) signToken(userID string) (string, error) {
	claims := userClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func bearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func isUUIDLike(v string) bool {
	if len(v) != 36 {
		return false
	}
	hyphenPos := map[int]struct{}{8: {}, 13: {}, 18: {}, 23: {}}
	for i := 0; i < len(v); i++ {
		ch := v[i]
		if _, ok := hyphenPos[i]; ok {
			if ch != '-' {
				return false
			}
			continue
		}
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}
