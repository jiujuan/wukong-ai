package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Writer 文件写入工具
type Writer struct {
	allowedPaths []string
}

// NewWriter 创建文件写入工具
func NewWriter(allowedPaths []string) *Writer {
	return &Writer{
		allowedPaths: allowedPaths,
	}
}

// Name 返回工具名称
func (w *Writer) Name() string {
	return "file_writer"
}

// Description 返回工具描述
func (w *Writer) Description() string {
	return "Write content to a file. Input format: <file_path>\n<content>"
}

// Execute 写入文件内容
func (w *Writer) Execute(ctx context.Context, input string) (string, error) {
	// 解析输入：路径和内容用换行分隔
	parts := strings.SplitN(input, "\n", 2)
	if len(parts) < 2 {
		return "", &tools.ToolError{
			ToolName: w.Name(),
			Message:  "invalid input format, expected: <file_path>\\n<content>",
		}
	}

	path := strings.TrimSpace(parts[0])
	content := parts[1]

	// 安全检查
	if !w.isPathAllowed(path) {
		return "", &tools.ToolError{
			ToolName: w.Name(),
			Message:  "path not allowed: " + path,
		}
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", &tools.ToolError{
			ToolName: w.Name(),
			Message:  "failed to create directory: " + err.Error(),
		}
	}

	// 写入文件
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", &tools.ToolError{
			ToolName: w.Name(),
			Message:  "failed to write file: " + err.Error(),
		}
	}

	logger.Debug("File written", "path", path, "size", len(content))
	return "File written successfully: " + path, nil
}

// isPathAllowed 检查路径是否在允许列表中
func (w *Writer) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, allowed := range w.allowedPaths {
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

// Ensure Writer 实现 Tool 接口
var _ tools.Tool = (*Writer)(nil)
