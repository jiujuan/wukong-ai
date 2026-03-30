package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// NodeExecutionLog 节点执行日志
type NodeExecutionLog struct {
	ID          int64          `json:"id"`
	TaskID      string         `json:"task_id"`
	NodeName    string         `json:"node_name"`
	Status      string         `json:"status"`
	Input       sql.NullString `json:"input"`
	Output      sql.NullString `json:"output"`
	ErrorMsg    sql.NullString `json:"error_msg"`
	DurationMs  sql.NullInt64  `json:"duration_ms"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     sql.NullTime   `json:"end_time"`
}

// CreateNodeLog 创建节点执行日志
func CreateNodeLog(log *NodeExecutionLog) (int64, error) {
	db := db.Get()
	query := `
		INSERT INTO node_execution_logs (
			task_id, node_name, status, input, output, error_msg,
			duration_ms, start_time, end_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var id int64
	err := db.QueryRow(query,
		log.TaskID, log.NodeName, log.Status, log.Input, log.Output,
		log.ErrorMsg, log.DurationMs, log.StartTime, log.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create node log: %w", err)
	}
	logger.Debug("node log created", "log_id", id, "task_id", log.TaskID, "node", log.NodeName)
	return id, nil
}

// UpdateNodeLog 更新节点执行日志
func UpdateNodeLog(id int64, status, output, errorMsg string, durationMs int64) error {
	db := db.Get()
	query := `
		UPDATE node_execution_logs SET
			status = $1,
			output = $2,
			error_msg = $3,
			duration_ms = $4,
			end_time = NOW()
		WHERE id = $5
	`
	_, err := db.Exec(query, status, output, errorMsg, durationMs, id)
	if err != nil {
		return fmt.Errorf("failed to update node log: %w", err)
	}
	return nil
}

// GetNodeLogsByTaskID 获取任务的所有节点日志
func GetNodeLogsByTaskID(taskID string) ([]*NodeExecutionLog, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, node_name, status, input, output, error_msg,
			   duration_ms, start_time, end_time
		FROM node_execution_logs
		WHERE task_id = $1
		ORDER BY start_time ASC
	`
	rows, err := db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node logs: %w", err)
	}
	defer rows.Close()

	var logs []*NodeExecutionLog
	for rows.Next() {
		var log NodeExecutionLog
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.NodeName, &log.Status, &log.Input,
			&log.Output, &log.ErrorMsg, &log.DurationMs, &log.StartTime, &log.EndTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetLatestNodeLog 获取任务最后执行的节点日志
func GetLatestNodeLog(taskID string) (*NodeExecutionLog, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, node_name, status, input, output, error_msg,
			   duration_ms, start_time, end_time
		FROM node_execution_logs
		WHERE task_id = $1
		ORDER BY end_time DESC NULLS LAST
		LIMIT 1
	`
	var log NodeExecutionLog
	err := db.QueryRow(query, taskID).Scan(
		&log.ID, &log.TaskID, &log.NodeName, &log.Status, &log.Input,
		&log.Output, &log.ErrorMsg, &log.DurationMs, &log.StartTime, &log.EndTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest node log: %w", err)
	}
	return &log, nil
}
