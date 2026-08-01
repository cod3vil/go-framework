// Package response 提供统一的 API 响应格式。
//
// 所有接口统一返回 {"code":0,"msg":"ok","data":...}，
// 分页数据统一为 {"list":[],"total":0,"page":1,"pageSize":10}。
package response

import (
	"net/http"

	"github.com/cod3vil/go-framework/pkg/errs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Body 统一响应结构。
type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// Page 统一分页结构。
type Page struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

// OK 返回成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: errs.CodeOK, Msg: "ok", Data: data})
}

// OKPage 返回分页成功响应。
func OKPage(c *gin.Context, list any, total int64, page, pageSize int) {
	OK(c, Page{List: list, Total: total, Page: page, PageSize: pageSize})
}

// Error 返回错误响应：业务错误按其错误码返回，
// 未知错误统一按 500 返回且不向客户端泄漏内部信息。
func Error(c *gin.Context, err error) {
	if e, ok := errs.From(err); ok {
		c.JSON(http.StatusOK, Body{Code: e.Code, Msg: e.Msg})
		return
	}
	if logger, exists := c.Get("logger"); exists {
		if l, ok := logger.(*zap.Logger); ok {
			l.Error("unhandled error", zap.String("path", c.FullPath()), zap.Error(err))
		}
	}
	c.JSON(http.StatusInternalServerError, Body{Code: errs.CodeInternal, Msg: errs.ErrInternal.Msg})
}

// BadRequest 返回参数错误响应，msg 为空时使用默认提示。
func BadRequest(c *gin.Context, msg string) {
	if msg == "" {
		msg = errs.ErrBadRequest.Msg
	}
	c.JSON(http.StatusOK, Body{Code: errs.CodeBadRequest, Msg: msg})
}
