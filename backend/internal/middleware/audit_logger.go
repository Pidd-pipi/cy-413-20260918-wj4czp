package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func AuditLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		b := make([]byte, 6)
		if _, e := rand.Read(b); e == nil {
			c.Set("request_id", hex.EncodeToString(b))
		}
		start := time.Now()
		logger.Info(constants.LogRequestStarted, "request_id", c.GetString("request_id"), "method", c.Request.Method, "path", c.Request.URL.Path)
		c.Next()
		logger.Info(constants.LogRequestCompleted, "request_id", c.GetString("request_id"), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "latency_ms", time.Since(start).Milliseconds())
		if c.Request.Method != "GET" && c.Writer.Status() < 400 {
			logger.Info(constants.LogAuditWrite, "method", c.Request.Method, "path", c.Request.URL.Path)
		}
	}
}
