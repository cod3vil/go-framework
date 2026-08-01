package service

import "github.com/cod3vil/go-framework/internal/system/model"

// buildMenuTree 将平铺菜单按 ParentID 组装为树（输入需已按 sort 排序）。
func buildMenuTree(menus []*model.SysMenu, parentID uint) []*model.SysMenu {
	tree := make([]*model.SysMenu, 0)
	for _, m := range menus {
		if m.ParentID == parentID {
			m.Children = buildMenuTree(menus, m.ID)
			tree = append(tree, m)
		}
	}
	return tree
}

// buildDeptTree 将平铺部门按 ParentID 组装为树。
func buildDeptTree(depts []*model.SysDept, parentID uint) []*model.SysDept {
	tree := make([]*model.SysDept, 0)
	for _, d := range depts {
		if d.ParentID == parentID {
			d.Children = buildDeptTree(depts, d.ID)
			tree = append(tree, d)
		}
	}
	return tree
}
