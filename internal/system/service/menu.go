package service

import (
	"context"
	"errors"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"gorm.io/gorm"
)

// MenuTree 查询全量菜单树（管理页使用，含按钮与隐藏项）。
func (s *Service) MenuTree(ctx context.Context) ([]*model.SysMenu, error) {
	var menus []*model.SysMenu
	if err := s.db(ctx).Order("sort, id").Find(&menus).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return buildMenuTree(menus, 0), nil
}

// MenuInput 创建/更新菜单参数。
type MenuInput struct {
	ParentID  uint
	Title     string
	Name      string
	Type      int8
	Path      string
	Component string
	Perm      string
	Icon      string
	Sort      int
	Visible   int8
	Status    int8
	KeepAlive bool
}

// CreateMenu 创建菜单。
func (s *Service) CreateMenu(ctx context.Context, in MenuInput, operator uint) (*model.SysMenu, error) {
	menu := model.SysMenu{
		ParentID: in.ParentID, Title: in.Title, Name: in.Name, Type: in.Type,
		Path: in.Path, Component: in.Component, Perm: in.Perm, Icon: in.Icon,
		Sort: in.Sort, Visible: in.Visible, Status: in.Status, KeepAlive: in.KeepAlive,
	}
	menu.CreatedBy = operator
	if err := s.db(ctx).Create(&menu).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &menu, nil
}

// UpdateMenu 更新菜单。
func (s *Service) UpdateMenu(ctx context.Context, id uint, in MenuInput, operator uint) error {
	var menu model.SysMenu
	err := s.db(ctx).First(&menu, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrNotFound
	}
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if in.ParentID == id {
		return errs.New(errs.CodeBadRequest, "上级菜单不能是自身")
	}
	err = s.db(ctx).Model(&menu).Updates(map[string]any{
		"parent_id": in.ParentID, "title": in.Title, "name": in.Name, "type": in.Type,
		"path": in.Path, "component": in.Component, "perm": in.Perm, "icon": in.Icon,
		"sort": in.Sort, "visible": in.Visible, "status": in.Status,
		"keep_alive": in.KeepAlive, "updated_by": operator,
	}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// DeleteMenu 删除菜单；存在子菜单时拒绝。
func (s *Service) DeleteMenu(ctx context.Context, id uint) error {
	var count int64
	s.db(ctx).Model(&model.SysMenu{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errs.New(errs.CodeConflict, "存在子菜单，无法删除")
	}
	err := s.db(ctx).Exec("DELETE FROM sys_role_menu WHERE sys_menu_id = ?", id).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	res := s.db(ctx).Unscoped().Delete(&model.SysMenu{}, id)
	if res.Error != nil {
		return errs.ErrInternal.WithCause(res.Error)
	}
	if res.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}
