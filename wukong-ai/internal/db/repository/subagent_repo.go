package repository

import (
	"database/sql"
	"fmt"

	"github.com/jiujuan/wukong-ai/internal/db"
)

// SubAgentRecord 子 Agent 执行记录
type SubAgentRecord struct {
	ID         int64          `json:"id"`
	TaskID     string         `json:"task_id"`
	AgentIndex int            `json:"agent_index"`
	SubTask    string         `json:"sub_task"`
	Result     sql.NullString `json:"result"`
	Status     string         `json:"status"`
	RetryCount int            `json:"retry_count"`
	ErrorMsg   sql.NullString `json:"error_msg"`
	StartTime  sql.NullTime   `json:"start_time"`
	EndTime    sql.NullTime   `json:"end_time"`
}

// CreateSubAgentRecord 创建子 Agent 记录
func CreateSubAgentRecord(record *SubAgentRecord) (int64, error) {
	db := db.Get()
	query := `
		INSERT INTO sub_agents (task_id, agent_index, sub_task, result, status, retry_count, error_msg, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var id int64
	err := db.QueryRow(query,
		record.TaskID, record.AgentIndex, record.SubTask, record.Result,
		record.Status, record.RetryCount, record.ErrorMsg, record.StartTime, record.EndTime,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create sub agent record: %w", err)
	}
	return id, nil
}

// UpdateSubAgentRecord 更新子 Agent 记录
func UpdateSubAgentRecord(id int64, status, result, errorMsg string) error {
	db := db.Get()
	query := `
		UPDATE sub_agents SET
			status = $1,
			result = $2,
			error_msg = $3,
			end_time = NOW()
		WHERE id = $4
	`
	_, err := db.Exec(query, status, result, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update sub agent record: %w", err)
	}
	return nil
}

// GetSubAgentRecordsByTaskID 获取任务的所有子 Agent 记录
func GetSubAgentRecordsByTaskID(taskID string) ([]*SubAgentRecord, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, agent_index, sub_task, result, status, retry_count, error_msg, start_time, end_time
		FROM sub_agents
		WHERE task_id = $1
		ORDER BY agent_index ASC
	`
	rows, err := db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub agent records: %w", err)
	}
	defer rows.Close()

	var records []*SubAgentRecord
	for rows.Next() {
		var record SubAgentRecord
		err := rows.Scan(
			&record.ID, &record.TaskID, &record.AgentIndex, &record.SubTask,
			&record.Result, &record.Status, &record.RetryCount,
			&record.ErrorMsg, &record.StartTime, &record.EndTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sub agent record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}
