package router

import (
	"github.com/blueship581/mindgarden/backend/internal/handler"
	"github.com/blueship581/mindgarden/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterUsers(g *gin.RouterGroup, h *handler.UserHandler, auth gin.HandlerFunc) {
	g.POST("/auth/register", h.Register)
	g.POST("/auth/login", h.Login)
	p := g.Group("/users", auth)
	p.GET("/me", h.Me)
	p.PUT("/me", h.Update)
	p.GET("/reports", h.Report)
	_ = middleware.UserID
}
