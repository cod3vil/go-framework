package app

import (
	"net/http"
	"time"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// setupRouter 注册全局中间件与各模块路由。
// 新增业务模块时在 registerModules 中挂载一行即可。
func (a *App) setupRouter() {
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
	a.registerModules(api)
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

// registerModules 挂载系统模块与业务模块。
// P2 起在此注册 system 模块；业务模块同样在此挂载：
//
//	article.Register(a.toolkit(), api)
func (a *App) registerModules(api *gin.RouterGroup) {
	_ = api
}
