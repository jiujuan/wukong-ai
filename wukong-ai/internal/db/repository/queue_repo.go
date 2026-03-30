package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// TaskQueueItem 任务队列项
type TaskQueueItem struct {
	ID           int64           `json:"id"`
	TaskID       string          `json:"task_id"`
	Status       string          `json:"status"`
	Priority     int             `json:"priority"`
	Payload      json.RawMessage `json:"payload"`
	RetryCount   int             `json:"retry_count"`
	MaxRetries   int             `json:"max_retries"`
	WorkerID     sql.NullString  `json:"worker_id"`
	EnqueueTime  time.Time       `json:"enqueue_time"`
	StartTime    sql.NullTime    `json:"start_time"`
	FinishTime   sql.NullTime    `json:"finish_time"`
	NextRetryAt  sql.NullTime   `json:"next_retry_at"`
}

// EnqueueTask 将任务加入队列
func EnqueueTask(taskID string, payload json.RawMessage, priority int) error {
	db := db.Get()
	query := `
		INSERT INTO task_queue (task_id, status, priority, payload, retry_count, max_retries, enqueue_time)
		VALUES ($1, 'queued', $2, $3, 0, 3, NOW())
		ON CONFLICT (task_id) DO UPDATE SET
			status = 'queued',
			payload = $3,
			enqueue_time = NOW(),
			finish_time = NULL
	`
	_, err := db.Exec(query, taskID, priority, payload)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	logger.Debug("task enqueued", "task_id", taskID)
	return nil
}

// DequeueTask 原子性取出待执行任务
func DequeueTask(workerID string) (*TaskQueueItem, error) {
	db := db.Get()
	query := `
		UPDATE task_queue
		SET status = 'running',
			worker_id = $1,
			start_time = NOW()
		WHERE id = (
			SELECT id FROM task_queue
			WHERE status = 'queued'
			  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
			ORDER BY priority DESC, enqueue_time ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, task_id, status, priority, payload, retry_count, max_retries, worker_id, enqueue_time, start_time, finish_time, next_retry_at
	`
	var item TaskQueueItem
	err := db.QueryRow(query, workerID).Scan(
		&item.ID, &item.TaskID, &item.Status, &item.Priority,
		&item.Payload, &item.RetryCount, &item.MaxRetries, &item.WorkerID,
		&item.EnqueueTime, &item.StartTime, &item.FinishTime, &item.NextRetryAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // 队列为空
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue task: %w", err)
	}
	logger.Debug("task dequeued", "task_id", item.TaskID, "worker", workerID)
	return &item, nil
}

// MarkTaskSuccess 标记任务成功
func MarkTaskSuccess(taskID string) error {
	db := db.Get()
	query := `
		UPDATE task_queue
		SET status = 'success',
			finish_time = NOW()
		WHERE task_id = $1
	`
	_, err := db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to mark task success: %w", err)
	}
	return nil
}

// MarkTaskFailed 标记任务失败
func MarkTaskFailed(taskID, errorMsg string, retryCount int) error {
	db := db.Get()
	query := `
		UPDATE task_queue
		SET status = CASE
			WHEN retry_count + 1 >= max_retries THEN 'failed'
			ELSE 'queued'
		  END,
			retry_count = retry_count + 1,
			next_retry_at = NOW() + (5 * (retry_count + 1) || ' seconds')::interval,
			finish_time = CASE
			WHEN retry_count + 1 >= max_retries THEN NOW()
			ELSE NULL
		  END
		WHERE task_id = $1
	`
	_, err := db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to mark task failed: %w", err)
	}
	return nil
}

// RecoverStaleJobs 恢复僵尸任务
func RecoverStaleJobs(staleDuration time.Duration) error {
	db := db.Get()
	query := `
		UPDATE task_queue
		SET status = 'queued',
			worker_id = NULL,
			retry_count = retry_count + 1,
			next_retry_at = NOW() + interval '5 seconds'
		WHERE status = 'running'
		  AND start_time < NOW() - $1::interval
	`
	_, err := db.Exec(query, staleDuration.String())
	if err != nil {
		return fmt.Errorf("failed to recover stale jobs: %w", err)
	}
	return nil
}

// GetQueueItem 获取队列项
func GetQueueItem(taskID string) (*TaskQueueItem, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, status, priority, payload, retry_count, max_retries, worker_id, enqueue_time, start_time, finish_time, next_retry_at
		FROM task_queue
		WHERE task_id = $1
	`
	var item TaskQueueItem
	err := db.QueryRow(query, taskID).Scan(
		&item.ID, &item.TaskID, &item.Status, &item.Priority,
		&item.Payload, &item.RetryCount, &item.MaxRetries, &item.WorkerID,
		&item.EnqueueTime, &item.StartTime, &item.FinishTime, &item.NextRetryAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get queue item: %w", err)
	}
	return &item, nil
}

// CancelTask 取消任务
func CancelTask(taskID string) error {
	db := db.Get()
	query := `
		UPDATE task_queue
		SET status = 'cancelled',
			finish_time = NOW()
		WHERE task_id = $1 AND status IN ('queued', 'running')
	`
	_, err := db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to cancel task: %w", err)
	}
	return nil
}
