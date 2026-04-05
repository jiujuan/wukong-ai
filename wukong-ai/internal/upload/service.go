package upload

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// UploadService 文件存储服务
type UploadService struct {
	baseDir   string // 上传根目录，默认 "./uploads"
	validator *FileValidator
}

// NewUploadService 创建存储服务
func NewUploadService(baseDir string) *UploadService {
	if baseDir == "" {
		baseDir = "./uploads"
	}
	return &UploadService{
		baseDir:   baseDir,
		validator: NewFileValidator(),
	}
}

// Save 校验并持久化文件，返回 task_attachments 记录
func (s *UploadService) Save(
	ctx context.Context,
	taskID string,
	header *multipart.FileHeader,
	file multipart.File,
) (*repository.TaskAttachment, error) {

	mimeType := NormalizeMIME(header.Header.Get("Content-Type"))

	// 按 task_id 隔离目录
	dir := filepath.Join(s.baseDir, taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	// 防文件名冲突：时间戳前缀
	safeFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	filePath := filepath.Join(dir, safeFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 写 DB 记录
	att := &repository.TaskAttachment{
		TaskID:        taskID,
		FileName:      header.Filename,
		FilePath:      filePath,
		MimeType:      mimeType,
		FileSize:      header.Size,
		ExtractStatus: "pending",
		IsImage:       IsImage(mimeType),
		UploadTime:    time.Now(),
	}
	id, err := repository.CreateAttachment(att)
	if err != nil {
		// 清理已写的文件
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save attachment record: %w", err)
	}
	att.ID = id
	logger.Info("file uploaded",
		"task_id", taskID,
		"file_name", header.Filename,
		"mime", mimeType,
		"size", header.Size,
		"attachment_id", id,
	)
	return att, nil
}
