package model

import "time"

// 登录结果。
const (
	LoginSuccess int8 = 1
	LoginFailed  int8 = 2
)

// SysLoginLog 登录日志。
type SysLoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"size:64;index" json:"username"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:512" json:"userAgent"`
	Status    int8      `json:"status"`
	Msg       string    `gorm:"size:255" json:"msg"`
	CreatedAt time.Time `gorm:"index" json:"createdAt"`
}

// TableName 指定表名。
func (SysLoginLog) TableName() string { return "sys_login_log" }
