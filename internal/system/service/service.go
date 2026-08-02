// Package service 系统管理模块的业务逻辑层。
package service

import (
	"context"

	"github.com/casbin/casbin/v3"
	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/captcha"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/cronx"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"github.com/cod3vil/go-framework/pkg/upload"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 系统模块服务，聚合全部依赖；各业务方法分布在同包其他文件。
type Service struct {
	DB       *gorm.DB
	Cache    cache.Cache
	Logger   *zap.Logger
	Config   *config.Config
	JWT      *jwtx.Manager
	Enforcer *casbin.Enforcer
	Captcha  *captcha.Captcha
	Cron     *cronx.Manager
	Uploader *upload.Uploader

	// Tenancy 多租户库管理器；MigrateTenant 为新租户 schema 建表+种子的回调。
	Tenancy       *tenancy.Manager
	MigrateTenant func(db *gorm.DB) error
}

// db 返回当前请求应使用的数据库句柄：多租户下为请求所属租户的库，否则为基础库。
func (s *Service) db(ctx context.Context) *gorm.DB {
	if tdb := tenancy.DBFrom(ctx); tdb != nil {
		return tdb.WithContext(ctx)
	}
	return s.DB.WithContext(ctx)
}
