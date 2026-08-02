package middleware

import (
	"net/http"

	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CtxTenantSchema 请求所属租户 schema 在 gin.Context 中的键。
const CtxTenantSchema = "tenant_schema"

// TenantResolver 按租户编码解析租户信息与其数据库句柄。
type TenantResolver interface {
	ResolveTenantByCode(code string) (*tenancy.Tenant, *gorm.DB, error)
}

// Tenant 多租户中间件：解析请求所属租户并把其数据库句柄注入 context。
//   - 已认证请求：租户编码来自 JWT（Auth 中间件注入的 claims）。
//   - 未认证请求（登录/验证码）：租户编码来自请求头 headerKey（默认 X-Tenant）。
//
// 解析失败（租户不存在/停用）直接拒绝；空编码回退主租户（public）。
func Tenant(resolver TenantResolver, headerKey string) gin.HandlerFunc {
	if headerKey == "" {
		headerKey = "X-Tenant"
	}
	return func(c *gin.Context) {
		code := tenantCodeFromRequest(c, headerKey)

		tenant, db, err := resolver.ResolveTenantByCode(code)
		if err != nil {
			if e, ok := errs.From(err); ok {
				c.AbortWithStatusJSON(http.StatusOK, response.Body{Code: e.Code, Msg: e.Msg})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Body{
				Code: errs.CodeInternal, Msg: errs.ErrInternal.Msg,
			})
			return
		}

		c.Set(CtxTenantSchema, tenant.Schema)
		ctx := tenancy.WithTenant(c.Request.Context(), tenant)
		ctx = tenancy.WithDB(ctx, db)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// tenantCodeFromRequest 优先取 JWT 中的租户编码，其次取请求头。
func tenantCodeFromRequest(c *gin.Context, headerKey string) string {
	if v, ok := c.Get(CtxClaims); ok {
		if claims, ok := v.(*jwtx.Claims); ok && claims.Tenant != "" {
			return claims.Tenant
		}
	}
	return c.GetHeader(headerKey)
}
