package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// CancelHandler 取消处理器
type CancelHandler struct {
	queue interface {
		Cancel(taskID string) error
	}
}

// NewCancelHandler 创建取消处理器
func NewCancelHandler(queue interface {
	Cancel(taskID string) error
}) *CancelHandler {
	return &CancelHandler{
		queue: queue,
	}
}

// CancelRequest 取消请求
type CancelRequest struct {
	TaskID string `json:"task_id" binding:"required"`
}

// CancelResponse 取消响应
type CancelResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// Handle 处理取消请求
func (h *CancelHandler) Handle(c *gin.Context) {
	var req CancelRequest
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

	// 只能取消 queued 或 running 状态的任务
	if task.Status != "queued" && task.Status != "running" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task cannot be cancelled in current status: " + task.Status})
		return
	}

	// 从队列取消
	if err := h.queue.Cancel(req.TaskID); err != nil {
		logger.Error("failed to cancel task in queue", "task_id", req.TaskID, "err", err)
	}

	// 更新任务状态
	if err := repository.FailTask(req.TaskID, "cancelled by user"); err != nil {
		logger.Error("failed to update task status", "task_id", req.TaskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel task"})
		return
	}

	c.JSON(http.StatusOK, CancelResponse{
		TaskID:  req.TaskID,
		Status:  "cancelled",
		Message: "task cancelled successfully",
	})
}
