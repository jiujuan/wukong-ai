package node

import (
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Reporter 报告节点 - 生成最终报告
type Reporter struct {
	llmProvider llm.LLM
	promptDir   string
}

// NewReporter 创建报告节点
func NewReporter(llmProvider llm.LLM, promptDir string) *Reporter {
	return &Reporter{
		llmProvider: llmProvider,
		promptDir:   promptDir,
	}
}

// Name 返回节点名称
func (r *Reporter) Name() string {
	return "reporter"
}

// Run 执行报告生成逻辑
func (r *Reporter) Run(ctx *workflow.WukongContext) error {
	logger.Info("Reporter running", "task_id", ctx.Config.TaskID)

	// 加载系统提示词
	systemPrompt := prompts.LoadPrompt(r.promptDir, "reporter.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a Reporter node for an AI task execution system.
Your role is to:
1. Synthesize all findings and results
2. Generate a comprehensive final report
3. Present information in a clear, structured format
4. Ensure the report addresses the user's original intent

Format the output as a well-structured markdown report.`
	}

	// 构建上下文内容
	var contextContent strings.Builder
	contextContent.WriteString("# Task Summary\n\n")
	contextContent.WriteString("**User Input:** " + ctx.UserInput + "\n\n")
	contextContent.WriteString("**Intention:** " + ctx.State.Intention + "\n\n")

	if ctx.State.Plan != "" {
		contextContent.WriteString("**Plan:**\n" + ctx.State.Plan + "\n\n")
	}

	if len(ctx.State.SubAgentResults) > 0 {
		contextContent.WriteString("## Sub-Agent Results\n\n")
		for i, result := range ctx.State.SubAgentResults {
			contextContent.WriteString("### Task " + string(rune('1'+i)) + "\n\n")
			contextContent.WriteString(result + "\n\n")
		}
	}

	if ctx.State.LastNodeOutput != "" {
		contextContent.WriteString("## Research Findings\n\n")
		contextContent.WriteString(ctx.State.LastNodeOutput + "\n\n")
	}

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: contextContent.String()},
	}

	response, err := r.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	// 设置最终输出
	ctx.State.SetFinalOutput(response)

	logger.Info("Reporter completed", "task_id", ctx.Config.TaskID, "output_length", len(response))
	return nil
}
