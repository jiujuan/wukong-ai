package conversation

import "time"

// Conversation 对话会话
type Conversation struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`       // 对话标题（首轮输入自动截取）
	Summary    string    `json:"summary"`     // 早期历史滚动压缩摘要
	TurnCount  int       `json:"turn_count"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// Turn 单轮对话
type Turn struct {
	ID             int64     `json:"id"`
	ConversationID string    `json:"conversation_id"`
	TaskID         string    `json:"task_id,omitempty"` // 关联的任务 ID
	TurnIndex      int       `json:"turn_index"`
	Role           string    `json:"role"`            // user / assistant
	Content        string    `json:"content"`         // 用户输入 或 输出摘要
	FullOutput     string    `json:"full_output,omitempty"` // 完整输出（仅 assistant）
	CreateTime     time.Time `json:"create_time"`
}

// MaxRecentTurns 注入 Prompt 时保留最近轮次数
const MaxRecentTurns = 10

// SummaryThreshold 超过此轮数时触发摘要压缩
const SummaryThreshold = 20
