package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mohit1157/lama-yatayat-backend/internal/connection"
	geoHandler "github.com/mohit1157/lama-yatayat-backend/internal/geo/handler"
	matchingHandler "github.com/mohit1157/lama-yatayat-backend/internal/matching/handler"
	matchingService "github.com/mohit1157/lama-yatayat-backend/internal/matching/service"
	notifHandler "github.com/mohit1157/lama-yatayat-backend/internal/notification/handler"
	notifRepo "github.com/mohit1157/lama-yatayat-backend/internal/notification/repository"
	notifService "github.com/mohit1157/lama-yatayat-backend/internal/notification/service"
	payHandler "github.com/mohit1157/lama-yatayat-backend/internal/payment/handler"
	payModels "github.com/mohit1157/lama-yatayat-backend/internal/payment/models"
	payRepo "github.com/mohit1157/lama-yatayat-backend/internal/payment/repository"
	payService "github.com/mohit1157/lama-yatayat-backend/internal/payment/service"
	pricingHandler "github.com/mohit1157/lama-yatayat-backend/internal/pricing/handler"
	pricingRepo "github.com/mohit1157/lama-yatayat-backend/internal/pricing/repository"
	pricingService "github.com/mohit1157/lama-yatayat-backend/internal/pricing/service"
	rideHandler "github.com/mohit1157/lama-yatayat-backend/internal/ride/handler"
	rideRepo "github.com/mohit1157/lama-yatayat-backend/internal/ride/repository"
	rideService "github.com/mohit1157/lama-yatayat-backend/internal/ride/service"
	userHandler "github.com/mohit1157/lama-yatayat-backend/internal/user/handler"
	userRepo "github.com/mohit1157/lama-yatayat-backend/internal/user/repository"
	userService "github.com/mohit1157/lama-yatayat-backend/internal/user/service"

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

	zapLog := logger.New("gateway", cfg.App.Debug)
	defer zapLog.Sync()

	// ─── Database ────────────────────────────────────
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// ─── Redis ───────────────────────────────────────
	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	// ─── Event Bus ───────────────────────────────────
	bus := events.NewChannelBus(1000)
	go bus.Start(context.Background())
	defer bus.Stop()

	// ─── JWT ─────────────────────────────────────────
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	// ─── WebSocket Hub ───────────────────────────────
	hub := connection.NewHub()
	go hub.Run()

	// ─── Service Layers ──────────────────────────────

	// User
	uRepo := userRepo.NewUserRepository(db)
	uSvc := userService.NewUserService(uRepo, jwtMgr, bus)
	uH := userHandler.NewUserHandler(uSvc)

	// Ride
	rRepo := rideRepo.NewRideRepository(db)
	rSvc := rideService.NewRideService(rRepo, bus)
	rH := rideHandler.NewRideHandler(rSvc, cfg.Pricing.BaseFareRoundTrip, cfg.Pricing.BaseFareOneWay)

	// Payment
	pRepo := payRepo.NewPaymentRepository(db)
	pSvc := payService.NewPaymentService(pRepo, bus, cfg.Stripe.SecretKey, float64(cfg.Pricing.PlatformCommissionPct)/100.0)
	pH := payHandler.NewPaymentHandler(pSvc)

	// Pricing
	prRepo := pricingRepo.NewPricingRepository(db)
	prSvc := pricingService.NewPricingService(prRepo, cfg.Pricing.BaseFareRoundTrip, cfg.Pricing.BaseFareOneWay)
	prH := pricingHandler.NewPricingHandler(prSvc)

	// Notification
	nRepo := notifRepo.NewNotificationRepository(db)
	nSvc := notifService.NewNotificationService(nRepo)
	nH := notifHandler.NewNotificationHandler(nSvc)

	// Geolocation
	gH := geoHandler.NewGeoHandler(rdb)

	// Matching
	mSvc := matchingService.NewMatchingService(rdb, bus, float64(cfg.Matching.CorridorMeters), cfg.Matching.MaxBatchSize)
	mSvc.SetupEventListeners()
	mH := matchingHandler.NewMatchingHandler(mSvc)

	// ─── Event Subscribers ───────────────────────────

	// On ride completed → charge rider + notify
	bus.Subscribe("ride.completed", func(ctx context.Context, event events.Event) error {
		var payload struct {
			RideID    string  `json:"ride_id"`
			RiderID   string  `json:"rider_id"`
			FareAmount float64 `json:"fare_amount"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}

		// Auto-charge rider
		go pSvc.ChargeRider(ctx, &payModels.ChargeRequest{
			RideID: payload.RideID,
			UserID: payload.RiderID,
			Amount: payload.FareAmount,
		})

		// Notify rider
		go nSvc.SendToUser(ctx, payload.RiderID, "Trip Complete!",
			"Your ride has been completed. Thanks for riding with LaMa Yatayat!",
			map[string]string{"ride_id": payload.RideID, "type": "ride_completed"})

		return nil
	})

	// On ride matched → notify rider via push + WebSocket
	bus.Subscribe("ride.matched", func(ctx context.Context, event events.Event) error {
		var payload struct {
			RideID   string `json:"ride_id"`
			RiderID  string `json:"rider_id"`
			DriverID string `json:"driver_id"`
		}
		json.Unmarshal(event.Payload, &payload)

		go nSvc.SendToUser(ctx, payload.RiderID, "Ride Matched!",
			"A driver has been assigned to your ride.",
			map[string]string{"ride_id": payload.RideID, "type": "ride_matched"})

		// Send real-time WebSocket notification to the rider
		hub.SendToUser(payload.RiderID, &connection.Message{
			Type:    "ride_matched",
			RideID:  payload.RideID,
			Payload: event.Payload,
		})

		return nil
	})

	// On batch.offer → send WebSocket notification to the driver
	bus.Subscribe("batch.offer", func(ctx context.Context, event events.Event) error {
		var payload struct {
			DriverID  string  `json:"driver_id"`
			RideID    string  `json:"ride_id"`
			RiderID   string  `json:"rider_id"`
			PickupLat float64 `json:"pickup_lat"`
			PickupLng float64 `json:"pickup_lng"`
			DropoffLat float64 `json:"dropoff_lat"`
			DropoffLng float64 `json:"dropoff_lng"`
			DistanceM float64 `json:"distance_m"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}

		hub.SendToUser(payload.DriverID, &connection.Message{
			Type:    "batch_request",
			RideID:  payload.RideID,
			Payload: event.Payload,
		})

		log.Printf("WS: sent batch_request to driver %s for ride %s", payload.DriverID, payload.RideID)
		return nil
	})

	// ─── Router ──────────────────────────────────────
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS(), middleware.RequestLogger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "lama-yatayat-gateway",
			"time":    time.Now().UTC(),
		})
	})

	// WebSocket
	r.GET("/ws", func(c *gin.Context) {
		hub.HandleWebSocket(c.Writer, c.Request)
	})

	api := r.Group("/api/v1")

	// ─── Auth (Public) ───────────────────────────────
	api.POST("/auth/register", uH.Register)
	api.POST("/auth/login", uH.Login)
	api.POST("/auth/refresh", uH.RefreshToken)

	// ─── Protected Routes ────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthRequired(jwtMgr))

	// Auth (protected)
	protected.GET("/auth/me", uH.GetMe)

	// User
	protected.GET("/users/:id", uH.GetUser)
	protected.PUT("/users/:id", uH.UpdateUser)
	protected.POST("/drivers/onboard", uH.OnboardDriver)
	protected.GET("/drivers/:id/status", uH.GetDriverStatus)
	protected.GET("/drivers/:id/profile", uH.GetDriverStatus) // alias for mobile apps

	// Rides
	protected.POST("/rides/request", rH.RequestRide)
	protected.GET("/rides/:id", rH.GetRide)
	protected.PUT("/rides/:id/cancel", rH.CancelRide)
	protected.DELETE("/rides/:id/cancel", rH.CancelRide) // alias for mobile app
	protected.POST("/rides/:id/pickup-confirm", rH.ConfirmPickup)
	protected.POST("/rides/:id/dropoff-confirm", rH.ConfirmDropoff)
	protected.POST("/rides/:id/pickup", rH.ConfirmPickup)   // alias for mobile app
	protected.POST("/rides/:id/dropoff", rH.ConfirmDropoff) // alias for mobile app
	protected.GET("/rides/active", rH.GetActiveRide)
	protected.GET("/rides/history", rH.GetRideHistory)
	protected.POST("/rides/:id/rate", rH.RateRide)
	protected.GET("/rides/estimate", rH.GetFareEstimate)

	// Batch ride management (for driver mobile app)
	protected.POST("/rides/batches/:id/accept", mH.AcceptBatch)
	protected.POST("/rides/batches/:id/decline", mH.DeclineBatch)

	// Payments
	protected.POST("/payments/charge", pH.ChargeRider)
	protected.POST("/payments/refund", pH.Refund)
	protected.GET("/payments/wallet/:userId", pH.GetWallet)
	protected.POST("/payments/payout", pH.PayoutDriver)
	protected.GET("/payments/history", pH.GetHistory)
	protected.POST("/payments/methods", pH.AddPaymentMethod)
	protected.GET("/payments/methods", pH.ListPaymentMethods)
	protected.GET("/payments/earnings/summary", pH.GetEarningsSummary)
	protected.GET("/payments/earnings/recent", pH.GetRecentEarnings)

	// Pricing
	protected.POST("/pricing/estimate", prH.EstimateFare)
	protected.GET("/pricing/zones", prH.GetZones)
	protected.POST("/pricing/promo/validate", prH.ValidatePromo)

	// Notifications
	protected.POST("/notifications/token", nH.RegisterToken)
	protected.GET("/notifications", nH.ListNotifications)
	protected.PUT("/notifications/:id/read", nH.MarkRead)

	// Geolocation
	protected.PUT("/geo/drivers/:id/location", gH.UpdateLocation)
	protected.PUT("/geo/drivers/:id/status", gH.UpdateDriverStatus)
	protected.GET("/geo/drivers/nearby", gH.GetNearbyDrivers)
	protected.POST("/geo/route", gH.GetRoute)
	protected.GET("/geo/eta", gH.GetETA)

	// Matching (internal but accessible for demo)
	protected.POST("/matching/find-riders", mH.FindRiders)
	protected.GET("/matching/batch/:id", mH.GetBatch)
	protected.POST("/matching/optimize", mH.OptimizeSequence)

	// ─── Admin Routes ────────────────────────────────
	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(jwtMgr), middleware.RoleRequired("admin"))

	admin.GET("/drivers", uH.ListDrivers)
	admin.PUT("/drivers/:id/approve", uH.ApproveDriver)
	admin.PUT("/drivers/:id/suspend", uH.SuspendDriver)
	admin.GET("/users", uH.ListUsers)
	admin.PUT("/users/:id/reset-password", uH.ResetUserPassword)
	admin.GET("/rides", rH.ListRidesAdmin)
	admin.GET("/stats", rH.GetRideStats)
	admin.PUT("/pricing/zones/:id", prH.UpdateZone)
	admin.GET("/pricing/promos", prH.ListPromos)
	admin.POST("/pricing/promos", prH.CreatePromo)
	admin.POST("/notifications/send", nH.SendPush)

	// ─── Server ──────────────────────────────────────
	port := cfg.App.ServicePort
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("🚀 LaMa Yatayat Gateway starting on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
