package models

type FareEstimateRequest struct {
	PickupLat   float64 `json:"pickup_lat"`
	PickupLng   float64 `json:"pickup_lng"`
	DropoffLat  float64 `json:"dropoff_lat"`
	DropoffLng  float64 `json:"dropoff_lng"`
	IsRoundTrip bool    `json:"is_round_trip"`
	PromoCode   string  `json:"promo_code,omitempty"`
}

type FareEstimateResponse struct {
	BaseFare    float64 `json:"base_fare"`
	Surcharges  float64 `json:"surcharges"`
	Discount    float64 `json:"discount"`
	Total       float64 `json:"total"`
	IsRoundTrip bool    `json:"is_round_trip"`
	PromoApplied string `json:"promo_applied,omitempty"`
}

type PricingZone struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Surcharge float64 `json:"surcharge"`
	CenterLat float64 `json:"center_lat"`
	CenterLng float64 `json:"center_lng"`
	RadiusM   float64 `json:"radius_meters"`
}

type PromoCode struct {
	Code       string  `json:"code"`
	Type       string  `json:"type"` // percent, fixed
	Value      float64 `json:"value"`
	MaxUses    int     `json:"max_uses"`
	UsedCount  int     `json:"used_count"`
	Active     bool    `json:"active"`
}
