// Package errs 定义业务错误类型与分段错误码。
//
// 错误码分段约定：
//   - 0        成功
//   - 1xxx     通用错误
//   - 2xxx     系统管理模块
//   - 10xxx 起 业务模块各自分配一段
package errs

import (
	"errors"
	"fmt"
)

// 通用错误码。
const (
	CodeOK           = 0
	CodeBadRequest   = 1000 // 参数错误
	CodeUnauthorized = 1001 // 未认证或凭证失效
	CodeForbidden    = 1003 // 无权限
	CodeNotFound     = 1004 // 资源不存在
	CodeConflict     = 1005 // 资源冲突（如唯一键重复）
	CodeTooManyReq   = 1006 // 请求过于频繁
	CodeInternal     = 1500 // 服务内部错误
)

// Error 业务错误，携带对外错误码与提示信息。
type Error struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	// cause 内部原因，仅记录日志，不返回给客户端。
	cause error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Msg, e.cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// Unwrap 支持 errors.Is / errors.As 链式匹配。
func (e *Error) Unwrap() error { return e.cause }

// WithCause 附加内部原因，返回新错误（不修改原错误，预定义错误可安全复用）。
func (e *Error) WithCause(cause error) *Error {
	return &Error{Code: e.Code, Msg: e.Msg, cause: cause}
}

// New 创建业务错误。
func New(code int, msg string) *Error {
	return &Error{Code: code, Msg: msg}
}

// Newf 创建带格式化消息的业务错误。
func Newf(code int, format string, args ...any) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// From 从任意 error 中提取业务错误；非业务错误返回 nil, false。
func From(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// 预定义通用错误，handler/service 可直接使用或通过 WithCause 附加细节。
var (
	ErrBadRequest   = New(CodeBadRequest, "请求参数错误")
	ErrUnauthorized = New(CodeUnauthorized, "未登录或登录已过期")
	ErrForbidden    = New(CodeForbidden, "没有操作权限")
	ErrNotFound     = New(CodeNotFound, "资源不存在")
	ErrConflict     = New(CodeConflict, "资源已存在")
	ErrTooManyReq   = New(CodeTooManyReq, "请求过于频繁，请稍后重试")
	ErrInternal     = New(CodeInternal, "服务开小差了，请稍后重试")
)
