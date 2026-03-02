-- Migration: 003 - Create payment tables

CREATE TABLE IF NOT EXISTS transactions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id    UUID REFERENCES rides(id),
    user_id    UUID REFERENCES users(id),
    type       VARCHAR(20) NOT NULL CHECK (type IN ('charge', 'refund', 'payout', 'credit')),
    amount     DECIMAL(10,2) NOT NULL,
    status     VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    stripe_id  VARCHAR(100),
    metadata   JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_txn_ride ON transactions(ride_id);
CREATE INDEX idx_txn_user ON transactions(user_id);
CREATE INDEX idx_txn_status ON transactions(status);

CREATE TABLE IF NOT EXISTS wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE REFERENCES users(id),
    balance         DECIMAL(10,2) DEFAULT 0.00,
    pending_balance DECIMAL(10,2) DEFAULT 0.00,
    currency        VARCHAR(3) DEFAULT 'USD',
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payment_methods (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id),
    type       VARCHAR(20) NOT NULL,
    stripe_pm  VARCHAR(100),
    last_four  VARCHAR(4),
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
