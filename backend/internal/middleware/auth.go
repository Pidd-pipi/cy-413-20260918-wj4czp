package middleware

import (
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func Auth(secret, issuer string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{Code: constants.CodeUnauthorized, Message: "Authorization token required"})
			return
		}
		claims, e := util.ParseToken(strings.TrimPrefix(h, "Bearer "), secret, issuer)
		if e != nil {
			logger.Warn(constants.LogErrorWrapped, "error", e)
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{Code: constants.CodeUnauthorized, Message: constants.MessageUnauthorized})
			return
		}
		id, e := strconv.ParseUint(claims.Subject, 10, 64)
		if e != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{Code: constants.CodeUnauthorized, Message: constants.MessageUnauthorized})
			return
		}
		c.Set("userID", uint(id))
		c.Set("role", claims.Role)
		c.Next()
	}
}
func RequireRole(role string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		got, _ := c.Get("role")
		if got != role {
			logger.Warn(constants.LogErrorWrapped, "reason", "role denied")
			c.AbortWithStatusJSON(http.StatusForbidden, dto.Response{Code: constants.CodeForbidden, Message: constants.MessageForbidden})
			return
		}
		c.Next()
	}
}
func UserID(c *gin.Context) uint { v, _ := c.Get("userID"); id, _ := v.(uint); return id }
