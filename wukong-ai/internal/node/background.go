package node

import (
	"fmt"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Background 后台任务节点 - 收集背景信息，按需调用搜索工具
type Background struct {
	llmProvider  llm.LLM
	promptDir    string
	toolRegistry *tools.ToolRegistry
}

// NewBackground 创建后台任务节点
func NewBackground(llmProvider llm.LLM, promptDir string, toolRegistry *tools.ToolRegistry) *Background {
	return &Background{
		llmProvider:  llmProvider,
		promptDir:    promptDir,
		toolRegistry: toolRegistry,
	}
}

// Name 返回节点名称
func (b *Background) Name() string {
	return "background"
}

// Run 执行后台任务逻辑
func (b *Background) Run(ctx *workflow.WukongContext) error {
	logger.Info("Background running", "task_id", ctx.Config.TaskID)

	// 加载系统提示词
	systemPrompt := prompts.LoadPrompt(b.promptDir, "background.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a Background research node.
Your role is to:
1. Gather relevant background information for the task
2. Identify key concepts and terminology
3. Find potential challenges or considerations
4. Prepare context for the researcher node

Provide a structured summary of background information.`
	}

	// ── Step 1：使用搜索工具补充背景资料 ────────────────────────
	searchContext := b.runSearchTool(ctx)

	// ── Step 2：携带搜索结果请求 LLM 分析 ───────────────────────
	var userContent strings.Builder
	userContent.WriteString("User Input: " + ctx.UserInput + "\n\n")
	userContent.WriteString("Intention: " + ctx.State.Intention + "\n\n")
	if searchContext != "" {
		userContent.WriteString("Search Results (use as reference):\n" + searchContext + "\n\n")
	}
	userContent.WriteString("Please provide a structured background summary based on the above information.")

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent.String()},
	}

	response, err := b.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	ctx.State.SetLastNodeOutput(response)
	logger.Info("Background completed", "task_id", ctx.Config.TaskID, "output_length", len(response))
	return nil
}

// runSearchTool 调用可用的搜索工具获取背景资料
func (b *Background) runSearchTool(ctx *workflow.WukongContext) string {
	if b.toolRegistry == nil {
		return ""
	}

	// 按优先级尝试搜索工具
	searchToolNames := []string{"tavily_search", "duckduckgo_search", "mock_search"}
	for _, toolName := range searchToolNames {
		tool, ok := b.toolRegistry.Get(toolName)
		if !ok {
			continue
		}
		query := fmt.Sprintf("%s background information", ctx.UserInput)
		result, err := tool.Execute(ctx.Context, query)
		if err != nil {
			logger.Warn("Background search tool failed", "tool", toolName, "err", err)
			continue
		}
		logger.Info("Background search completed", "task_id", ctx.Config.TaskID, "tool", toolName)
		return result
	}
	return ""
}
