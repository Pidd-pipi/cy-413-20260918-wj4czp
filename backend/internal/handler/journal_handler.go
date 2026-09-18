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

type JournalHandler struct {
	s      *service.JournalService
	logger *slog.Logger
}

func NewJournalHandler(s *service.JournalService, l *slog.Logger) *JournalHandler {
	return &JournalHandler{s, l}
}
func (h *JournalHandler) List(c *gin.Context) {
	lvl, _ := strconv.Atoi(c.Query("mood_level"))
	v, e := h.s.List(middleware.UserID(c), lvl)
	if e != nil {
		c.Error(e)
		return
	}
	h.logger.Info(constants.LogJournalListed)
	ok(c, v)
}
func (h *JournalHandler) Create(c *gin.Context) {
	var r dto.JournalRequest
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
func (h *JournalHandler) Update(c *gin.Context) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		c.Error(util.NewAppError(constants.CodeValidation, "Journal[id] update failed: invalid id", e))
		return
	}
	var r dto.JournalRequest
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
func (h *JournalHandler) Delete(c *gin.Context) {
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
