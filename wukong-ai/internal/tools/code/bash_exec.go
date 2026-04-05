package code

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// BashExec Bash 执行工具
type BashExec struct {
	sandboxDir string
	enabled    bool
}

// NewBashExec 创建 Bash 执行工具
func NewBashExec(sandboxDir string, enabled bool) *BashExec {
	return &BashExec{
		sandboxDir: sandboxDir,
		enabled:    enabled,
	}
}

// Name 返回工具名称
func (b *BashExec) Name() string { return "bash_exec" }

// Description 返回工具描述
func (b *BashExec) Description() string {
	return "Execute bash commands in a sandboxed environment. Input: bash command to execute."
}

// Execute 执行 Bash 命令（符合 tools.Tool 接口：只接受 ctx + input）
// input 格式：直接是 bash 命令字符串
func (b *BashExec) Execute(ctx context.Context, input string) (string, error) {
	if !b.enabled {
		return "", &tools.ToolError{ToolName: b.Name(), Message: "Bash execution is disabled"}
	}

	command := strings.TrimSpace(input)
	if b.isDangerousCommand(command) {
		return "", &tools.ToolError{ToolName: b.Name(), Message: "command not allowed for security reasons"}
	}

	// 使用通用沙箱目录（不再依赖 taskID）
	dir := b.sandboxDir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", &tools.ToolError{ToolName: b.Name(), Message: "failed to create sandbox directory: " + err.Error()}
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &tools.ToolError{
			ToolName: b.Name(),
			Message:  fmt.Sprintf("execution failed: %v\nstderr: %s", err, stderr.String()),
		}
	}

	result := stdout.String()
	if stderr.Len() > 0 {
		result += "\nStderr:\n" + stderr.String()
	}
	logger.Debug("Bash executed", "output_length", len(result))
	return result, nil
}

// ExecuteInDir 带 taskID 目录隔离的扩展调用（供内部直接调用，不通过 Tool 接口）
func (b *BashExec) ExecuteInDir(ctx context.Context, taskID, input string) (string, error) {
	if !b.enabled {
		return "", &tools.ToolError{ToolName: b.Name(), Message: "Bash execution is disabled"}
	}
	command := strings.TrimSpace(input)
	if b.isDangerousCommand(command) {
		return "", &tools.ToolError{ToolName: b.Name(), Message: "command not allowed for security reasons"}
	}

	dir := filepath.Join(b.sandboxDir, taskID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", &tools.ToolError{ToolName: b.Name(), Message: "failed to create sandbox directory: " + err.Error()}
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TASK_ID="+taskID)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &tools.ToolError{
			ToolName: b.Name(),
			Message:  fmt.Sprintf("execution failed: %v\nstderr: %s", err, stderr.String()),
		}
	}

	result := stdout.String()
	if stderr.Len() > 0 {
		result += "\nStderr:\n" + stderr.String()
	}
	logger.Debug("Bash executed", "task_id", taskID, "output_length", len(result))
	return result, nil
}

func (b *BashExec) isDangerousCommand(command string) bool {
	dangerousPatterns := []string{
		"rm -rf /", "rm -rf ~", ":(){ :|:& };:",
		"dd if=", "mkfs", "fdisk",
		"curl | sh", "wget | sh", "curl -s", "wget -O-",
	}
	lower := strings.ToLower(command)
	for _, p := range dangerousPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// 编译期接口检查
var _ tools.Tool = (*BashExec)(nil)
