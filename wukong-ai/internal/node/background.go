package node

import (
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Background 后台任务节点 - 在后台执行信息收集
type Background struct {
	llmProvider llm.LLM
	promptDir   string
}

// NewBackground 创建后台任务节点
func NewBackground(llmProvider llm.LLM, promptDir string) *Background {
	return &Background{
		llmProvider: llmProvider,
		promptDir:   promptDir,
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

	// 构建消息
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "User Input: " + ctx.UserInput + "\n\nIntention: " + ctx.State.Intention},
	}

	response, err := b.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	ctx.State.SetLastNodeOutput(response)
	logger.Info("Background completed", "task_id", ctx.Config.TaskID, "output_length", len(response))
	return nil
}
