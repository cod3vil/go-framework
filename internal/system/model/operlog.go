package model

import "time"

// SysOperLog 操作审计日志，记录写操作（POST/PUT/DELETE/PATCH）。
type SysOperLog struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Username string `gorm:"size:64;index" json:"username"`
	UserID   uint   `gorm:"index" json:"userId"`
	Method   string `gorm:"size:16" json:"method"`
	Path     string `gorm:"size:255" json:"path"`
	// Query URL 查询串；Body 请求体（已脱敏 password 字段，超长截断）。
	Query string `gorm:"size:1024" json:"query"`
	Body  string `gorm:"type:text" json:"body"`
	IP    string `gorm:"size:64" json:"ip"`
	// Status HTTP 状态码；Code 业务响应码。
	Status int `json:"status"`
	Code   int `json:"code"`
	// LatencyMS 耗时毫秒。
	LatencyMS int64     `json:"latencyMs"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
}

// TableName 指定表名。
func (SysOperLog) TableName() string { return "sys_oper_log" }
