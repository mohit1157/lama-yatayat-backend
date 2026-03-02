package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/handler"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/repository"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/service"
	"github.com/mohit1157/lama-yatayat-backend/pkg/auth"
	"github.com/mohit1157/lama-yatayat-backend/pkg/config"
	"github.com/mohit1157/lama-yatayat-backend/pkg/database"
	"github.com/mohit1157/lama-yatayat-backend/pkg/events"
	"github.com/mohit1157/lama-yatayat-backend/pkg/logger"
	"github.com/mohit1157/lama-yatayat-backend/pkg/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLog := logger.New("user-service", cfg.App.Debug)
	defer zapLog.Sync()

	// Database
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// Event bus
	bus := events.NewChannelBus(1000)
	go bus.Start(context.Background())
	defer bus.Stop()

	// JWT
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	// Layers
	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo, jwtMgr, bus)
	h := handler.NewUserHandler(svc)

	// Router
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestLogger())

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user-service"})
	})

	// Public routes
	api := r.Group("/api/v1")
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/refresh", h.RefreshToken)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(jwtMgr))
	protected.GET("/users/:id", h.GetUser)
	protected.PUT("/users/:id", h.UpdateUser)
	protected.POST("/drivers/onboard", h.OnboardDriver)
	protected.GET("/drivers/:id/status", h.GetDriverStatus)

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(jwtMgr), middleware.RoleRequired("admin"))
	admin.GET("/drivers", h.ListDrivers)
	admin.PUT("/drivers/:id/approve", h.ApproveDriver)
	admin.PUT("/drivers/:id/suspend", h.SuspendDriver)
	admin.GET("/users", h.ListUsers)

	// Server
	port := cfg.App.ServicePort
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("🚀 User Service starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
