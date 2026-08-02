// Package modkit 为业务模块提供统一的依赖工具箱与路由分组助手，
// 使业务模块只依赖本包，而不直接耦合 system 模块内部。
package modkit

import (
	"context"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/datascope"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Kit 业务模块可用的核心依赖与共享中间件。
type Kit struct {
	DB     *gorm.DB
	Cache  cache.Cache
	Logger *zap.Logger
	Config *config.Config

	// API 是 /api/v1 分组。
	API *gin.RouterGroup
	// 共享中间件：认证、RBAC 鉴权、写操作审计、租户路由。
	Auth    gin.HandlerFunc
	RBAC    gin.HandlerFunc
	OperLog gin.HandlerFunc
	Tenant  gin.HandlerFunc

	// ResolveScope 解析指定用户的行级数据范围（由 system 模块提供）。
	ResolveScope func(ctx context.Context, userID uint) (datascope.Scope, error)

	// tenantModels 业务模块登记的模型，供多租户开通新租户时在其 schema 内建表。
	tenantModels *[]any
}

// RegisterModels 登记业务模块的 GORM 模型，用于多租户开通新租户时自动建表。
// 业务模块在 Register 中调用：kit.RegisterModels(&Article{})。
func (k *Kit) RegisterModels(models ...any) {
	if k.tenantModels == nil {
		s := make([]any, 0)
		k.tenantModels = &s
	}
	*k.tenantModels = append(*k.tenantModels, models...)
}

// TenantModels 返回已登记的业务模型（供租户开通时迁移）。
func (k *Kit) TenantModels() []any {
	if k.tenantModels == nil {
		return nil
	}
	return *k.tenantModels
}

// SetTenantModelSink 由框架注入共享的模型登记表指针（内部使用）。
func (k *Kit) SetTenantModelSink(sink *[]any) { k.tenantModels = sink }

// DBOf 返回当前请求应使用的数据库句柄：多租户下为请求所属租户的库，否则为基础库。
// 业务模块的 service 应始终用它取库，而非缓存 kit.DB，以获得租户隔离：
//
//	func (s *Service) List(ctx context.Context) { s.kit.DBOf(ctx).Find(&rows) }
func (k *Kit) DBOf(ctx context.Context) *gorm.DB {
	if tdb := tenancy.DBFrom(ctx); tdb != nil {
		return tdb.WithContext(ctx)
	}
	return k.DB.WithContext(ctx)
}

// ScopeOf 解析当前请求登录用户的数据范围，业务模块用它做行级过滤：
//
//	scope, _ := kit.ScopeOf(c)
//	db.Scopes(scope.GormScope("dept_id", "created_by")).Find(&rows)
//
// 未配置解析器时返回“全部”，保证在无数据权限需求时不影响查询。
func (k *Kit) ScopeOf(c *gin.Context) (datascope.Scope, error) {
	if k.ResolveScope == nil {
		return datascope.Scope{All: true}, nil
	}
	return k.ResolveScope(c.Request.Context(), middleware.UserID(c))
}

// Secured 返回挂载了「认证 + 租户路由 + RBAC + 审计」的路由分组，
// 业务模块用它注册需要登录与权限控制的接口：
//
//	g := kit.Secured("/articles")
//	g.GET("", h.List)
func (k *Kit) Secured(relativePath string) *gin.RouterGroup {
	mws := []gin.HandlerFunc{k.Auth}
	if k.Tenant != nil {
		mws = append(mws, k.Tenant)
	}
	mws = append(mws, k.RBAC, k.OperLog)
	return k.API.Group(relativePath, mws...)
}

// Public 返回仅在 /api/v1 下、不做鉴权的路由分组（如对外只读接口）。
func (k *Kit) Public(relativePath string) *gin.RouterGroup {
	return k.API.Group(relativePath)
}
