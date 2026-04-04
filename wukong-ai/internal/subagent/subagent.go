package subagent

import (
	"fmt"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// SubAgent 子代理结构体
type SubAgent struct {
	llmProvider  llm.LLM
	promptDir    string
	toolRegistry *tools.ToolRegistry // 按需调用工具
}

// SubTask 子任务
type SubTask struct {
	ID     int64  `json:"id"`
	TaskID string `json:"task_id"`
	Index  int    `json:"index"`
	Input  string `json:"input"`
}

// NewSubAgent 创建新的子代理
func NewSubAgent(llmProvider llm.LLM, promptDir string, toolRegistry *tools.ToolRegistry) *SubAgent {
	return &SubAgent{
		llmProvider:  llmProvider,
		promptDir:    promptDir,
		toolRegistry: toolRegistry,
	}
}

// Execute 执行子代理任务：先用工具搜集资料，再交给 LLM 分析
func (s *SubAgent) Execute(task *SubTask) (string, error) {
	logger.Debug("SubAgent executing", "task_id", task.TaskID, "index", task.Index)

	systemPrompt := prompts.LoadPrompt(s.promptDir, "subagent.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a SubAgent executing a specific subtask.
Your role is to:
1. Complete the assigned subtask thoroughly
2. Provide detailed and accurate results
3. Follow best practices for the task type
4. Report your findings clearly

Execute the task and provide your results.`
	}

	// ── Step 1：搜索工具补充资料 ─────────────────────────────────
	searchContext := s.runSearchTool(task)

	// ── Step 2：携带工具结果调用 LLM ─────────────────────────────
	var userContent strings.Builder
	userContent.WriteString(task.Input)
	if searchContext != "" {
		userContent.WriteString("\n\nSearch Results (use as reference):\n" + searchContext)
	}

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent.String()},
	}

	response, err := s.llmProvider.ChatWithHistory(nil, messages)
	if err != nil {
		logger.Error("SubAgent LLM call failed", "task_id", task.TaskID, "index", task.Index, "err", err)
		return "", err
	}

	logger.Debug("SubAgent completed", "task_id", task.TaskID, "index", task.Index)
	return response, nil
}

// runSearchTool 调用可用搜索工具为子任务补充信息
func (s *SubAgent) runSearchTool(task *SubTask) string {
	if s.toolRegistry == nil {
		return ""
	}
	for _, name := range []string{"tavily_search", "duckduckgo_search", "mock_search"} {
		tool, ok := s.toolRegistry.Get(name)
		if !ok {
			continue
		}
		result, err := tool.Execute(nil, fmt.Sprintf("%s", task.Input))
		if err != nil {
			logger.Warn("SubAgent search failed", "tool", name, "task_id", task.TaskID, "err", err)
			continue
		}
		logger.Debug("SubAgent search done", "tool", name, "task_id", task.TaskID, "index", task.Index)
		return result
	}
	return ""
}
