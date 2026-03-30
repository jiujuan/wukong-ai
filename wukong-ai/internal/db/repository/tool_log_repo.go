package repository

import (
	"database/sql"
	"fmt"

	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// ToolCallLog 工具调用日志
type ToolCallLog struct {
	ID         int64          `json:"id"`
	TaskID     string         `json:"task_id"`
	NodeName   string         `json:"node_name"`
	ToolName   string         `json:"tool_name"`
	Input      sql.NullString `json:"input"`
	Output     sql.NullString `json:"output"`
	Success    bool           `json:"success"`
	ErrorMsg   sql.NullString `json:"error_msg"`
	DurationMs sql.NullInt64  `json:"duration_ms"`
	CallTime   sql.NullString `json:"call_time"`
}

// CreateToolLog 创建工具调用日志
func CreateToolLog(log *ToolCallLog) (int64, error) {
	db := db.Get()
	query := `
		INSERT INTO tool_call_logs (
			task_id, node_name, tool_name, input, output,
			success, error_msg, duration_ms, call_time
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id
	`
	var id int64
	err := db.QueryRow(query,
		log.TaskID, log.NodeName, log.ToolName, log.Input, log.Output,
		log.Success, log.ErrorMsg, log.DurationMs,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create tool log: %w", err)
	}
	logger.Debug("tool log created", "log_id", id, "tool", log.ToolName)
	return id, nil
}

// GetToolLogsByTaskID 获取任务的所有工具调用日志
func GetToolLogsByTaskID(taskID string) ([]*ToolCallLog, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, node_name, tool_name, input, output,
			   success, error_msg, duration_ms, call_time
		FROM tool_call_logs
		WHERE task_id = $1
		ORDER BY call_time ASC
	`
	rows, err := db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool logs: %w", err)
	}
	defer rows.Close()

	var logs []*ToolCallLog
	for rows.Next() {
		var log ToolCallLog
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.NodeName, &log.ToolName,
			&log.Input, &log.Output, &log.Success, &log.ErrorMsg,
			&log.DurationMs, &log.CallTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tool log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// UpdateToolLog 更新工具调用日志
func UpdateToolLog(id int64, success bool, output, errorMsg string, durationMs int64) error {
	db := db.Get()
	query := `
		UPDATE tool_call_logs SET
			success = $1,
			output = $2,
			error_msg = $3,
			duration_ms = $4
		WHERE id = $5
	`
	_, err := db.Exec(query, success, output, errorMsg, durationMs, id)
	if err != nil {
		return fmt.Errorf("failed to update tool log: %w", err)
	}
	return nil
}
