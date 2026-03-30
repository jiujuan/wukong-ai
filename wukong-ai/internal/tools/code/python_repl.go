package code

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// PythonREPL Python REPL 执行工具
type PythonREPL struct {
	sandboxDir string
	enabled    bool
}

// NewPythonREPL 创建 Python REPL 工具
func NewPythonREPL(sandboxDir string, enabled bool) *PythonREPL {
	return &PythonREPL{
		sandboxDir: sandboxDir,
		enabled:    enabled,
	}
}

// Name 返回工具名称
func (p *PythonREPL) Name() string {
	return "python_repl"
}

// Description 返回工具描述
func (p *PythonREPL) Description() string {
	return "Execute Python code in a sandboxed environment. Input: Python code to execute."
}

// Execute 执行 Python 代码
func (p *PythonREPL) Execute(ctx context.Context, taskID, input string) (string, error) {
	if !p.enabled {
		return "", &tools.ToolError{
			ToolName: p.Name(),
			Message:  "Python REPL is disabled",
		}
	}

	// 创建临时目录
	dir := filepath.Join(p.sandboxDir, taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", &tools.ToolError{
			ToolName: p.Name(),
			Message:  "failed to create sandbox directory: " + err.Error(),
		}
	}

	// 创建临时文件
	tmpFile := filepath.Join(dir, "temp_script.py")
	if err := os.WriteFile(tmpFile, []byte(input), 0644); err != nil {
		return "", &tools.ToolError{
			ToolName: p.Name(),
			Message:  "failed to write script: " + err.Error(),
		}
	}

	// 执行 Python 代码
	cmd := exec.CommandContext(ctx, "python3", tmpFile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &tools.ToolError{
			ToolName: p.Name(),
			Message:  fmt.Sprintf("execution failed: %v\nstderr: %s", err, stderr.String()),
		}
	}

	result := stdout.String()
	if stderr.Len() > 0 {
		result += "\nStderr:\n" + stderr.String()
	}

	logger.Debug("Python REPL executed", "task_id", taskID, "output_length", len(result))
	return result, nil
}

// ExecuteWithTimeout 带超时的执行
func (p *PythonREPL) ExecuteWithTimeout(ctx context.Context, taskID, input string, timeoutSeconds int) (string, error) {
	// 创建超时上下文
	execCtx, cancel := context.WithTimeout(ctx, context.Duration(timeoutSeconds)*1000000000)
	defer cancel()

	return p.Execute(execCtx, taskID, input)
}

// Ensure PythonREPL 实现 Tool 接口
var _ tools.Tool = (*PythonREPL)(nil)
