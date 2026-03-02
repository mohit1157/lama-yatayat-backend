-- Migration: 005 - Seed admin user and demo data

-- Admin user (password: admin123456)
INSERT INTO users (id, email, phone, password_hash, name, role, status) VALUES
('00000000-0000-0000-0000-000000000001', 'admin@lamayatayat.com', '+1234567890',
 '$2a$10$XQxBj4vGNXr1RZiHKqcjLuZpXvDKxNRiDmhLGkm6xBJFGxJhKZP2e',
 'Admin User', 'admin', 'active')
ON CONFLICT (email) DO NOTHING;

-- Sample pricing zones
INSERT INTO pricing_zones (name, surcharge, center_lat, center_lng, radius_m) VALUES
('Airport', 3.00, 32.8998, -97.0403, 3000),
('Stadium', 2.00, 32.7480, -97.0925, 2000),
('Downtown', 0.00, 32.7555, -97.3308, 5000)
ON CONFLICT DO NOTHING;

-- Sample promo codes
INSERT INTO promo_codes (code, type, value, max_uses) VALUES
('FIRST_RIDE', 'fixed', 5.00, 1000),
('DEMO2026', 'percent', 50.00, 500),
('LAUNCH', 'fixed', 10.00, 100)
ON CONFLICT (code) DO NOTHING;
