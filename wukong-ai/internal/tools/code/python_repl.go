package code

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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
	return &PythonREPL{sandboxDir: sandboxDir, enabled: enabled}
}

// Name 返回工具名称
func (p *PythonREPL) Name() string { return "python_repl" }

// Description 返回工具描述
func (p *PythonREPL) Description() string {
	return "Execute Python code in a sandboxed environment. Input: Python code to execute."
}

// Execute 执行 Python 代码（符合 tools.Tool 接口：只接受 ctx + input）
// input：Python 源代码字符串
func (p *PythonREPL) Execute(ctx context.Context, input string) (string, error) {
	if !p.enabled {
		return "", &tools.ToolError{ToolName: p.Name(), Message: "Python REPL is disabled"}
	}

	dir := p.sandboxDir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", &tools.ToolError{ToolName: p.Name(), Message: "failed to create sandbox directory: " + err.Error()}
	}

	tmpFile := filepath.Join(dir, "temp_script.py")
	if err := os.WriteFile(tmpFile, []byte(input), 0644); err != nil {
		return "", &tools.ToolError{ToolName: p.Name(), Message: "failed to write script: " + err.Error()}
	}
	defer os.Remove(tmpFile)

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
	logger.Debug("Python REPL executed", "output_length", len(result))
	return result, nil
}

// ExecuteInDir 带 taskID 目录隔离的扩展调用（供内部直接调用，不通过 Tool 接口）
func (p *PythonREPL) ExecuteInDir(ctx context.Context, taskID, input string) (string, error) {
	if !p.enabled {
		return "", &tools.ToolError{ToolName: p.Name(), Message: "Python REPL is disabled"}
	}

	dir := filepath.Join(p.sandboxDir, taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", &tools.ToolError{ToolName: p.Name(), Message: "failed to create sandbox directory: " + err.Error()}
	}

	tmpFile := filepath.Join(dir, "temp_script.py")
	if err := os.WriteFile(tmpFile, []byte(input), 0644); err != nil {
		return "", &tools.ToolError{ToolName: p.Name(), Message: "failed to write script: " + err.Error()}
	}
	defer os.Remove(tmpFile)

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

// ExecuteWithTimeout 带超时的执行（使用标准 time.Duration）
func (p *PythonREPL) ExecuteWithTimeout(ctx context.Context, input string, timeout time.Duration) (string, error) {
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return p.Execute(execCtx, input)
}

// 编译期接口检查
var _ tools.Tool = (*PythonREPL)(nil)
