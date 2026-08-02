package handler

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListRoles 分页查询角色。
// GET /api/v1/system/roles
func (h *Handler) ListRoles(c *gin.Context) {
	page, pageSize := getPage(c)
	status, _ := strconv.Atoi(c.Query("status"))
	roles, total, err := h.svc.ListRoles(c.Request.Context(), service.RoleQuery{
		Page: page, PageSize: pageSize,
		Name: c.Query("name"), Status: int8(status),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, roles, total, page, pageSize)
}

type roleReq struct {
	Name      string `json:"name" binding:"required,max=64"`
	Key       string `json:"key" binding:"required,max=64"`
	Sort      int    `json:"sort"`
	Status    int8   `json:"status" binding:"omitempty,oneof=1 2"`
	Remark    string `json:"remark" binding:"max=255"`
	DataScope int8   `json:"dataScope" binding:"omitempty,oneof=1 2 3 4 5"`
	DeptIDs   []uint `json:"deptIds"`
}

func (r roleReq) toInput() service.RoleInput {
	status := r.Status
	if status == 0 {
		status = 1
	}
	return service.RoleInput{
		Name: r.Name, Key: r.Key, Sort: r.Sort, Status: status, Remark: r.Remark,
		DataScope: r.DataScope, DeptIDs: r.DeptIDs,
	}
}

// CreateRole 创建角色。
// POST /api/v1/system/roles
func (h *Handler) CreateRole(c *gin.Context) {
	var req roleReq
	if !bindJSON(c, &req) {
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, role)
}

// GetRole 角色详情。
// GET /api/v1/system/roles/:id
func (h *Handler) GetRole(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	role, err := h.svc.GetRole(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, role)
}

// UpdateRole 更新角色。
// PUT /api/v1/system/roles/:id
func (h *Handler) UpdateRole(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req roleReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateRole(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteRole 删除角色。
// DELETE /api/v1/system/roles/:id
func (h *Handler) DeleteRole(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// GetRoleDepts 查询角色自定义数据权限的部门 ID。
// GET /api/v1/system/roles/:id/depts
func (h *Handler) GetRoleDepts(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	ids, err := h.svc.GetRoleDeptIDs(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, ids)
}

// GetRoleMenus 查询角色绑定的菜单 ID。
// GET /api/v1/system/roles/:id/menus
func (h *Handler) GetRoleMenus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	ids, err := h.svc.GetRoleMenuIDs(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, ids)
}

type roleMenusReq struct {
	MenuIDs []uint `json:"menuIds"`
}

// SetRoleMenus 绑定角色菜单。
// PUT /api/v1/system/roles/:id/menus
func (h *Handler) SetRoleMenus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req roleMenusReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.SetRoleMenus(c.Request.Context(), id, req.MenuIDs); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// GetRoleAPIs 查询角色的 API 权限。
// GET /api/v1/system/roles/:id/apis
func (h *Handler) GetRoleAPIs(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	apis, err := h.svc.GetRoleAPIs(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, apis)
}

type roleAPIsReq struct {
	APIs []service.APIPermission `json:"apis"`
}

// SetRoleAPIs 设置角色的 API 权限（Casbin 策略，实时生效）。
// PUT /api/v1/system/roles/:id/apis
func (h *Handler) SetRoleAPIs(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req roleAPIsReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.SetRoleAPIs(c.Request.Context(), id, req.APIs); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
