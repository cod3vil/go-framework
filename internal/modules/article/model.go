// Package article 是一个示例业务模块，演示基于框架开发 CRUD 模块的完整流程：
// model → service → handler → router，并通过 modkit.Kit 复用认证/鉴权/审计能力。
package article

import "time"

// 文章状态。
const (
	StatusDraft     int8 = 1 // 草稿
	StatusPublished int8 = 2 // 已发布
)

// Article 文章模型。业务表以模块名为前缀，避免与系统表冲突。
type Article struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Author    string    `gorm:"size:64" json:"author"`
	Content   string    `gorm:"type:text" json:"content"`
	Status    int8      `gorm:"default:1;index" json:"status"`
	Views     int       `gorm:"default:0" json:"views"`
	CreatedBy uint      `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName 指定表名。
func (Article) TableName() string { return "biz_article" }
