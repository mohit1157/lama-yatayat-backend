package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/matching/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/matching/service"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type MatchingHandler struct {
	svc *service.MatchingService
}

func NewMatchingHandler(svc *service.MatchingService) *MatchingHandler {
	return &MatchingHandler{svc: svc}
}

func (h *MatchingHandler) FindRiders(c *gin.Context) {
	var req models.MatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	batch, err := h.svc.FindRiders(c.Request.Context(), &req)
	if err != nil {
		response.Success(c, gin.H{"batch": nil, "message": err.Error()})
		return
	}
	response.Success(c, batch)
}

func (h *MatchingHandler) GetBatch(c *gin.Context) {
	response.Success(c, gin.H{"batch_id": c.Param("id"), "message": "batch details"})
}

func (h *MatchingHandler) OptimizeSequence(c *gin.Context) {
	response.Success(c, gin.H{"message": "sequence optimized"})
}
