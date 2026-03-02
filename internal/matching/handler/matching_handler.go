package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type MatchingHandler struct {
	// engine *service.MatchingService // TODO: wire up
}

func NewMatchingHandler() *MatchingHandler {
	return &MatchingHandler{}
}

func (h *MatchingHandler) FindRiders(c *gin.Context) {
	// TODO: Called by Ride Service when driver goes online
	// 1. Get driver route polyline
	// 2. Decompose into geohash segments
	// 3. Query pending rides in corridor
	// 4. Filter by detour tolerance
	// 5. Run TSP solver
	// 6. Return batch
	response.Success(c, gin.H{"batch": nil, "message": "matching engine ready"})
}

func (h *MatchingHandler) GetBatch(c *gin.Context) {
	batchID := c.Param("id")
	response.Success(c, gin.H{"batch_id": batchID})
}

func (h *MatchingHandler) OptimizeSequence(c *gin.Context) {
	response.Success(c, gin.H{"message": "sequence optimized"})
}
