package service

import (
	"context"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"go.uber.org/zap"
)

const configCachePrefix = "sys:config:"

// ConfigQuery 参数列表查询条件。
type ConfigQuery struct {
	Page     int
	PageSize int
	Name     string
	Key      string
}

// ListConfigs 分页查询参数。
func (s *Service) ListConfigs(ctx context.Context, q ConfigQuery) ([]model.SysConfig, int64, error) {
	db := s.db(ctx).Model(&model.SysConfig{})
	if q.Name != "" {
		db = db.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.Key != "" {
		db = db.Where("key LIKE ?", "%"+q.Key+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var configs []model.SysConfig
	err := db.Order("id").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&configs).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return configs, total, nil
}

// GetConfigValue 按 key 读取参数值，优先走缓存。供业务代码读取运行参数。
func (s *Service) GetConfigValue(ctx context.Context, key string) (string, error) {
	if val, err := s.Cache.Get(ctx, configCachePrefix+key); err == nil {
		return val, nil
	}
	var cfg model.SysConfig
	if err := s.db(ctx).Where("key = ?", key).First(&cfg).Error; err != nil {
		return "", errs.ErrNotFound
	}
	_ = s.Cache.Set(ctx, configCachePrefix+key, cfg.Value, 0)
	return cfg.Value, nil
}

// ConfigInput 创建/更新参数。
type ConfigInput struct {
	Name   string
	Key    string
	Value  string
	Remark string
}

// CreateConfig 创建参数。
func (s *Service) CreateConfig(ctx context.Context, in ConfigInput, operator uint) (*model.SysConfig, error) {
	var count int64
	s.db(ctx).Model(&model.SysConfig{}).Where("key = ?", in.Key).Count(&count)
	if count > 0 {
		return nil, errs.New(errs.CodeConflict, "参数键已存在")
	}
	cfg := model.SysConfig{Name: in.Name, Key: in.Key, Value: in.Value, Remark: in.Remark}
	cfg.CreatedBy = operator
	if err := s.db(ctx).Create(&cfg).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	_ = s.Cache.Set(ctx, configCachePrefix+cfg.Key, cfg.Value, 0)
	return &cfg, nil
}

// UpdateConfig 更新参数值并刷新缓存。
func (s *Service) UpdateConfig(ctx context.Context, id uint, in ConfigInput, operator uint) error {
	var cfg model.SysConfig
	if err := s.firstByID(ctx, &cfg, id); err != nil {
		return err
	}
	err := s.db(ctx).Model(&cfg).Updates(map[string]any{
		"name": in.Name, "value": in.Value, "remark": in.Remark, "updated_by": operator,
	}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if err := s.Cache.Set(ctx, configCachePrefix+cfg.Key, in.Value, 0); err != nil {
		s.Logger.Warn("刷新参数缓存失败", zap.String("key", cfg.Key), zap.Error(err))
	}
	return nil
}

// DeleteConfig 删除参数（内置参数不允许删除）并清缓存。
func (s *Service) DeleteConfig(ctx context.Context, id uint) error {
	var cfg model.SysConfig
	if err := s.firstByID(ctx, &cfg, id); err != nil {
		return err
	}
	if cfg.Builtin {
		return errs.New(errs.CodeForbidden, "内置参数不允许删除")
	}
	if err := s.db(ctx).Unscoped().Delete(&cfg).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	_ = s.Cache.Del(ctx, configCachePrefix+cfg.Key)
	return nil
}
