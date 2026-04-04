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
	appCfg *config.AppConfig
}

func NewRunHandler(queue *queue.PersistentQueue, appCfg *config.AppConfig) *RunHandler {
	return &RunHandler{queue: queue, appCfg: appCfg}
}

// RunRequest HTTP 请求体
type RunRequest struct {
	UserInput       string `json:"user_input" binding:"required"`
	ThinkingEnabled bool   `json:"thinking_enabled"`
	PlanEnabled     bool   `json:"plan_enabled"`
	SubAgentEnabled bool   `json:"subagent_enabled"`
	LLMProvider     string `json:"llm_provider,omitempty"`
	MaxSubAgents    int    `json:"max_sub_agents,omitempty"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
	ConversationID  string `json:"conversation_id,omitempty"` // 多轮对话关联 ID
}

// RunResponse 运行响应
type RunResponse struct {
	TaskID         string `json:"task_id"`
	Status         string `json:"status"`
	Mode           string `json:"mode"`
	StreamURL      string `json:"stream_url,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	CreateTime     string `json:"create_time"`
}

// Handle POST /api/run
func (h *RunHandler) Handle(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.handleRequest(c, req)
}

// HandleWithConvID POST /api/conversation/:id/run
// 将路径中的 conversation_id 注入请求，其余逻辑与 Handle 完全相同
func (h *RunHandler) HandleWithConvID(c *gin.Context, convID string) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ConversationID = convID // 路径参数优先
	h.handleRequest(c, req)
}

// handleRequest 核心执行逻辑（Handle / HandleWithConvID 共用）
func (h *RunHandler) handleRequest(c *gin.Context, req RunRequest) {
	// 如果指定了 conversation_id，校验对话是否存在
	if req.ConversationID != "" {
		conv, err := repository.GetConversation(req.ConversationID)
		if err != nil {
			logger.Error("failed to validate conversation",
				"conversation_id", req.ConversationID, "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate conversation"})
			return
		}
		if conv == nil {
			c.JSON(http.StatusBadRequest,
				gin.H{"error": "conversation not found: " + req.ConversationID})
			return
		}
	}

	taskID := uuid.NewTaskID()
	createTime := time.Now().Format(time.RFC3339)

	task := &repository.Task{
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

	if err := repository.CreateTask(task); err != nil {
		logger.Error("failed to create task in database", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

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
		TaskID:         taskID,
		Status:         "queued",
		Mode:           task.Mode,
		StreamURL:      "/api/task/stream?task_id=" + taskID,
		ConversationID: req.ConversationID,
		CreateTime:     createTime,
	})
}

// buildWorkerRequest 从 HTTP 请求 + 全局配置组合 worker.RunRequest
func (h *RunHandler) buildWorkerRequest(req RunRequest) worker.RunRequest {
	toolsCfg := h.appCfg.Tools
	sandboxCfg := h.appCfg.Sandbox
	return worker.RunRequest{
		UserInput:         req.UserInput,
		ThinkingEnabled:   req.ThinkingEnabled,
		PlanEnabled:       req.PlanEnabled,
		SubAgentEnabled:   req.SubAgentEnabled,
		MaxSubAgents:      req.MaxSubAgents,
		TimeoutSeconds:    req.TimeoutSeconds,
		ConversationID:    req.ConversationID,
		TavilyAPIKey:      toolsCfg.Search.TavilyAPIKey,
		DuckDuckGoEnabled: toolsCfg.Search.DuckDuckGoEnabled,
		FileAllowedPaths:  toolsCfg.File.AllowedPaths,
		SandboxDir:        sandboxCfg.BaseDir,
		PythonReplEnabled: sandboxCfg.PythonReplEnabled,
		BashEnabled:       sandboxCfg.BashEnabled,
	}
}

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
