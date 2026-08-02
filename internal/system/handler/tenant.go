package handler

import (
	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListTenants 分页查询租户。
// GET /api/v1/system/tenants
func (h *Handler) ListTenants(c *gin.Context) {
	page, pageSize := getPage(c)
	list, total, err := h.svc.ListTenants(c.Request.Context(), page, pageSize, c.Query("name"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, list, total, page, pageSize)
}

type tenantReq struct {
	Code    string `json:"code" binding:"required,max=30"`
	Name    string `json:"name" binding:"required,max=128"`
	Contact string `json:"contact" binding:"max=64"`
	Remark  string `json:"remark" binding:"max=255"`
}

// CreateTenant 开通租户（建 schema + 迁移 + 种子）。
// POST /api/v1/system/tenants
func (h *Handler) CreateTenant(c *gin.Context) {
	var req tenantReq
	if !bindJSON(c, &req) {
		return
	}
	t, err := h.svc.CreateTenant(c.Request.Context(), service.TenantInput{
		Code: req.Code, Name: req.Name, Contact: req.Contact, Remark: req.Remark,
	}, middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

// SetTenantStatus 启用/停用租户。
// PUT /api/v1/system/tenants/:id/status
func (h *Handler) SetTenantStatus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req statusReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.SetTenantStatus(c.Request.Context(), id, req.Status, middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteTenant 删除租户并销毁其 schema。
// DELETE /api/v1/system/tenants/:id
func (h *Handler) DeleteTenant(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteTenant(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// TenantEnabled 返回多租户是否启用（供前端登录页决定是否显示租户输入）。
// GET /api/v1/tenant-enabled
func (h *Handler) TenantEnabled(c *gin.Context) {
	response.OK(c, gin.H{"enabled": h.svc.TenantEnabled()})
}
