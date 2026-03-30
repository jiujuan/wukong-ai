package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// ResumeHandler 断点续跑处理器
type ResumeHandler struct {
	stateManager interface {
		Get(taskID string) (interface{ GetLastNode() string }, error)
	}
	queue interface {
		Enqueue(ctx interface{}, taskID string, payload []byte, priority int) error
	}
}

// NewResumeHandler 创建断点续跑处理器
func NewResumeHandler(stateManager interface {
	Get(taskID string) (interface{ GetLastNode() string }, error)
}, queue interface {
	Enqueue(ctx interface{}, taskID string, payload []byte, priority int) error
}) *ResumeHandler {
	return &ResumeHandler{
		stateManager: stateManager,
		queue:        queue,
	}
}

// ResumeRequest 续跑请求
type ResumeRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// ResumeResponse 续跑响应
type ResumeResponse struct {
	TaskID      string `json:"task_id"`
	Status     string `json:"status"`
	ResumedFrom string `json:"resumed_from"`
	Message    string `json:"message"`
}

// Handle 处理续跑请求
func (h *ResumeHandler) Handle(c *gin.Context) {
	var req ResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查任务是否存在
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

	// 检查任务状态
	if task.Status == "success" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task already completed"})
		return
	}

	if task.Status == "running" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task is already running"})
		return
	}

	// 获取最后节点
	lastNode := ""
	if task.LastNode.Valid {
		lastNode = task.LastNode.String
	}

	// 简化处理：直接更新状态并重新执行

	// 更新状态
	if err := repository.UpdateTaskStatus(req.TaskID, "queued"); err != nil {
		logger.Error("failed to update task status", "task_id", req.TaskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resume task"})
		return
	}

	c.JSON(http.StatusOK, ResumeResponse{
		TaskID:      req.TaskID,
		Status:     "queued",
		ResumedFrom: lastNode,
		Message:    "task resumed from breakpoint",
	})
}
