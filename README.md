# 🚐 LaMa Yatayat — Backend

> Multi-Passenger Route-Optimized Rideshare Platform

**Go 1.22+** | **PostgreSQL** | **Redis** | **Docker Compose**

---

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Rider App  │────▶│  API Gateway │────▶│  User Service    │ :8001
│  Driver App │     │  (Kong/Nginx)│     │  Ride Service    │ :8002
│  Admin Web  │     └──────────────┘     │  Route Matching  │ :8003
└─────────────┘                          │  Geolocation     │ :8004
                                         │  Payment         │ :8005
                                         │  Pricing         │ :8006
                                         │  Notification    │ :8007
                                         │  Connection (WS) │ :8008
                                         └─────────────────┘
                                                │
                                         ┌──────┴──────┐
                                         │ PostgreSQL  │
                                         │ Redis       │
                                         └─────────────┘
```

## Quick Start

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- Make

### 1. Clone & Configure

```bash
git clone https://github.com/mohit1157/lama-yatayat-backend.git
cd lama-yatayat-backend
cp .env.example .env
# Edit .env: add your Stripe test key, Google Maps API key
```

### 2. Start Infrastructure

```bash
make infra-up    # Starts PostgreSQL + Redis in Docker
```

### 3. Run Migrations

```bash
make migrate     # Creates all tables
```

### 4. Seed Demo Data

```bash
make seed        # Creates admin, 10 drivers, 50 riders, 200 rides
```

### 5. Run All Services

```bash
make run         # Starts all 8 services concurrently
```

### 6. Verify

```bash
# Health check
curl http://localhost:8001/health

# Register a rider
curl -X POST http://localhost:8001/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User","role":"rider"}'

# Login
curl -X POST http://localhost:8001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

## Service Ports

| Service | Port | Description |
|---------|------|-------------|
| User Service | 8001 | Auth, profiles, driver onboarding |
| Ride Service | 8002 | Ride lifecycle, batches |
| Route Matching | 8003 | Corridor matching, TSP optimization |
| Geolocation | 8004 | GPS ingestion, nearby drivers |
| Payment | 8005 | Stripe charges, payouts, wallets |
| Pricing | 8006 | Fare estimation, zones, promos |
| Notification | 8007 | Push notifications, in-app messages |
| Connection | 8008 | WebSocket for live tracking |

## Project Structure

```
├── cmd/                    # Service entry points (main.go per service)
├── internal/               # Private packages per service
│   ├── user/               # handler/ service/ repository/ models/
│   ├── ride/
│   ├── matching/engine/    # Corridor matching + TSP solver
│   ├── geo/
│   ├── payment/
│   ├── pricing/
│   ├── notification/
│   └── connection/         # WebSocket hub
├── pkg/                    # Shared packages
│   ├── auth/               # JWT creation + validation
│   ├── middleware/          # CORS, auth, logging
│   ├── database/           # PostgreSQL + Redis helpers
│   ├── events/             # Event bus (channels for demo, Kafka for prod)
│   ├── config/             # Viper configuration
│   ├── response/           # Standardized API responses
│   ├── geohash/            # Geohash encoding
│   └── logger/             # Zap structured logging
├── migrations/             # SQL migration files (run in order)
├── scripts/                # Seed data, driver simulation
├── docker-compose.yml      # Full local development stack
├── Dockerfile              # Multi-stage build (any service via --build-arg)
└── Makefile                # All dev commands
```

## Key API Endpoints

### Auth
- `POST /api/v1/auth/register` — Register rider or driver
- `POST /api/v1/auth/login` — Login, returns JWT
- `POST /api/v1/auth/refresh` — Refresh access token

### Rides
- `POST /api/v1/rides/request` — Request a ride
- `GET /api/v1/rides/:id` — Get ride details
- `PUT /api/v1/rides/:id/cancel` — Cancel ride
- `POST /api/v1/rides/:id/pickup-confirm` — Confirm pickup
- `POST /api/v1/rides/:id/dropoff-confirm` — Confirm dropoff

### Geolocation
- `PUT /api/v1/geo/drivers/:id/location` — Update driver GPS
- `GET /api/v1/geo/drivers/nearby?lat=&lng=&radius=` — Find nearby drivers

### Admin
- `GET /api/v1/admin/drivers` — List all drivers
- `PUT /api/v1/admin/drivers/:id/approve` — Approve driver
- `GET /api/v1/admin/users` — List all users

## Demo Credentials

| Role | Email | Password |
|------|-------|----------|
| Admin | admin@lamayatayat.com | admin123456 |
| Driver | driver1@demo.lamayatayat.com | password123 |
| Rider | rider1@demo.lamayatayat.com | password123 |

## Docker (Full Stack)

```bash
# Build and run everything in Docker
docker-compose up --build -d

# View logs
docker-compose logs -f

# Stop
docker-compose down -v
```

## Testing

```bash
make test             # Run all tests
make test-coverage    # Generate coverage report
```

## Deploy to Render (Free Demo)

1. Push this repo to GitHub
2. Create a new Web Service on [Render](https://render.com)
3. Connect your GitHub repo
4. Set build command: `go build -o bin/service ./cmd/user-service`
5. Set start command: `./bin/service`
6. Add environment variables from `.env.example`
7. Add a PostgreSQL database (free tier: 256MB)
8. Add a Redis instance (free tier via Upstash)

## Companion Repos

- **Mobile App**: [lama-yatayat-mobile](https://github.com/mohit1157/lama-yatayat-mobile) (Expo / React Native)
- **Admin Portal**: [lama-yatayat-admin](https://github.com/mohit1157/lama-yatayat-admin) (Next.js)

## License

Proprietary — LaMa Yatayat © 2026
