// 框架入口：基于 cobra 的 CLI。
//
//	server  启动 HTTP 服务（默认命令）
//	version 打印版本号
//
// P2 将新增 migrate（建表与种子数据）、create-admin（创建管理员）等子命令。
package main

import (
	"errors"
	"fmt"
	"os"

	goframework "github.com/cod3vil/go-framework"
	"github.com/cod3vil/go-framework/internal/app"
	"github.com/cod3vil/go-framework/internal/system"
	"github.com/cod3vil/go-framework/internal/system/model"
	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/database"
	"github.com/cod3vil/go-framework/pkg/utils"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var configPath string

func main() {
	root := &cobra.Command{
		Use:   "server",
		Short: "go-framework 企业级应用框架",
		// 不带子命令时默认启动服务。
		RunE: func(cmd *cobra.Command, args []string) error { return runServer() },
	}
	root.PersistentFlags().StringVarP(&configPath, "config", "c", "configs/config.yaml", "配置文件路径")

	adminCmd := &cobra.Command{
		Use:   "create-admin",
		Short: "创建或重置管理员账号密码",
		RunE:  func(cmd *cobra.Command, args []string) error { return runCreateAdmin(cmd) },
	}
	adminCmd.Flags().String("username", "admin", "管理员账号")
	adminCmd.Flags().String("password", "", "管理员密码（必填）")
	_ = adminCmd.MarkFlagRequired("password")

	root.AddCommand(
		&cobra.Command{
			Use:   "serve",
			Short: "启动 HTTP 服务",
			RunE:  func(cmd *cobra.Command, args []string) error { return runServer() },
		},
		&cobra.Command{
			Use:   "migrate",
			Short: "执行数据库迁移与种子数据（默认管理员 admin/admin123、默认角色与菜单）",
			RunE:  func(cmd *cobra.Command, args []string) error { return runMigrate() },
		},
		adminCmd,
		&cobra.Command{
			Use:   "version",
			Short: "打印版本号",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(app.Version)
			},
		},
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runServer() error {
	opts := []app.Option{}
	if adminFS, err := goframework.AdminFS(); err == nil {
		opts = append(opts, app.WithAdminFS(adminFS))
	}
	a, err := app.New(configPath, opts...)
	if err != nil {
		return err
	}
	return a.Run()
}

// openDB 供 migrate/create-admin 等离线命令使用，仅初始化数据库连接。
func openDB() (*gorm.DB, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	return database.New(cfg.Database)
}

func runMigrate() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	if err := system.Migrate(db); err != nil {
		return err
	}
	fmt.Println("迁移完成。默认管理员: admin / admin123（生产环境请立即用 create-admin 修改密码）")
	return nil
}

func runCreateAdmin(cmd *cobra.Command) error {
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	if len(password) < 6 {
		return fmt.Errorf("密码长度不能少于 6 位")
	}
	db, err := openDB()
	if err != nil {
		return err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	var user model.SysUser
	err = db.Where("username = ?", username).First(&user).Error
	switch {
	case err == nil:
		if err := db.Model(&user).Update("password", hash).Error; err != nil {
			return err
		}
		fmt.Printf("已重置账号 %s 的密码\n", username)
	case errors.Is(err, gorm.ErrRecordNotFound):
		var adminRole model.SysRole
		if err := db.Where("key = ?", model.AdminRoleKey).First(&adminRole).Error; err != nil {
			return fmt.Errorf("管理员角色不存在，请先执行 migrate: %w", err)
		}
		user = model.SysUser{
			Username: username, Password: hash, Nickname: username,
			Status: model.StatusEnabled, Roles: []model.SysRole{adminRole},
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
		fmt.Printf("已创建管理员账号 %s\n", username)
	default:
		return err
	}
	return nil
}
