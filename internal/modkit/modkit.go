// Package modkit 为业务模块提供统一的依赖工具箱与路由分组助手，
// 使业务模块只依赖本包，而不直接耦合 system 模块内部。
package modkit

import (
	"context"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/datascope"
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
	// 共享中间件：认证、RBAC 鉴权、写操作审计。
	Auth    gin.HandlerFunc
	RBAC    gin.HandlerFunc
	OperLog gin.HandlerFunc

	// ResolveScope 解析指定用户的行级数据范围（由 system 模块提供）。
	ResolveScope func(ctx context.Context, userID uint) (datascope.Scope, error)
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

// Secured 返回挂载了「认证 + RBAC + 审计」的路由分组，
// 业务模块用它注册需要登录与权限控制的接口：
//
//	g := kit.Secured("/articles")
//	g.GET("", h.List)
func (k *Kit) Secured(relativePath string) *gin.RouterGroup {
	return k.API.Group(relativePath, k.Auth, k.RBAC, k.OperLog)
}

// Public 返回仅在 /api/v1 下、不做鉴权的路由分组（如对外只读接口）。
func (k *Kit) Public(relativePath string) *gin.RouterGroup {
	return k.API.Group(relativePath)
}
