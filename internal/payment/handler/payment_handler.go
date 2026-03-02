package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/payment/models"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type PaymentHandler struct {}

func NewPaymentHandler() *PaymentHandler { return &PaymentHandler{} }

func (h *PaymentHandler) ChargeRider(c *gin.Context) {
	var req models.ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// TODO: Create Stripe PaymentIntent
	response.Success(c, gin.H{"message": "charge initiated", "amount": req.Amount})
}

func (h *PaymentHandler) Refund(c *gin.Context) {
	response.Success(c, gin.H{"message": "refund processed"})
}

func (h *PaymentHandler) GetWallet(c *gin.Context) {
	userID := c.Param("userId")
	response.Success(c, models.Wallet{UserID: userID, Balance: 0, PendingBalance: 0})
}

func (h *PaymentHandler) PayoutDriver(c *gin.Context) {
	response.Success(c, gin.H{"message": "payout initiated"})
}

func (h *PaymentHandler) GetHistory(c *gin.Context) {
	response.Success(c, gin.H{"transactions": []interface{}{}, "total": 0})
}
