package model

import "time"

// AdminUsername 内置超级管理员账号，禁止删除/停用。
const AdminUsername = "admin"

// SysUser 系统用户。
type SysUser struct {
	BaseModel
	Username    string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password    string     `gorm:"size:128;not null" json:"-"`
	Nickname    string     `gorm:"size:64" json:"nickname"`
	Email       string     `gorm:"size:128" json:"email"`
	Phone       string     `gorm:"size:32" json:"phone"`
	Avatar      string     `gorm:"size:255" json:"avatar"`
	Status      int8       `gorm:"default:1" json:"status"`
	DeptID      uint       `gorm:"index" json:"deptId"`
	Remark      string     `gorm:"size:255" json:"remark"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	LastLoginIP string     `gorm:"size:64" json:"lastLoginIp"`

	Dept  *SysDept  `gorm:"foreignKey:DeptID" json:"dept,omitempty"`
	Roles []SysRole `gorm:"many2many:sys_user_role" json:"roles,omitempty"`
}

// TableName 指定表名。
func (SysUser) TableName() string { return "sys_user" }

// IsAdmin 是否内置超级管理员账号。
func (u *SysUser) IsAdmin() bool { return u.Username == AdminUsername }

// RoleKeys 返回用户的角色标识列表。
func (u *SysUser) RoleKeys() []string {
	keys := make([]string, 0, len(u.Roles))
	for _, r := range u.Roles {
		keys = append(keys, r.Key)
	}
	return keys
}
