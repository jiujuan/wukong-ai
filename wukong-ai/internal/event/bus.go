package event

import (
	"sync"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// ProgressEvent 进度事件
type ProgressEvent struct {
	Type      string `json:"type"`      // node_start/node_done/sub_agent_update/task_done/task_failed
	Node      string `json:"node,omitempty"`
	Status    string `json:"status,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	Total     int    `json:"total,omitempty"`
	Done      int    `json:"done,omitempty"`
	Latest    string `json:"latest,omitempty"`
	Output    string `json:"final_output,omitempty"`
	Timestamp string `json:"timestamp"`
}

// EventBus 事件总线
type EventBus struct {
	mu   sync.RWMutex
	subs map[string][]chan ProgressEvent
}

// NewEventBus 创建新的事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[string][]chan ProgressEvent),
	}
}

// Subscribe 订阅任务事件
func (b *EventBus) Subscribe(taskID string) chan ProgressEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan ProgressEvent, 100)
	b.subs[taskID] = append(b.subs[taskID], ch)

	logger.Debug("subscribed to task events", "task_id", taskID)
	return ch
}

// Unsubscribe 取消订阅
func (b *EventBus) Unsubscribe(taskID string, ch chan ProgressEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subs[taskID]; ok {
		for i, sub := range subs {
			if sub == ch {
				b.subs[taskID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
	}
}

// Publish 发布事件
func (b *EventBus) Publish(taskID string, event ProgressEvent) {
	event.Timestamp = time.Now().Format(time.RFC3339)

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subs[taskID] {
		select {
		case ch <- event:
		default:
			// 消费者慢时丢弃，不阻塞主流程
			logger.Warn("event dropped, channel full", "task_id", taskID)
		}
	}
}

// CloseTask 关闭任务的所有订阅
func (b *EventBus) CloseTask(taskID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subs[taskID]; ok {
		for _, ch := range subs {
			close(ch)
		}
		delete(b.subs, taskID)
	}
}

// GetSubscriberCount 获取订阅者数量
func (b *EventBus) GetSubscriberCount(taskID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subs[taskID]; ok {
		return len(subs)
	}
	return 0
}
