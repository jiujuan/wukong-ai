package queue

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// TaskJob 任务作业
type TaskJob struct {
	TaskID  string          `json:"task_id"`
	Payload json.RawMessage `json:"payload"`
}

// TaskQueue 任务队列
type TaskQueue struct {
	ch     chan TaskJob
	buffer int
	wg     sync.WaitGroup
	stopCh chan struct{}
}

// NewTaskQueue 创建新的任务队列
func NewTaskQueue(buffer int) *TaskQueue {
	return &TaskQueue{
		ch:     make(chan TaskJob, buffer),
		buffer: buffer,
		stopCh: make(chan struct{}),
	}
}

// Enqueue 入队
func (q *TaskQueue) Enqueue(job TaskJob) error {
	select {
	case q.ch <- job:
		logger.Debug("job enqueued", "task_id", job.TaskID)
		return nil
	case <-time.After(5 * time.Second):
		return ErrQueueFull
	}
}

// Dequeue 出队
func (q *TaskQueue) Dequeue(ctx context.Context) (*TaskJob, error) {
	select {
	case job := <-q.ch:
		return &job, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-q.stopCh:
		return nil, ErrQueueStopped
	}
}

// Len 获取队列长度
func (q *TaskQueue) Len() int {
	return len(q.ch)
}

// Stop 停止队列
func (q *TaskQueue) Stop() {
	close(q.stopCh)
}

// Wait 等待队列处理完成
func (q *TaskQueue) Wait() {
	q.wg.Wait()
}

// PersistentQueue 持久化任务队列
type PersistentQueue struct{}

// NewPersistentQueue 创建持久化任务队列
func NewPersistentQueue() *PersistentQueue {
	return &PersistentQueue{}
}

// Enqueue 入队到数据库
func (q *PersistentQueue) Enqueue(ctx context.Context, taskID string, payload json.RawMessage, priority int) error {
	return repository.EnqueueTask(taskID, payload, priority)
}

// Dequeue 出队（原子性操作）
func (q *PersistentQueue) Dequeue(ctx context.Context, workerID string) (*TaskJob, error) {
	item, err := repository.DequeueTask(workerID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, nil
	}

	return &TaskJob{
		TaskID:  item.TaskID,
		Payload: item.Payload,
	}, nil
}

// MarkSuccess 标记成功
func (q *PersistentQueue) MarkSuccess(taskID string) error {
	return repository.MarkTaskSuccess(taskID)
}

// MarkFailed 标记失败
func (q *PersistentQueue) MarkFailed(taskID, errorMsg string) error {
	return repository.MarkTaskFailed(taskID, errorMsg, 0)
}

// Cancel 取消任务
func (q *PersistentQueue) Cancel(taskID string) error {
	return repository.CancelTask(taskID)
}

// GetItem 获取队列项
func (q *PersistentQueue) GetItem(taskID string) (*repository.TaskQueueItem, error) {
	return repository.GetQueueItem(taskID)
}

// RecoverStaleJobs 恢复僵尸任务
func (q *PersistentQueue) RecoverStaleJobs(staleDuration time.Duration) error {
	return repository.RecoverStaleJobs(staleDuration)
}

// ErrQueueFull 队列满错误
var ErrQueueFull = &QueueError{Message: "queue is full"}

// ErrQueueStopped 队列已停止错误
var ErrQueueStopped = &QueueError{Message: "queue is stopped"}

// QueueError 队列错误
type QueueError struct {
	Message string
}

func (e *QueueError) Error() string {
	return e.Message
}
