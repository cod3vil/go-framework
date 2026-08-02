// Package cronx 封装 robfig/cron，提供任务注册表与动态调度管理。
//
// 使用方式：先用 Register 注册命名任务处理器（Go 函数），
// 再用 Schedule 按 cron 表达式将某个任务加入调度；每次执行结果通过 OnResult 回调上报。
package cronx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Task 任务处理器。返回 error 表示执行失败，会被记录到执行日志。
type Task func(ctx context.Context) error

// Result 一次任务执行的结果，供上层落库为执行日志。
type Result struct {
	JobID    uint
	JobKey   string
	Success  bool
	Message  string
	Duration time.Duration
	RunAt    time.Time
}

// Manager 定时任务管理器。
type Manager struct {
	cron     *cron.Cron
	mu       sync.RWMutex
	tasks    map[string]Task       // jobKey -> 处理器
	entries  map[uint]cron.EntryID // 业务 jobID -> cron 内部 entry
	onResult func(Result)
}

// New 创建管理器（秒级精度关闭，使用标准 5 段 cron 表达式）。
func New() *Manager {
	return &Manager{
		cron:    cron.New(),
		tasks:   make(map[string]Task),
		entries: make(map[uint]cron.EntryID),
	}
}

// SetResultHook 设置执行结果回调（用于落库执行日志）。
func (m *Manager) SetResultHook(fn func(Result)) {
	m.mu.Lock()
	m.onResult = fn
	m.mu.Unlock()
}

// Register 注册一个命名任务处理器。
func (m *Manager) Register(key string, task Task) {
	m.mu.Lock()
	m.tasks[key] = task
	m.mu.Unlock()
}

// TaskKeys 返回已注册的任务键列表，供前端下拉选择。
func (m *Manager) TaskKeys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.tasks))
	for k := range m.tasks {
		keys = append(keys, k)
	}
	return keys
}

// HasTask 判断任务键是否已注册。
func (m *Manager) HasTask(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.tasks[key]
	return ok
}

// Schedule 按 cron 表达式调度业务任务；若该 jobID 已在调度中，先移除再加入。
func (m *Manager) Schedule(jobID uint, jobKey, spec string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	task, ok := m.tasks[jobKey]
	if !ok {
		return fmt.Errorf("任务 %s 未注册", jobKey)
	}
	if eid, exists := m.entries[jobID]; exists {
		m.cron.Remove(eid)
		delete(m.entries, jobID)
	}
	eid, err := m.cron.AddFunc(spec, m.wrap(jobID, jobKey, task))
	if err != nil {
		return fmt.Errorf("无效的 cron 表达式: %w", err)
	}
	m.entries[jobID] = eid
	return nil
}

// Unschedule 从调度中移除业务任务。
func (m *Manager) Unschedule(jobID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if eid, exists := m.entries[jobID]; exists {
		m.cron.Remove(eid)
		delete(m.entries, jobID)
	}
}

// RunOnce 立即同步执行一次指定任务（手动触发），结果同样上报回调。
func (m *Manager) RunOnce(jobID uint, jobKey string) error {
	m.mu.RLock()
	task, ok := m.tasks[jobKey]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("任务 %s 未注册", jobKey)
	}
	m.wrap(jobID, jobKey, task)()
	return nil
}

// ValidateSpec 校验 cron 表达式是否合法。
func (m *Manager) ValidateSpec(spec string) error {
	_, err := cron.ParseStandard(spec)
	return err
}

// wrap 包装任务：计时、捕获 panic、上报结果。
func (m *Manager) wrap(jobID uint, jobKey string, task Task) func() {
	return func() {
		start := time.Now()
		res := Result{JobID: jobID, JobKey: jobKey, RunAt: start, Success: true, Message: "执行成功"}
		func() {
			defer func() {
				if r := recover(); r != nil {
					res.Success = false
					res.Message = fmt.Sprintf("任务 panic: %v", r)
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := task(ctx); err != nil {
				res.Success = false
				res.Message = err.Error()
			}
		}()
		res.Duration = time.Since(start)

		m.mu.RLock()
		hook := m.onResult
		m.mu.RUnlock()
		if hook != nil {
			hook(res)
		}
	}
}

// Start 启动调度（非阻塞）。
func (m *Manager) Start() { m.cron.Start() }

// Stop 停止调度并等待执行中的任务完成。
func (m *Manager) Stop() context.Context { return m.cron.Stop() }
