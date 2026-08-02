package handler

import (
	"strconv"

	"github.com/cod3vil/go-framework/internal/middleware"
	"github.com/cod3vil/go-framework/internal/system/service"
	"github.com/cod3vil/go-framework/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListJobs 分页查询定时任务。
// GET /api/v1/system/jobs
func (h *Handler) ListJobs(c *gin.Context) {
	page, pageSize := getPage(c)
	status, _ := strconv.Atoi(c.Query("status"))
	jobs, total, err := h.svc.ListJobs(c.Request.Context(), service.JobQuery{
		Page: page, PageSize: pageSize, Name: c.Query("name"), Status: int8(status),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, jobs, total, page, pageSize)
}

// ListJobTasks 返回已注册的任务处理器键，供创建任务时选择。
// GET /api/v1/system/jobs/tasks
func (h *Handler) ListJobTasks(c *gin.Context) {
	response.OK(c, h.svc.RegisteredTasks())
}

type jobReq struct {
	Name     string `json:"name" binding:"required,max=64"`
	JobKey   string `json:"jobKey" binding:"required,max=64"`
	CronExpr string `json:"cronExpr" binding:"required,max=64"`
	Status   int8   `json:"status" binding:"omitempty,oneof=1 2"`
	Remark   string `json:"remark" binding:"max=255"`
}

func (r jobReq) toInput() service.JobInput {
	status := r.Status
	if status == 0 {
		status = 2 // 默认停用
	}
	return service.JobInput{Name: r.Name, JobKey: r.JobKey, CronExpr: r.CronExpr, Status: status, Remark: r.Remark}
}

// CreateJob 创建定时任务。
// POST /api/v1/system/jobs
func (h *Handler) CreateJob(c *gin.Context) {
	var req jobReq
	if !bindJSON(c, &req) {
		return
	}
	job, err := h.svc.CreateJob(c.Request.Context(), req.toInput(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, job)
}

// UpdateJob 更新定时任务。
// PUT /api/v1/system/jobs/:id
func (h *Handler) UpdateJob(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req jobReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.UpdateJob(c.Request.Context(), id, req.toInput(), middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// SetJobStatus 启用/停用任务。
// PUT /api/v1/system/jobs/:id/status
func (h *Handler) SetJobStatus(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req statusReq
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.SetJobStatus(c.Request.Context(), id, req.Status, middleware.UserID(c)); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// RunJob 手动触发一次任务。
// POST /api/v1/system/jobs/:id/run
func (h *Handler) RunJob(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.RunJobOnce(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// DeleteJob 删除任务。
// DELETE /api/v1/system/jobs/:id
func (h *Handler) DeleteJob(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteJob(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// ListJobLogs 分页查询任务执行日志。
// GET /api/v1/system/job-logs
func (h *Handler) ListJobLogs(c *gin.Context) {
	page, pageSize := getPage(c)
	jobID, _ := strconv.ParseUint(c.Query("jobId"), 10, 64)
	logs, total, err := h.svc.ListJobLogs(c.Request.Context(), service.JobLogQuery{
		Page: page, PageSize: pageSize, JobID: uint(jobID),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPage(c, logs, total, page, pageSize)
}
