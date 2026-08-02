package service

import (
	"context"
	"errors"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/database"
	"github.com/cod3vil/go-framework/pkg/datascope"
	"github.com/cod3vil/go-framework/pkg/errs"
	"gorm.io/gorm"
)

// RoleQuery 角色列表查询条件。
type RoleQuery struct {
	Page     int
	PageSize int
	Name     string
	Status   int8
}

// ListRoles 分页查询角色。
func (s *Service) ListRoles(ctx context.Context, q RoleQuery) ([]model.SysRole, int64, error) {
	db := s.db(ctx).Model(&model.SysRole{})
	if q.Name != "" {
		db = db.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Status != 0 {
		db = db.Where("status = ?", q.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var roles []model.SysRole
	err := db.Order("sort, id").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&roles).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return roles, total, nil
}

// RoleInput 创建/更新角色参数。
type RoleInput struct {
	Name   string
	Key    string
	Sort   int
	Status int8
	Remark string
	// DataScope 数据范围（见 pkg/datascope）；DeptIDs 仅在自定义(2)时生效。
	DataScope int8
	DeptIDs   []uint
}

// CreateRole 创建角色。
func (s *Service) CreateRole(ctx context.Context, in RoleInput, operator uint) (*model.SysRole, error) {
	var count int64
	s.db(ctx).Model(&model.SysRole{}).Where("key = ?", in.Key).Count(&count)
	if count > 0 {
		return nil, errs.New(errs.CodeConflict, "角色标识已存在")
	}
	if in.DataScope == 0 {
		in.DataScope = datascope.ScopeAll
	}
	role := model.SysRole{
		Name: in.Name, Key: in.Key, Sort: in.Sort, Status: in.Status,
		Remark: in.Remark, DataScope: in.DataScope,
	}
	role.CreatedBy = operator
	err := database.Tx(ctx, s.db(ctx), func(tx *gorm.DB) error {
		if err := tx.Create(&role).Error; err != nil {
			return err
		}
		return s.replaceRoleDepts(tx, &role, in)
	})
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &role, nil
}

// GetRole 查询角色详情。
func (s *Service) GetRole(ctx context.Context, id uint) (*model.SysRole, error) {
	var role model.SysRole
	err := s.db(ctx).First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ErrNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &role, nil
}

// UpdateRole 更新角色（角色标识 Key 创建后不可修改，避免策略失联）。
func (s *Service) UpdateRole(ctx context.Context, id uint, in RoleInput, operator uint) error {
	role, err := s.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if role.Key == model.AdminRoleKey && in.Status == model.StatusDisabled {
		return errs.New(errs.CodeForbidden, "超级管理员角色不允许停用")
	}
	if in.DataScope == 0 {
		in.DataScope = datascope.ScopeAll
	}
	err = database.Tx(ctx, s.db(ctx), func(tx *gorm.DB) error {
		if err := tx.Model(role).Updates(map[string]any{
			"name": in.Name, "sort": in.Sort, "status": in.Status,
			"remark": in.Remark, "data_scope": in.DataScope, "updated_by": operator,
		}).Error; err != nil {
			return err
		}
		return s.replaceRoleDepts(tx, role, in)
	})
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// replaceRoleDepts 同步角色自定义部门：非自定义范围时清空，自定义时整体替换。
func (s *Service) replaceRoleDepts(tx *gorm.DB, role *model.SysRole, in RoleInput) error {
	if in.DataScope != datascope.ScopeCustom {
		return tx.Model(role).Association("Depts").Clear()
	}
	depts := make([]model.SysDept, 0, len(in.DeptIDs))
	if len(in.DeptIDs) > 0 {
		if err := tx.Find(&depts, in.DeptIDs).Error; err != nil {
			return err
		}
	}
	return tx.Model(role).Association("Depts").Replace(depts)
}

// GetRoleDeptIDs 查询角色自定义数据权限的部门 ID 列表（供前端回显）。
func (s *Service) GetRoleDeptIDs(ctx context.Context, id uint) ([]uint, error) {
	if _, err := s.GetRole(ctx, id); err != nil {
		return nil, err
	}
	return s.roleDeptIDs(ctx, id), nil
}

// DeleteRole 删除角色；有用户关联时拒绝。
func (s *Service) DeleteRole(ctx context.Context, id uint) error {
	role, err := s.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if role.Key == model.AdminRoleKey {
		return errs.New(errs.CodeForbidden, "超级管理员角色不允许删除")
	}
	var cnt int64
	s.db(ctx).Table("sys_user_role").Where("sys_role_id = ?", id).Count(&cnt)
	if cnt > 0 {
		return errs.New(errs.CodeConflict, "角色已分配给用户，无法删除")
	}
	err = database.Tx(ctx, s.db(ctx), func(tx *gorm.DB) error {
		if err := tx.Model(role).Association("Menus").Clear(); err != nil {
			return err
		}
		if err := tx.Model(role).Association("Depts").Clear(); err != nil {
			return err
		}
		if _, err := s.Enforcer.RemoveFilteredPolicy(0, role.Key); err != nil {
			return err
		}
		return tx.Unscoped().Delete(role).Error
	})
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// GetRoleMenuIDs 查询角色绑定的菜单 ID 列表。
func (s *Service) GetRoleMenuIDs(ctx context.Context, id uint) ([]uint, error) {
	if _, err := s.GetRole(ctx, id); err != nil {
		return nil, err
	}
	ids := make([]uint, 0)
	err := s.db(ctx).Table("sys_role_menu").
		Where("sys_role_id = ?", id).Pluck("sys_menu_id", &ids).Error
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return ids, nil
}

// SetRoleMenus 绑定角色菜单（整体替换）。
func (s *Service) SetRoleMenus(ctx context.Context, id uint, menuIDs []uint) error {
	role, err := s.GetRole(ctx, id)
	if err != nil {
		return err
	}
	menus := make([]model.SysMenu, 0, len(menuIDs))
	if len(menuIDs) > 0 {
		if err := s.db(ctx).Find(&menus, menuIDs).Error; err != nil {
			return errs.ErrInternal.WithCause(err)
		}
	}
	if err := s.db(ctx).Model(role).Association("Menus").Replace(menus); err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// APIPermission 一条 API 权限（Casbin 策略）。
type APIPermission struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
}

// GetRoleAPIs 查询角色的 API 权限列表。
func (s *Service) GetRoleAPIs(ctx context.Context, id uint) ([]APIPermission, error) {
	role, err := s.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	policies, err := s.Enforcer.GetFilteredPolicy(0, role.Key)
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	apis := make([]APIPermission, 0, len(policies))
	for _, p := range policies {
		if len(p) >= 3 {
			apis = append(apis, APIPermission{Path: p[1], Method: p[2]})
		}
	}
	return apis, nil
}

// SetRoleAPIs 设置角色的 API 权限（整体替换 Casbin 策略，实时生效）。
func (s *Service) SetRoleAPIs(ctx context.Context, id uint, apis []APIPermission) error {
	role, err := s.GetRole(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.Enforcer.RemoveFilteredPolicy(0, role.Key); err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if len(apis) == 0 {
		return nil
	}
	rules := make([][]string, 0, len(apis))
	for _, a := range apis {
		rules = append(rules, []string{role.Key, a.Path, a.Method})
	}
	if _, err := s.Enforcer.AddPolicies(rules); err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}
