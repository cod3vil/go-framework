package article

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/modkit"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// Handler 文章 HTTP 处理层。
type Handler struct {
	svc *Service
	kit *modkit.Kit
}

// NewHandler 创建处理器。kit 用于解析当前用户的数据范围。
func NewHandler(svc *Service, kit *modkit.Kit) *Handler { return &Handler{svc: svc, kit: kit} }

// List 分页查询文章。
// GET /api/v1/articles
func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 10
	}
	status, _ := strconv.Atoi(c.Query("status"))
	scope, err := h.kit.ScopeOf(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), Query{
		Page: page, PageSize: pageSize, Title: c.Query("title"), Status: int8(status),
	}, scope)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, list, total, page, pageSize)
}

// Get 文章详情。
// GET /api/v1/articles/:id
func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}
	a, err := h.svc.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, a)
}

type articleReq struct {
	Title   string `json:"title" binding:"required,max=200"`
	Author  string `json:"author" binding:"max=64"`
	Content string `json:"content"`
	Status  int8   `json:"status" binding:"omitempty,oneof=1 2"`
}

func (r articleReq) toInput() Input {
	status := r.Status
	if status == 0 {
		status = StatusDraft
	}
	return Input{Title: r.Title, Author: r.Author, Content: r.Content, Status: status}
}

// Create 新建文章。
// POST /api/v1/articles
func (h *Handler) Create(c *gin.Context) {
	var req articleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	a, err := h.svc.Create(c.Request.Context(), req.toInput(), middleware.UserID(c), middleware.DeptID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, a)
}

// Update 更新文章。
// PUT /api/v1/articles/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}
	var req articleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	if err := h.svc.Update(c.Request.Context(), uint(id), req.toInput()); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// Delete 删除文章。
// DELETE /api/v1/articles/:id
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
