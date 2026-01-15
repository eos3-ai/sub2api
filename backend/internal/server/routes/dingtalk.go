package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"

	"github.com/gin-gonic/gin"
)

// RegisterDingtalkBotRoutes registers DingTalk bot webhooks (no auth).
func RegisterDingtalkBotRoutes(r *gin.Engine, h *handler.Handlers) {
	if r == nil || h == nil || h.DingtalkBot == nil {
		return
	}

	group := r.Group("/hooks/dingtalk")
	group.GET("/recharge", h.DingtalkBot.RechargeStatus)
	group.POST("/recharge", h.DingtalkBot.Recharge)
}
