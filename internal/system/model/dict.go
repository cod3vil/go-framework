package model

// SysDict 字典类型，如 "用户性别"(sys_user_sex)。
type SysDict struct {
	BaseModel
	Name   string `gorm:"size:64;not null" json:"name"`
	Type   string `gorm:"size:64;uniqueIndex;not null" json:"type"`
	Status int8   `gorm:"default:1" json:"status"`
	Remark string `gorm:"size:255" json:"remark"`

	Items []SysDictItem `gorm:"-" json:"items,omitempty"`
}

// TableName 指定表名。
func (SysDict) TableName() string { return "sys_dict" }

// SysDictItem 字典项，归属某个字典类型，作为前端下拉框数据源。
type SysDictItem struct {
	BaseModel
	DictType string `gorm:"size:64;index;not null" json:"dictType"`
	Label    string `gorm:"size:64;not null" json:"label"`
	Value    string `gorm:"size:64;not null" json:"value"`
	Sort     int    `gorm:"default:0" json:"sort"`
	// CSSClass / ListClass 供前端标签样式（如 success/danger）使用。
	CSSClass  string `gorm:"size:64" json:"cssClass"`
	ListClass string `gorm:"size:64" json:"listClass"`
	IsDefault bool   `gorm:"default:false" json:"isDefault"`
	Status    int8   `gorm:"default:1" json:"status"`
	Remark    string `gorm:"size:255" json:"remark"`
}

// TableName 指定表名。
func (SysDictItem) TableName() string { return "sys_dict_item" }
