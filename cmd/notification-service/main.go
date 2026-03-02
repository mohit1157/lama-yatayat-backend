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

	zapLog := logger.New("notification-service", cfg.App.Debug)
	defer zapLog.Sync()

	// Database (skip for geo/connection which are Redis-only)
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Printf("warning: database not available: %v", err)
	}
	if db != nil {
		defer db.Close()
	}

	// Redis
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Printf("warning: redis not available: %v", err)
	}
	if rdb != nil {
		defer rdb.Close()
	}

	// Event bus
	bus := events.NewChannelBus(1000)
	go bus.Start(context.Background())
	defer bus.Stop()

	_ = db   // Available for service layer
	_ = rdb  // Available for service layer
	_ = bus  // Available for event publishing/subscribing

	// Router
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestLogger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notification-service"})
	})

	// TODO: Register service-specific routes here
	// See internal/notification-service/ for handler, service, repository layers
	api := r.Group("/api/v1")
	_ = api

	// Server
	port := cfg.App.ServicePort
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("🚀 notification-service starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
