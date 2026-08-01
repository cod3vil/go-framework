package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HeaderRequestID 请求 ID 的 Header 名。
const HeaderRequestID = "X-Request-ID"

// CtxRequestID 请求 ID 在 gin.Context 中的键。
const CtxRequestID = "request_id"

// RequestID 为每个请求生成/透传请求 ID，写入响应头并注入 context，
// 便于日志串联一次请求的完整链路。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(CtxRequestID, id)
		c.Header(HeaderRequestID, id)
		c.Next()
	}
}
