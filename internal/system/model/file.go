package model

// SysFile 上传文件记录。
type SysFile struct {
	BaseModel
	// Name 原始文件名；Key 存储相对路径；URL 对外访问地址。
	Name string `gorm:"size:255;not null" json:"name"`
	Key  string `gorm:"size:255;not null" json:"key"`
	URL  string `gorm:"size:512;not null" json:"url"`
	Ext  string `gorm:"size:32" json:"ext"`
	Size int64  `json:"size"`
	// Storage 存储驱动（local/oss/s3）。
	Storage string `gorm:"size:16" json:"storage"`
}

// TableName 指定表名。
func (SysFile) TableName() string { return "sys_file" }
