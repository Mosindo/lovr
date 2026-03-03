package main

import (
	"context"
	"log"
	"net/http"
	"os"
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
