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

// Researcher 研究节点 - 深度分析，按需调用工具
type Researcher struct {
	llmProvider  llm.LLM
	promptDir    string
	toolRegistry *tools.ToolRegistry
}

// NewResearcher 创建研究节点
func NewResearcher(llmProvider llm.LLM, promptDir string, toolRegistry *tools.ToolRegistry) *Researcher {
	return &Researcher{
		llmProvider:  llmProvider,
		promptDir:    promptDir,
		toolRegistry: toolRegistry,
	}
}

// Name 返回节点名称
func (r *Researcher) Name() string {
	return "researcher"
}

// Run 执行研究逻辑
func (r *Researcher) Run(ctx *workflow.WukongContext) error {
	logger.Info("Researcher running", "task_id", ctx.Config.TaskID)

	systemPrompt := prompts.LoadPrompt(r.promptDir, "researcher.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a Researcher node for an AI task execution system.
Your role is to:
1. Conduct thorough research on the given tasks
2. Use available tools to gather information
3. Synthesize findings into structured insights
4. Prepare comprehensive research results`
	}

	// ── Step 1：针对每个子任务调用搜索工具 ──────────────────────
	toolResults := r.runToolsForTasks(ctx)

	// ── Step 2：汇总工具结果 + LLM 深度分析 ─────────────────────
	var userContent strings.Builder
	userContent.WriteString("User Input: " + ctx.UserInput + "\n\n")
	userContent.WriteString("Intention: " + ctx.State.Intention + "\n\n")

	if ctx.State.Plan != "" {
		userContent.WriteString("Plan: " + ctx.State.Plan + "\n\n")
	}
	if len(ctx.State.Tasks) > 0 {
		userContent.WriteString("Tasks:\n")
		for i, task := range ctx.State.Tasks {
			userContent.WriteString(fmt.Sprintf("  %d. %s\n", i+1, task))
		}
		userContent.WriteString("\n")
	}
	if toolResults != "" {
		userContent.WriteString("Tool Research Results (use as reference):\n" + toolResults + "\n\n")
	}
	userContent.WriteString("Please synthesize all the above into comprehensive research findings.")

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userContent.String()},
	}

	response, err := r.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	ctx.State.SetLastNodeOutput(response)
	logger.Info("Researcher completed", "task_id", ctx.Config.TaskID, "output_length", len(response))
	return nil
}

// runToolsForTasks 对每个任务依次调用搜索工具，汇总所有结果
func (r *Researcher) runToolsForTasks(ctx *workflow.WukongContext) string {
	if r.toolRegistry == nil {
		return ""
	}

	// 确定搜索工具
	var searchTool tools.Tool
	for _, name := range []string{"tavily_search", "duckduckgo_search", "mock_search"} {
		if t, ok := r.toolRegistry.Get(name); ok {
			searchTool = t
			break
		}
	}
	if searchTool == nil {
		return ""
	}

	// 确定要搜索的查询列表（有子任务就按子任务查，否则用用户输入）
	queries := ctx.State.Tasks
	if len(queries) == 0 {
		queries = []string{ctx.UserInput}
	}

	var allResults strings.Builder
	for i, query := range queries {
		result, err := searchTool.Execute(ctx.Context, query)
		if err != nil {
			logger.Warn("Researcher search failed", "task_id", ctx.Config.TaskID, "query", query, "err", err)
			continue
		}
		allResults.WriteString(fmt.Sprintf("--- Search Result %d: %s ---\n%s\n\n", i+1, query, result))
		logger.Info("Researcher search done", "task_id", ctx.Config.TaskID, "tool", searchTool.Name(), "query_index", i+1)
	}
	return allResults.String()
}
