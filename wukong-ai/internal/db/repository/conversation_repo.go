package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jiujuan/wukong-ai/internal/conversation"
	"github.com/jiujuan/wukong-ai/internal/db"
)

// ── Conversation CRUD ────────────────────────────────────────────────────────

// CreateConversation 创建新对话
func CreateConversation(conv *conversation.Conversation) error {
	d := db.Get()
	_, err := d.Exec(`
		INSERT INTO conversations (id, title, summary, turn_count, create_time, update_time)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		conv.ID, conv.Title, conv.Summary, conv.TurnCount,
		conv.CreateTime, conv.UpdateTime,
	)
	return err
}

// GetConversation 按 ID 查询对话
func GetConversation(id string) (*conversation.Conversation, error) {
	d := db.Get()
	var conv conversation.Conversation
	var summary sql.NullString
	err := d.QueryRow(`
		SELECT id, title, summary, turn_count, create_time, update_time
		FROM conversations WHERE id = $1`, id,
	).Scan(&conv.ID, &conv.Title, &summary, &conv.TurnCount,
		&conv.CreateTime, &conv.UpdateTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	conv.Summary = summary.String
	return &conv, nil
}

// ListConversations 列出对话（分页，按更新时间倒序）
func ListConversations(page, size int) ([]*conversation.Conversation, int, error) {
	d := db.Get()
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	var total int
	if err := d.QueryRow(`SELECT COUNT(*) FROM conversations`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := d.Query(`
		SELECT id, title, summary, turn_count, create_time, update_time
		FROM conversations
		ORDER BY update_time DESC
		LIMIT $1 OFFSET $2`, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var convs []*conversation.Conversation
	for rows.Next() {
		var conv conversation.Conversation
		var summary sql.NullString
		if err := rows.Scan(&conv.ID, &conv.Title, &summary, &conv.TurnCount,
			&conv.CreateTime, &conv.UpdateTime); err != nil {
			return nil, 0, err
		}
		conv.Summary = summary.String
		convs = append(convs, &conv)
	}
	return convs, total, nil
}

// UpdateConversationSummary 更新对话摘要（历史压缩时调用）
func UpdateConversationSummary(id, summary string) error {
	d := db.Get()
	_, err := d.Exec(`
		UPDATE conversations SET summary=$1, update_time=$2 WHERE id=$3`,
		summary, time.Now(), id)
	return err
}

// IncrTurnCount 轮次数 +1 并刷新 update_time
func IncrTurnCount(conversationID string) error {
	d := db.Get()
	_, err := d.Exec(`
		UPDATE conversations
		SET turn_count = turn_count + 1, update_time = $1
		WHERE id = $2`, time.Now(), conversationID)
	return err
}

// UpdateConversationTitle 更新对话标题
func UpdateConversationTitle(id, title string) error {
	d := db.Get()
	_, err := d.Exec(`UPDATE conversations SET title=$1, update_time=$2 WHERE id=$3`,
		title, time.Now(), id)
	return err
}

// DeleteConversation 删除对话（级联删除 turns）
func DeleteConversation(id string) error {
	d := db.Get()
	_, err := d.Exec(`DELETE FROM conversations WHERE id=$1`, id)
	return err
}

// ── ConversationTurn CRUD ────────────────────────────────────────────────────

// AddTurn 追加一条轮次记录
func AddTurn(turn *conversation.Turn) (int64, error) {
	d := db.Get()
	var id int64
	err := d.QueryRow(`
		INSERT INTO conversation_turns
			(conversation_id, task_id, turn_index, role, content, full_output, create_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		turn.ConversationID, nullString(turn.TaskID), turn.TurnIndex,
		turn.Role, turn.Content, nullString(turn.FullOutput), turn.CreateTime,
	).Scan(&id)
	return id, err
}

// GetTurns 按对话 ID 获取所有轮次（按 turn_index 升序）
func GetTurns(conversationID string) ([]conversation.Turn, error) {
	d := db.Get()
	rows, err := d.Query(`
		SELECT id, conversation_id, COALESCE(task_id,''), turn_index, role,
		       content, COALESCE(full_output,''), create_time
		FROM conversation_turns
		WHERE conversation_id = $1
		ORDER BY turn_index ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []conversation.Turn
	for rows.Next() {
		var t conversation.Turn
		if err := rows.Scan(&t.ID, &t.ConversationID, &t.TaskID, &t.TurnIndex,
			&t.Role, &t.Content, &t.FullOutput, &t.CreateTime); err != nil {
			return nil, err
		}
		turns = append(turns, t)
	}
	return turns, nil
}

// GetRecentTurns 获取最近 N 轮（用于注入 Prompt）
func GetRecentTurns(conversationID string, limit int) ([]conversation.Turn, error) {
	d := db.Get()
	rows, err := d.Query(`
		SELECT id, conversation_id, COALESCE(task_id,''), turn_index, role,
		       content, COALESCE(full_output,''), create_time
		FROM conversation_turns
		WHERE conversation_id = $1
		ORDER BY turn_index DESC
		LIMIT $2`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []conversation.Turn
	for rows.Next() {
		var t conversation.Turn
		if err := rows.Scan(&t.ID, &t.ConversationID, &t.TaskID, &t.TurnIndex,
			&t.Role, &t.Content, &t.FullOutput, &t.CreateTime); err != nil {
			return nil, err
		}
		turns = append(turns, t)
	}
	// 反转为升序
	for i, j := 0, len(turns)-1; i < j; i, j = i+1, j-1 {
		turns[i], turns[j] = turns[j], turns[i]
	}
	return turns, nil
}

// NextTurnIndex 获取当前最大 turn_index + 1
func NextTurnIndex(conversationID string) (int, error) {
	d := db.Get()
	var maxIdx sql.NullInt64
	err := d.QueryRow(`
		SELECT MAX(turn_index) FROM conversation_turns WHERE conversation_id=$1`,
		conversationID).Scan(&maxIdx)
	if err != nil {
		return 0, fmt.Errorf("NextTurnIndex: %w", err)
	}
	if !maxIdx.Valid {
		return 0, nil
	}
	return int(maxIdx.Int64) + 1, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
