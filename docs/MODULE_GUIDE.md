# 业务模块开发指南

本框架采用**模块化单体**架构：系统管理能力（认证、RBAC、审计等）内置在 `internal/system`，
业务功能以自治模块的形式放在 `internal/modules/<模块名>`。本文以内置的示例模块 `article`
为参照，说明如何新增一个业务 CRUD 模块。

> 示例源码：后端 `internal/modules/article/`，前端 `web/src/pages/article/ArticlePage.tsx`。

## 1. 后端：四个文件

每个模块是一个自治的 package，约定包含四个文件：

```
internal/modules/article/
├── model.go     # GORM 模型 + 表名（表名以业务前缀，如 biz_article）
├── service.go   # 业务逻辑与事务，返回 (result, error)，error 用 pkg/errs
├── handler.go   # 解析参数 → 调 service → 写统一响应，不含业务逻辑
└── router.go    # Register(kit) 装配并注册路由
```

### model.go

```go
type Article struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Title     string    `gorm:"size:200;not null" json:"title"`
    Status    int8      `gorm:"default:1;index" json:"status"`
    CreatedBy uint      `json:"createdBy"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

func (Article) TableName() string { return "biz_article" }
```

### service.go

- 构造函数 `NewService(db *gorm.DB) *Service`；
- 方法返回 `(结果, error)`，业务错误用 `errs.New(code, msg)`；
- **业务错误码从 `10xxx` 起**各模块分配一段（见 `pkg/errs` 分段约定），例如：

```go
const codeArticleNotFound = 10001
// ...
return nil, errs.New(codeArticleNotFound, "文章不存在")
```

- 简单 CRUD 允许 service 直接用 GORM，不强制 repository 层；需要事务时用
  `database.Tx(ctx, db, func(tx *gorm.DB) error { ... })`。

### handler.go

- 构造函数 `NewHandler(svc *Service) *Handler`；
- 只做「解析参数 → 调 service → 写响应」，用 `response.OK/OKPage/Error/BadRequest`；
- 当前登录用户 ID 用 `middleware.UserID(c)` 获取。

### router.go —— 模块入口

```go
func Register(kit *modkit.Kit) {
    svc := NewService(kit.DB)
    h := NewHandler(svc)

    // 业务模块自行迁移自己的表
    _ = kit.DB.AutoMigrate(&Article{})

    g := kit.Secured("/articles") // 已挂载 认证 + RBAC + 写操作审计
    g.GET("", h.List)
    g.POST("", h.Create)
    g.GET("/:id", h.Get)
    g.PUT("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
}
```

`modkit.Kit` 提供业务模块所需的一切：

| 字段/方法 | 说明 |
|-----------|------|
| `kit.DB` / `kit.Cache` / `kit.Logger` / `kit.Config` | 核心资源 |
| `kit.Secured(path)` | 返回挂载了 **认证 + RBAC 鉴权 + 写操作审计** 的路由分组 |
| `kit.Public(path)` | 返回 `/api/v1` 下不鉴权的分组（如对外只读接口） |

## 2. 挂载模块（一行）

在 `internal/app/router.go` 的 `registerModules` 中，业务模块注册区加一行：

```go
// —— 在此注册业务模块（每个模块一行）——
article.Register(kit)
// yourmodule.Register(kit)
```

至此后端接口即生效，且自动获得：统一响应格式、认证、RBAC 鉴权、写操作审计、
Swagger 文档（`/swagger` 自动收录新路由）。

## 3. 菜单与权限

在后台「菜单管理」中为模块新增菜单与按钮权限点（权限标识如 `article:add`），
再到「角色管理」把菜单与 API 权限分配给角色。API 权限基于 Casbin，保存后**实时生效**。

> 示例模块通过 `internal/system/migrate.go` 的 `seedArticleMenu` 预置了「内容管理」菜单
> 与 `article:*` 权限点，可作为业务模块登记菜单的参照。

## 4. 前端：一个页面 + 两处登记

1. **API**：在 `web/src/api/index.ts` 增加模块 API（参照 `articleApi`）；类型加到 `web/src/types.ts`。
2. **页面**：复制 `web/src/pages/article/ArticlePage.tsx`，用 `usePagedList` 拉数据，
   `Table` + `Modal` 表单实现 CRUD，`<Auth perm="article:add">` 控制按钮级权限。
3. **路由登记**：在 `web/src/router/routes.tsx` 增加一行，`path` 与后端菜单的 `path` 对齐：

```tsx
{ path: 'article', element: <ArticlePage /> },
```

侧边菜单由菜单接口动态生成，无需手写；按钮显隐由权限标识自动控制。

前端开发用 `make web-dev`（热更新，代理到本地 Go 服务）；提交前 `make web` 重建
`web/dist` 并随二进制 embed。

## 5. 检查清单

- [ ] `internal/modules/<name>/` 下 model/service/handler/router 四文件
- [ ] 业务错误码从 `10xxx` 起分配，不与其他模块冲突
- [ ] `router.go` 用 `kit.Secured` 注册，`AutoMigrate` 自己的表
- [ ] `internal/app/router.go` 中 `registerModules` 挂载一行
- [ ] 后台配置菜单与权限点并分配给角色
- [ ] 前端加 API/类型、复制页面、`routes.tsx` 登记一行
- [ ] 写一个 `scripts/smoke-test-<name>.sh` 冒烟测试（可选，参照 article）
