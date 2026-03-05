package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/mohit1157/lama-yatayat-backend/internal/geo/models"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type GeoHandler struct {
	rdb *redis.Client
}

func NewGeoHandler(rdb *redis.Client) *GeoHandler {
	return &GeoHandler{rdb: rdb}
}

const driverLocationKey = "drivers:locations"

func (h *GeoHandler) UpdateLocation(c *gin.Context) {
	driverID := c.Param("id")
	var req models.DriverLocation
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Store in Redis geospatial index
	ctx := context.Background()
	err := h.rdb.GeoAdd(ctx, driverLocationKey, &redis.GeoLocation{
		Name:      driverID,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	}).Err()
	if err != nil {
		response.InternalError(c, fmt.Sprintf("failed to update location: %v", err))
		return
	}

	// Also store metadata (heading, speed, timestamp) in a hash
	h.rdb.HSet(ctx, "driver:meta:"+driverID, map[string]interface{}{
		"lat": req.Latitude, "lng": req.Longitude,
		"heading": req.Heading, "speed": req.Speed, "ts": req.Timestamp,
	})

	// Publish for real-time subscribers
	h.rdb.Publish(ctx, "driver.location."+driverID, fmt.Sprintf(
		`{"driver_id":"%s","lat":%f,"lng":%f,"heading":%f,"speed":%f}`,
		driverID, req.Latitude, req.Longitude, req.Heading, req.Speed))

	response.Success(c, gin.H{"message": "location updated"})
}

func (h *GeoHandler) UpdateDriverStatus(c *gin.Context) {
	driverID := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"` // "online" or "offline"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := context.Background()
	h.rdb.HSet(ctx, "driver:meta:"+driverID, "status", req.Status)

	if req.Status == "offline" {
		// Remove from geospatial index when going offline
		h.rdb.ZRem(ctx, driverLocationKey, driverID)
	}

	response.Success(c, gin.H{"message": "status updated", "status": req.Status})
}

func (h *GeoHandler) GetNearbyDrivers(c *gin.Context) {
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	radiusStr := c.DefaultQuery("radius", "5000") // meters
	radius, _ := strconv.ParseFloat(radiusStr, 64)

	ctx := context.Background()
	results, err := h.rdb.GeoSearchLocation(ctx, driverLocationKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude: lng, Latitude: lat,
			Radius: radius, RadiusUnit: "m",
			Sort: "ASC", Count: 20,
		},
		WithCoord: true, WithDist: true,
	}).Result()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	drivers := make([]models.NearbyDriver, 0, len(results))
	for _, r := range results {
		drivers = append(drivers, models.NearbyDriver{
			DriverID: r.Name, Latitude: r.Latitude, Longitude: r.Longitude, Distance: r.Dist,
		})
	}

	response.Success(c, drivers)
}

func (h *GeoHandler) GetRoute(c *gin.Context) {
	var req models.RouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// TODO: Call Google Maps Directions API
	response.Success(c, models.RouteResponse{
		Polyline: "TODO_encoded_polyline", DistanceM: 12500, DurationSec: 1200,
	})
}

func (h *GeoHandler) GetETA(c *gin.Context) {
	// TODO: Call Google Maps Distance Matrix API
	response.Success(c, models.ETAResponse{DurationSec: 480, DistanceM: 3200})
}
