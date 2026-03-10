package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	authmw "example.com/api/internal/auth"
	"example.com/api/internal/db"
	"example.com/api/internal/handlers"
	"example.com/api/internal/repositories"
	"example.com/api/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

type meResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type metricKey struct {
	Method string
	Path   string
	Status int
}

type metricValue struct {
	Count         int64
	LatencyTotal  int64
	LatencyMax    int64
	LastUpdatedAt time.Time
}

type metricsStore struct {
	mu   sync.Mutex
	data map[metricKey]metricValue
}

var httpMetrics = metricsStore{
	data: make(map[metricKey]metricValue),
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
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestIDMiddleware())
	r.Use(requestStartMiddleware())
	r.Use(requestMetricsMiddleware())
	r.Use(requestLogger())
	authRepo := repositories.NewPGAuthRepository(a.dbPool)
	authService := services.NewAuthService(authRepo, a.jwtSecret)
	authHandler := handlers.NewAuthHandler(authService)
	socialRepo := repositories.NewPGSocialRepository(a.dbPool)
	socialService := services.NewSocialService(socialRepo)
	socialHandler := handlers.NewSocialHandler(socialService)
	chatRepo := repositories.NewPGChatRepository(a.dbPool)
	chatService := services.NewChatService(chatRepo)
	chatHandler := handlers.NewChatHandler(chatService)
	requireUser := authmw.RequireUser(a.jwtSecret)

	r.GET("/health", a.health)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)
	r.GET("/me", requireUser, authHandler.Me)
	r.GET("/discover", requireUser, socialHandler.Discover)
	r.POST("/likes", requireUser, socialHandler.Like)
	r.POST("/block", requireUser, socialHandler.Block)
	r.GET("/matches", requireUser, socialHandler.Matches)
	r.GET("/chats", requireUser, chatHandler.Chats)
	r.GET("/chats/:userId/messages", requireUser, chatHandler.ChatMessages)
	r.POST("/chats/:userId/messages", requireUser, chatHandler.SendChatMessage)
	return r
}

func requestMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		method := c.Request.Method

		key := metricKey{Method: method, Path: path, Status: status}
		httpMetrics.mu.Lock()
		current := httpMetrics.data[key]
		current.Count++
		current.LatencyTotal += latency
		if latency > current.LatencyMax {
			current.LatencyMax = latency
		}
		current.LastUpdatedAt = time.Now().UTC()
		httpMetrics.data[key] = current
		total := current.Count
		avg := int64(0)
		if current.Count > 0 {
			avg = current.LatencyTotal / current.Count
		}
		httpMetrics.mu.Unlock()

		if total%50 == 0 {
			log.Printf(
				`{"event":"http_metric","method":%q,"path":%q,"status":%d,"count":%d,"avg_latency_ms":%d,"max_latency_ms":%d}`,
				method,
				path,
				status,
				total,
				avg,
				current.LatencyMax,
			)
		}
	}
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		userID := c.GetString("userID")

		log.Printf(
			`{"event":"http_request","request_id":%q,"method":%q,"path":%q,"status":%d,"latency_ms":%d,"client_ip":%q,"user_id":%q}`,
			c.GetString("request_id"),
			c.Request.Method,
			path,
			c.Writer.Status(),
			requestLatencyFromContext(c),
			c.ClientIP(),
			userID,
		)
	}
}

func requestStartMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("request_started_at", time.Now())
		c.Next()
	}
}

func requestLatencyFromContext(c *gin.Context) int64 {
	startedAt, ok := c.Get("request_started_at")
	if !ok {
		return 0
	}
	start, ok := startedAt.(time.Time)
	if !ok {
		return 0
	}
	return time.Since(start).Milliseconds()
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func generateRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

func (a *app) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
