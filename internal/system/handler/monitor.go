package handler

import (
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ServerMonitor 返回服务运行监控指标。
// GET /api/v1/system/monitor/server
func (h *Handler) ServerMonitor(c *gin.Context) {
	response.OK(c, h.svc.ServerStat(c.Request.Context()))
}
