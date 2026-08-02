package app

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// registerOpenAPI 挂载 API 文档：
//   - GET /swagger           Swagger UI 页面（UI 资源引用官方 CDN，规范数据来自本服务）
//   - GET /swagger/doc.json  由已注册路由运行时生成的 OpenAPI 3.0 规范
//
// 采用运行时生成而非注释代码生成，保证文档与实际路由始终一致、零维护成本。
// 代价是自动生成的文档不含请求体字段级 schema（如需精细化可后续引入 swaggo 注释）。
func (a *App) registerOpenAPI() {
	a.Engine.GET("/swagger/doc.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, a.buildOpenAPI())
	})
	a.Engine.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
	a.Engine.GET("/swagger/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
}

// buildOpenAPI 根据 gin 路由表构造 OpenAPI 3.0 文档。
// 路径参数 :id 转为 {id}，并按第 3 段路径归类 tag（如 system/auth）。
func (a *App) buildOpenAPI() map[string]any {
	type operation map[string]any
	paths := map[string]map[string]any{}

	bearer := []map[string][]string{{"bearerAuth": {}}}

	for _, r := range a.Engine.Routes() {
		if !strings.HasPrefix(r.Path, "/api/") {
			continue
		}
		oaPath := ginPathToOpenAPI(r.Path)
		tag := routeTag(r.Path)

		op := operation{
			"tags":    []string{tag},
			"summary": r.Method + " " + r.Path,
			"responses": map[string]any{
				"200": map[string]any{"description": "统一响应 {code,msg,data}"},
			},
		}
		// 非登录/验证码接口标注需要 Bearer 鉴权。
		if !strings.HasPrefix(r.Path, "/api/v1/auth/login") &&
			!strings.HasPrefix(r.Path, "/api/v1/auth/captcha") &&
			!strings.HasPrefix(r.Path, "/api/v1/auth/refresh") &&
			!strings.HasPrefix(r.Path, "/api/v1/health") &&
			!strings.HasPrefix(r.Path, "/api/v1/ping") {
			op["security"] = bearer
		}
		if params := pathParams(oaPath); len(params) > 0 {
			op["parameters"] = params
		}

		if paths[oaPath] == nil {
			paths[oaPath] = map[string]any{}
		}
		paths[oaPath][strings.ToLower(r.Method)] = op
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "go-framework API",
			"version":     Version,
			"description": "由已注册路由自动生成。系统接口需先经 /api/v1/auth/login 获取 accessToken，在右上角 Authorize 填入。",
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"bearerAuth": map[string]any{
					"type": "http", "scheme": "bearer", "bearerFormat": "JWT",
				},
			},
		},
		"paths": paths,
	}
}

// ginPathToOpenAPI 将 /a/:id 转为 /a/{id}。
func ginPathToOpenAPI(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, ":") {
			segs[i] = "{" + s[1:] + "}"
		} else if strings.HasPrefix(s, "*") {
			segs[i] = "{" + s[1:] + "}"
		}
	}
	return strings.Join(segs, "/")
}

// pathParams 从 OpenAPI 路径中提取 {xxx} 作为 path 参数定义。
func pathParams(oaPath string) []map[string]any {
	var params []map[string]any
	for _, s := range strings.Split(oaPath, "/") {
		if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
			name := s[1 : len(s)-1]
			params = append(params, map[string]any{
				"name": name, "in": "path", "required": true,
				"schema": map[string]any{"type": "string"},
			})
		}
	}
	return params
}

// routeTag 用路径中 /api/v1/<tag> 的段作为分组标签。
func routeTag(p string) string {
	segs := strings.Split(strings.TrimPrefix(p, "/"), "/")
	if len(segs) >= 3 {
		return segs[2]
	}
	return "default"
}

var _ = sort.Strings

const swaggerHTML = `<!DOCTYPE html>
<html lang="zh">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>go-framework API 文档</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: 'doc.json',
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis],
        layout: 'BaseLayout'
      });
    };
  </script>
</body>
</html>`
