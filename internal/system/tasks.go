package system

import (
	"context"
	"time"

	"github.com/cod3vil/go-framework/pkg/cronx"
	"go.uber.org/zap"
)

// registerBuiltinTasks 注册框架内置的定时任务处理器。
// 业务方可在自己的模块中调用 manager.Register 注册更多任务。
func registerBuiltinTasks(manager *cronx.Manager, logger *zap.Logger) {
	// 心跳任务：用于演示与验证调度链路可用。
	manager.Register("system:heartbeat", func(_ context.Context) error {
		logger.Info("定时任务心跳", zap.Time("at", time.Now()))
		return nil
	})
}
