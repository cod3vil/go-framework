package service

import (
	"context"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/datascope"
)

// ResolveDataScope 解析用户的有效数据范围。多角色时取并集（越宽越优先）：
// 任一角色为“全部”即全部；否则合并各角色的可见部门与“仅本人”标记。
// 超级管理员始终为“全部”。
func (s *Service) ResolveDataScope(ctx context.Context, userID uint) (datascope.Scope, error) {
	var user model.SysUser
	err := s.db(ctx).Preload("Roles").First(&user, userID).Error
	if err != nil {
		return datascope.Scope{}, err
	}
	if user.IsAdmin() {
		return datascope.Scope{All: true}, nil
	}

	scope := datascope.Scope{UserID: userID}
	deptSet := make(map[uint]struct{})
	var descendantsCache map[uint][]uint

	for _, role := range user.Roles {
		if role.Status != model.StatusEnabled {
			continue
		}
		switch role.DataScope {
		case datascope.ScopeAll:
			return datascope.Scope{All: true}, nil
		case datascope.ScopeSelf:
			scope.IncludeSelf = true
		case datascope.ScopeDept:
			if user.DeptID != 0 {
				deptSet[user.DeptID] = struct{}{}
			}
		case datascope.ScopeDeptAndSub:
			if user.DeptID != 0 {
				if descendantsCache == nil {
					descendantsCache = s.deptDescendants(ctx)
				}
				deptSet[user.DeptID] = struct{}{}
				for _, id := range descendantsCache[user.DeptID] {
					deptSet[id] = struct{}{}
				}
			}
		case datascope.ScopeCustom:
			for _, id := range s.roleDeptIDs(ctx, role.ID) {
				deptSet[id] = struct{}{}
			}
		}
	}

	for id := range deptSet {
		scope.DeptIDs = append(scope.DeptIDs, id)
	}
	return scope, nil
}

// deptDescendants 构建 部门ID -> 全部后代部门ID 的映射（一次查询全量部门后在内存中展开）。
func (s *Service) deptDescendants(ctx context.Context) map[uint][]uint {
	var depts []model.SysDept
	s.db(ctx).Select("id", "parent_id").Find(&depts)
	children := make(map[uint][]uint)
	for _, d := range depts {
		children[d.ParentID] = append(children[d.ParentID], d.ID)
	}
	result := make(map[uint][]uint)
	var walk func(id uint) []uint
	walk = func(id uint) []uint {
		var acc []uint
		for _, c := range children[id] {
			acc = append(acc, c)
			acc = append(acc, walk(c)...)
		}
		return acc
	}
	for _, d := range depts {
		result[d.ID] = walk(d.ID)
	}
	return result
}

// roleDeptIDs 查询角色自定义数据权限授权的部门 ID 列表。
func (s *Service) roleDeptIDs(ctx context.Context, roleID uint) []uint {
	ids := make([]uint, 0)
	s.db(ctx).Table("sys_role_dept").
		Where("sys_role_id = ?", roleID).Pluck("sys_dept_id", &ids)
	return ids
}
