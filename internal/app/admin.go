package app

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// registerAdmin 将管理后台前端挂载到 /admin。
// 前端为 SPA（HashRouter），静态资源命中即返回，未命中回退到 index.html。
func (a *App) registerAdmin() {
	if a.adminFS == nil {
		return
	}
	index, err := fs.ReadFile(a.adminFS, "index.html")
	if err != nil {
		a.Logger.Warn("管理后台前端未内置，跳过 /admin 挂载: " + err.Error())
		return
	}
	fileServer := http.FileServer(http.FS(a.adminFS))

	// 根路径重定向到 /admin/。
	a.Engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/admin/")
	})

	handler := func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/admin")
		p = strings.TrimPrefix(p, "/")
		if p == "" || p == "index.html" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
			return
		}
		// 资源存在则交给文件服务，否则回退到 index.html（SPA 前端路由）。
		if _, err := fs.Stat(a.adminFS, p); err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", index)
			return
		}
		c.Request.URL.Path = "/" + p
		fileServer.ServeHTTP(c.Writer, c.Request)
	}

	a.Engine.GET("/admin", func(c *gin.Context) { c.Redirect(http.StatusFound, "/admin/") })
	a.Engine.GET("/admin/*filepath", handler)
}
