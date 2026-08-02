package service

import (
	"context"
	"time"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/errs"
	"go.uber.org/zap"
)

// RecordOperLog 异步落库一条操作日志，作为 middleware.OperLogRecorder 使用。
// 用独立 context，避免请求结束后 context 取消导致写入失败。
func (s *Service) RecordOperLog(e middleware.OperLogEntry) {
	go func() {
		log := model.SysOperLog{
			Username: e.Username, UserID: e.UserID, Method: e.Method,
			Path: e.Path, Query: e.Query, Body: e.Body, IP: e.IP,
			Status: e.Status, Code: e.Code, LatencyMS: e.LatencyMS, CreatedAt: e.CreatedAt,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.DB.WithContext(ctx).Create(&log).Error; err != nil {
			s.Logger.Warn("写操作日志失败", zap.Error(err))
		}
	}()
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
	db := s.DB.WithContext(ctx).Model(&model.SysOperLog{})
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
	if err := s.DB.WithContext(ctx).Where("1 = 1").Delete(&model.SysOperLog{}).Error; err != nil {
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
	db := s.DB.WithContext(ctx).Model(&model.SysLoginLog{})
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
	if err := s.DB.WithContext(ctx).Where("1 = 1").Delete(&model.SysLoginLog{}).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}
