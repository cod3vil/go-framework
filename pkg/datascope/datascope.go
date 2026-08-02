// Package datascope 提供行级数据权限（数据范围）能力。
//
// RBAC 解决“能不能访问这个接口”（功能权限），数据权限解决“能看到哪些行”（数据范围）。
// 根据用户所属角色的数据范围与其部门，过滤查询结果。
//
// 用法：解析出当前用户的 Scope 后，作为 GORM Scopes 应用到查询：
//
//	scope := ...            // 由 system 服务解析
//	db.Scopes(scope.GormScope("dept_id", "created_by")).Find(&rows)
package datascope

import "gorm.io/gorm"

// 数据范围类型（与 SysRole.DataScope 取值对应）。
const (
	ScopeAll        int8 = 1 // 全部数据
	ScopeCustom     int8 = 2 // 自定数据权限（角色关联的部门集合）
	ScopeDept       int8 = 3 // 本部门数据
	ScopeDeptAndSub int8 = 4 // 本部门及以下数据
	ScopeSelf       int8 = 5 // 仅本人数据
)

// Scope 解析后的有效数据范围。多角色时取并集（越宽越优先）。
type Scope struct {
	// All 为 true 时不施加任何过滤（拥有全部数据权限）。
	All bool
	// DeptIDs 可见部门集合（本部门/本部门及以下/自定义 合并去重）。
	DeptIDs []uint
	// IncludeSelf 为 true 时额外放行“本人创建”的行。
	IncludeSelf bool
	// UserID 当前用户 ID（IncludeSelf 时使用）。
	UserID uint
}

// GormScope 返回一个 GORM Scopes 函数，按 deptCol / userCol 施加数据范围过滤。
//   - All：不加条件
//   - 有部门集合或仅本人：拼成 (dept_col IN (...) OR user_col = ?) 的并集条件
//   - 既无部门也非本人（无任何可见范围）：用恒假条件，保证查不到数据
func (s Scope) GormScope(deptCol, userCol string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if s.All {
			return db
		}
		hasDept := len(s.DeptIDs) > 0
		switch {
		case hasDept && s.IncludeSelf:
			return db.Where(db.Session(&gorm.Session{NewDB: true}).
				Where(deptCol+" IN ?", s.DeptIDs).Or(userCol+" = ?", s.UserID))
		case hasDept:
			return db.Where(deptCol+" IN ?", s.DeptIDs)
		case s.IncludeSelf:
			return db.Where(userCol+" = ?", s.UserID)
		default:
			// 无任何可见范围：恒假，避免误放行全部数据。
			return db.Where("1 = 0")
		}
	}
}
