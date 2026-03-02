package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/payment/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/payment/service"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) ChargeRider(c *gin.Context) {
	var req models.ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	txn, err := h.svc.ChargeRider(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, txn)
}

func (h *PaymentHandler) Refund(c *gin.Context) {
	var req models.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	txn, err := h.svc.Refund(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, txn)
}

func (h *PaymentHandler) GetWallet(c *gin.Context) {
	userID := c.Param("userId")
	wallet, err := h.svc.GetWallet(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, wallet)
}

func (h *PaymentHandler) PayoutDriver(c *gin.Context) {
	var req models.PayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	txn, err := h.svc.PayoutDriver(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, txn)
}

func (h *PaymentHandler) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit > 100 {
		limit = 100
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	txns, total, err := h.svc.GetHistory(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": txns, "meta": gin.H{"total": total}})
}

func (h *PaymentHandler) AddPaymentMethod(c *gin.Context) {
	var req models.AddPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userID, _ := c.Get("user_id")
	pm, err := h.svc.AddPaymentMethod(c.Request.Context(), userID.(string), &req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Created(c, pm)
}

func (h *PaymentHandler) ListPaymentMethods(c *gin.Context) {
	userID, _ := c.Get("user_id")
	methods, err := h.svc.ListPaymentMethods(c.Request.Context(), userID.(string))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, methods)
}
