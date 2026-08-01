// Package system 内置系统管理模块：认证、用户、角色、菜单、部门与 RBAC。
package system

import (
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/handler"
	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/internal/system/rbac"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/cache"
	"github.com/cod3vil/go-framework/pkg/captcha"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Options 模块依赖，由应用容器装配传入。
type Options struct {
	DB     *gorm.DB
	Cache  cache.Cache
	Logger *zap.Logger
	Config *config.Config
	Engine *gin.Engine
}

// Module 已装配的系统模块，暴露 Enforcer 等供其他模块复用。
type Module struct {
	Service  *service.Service
	Enforcer *casbin.Enforcer
}

// Register 装配系统模块并注册路由到 api 分组（通常为 /api/v1）。
func Register(api *gin.RouterGroup, opt Options) (*Module, error) {
	enforcer, err := rbac.NewEnforcer(opt.DB)
	if err != nil {
		return nil, err
	}

	svc := &service.Service{
		DB:       opt.DB,
		Cache:    opt.Cache,
		Logger:   opt.Logger,
		Config:   opt.Config,
		JWT:      jwtx.NewManager(opt.Config.JWT),
		Enforcer: enforcer,
		Captcha:  captcha.New(opt.Cache),
	}
	h := handler.New(svc)

	authMW := middleware.Auth(svc.JWT, opt.Cache)
	rbacMW := middleware.Casbin(enforcer, model.AdminRoleKey, opt.Logger)

	// 认证接口：无需登录。
	auth := api.Group("/auth")
	{
		auth.GET("/captcha", h.Captcha)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
	}
	// 需登录、无需 RBAC 的接口。
	authed := api.Group("/auth", authMW)
	{
		authed.POST("/logout", h.Logout)
		authed.GET("/userinfo", h.UserInfo)
	}

	// 系统管理接口：登录 + RBAC。
	sys := api.Group("/system", authMW, rbacMW)
	{
		users := sys.Group("/users")
		{
			users.GET("", h.ListUsers)
			users.POST("", h.CreateUser)
			users.GET("/:id", h.GetUser)
			users.PUT("/:id", h.UpdateUser)
			users.DELETE("/:id", h.DeleteUser)
			users.PUT("/:id/password", h.ResetUserPassword)
			users.PUT("/:id/status", h.SetUserStatus)
		}
		roles := sys.Group("/roles")
		{
			roles.GET("", h.ListRoles)
			roles.POST("", h.CreateRole)
			roles.GET("/:id", h.GetRole)
			roles.PUT("/:id", h.UpdateRole)
			roles.DELETE("/:id", h.DeleteRole)
			roles.GET("/:id/menus", h.GetRoleMenus)
			roles.PUT("/:id/menus", h.SetRoleMenus)
			roles.GET("/:id/apis", h.GetRoleAPIs)
			roles.PUT("/:id/apis", h.SetRoleAPIs)
		}
		menus := sys.Group("/menus")
		{
			menus.GET("/tree", h.MenuTree)
			menus.POST("", h.CreateMenu)
			menus.PUT("/:id", h.UpdateMenu)
			menus.DELETE("/:id", h.DeleteMenu)
		}
		depts := sys.Group("/depts")
		{
			depts.GET("/tree", h.DeptTree)
			depts.POST("", h.CreateDept)
			depts.PUT("/:id", h.UpdateDept)
			depts.DELETE("/:id", h.DeleteDept)
		}
		// 已注册 API 清单，供角色管理页勾选 API 权限。
		sys.GET("/apis", listAPIs(opt.Engine))
	}

	return &Module{Service: svc, Enforcer: enforcer}, nil
}

// apiInfo 一条已注册的 API 路由。
type apiInfo struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// listAPIs 枚举需要授权的已注册路由（排除认证与健康检查等公开接口）。
// 延迟到请求时读取 Routes()，确保全部模块注册完毕。
func listAPIs(engine *gin.Engine) gin.HandlerFunc {
	skip := []string{"/api/v1/auth/", "/api/v1/health", "/api/v1/ping"}
	return func(c *gin.Context) {
		apis := make([]apiInfo, 0)
	next:
		for _, r := range engine.Routes() {
			if !strings.HasPrefix(r.Path, "/api/v1/") {
				continue
			}
			for _, s := range skip {
				if strings.HasPrefix(r.Path, s) {
					continue next
				}
			}
			apis = append(apis, apiInfo{Path: r.Path, Method: r.Method})
		}
		response.OK(c, apis)
	}
}
