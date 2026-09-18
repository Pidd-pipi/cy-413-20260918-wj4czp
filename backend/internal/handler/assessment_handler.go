package handler

import (
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/middleware"
	"github.com/blueship581/mindgarden/backend/internal/service"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"github.com/gin-gonic/gin"
	"log/slog"
	"strconv"
)

type AssessmentHandler struct {
	s      *service.AssessmentService
	logger *slog.Logger
}

func NewAssessmentHandler(s *service.AssessmentService, l *slog.Logger) *AssessmentHandler {
	return &AssessmentHandler{s, l}
}
func (h *AssessmentHandler) List(c *gin.Context) {
	v, e := h.s.List()
	if e != nil {
		c.Error(e)
		return
	}
	h.logger.Info(constants.LogAssessmentListed)
	ok(c, v)
}
func (h *AssessmentHandler) Create(c *gin.Context) {
	var r dto.AssessmentRequest
	if !bind(c, &r) {
		return
	}
	v, e := h.s.Create(r)
	if e != nil {
		c.Error(e)
		return
	}
	created(c, v)
}
func (h *AssessmentHandler) Take(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.Error(util.NewAppError(constants.CodeValidation, "Assessment[id] submit failed: invalid id", e))
		return
	}
	var r dto.TakeAssessmentRequest
	if !bind(c, &r) {
		return
	}
	v, e := h.s.Take(middleware.UserID(c), uint(id), r)
	if e != nil {
		c.Error(e)
		return
	}
	ok(c, v)
}
