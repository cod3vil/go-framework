package service

import (
	"context"
	"errors"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"gorm.io/gorm"
)

// DictQuery 字典类型列表查询条件。
type DictQuery struct {
	Page     int
	PageSize int
	Name     string
	Type     string
	Status   int8
}

// ListDicts 分页查询字典类型。
func (s *Service) ListDicts(ctx context.Context, q DictQuery) ([]model.SysDict, int64, error) {
	db := s.DB.WithContext(ctx).Model(&model.SysDict{})
	if q.Name != "" {
		db = db.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Type != "" {
		db = db.Where("type LIKE ?", "%"+q.Type+"%")
	}
	if q.Status != 0 {
		db = db.Where("status = ?", q.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var dicts []model.SysDict
	err := db.Order("id").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&dicts).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return dicts, total, nil
}

// DictInput 创建/更新字典类型参数。
type DictInput struct {
	Name   string
	Type   string
	Status int8
	Remark string
}

// CreateDict 创建字典类型。
func (s *Service) CreateDict(ctx context.Context, in DictInput, operator uint) (*model.SysDict, error) {
	var count int64
	s.DB.WithContext(ctx).Model(&model.SysDict{}).Where("type = ?", in.Type).Count(&count)
	if count > 0 {
		return nil, errs.New(errs.CodeConflict, "字典类型已存在")
	}
	dict := model.SysDict{Name: in.Name, Type: in.Type, Status: in.Status, Remark: in.Remark}
	dict.CreatedBy = operator
	if err := s.DB.WithContext(ctx).Create(&dict).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &dict, nil
}

// UpdateDict 更新字典类型（type 不可改，避免与字典项失联）。
func (s *Service) UpdateDict(ctx context.Context, id uint, in DictInput, operator uint) error {
	var dict model.SysDict
	if err := s.firstByID(ctx, &dict, id); err != nil {
		return err
	}
	err := s.DB.WithContext(ctx).Model(&dict).Updates(map[string]any{
		"name": in.Name, "status": in.Status, "remark": in.Remark, "updated_by": operator,
	}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// DeleteDict 删除字典类型及其全部字典项。
func (s *Service) DeleteDict(ctx context.Context, id uint) error {
	var dict model.SysDict
	if err := s.firstByID(ctx, &dict, id); err != nil {
		return err
	}
	err := s.DB.WithContext(ctx).Where("dict_type = ?", dict.Type).Unscoped().Delete(&model.SysDictItem{}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if err := s.DB.WithContext(ctx).Unscoped().Delete(&dict).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// ListDictItems 查询某字典类型下的字典项（启用项，按 sort 排序），供前端下拉框使用。
func (s *Service) ListDictItems(ctx context.Context, dictType string) ([]model.SysDictItem, error) {
	var items []model.SysDictItem
	err := s.DB.WithContext(ctx).Where("dict_type = ? AND status = ?", dictType, model.StatusEnabled).
		Order("sort, id").Find(&items).Error
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return items, nil
}

// AllDictItems 查询某字典类型下全部字典项（含停用，管理页使用）。
func (s *Service) AllDictItems(ctx context.Context, dictType string) ([]model.SysDictItem, error) {
	var items []model.SysDictItem
	err := s.DB.WithContext(ctx).Where("dict_type = ?", dictType).Order("sort, id").Find(&items).Error
	if err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return items, nil
}

// DictItemInput 创建/更新字典项参数。
type DictItemInput struct {
	DictType  string
	Label     string
	Value     string
	Sort      int
	CSSClass  string
	ListClass string
	IsDefault bool
	Status    int8
	Remark    string
}

// CreateDictItem 创建字典项。
func (s *Service) CreateDictItem(ctx context.Context, in DictItemInput, operator uint) (*model.SysDictItem, error) {
	var count int64
	s.DB.WithContext(ctx).Model(&model.SysDict{}).Where("type = ?", in.DictType).Count(&count)
	if count == 0 {
		return nil, errs.New(errs.CodeBadRequest, "字典类型不存在")
	}
	item := model.SysDictItem{
		DictType: in.DictType, Label: in.Label, Value: in.Value, Sort: in.Sort,
		CSSClass: in.CSSClass, ListClass: in.ListClass, IsDefault: in.IsDefault,
		Status: in.Status, Remark: in.Remark,
	}
	item.CreatedBy = operator
	if err := s.DB.WithContext(ctx).Create(&item).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	return &item, nil
}

// UpdateDictItem 更新字典项。
func (s *Service) UpdateDictItem(ctx context.Context, id uint, in DictItemInput, operator uint) error {
	var item model.SysDictItem
	if err := s.firstByID(ctx, &item, id); err != nil {
		return err
	}
	err := s.DB.WithContext(ctx).Model(&item).Updates(map[string]any{
		"label": in.Label, "value": in.Value, "sort": in.Sort,
		"css_class": in.CSSClass, "list_class": in.ListClass, "is_default": in.IsDefault,
		"status": in.Status, "remark": in.Remark, "updated_by": operator,
	}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// DeleteDictItem 删除字典项。
func (s *Service) DeleteDictItem(ctx context.Context, id uint) error {
	res := s.DB.WithContext(ctx).Unscoped().Delete(&model.SysDictItem{}, id)
	if res.Error != nil {
		return errs.ErrInternal.WithCause(res.Error)
	}
	if res.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// firstByID 通用按主键查询，未找到返回 ErrNotFound。
func (s *Service) firstByID(ctx context.Context, dest any, id uint) error {
	err := s.DB.WithContext(ctx).First(dest, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ErrNotFound
	}
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}
