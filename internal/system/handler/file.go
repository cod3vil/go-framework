package handler

import (
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// UploadFile 上传文件（multipart/form-data，字段名 file）。
// POST /api/v1/system/files
func (h *Handler) UploadFile(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件（表单字段名 file）")
		return
	}
	file, err := h.svc.UploadFile(c.Request.Context(), fh, middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, file)
}

// ListFiles 分页查询文件。
// GET /api/v1/system/files
func (h *Handler) ListFiles(c *gin.Context) {
	page, pageSize := getPage(c)
	files, total, err := h.svc.ListFiles(c.Request.Context(), service.FileQuery{
		Page: page, PageSize: pageSize, Name: c.Query("name"), Ext: c.Query("ext"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, files, total, page, pageSize)
}

// DownloadFile 下载文件（以原始文件名作为附件名）。
// GET /api/v1/system/files/:id/download
func (h *Handler) DownloadFile(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	file, err := h.svc.GetFile(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	path, ok := h.svc.Uploader.Storage().LocalPath(file.Key)
	if !ok {
		// 非本地存储：重定向到对象 URL。
		c.Redirect(http.StatusFound, file.URL)
		return
	}
	c.Header("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(filepath.Base(file.Name)))
	c.File(path)
}

// DeleteFile 删除文件。
// DELETE /api/v1/system/files/:id
func (h *Handler) DeleteFile(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteFile(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
