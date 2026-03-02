package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/mohit1157/lama-yatayat-backend/internal/pricing/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/pricing/repository"
)

type PricingService struct {
	repo           *repository.PricingRepository
	baseFareRT     float64
	baseFareOneWay float64
}

func NewPricingService(repo *repository.PricingRepository, baseFareRT, baseFareOneWay float64) *PricingService {
	return &PricingService{repo: repo, baseFareRT: baseFareRT, baseFareOneWay: baseFareOneWay}
}

func (s *PricingService) EstimateFare(ctx context.Context, req *models.FareEstimateRequest) (*models.FareEstimateResponse, error) {
	base := s.baseFareOneWay
	if req.IsRoundTrip {
		base = s.baseFareRT
	}

	// Zone surcharge
	var surcharges float64
	zones, _ := s.repo.ListZones(ctx)
	for _, z := range zones {
		if isPointInZone(req.PickupLat, req.PickupLng, z) || isPointInZone(req.DropoffLat, req.DropoffLng, z) {
			if z.Surcharge > surcharges {
				surcharges = z.Surcharge // Use highest applicable surcharge
			}
		}
	}

	// Promo discount
	var discount float64
	var promoApplied string
	if req.PromoCode != "" {
		promo, err := s.repo.GetPromoByCode(ctx, req.PromoCode)
		if err == nil && promo.Active && promo.UsedCount < promo.MaxUses {
			if promo.ExpiresAt != nil && promo.ExpiresAt.Before(time.Now()) {
				// Expired — skip
			} else {
				switch promo.Type {
				case "fixed":
					discount = promo.Value
				case "percent":
					discount = (base + surcharges) * promo.Value / 100
				}
				promoApplied = promo.Code
			}
		}
	}

	total := base + surcharges - discount
	if total < 0 {
		total = 0
	}

	return &models.FareEstimateResponse{
		BaseFare:     base,
		Surcharges:   surcharges,
		Discount:     discount,
		Total:        math.Round(total*100) / 100,
		IsRoundTrip:  req.IsRoundTrip,
		PromoApplied: promoApplied,
	}, nil
}

func (s *PricingService) GetZones(ctx context.Context) ([]models.PricingZone, error) {
	return s.repo.ListZones(ctx)
}

func (s *PricingService) UpdateZone(ctx context.Context, id string, surcharge float64) error {
	return s.repo.UpdateZone(ctx, id, surcharge)
}

func (s *PricingService) ValidatePromo(ctx context.Context, code string) (*models.PromoCode, error) {
	promo, err := s.repo.GetPromoByCode(ctx, code)
	if err != nil {
		return nil, errors.New("promo code not found")
	}
	if !promo.Active {
		return nil, errors.New("promo code is inactive")
	}
	if promo.UsedCount >= promo.MaxUses {
		return nil, errors.New("promo code has reached maximum uses")
	}
	if promo.ExpiresAt != nil && promo.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("promo code has expired")
	}
	return promo, nil
}

func (s *PricingService) ListPromos(ctx context.Context) ([]models.PromoCode, error) {
	return s.repo.ListPromos(ctx)
}

func (s *PricingService) CreatePromo(ctx context.Context, req *models.CreatePromoRequest) (*models.PromoCode, error) {
	p := &models.PromoCode{
		ID:       uuid.New().String(),
		Code:     req.Code,
		Type:     req.Type,
		Value:    req.Value,
		MaxUses:  req.MaxUses,
		Active:   true,
		ExpiresAt: req.ExpiresAt,
	}
	if err := s.repo.CreatePromo(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// isPointInZone checks if a lat/lng falls within a circular pricing zone
func isPointInZone(lat, lng float64, zone models.PricingZone) bool {
	const R = 6371000.0 // Earth radius in meters
	dLat := (lat - zone.CenterLat) * math.Pi / 180
	dLng := (lng - zone.CenterLng) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(zone.CenterLat*math.Pi/180)*math.Cos(lat*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c
	return distance <= zone.RadiusM
}
