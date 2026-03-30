package memory

import (
	"sync"
)

// ShortTermMemory 短期上下文记忆
type ShortTermMemory struct {
	messages     []Message
	maxMessages  int
	mu           sync.RWMutex
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewShortTermMemory 创建短期记忆
func NewShortTermMemory(maxMessages int) *ShortTermMemory {
	return &ShortTermMemory{
		messages:    make([]Message, 0, maxMessages),
		maxMessages: maxMessages,
	}
}

// Add 添加消息
func (m *ShortTermMemory) Add(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, Message{Role: role, Content: content})

	// 如果超过最大消息数，删除最旧的消息
	if len(m.messages) > m.maxMessages {
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}
}

// GetAll 获取所有消息
func (m *ShortTermMemory) GetAll() []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Message, len(m.messages))
	copy(result, m.messages)
	return result
}

// GetRecent 获取最近 n 条消息
func (m *ShortTermMemory) GetRecent(n int) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n > len(m.messages) {
		n = len(m.messages)
	}

	result := make([]Message, n)
	copy(result, m.messages[len(m.messages)-n:])
	return result
}

// Clear 清空记忆
func (m *ShortTermMemory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]Message, 0, m.maxMessages)
}

// Size 获取当前消息数
func (m *ShortTermMemory) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}
