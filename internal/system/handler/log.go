package handler

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListOperLogs 分页查询操作日志。
// GET /api/v1/system/oper-logs
func (h *Handler) ListOperLogs(c *gin.Context) {
	page, pageSize := getPage(c)
	logs, total, err := h.svc.ListOperLogs(c.Request.Context(), service.OperLogQuery{
		Page: page, PageSize: pageSize,
		Username: c.Query("username"), Path: c.Query("path"), Method: c.Query("method"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, logs, total, page, pageSize)
}

// ClearOperLogs 清空操作日志。
// DELETE /api/v1/system/oper-logs
func (h *Handler) ClearOperLogs(c *gin.Context) {
	if err := h.svc.ClearOperLogs(c.Request.Context()); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListLoginLogs 分页查询登录日志。
// GET /api/v1/system/login-logs
func (h *Handler) ListLoginLogs(c *gin.Context) {
	page, pageSize := getPage(c)
	status, _ := strconv.Atoi(c.Query("status"))
	logs, total, err := h.svc.ListLoginLogs(c.Request.Context(), service.LoginLogQuery{
		Page: page, PageSize: pageSize,
		Username: c.Query("username"), Status: int8(status),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, logs, total, page, pageSize)
}

// ClearLoginLogs 清空登录日志。
// DELETE /api/v1/system/login-logs
func (h *Handler) ClearLoginLogs(c *gin.Context) {
	if err := h.svc.ClearLoginLogs(c.Request.Context()); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
