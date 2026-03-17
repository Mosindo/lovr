package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r gin.IRouter, handler *Handler, requireUser gin.HandlerFunc) {
	r.POST("/auth/register", handler.Register)
	r.POST("/auth/login", handler.Login)
	r.GET("/me", requireUser, handler.Me)
}
