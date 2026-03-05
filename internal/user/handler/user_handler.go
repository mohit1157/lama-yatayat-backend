package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/service"
	"github.com/mohit1157/lama-yatayat-backend/pkg/response"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Created(c, res)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, res)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.Success(c, res)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := h.svc.GetUser(c.Request.Context(), userID.(string))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	user, err := h.svc.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}
	response.Success(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdateUser(c.Request.Context(), c.Param("id"), &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "updated"})
}

func (h *UserHandler) OnboardDriver(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req models.OnboardDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	dp, err := h.svc.OnboardDriver(c.Request.Context(), userID.(string), &req)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Created(c, dp)
}

func (h *UserHandler) GetDriverStatus(c *gin.Context) {
	dp, err := h.svc.GetDriverStatus(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.NotFound(c, "driver profile not found")
		return
	}
	response.Success(c, dp)
}

func (h *UserHandler) ApproveDriver(c *gin.Context) {
	if err := h.svc.ApproveDriver(c.Request.Context(), c.Param("id")); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "driver approved"})
}

func (h *UserHandler) SuspendDriver(c *gin.Context) {
	if err := h.svc.SuspendDriver(c.Request.Context(), c.Param("id")); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "driver suspended"})
}

func getPagination(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if limit > 100 { limit = 100 }
	if page < 1 { page = 1 }
	return limit, (page - 1) * limit
}

func (h *UserHandler) ListDrivers(c *gin.Context) {
	limit, offset := getPagination(c)
	drivers, total, err := h.svc.ListDrivers(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": drivers, "meta": gin.H{"total": total}})
}

func (h *UserHandler) ResetUserPassword(c *gin.Context) {
	var req models.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.ResetUserPassword(c.Request.Context(), c.Param("id"), req.NewPassword); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "password reset successfully"})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	limit, offset := getPagination(c)
	users, total, err := h.svc.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": users, "meta": gin.H{"total": total}})
}
