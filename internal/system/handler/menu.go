package handler

import (
	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// MenuTree 全量菜单树。
// GET /api/v1/system/menus/tree
func (h *Handler) MenuTree(c *gin.Context) {
	tree, err := h.svc.MenuTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tree)
}

type menuReq struct {
	ParentID  uint   `json:"parentId"`
	Title     string `json:"title" binding:"required,max=64"`
	Name      string `json:"name" binding:"max=64"`
	Type      int8   `json:"type" binding:"required,oneof=1 2 3"`
	Path      string `json:"path" binding:"max=255"`
	Component string `json:"component" binding:"max=255"`
	Perm      string `json:"perm" binding:"max=128"`
	Icon      string `json:"icon" binding:"max=64"`
	Sort      int    `json:"sort"`
	Visible   int8   `json:"visible" binding:"omitempty,oneof=1 2"`
	Status    int8   `json:"status" binding:"omitempty,oneof=1 2"`
	KeepAlive bool   `json:"keepAlive"`
}

func (r menuReq) toInput() service.MenuInput {
	visible, status := r.Visible, r.Status
	if visible == 0 {
		visible = 1
	}
	if status == 0 {
		status = 1
	}
	return service.MenuInput{
		ParentID: r.ParentID, Title: r.Title, Name: r.Name, Type: r.Type,
		Path: r.Path, Component: r.Component, Perm: r.Perm, Icon: r.Icon,
		Sort: r.Sort, Visible: visible, Status: status, KeepAlive: r.KeepAlive,
	}
}

// CreateMenu 创建菜单。
// POST /api/v1/system/menus
func (h *Handler) CreateMenu(c *gin.Context) {
	var req menuReq
	if !bindJSON(c, &req) {
		return
	}
	menu, err := h.svc.CreateMenu(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, menu)
}

// UpdateMenu 更新菜单。
// PUT /api/v1/system/menus/:id
func (h *Handler) UpdateMenu(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req menuReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateMenu(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteMenu 删除菜单。
// DELETE /api/v1/system/menus/:id
func (h *Handler) DeleteMenu(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteMenu(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
