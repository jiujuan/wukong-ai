package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Reader 文件读取工具
type Reader struct {
	allowedPaths []string
}

// NewReader 创建文件读取工具
func NewReader(allowedPaths []string) *Reader {
	return &Reader{
		allowedPaths: allowedPaths,
	}
}

// Name 返回工具名称
func (r *Reader) Name() string {
	return "file_reader"
}

// Description 返回工具描述
func (r *Reader) Description() string {
	return "Read content from a file. Input should be the file path."
}

// Execute 读取文件内容
func (r *Reader) Execute(ctx context.Context, input string) (string, error) {
	// 清理输入路径
	path := strings.TrimSpace(input)

	// 安全检查
	if !r.isPathAllowed(path) {
		return "", &tools.ToolError{
			ToolName: r.Name(),
			Message:  "path not allowed: " + path,
		}
	}

	// 读取文件
	content, err := os.ReadFile(path)
	if err != nil {
		return "", &tools.ToolError{
			ToolName: r.Name(),
			Message:  "failed to read file: " + err.Error(),
		}
	}

	logger.Debug("File read", "path", path, "size", len(content))
	return string(content), nil
}

// isPathAllowed 检查路径是否在允许列表中
func (r *Reader) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, allowed := range r.allowedPaths {
		absAllowed, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absAllowed) {
			return true
		}
	}

	return false
}

// Ensure Reader 实现 Tool 接口
var _ tools.Tool = (*Reader)(nil)
