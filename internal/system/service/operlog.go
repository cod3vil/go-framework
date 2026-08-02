package service

import (
	"context"
	"time"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/cod3vil/go-framework/pkg/tenancy"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RecordOperLog 异步落库一条操作日志，作为 middleware.OperLogRecorder 使用。
// 用独立 context，避免请求结束后 context 取消导致写入失败；
// 多租户下按 entry 携带的 schema 写入对应租户库。
func (s *Service) RecordOperLog(e middleware.OperLogEntry) {
	go func() {
		log := model.SysOperLog{
			Username: e.Username, UserID: e.UserID, Method: e.Method,
			Path: e.Path, Query: e.Query, Body: e.Body, IP: e.IP,
			Status: e.Status, Code: e.Code, LatencyMS: e.LatencyMS, CreatedAt: e.CreatedAt,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		db := s.dbForSchema(e.TenantSchema).WithContext(ctx)
		if err := db.Create(&log).Error; err != nil {
			s.Logger.Warn("写操作日志失败", zap.Error(err))
		}
	}()
}

// dbForSchema 返回指定 schema 的库句柄；空或 public 或未启用多租户时返回基础库。
func (s *Service) dbForSchema(schema string) *gorm.DB {
	if schema == "" || schema == tenancy.PrimarySchema || !s.TenantEnabled() {
		return s.DB
	}
	if db, err := s.Tenancy.DB(schema); err == nil {
		return db
	}
	return s.DB
}

// OperLogQuery 操作日志查询条件。
type OperLogQuery struct {
	Page     int
	PageSize int
	Username string
	Path     string
	Method   string
}

// ListOperLogs 分页查询操作日志（按时间倒序）。
func (s *Service) ListOperLogs(ctx context.Context, q OperLogQuery) ([]model.SysOperLog, int64, error) {
	db := s.db(ctx).Model(&model.SysOperLog{})
	if q.Username != "" {
		db = db.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Path != "" {
		db = db.Where("path LIKE ?", "%"+q.Path+"%")
	}
	if q.Method != "" {
		db = db.Where("method = ?", q.Method)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var logs []model.SysOperLog
	err := db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&logs).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return logs, total, nil
}

// ClearOperLogs 清空全部操作日志。
func (s *Service) ClearOperLogs(ctx context.Context) error {
	if err := s.db(ctx).Where("1 = 1").Delete(&model.SysOperLog{}).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// LoginLogQuery 登录日志查询条件。
type LoginLogQuery struct {
	Page     int
	PageSize int
	Username string
	Status   int8
}

// ListLoginLogs 分页查询登录日志（按时间倒序）。
func (s *Service) ListLoginLogs(ctx context.Context, q LoginLogQuery) ([]model.SysLoginLog, int64, error) {
	db := s.db(ctx).Model(&model.SysLoginLog{})
	if q.Username != "" {
		db = db.Where("username LIKE ?", "%"+q.Username+"%")
	}
	if q.Status != 0 {
		db = db.Where("status = ?", q.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var logs []model.SysLoginLog
	err := db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&logs).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return logs, total, nil
}

// ClearLoginLogs 清空全部登录日志。
func (s *Service) ClearLoginLogs(ctx context.Context) error {
	if err := s.db(ctx).Where("1 = 1").Delete(&model.SysLoginLog{}).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}
