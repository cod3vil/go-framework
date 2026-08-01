// Package model 定义系统管理模块的数据模型，表名统一前缀 sys_。
package model

import (
	"time"

	"gorm.io/gorm"
)

// 通用状态值。
const (
	StatusEnabled  int8 = 1 // 启用
	StatusDisabled int8 = 2 // 停用
)

// BaseModel 业务表通用字段。
// 注意：username/role key 等带唯一索引的实体删除时使用 Unscoped 硬删除，
// 避免软删除残留记录占用唯一键。
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy uint           `json:"createdBy,omitempty"`
	UpdatedBy uint           `json:"updatedBy,omitempty"`
}
