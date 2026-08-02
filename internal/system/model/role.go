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
	// DataScope 数据权限范围（见 pkg/datascope）：
	// 1 全部 2 自定义部门 3 本部门 4 本部门及以下 5 仅本人。
	DataScope int8 `gorm:"default:1" json:"dataScope"`

	Menus []SysMenu `gorm:"many2many:sys_role_menu" json:"menus,omitempty"`
	// Depts 自定义数据权限时授权的部门集合（DataScope=2 时生效）。
	Depts []SysDept `gorm:"many2many:sys_role_dept" json:"depts,omitempty"`
}

// TableName 指定表名。
func (SysRole) TableName() string { return "sys_role" }
