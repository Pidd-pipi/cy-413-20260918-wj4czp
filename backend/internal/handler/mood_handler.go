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

type MoodHandler struct {
	s      *service.MoodService
	logger *slog.Logger
}

func NewMoodHandler(s *service.MoodService, l *slog.Logger) *MoodHandler { return &MoodHandler{s, l} }
func (h *MoodHandler) List(c *gin.Context) {
	v, e := h.s.List(middleware.UserID(c), c.Query("date"))
	if e != nil {
		c.Error(e)
		return
	}
	h.logger.Info(constants.LogMoodListed)
	ok(c, v)
}
func (h *MoodHandler) Create(c *gin.Context) {
	var r dto.MoodRequest
	if !bind(c, &r) {
		return
	}
	v, e := h.s.Create(middleware.UserID(c), r)
	if e != nil {
		c.Error(e)
		return
	}
	created(c, v)
}
func (h *MoodHandler) Update(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.Error(util.NewAppError(constants.CodeValidation, "Mood[id] update failed: invalid id", e))
		return
	}
	var r dto.MoodRequest
	if !bind(c, &r) {
		return
	}
	v, e := h.s.Update(middleware.UserID(c), uint(id), r)
	if e != nil {
		c.Error(e)
		return
	}
	ok(c, v)
}
func (h *MoodHandler) Delete(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.Error(e)
		return
	}
	if e = h.s.Delete(middleware.UserID(c), uint(id)); e != nil {
		c.Error(e)
		return
	}
	ok(c, gin.H{"deleted": id})
}
