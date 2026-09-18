package router

import (
	"github.com/blueship581/mindgarden/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterMoods(g *gin.RouterGroup, h *handler.MoodHandler, auth gin.HandlerFunc) {
	p := g.Group("/moods", auth)
	p.GET("", h.List)
	p.POST("", h.Create)
	p.PUT("/:id", h.Update)
	p.DELETE("/:id", h.Delete)
}
