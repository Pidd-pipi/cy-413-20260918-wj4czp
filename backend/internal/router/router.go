package router

import (
	"github.com/blueship581/mindgarden/backend/internal/config"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/handler"
	"github.com/blueship581/mindgarden/backend/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

type Handlers struct {
	User       *handler.UserHandler
	Mood       *handler.MoodHandler
	Assessment *handler.AssessmentHandler
	Journal    *handler.JournalHandler
}

func New(cfg config.Config, h Handlers, l *slog.Logger) *gin.Engine {
	r := gin.New()
	r.Use(cors.New(cors.Config{AllowOrigins: []string{cfg.CORSOrigin}, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Authorization", "Content-Type"}}))
	r.Use(middleware.ErrorHandler(l), middleware.AuditLogger(l))
	r.GET("/healthz", func(c *gin.Context) {
		l.Info(constants.LogHealthChecked)
		c.JSON(http.StatusOK, dto.Response{Code: 0, Message: "ok", Data: gin.H{"status": "healthy"}})
	})
	auth := middleware.Auth(cfg.JWTSecret, cfg.JWTIssuer, l)
	for _, prefix := range []string{"/v1", "/api/v1"} {
		g := r.Group(prefix)
		RegisterUsers(g, h.User, auth)
		RegisterMoods(g, h.Mood, auth)
		RegisterAssessments(g, h.Assessment, auth, l)
		RegisterJournals(g, h.Journal, auth)
	}
	return r
}
