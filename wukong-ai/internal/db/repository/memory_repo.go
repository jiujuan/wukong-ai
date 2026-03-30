package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/pgvector/pgvector-go"
)

// Memory 记忆结构体
type Memory struct {
	ID         int64           `json:"id"`
	TaskID     sql.NullString  `json:"task_id"`
	SessionID  sql.NullString  `json:"session_id"`
	Content    string          `json:"content"`
	Embedding  pgvector.Vector `json:"embedding"`
	MemoryType string          `json:"memory_type"`
	Metadata   json.RawMessage `json:"metadata"`
	CreateTime string          `json:"create_time"`
}

// SaveMemory 保存记忆
func SaveMemory(memory *Memory) (int64, error) {
	db := db.Get()
	query := `
		INSERT INTO memories (
			task_id, session_id, content, embedding, memory_type, metadata, create_time
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id
	`
	var id int64
	err := db.QueryRow(query,
		memory.TaskID, memory.SessionID, memory.Content,
		memory.Embedding, memory.MemoryType, memory.Metadata,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to save memory: %w", err)
	}
	logger.Debug("memory saved", "memory_id", id)
	return id, nil
}

// SearchMemoriesByEmbedding 根据向量搜索相似记忆
func SearchMemoriesByEmbedding(embedding pgvector.Vector, taskID string, topK int) ([]*Memory, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, session_id, content, embedding, memory_type, metadata, create_time
		FROM memories
		WHERE task_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3
	`
	rows, err := db.Query(query, taskID, embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.TaskID, &m.SessionID, &m.Content,
			&m.Embedding, &m.MemoryType, &m.Metadata, &m.CreateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		memories = append(memories, &m)
	}

	return memories, nil
}

// SearchMemoriesBySession 根据 session 搜索记忆
func SearchMemoriesBySession(sessionID string, topK int) ([]*Memory, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, session_id, content, embedding, memory_type, metadata, create_time
		FROM memories
		WHERE session_id = $1
		ORDER BY create_time DESC
		LIMIT $2
	`
	rows, err := db.Query(query, sessionID, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories by session: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.TaskID, &m.SessionID, &m.Content,
			&m.Embedding, &m.MemoryType, &m.Metadata, &m.CreateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		memories = append(memories, &m)
	}

	return memories, nil
}

// GetMemoriesByTaskID 获取任务的所有记忆
func GetMemoriesByTaskID(taskID string) ([]*Memory, error) {
	db := db.Get()
	query := `
		SELECT id, task_id, session_id, content, embedding, memory_type, metadata, create_time
		FROM memories
		WHERE task_id = $1
		ORDER BY create_time ASC
	`
	rows, err := db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		var m Memory
		err := rows.Scan(
			&m.ID, &m.TaskID, &m.SessionID, &m.Content,
			&m.Embedding, &m.MemoryType, &m.Metadata, &m.CreateTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan memory: %w", err)
		}
		memories = append(memories, &m)
	}

	return memories, nil
}

// DeleteMemoriesByTaskID 删除任务的所有记忆
func DeleteMemoriesByTaskID(taskID string) error {
	db := db.Get()
	query := `DELETE FROM memories WHERE task_id = $1`
	_, err := db.Exec(query, taskID)
	if err != nil {
		return fmt.Errorf("failed to delete memories: %w", err)
	}
	return nil
}
