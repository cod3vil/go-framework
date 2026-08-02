package handler

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListUsers 分页查询用户。
// GET /api/v1/system/users
func (h *Handler) ListUsers(c *gin.Context) {
	page, pageSize := getPage(c)
	status, _ := strconv.Atoi(c.Query("status"))
	deptID, _ := strconv.ParseUint(c.Query("deptId"), 10, 64)
	users, total, err := h.svc.ListUsers(c.Request.Context(), service.UserQuery{
		Page:     page,
		PageSize: pageSize,
		Username: c.Query("username"),
		Nickname: c.Query("nickname"),
		Status:   int8(status),
		DeptID:   uint(deptID),
		Operator: middleware.UserID(c),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, users, total, page, pageSize)
}

type userReq struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password"`
	Nickname string `json:"nickname" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"max=32"`
	Status   int8   `json:"status" binding:"omitempty,oneof=1 2"`
	DeptID   uint   `json:"deptId"`
	Remark   string `json:"remark" binding:"max=255"`
	RoleIDs  []uint `json:"roleIds"`
}

func (r userReq) toInput() service.UserInput {
	status := r.Status
	if status == 0 {
		status = 1
	}
	return service.UserInput{
		Username: r.Username, Password: r.Password, Nickname: r.Nickname,
		Email: r.Email, Phone: r.Phone, Status: status, DeptID: r.DeptID,
		Remark: r.Remark, RoleIDs: r.RoleIDs,
	}
}

// CreateUser 创建用户。
// POST /api/v1/system/users
func (h *Handler) CreateUser(c *gin.Context) {
	var req userReq
	if !bindJSON(c, &req) {
		return
	}
	if len(req.Password) < 6 {
		response.BadRequest(c, "密码长度不能少于 6 位")
		return
	}
	user, err := h.svc.CreateUser(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

// GetUser 用户详情。
// GET /api/v1/system/users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	user, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

// UpdateUser 更新用户。
// PUT /api/v1/system/users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req userReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateUser(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteUser 删除用户。
// DELETE /api/v1/system/users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteUser(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

type resetPwdReq struct {
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// ResetUserPassword 重置用户密码。
// PUT /api/v1/system/users/:id/password
func (h *Handler) ResetUserPassword(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req resetPwdReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), id, req.Password, middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

type statusReq struct {
	Status int8 `json:"status" binding:"required,oneof=1 2"`
}

// SetUserStatus 启用/停用用户。
// PUT /api/v1/system/users/:id/status
func (h *Handler) SetUserStatus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req statusReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.SetUserStatus(c.Request.Context(), id, req.Status, middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
