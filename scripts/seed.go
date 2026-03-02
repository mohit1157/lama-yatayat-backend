package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/mohit1157/lama-yatayat-backend/pkg/config"
	"github.com/mohit1157/lama-yatayat-backend/pkg/database"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, _ := config.Load()
	db, err := database.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Fort Worth area coordinates for demo
	baseLat, baseLng := 32.7555, -97.3308

	// Seed 10 drivers
	for i := 1; i <= 10; i++ {
		uid := uuid.New().String()
		email := fmt.Sprintf("driver%d@demo.lamayatayat.com", i)
		name := fmt.Sprintf("Driver %d", i)
		status := "approved"
		if i > 7 { status = "pending" }
		if i > 9 { status = "suspended" }

		db.Exec(ctx, `INSERT INTO users (id, email, password_hash, name, role, status)
			VALUES ($1, $2, $3, $4, 'driver', 'active') ON CONFLICT (email) DO NOTHING`,
			uid, email, string(hash), name)
		
		db.Exec(ctx, `INSERT INTO driver_profiles (id, user_id, license_number, vehicle_make,
			vehicle_model, vehicle_year, vehicle_plate, vehicle_color, capacity, bg_check_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT DO NOTHING`,
			uuid.New().String(), uid, fmt.Sprintf("DL%06d", i),
			"Toyota", "Camry", 2022+rng.Intn(3), fmt.Sprintf("LY-%04d", i),
			[]string{"White", "Black", "Silver", "Blue", "Red"}[rng.Intn(5)],
			4, status)
	}

	// Seed 50 riders
	for i := 1; i <= 50; i++ {
		uid := uuid.New().String()
		email := fmt.Sprintf("rider%d@demo.lamayatayat.com", i)
		name := fmt.Sprintf("Rider %d", i)
		db.Exec(ctx, `INSERT INTO users (id, email, password_hash, name, role, status)
			VALUES ($1, $2, $3, $4, 'rider', 'active') ON CONFLICT (email) DO NOTHING`,
			uid, email, string(hash), name)
	}

	// Seed 200 historical rides (for analytics)
	riders, _ := db.Query(ctx, `SELECT id FROM users WHERE role = 'rider' LIMIT 50`)
	var riderIDs []string
	for riders.Next() {
		var id string
		riders.Scan(&id)
		riderIDs = append(riderIDs, id)
	}
	riders.Close()

	for i := 0; i < 200; i++ {
		if len(riderIDs) == 0 { break }
		riderID := riderIDs[rng.Intn(len(riderIDs))]
		daysAgo := rng.Intn(30)
		createdAt := time.Now().AddDate(0, 0, -daysAgo)
		fare := 10.0
		if rng.Float64() > 0.6 { fare = 20.0 }

		pLat := baseLat + (rng.Float64()-0.5)*0.1
		pLng := baseLng + (rng.Float64()-0.5)*0.1
		dLat := baseLat + (rng.Float64()-0.5)*0.1
		dLng := baseLng + (rng.Float64()-0.5)*0.1

		db.Exec(ctx, `INSERT INTO rides (rider_id, status, pickup_lat, pickup_lng, dropoff_lat, dropoff_lng,
			fare_amount, is_round_trip, created_at, completed_at) VALUES ($1, 'completed', $2, $3, $4, $5, $6, $7, $8, $9)`,
			riderID, pLat, pLng, dLat, dLng, fare, fare == 20.0, createdAt, createdAt.Add(30*time.Minute))
	}

	fmt.Println("✅ Seed complete: 1 admin + 10 drivers + 50 riders + 200 historical rides")
	fmt.Println("   Admin login: admin@lamayatayat.com / admin123456")
	fmt.Println("   Driver login: driver1@demo.lamayatayat.com / password123")
	fmt.Println("   Rider login: rider1@demo.lamayatayat.com / password123")
}
