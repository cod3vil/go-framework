package middleware

import (
	"net/http"
	"strings"

	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// 认证信息在 gin.Context 中的键。
const (
	CtxUserID   = "auth_user_id"
	CtxUsername = "auth_username"
	CtxDeptID   = "auth_dept_id"
	CtxRoleKeys = "auth_role_keys"
	CtxClaims   = "auth_claims"
)

// BlacklistKeyPrefix 令牌黑名单缓存键前缀（登出/刷新后失效的令牌 jti）。
const BlacklistKeyPrefix = "auth:blacklist:"

// Auth JWT 认证中间件：解析 Bearer 令牌，校验黑名单，注入用户身份。
func Auth(jm *jwtx.Manager, store cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			abortUnauthorized(c)
			return
		}
		claims, err := jm.Parse(token, jwtx.TypeAccess)
		if err != nil {
			abortUnauthorized(c)
			return
		}
		if banned, _ := store.Exists(c.Request.Context(), BlacklistKeyPrefix+claims.ID); banned {
			abortUnauthorized(c)
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxUsername, claims.Username)
		c.Set(CtxDeptID, claims.DeptID)
		c.Set(CtxRoleKeys, claims.RoleKeys)
		c.Set(CtxClaims, claims)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
		return after
	}
	return ""
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, response.Body{
		Code: errs.CodeUnauthorized,
		Msg:  errs.ErrUnauthorized.Msg,
	})
}

// UserID 从 context 取当前登录用户 ID。
func UserID(c *gin.Context) uint {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// DeptID 从 context 取当前登录用户所属部门 ID。
func DeptID(c *gin.Context) uint {
	if v, ok := c.Get(CtxDeptID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// RoleKeys 从 context 取当前用户角色标识。
func RoleKeys(c *gin.Context) []string {
	if v, ok := c.Get(CtxRoleKeys); ok {
		if keys, ok := v.([]string); ok {
			return keys
		}
	}
	return nil
}
