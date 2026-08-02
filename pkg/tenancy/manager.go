package tenancy

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/cod3vil/go-framework/pkg/config"
	"github.com/cod3vil/go-framework/pkg/database"
	"gorm.io/gorm"
)

// searchPathRe 用于剥离基础 DSN 中已有的 search_path，避免重复。
var searchPathRe = regexp.MustCompile(`(?i)\s*search_path=\S+`)

// Manager 管理各租户 schema 的数据库句柄，每个 schema 一个独立连接池并缓存复用。
type Manager struct {
	baseCfg config.DatabaseConfig
	maxConn int
	mu      sync.RWMutex
	handles map[string]*gorm.DB
}

// NewManager 创建租户库管理器。base 为基础数据库配置（必须是 postgres）。
func NewManager(base config.DatabaseConfig, maxConnPerTenant int) *Manager {
	if maxConnPerTenant <= 0 {
		maxConnPerTenant = 10
	}
	return &Manager{
		baseCfg: base,
		maxConn: maxConnPerTenant,
		handles: make(map[string]*gorm.DB),
	}
}

// Supported 当前数据库驱动是否支持多租户（仅 postgres）。
func (m *Manager) Supported() bool {
	return m.baseCfg.Driver == "postgres"
}

// DB 返回指定 schema 的数据库句柄（带缓存）。首次调用时创建独立连接池。
func (m *Manager) DB(schema string) (*gorm.DB, error) {
	if !m.Supported() {
		return nil, fmt.Errorf("多租户仅支持 PostgreSQL，当前驱动: %s", m.baseCfg.Driver)
	}
	if !validSchema(schema) {
		return nil, fmt.Errorf("非法的 schema 名: %s", schema)
	}

	m.mu.RLock()
	if db, ok := m.handles[schema]; ok {
		m.mu.RUnlock()
		return db, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	if db, ok := m.handles[schema]; ok { // 双检
		return db, nil
	}

	cfg := m.baseCfg
	cfg.DSN = withSearchPath(m.baseCfg.DSN, schema)
	// 每租户连接池独立且较小，避免租户数量多时连接爆炸。
	cfg.MaxOpenConns = m.maxConn
	if cfg.MaxIdleConns > m.maxConn {
		cfg.MaxIdleConns = m.maxConn
	}
	db, err := database.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化租户 %s 数据库失败: %w", schema, err)
	}
	m.handles[schema] = db
	return db, nil
}

// Close 关闭所有租户连接池。
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, db := range m.handles {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
	m.handles = make(map[string]*gorm.DB)
}

// withSearchPath 在基础 DSN 上设置 search_path=<schema>,public，
// 使 schema 内对象优先、并回退 public（如共享的扩展/类型）。
func withSearchPath(dsn, schema string) string {
	cleaned := strings.TrimSpace(searchPathRe.ReplaceAllString(dsn, ""))
	return fmt.Sprintf("%s search_path=%s,public", cleaned, schema)
}

// validSchema 仅允许字母数字下划线，防止 SQL 注入到 DSN / DDL。
func validSchema(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}
