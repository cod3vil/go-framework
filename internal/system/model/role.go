package model

// AdminRoleKey 超级管理员角色标识，拥有全部权限（绕过 Casbin 检查）。
const AdminRoleKey = "admin"

// SysRole 角色。
type SysRole struct {
	BaseModel
	Name   string `gorm:"size:64;not null" json:"name"`
	Key    string `gorm:"size:64;uniqueIndex;not null" json:"key"`
	Sort   int    `gorm:"default:0" json:"sort"`
	Status int8   `gorm:"default:1" json:"status"`
	Remark string `gorm:"size:255" json:"remark"`
	// DataScope 数据权限范围（预留）：1 全部 2 本部门 3 本部门及以下 4 仅本人。
	DataScope int8 `gorm:"default:1" json:"dataScope"`

	Menus []SysMenu `gorm:"many2many:sys_role_menu" json:"menus,omitempty"`
}

// TableName 指定表名。
func (SysRole) TableName() string { return "sys_role" }
