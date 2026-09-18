package middleware

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/repository"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(constants.LogErrorWrapped, "panic", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Response{Code: constants.CodeInternal, Message: constants.MessageInternal})
			}
		}()
		c.Next()
		if len(c.Errors) == 0 || c.IsAborted() {
			return
		}
		e := c.Errors.Last().Err
		var app *util.AppError
		if errors.As(e, &app) {
			status := http.StatusBadRequest
			if app.Code == constants.CodeUnauthorized {
				status = http.StatusUnauthorized
			}
			c.JSON(status, dto.Response{Code: app.Code, Message: app.Message})
			return
		}
		if errors.Is(e, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Response{Code: constants.CodeNotFound, Message: "resource not found"})
			return
		}
		logger.Error(constants.LogErrorWrapped, "error", e)
		c.JSON(http.StatusInternalServerError, dto.Response{Code: constants.CodeInternal, Message: constants.MessageInternal})
	}
}
