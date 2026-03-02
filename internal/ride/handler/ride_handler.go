package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/ride/models"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type RideHandler struct {
	// svc *service.RideService  // TODO: wire up
}

func NewRideHandler() *RideHandler {
	return &RideHandler{}
}

func (h *RideHandler) RequestRide(c *gin.Context) {
	var req models.RideRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// TODO: Call service layer to create ride, publish ride.requested event
	response.Created(c, gin.H{"message": "ride requested", "ride_id": "TODO"})
}

func (h *RideHandler) GetRide(c *gin.Context) {
	rideID := c.Param("id")
	// TODO: Fetch from service
	response.Success(c, gin.H{"ride_id": rideID, "status": "requested"})
}

func (h *RideHandler) CancelRide(c *gin.Context) {
	rideID := c.Param("id")
	// TODO: Cancel logic with grace period
	response.Success(c, gin.H{"message": "ride cancelled", "ride_id": rideID})
}

func (h *RideHandler) ConfirmPickup(c *gin.Context) {
	rideID := c.Param("id")
	response.Success(c, gin.H{"message": "pickup confirmed", "ride_id": rideID})
}

func (h *RideHandler) ConfirmDropoff(c *gin.Context) {
	rideID := c.Param("id")
	response.Success(c, gin.H{"message": "dropoff confirmed", "ride_id": rideID})
}

func (h *RideHandler) GetActiveRide(c *gin.Context) {
	response.Success(c, gin.H{"ride": nil})
}

func (h *RideHandler) GetRideHistory(c *gin.Context) {
	response.Success(c, gin.H{"rides": []interface{}{}, "total": 0})
}

func (h *RideHandler) RateRide(c *gin.Context) {
	response.Success(c, gin.H{"message": "rating submitted"})
}

func (h *RideHandler) GetFareEstimate(c *gin.Context) {
	response.Success(c, models.FareEstimate{
		BaseFare: 10.00, Total: 10.00, IsRoundTrip: false,
	})
}

// Admin endpoints
func (h *RideHandler) ListRidesAdmin(c *gin.Context) {
	response.Success(c, gin.H{"rides": []interface{}{}, "total": 0})
}

func (h *RideHandler) GetRideStats(c *gin.Context) {
	response.Success(c, gin.H{
		"rides_today": 0, "rides_week": 0, "revenue_today": 0,
		"active_rides": 0, "avg_passengers_per_batch": 0,
	})
}
