// Package handler 系统管理模块的 HTTP 处理层：解析参数、调 service、写响应。
package handler

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler 系统模块处理器。
type Handler struct {
	svc *service.Service
}

// New 创建处理器。
func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// getPage 解析分页参数，page 默认 1，pageSize 默认 10、上限 200。
func getPage(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}

// paramID 解析路径中的 :id。
func paramID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "无效的 ID")
		return 0, false
	}
	return uint(id), true
}

// bindJSON 绑定并校验 JSON 请求体，失败时直接写参数错误响应。
func bindJSON(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return false
	}
	return true
}
