// Package config 基于 Viper 提供 YAML 配置加载，支持环境变量覆盖。
// 环境变量前缀为 APP，层级用下划线连接，如 APP_SERVER_PORT=9000 覆盖 server.port。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用总配置。
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// AppConfig 应用基本信息。
type AppConfig struct {
	Name string `mapstructure:"name"`
	// Mode 运行模式: debug / release / test，对应 gin 的模式。
	Mode string `mapstructure:"mode"`
	// CaptchaEnabled 登录是否启用图形验证码。
	CaptchaEnabled bool `mapstructure:"captcha_enabled"`
}

// ServerConfig HTTP 服务配置。
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	// ReadTimeout / WriteTimeout 单位秒。
	ReadTimeout  int `mapstructure:"read_timeout"`
	WriteTimeout int `mapstructure:"write_timeout"`
	// ShutdownTimeout 优雅关闭最长等待秒数。
	ShutdownTimeout int `mapstructure:"shutdown_timeout"`
	// RateLimit 每秒允许的请求数，0 表示不限流。
	RateLimit int `mapstructure:"rate_limit"`
	// RateBurst 令牌桶突发容量。
	RateBurst int `mapstructure:"rate_burst"`
}

// Addr 返回监听地址。
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// LogConfig 日志配置。
type LogConfig struct {
	// Level: debug / info / warn / error。
	Level string `mapstructure:"level"`
	// Format: console / json。
	Format string `mapstructure:"format"`
	// Filename 日志文件路径，为空时仅输出到 stdout。
	Filename   string `mapstructure:"filename"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

// DatabaseConfig 数据库配置。
type DatabaseConfig struct {
	// Driver: sqlite / mysql / postgres。
	Driver string `mapstructure:"driver"`
	// DSN 连接串；sqlite 时为文件路径。
	DSN             string `mapstructure:"dsn"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // 秒
	// LogLevel: silent / error / warn / info。
	LogLevel string `mapstructure:"log_level"`
}

// RedisConfig Redis 配置，Addr 为空时框架降级为内存缓存。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT 配置（P2 使用，先占位统一管理）。
type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	// AccessExpire / RefreshExpire 单位分钟。
	AccessExpire  int    `mapstructure:"access_expire"`
	RefreshExpire int    `mapstructure:"refresh_expire"`
	Issuer        string `mapstructure:"issuer"`
}

// AccessTTL Access Token 有效期。
func (j JWTConfig) AccessTTL() time.Duration { return time.Duration(j.AccessExpire) * time.Minute }

// RefreshTTL Refresh Token 有效期。
func (j JWTConfig) RefreshTTL() time.Duration { return time.Duration(j.RefreshExpire) * time.Minute }

// Load 从指定文件加载配置并应用环境变量覆盖。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}
	cfg := new(Config)
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "go-framework")
	v.SetDefault("app.mode", "debug")
	v.SetDefault("app.captcha_enabled", true)

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 60)
	v.SetDefault("server.write_timeout", 60)
	v.SetDefault("server.shutdown_timeout", 15)
	v.SetDefault("server.rate_limit", 0)
	v.SetDefault("server.rate_burst", 100)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.max_size_mb", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age_days", 30)

	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.dsn", "host=127.0.0.1 user=postgres password=postgres dbname=app port=5432 sslmode=disable")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 3600)
	v.SetDefault("database.log_level", "warn")

	v.SetDefault("jwt.access_expire", 120)
	v.SetDefault("jwt.refresh_expire", 10080)
	v.SetDefault("jwt.issuer", "go-framework")
}
