// Package system 内置系统管理模块：认证、用户、角色、菜单、部门与 RBAC。
package system

import (
	"context"
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
	"github.com/cod3vil/go-framework/pkg/cronx"
	"github.com/cod3vil/go-framework/pkg/jwtx"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"github.com/cod3vil/go-framework/pkg/upload"
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
	// Tenancy 多租户库管理器（可为 nil，表示单租户）。
	Tenancy *tenancy.Manager
	// TenantModels 业务模块登记的模型指针，开通新租户时在其 schema 内建表。
	TenantModels *[]any
}

// Module 已装配的系统模块，暴露 Service/Enforcer/Cron 供其他模块复用与生命周期管理。
type Module struct {
	Service  *service.Service
	Enforcer *casbin.Enforcer
	Cron     *cronx.Manager
	// 共享中间件，供业务模块经 modkit 复用（认证/RBAC/写操作审计/租户路由）。
	AuthMW    gin.HandlerFunc
	RBACMW    gin.HandlerFunc
	OperLogMW gin.HandlerFunc
	TenantMW  gin.HandlerFunc
}

// Register 装配系统模块并注册路由到 api 分组（通常为 /api/v1）。
func Register(api *gin.RouterGroup, opt Options) (*Module, error) {
	enforcer, err := rbac.NewEnforcer(opt.DB)
	if err != nil {
		return nil, err
	}

	uploader, err := upload.New(opt.Config.Upload)
	if err != nil {
		return nil, err
	}

	cronManager := cronx.New()
	registerBuiltinTasks(cronManager, opt.Logger)

	svc := &service.Service{
		DB:       opt.DB,
		Cache:    opt.Cache,
		Logger:   opt.Logger,
		Config:   opt.Config,
		JWT:      jwtx.NewManager(opt.Config.JWT),
		Enforcer: enforcer,
		Captcha:  captcha.New(opt.Cache),
		Cron:     cronManager,
		Uploader: uploader,
		Tenancy:  opt.Tenancy,
	}
	// 开通新租户时在其 schema 内建系统表+种子，并迁移业务模块登记的模型。
	svc.MigrateTenant = func(tdb *gorm.DB) error {
		if err := Migrate(tdb); err != nil {
			return err
		}
		if opt.TenantModels != nil {
			for _, m := range *opt.TenantModels {
				if err := tdb.AutoMigrate(m); err != nil {
					return err
				}
			}
		}
		return nil
	}
	h := handler.New(svc)

	// 上传文件静态访问：/uploads/** 映射到本地存储目录。
	if opt.Config.Upload.Driver == "local" || opt.Config.Upload.Driver == "" {
		opt.Engine.Static(opt.Config.Upload.URLPrefix, opt.Config.Upload.Dir)
	}

	authMW := middleware.Auth(svc.JWT, opt.Cache)
	rbacMW := middleware.Casbin(enforcer, model.AdminRoleKey, opt.Logger)
	operLogMW := middleware.OperLog(svc.RecordOperLog)

	// 租户路由中间件：启用多租户时按 JWT/请求头解析租户并注入库句柄；否则空操作。
	tenantMW := func(c *gin.Context) { c.Next() }
	if svc.TenantEnabled() {
		tenantMW = middleware.Tenant(svc, opt.Config.Tenant.HeaderKey)
	}

	// 多租户开关探测（公开，供前端登录页决定是否显示租户输入）。
	api.GET("/tenant-enabled", h.TenantEnabled)

	// 认证接口：无需登录，但需先解析租户（登录时按请求头路由到对应租户库）。
	auth := api.Group("/auth", tenantMW)
	{
		auth.GET("/captcha", h.Captcha)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
	}
	// 需登录、无需 RBAC 的接口。
	authed := api.Group("/auth", authMW, tenantMW)
	{
		authed.POST("/logout", h.Logout)
		authed.GET("/userinfo", h.UserInfo)
	}

	// 系统管理接口：登录 + 租户路由 + RBAC + 写操作审计。
	sys := api.Group("/system", authMW, tenantMW, rbacMW, operLogMW)
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
			roles.GET("/:id/depts", h.GetRoleDepts)
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
		dicts := sys.Group("/dicts")
		{
			dicts.GET("", h.ListDicts)
			dicts.POST("", h.CreateDict)
			dicts.PUT("/:id", h.UpdateDict)
			dicts.DELETE("/:id", h.DeleteDict)
			dicts.GET("/:type/items", h.ListDictItems)
		}
		dictItems := sys.Group("/dict-items")
		{
			dictItems.POST("", h.CreateDictItem)
			dictItems.PUT("/:id", h.UpdateDictItem)
			dictItems.DELETE("/:id", h.DeleteDictItem)
		}
		configs := sys.Group("/configs")
		{
			configs.GET("", h.ListConfigs)
			configs.POST("", h.CreateConfig)
			configs.GET("/key/:key", h.GetConfigByKey)
			configs.PUT("/:id", h.UpdateConfig)
			configs.DELETE("/:id", h.DeleteConfig)
		}
		operLogs := sys.Group("/oper-logs")
		{
			operLogs.GET("", h.ListOperLogs)
			operLogs.DELETE("", h.ClearOperLogs)
		}
		loginLogs := sys.Group("/login-logs")
		{
			loginLogs.GET("", h.ListLoginLogs)
			loginLogs.DELETE("", h.ClearLoginLogs)
		}
		jobs := sys.Group("/jobs")
		{
			jobs.GET("", h.ListJobs)
			jobs.GET("/tasks", h.ListJobTasks)
			jobs.POST("", h.CreateJob)
			jobs.PUT("/:id", h.UpdateJob)
			jobs.PUT("/:id/status", h.SetJobStatus)
			jobs.POST("/:id/run", h.RunJob)
			jobs.DELETE("/:id", h.DeleteJob)
		}
		sys.GET("/job-logs", h.ListJobLogs)
		files := sys.Group("/files")
		{
			files.GET("", h.ListFiles)
			files.POST("", h.UploadFile)
			files.GET("/:id/download", h.DownloadFile)
			files.DELETE("/:id", h.DeleteFile)
		}
		sys.GET("/monitor/server", h.ServerMonitor)

		// 租户管理（平台级，作用于 public 注册表）。
		tenants := sys.Group("/tenants")
		{
			tenants.GET("", h.ListTenants)
			tenants.POST("", h.CreateTenant)
			tenants.PUT("/:id/status", h.SetTenantStatus)
			tenants.DELETE("/:id", h.DeleteTenant)
		}

		// 已注册 API 清单，供角色管理页勾选 API 权限。
		sys.GET("/apis", listAPIs(opt.Engine))
	}

	return &Module{
		Service:   svc,
		Enforcer:  enforcer,
		Cron:      cronManager,
		AuthMW:    authMW,
		RBACMW:    rbacMW,
		OperLogMW: operLogMW,
		TenantMW:  tenantMW,
	}, nil
}

// Init 在服务启动后调用：加载并启动定时任务调度。
func (m *Module) Init(ctx context.Context) error {
	return m.Service.InitJobs(ctx)
}

// Stop 在服务关闭时调用：停止定时任务调度，等待执行中的任务完成。
func (m *Module) Stop() {
	if m.Cron != nil {
		m.Cron.Stop()
	}
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
