# Go Framework 设计文档

> 一个架构简单但功能强大的企业级 Go 开发框架，内置后端 API 框架与后台 Web 管理系统。

## 1. 目标与设计原则

**目标**：让团队用最少的心智负担快速搭建企业级后端应用 —— 开箱即有认证授权、RBAC 权限、系统管理后台等企业级通用能力，业务方只需专注写业务模块。

**设计原则**：

1. **简单优先**：三层结构（handler → service → model），不引入 repository/usecase 等额外抽象层，不做过度设计。需要抽象时再抽象。
2. **约定优于配置**：目录结构、命名、响应格式、错误码都有统一约定，模块之间长得一样。
3. **单二进制交付**：前端构建产物通过 `go:embed` 打进二进制，一个文件即可部署（也支持前后端分离部署）。
4. **模块化单体**：按业务域分模块（package），模块内自治、模块间通过 service 接口调用，为将来可能的拆分保留边界。
5. **依赖显式传递**：不用全局变量魔法，核心资源（DB、Redis、Logger、Config）通过应用容器显式注入。

## 2. 技术选型

| 领域 | 选型 | 说明 |
|------|------|------|
| HTTP 框架 | [Gin](https://github.com/gin-gonic/gin) | 生态成熟、中间件丰富 |
| ORM | [GORM](https://gorm.io) | 默认 PostgreSQL，兼容 MySQL / SQLite，自动迁移 |
| 配置 | [Viper](https://github.com/spf13/viper) | YAML 配置 + 环境变量覆盖 |
| 日志 | [Zap](https://github.com/uber-go/zap) + lumberjack | 结构化日志、按大小滚动切割 |
| 认证 | JWT（golang-jwt/jwt/v5） | Access Token + Refresh Token 双令牌 |
| 权限 | [Casbin](https://casbin.org) | RBAC 模型，策略存 DB，动态变更 |
| 缓存 | go-redis v9，默认启用 Redis | 统一 Cache 接口，未配置 Redis 时降级为内存缓存 |
| 定时任务 | robfig/cron/v3 | 后台可查看/启停任务 |
| 参数校验 | validator/v10（Gin 内置） | 统一翻译为中文错误信息 |
| API 文档 | swaggo/swag | 注释生成 Swagger UI |
| 管理前端 | React 18 + TypeScript + Vite + Ant Design 5 | 前端产物 embed 进二进制 |
| CLI | cobra | `server`、`migrate`、`create-admin` 等子命令 |

## 3. 总体架构

```
┌────────────────────────────────────────────────────────┐
│                      客户端                              │
│   管理后台 (React + AntD)   │   业务前端 / 第三方调用      │
└──────────────┬──────────────────────┬──────────────────┘
               │  /admin (embed 静态)  │  /api/v1
┌──────────────▼──────────────────────▼──────────────────┐
│                    Gin HTTP Server                     │
│  全局中间件: Recovery → RequestID → Logger → CORS       │
│            → RateLimit → JWT Auth → Casbin RBAC        │
│            → OperationLog                              │
├────────────────────────────────────────────────────────┤
│   系统模块 (内置)              │   业务模块 (使用方开发)    │
│  ┌──────────────────────┐    │  ┌──────────────────┐   │
│  │ user / role / menu   │    │  │ handler          │   │
│  │ dept / dict / config │    │  │   ↓              │   │
│  │ loginlog / operlog   │    │  │ service          │   │
│  │ job / file / monitor │    │  │   ↓              │   │
│  └──────────────────────┘    │  │ model (GORM)     │   │
│                              │  └──────────────────┘   │
├────────────────────────────────────────────────────────┤
│  核心层 pkg/: config·logger·db·cache·jwt·resp·errs·... │
├────────────────────────────────────────────────────────┤
│        MySQL / PostgreSQL / SQLite    │    Redis(可选)  │
└────────────────────────────────────────────────────────┘
```

## 4. 目录结构

```
go-framework/
├── cmd/
│   └── server/
│       └── main.go              # 入口：cobra 命令（server / migrate / create-admin / version）
├── configs/
│   ├── config.yaml              # 默认配置
│   └── config.example.yaml
├── internal/
│   ├── app/
│   │   ├── app.go               # 应用容器：装配 Config/Logger/DB/Cache/Cron，优雅启停
│   │   └── router.go            # 路由注册总入口，聚合各模块路由
│   ├── middleware/               # 框架中间件
│   │   ├── auth.go              # JWT 认证
│   │   ├── casbin.go            # RBAC 鉴权
│   │   ├── cors.go
│   │   ├── logger.go            # 访问日志 + RequestID
│   │   ├── ratelimit.go         # 令牌桶限流
│   │   ├── recovery.go
│   │   └── operlog.go           # 写操作审计日志
│   ├── system/                   # 内置系统管理模块（后台管理的后端）
│   │   ├── model/               # SysUser, SysRole, SysMenu, SysDept, SysDict,
│   │   │                        # SysConfig, SysLoginLog, SysOperLog, SysJob, SysFile
│   │   ├── service/
│   │   ├── handler/
│   │   └── router.go
│   └── modules/                  # 业务模块目录（示例：demo 文章模块）
│       └── article/
│           ├── model.go
│           ├── service.go
│           ├── handler.go
│           └── router.go
├── pkg/                          # 可复用核心库（不依赖 internal）
│   ├── config/                  # Viper 封装
│   ├── logger/                  # Zap 封装
│   ├── database/                # GORM 初始化、事务助手
│   ├── cache/                   # Cache 接口 + redis/memory 两种实现
│   ├── jwtx/                    # 双令牌签发与校验
│   ├── response/                # 统一响应
│   ├── errs/                    # 业务错误码
│   ├── cronx/                   # 定时任务管理器
│   ├── captcha/                 # 登录验证码
│   ├── upload/                  # 文件上传（本地磁盘，接口预留 OSS/S3）
│   └── utils/
├── web/                          # 管理后台前端
│   ├── src/
│   │   ├── api/                 # axios 封装 + 各模块 API
│   │   ├── components/
│   │   ├── layouts/             # 侧边菜单 + 顶栏布局
│   │   ├── pages/               # login / dashboard / system/* 各管理页
│   │   ├── router/              # 静态路由 + 按菜单接口生成的动态路由
│   │   └── store/               # zustand：用户信息、权限、菜单
│   ├── dist/                    # 构建产物（embed 目标）
│   └── vite.config.ts
├── web_embed.go                  # //go:embed web/dist
├── docs/
├── Makefile                      # make dev / build / swag / web / lint
├── go.mod
└── README.md
```

**分层约定**（保持简单的关键）：

- `handler`：解析参数、调 service、写响应。不写业务逻辑。
- `service`：业务逻辑、事务边界。返回 `(result, error)`，error 用 `errs` 包的业务错误。
- `model`：GORM 模型 + 表级别的查询方法。简单 CRUD 允许 service 直接操作 GORM，**不强制 repository 层**。

## 5. 核心机制设计

### 5.1 应用容器与生命周期

```go
type App struct {
    Config *config.Config
    Logger *zap.Logger
    DB     *gorm.DB
    Cache  cache.Cache
    Cron   *cronx.Manager
    Engine *gin.Engine
}
```

`app.New()` 按序初始化各组件，`app.Run()` 启动 HTTP 服务并监听信号，收到 SIGTERM 后优雅关闭（停止接收请求 → 等待处理中请求 → 停 cron → 关 DB/Redis）。模块注册路由时拿到 `*App`，需要什么用什么。

### 5.2 统一响应与错误码

所有 API 返回统一结构：

```json
{ "code": 0, "msg": "ok", "data": { ... } }
```

分页统一为 `data: { list: [], total: 100, page: 1, pageSize: 10 }`。

错误码分段：`0` 成功；`1xxx` 通用错误（1000 参数错误、1001 未认证、1003 无权限、1004 资源不存在）；`2xxx` 系统模块；业务模块从 `10xxx` 起各自分配一段。`errs.New(code, msg)` 创建业务错误，handler 层用 `response.Error(c, err)` 自动识别错误码；非业务 error 统一返回 500 且不向客户端泄漏内部信息。

### 5.3 认证与授权

- **登录**：账号密码（bcrypt）+ 图形验证码 → 签发 Access Token（2h）与 Refresh Token（7d）；连续失败 5 次锁定 10 分钟；登录日志入库。
- **认证中间件**：解析 `Authorization: Bearer`，将 `userID / username / roleKeys` 注入 context。
- **登出/踢人**：Redis 维护 token 黑名单（无 Redis 时降级为内存实现）。
- **授权**：Casbin RBAC，`p, role, /api/v1/system/users, GET`。角色-菜单-API 权限在后台配置，保存后实时生效。前端按钮级权限用权限标识（如 `system:user:add`）由菜单接口下发。
- **超级管理员**（`admin` 角色）绕过 Casbin 检查。

### 5.4 系统管理模块（内置企业级功能）

| 模块 | 功能 |
|------|------|
| 用户管理 | CRUD、分配角色/部门、重置密码、启用停用、导入导出 |
| 角色管理 | CRUD、绑定菜单权限与 API 权限（数据权限范围预留字段） |
| 菜单管理 | 树形菜单，支持目录/菜单/按钮三种类型，驱动前端动态路由与按钮权限 |
| 部门管理 | 树形部门 |
| 字典管理 | 通用键值字典（前端下拉框数据源） |
| 参数配置 | 系统运行参数在线修改，带缓存 |
| 操作日志 | 写操作（POST/PUT/DELETE）自动审计：谁、何时、什么接口、请求体、耗时、结果 |
| 登录日志 | 登录成功/失败记录，IP、UA、地点 |
| 定时任务 | cron 表达式任务的查看、启停、手动触发、执行日志 |
| 文件管理 | 上传下载、类型/大小白名单，存储接口预留 OSS/S3 扩展 |
| 服务监控 | Goroutine 数、内存、CPU、磁盘、在线用户 |

### 5.5 数据库约定

- 基础模型 `BaseModel`：`ID`（自增）、`CreatedAt`、`UpdatedAt`、`DeletedAt`（软删除）、`CreatedBy`、`UpdatedBy`。
- 系统表前缀 `sys_`，业务表按模块前缀。
- 开发环境用 GORM `AutoMigrate`；`migrate` 子命令负责建表和种子数据（默认管理员、默认菜单、Casbin 策略）。
- 事务：`database.Tx(ctx, func(tx *gorm.DB) error)` 助手，service 层控制事务边界。

### 5.6 管理后台前端

- React 18 + TS + Vite + Ant Design 5，状态用 zustand，请求用 axios（统一拦截器处理 token 刷新与错误码提示）。
- 登录后拉取 `/api/v1/auth/userinfo`（含菜单树与权限标识）→ 动态生成路由与侧边菜单；`<Auth perm="system:user:add">` 组件控制按钮显隐。
- 内置页面：登录、工作台，以及 5.4 中全部系统管理页面；提供一个 CRUD 页面模板（搜索表单 + 表格 + 弹窗表单），业务页面照抄即可。
- `make build` 时先 `vite build` 再 `go build`，产物经 `go:embed` 挂在 `/admin` 路径下。

### 5.7 业务模块开发流程（使用方视角）

1. 在 `internal/modules/` 下建包：`model.go`、`service.go`、`handler.go`、`router.go`；
2. 在 `router.go` 的 `Register(app, group)` 中注册路由；
3. 在 `internal/app/router.go` 中挂载该模块（一行）；
4. 后台「菜单管理」中配置菜单与权限点，分配给角色；
5. 前端 `pages/` 下按模板复制一个 CRUD 页面。

## 6. 实施路线图

| 阶段 | 内容 | 产出 |
|------|------|------|
| **P1 框架骨架** | 目录结构、config/logger/database/cache/response/errs、App 容器、优雅启停、基础中间件（recovery/cors/requestid/访问日志）、Makefile | 可跑通的 HTTP 服务 + 健康检查接口 |
| **P2 认证授权** | JWT 双令牌、验证码、登录/登出/刷新、Casbin RBAC、用户/角色/菜单/部门模型与接口、migrate 种子数据 | 完整的认证 + RBAC API |
| **P3 系统管理完善** | 字典、参数配置、操作/登录日志、定时任务、文件上传、服务监控、Swagger | 系统模块 API 全量完成 |
| **P4 管理后台前端** | React 工程搭建、登录、动态菜单路由、全部系统管理页面、embed 集成 | 单二进制含完整管理后台 |
| **P5 示例与文档** | article 示例模块（前后端）、README、模块开发指南 | 可交付的框架 v0.1.0 |

## 7. 明确不做的事（保持简单）

- 不做微服务/RPC/服务注册发现 —— 模块边界清晰即可，需要时再拆。
- 不做多租户（表结构预留 `tenant_id` 的决定推迟到有真实需求时）。
- 不做消息队列抽象 —— 业务需要时直接引入具体实现。
- 不做代码生成器（P5 之后如有需要再评估）。
