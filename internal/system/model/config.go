package model

// SysConfig 系统参数配置，运行期可在线修改，读取带缓存。
type SysConfig struct {
	BaseModel
	Name  string `gorm:"size:64;not null" json:"name"`
	Key   string `gorm:"size:128;uniqueIndex;not null" json:"key"`
	Value string `gorm:"size:1024;not null" json:"value"`
	// Builtin 内置参数不允许删除。
	Builtin bool   `gorm:"default:false" json:"builtin"`
	Remark  string `gorm:"size:255" json:"remark"`
}

// TableName 指定表名。
func (SysConfig) TableName() string { return "sys_config" }
