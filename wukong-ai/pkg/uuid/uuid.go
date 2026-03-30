package uuid

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// New 生成通用 UUID 字符串（格式：xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx）
func New() string {
	return uuid.New().String()
}

// NewShort 生成短 UUID（8位）
func NewShort() string {
	return uuid.New().String()[:8]
}

// NewTaskID 生成带时间前缀的任务 ID（格式：task_20250301_xxxxxxxx）
func NewTaskID() string {
	date := time.Now().Format("20060102")
	short := uuid.New().String()[:8]
	return fmt.Sprintf("task_%s_%s", date, short)
}

// NewSessionID 生成会话 ID（格式：sess_xxxxxxxx_xxxxxxxx）
func NewSessionID() string {
	return fmt.Sprintf("sess_%s_%s", uuid.New().String()[:8], uuid.New().String()[9:17])
}

// ParseTaskID 从 task_id 中提取日期部分
func ParseTaskID(taskID string) (date string, ok bool) {
	if len(taskID) < 14 || taskID[:5] != "task_" {
		return "", false
	}
	return taskID[5:13], true
}

// IsValidTaskID 检查 task_id 格式是否有效
func IsValidTaskID(taskID string) bool {
	if len(taskID) < 15 || taskID[:5] != "task_" {
		return false
	}
	// 检查日期部分是否为有效日期
	dateStr := taskID[5:13]
	if _, err := time.Parse("20060102", dateStr); err != nil {
		return false
	}
	return true
}
