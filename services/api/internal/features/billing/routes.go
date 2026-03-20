package billing

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter, handler *Handler, requireUser gin.HandlerFunc) {
	r.POST("/billing/checkout", requireUser, handler.Checkout)
	r.POST("/billing/webhook", handler.Webhook)
}
