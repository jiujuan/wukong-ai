package repository

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Task 任务结构体
type Task struct {
	ID              string          `json:"id"`
	Status          string          `json:"status"`
	Mode            string          `json:"mode"`
	UserInput       string          `json:"user_input"`
	Intention       sql.NullString  `json:"intention"`
	Plan            sql.NullString  `json:"plan"`
	TasksList       json.RawMessage `json:"tasks_list"`
	SubResults      json.RawMessage `json:"sub_results"`
	FinalOutput     sql.NullString  `json:"final_output"`
	LastNode        sql.NullString  `json:"last_node"`
	RetryCount      int             `json:"retry_count"`
	ErrorMsg        sql.NullString  `json:"error_msg"`
	ThinkingEnabled  bool            `json:"thinking_enabled"`
	PlanEnabled     bool            `json:"plan_enabled"`
	SubagentEnabled bool            `json:"subagent_enabled"`
	CreateTime      time.Time       `json:"create_time"`
	UpdateTime      time.Time       `json:"update_time"`
	FinishTime      sql.NullTime    `json:"finish_time"`
}

// CreateTask 创建新任务
func CreateTask(task *Task) error {
	db := db.Get()
	tasksList, err := normalizeJSONValue(task.TasksList)
	if err != nil {
		return fmt.Errorf("failed to normalize tasks_list: %w", err)
	}
	subResults, err := normalizeJSONValue(task.SubResults)
	if err != nil {
		return fmt.Errorf("failed to normalize sub_results: %w", err)
	}
	query := `
		INSERT INTO tasks (
			id, status, mode, user_input, intention, plan, tasks_list,
			sub_results, final_output, last_node, retry_count, error_msg,
			thinking_enabled, plan_enabled, subagent_enabled,
			create_time, update_time, finish_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err = db.Exec(query,
		task.ID, task.Status, task.Mode, task.UserInput, task.Intention, task.Plan,
		tasksList, subResults, task.FinalOutput, task.LastNode,
		task.RetryCount, task.ErrorMsg, task.ThinkingEnabled, task.PlanEnabled,
		task.SubagentEnabled, task.CreateTime, task.UpdateTime, task.FinishTime,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	logger.Debug("task created", "task_id", task.ID)
	return nil
}

// GetTaskByID 根据 ID 获取任务
func GetTaskByID(taskID string) (*Task, error) {
	db := db.Get()
	query := `
		SELECT id, status, mode, user_input, intention, plan, tasks_list,
			   sub_results, final_output, last_node, retry_count, error_msg,
			   thinking_enabled, plan_enabled, subagent_enabled,
			   create_time, update_time, finish_time
		FROM tasks WHERE id = $1
	`
	var task Task
	err := db.QueryRow(query, taskID).Scan(
		&task.ID, &task.Status, &task.Mode, &task.UserInput, &task.Intention,
		&task.Plan, &task.TasksList, &task.SubResults, &task.FinalOutput,
		&task.LastNode, &task.RetryCount, &task.ErrorMsg, &task.ThinkingEnabled,
		&task.PlanEnabled, &task.SubagentEnabled, &task.CreateTime,
		&task.UpdateTime, &task.FinishTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(taskID, status string) error {
	db := db.Get()
	query := `UPDATE tasks SET status = $1, update_time = NOW() WHERE id = $2`
	_, err := db.Exec(query, status, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

// UpdateTaskResult 更新任务结果
func UpdateTaskResult(taskID, intention, plan, finalOutput string) error {
	db := db.Get()
	query := `
		UPDATE tasks SET
			intention = $1,
			plan = $2,
			final_output = $3,
			update_time = NOW()
		WHERE id = $4
	`
	_, err := db.Exec(query, intention, plan, finalOutput, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}
	return nil
}

// UpdateTaskLastNode 更新任务最后节点（断点续跑）
func UpdateTaskLastNode(taskID, lastNode string) error {
	db := db.Get()
	query := `UPDATE tasks SET last_node = $1, update_time = NOW() WHERE id = $2`
	_, err := db.Exec(query, lastNode, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task last node: %w", err)
	}
	return nil
}

// UpdateTaskError 更新任务错误信息
func UpdateTaskError(taskID, errorMsg string, retryCount int) error {
	db := db.Get()
	query := `
		UPDATE tasks SET
			error_msg = $1,
			retry_count = $2,
			update_time = NOW()
		WHERE id = $3
	`
	_, err := db.Exec(query, errorMsg, retryCount, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task error: %w", err)
	}
	return nil
}

// CompleteTask 完成任务
func CompleteTask(taskID, finalOutput string) error {
	db := db.Get()
	query := `
		UPDATE tasks SET
			status = 'success',
			final_output = $1,
			finish_time = NOW(),
			update_time = NOW()
		WHERE id = $2
	`
	_, err := db.Exec(query, finalOutput, taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}
	return nil
}

// FailTask 标记任务失败
func FailTask(taskID, errorMsg string) error {
	db := db.Get()
	query := `
		UPDATE tasks SET
			status = 'failed',
			error_msg = $1,
			finish_time = NOW(),
			update_time = NOW()
		WHERE id = $2
	`
	_, err := db.Exec(query, errorMsg, taskID)
	if err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}
	return nil
}

// ListTasks 获取任务列表
func ListTasks(page, size int, status string) ([]*Task, int, error) {
	db := db.Get()

	// 构建查询条件
	whereClause := ""
	args := []any{}
	argIndex := 1

	if status != "" {
		whereClause = fmt.Sprintf("WHERE status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// 获取总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// 获取任务列表
	offset := (page - 1) * size
	query := fmt.Sprintf(`
		SELECT id, status, mode, user_input, intention, plan, tasks_list,
			   sub_results, final_output, last_node, retry_count, error_msg,
			   thinking_enabled, plan_enabled, subagent_enabled,
			   create_time, update_time, finish_time
		FROM tasks %s
		ORDER BY create_time DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, size, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID, &task.Status, &task.Mode, &task.UserInput, &task.Intention,
			&task.Plan, &task.TasksList, &task.SubResults, &task.FinalOutput,
			&task.LastNode, &task.RetryCount, &task.ErrorMsg, &task.ThinkingEnabled,
			&task.PlanEnabled, &task.SubagentEnabled, &task.CreateTime,
			&task.UpdateTime, &task.FinishTime,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	return tasks, total, nil
}

