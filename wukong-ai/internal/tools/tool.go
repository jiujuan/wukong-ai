package tools

import (
	"context"
)

// Tool 接口定义
type Tool interface {
	// Name 返回工具名称
	Name() string
	// Description 返回工具描述
	Description() string
	// Execute 执行工具
	Execute(ctx context.Context, input string) (string, error)
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List 获取所有工具
func (r *ToolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetNames 获取所有工具名称
func (r *ToolRegistry) GetNames() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Execute 执行工具
func (r *ToolRegistry) Execute(ctx context.Context, name, input string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", ErrToolNotFound(name)
	}
	return tool.Execute(ctx, input)
}

// ToolError 工具错误
type ToolError struct {
	ToolName string
	Message  string
}

func (e *ToolError) Error() string {
	return e.ToolName + ": " + e.Message
}

// ErrToolNotFound 工具未找到错误
func ErrToolNotFound(name string) error {
	return &ToolError{
		ToolName: name,
		Message:  "tool not found",
	}
}
