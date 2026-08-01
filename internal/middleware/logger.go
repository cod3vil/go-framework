package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CtxLogger 携带 request_id 的 logger 在 gin.Context 中的键，
// response.Error 等处会从 context 取它记录未知错误。
const CtxLogger = "logger"

// AccessLog 记录访问日志，并将带 request_id 字段的 logger 注入 context。
func AccessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqLogger := logger.With(zap.String("request_id", c.GetString(CtxRequestID)))
		c.Set(CtxLogger, reqLogger)

		c.Next()

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", time.Since(start)),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}
		switch {
		case c.Writer.Status() >= 500:
			reqLogger.Error("access", fields...)
		case c.Writer.Status() >= 400:
			reqLogger.Warn("access", fields...)
		default:
			reqLogger.Info("access", fields...)
		}
	}
}
