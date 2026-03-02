-- Migration: 002 - Create rides and batches

CREATE TABLE IF NOT EXISTS ride_batches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id       UUID REFERENCES users(id),
    status          VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'in_progress', 'completed', 'cancelled')),
    route_polyline  TEXT,
    max_passengers  INT DEFAULT 4,
    current_count   INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rides (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id      UUID REFERENCES users(id),
    batch_id      UUID REFERENCES ride_batches(id),
    status        VARCHAR(30) NOT NULL DEFAULT 'requested'
                  CHECK (status IN ('requested', 'matching', 'matched', 'driver_en_route',
                         'pickup_arrived', 'in_progress', 'completed', 'cancelled', 'disputed')),
    pickup_lat    DECIMAL(10,7) NOT NULL,
    pickup_lng    DECIMAL(10,7) NOT NULL,
    pickup_addr   TEXT,
    dropoff_lat   DECIMAL(10,7) NOT NULL,
    dropoff_lng   DECIMAL(10,7) NOT NULL,
    dropoff_addr  TEXT,
    fare_amount   DECIMAL(10,2),
    is_round_trip BOOLEAN DEFAULT FALSE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    matched_at    TIMESTAMPTZ,
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    cancelled_at  TIMESTAMPTZ
);

CREATE INDEX idx_rides_rider ON rides(rider_id);
CREATE INDEX idx_rides_batch ON rides(batch_id);
CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_created ON rides(created_at DESC);

CREATE TABLE IF NOT EXISTS ride_ratings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id      UUID REFERENCES rides(id),
    from_user_id UUID REFERENCES users(id),
    to_user_id   UUID REFERENCES users(id),
    score        INT NOT NULL CHECK (score BETWEEN 1 AND 5),
    comment      TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ratings_ride ON ride_ratings(ride_id);
