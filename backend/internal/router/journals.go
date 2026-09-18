package router

import (
	"github.com/blueship581/mindgarden/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterJournals(g *gin.RouterGroup, h *handler.JournalHandler, auth gin.HandlerFunc) {
	p := g.Group("/journals", auth)
	p.GET("", h.List)
	p.POST("", h.Create)
	p.PUT("/:id", h.Update)
	p.DELETE("/:id", h.Delete)
}
