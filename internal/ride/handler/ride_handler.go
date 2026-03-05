package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/ride/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/ride/service"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type RideHandler struct {
	svc            *service.RideService
	baseFareRT     float64
	baseFareOneWay float64
}

func NewRideHandler(svc *service.RideService, baseFareRT, baseFareOneWay float64) *RideHandler {
	return &RideHandler{svc: svc, baseFareRT: baseFareRT, baseFareOneWay: baseFareOneWay}
}

func (h *RideHandler) RequestRide(c *gin.Context) {
	var req models.RideRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	riderID, _ := c.Get("user_id")
	fare := h.baseFareOneWay
	if req.IsRoundTrip {
		fare = h.baseFareRT
	}

	ride, err := h.svc.RequestRide(c.Request.Context(), riderID.(string), &req, fare)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, ride)
}

func (h *RideHandler) GetRide(c *gin.Context) {
	ride, err := h.svc.GetRide(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.NotFound(c, "ride not found")
		return
	}
	response.Success(c, ride)
}

func (h *RideHandler) CancelRide(c *gin.Context) {
	userID, _ := c.Get("user_id")
	if err := h.svc.CancelRide(c.Request.Context(), c.Param("id"), userID.(string)); err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "ride cancelled"})
}

func (h *RideHandler) ConfirmPickup(c *gin.Context) {
	if err := h.svc.ConfirmPickup(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "pickup confirmed"})
}

func (h *RideHandler) ConfirmDropoff(c *gin.Context) {
	if err := h.svc.ConfirmDropoff(c.Request.Context(), c.Param("id")); err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "dropoff confirmed"})
}

func (h *RideHandler) GetActiveRide(c *gin.Context) {
	userID, _ := c.Get("user_id")
	ride, err := h.svc.GetActiveRide(c.Request.Context(), userID.(string))
	if err != nil || ride == nil {
		response.NotFound(c, "no active ride")
		return
	}
	response.Success(c, ride)
}

func (h *RideHandler) GetRideHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	limit, offset := getPagination(c)
	rides, total, err := h.svc.GetHistory(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rides, "meta": gin.H{"total": total}})
}

func (h *RideHandler) RateRide(c *gin.Context) {
	var input models.RateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.svc.RateRide(c.Request.Context(), c.Param("id"), userID.(string), &input); err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "rating submitted"})
}

func (h *RideHandler) GetFareEstimate(c *gin.Context) {
	isRoundTrip := c.Query("round_trip") == "true"
	estimate := h.svc.GetFareEstimate(c.Request.Context(), isRoundTrip, h.baseFareRT, h.baseFareOneWay)
	response.Success(c, estimate)
}

// Admin endpoints

func (h *RideHandler) ListRidesAdmin(c *gin.Context) {
	limit, offset := getPagination(c)
	status := c.Query("status")
	rides, total, err := h.svc.ListAll(c.Request.Context(), status, limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": rides, "meta": gin.H{"total": total}})
}

func (h *RideHandler) GetRideStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func getPagination(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}
	return limit, (page - 1) * limit
}
