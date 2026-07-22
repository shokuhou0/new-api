package middleware

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

func CanvasHandoffCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin, err := system_setting.GetCanvasOrigin()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"code":    "CANVAS_CONFIG_INVALID",
				"message": err.Error(),
			})
			return
		}
		requestOrigin, err := common.NormalizeOrigin(c.GetHeader("Origin"))
		if err != nil || requestOrigin != allowedOrigin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"code":    "CANVAS_ORIGIN_FORBIDDEN",
				"message": "canvas handoff origin is not allowed",
			})
			return
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Max-Age", "600")
		c.Header("Vary", "Origin")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
