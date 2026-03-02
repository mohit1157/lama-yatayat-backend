-- Migration: 001 - Create users and driver profiles
-- LaMa Yatayat Backend

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    phone         VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(100) NOT NULL,
    role          VARCHAR(20) NOT NULL CHECK (role IN ('rider', 'driver', 'admin')),
    status        VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deactivated')),
    avatar_url    TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

CREATE TABLE IF NOT EXISTS driver_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    license_number  VARCHAR(50),
    license_doc_url TEXT,
    vehicle_make    VARCHAR(50),
    vehicle_model   VARCHAR(50),
    vehicle_year    INT,
    vehicle_plate   VARCHAR(20),
    vehicle_color   VARCHAR(30),
    capacity        INT DEFAULT 4 CHECK (capacity BETWEEN 1 AND 6),
    bg_check_status VARCHAR(20) DEFAULT 'pending' CHECK (bg_check_status IN ('pending', 'approved', 'rejected', 'suspended')),
    rating_avg      DECIMAL(3,2) DEFAULT 5.00,
    rating_count    INT DEFAULT 0,
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_driver_profiles_status ON driver_profiles(bg_check_status);
