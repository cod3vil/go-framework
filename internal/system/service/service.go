// Package service 系统管理模块的业务逻辑层。
package service

import (
	"github.com/casbin/casbin/v3"
	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/captcha"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/jwtx"
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
}
