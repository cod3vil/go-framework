// Package goframework 是模块根包，持有管理后台前端的嵌入产物。
package goframework

import (
	"embed"
	"io/fs"
)

// 管理后台前端构建产物，通过 go:embed 打进二进制实现单文件部署。
// 构建流程：cd web && pnpm build（或 make web），产物在 web/dist。
// 为保证 `go build` 无需 Node 即可编译，web/dist 已提交入库。
//
//go:embed all:web/dist
var webDist embed.FS

// AdminFS 返回以 web/dist 为根的文件系统，供 HTTP 静态服务挂载到 /admin。
func AdminFS() (fs.FS, error) {
	return fs.Sub(webDist, "web/dist")
}
