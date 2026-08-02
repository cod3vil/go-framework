package model

import "time"

// SysJob 定时任务定义。JobKey 对应 cronx 中注册的任务处理器。
type SysJob struct {
	BaseModel
	Name     string `gorm:"size:64;not null" json:"name"`
	JobKey   string `gorm:"size:64;not null" json:"jobKey"`
	CronExpr string `gorm:"size:64;not null" json:"cronExpr"`
	Status   int8   `gorm:"default:2" json:"status"` // 默认停用，避免误触发
	Remark   string `gorm:"size:255" json:"remark"`
}

// TableName 指定表名。
func (SysJob) TableName() string { return "sys_job" }

// SysJobLog 定时任务执行日志。
type SysJobLog struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	JobID      uint      `gorm:"index" json:"jobId"`
	JobKey     string    `gorm:"size:64" json:"jobKey"`
	Success    bool      `json:"success"`
	Message    string    `gorm:"size:1024" json:"message"`
	DurationMS int64     `json:"durationMs"`
	CreatedAt  time.Time `gorm:"index" json:"createdAt"`
}

// TableName 指定表名。
func (SysJobLog) TableName() string { return "sys_job_log" }
