// Package app 提供应用容器：装配核心组件（Config/Logger/DB/Cache），
// 负责 HTTP 服务的启动与优雅关闭。
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cod3vil/go-framework/internal/system"
	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/database"
	"github.com/cod3vil/go-framework/pkg/logger"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 应用容器，持有全部核心资源，模块注册路由时按需取用。
type App struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Cache  cache.Cache
	Engine *gin.Engine
	// System 系统管理模块，暴露 Enforcer/Service 供业务模块复用。
	System *system.Module
	// Tenancy 多租户库管理器，未启用时为 nil。
	Tenancy *tenancy.Manager
	// adminFS 管理后台前端文件系统，为空时不挂载 /admin。
	adminFS fs.FS
}

// Option 配置应用容器的可选项。
type Option func(*App)

// WithAdminFS 注入管理后台前端文件系统，挂载到 /admin。
// 由 main 传入嵌入的 web/dist，使 app 包不依赖前端产物。
func WithAdminFS(fsys fs.FS) Option {
	return func(a *App) { a.adminFS = fsys }
}

// New 按序初始化各组件：配置 → 日志 → 数据库 → 缓存 → HTTP 引擎与路由。
func New(configPath string, opts ...Option) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("初始化日志失败: %w", err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, err
	}
	log.Info("数据库连接成功", zap.String("driver", cfg.Database.Driver))

	var c cache.Cache
	if cfg.Redis.Addr != "" {
		c, err = cache.NewRedis(cfg.Redis)
		if err != nil {
			return nil, err
		}
		log.Info("Redis 连接成功", zap.String("addr", cfg.Redis.Addr))
	} else {
		c = cache.NewMemory()
		log.Info("未配置 Redis，使用内存缓存")
	}

	gin.SetMode(ginMode(cfg.App.Mode))
	app := &App{
		Config: cfg,
		Logger: log,
		DB:     db,
		Cache:  c,
		Engine: gin.New(),
	}

	// 多租户：启用且为 PostgreSQL 时创建租户库管理器。
	if cfg.Tenant.Enabled {
		mgr := tenancy.NewManager(cfg.Database, cfg.Tenant.MaxConnsPerTenant)
		if !mgr.Supported() {
			return nil, fmt.Errorf("多租户仅支持 PostgreSQL，当前驱动: %s", cfg.Database.Driver)
		}
		app.Tenancy = mgr
		log.Info("多租户已启用（PostgreSQL schema 隔离）")
	}

	for _, opt := range opts {
		opt(app)
	}
	if err := app.setupRouter(); err != nil {
		return nil, err
	}
	return app, nil
}

func ginMode(mode string) string {
	switch mode {
	case "release":
		return gin.ReleaseMode
	case "test":
		return gin.TestMode
	default:
		return gin.DebugMode
	}
}

// Run 启动 HTTP 服务并阻塞，收到 SIGINT/SIGTERM 后优雅关闭：
// 停止接收新请求 → 等待处理中请求（最长 shutdown_timeout 秒）→ 释放资源。
func (a *App) Run() error {
	srv := &http.Server{
		Addr:         a.Config.Server.Addr(),
		Handler:      a.Engine,
		ReadTimeout:  time.Duration(a.Config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(a.Config.Server.WriteTimeout) * time.Second,
	}

	// 启动定时任务调度（加载已启用任务）。
	if a.System != nil {
		if err := a.System.Init(context.Background()); err != nil {
			a.Logger.Warn("初始化定时任务失败", zap.Error(err))
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		a.Logger.Info("HTTP 服务启动", zap.String("addr", srv.Addr), zap.String("mode", a.Config.App.Mode))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		a.close()
		return fmt.Errorf("HTTP 服务异常退出: %w", err)
	case <-ctx.Done():
	}

	a.Logger.Info("收到退出信号，开始优雅关闭...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(),
		time.Duration(a.Config.Server.ShutdownTimeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		a.Logger.Warn("优雅关闭超时，强制退出", zap.Error(err))
	}
	a.close()
	a.Logger.Info("服务已退出")
	return nil
}

// close 释放数据库、缓存等资源。
func (a *App) close() {
	if a.System != nil {
		a.System.Stop() // 停止 cron 调度，等待执行中的任务完成
	}
	if a.Tenancy != nil {
		a.Tenancy.Close() // 关闭所有租户连接池
	}
	if a.Cache != nil {
		if err := a.Cache.Close(); err != nil {
			a.Logger.Warn("关闭缓存失败", zap.Error(err))
		}
	}
	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				a.Logger.Warn("关闭数据库失败", zap.Error(err))
			}
		}
	}
	_ = a.Logger.Sync()
}
