package controller

import (
	"errors"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type canvasHandoffExchangeRequest struct {
	Ticket string `json:"ticket" binding:"required"`
}

func CreateCanvasHandoff(c *gin.Context) {
	canvasURL, err := system_setting.GetCanvasURL()
	if err != nil {
		writeCanvasHandoffError(c, http.StatusOK, "CANVAS_CONFIG_INVALID", err.Error())
		return
	}
	if _, err := system_setting.GetCanvasAPIBaseURL(); err != nil {
		writeCanvasHandoffError(c, http.StatusOK, "CANVAS_CONFIG_INVALID", err.Error())
		return
	}
	group := system_setting.GetCanvasTokenGroup()
	userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
	if !service.GroupInUserUsableGroups(userGroup, group) {
		writeCanvasHandoffError(c, http.StatusOK, "CANVAS_GROUP_UNAVAILABLE", "the configured canvas token group is unavailable for this user")
		return
	}

	userID := common.GetContextKeyInt(c, constant.ContextKeyUserId)
	token, err := model.GetUsableUserTokenByGroup(userID, group, common.GetTimestamp())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeCanvasHandoffError(c, http.StatusOK, "CANVAS_TOKEN_NOT_FOUND", "no usable token was found for the configured canvas group")
		return
	}
	if err != nil {
		common.ApiError(c, err)
		return
	}

	ticket, expiresAt, err := service.CreateCanvasHandoffTicket(userID, token.Id, system_setting.GetCanvasHandoffTTL())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"ticket":     ticket,
		"canvas_url": canvasURL,
		"expires_at": expiresAt,
	})
}

func ExchangeCanvasHandoff(c *gin.Context) {
	request := canvasHandoffExchangeRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		writeCanvasHandoffError(c, http.StatusBadRequest, "CANVAS_TICKET_REQUIRED", "a canvas handoff ticket is required")
		return
	}

	payload, err := service.ConsumeCanvasHandoffTicket(request.Ticket)
	if errors.Is(err, service.ErrCanvasHandoffTicketInvalid) {
		writeCanvasHandoffError(c, http.StatusUnauthorized, "CANVAS_TICKET_INVALID", "the canvas handoff ticket is invalid or expired")
		return
	}
	if err != nil {
		writeCanvasHandoffError(c, http.StatusInternalServerError, "CANVAS_TICKET_EXCHANGE_FAILED", "failed to exchange the canvas handoff ticket")
		return
	}

	user, err := model.GetUserById(payload.UserID, false)
	if err != nil || user.Status != common.UserStatusEnabled {
		writeCanvasHandoffError(c, http.StatusUnauthorized, "CANVAS_USER_INVALID", "the user is no longer available")
		return
	}
	group := system_setting.GetCanvasTokenGroup()
	if !service.GroupInUserUsableGroups(user.Group, group) {
		writeCanvasHandoffError(c, http.StatusForbidden, "CANVAS_GROUP_UNAVAILABLE", "the configured canvas token group is unavailable for this user")
		return
	}
	token, err := model.GetUsableUserTokenByIdAndGroup(payload.TokenID, payload.UserID, group, common.GetTimestamp())
	if err != nil {
		writeCanvasHandoffError(c, http.StatusUnauthorized, "CANVAS_TOKEN_INVALID", "the canvas token is no longer usable")
		return
	}

	models := []string{}
	if token.ModelLimitsEnabled {
		models = token.GetModelLimits()
	}
	baseURL, err := system_setting.GetCanvasAPIBaseURL()
	if err != nil {
		writeCanvasHandoffError(c, http.StatusServiceUnavailable, "CANVAS_CONFIG_INVALID", err.Error())
		return
	}
	common.ApiSuccess(c, gin.H{
		"base_url": baseURL,
		"api_key":  "sk-" + token.GetFullKey(),
		"group":    group,
		"models":   models,
	})
}

func writeCanvasHandoffError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"code":    code,
		"message": message,
	})
}
