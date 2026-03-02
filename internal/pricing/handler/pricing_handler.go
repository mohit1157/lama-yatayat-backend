package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/pricing/models"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type PricingHandler struct {
	baseFareRoundTrip float64
	baseFareOneWay    float64
}

func NewPricingHandler(roundTrip, oneWay float64) *PricingHandler {
	return &PricingHandler{baseFareRoundTrip: roundTrip, baseFareOneWay: oneWay}
}

func (h *PricingHandler) EstimateFare(c *gin.Context) {
	var req models.FareEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	base := h.baseFareOneWay
	if req.IsRoundTrip {
		base = h.baseFareRoundTrip
	}
	// TODO: Check zone surcharges, apply promo codes
	response.Success(c, models.FareEstimateResponse{
		BaseFare: base, Surcharges: 0, Discount: 0, Total: base, IsRoundTrip: req.IsRoundTrip,
	})
}

func (h *PricingHandler) GetZones(c *gin.Context) {
	response.Success(c, gin.H{"zones": []interface{}{}})
}

func (h *PricingHandler) UpdateZone(c *gin.Context) {
	response.Success(c, gin.H{"message": "zone updated"})
}

func (h *PricingHandler) ValidatePromo(c *gin.Context) {
	response.Success(c, gin.H{"valid": false, "message": "promo code not found"})
}

func (h *PricingHandler) ListPromos(c *gin.Context) {
	response.Success(c, gin.H{"promos": []interface{}{}})
}

func (h *PricingHandler) CreatePromo(c *gin.Context) {
	response.Success(c, gin.H{"message": "promo created"})
}
