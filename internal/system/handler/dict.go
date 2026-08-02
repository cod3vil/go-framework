package handler

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListDicts 分页查询字典类型。
// GET /api/v1/system/dicts
func (h *Handler) ListDicts(c *gin.Context) {
	page, pageSize := getPage(c)
	status, _ := strconv.Atoi(c.Query("status"))
	dicts, total, err := h.svc.ListDicts(c.Request.Context(), service.DictQuery{
		Page: page, PageSize: pageSize,
		Name: c.Query("name"), Type: c.Query("type"), Status: int8(status),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, dicts, total, page, pageSize)
}

type dictReq struct {
	Name   string `json:"name" binding:"required,max=64"`
	Type   string `json:"type" binding:"required,max=64"`
	Status int8   `json:"status" binding:"omitempty,oneof=1 2"`
	Remark string `json:"remark" binding:"max=255"`
}

func (r dictReq) toInput() service.DictInput {
	status := r.Status
	if status == 0 {
		status = 1
	}
	return service.DictInput{Name: r.Name, Type: r.Type, Status: status, Remark: r.Remark}
}

// CreateDict 创建字典类型。
// POST /api/v1/system/dicts
func (h *Handler) CreateDict(c *gin.Context) {
	var req dictReq
	if !bindJSON(c, &req) {
		return
	}
	dict, err := h.svc.CreateDict(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dict)
}

// UpdateDict 更新字典类型。
// PUT /api/v1/system/dicts/:id
func (h *Handler) UpdateDict(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req dictReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateDict(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteDict 删除字典类型及其字典项。
// DELETE /api/v1/system/dicts/:id
func (h *Handler) DeleteDict(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteDict(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListDictItems 查询字典项。管理页传 all=1 含停用项，前端下拉默认只取启用项。
// GET /api/v1/system/dicts/:type/items
func (h *Handler) ListDictItems(c *gin.Context) {
	dictType := c.Param("type")
	ctx := c.Request.Context()
	var (
		items any
		err   error
	)
	if c.Query("all") == "1" {
		items, err = h.svc.AllDictItems(ctx, dictType)
	} else {
		items, err = h.svc.ListDictItems(ctx, dictType)
	}
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

type dictItemReq struct {
	DictType  string `json:"dictType" binding:"required,max=64"`
	Label     string `json:"label" binding:"required,max=64"`
	Value     string `json:"value" binding:"required,max=64"`
	Sort      int    `json:"sort"`
	CSSClass  string `json:"cssClass" binding:"max=64"`
	ListClass string `json:"listClass" binding:"max=64"`
	IsDefault bool   `json:"isDefault"`
	Status    int8   `json:"status" binding:"omitempty,oneof=1 2"`
	Remark    string `json:"remark" binding:"max=255"`
}

func (r dictItemReq) toInput() service.DictItemInput {
	status := r.Status
	if status == 0 {
		status = 1
	}
	return service.DictItemInput{
		DictType: r.DictType, Label: r.Label, Value: r.Value, Sort: r.Sort,
		CSSClass: r.CSSClass, ListClass: r.ListClass, IsDefault: r.IsDefault,
		Status: status, Remark: r.Remark,
	}
}

// CreateDictItem 创建字典项。
// POST /api/v1/system/dict-items
func (h *Handler) CreateDictItem(c *gin.Context) {
	var req dictItemReq
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.CreateDictItem(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

// UpdateDictItem 更新字典项。
// PUT /api/v1/system/dict-items/:id
func (h *Handler) UpdateDictItem(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req dictItemReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateDictItem(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteDictItem 删除字典项。
// DELETE /api/v1/system/dict-items/:id
func (h *Handler) DeleteDictItem(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteDictItem(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
