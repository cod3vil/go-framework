package middleware

import (
	"net/http"

	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit 全局令牌桶限流；rps <= 0 时返回空中间件（不限流）。
func RateLimit(rps, burst int) gin.HandlerFunc {
	if rps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	if burst <= 0 {
		burst = rps
	}
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, response.Body{
				Code: errs.CodeTooManyReq,
				Msg:  errs.ErrTooManyReq.Msg,
			})
			return
		}
		c.Next()
	}
}
