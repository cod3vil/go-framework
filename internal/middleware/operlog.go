package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// OperLogEntry 一条操作日志，交由 recorder 落库，避免中间件耦合具体存储。
type OperLogEntry struct {
	Username  string
	UserID    uint
	Method    string
	Path      string
	Query     string
	Body      string
	IP        string
	Status    int
	Code      int
	LatencyMS int64
	CreatedAt time.Time
}

// OperLogRecorder 落库回调，实现方决定同步/异步写入。
type OperLogRecorder func(OperLogEntry)

const (
	maxBodyLog = 2048 // 请求体记录上限，超出截断
)

// 脱敏字段：请求体中匹配到这些键的值会被替换为 ***。
var sensitiveField = regexp.MustCompile(`("(?:password|newPassword|oldPassword|captchaCode)"\s*:\s*)"[^"]*"`)

// bodyCaptureWriter 包装 ResponseWriter 以捕获响应体（用于提取业务 code）。
type bodyCaptureWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w bodyCaptureWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

// OperLog 审计写操作（POST/PUT/DELETE/PATCH）。放在 Auth 之后以拿到用户身份。
func OperLog(record OperLogRecorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isWrite(c.Request.Method) {
			c.Next()
			return
		}
		start := time.Now()

		var reqBody string
		if c.Request.Body != nil {
			raw, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(raw)) // 复位供后续 handler 读取
			reqBody = desensitize(raw)
		}

		respBuf := &bytes.Buffer{}
		c.Writer = bodyCaptureWriter{ResponseWriter: c.Writer, buf: respBuf}

		c.Next()

		record(OperLogEntry{
			Username:  c.GetString(CtxUsername),
			UserID:    UserID(c),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Query:     c.Request.URL.RawQuery,
			Body:      reqBody,
			IP:        c.ClientIP(),
			Status:    c.Writer.Status(),
			Code:      extractCode(respBuf.Bytes()),
			LatencyMS: time.Since(start).Milliseconds(),
			CreatedAt: start,
		})
	}
}

func isWrite(method string) bool {
	switch method {
	case "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}

func desensitize(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	s := sensitiveField.ReplaceAllString(string(raw), `${1}"***"`)
	if len(s) > maxBodyLog {
		return s[:maxBodyLog] + "...(truncated)"
	}
	return s
}

// extractCode 从统一响应体中解析业务 code 字段。
func extractCode(resp []byte) int {
	if len(resp) == 0 {
		return 0
	}
	var body struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(resp, &body); err != nil {
		return 0
	}
	return body.Code
}
