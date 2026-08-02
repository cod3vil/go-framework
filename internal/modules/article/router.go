package article

import "github.com/cod3vil/go-framework/internal/modkit"

// Register 装配文章模块并注册路由。这是业务模块的统一入口：
// 用 kit.DB 建 service/handler，用 kit.Secured 注册受保护路由。
//
// 在 internal/app/router.go 的 registerModules 中挂载一行即可启用：
//
//	article.Register(kit)
func Register(kit *modkit.Kit) {
	svc := NewService(kit)
	h := NewHandler(svc, kit)

	// 建表：业务模块自行迁移自己的表（主租户/public）。
	if err := kit.DB.AutoMigrate(&Article{}); err != nil {
		kit.Logger.Error("article 模块建表失败: " + err.Error())
	}
	// 登记模型：多租户开通新租户时在其 schema 内自动建表。
	kit.RegisterModels(&Article{})

	g := kit.Secured("/articles")
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.PUT("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}
