package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// localStorage 本地磁盘存储。
type localStorage struct {
	baseDir   string
	urlPrefix string
}

func newLocalStorage(baseDir, urlPrefix string) (*localStorage, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("upload.dir 不能为空")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}
	return &localStorage{
		baseDir:   baseDir,
		urlPrefix: strings.TrimRight(urlPrefix, "/"),
	}, nil
}

// Save 将内容写入 baseDir/key，返回 urlPrefix/key。
func (l *localStorage) Save(_ context.Context, key string, r io.Reader) (string, error) {
	fullPath := filepath.Join(l.baseDir, filepath.Clean("/"+key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, r); err != nil {
		return "", err
	}
	return l.urlPrefix + "/" + key, nil
}

// Delete 删除本地文件。
func (l *localStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(l.baseDir, filepath.Clean("/"+key))
	err := os.Remove(fullPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// LocalPath 返回本地绝对路径供下载。
func (l *localStorage) LocalPath(key string) (string, bool) {
	return filepath.Join(l.baseDir, filepath.Clean("/"+key)), true
}

// BaseDir 返回存储根目录（供静态文件服务挂载）。
func (l *localStorage) BaseDir() string { return l.baseDir }
