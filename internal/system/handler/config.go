package handler

import (
	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListConfigs 分页查询参数。
// GET /api/v1/system/configs
func (h *Handler) ListConfigs(c *gin.Context) {
	page, pageSize := getPage(c)
	configs, total, err := h.svc.ListConfigs(c.Request.Context(), service.ConfigQuery{
		Page: page, PageSize: pageSize,
		Name: c.Query("name"), Key: c.Query("key"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, configs, total, page, pageSize)
}

// GetConfigByKey 按 key 读取参数值（公开给已登录用户，如前端读取系统标题）。
// GET /api/v1/system/configs/key/:key
func (h *Handler) GetConfigByKey(c *gin.Context) {
	val, err := h.svc.GetConfigValue(c.Request.Context(), c.Param("key"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"key": c.Param("key"), "value": val})
}

type configReq struct {
	Name   string `json:"name" binding:"required,max=64"`
	Key    string `json:"key" binding:"required,max=128"`
	Value  string `json:"value" binding:"required,max=1024"`
	Remark string `json:"remark" binding:"max=255"`
}

func (r configReq) toInput() service.ConfigInput {
	return service.ConfigInput{Name: r.Name, Key: r.Key, Value: r.Value, Remark: r.Remark}
}

// CreateConfig 创建参数。
// POST /api/v1/system/configs
func (h *Handler) CreateConfig(c *gin.Context) {
	var req configReq
	if !bindJSON(c, &req) {
		return
	}
	cfg, err := h.svc.CreateConfig(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cfg)
}

// UpdateConfig 更新参数。
// PUT /api/v1/system/configs/:id
func (h *Handler) UpdateConfig(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req configReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateConfig(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteConfig 删除参数。
// DELETE /api/v1/system/configs/:id
func (h *Handler) DeleteConfig(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteConfig(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
