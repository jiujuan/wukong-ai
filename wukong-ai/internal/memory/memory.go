package memory

import (
	"context"
)

// Memory 记忆系统接口
type Memory interface {
	// Save 保存记忆
	Save(ctx context.Context, content string) error
	// Query 查询记忆
	Query(ctx context.Context, query string, topK int) ([]string, error)
}

// MemoryType 记忆类型
type MemoryType string

const (
	MemoryTypeShortTerm  MemoryType = "short_term"
	MemoryTypeLongTerm   MemoryType = "long_term"
	MemoryTypeStructured MemoryType = "structured"
)

// MemoryItem 记忆项
type MemoryItem struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	MemoryType MemoryType `json:"memory_type"`
	TaskID    string     `json:"task_id,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}
