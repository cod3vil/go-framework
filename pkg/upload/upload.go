// Package upload 提供文件上传能力：Storage 接口抽象底层存储，
// 内置本地磁盘实现，接口预留以便扩展 OSS/S3。
package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/google/uuid"
)

// datePath 返回 yyyy/MM/dd 形式的日期目录，用于分散存储文件。
func datePath() string {
	return time.Now().Format("2006/01/02")
}

// Storage 存储后端抽象。key 为存储相对路径。
type Storage interface {
	// Save 保存内容，返回可对外访问的 URL。
	Save(ctx context.Context, key string, r io.Reader) (url string, err error)
	// Delete 删除对象。
	Delete(ctx context.Context, key string) error
	// LocalPath 返回本地文件路径用于下载；非本地存储返回空串与 false。
	LocalPath(key string) (string, bool)
}

// Uploader 处理上传：校验白名单、生成存储 key、委托 Storage 保存。
type Uploader struct {
	storage     Storage
	maxSize     int64
	allowedExts []string
}

// Result 上传结果。
type Result struct {
	Key      string
	URL      string
	Filename string
	Size     int64
	Ext      string
}

// New 根据配置创建上传器。当前仅支持 local 驱动。
func New(cfg config.UploadConfig) (*Uploader, error) {
	var storage Storage
	switch cfg.Driver {
	case "local", "":
		s, err := newLocalStorage(cfg.Dir, cfg.URLPrefix)
		if err != nil {
			return nil, err
		}
		storage = s
	default:
		return nil, fmt.Errorf("暂不支持的上传驱动: %s", cfg.Driver)
	}
	return &Uploader{
		storage:     storage,
		maxSize:     int64(cfg.MaxSizeMB) * 1024 * 1024,
		allowedExts: cfg.AllowedExts,
	}, nil
}

// Storage 暴露底层存储（供下载定位文件）。
func (u *Uploader) Storage() Storage { return u.storage }

// Save 校验并保存一个上传文件头。key 形如 2026/08/02/<uuid>.<ext>。
func (u *Uploader) Save(ctx context.Context, fh *multipart.FileHeader) (*Result, error) {
	if u.maxSize > 0 && fh.Size > u.maxSize {
		return nil, fmt.Errorf("文件大小超过限制（最大 %d MB）", u.maxSize/1024/1024)
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if len(u.allowedExts) > 0 && !slices.Contains(u.allowedExts, ext) {
		return nil, fmt.Errorf("不允许的文件类型: %s", ext)
	}

	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	key := datePath() + "/" + uuid.NewString() + ext
	url, err := u.storage.Save(ctx, key, src)
	if err != nil {
		return nil, err
	}
	return &Result{Key: key, URL: url, Filename: fh.Filename, Size: fh.Size, Ext: ext}, nil
}

// Delete 删除已保存的对象。
func (u *Uploader) Delete(ctx context.Context, key string) error {
	return u.storage.Delete(ctx, key)
}
