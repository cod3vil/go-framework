// Package tenancy 实现基于 PostgreSQL schema 的多租户（SaaS）隔离。
//
// 隔离机制：每个租户对应一个独立的 PostgreSQL schema，并为其维护一个独立的
// 连接池，池内所有连接通过 DSN 的 search_path 固定到该 schema。由于连接池
// 天然隔离且 search_path 随连接固定，跨租户查询在连接池下也不会串号。
//
// public schema 承载租户注册表（sys_tenant）与主租户（primary）的全部数据。
package tenancy

import (
	"context"

	"gorm.io/gorm"
)

// PrimarySchema 主租户所在的 schema（复用现有 public，实现向后兼容）。
const PrimarySchema = "public"

// PrimaryCode 主租户编码。
const PrimaryCode = "primary"

// SchemaOf 返回租户编码对应的 schema 名。主租户用 public，其余为 tenant_<code>。
func SchemaOf(code string) string {
	if code == "" || code == PrimaryCode {
		return PrimarySchema
	}
	return "tenant_" + code
}

// Tenant 请求所属租户。
type Tenant struct {
	Code   string
	Schema string
}

type tenantCtxKey struct{}
type dbCtxKey struct{}

// WithTenant 将租户信息写入 context。
func WithTenant(ctx context.Context, t *Tenant) context.Context {
	return context.WithValue(ctx, tenantCtxKey{}, t)
}

// TenantFrom 从 context 取租户信息，未设置返回 nil。
func TenantFrom(ctx context.Context) *Tenant {
	if v, ok := ctx.Value(tenantCtxKey{}).(*Tenant); ok {
		return v
	}
	return nil
}

// WithDB 将租户数据库句柄写入 context。
func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbCtxKey{}, db)
}

// DBFrom 从 context 取租户数据库句柄，未设置返回 nil（调用方回退到基础库）。
func DBFrom(ctx context.Context) *gorm.DB {
	if v, ok := ctx.Value(dbCtxKey{}).(*gorm.DB); ok {
		return v
	}
	return nil
}
