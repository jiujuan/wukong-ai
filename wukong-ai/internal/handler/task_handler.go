package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// TaskHandler 任务处理器
type TaskHandler struct{}

// NewTaskHandler 创建任务处理器
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

// TaskResponse 任务响应
type TaskResponse struct {
	TaskID          string   `json:"task_id"`
	Status          string   `json:"status"`
	Mode            string   `json:"mode"`
	UserInput       string   `json:"user_input"`
	Intention       string   `json:"intention,omitempty"`
	Plan            string   `json:"plan,omitempty"`
	Tasks           []string `json:"tasks,omitempty"`
	SubResults      []string `json:"sub_results,omitempty"`
	FinalOutput     string   `json:"final_output,omitempty"`
	LastNode        string   `json:"last_node,omitempty"`
	CreateTime      string   `json:"create_time"`
	FinishTime      string   `json:"finish_time,omitempty"`
	ErrorMsg        string   `json:"error_msg,omitempty"`
}

// Handle 处理获取任务请求
func (h *TaskHandler) Handle(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	task, err := repository.GetTaskByID(taskID)
	if err != nil {
		logger.Error("failed to get task", "task_id", taskID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 构建响应
	response := TaskResponse{
		TaskID:     task.ID,
		Status:     task.Status,
		Mode:       task.Mode,
		UserInput:  task.UserInput,
		CreateTime: task.CreateTime.Format("2006-01-02T15:04:05Z07:00"),
	}

	if task.Intention.Valid {
		response.Intention = task.Intention.String
	}
	if task.Plan.Valid {
		response.Plan = task.Plan.String
	}
	if task.FinalOutput.Valid {
		response.FinalOutput = task.FinalOutput.String
	}
	if task.LastNode.Valid {
		response.LastNode = task.LastNode.String
	}
	if task.FinishTime.Valid {
		response.FinishTime = task.FinishTime.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if task.ErrorMsg.Valid {
		response.ErrorMsg = task.ErrorMsg.String
	}

	// 解析 tasks_list
	if task.TasksList != nil {
		var tasks []string
		if err := json.Unmarshal(task.TasksList, &tasks); err == nil {
			response.Tasks = tasks
		}
	}

	// 解析 sub_results
	if task.SubResults != nil {
		var subResults []string
		if err := json.Unmarshal(task.SubResults, &subResults); err == nil {
			response.SubResults = subResults
		}
	}

	c.JSON(http.StatusOK, response)
}
