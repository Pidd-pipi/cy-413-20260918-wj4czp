package router

import (
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/handler"
	"github.com/blueship581/mindgarden/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func RegisterAssessments(g *gin.RouterGroup, h *handler.AssessmentHandler, auth gin.HandlerFunc, logger *slog.Logger) {
	g.GET("/assessments", h.List)
	g.POST("/assessments", auth, middleware.RequireRole(constants.RoleAdmin, logger), h.Create)
	g.POST("/assessments/:id/take", auth, h.Take)
}
