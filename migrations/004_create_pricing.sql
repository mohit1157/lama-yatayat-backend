-- Migration: 004 - Create pricing tables

CREATE TABLE IF NOT EXISTS pricing_zones (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name      VARCHAR(100) NOT NULL,
    surcharge DECIMAL(10,2) DEFAULT 0.00,
    center_lat DECIMAL(10,7),
    center_lng DECIMAL(10,7),
    radius_m  DECIMAL(10,2) DEFAULT 5000,
    active    BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS promo_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code       VARCHAR(50) UNIQUE NOT NULL,
    type       VARCHAR(20) NOT NULL CHECK (type IN ('percent', 'fixed')),
    value      DECIMAL(10,2) NOT NULL,
    max_uses   INT DEFAULT 100,
    used_count INT DEFAULT 0,
    active     BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id),
    type       VARCHAR(50) NOT NULL,
    title      VARCHAR(200),
    body       TEXT,
    read       BOOLEAN DEFAULT FALSE,
    metadata   JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notif_user ON notifications(user_id, read);
