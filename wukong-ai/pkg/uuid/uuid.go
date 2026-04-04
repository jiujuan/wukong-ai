package uuid

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// New 生成通用 UUID 字符串
func New() string {
	return uuid.New().String()
}

// NewTaskID 生成带日期前缀的任务 ID：task_20250301_xxxxxxxx
func NewTaskID() string {
	date := time.Now().Format("20060102")
	short := uuid.New().String()[:8]
	return fmt.Sprintf("task_%s_%s", date, short)
}

// NewConversationID 生成带前缀的对话 ID：conv_20250301_xxxxxxxx
func NewConversationID() string {
	date := time.Now().Format("20060102")
	short := uuid.New().String()[:8]
	return fmt.Sprintf("conv_%s_%s", date, short)
}
