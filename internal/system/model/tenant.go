package model

// SysTenant 租户注册表，存于 public schema。
type SysTenant struct {
	BaseModel
	// Code 租户编码（字母数字下划线），决定其 schema 名 tenant_<code>。
	Code string `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Name string `gorm:"size:128;not null" json:"name"`
	// Schema 实际使用的 PostgreSQL schema 名。
	Schema  string `gorm:"size:64;not null" json:"schema"`
	Status  int8   `gorm:"default:1" json:"status"`
	Contact string `gorm:"size:64" json:"contact"`
	Remark  string `gorm:"size:255" json:"remark"`
	// Primary 主租户（public），不可删除。
	Primary bool `gorm:"default:false" json:"primary"`
}

// TableName 指定表名。
func (SysTenant) TableName() string { return "sys_tenant" }
