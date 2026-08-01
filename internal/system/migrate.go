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
	return seedMenus(db)
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
	}
	buttonTitles := map[string]string{
		"query": "查询", "add": "新增", "edit": "修改", "del": "删除",
		"resetPwd": "重置密码", "status": "启用停用", "perm": "分配权限",
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
