package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// setupRouter 注册全局中间件与各模块路由。
// 新增业务模块时在 registerModules 中挂载一行即可。
func (a *App) setupRouter() error {
	e := a.Engine

	e.Use(
		middleware.RequestID(),
		middleware.Recovery(a.Logger),
		middleware.AccessLog(a.Logger),
		middleware.CORS(),
		middleware.RateLimit(a.Config.Server.RateLimit, a.Config.Server.RateBurst),
	)

	e.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.Body{Code: 1004, Msg: "接口不存在"})
	})

	api := e.Group("/api/v1")
	a.registerHealth(api)
	if err := a.registerModules(api); err != nil {
		return err
	}
	a.registerOpenAPI()
	a.registerAdmin()
	return nil
}

var startTime = time.Now()

// registerHealth 健康检查与存活探针。
func (a *App) registerHealth(api *gin.RouterGroup) {
	api.GET("/health", func(c *gin.Context) {
		dbStatus := "up"
		if sqlDB, err := a.DB.DB(); err != nil || sqlDB.Ping() != nil {
			dbStatus = "down"
		}
		response.OK(c, gin.H{
			"status":   "up",
			"database": dbStatus,
			"uptime":   time.Since(startTime).Round(time.Second).String(),
			"version":  Version,
		})
	})
	api.GET("/ping", func(c *gin.Context) {
		response.OK(c, "pong")
	})
}

// registerModules 挂载系统模块与业务模块，新模块在此追加注册。
func (a *App) registerModules(api *gin.RouterGroup) error {
	sysModule, err := system.Register(api, system.Options{
		DB:     a.DB,
		Cache:  a.Cache,
		Logger: a.Logger,
		Config: a.Config,
		Engine: a.Engine,
	})
	if err != nil {
		return fmt.Errorf("注册 system 模块失败: %w", err)
	}
	a.System = sysModule
	return nil
}
