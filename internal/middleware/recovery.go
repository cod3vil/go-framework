package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕获 panic，记录堆栈并返回统一 500 响应。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("request_id", c.GetString(CtxRequestID)),
					zap.String("path", c.Request.URL.Path),
					zap.ByteString("stack", debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Body{
					Code: errs.CodeInternal,
					Msg:  errs.ErrInternal.Msg,
				})
			}
		}()
		c.Next()
	}
}
