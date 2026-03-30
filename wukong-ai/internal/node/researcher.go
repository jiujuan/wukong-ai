package node

import (
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Researcher 研究节点 - 执行任务研究
type Researcher struct {
	llmProvider llm.LLM
	promptDir   string
}

// NewResearcher 创建研究节点
func NewResearcher(llmProvider llm.LLM, promptDir string) *Researcher {
	return &Researcher{
		llmProvider: llmProvider,
		promptDir:   promptDir,
	}
}

// Name 返回节点名称
func (r *Researcher) Name() string {
	return "researcher"
}

// Run 执行研究逻辑
func (r *Researcher) Run(ctx *workflow.WukongContext) error {
	logger.Info("Researcher running", "task_id", ctx.Config.TaskID)

	// 加载系统提示词
	systemPrompt := prompts.LoadPrompt(r.promptDir, "researcher.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a Researcher node for an AI task execution system.
Your role is to:
1. Conduct thorough research on the given tasks
2. Use available tools to gather information
3. Synthesize findings into structured insights
4. Prepare comprehensive research results`
	}

	// 构建消息
	var userContent strings.Builder
	userContent.WriteString("User Input: " + ctx.UserInput + "\n\n")
	userContent.WriteString("Intention: " + ctx.State.Intention + "\n\n")
	if ctx.State.Plan != "" {
		userContent.WriteString("Plan: " + ctx.State.Plan + "\n\n")
	}
	if len(ctx.State.Tasks) > 0 {
		userContent.WriteString("Tasks:\n")
		for i, task := range ctx.State.Tasks {
			userContent.WriteString("  " + string(rune('1'+i)) + ". " + task + "\n")
		}
	}

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
