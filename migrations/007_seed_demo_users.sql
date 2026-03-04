-- Migration: 007 - Seed demo rider and driver accounts

-- Demo rider (password: rider12345)
INSERT INTO users (id, email, phone, password_hash, name, role, status, avatar_url) VALUES
('00000000-0000-0000-0000-000000000002', 'rider@demo.com', '+9779800000001',
 '$2a$10$rzMsotwiiUfNC/KW.mfVAOlUdEUq2JIqFEKUUbWqgmjsf38T.Dmke',
 'Demo Rider', 'rider', 'active', '')
ON CONFLICT (email) DO NOTHING;

-- Demo driver (password: driver12345)
INSERT INTO users (id, email, phone, password_hash, name, role, status, avatar_url) VALUES
('00000000-0000-0000-0000-000000000003', 'driver@demo.com', '+9779800000002',
 '$2a$10$6j38pcZD8FhStf5eKw0TEOokydM9O0.bXVQQmonK/TMdqPIRWUlJW',
 'Demo Driver', 'driver', 'active', '')
ON CONFLICT (email) DO NOTHING;

-- Demo driver profile (approved, with vehicle info)
INSERT INTO driver_profiles (id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year, vehicle_plate, vehicle_color, capacity, bg_check_status, rating_avg, rating_count, verified_at) VALUES
('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000003',
 'DL-2026-DEMO', 'Toyota', 'Prius', 2024, 'BA 1 JA 2026', 'White', 4,
 'approved', 4.85, 42, NOW())
ON CONFLICT (user_id) DO NOTHING;
