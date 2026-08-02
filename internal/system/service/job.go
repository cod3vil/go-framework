package service

import (
	"context"
	"time"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/cronx"
	"github.com/cod3vil/go-framework/pkg/errs"
	"go.uber.org/zap"
)

// InitJobs 启动时装配定时任务：设置执行日志回调、加载已启用任务并启动调度器。
func (s *Service) InitJobs(ctx context.Context) error {
	s.Cron.SetResultHook(s.recordJobResult)

	var jobs []model.SysJob
	if err := s.DB.WithContext(ctx).Where("status = ?", model.StatusEnabled).Find(&jobs).Error; err != nil {
		return err
	}
	for _, job := range jobs {
		if !s.Cron.HasTask(job.JobKey) {
			s.Logger.Warn("跳过未注册的定时任务", zap.String("jobKey", job.JobKey))
			continue
		}
		if err := s.Cron.Schedule(job.ID, job.JobKey, job.CronExpr); err != nil {
			s.Logger.Warn("调度定时任务失败", zap.String("name", job.Name), zap.Error(err))
		}
	}
	s.Cron.Start()
	return nil
}

func (s *Service) recordJobResult(r cronx.Result) {
	log := model.SysJobLog{
		JobID: r.JobID, JobKey: r.JobKey, Success: r.Success,
		Message: r.Message, DurationMS: r.Duration.Milliseconds(), CreatedAt: r.RunAt,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.DB.WithContext(ctx).Create(&log).Error; err != nil {
		s.Logger.Warn("写任务执行日志失败", zap.Error(err))
	}
}

// JobQuery 任务列表查询条件。
type JobQuery struct {
	Page     int
	PageSize int
	Name     string
	Status   int8
}

// ListJobs 分页查询任务。
func (s *Service) ListJobs(ctx context.Context, q JobQuery) ([]model.SysJob, int64, error) {
	db := s.DB.WithContext(ctx).Model(&model.SysJob{})
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
	var jobs []model.SysJob
	err := db.Order("id").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&jobs).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return jobs, total, nil
}

// RegisteredTasks 返回已注册的任务键，供前端选择。
func (s *Service) RegisteredTasks() []string {
	return s.Cron.TaskKeys()
}

// JobInput 创建/更新任务参数。
type JobInput struct {
	Name     string
	JobKey   string
	CronExpr string
	Status   int8
	Remark   string
}

func (s *Service) validateJob(in JobInput) error {
	if !s.Cron.HasTask(in.JobKey) {
		return errs.New(errs.CodeBadRequest, "任务处理器未注册: "+in.JobKey)
	}
	if err := s.Cron.ValidateSpec(in.CronExpr); err != nil {
		return errs.New(errs.CodeBadRequest, "无效的 cron 表达式")
	}
	return nil
}

// CreateJob 创建任务；若状态为启用则立即加入调度。
func (s *Service) CreateJob(ctx context.Context, in JobInput, operator uint) (*model.SysJob, error) {
	if err := s.validateJob(in); err != nil {
		return nil, err
	}
	job := model.SysJob{Name: in.Name, JobKey: in.JobKey, CronExpr: in.CronExpr, Status: in.Status, Remark: in.Remark}
	job.CreatedBy = operator
	if err := s.DB.WithContext(ctx).Create(&job).Error; err != nil {
		return nil, errs.ErrInternal.WithCause(err)
	}
	if job.Status == model.StatusEnabled {
		if err := s.Cron.Schedule(job.ID, job.JobKey, job.CronExpr); err != nil {
			s.Logger.Warn("调度新任务失败", zap.Error(err))
		}
	}
	return &job, nil
}

// UpdateJob 更新任务并同步调度状态。
func (s *Service) UpdateJob(ctx context.Context, id uint, in JobInput, operator uint) error {
	if err := s.validateJob(in); err != nil {
		return err
	}
	var job model.SysJob
	if err := s.firstByID(ctx, &job, id); err != nil {
		return err
	}
	err := s.DB.WithContext(ctx).Model(&job).Updates(map[string]any{
		"name": in.Name, "job_key": in.JobKey, "cron_expr": in.CronExpr,
		"status": in.Status, "remark": in.Remark, "updated_by": operator,
	}).Error
	if err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	// 重新同步调度：先移除，启用则按新表达式重新加入。
	s.Cron.Unschedule(id)
	if in.Status == model.StatusEnabled {
		if err := s.Cron.Schedule(id, in.JobKey, in.CronExpr); err != nil {
			return errs.New(errs.CodeBadRequest, err.Error())
		}
	}
	return nil
}

// SetJobStatus 启用/停用任务并同步调度。
func (s *Service) SetJobStatus(ctx context.Context, id uint, status int8, operator uint) error {
	var job model.SysJob
	if err := s.firstByID(ctx, &job, id); err != nil {
		return err
	}
	if err := s.DB.WithContext(ctx).Model(&job).Updates(map[string]any{"status": status, "updated_by": operator}).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	if status == model.StatusEnabled {
		if err := s.Cron.Schedule(id, job.JobKey, job.CronExpr); err != nil {
			return errs.New(errs.CodeBadRequest, err.Error())
		}
	} else {
		s.Cron.Unschedule(id)
	}
	return nil
}

// DeleteJob 删除任务并从调度中移除。
func (s *Service) DeleteJob(ctx context.Context, id uint) error {
	var job model.SysJob
	if err := s.firstByID(ctx, &job, id); err != nil {
		return err
	}
	s.Cron.Unschedule(id)
	if err := s.DB.WithContext(ctx).Unscoped().Delete(&job).Error; err != nil {
		return errs.ErrInternal.WithCause(err)
	}
	return nil
}

// RunJobOnce 手动立即执行一次任务。
func (s *Service) RunJobOnce(ctx context.Context, id uint) error {
	var job model.SysJob
	if err := s.firstByID(ctx, &job, id); err != nil {
		return err
	}
	if err := s.Cron.RunOnce(job.ID, job.JobKey); err != nil {
		return errs.New(errs.CodeBadRequest, err.Error())
	}
	return nil
}

// JobLogQuery 任务日志查询条件。
type JobLogQuery struct {
	Page     int
	PageSize int
	JobID    uint
}

// ListJobLogs 分页查询任务执行日志（按时间倒序）。
func (s *Service) ListJobLogs(ctx context.Context, q JobLogQuery) ([]model.SysJobLog, int64, error) {
	db := s.DB.WithContext(ctx).Model(&model.SysJobLog{})
	if q.JobID != 0 {
		db = db.Where("job_id = ?", q.JobID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	var logs []model.SysJobLog
	err := db.Order("id DESC").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&logs).Error
	if err != nil {
		return nil, 0, errs.ErrInternal.WithCause(err)
	}
	return logs, total, nil
}
