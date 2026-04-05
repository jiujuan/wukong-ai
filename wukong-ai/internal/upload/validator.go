package upload

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
)

// 允许的 MIME 类型白名单
var allowedMIME = map[string]bool{
	"text/plain":    true,
	"text/markdown": true,
	"text/csv":      true,
	"application/pdf": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true, // .xlsx
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
	"image/gif":  true,
}

// 文件魔数（Magic Bytes）——防止篡改扩展名伪装
var magicBytes = map[string][]byte{
	"application/pdf": {0x25, 0x50, 0x44, 0x46},    // %PDF
	"image/png":       {0x89, 0x50, 0x4E, 0x47},    // PNG
	"image/jpeg":      {0xFF, 0xD8, 0xFF},           // JPEG
	"image/webp":      {0x52, 0x49, 0x46, 0x46},    // RIFF (WEBP)
	"image/gif":       {0x47, 0x49, 0x46, 0x38},    // GIF8
	// docx / xlsx 本质是 ZIP
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {0x50, 0x4B, 0x03, 0x04},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       {0x50, 0x4B, 0x03, 0x04},
}

const (
	MaxDocumentSize int64 = 50 * 1024 * 1024 // 文档类最大 50 MB
	MaxImageSize    int64 = 20 * 1024 * 1024 // 图片类最大 20 MB
	MaxFileCount          = 10               // 单次任务最多 10 个文件
)

// FileValidator 文件安全校验器
type FileValidator struct{}

// NewFileValidator 创建校验器
func NewFileValidator() *FileValidator {
	return &FileValidator{}
}

// Validate 执行三层校验：大小 → MIME 白名单 → 魔数验证
func (v *FileValidator) Validate(header *multipart.FileHeader, file multipart.File) error {
	// 1. 文件大小
	maxSize := MaxDocumentSize
	declaredMIME := header.Header.Get("Content-Type")
	// Content-Type 可能带 charset 参数，取第一段
	if idx := strings.Index(declaredMIME, ";"); idx > 0 {
		declaredMIME = strings.TrimSpace(declaredMIME[:idx])
	}

	if strings.HasPrefix(declaredMIME, "image/") {
		maxSize = MaxImageSize
	}
	if header.Size > maxSize {
		return fmt.Errorf("file too large: %d bytes (max %d)", header.Size, maxSize)
	}

	// 2. MIME 白名单
	if !allowedMIME[declaredMIME] {
		return fmt.Errorf("unsupported file type: %s", declaredMIME)
	}

	// 3. 魔数验证（防扩展名伪造）
	if magic, ok := magicBytes[declaredMIME]; ok {
		buf := make([]byte, 8)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read file header: %w", err)
		}
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}
		if n < len(magic) || !bytes.HasPrefix(buf[:n], magic) {
			return fmt.Errorf("file content does not match declared type %s", declaredMIME)
		}
	}

	return nil
}

// NormalizeMIME 标准化 Content-Type（去掉 charset 等参数）
func NormalizeMIME(contentType string) string {
	if idx := strings.Index(contentType, ";"); idx > 0 {
		return strings.TrimSpace(contentType[:idx])
	}
	return strings.TrimSpace(contentType)
}

// IsImage 判断 MIME 是否为图片类型
func IsImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}
