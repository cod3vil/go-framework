// Package rbac 基于 Casbin 提供 RBAC 权限引擎。
//
// 策略模型：p, 角色key, URL路径, HTTP方法；
// 路径匹配用 keyMatch2，支持 /api/v1/system/users/:id 形式的通配。
// 超级管理员角色（admin）在中间件层直接放行，不写入策略。
package rbac

import (
	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

const modelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch2(r.obj, p.obj) && r.act == p.act
`

// NewEnforcer 创建 Casbin enforcer，策略存储于 sys_casbin_rule 表。
func NewEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &gormadapter.CasbinRule{}, "sys_casbin_rule")
	if err != nil {
		return nil, err
	}
	m, err := casbinmodel.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}
	return e, nil
}
