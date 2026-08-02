package system

import (
	"fmt"

	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/internal/system/rbac"
	"github.com/cod3vil/go-framework/pkg/utils"
	"gorm.io/gorm"
)

// Migrate 建表（含 Casbin 策略表）并写入种子数据（仅在对应表为空时执行）。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.SysDept{},
		&model.SysUser{},
		&model.SysRole{},
		&model.SysMenu{},
		&model.SysLoginLog{},
		&model.SysDict{},
		&model.SysDictItem{},
		&model.SysConfig{},
		&model.SysOperLog{},
		&model.SysJob{},
		&model.SysJobLog{},
		&model.SysFile{},
	); err != nil {
		return fmt.Errorf("自动迁移失败: %w", err)
	}
	// Casbin 策略表由 adapter 建立。
	if _, err := rbac.NewEnforcer(db); err != nil {
		return fmt.Errorf("初始化权限表失败: %w", err)
	}
	return seed(db)
}

func seed(db *gorm.DB) error {
	if err := seedDept(db); err != nil {
		return err
	}
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedAdmin(db); err != nil {
		return err
	}
	if err := seedMenus(db); err != nil {
		return err
	}
	if err := seedConfigs(db); err != nil {
		return err
	}
	return seedDicts(db)
}

func seedConfigs(db *gorm.DB) error {
	var count int64
	db.Model(&model.SysConfig{}).Count(&count)
	if count > 0 {
		return nil
	}
	configs := []model.SysConfig{
		{Name: "系统名称", Key: "sys.name", Value: "go-framework 管理后台", Builtin: true, Remark: "后台标题"},
		{Name: "初始密码", Key: "sys.user.initPassword", Value: "123456", Builtin: true, Remark: "用户重置后的默认密码"},
	}
	return db.Create(&configs).Error
}

func seedDicts(db *gorm.DB) error {
	var count int64
	db.Model(&model.SysDict{}).Count(&count)
	if count > 0 {
		return nil
	}
	// 通用状态字典，供各管理页启用/停用下拉使用。
	dict := model.SysDict{Name: "通用状态", Type: "sys_common_status", Status: model.StatusEnabled, Remark: "启用/停用"}
	if err := db.Create(&dict).Error; err != nil {
		return err
	}
	items := []model.SysDictItem{
		{DictType: "sys_common_status", Label: "启用", Value: "1", Sort: 1, ListClass: "success", Status: model.StatusEnabled},
		{DictType: "sys_common_status", Label: "停用", Value: "2", Sort: 2, ListClass: "danger", Status: model.StatusEnabled},
	}
	return db.Create(&items).Error
}

func seedDept(db *gorm.DB) error {
	var count int64
	db.Model(&model.SysDept{}).Count(&count)
	if count > 0 {
		return nil
	}
	return db.Create(&model.SysDept{ParentID: 0, Name: "总公司", Sort: 1, Status: model.StatusEnabled}).Error
}

func seedRoles(db *gorm.DB) error {
	var count int64
	db.Model(&model.SysRole{}).Count(&count)
	if count > 0 {
		return nil
	}
	roles := []model.SysRole{
		{Name: "超级管理员", Key: model.AdminRoleKey, Sort: 1, Status: model.StatusEnabled, Remark: "拥有全部权限"},
		{Name: "普通用户", Key: "common", Sort: 2, Status: model.StatusEnabled, Remark: "默认角色，权限由管理员分配"},
	}
	return db.Create(&roles).Error
}

func seedAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&model.SysUser{}).Count(&count)
	if count > 0 {
		return nil
	}
	hash, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}
	var adminRole model.SysRole
	if err := db.Where("key = ?", model.AdminRoleKey).First(&adminRole).Error; err != nil {
		return err
	}
	var dept model.SysDept
	_ = db.Order("id").First(&dept).Error
	admin := model.SysUser{
		Username: model.AdminUsername,
		Password: hash,
		Nickname: "超级管理员",
		Status:   model.StatusEnabled,
		DeptID:   dept.ID,
		Roles:    []model.SysRole{adminRole},
	}
	return db.Create(&admin).Error
}

// seedMenus 默认菜单：系统管理目录 + 用户/角色/菜单/部门管理页面及其按钮权限。
func seedMenus(db *gorm.DB) error {
	var count int64
	db.Model(&model.SysMenu{}).Count(&count)
	if count > 0 {
		return nil
	}

	sysDir := model.SysMenu{ParentID: 0, Title: "系统管理", Name: "System", Type: model.MenuTypeDir,
		Path: "/system", Icon: "setting", Sort: 1, Visible: 1, Status: model.StatusEnabled}
	if err := db.Create(&sysDir).Error; err != nil {
		return err
	}

	type page struct {
		title, name, path, component, permPrefix string
		buttons                                  []string
	}
	pages := []page{
		{"用户管理", "SystemUser", "user", "system/user/index", "system:user",
			[]string{"query", "add", "edit", "del", "resetPwd", "status"}},
		{"角色管理", "SystemRole", "role", "system/role/index", "system:role",
			[]string{"query", "add", "edit", "del", "perm"}},
		{"菜单管理", "SystemMenu", "menu", "system/menu/index", "system:menu",
			[]string{"query", "add", "edit", "del"}},
		{"部门管理", "SystemDept", "dept", "system/dept/index", "system:dept",
			[]string{"query", "add", "edit", "del"}},
		{"字典管理", "SystemDict", "dict", "system/dict/index", "system:dict",
			[]string{"query", "add", "edit", "del"}},
		{"参数配置", "SystemConfig", "config", "system/config/index", "system:config",
			[]string{"query", "add", "edit", "del"}},
		{"定时任务", "SystemJob", "job", "system/job/index", "system:job",
			[]string{"query", "add", "edit", "del", "status", "run"}},
		{"操作日志", "SystemOperLog", "oper-log", "system/log/oper", "system:operlog",
			[]string{"query", "del"}},
		{"登录日志", "SystemLoginLog", "login-log", "system/log/login", "system:loginlog",
			[]string{"query", "del"}},
		{"文件管理", "SystemFile", "file", "system/file/index", "system:file",
			[]string{"query", "upload", "del"}},
		{"服务监控", "SystemMonitor", "monitor", "system/monitor/index", "system:monitor",
			[]string{"query"}},
	}
	buttonTitles := map[string]string{
		"query": "查询", "add": "新增", "edit": "修改", "del": "删除",
		"resetPwd": "重置密码", "status": "启用停用", "perm": "分配权限",
		"run": "执行", "upload": "上传",
	}

	for i, p := range pages {
		menu := model.SysMenu{
			ParentID: sysDir.ID, Title: p.title, Name: p.name, Type: model.MenuTypeMenu,
			Path: p.path, Component: p.component, Perm: p.permPrefix + ":list",
			Sort: i + 1, Visible: 1, Status: model.StatusEnabled,
		}
		if err := db.Create(&menu).Error; err != nil {
			return err
		}
		for j, b := range p.buttons {
			btn := model.SysMenu{
				ParentID: menu.ID, Title: buttonTitles[b], Type: model.MenuTypeButton,
				Perm: p.permPrefix + ":" + b, Sort: j + 1, Visible: 1, Status: model.StatusEnabled,
			}
			if err := db.Create(&btn).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
