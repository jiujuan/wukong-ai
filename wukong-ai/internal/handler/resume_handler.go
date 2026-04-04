package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/worker"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// ResumeHandler 断点续跑处理器
type ResumeHandler struct {
	queue  *queue.PersistentQueue
	appCfg *config.AppConfig
}

// NewResumeHandler 创建断点续跑处理器
func NewResumeHandler(q *queue.PersistentQueue, appCfg *config.AppConfig) *ResumeHandler {
	return &ResumeHandler{queue: q, appCfg: appCfg}
}

// ResumeRequest 续跑请求
type ResumeRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// ResumeResponse 续跑响应
type ResumeResponse struct {
	TaskID      string `json:"task_id"`
	Status      string `json:"status"`
	ResumedFrom string `json:"resumed_from"`
	Message     string `json:"message"`
}

// Handle 处理续跑请求
func (h *ResumeHandler) Handle(c *gin.Context) {
	var req ResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := repository.GetTaskByID(req.TaskID)
	if err != nil {
		logger.Error("failed to get task", "task_id", req.TaskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	if task.Status == "success" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already completed"})
		return
	}
	if task.Status == "running" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task is already running"})
		return
	}

	lastNode := ""
	if task.LastNode.Valid {
		lastNode = task.LastNode.String
	}

	// 重建 worker.RunRequest（含工具配置），标记断点续跑节点
	toolsCfg := h.appCfg.Tools
	sandboxCfg := h.appCfg.Sandbox
	workerReq := worker.RunRequest{
		UserInput:         task.UserInput,
		ThinkingEnabled:   task.ThinkingEnabled,
		PlanEnabled:       task.PlanEnabled,
		SubAgentEnabled:   task.SubagentEnabled,
		ResumeFrom:        lastNode,
		TavilyAPIKey:      toolsCfg.Search.TavilyAPIKey,
		DuckDuckGoEnabled: toolsCfg.Search.DuckDuckGoEnabled,
		FileAllowedPaths:  toolsCfg.File.AllowedPaths,
		SandboxDir:        sandboxCfg.BaseDir,
		PythonReplEnabled: sandboxCfg.PythonReplEnabled,
		BashEnabled:       sandboxCfg.BashEnabled,
	}

	payload, err := json.Marshal(workerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal resume request"})
		return
	}

	if err := repository.UpdateTaskStatus(req.TaskID, "queued"); err != nil {
		logger.Error("failed to update task status", "task_id", req.TaskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resume task"})
		return
	}

	if err := h.queue.Enqueue(c.Request.Context(), req.TaskID, payload, 1); err != nil {
		logger.Error("failed to re-enqueue task", "task_id", req.TaskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to re-enqueue task"})
		return
	}

	c.JSON(http.StatusOK, ResumeResponse{
		TaskID:      req.TaskID,
		Status:      "queued",
		ResumedFrom: lastNode,
		Message:     "task resumed from breakpoint",
	})
}
