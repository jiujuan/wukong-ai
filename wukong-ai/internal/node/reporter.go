package node

import (
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Reporter 报告节点 - 生成最终报告，按需使用 summarize skill 压缩中间结果
type Reporter struct {
	llmProvider   llm.LLM
	promptDir     string
	skillRegistry *skills.SkillRegistry
}

// NewReporter 创建报告节点
func NewReporter(llmProvider llm.LLM, promptDir string, skillRegistry *skills.SkillRegistry) *Reporter {
	return &Reporter{
		llmProvider:   llmProvider,
		promptDir:     promptDir,
		skillRegistry: skillRegistry,
	}
}

// Name 返回节点名称
func (r *Reporter) Name() string {
	return "reporter"
}

// Run 执行报告生成逻辑
func (r *Reporter) Run(ctx *workflow.WukongContext) error {
	logger.Info("Reporter running", "task_id", ctx.Config.TaskID)

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

	// ── Step 1：对过长的中间结果用 summarize skill 先压缩 ─────────
	subAgentSummary := r.summarizeSubResults(ctx)
	researchSummary := r.summarizeResearch(ctx)

	// ── Step 2：汇总上下文，调用 LLM 生成最终报告 ─────────────────
	var contextContent strings.Builder
	contextContent.WriteString("# Task Summary\n\n")
	contextContent.WriteString("**User Input:** " + ctx.UserInput + "\n\n")
	contextContent.WriteString("**Intention:** " + ctx.State.Intention + "\n\n")

	if ctx.State.Plan != "" {
		contextContent.WriteString("**Plan:**\n" + ctx.State.Plan + "\n\n")
	}
	if subAgentSummary != "" {
		contextContent.WriteString("## Sub-Agent Results\n\n" + subAgentSummary + "\n\n")
	}
	if researchSummary != "" {
		contextContent.WriteString("## Research Findings\n\n" + researchSummary + "\n\n")
	}

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: contextContent.String()},
	}

	response, err := r.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	ctx.State.SetFinalOutput(response)
	logger.Info("Reporter completed", "task_id", ctx.Config.TaskID, "output_length", len(response))
	return nil
}

// summarizeSubResults 当子任务结果较多时，用 summarize skill 先压缩
func (r *Reporter) summarizeSubResults(ctx *workflow.WukongContext) string {
	if len(ctx.State.SubAgentResults) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, result := range ctx.State.SubAgentResults {
		sb.WriteString("### Task " + string(rune('1'+i)) + "\n\n")
		sb.WriteString(result + "\n\n")
	}
	combined := sb.String()

	// 超过 3000 字符时启用 summarize skill 压缩
	if len(combined) > 3000 && r.skillRegistry != nil {
		if skill, ok := r.skillRegistry.Get("summarize"); ok {
			summary, err := skill.Execute(ctx.Context, combined)
			if err == nil && summary != "" {
				logger.Info("Reporter: sub-results summarized by skill",
					"task_id", ctx.Config.TaskID,
					"original_len", len(combined),
					"summary_len", len(summary),
				)
				return summary
			}
			logger.Warn("Reporter: summarize skill failed, using original", "err", err)
		}
	}
	return combined
}

// summarizeResearch 对 Researcher 输出做同样的压缩处理
func (r *Reporter) summarizeResearch(ctx *workflow.WukongContext) string {
	research := ctx.State.LastNodeOutput
	if research == "" {
		return ""
	}
	if len(research) > 3000 && r.skillRegistry != nil {
		if skill, ok := r.skillRegistry.Get("summarize"); ok {
			summary, err := skill.Execute(ctx.Context, research)
			if err == nil && summary != "" {
				logger.Info("Reporter: research summarized by skill",
					"task_id", ctx.Config.TaskID,
					"original_len", len(research),
					"summary_len", len(summary),
				)
				return summary
			}
		}
	}
	return research
}
