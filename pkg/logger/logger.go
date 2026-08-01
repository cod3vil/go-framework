// Package logger 基于 Zap 封装结构化日志，支持 lumberjack 文件滚动。
package logger

import (
	"os"

	"github.com/cod3vil/go-framework/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New 根据配置构建 *zap.Logger。Filename 为空时仅输出到 stdout，
// 配置了文件时同时输出到 stdout 与滚动文件。
func New(cfg config.LogConfig) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.TimeKey = "time"

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encCfg)
	} else {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level),
	}
	if cfg.Filename != "" {
		fileEncCfg := zap.NewProductionEncoderConfig()
		fileEncCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncCfg.TimeKey = "time"
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
		})
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(fileEncCfg), fileWriter, level))
	}

	return zap.New(zapcore.NewTee(cores...), zap.AddCaller()), nil
}
