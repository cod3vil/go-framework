package service

import (
	"context"
	"errors"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"gorm.io/gorm"
)

// DeptTree 查询部门树。
func (s *Service) DeptTree(ctx context.Context) ([]*model.SysDept, error) {
	var depts []*model.SysDept
	if err := s.DB.WithContext(ctx).Order("sort, id").Find(&depts).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return buildDeptTree(depts, 0), nil
}

// DeptInput 创建/更新部门参数。
type DeptInput struct {
	ParentID uint
	Name     string
	Sort     int
	Leader   string
	Phone    string
	Email    string
	Status   int8
}

// CreateDept 创建部门。
func (s *Service) CreateDept(ctx context.Context, in DeptInput, operator uint) (*model.SysDept, error) {
	dept := model.SysDept{
		ParentID: in.ParentID, Name: in.Name, Sort: in.Sort,
		Leader: in.Leader, Phone: in.Phone, Email: in.Email, Status: in.Status,
	}
	dept.CreatedBy = operator
	if err := s.DB.WithContext(ctx).Create(&dept).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &dept, nil
}

// UpdateDept 更新部门。
func (s *Service) UpdateDept(ctx context.Context, id uint, in DeptInput, operator uint) error {
	var dept model.SysDept
	err := s.DB.WithContext(ctx).First(&dept, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrNotFound
	}
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if in.ParentID == id {
		return errs.New(errs.CodeBadRequest, "上级部门不能是自身")
	}
	err = s.DB.WithContext(ctx).Model(&dept).Updates(map[string]any{
		"parent_id": in.ParentID, "name": in.Name, "sort": in.Sort,
		"leader": in.Leader, "phone": in.Phone, "email": in.Email,
		"status": in.Status, "updated_by": operator,
	}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// DeleteDept 删除部门；存在子部门或部门下有用户时拒绝。
func (s *Service) DeleteDept(ctx context.Context, id uint) error {
	var count int64
	s.DB.WithContext(ctx).Model(&model.SysDept{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return errs.New(errs.CodeConflict, "存在子部门，无法删除")
	}
	s.DB.WithContext(ctx).Model(&model.SysUser{}).Where("dept_id = ?", id).Count(&count)
	if count > 0 {
		return errs.New(errs.CodeConflict, "部门下存在用户，无法删除")
	}
	res := s.DB.WithContext(ctx).Unscoped().Delete(&model.SysDept{}, id)
	if res.Error != nil {
		return errs.ErrInternal.WithCause(res.Error)
	}
	if res.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}
