package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/pricing/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/pricing/service"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type PricingHandler struct {
	svc *service.PricingService
}

func NewPricingHandler(svc *service.PricingService) *PricingHandler {
	return &PricingHandler{svc: svc}
}

func (h *PricingHandler) EstimateFare(c *gin.Context) {
	var req models.FareEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	estimate, err := h.svc.EstimateFare(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, estimate)
}

func (h *PricingHandler) GetZones(c *gin.Context) {
	zones, err := h.svc.GetZones(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, zones)
}

func (h *PricingHandler) UpdateZone(c *gin.Context) {
	var req models.UpdateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdateZone(c.Request.Context(), c.Param("id"), req.Surcharge); err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "zone updated"})
}

func (h *PricingHandler) ValidatePromo(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	promo, err := h.svc.ValidatePromo(c.Request.Context(), req.Code)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, promo)
}

func (h *PricingHandler) ListPromos(c *gin.Context) {
	promos, err := h.svc.ListPromos(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, promos)
}

func (h *PricingHandler) CreatePromo(c *gin.Context) {
	var req models.CreatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	promo, err := h.svc.CreatePromo(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, promo)
}
