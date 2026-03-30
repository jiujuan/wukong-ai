package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// ListHandler 列表处理器
type ListHandler struct{}

// NewListHandler 创建列表处理器
func NewListHandler() *ListHandler {
	return &ListHandler{}
}

// ListResponse 列表响应
type ListResponse struct {
	Total int                    `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
	Tasks []TaskListItem         `json:"tasks"`
}

// TaskListItem 任务列表项
type TaskListItem struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	Mode       string `json:"mode"`
	UserInput  string `json:"user_input"`
	CreateTime string `json:"create_time"`
	FinishTime string `json:"finish_time,omitempty"`
}

// Handle 处理列表请求
func (h *ListHandler) Handle(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	tasks, total, err := repository.ListTasks(page, size, status)
	if err != nil {
		logger.Error("failed to list tasks", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	// 构建响应
	items := make([]TaskListItem, len(tasks))
	for i, task := range tasks {
		items[i] = TaskListItem{
			TaskID:     task.ID,
			Status:     task.Status,
			Mode:       task.Mode,
			UserInput:  task.UserInput,
			CreateTime: task.CreateTime.Format("2006-01-02T15:04:05Z07:00"),
		}
		if task.FinishTime.Valid {
			items[i].FinishTime = task.FinishTime.Time.Format("2006-01-02T15:04:05Z07:00")
		}
	}

	c.JSON(http.StatusOK, ListResponse{
		Total: total,
		Page:  page,
		Size:  size,
		Tasks: items,
	})
}
