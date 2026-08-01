package model

// 菜单类型。
const (
	MenuTypeDir    int8 = 1 // 目录
	MenuTypeMenu   int8 = 2 // 菜单（对应前端页面路由）
	MenuTypeButton int8 = 3 // 按钮（仅权限标识，不参与路由）
)

// SysMenu 菜单，驱动前端动态路由与按钮级权限。
type SysMenu struct {
	BaseModel
	ParentID uint   `gorm:"index;default:0" json:"parentId"`
	Title    string `gorm:"size:64;not null" json:"title"`
	// Name 前端路由名，Type 为按钮时为空。
	Name string `gorm:"size:64" json:"name"`
	Type int8   `gorm:"not null" json:"type"`
	Path string `gorm:"size:255" json:"path"`
	// Component 前端组件路径，如 system/user/index。
	Component string `gorm:"size:255" json:"component"`
	// Perm 权限标识，如 system:user:add，前端按钮显隐依据。
	Perm      string `gorm:"size:128" json:"perm"`
	Icon      string `gorm:"size:64" json:"icon"`
	Sort      int    `gorm:"default:0" json:"sort"`
	Visible   int8   `gorm:"default:1" json:"visible"` // 1 显示 2 隐藏
	Status    int8   `gorm:"default:1" json:"status"`
	KeepAlive bool   `gorm:"default:false" json:"keepAlive"`

	Children []*SysMenu `gorm:"-" json:"children,omitempty"`
}

// TableName 指定表名。
func (SysMenu) TableName() string { return "sys_menu" }
