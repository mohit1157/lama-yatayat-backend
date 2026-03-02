package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohit1157/lama-yatayat-backend/internal/pricing/models"
)

type PricingRepository struct {
	db *pgxpool.Pool
}

func NewPricingRepository(db *pgxpool.Pool) *PricingRepository {
	return &PricingRepository{db: db}
}

// Zones

func (r *PricingRepository) ListZones(ctx context.Context) ([]models.PricingZone, error) {
	query := `SELECT id, name, surcharge, center_lat, center_lng, radius_m, active FROM pricing_zones WHERE active = true`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []models.PricingZone
	for rows.Next() {
		var z models.PricingZone
		rows.Scan(&z.ID, &z.Name, &z.Surcharge, &z.CenterLat, &z.CenterLng, &z.RadiusM, &z.Active)
		zones = append(zones, z)
	}
	return zones, nil
}

func (r *PricingRepository) UpdateZone(ctx context.Context, id string, surcharge float64) error {
	tag, err := r.db.Exec(ctx, `UPDATE pricing_zones SET surcharge = $2 WHERE id = $1`, id, surcharge)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("zone not found")
	}
	return nil
}

// Promo codes

func (r *PricingRepository) GetPromoByCode(ctx context.Context, code string) (*models.PromoCode, error) {
	p := &models.PromoCode{}
	query := `SELECT id, code, type, value, max_uses, used_count, active, expires_at, created_at
		FROM promo_codes WHERE code = $1`
	err := r.db.QueryRow(ctx, query, code).Scan(
		&p.ID, &p.Code, &p.Type, &p.Value, &p.MaxUses, &p.UsedCount, &p.Active, &p.ExpiresAt, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("promo not found: %w", err)
	}
	return p, nil
}

func (r *PricingRepository) ListPromos(ctx context.Context) ([]models.PromoCode, error) {
	query := `SELECT id, code, type, value, max_uses, used_count, active, expires_at, created_at
		FROM promo_codes ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promos []models.PromoCode
	for rows.Next() {
		var p models.PromoCode
		rows.Scan(&p.ID, &p.Code, &p.Type, &p.Value, &p.MaxUses, &p.UsedCount, &p.Active, &p.ExpiresAt, &p.CreatedAt)
		promos = append(promos, p)
	}
	return promos, nil
}

func (r *PricingRepository) CreatePromo(ctx context.Context, p *models.PromoCode) error {
	query := `INSERT INTO promo_codes (id, code, type, value, max_uses, used_count, active, expires_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7)`
	_, err := r.db.Exec(ctx, query, p.ID, p.Code, p.Type, p.Value, p.MaxUses, p.Active, p.ExpiresAt)
	return err
}

func (r *PricingRepository) IncrementPromoUsage(ctx context.Context, code string) error {
	_, err := r.db.Exec(ctx, `UPDATE promo_codes SET used_count = used_count + 1 WHERE code = $1`, code)
	return err
}
