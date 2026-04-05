package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db"
)

// TaskAttachment 附件元信息
type TaskAttachment struct {
	ID            int64        `json:"id"`
	TaskID        string       `json:"task_id"`
	FileName      string       `json:"file_name"`
	FilePath      string       `json:"file_path"`
	MimeType      string       `json:"mime_type"`
	FileSize      int64        `json:"file_size"`
	ExtractStatus string       `json:"extract_status"` // pending/extracting/done/failed
	IsImage       bool         `json:"is_image"`
	ChunkCount    int          `json:"chunk_count"`
	ErrorMsg      string       `json:"error_msg,omitempty"`
	UploadTime    time.Time    `json:"upload_time"`
	ExtractTime   sql.NullTime `json:"extract_time,omitempty"`
}

// CreateAttachment 插入附件记录，返回自增 ID
func CreateAttachment(att *TaskAttachment) (int64, error) {
	d := db.Get()
	var id int64
	err := d.QueryRow(`
		INSERT INTO task_attachments
			(task_id, file_name, file_path, mime_type, file_size,
			 extract_status, is_image, chunk_count, upload_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`,
		att.TaskID, att.FileName, att.FilePath, att.MimeType, att.FileSize,
		att.ExtractStatus, att.IsImage, att.ChunkCount, att.UploadTime,
	).Scan(&id)
	return id, err
}

// GetAttachment 按 ID 查询附件
func GetAttachment(id int64) (*TaskAttachment, error) {
	d := db.Get()
	att := &TaskAttachment{}
	err := d.QueryRow(`
		SELECT id, task_id, file_name, file_path, mime_type, file_size,
		       extract_status, is_image, chunk_count,
		       COALESCE(error_msg,''), upload_time, extract_time
		FROM task_attachments WHERE id=$1`, id,
	).Scan(&att.ID, &att.TaskID, &att.FileName, &att.FilePath,
		&att.MimeType, &att.FileSize, &att.ExtractStatus,
		&att.IsImage, &att.ChunkCount, &att.ErrorMsg,
		&att.UploadTime, &att.ExtractTime)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return att, err
}

// GetAttachmentsByTaskID 获取任务的所有附件
func GetAttachmentsByTaskID(taskID string) ([]*TaskAttachment, error) {
	d := db.Get()
	rows, err := d.Query(`
		SELECT id, task_id, file_name, file_path, mime_type, file_size,
		       extract_status, is_image, chunk_count,
		       COALESCE(error_msg,''), upload_time, extract_time
		FROM task_attachments
		WHERE task_id=$1
		ORDER BY upload_time ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*TaskAttachment
	for rows.Next() {
		att := &TaskAttachment{}
		if err := rows.Scan(&att.ID, &att.TaskID, &att.FileName, &att.FilePath,
			&att.MimeType, &att.FileSize, &att.ExtractStatus,
			&att.IsImage, &att.ChunkCount, &att.ErrorMsg,
			&att.UploadTime, &att.ExtractTime); err != nil {
			return nil, err
		}
		list = append(list, att)
	}
	return list, nil
}

// UpdateAttachmentStatus 更新提取状态
func UpdateAttachmentStatus(id int64, status, errMsg string, chunkCount int) error {
	d := db.Get()
	if status == "done" || status == "failed" {
		_, err := d.Exec(`
			UPDATE task_attachments
			SET extract_status=$1, error_msg=$2, chunk_count=$3, extract_time=NOW()
			WHERE id=$4`,
			status, errMsg, chunkCount, id)
		return err
	}
	_, err := d.Exec(`
		UPDATE task_attachments SET extract_status=$1, error_msg=$2 WHERE id=$3`,
		status, errMsg, id)
	return err
}

// GetPendingAttachments 获取所有待提取的附件（服务重启恢复用）
func GetPendingAttachments() ([]*TaskAttachment, error) {
	d := db.Get()
	rows, err := d.Query(`
		SELECT id, task_id, file_name, file_path, mime_type, file_size,
		       extract_status, is_image, chunk_count,
		       COALESCE(error_msg,''), upload_time, extract_time
		FROM task_attachments
		WHERE extract_status IN ('pending','extracting')
		ORDER BY upload_time ASC`)
	if err != nil {
		return nil, fmt.Errorf("GetPendingAttachments: %w", err)
	}
	defer rows.Close()

	var list []*TaskAttachment
	for rows.Next() {
		att := &TaskAttachment{}
		if err := rows.Scan(&att.ID, &att.TaskID, &att.FileName, &att.FilePath,
			&att.MimeType, &att.FileSize, &att.ExtractStatus,
			&att.IsImage, &att.ChunkCount, &att.ErrorMsg,
			&att.UploadTime, &att.ExtractTime); err != nil {
			return nil, err
		}
		list = append(list, att)
	}
	return list, nil
}
