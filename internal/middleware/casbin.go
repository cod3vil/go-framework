package middleware

import (
	"net/http"
	"slices"

	"github.com/casbin/casbin/v3"
	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Casbin RBAC 鉴权中间件：按 (角色, 路径, 方法) 检查策略。
// adminRoleKey 角色直接放行；用户任一角色命中策略即放行。
func Casbin(enforcer *casbin.Enforcer, adminRoleKey string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := RoleKeys(c)
		if slices.Contains(roles, adminRoleKey) {
			c.Next()
			return
		}
		obj := c.Request.URL.Path
		act := c.Request.Method
		for _, role := range roles {
			ok, err := enforcer.Enforce(role, obj, act)
			if err != nil {
				logger.Error("casbin enforce error", zap.Error(err),
					zap.String("role", role), zap.String("obj", obj), zap.String("act", act))
				continue
			}
			if ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, response.Body{
			Code: errs.CodeForbidden,
			Msg:  errs.ErrForbidden.Msg,
		})
	}
}
