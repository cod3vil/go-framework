package handler

import (
	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// DeptTree 部门树。
// GET /api/v1/system/depts/tree
func (h *Handler) DeptTree(c *gin.Context) {
	tree, err := h.svc.DeptTree(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tree)
}

type deptReq struct {
	ParentID uint   `json:"parentId"`
	Name     string `json:"name" binding:"required,max=64"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader" binding:"max=64"`
	Phone    string `json:"phone" binding:"max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Status   int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

func (r deptReq) toInput() service.DeptInput {
	status := r.Status
	if status == 0 {
		status = 1
	}
	return service.DeptInput{
		ParentID: r.ParentID, Name: r.Name, Sort: r.Sort,
		Leader: r.Leader, Phone: r.Phone, Email: r.Email, Status: status,
	}
}

// CreateDept 创建部门。
// POST /api/v1/system/depts
func (h *Handler) CreateDept(c *gin.Context) {
	var req deptReq
	if !bindJSON(c, &req) {
		return
	}
	dept, err := h.svc.CreateDept(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dept)
}

// UpdateDept 更新部门。
// PUT /api/v1/system/depts/:id
func (h *Handler) UpdateDept(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req deptReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateDept(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteDept 删除部门。
// DELETE /api/v1/system/depts/:id
func (h *Handler) DeleteDept(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteDept(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
