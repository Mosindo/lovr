package main

import (
	"context"
	"log"
	"net/http"
	"time"

	authfeature "example.com/api/internal/features/auth"
	billingfeature "example.com/api/internal/features/billing"
	chatfeature "example.com/api/internal/features/chat"
	commentsfeature "example.com/api/internal/features/comments"
	filesfeature "example.com/api/internal/features/files"
	notificationsfeature "example.com/api/internal/features/notifications"
	postsfeature "example.com/api/internal/features/posts"
	usersfeature "example.com/api/internal/features/users"
	"example.com/api/internal/platform/config"
	"example.com/api/internal/platform/db"
	platformhandlers "example.com/api/internal/platform/handlers"
	"example.com/api/internal/platform/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type app struct {
	dbPool              *pgxpool.Pool
	jwtSecret           []byte
	stripeSecretKey     string
	stripeWebhookSecret string
	stripePriceID       string
	appBaseURL          string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := db.ConnectWithRetry(ctx, cfg.DatabaseURL, 10*time.Second, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatal(err)
	}

	a := &app{
		dbPool:              pool,
		jwtSecret:           []byte(cfg.JWTSecret),
		stripeSecretKey:     cfg.StripeSecretKey,
		stripeWebhookSecret: cfg.StripeWebhookSecret,
		stripePriceID:       cfg.StripePriceID,
		appBaseURL:          cfg.AppBaseURL,
	}

	r := setupRouter(a)

	addr := ":" + cfg.Port
	log.Printf("api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func setupRouter(a *app) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestStart())
	r.Use(middleware.RequestMetrics())
	r.Use(middleware.RequestLogger())
	authRepo := authfeature.NewPGRepository(a.dbPool)
	authService := authfeature.NewService(authRepo, a.jwtSecret)
	authHandler := authfeature.NewHandler(authService)
	usersRepo := usersfeature.NewPGRepository(a.dbPool)
	usersService := usersfeature.NewService(usersRepo)
	usersHandler := usersfeature.NewHandler(usersService)
	chatRepo := chatfeature.NewPGRepository(a.dbPool)
	chatService := chatfeature.NewService(chatRepo)
	chatHandler := chatfeature.NewHandler(chatService)
	postsRepo := postsfeature.NewPGRepository(a.dbPool)
	postsService := postsfeature.NewService(postsRepo)
	postsHandler := postsfeature.NewHandler(postsService)
	commentsRepo := commentsfeature.NewPGRepository(a.dbPool)
	commentsService := commentsfeature.NewService(commentsRepo)
	commentsHandler := commentsfeature.NewHandler(commentsService)
	notificationsRepo := notificationsfeature.NewPGRepository(a.dbPool)
	notificationsService := notificationsfeature.NewService(notificationsRepo)
	notificationsHandler := notificationsfeature.NewHandler(notificationsService)
	filesRepo := filesfeature.NewPGRepository(a.dbPool)
	filesService := filesfeature.NewService(filesRepo)
	filesHandler := filesfeature.NewHandler(filesService)
	billingRepo := billingfeature.NewPGRepository(a.dbPool)
	billingService := billingfeature.NewService(billingRepo, billingfeature.Config{
		StripeSecretKey:     a.stripeSecretKey,
		StripeWebhookSecret: a.stripeWebhookSecret,
		StripePriceID:       a.stripePriceID,
		AppBaseURL:          a.appBaseURL,
		HTTPClient:          &http.Client{Timeout: 10 * time.Second},
	})
	billingHandler := billingfeature.NewHandler(billingService)
	requireUser := middleware.RequireUser(a.jwtSecret)

	r.GET("/health", platformhandlers.NewHealthHandler(a.dbPool))
	authfeature.RegisterRoutes(r, authHandler, requireUser)
	usersfeature.RegisterRoutes(r, usersHandler, requireUser)
	chatfeature.RegisterRoutes(r, chatHandler, requireUser)
	postsfeature.RegisterRoutes(r, postsHandler, requireUser)
	commentsfeature.RegisterRoutes(r, commentsHandler, requireUser)
	notificationsfeature.RegisterRoutes(r, notificationsHandler, requireUser)
	filesfeature.RegisterRoutes(r, filesHandler, requireUser)
	billingfeature.RegisterRoutes(r, billingHandler, requireUser)
	return r
}
