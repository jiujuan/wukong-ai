package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/worker"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/uuid"
)

// RunHandler 运行处理器
type RunHandler struct {
	queue  *queue.PersistentQueue
	appCfg *config.AppConfig // 持有全局配置，用于填充工具默认值
}

// NewRunHandler 创建运行处理器
func NewRunHandler(queue *queue.PersistentQueue, appCfg *config.AppConfig) *RunHandler {
	return &RunHandler{
		queue:  queue,
		appCfg: appCfg,
	}
}

// RunRequest HTTP 请求体（用户可选覆盖工具开关）
type RunRequest struct {
	UserInput       string `json:"user_input" binding:"required"`
	ThinkingEnabled bool   `json:"thinking_enabled"`
	PlanEnabled     bool   `json:"plan_enabled"`
	SubAgentEnabled bool   `json:"subagent_enabled"`
	LLMProvider     string `json:"llm_provider,omitempty"`
	MaxSubAgents    int    `json:"max_sub_agents,omitempty"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
}

// RunResponse 运行响应
type RunResponse struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	Mode       string `json:"mode"`
	StreamURL  string `json:"stream_url,omitempty"`
	CreateTime string `json:"create_time"`
}

// Handle 处理运行请求
func (h *RunHandler) Handle(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID := uuid.NewTaskID()
	createTime := time.Now().Format(time.RFC3339)

	state := &repository.Task{
		ID:              taskID,
		Status:          "queued",
		Mode:            h.determineMode(req),
		UserInput:       req.UserInput,
		ThinkingEnabled: req.ThinkingEnabled,
		PlanEnabled:     req.PlanEnabled,
		SubagentEnabled: req.SubAgentEnabled,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}

	if err := repository.CreateTask(state); err != nil {
		logger.Error("failed to create task in database", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	// 构造 worker.RunRequest，从全局配置填充工具默认值
	workerReq := h.buildWorkerRequest(req)
	payload, err := json.Marshal(workerReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	if err := h.queue.Enqueue(c.Request.Context(), taskID, payload, 0); err != nil {
		logger.Error("failed to enqueue task", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue task"})
		return
	}

	c.JSON(http.StatusAccepted, RunResponse{
		TaskID:     taskID,
		Status:     "queued",
		Mode:       state.Mode,
		StreamURL:  "/api/task/stream?task_id=" + taskID,
		CreateTime: createTime,
	})
}

// buildWorkerRequest 从 HTTP 请求 + 全局配置组合出 worker.RunRequest
func (h *RunHandler) buildWorkerRequest(req RunRequest) worker.RunRequest {
	toolsCfg := h.appCfg.Tools
	sandboxCfg := h.appCfg.Sandbox

	return worker.RunRequest{
		UserInput:       req.UserInput,
		ThinkingEnabled: req.ThinkingEnabled,
		PlanEnabled:     req.PlanEnabled,
		SubAgentEnabled: req.SubAgentEnabled,
		MaxSubAgents:    req.MaxSubAgents,
		TimeoutSeconds:  req.TimeoutSeconds,
		// 工具配置 —— 读自全局 config
		TavilyAPIKey:      toolsCfg.Search.TavilyAPIKey,
		DuckDuckGoEnabled: toolsCfg.Search.DuckDuckGoEnabled,
		FileAllowedPaths:  toolsCfg.File.AllowedPaths,
		SandboxDir:        sandboxCfg.BaseDir,
		PythonReplEnabled: sandboxCfg.PythonReplEnabled,
		BashEnabled:       sandboxCfg.BashEnabled,
	}
}

// determineMode 确定执行模式字符串
func (h *RunHandler) determineMode(req RunRequest) string {
	if !req.ThinkingEnabled && !req.PlanEnabled && !req.SubAgentEnabled {
		return "flash"
	}
	if req.ThinkingEnabled && !req.PlanEnabled && !req.SubAgentEnabled {
		return "standard"
	}
	if req.ThinkingEnabled && req.PlanEnabled && !req.SubAgentEnabled {
		return "pro"
	}
	return "ultra"
}
