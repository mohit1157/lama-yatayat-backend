# ══════════════════════════════════════════════════
# LaMa Yatayat - Multi-stage Dockerfile
# Builds any service via --build-arg SERVICE=<name>
# ══════════════════════════════════════════════════

# ─── Stage 1: Build ─────────────────────────────
FROM golang:1.22-alpine AS builder

ARG SERVICE=user-service

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Copy source
COPY . .

# Build the specific service
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/service \
    ./cmd/${SERVICE}/

# ─── Stage 2: Runtime ───────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./service"]
