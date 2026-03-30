package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/uuid"
)

// RunHandler 运行处理器
type RunHandler struct {
	queue *queue.PersistentQueue
}

// NewRunHandler 创建运行处理器
func NewRunHandler(queue *queue.PersistentQueue) *RunHandler {
	return &RunHandler{
		queue: queue,
	}
}

// RunRequest 运行请求
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

	// 生成 task_id
	taskID := uuid.NewTaskID()
	createTime := time.Now().Format(time.RFC3339)

	// 创建初始状态
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

	// 序列化请求作为 payload
	payload, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal request"})
		return
	}

	// 入队
	if err := h.queue.Enqueue(c.Request.Context(), taskID, payload, 0); err != nil {
		logger.Error("failed to enqueue task", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue task"})
		return
	}

	// 立即返回
	c.JSON(http.StatusAccepted, RunResponse{
		TaskID:     taskID,
		Status:     "queued",
		Mode:       state.Mode,
		StreamURL:  "/api/task/stream?task_id=" + taskID,
		CreateTime: createTime,
	})
}

// determineMode 确定执行模式
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
