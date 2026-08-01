package model

// SysDept 部门（树形）。
type SysDept struct {
	BaseModel
	ParentID uint   `gorm:"index;default:0" json:"parentId"`
	Name     string `gorm:"size:64;not null" json:"name"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Leader   string `gorm:"size:64" json:"leader"`
	Phone    string `gorm:"size:32" json:"phone"`
	Email    string `gorm:"size:128" json:"email"`
	Status   int8   `gorm:"default:1" json:"status"`

	Children []*SysDept `gorm:"-" json:"children,omitempty"`
}

// TableName 指定表名。
func (SysDept) TableName() string { return "sys_dept" }
