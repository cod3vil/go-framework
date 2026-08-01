// Package database 封装 GORM 初始化与事务助手，支持 sqlite / mysql / postgres。
// sqlite 使用纯 Go 驱动（glebarez/sqlite），保证 CGO_ENABLED=0 下可编译单二进制。
package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// New 根据配置初始化数据库连接。
func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dialector, err := openDialector(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(parseLogLevel(cfg.LogLevel)),
		// 关联完整性由应用层保证，不生成数据库外键，
		// 避免 0 值关联与删除顺序问题，也便于分库分表演进。
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接检测失败: %w", err)
	}
	return db, nil
}

func openDialector(cfg config.DatabaseConfig) (gorm.Dialector, error) {
	switch cfg.Driver {
	case "sqlite":
		if dir := filepath.Dir(cfg.DSN); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("创建 sqlite 数据目录失败: %w", err)
			}
		}
		return sqlite.Open(cfg.DSN), nil
	case "mysql":
		return mysql.Open(cfg.DSN), nil
	case "postgres":
		return postgres.Open(cfg.DSN), nil
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s（可选 sqlite/mysql/postgres）", cfg.Driver)
	}
}

func parseLogLevel(level string) gormlogger.LogLevel {
	switch level {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}

// Tx 在事务中执行 fn，fn 返回错误时回滚。service 层用它控制事务边界：
//
//	err := database.Tx(ctx, db, func(tx *gorm.DB) error { ... })
func Tx(ctx context.Context, db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(fn)
}
