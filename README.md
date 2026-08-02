# go-framework

架构简单但功能强大的企业级 Go 开发框架，内置后端 API 框架与后台 Web 管理系统。

设计文档见 [docs/DESIGN.md](docs/DESIGN.md)，业务模块开发见 [docs/MODULE_GUIDE.md](docs/MODULE_GUIDE.md)。

## 技术栈

Gin · GORM · Casbin · Viper · Zap · JWT · React 18 + Ant Design 5（管理后台，embed 单二进制部署）

## 快速开始

```bash
# 开发模式启动（默认 PostgreSQL + Redis，见 configs/config.yaml）
make dev

# 无 PG/Redis 的本地快速体验：改 database.driver 为 sqlite、redis.addr 留空即可
# （sqlite 用纯 Go 驱动，Redis 缺省时自动降级为内存缓存）

# 编译单二进制
make build && ./bin/server
```

初始化与验证：

```bash
# 建表 + 种子数据（默认管理员 admin/admin123、默认角色/菜单/部门）
go run ./cmd/server migrate

# 修改管理员密码
go run ./cmd/server create-admin --password '新密码'

curl http://127.0.0.1:8080/api/v1/health
# {"code":0,"msg":"ok","data":{"database":"up","status":"up",...}}

# 端到端冒烟测试（需本机 Redis，用于读取验证码答案）
bash scripts/smoke-test.sh
```

已内置接口：登录（验证码/失败锁定/登录日志）、登出与令牌刷新（黑名单/轮换）、
用户/角色/菜单/部门管理、角色绑定菜单与 API 权限（Casbin 实时生效）、
字典与参数配置（带缓存）、操作审计日志、定时任务（cron 表达式，可启停/手动触发）、
文件上传下载、服务监控（CPU/内存/磁盘/运行时）。

API 文档（Swagger UI，由路由自动生成）：启动后访问 `http://127.0.0.1:8080/swagger`。

## 管理后台

React 18 + TypeScript + Ant Design 5，构建产物通过 `go:embed` 打进二进制。启动服务后访问
`http://127.0.0.1:8080/admin`（默认 admin / admin123）。已内置：登录（验证码）、动态菜单路由与
按钮级权限、工作台，以及用户/角色/菜单/部门/字典/参数/定时任务/操作日志/登录日志/文件/服务监控
全部管理页面。

前端开发（热更新，代理到本地 Go 服务）：

```bash
make web-dev   # 或 cd web && pnpm dev，访问 http://127.0.0.1:5173/admin/
```

重建前端并打包单二进制：

```bash
make all       # = make web + make build
```

> `web/dist` 已提交入库，因此在没有 Node 环境时 `go build` 也能直接产出含完整后台的单二进制。

## 多租户（SaaS）

内置基于 PostgreSQL schema 的多租户隔离：每个租户独立 schema + 独立连接池，跨租户零串号；
`public` 承载租户注册表与主租户。默认关闭（单租户），开启：

```yaml
tenant:
  enabled: true   # 或环境变量 APP_TENANT_ENABLED=true（仅 PostgreSQL）
```

在后台「租户管理」开通租户会自动创建其 schema、迁移系统表与业务表并写入种子。租户用户登录时
在登录页填写租户编码（或请求头 `X-Tenant`）。业务模块接入只需 service 用 `kit.DBOf(ctx)`
取库并 `kit.RegisterModels(...)` 登记模型。详见 [docs/MODULE_GUIDE.md](docs/MODULE_GUIDE.md)。

配置文件位于 `configs/config.yaml`，所有配置可用环境变量覆盖（前缀 `APP`，层级用下划线，如 `APP_SERVER_PORT=9000`）。

## 目录结构

```
cmd/server/        CLI 入口（serve / version）
configs/           配置文件
internal/app/      应用容器、路由总入口、优雅启停
internal/middleware/ 框架中间件（recovery/requestid/访问日志/cors/限流）
internal/system/   内置系统管理模块（P2+）
internal/modules/  业务模块目录
pkg/               可复用核心库（config/logger/database/cache/response/errs）
web/               管理后台前端（P4）
```

## 业务模块开发

在 `internal/modules/<模块名>/` 下按 model/service/handler/router 四文件组织，
用 `kit.Secured(path)` 注册受保护路由（自动获得认证 + RBAC + 审计），再到
`internal/app/router.go` 的 `registerModules` 挂载一行即可。内置示例模块 `article`
（后端 `internal/modules/article/`，前端 `web/src/pages/article/ArticlePage.tsx`，
冒烟测试 `scripts/smoke-test-article.sh`）演示了完整前后端流程。详见
[docs/MODULE_GUIDE.md](docs/MODULE_GUIDE.md)。

## 路线图

- [x] P1 框架骨架：核心库、App 容器、中间件、健康检查
- [x] P2 认证授权：JWT 双令牌、Casbin RBAC、用户/角色/菜单/部门
- [x] P3 系统管理：字典、参数、审计日志、定时任务、文件上传、监控、Swagger
- [x] P4 管理后台前端：React + Ant Design，embed 单二进制
- [x] P5 示例模块与文档：article 示例模块 + 模块开发指南，v0.1.0