// UpdateTasksList 更新子任务列表
func UpdateTasksList(taskID string, tasksList json.RawMessage) error {
	db := db.Get()
	normalized, err := normalizeJSONValue(tasksList)
	if err != nil {
		return fmt.Errorf("failed to normalize tasks list: %w", err)
	}
	query := `UPDATE tasks SET tasks_list = $1, update_time = NOW() WHERE id = $2`
	_, err = db.Exec(query, normalized, taskID)
	if err != nil {
		return fmt.Errorf("failed to update tasks list: %w", err)
	}
	return nil
}

// UpdateSubResults 更新子 Agent 结果
func UpdateSubResults(taskID string, subResults json.RawMessage) error {
	db := db.Get()
	normalized, err := normalizeJSONValue(subResults)
	if err != nil {
		return fmt.Errorf("failed to normalize sub results: %w", err)
	}
	query := `UPDATE tasks SET sub_results = $1, update_time = NOW() WHERE id = $2`
	_, err = db.Exec(query, normalized, taskID)
	if err != nil {
		return fmt.Errorf("failed to update sub results: %w", err)
	}
	return nil
}

func normalizeJSONValue(raw json.RawMessage) ([]byte, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return []byte("null"), nil
	}
	if !json.Valid(trimmed) {
		return nil, fmt.Errorf("invalid json payload")
	}
	return trimmed, nil
}
