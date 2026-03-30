package node

import (
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Coordinator 协调器节点 - 分析用户输入并决定后续流程
type Coordinator struct {
	llmProvider llm.LLM
	promptDir   string
}

// NewCoordinator 创建协调器节点
func NewCoordinator(llmProvider llm.LLM, promptDir string) *Coordinator {
	return &Coordinator{
		llmProvider: llmProvider,
		promptDir:   promptDir,
	}
}

// Name 返回节点名称
func (c *Coordinator) Name() string {
	return "coordinator"
}

// Run 执行协调器逻辑
func (c *Coordinator) Run(ctx *workflow.WukongContext) error {
	logger.Info("Coordinator running", "task_id", ctx.Config.TaskID)
	if ctx.Config.Mode == workflow.ModeFlash {
		return c.runFlashMode(ctx)
	}

	// 加载系统提示词
	systemPrompt := prompts.LoadPrompt(c.promptDir, "coordinator.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a Coordinator for an AI task execution system.
Your role is to:
1. Analyze the user's input to understand their intention
2. Determine if planning is needed based on task complexity
3. Decide if sub-agents should be used for parallel execution
4. Provide a clear intention summary

Respond with a JSON object containing:
{
  "intention": "A clear summary of what the user wants",
  "needs_planning": true/false,
  "needs_subagents": true/false,
  "complexity": "low/medium/high"
}`
	}

	// 构建用户消息
	userMessage := ctx.UserInput

	// 发送请求
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	response, err := c.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	// 解析响应
	intention := c.parseIntention(response)
	ctx.State.SetIntention(intention)

	// 根据响应更新配置
	c.updateConfigBasedOnResponse(ctx, response)

	logger.Info("Coordinator completed", "task_id", ctx.Config.TaskID, "intention", intention)
	return nil
}

func (c *Coordinator) runFlashMode(ctx *workflow.WukongContext) error {
	messages := []llm.Message{
		{Role: "user", Content: ctx.UserInput},
	}
	var outputBuilder strings.Builder

	streamProvider, ok := c.llmProvider.(llm.StreamLLM)
	if ok {
		err := streamProvider.ChatWithHistoryStream(ctx.Context, messages, func(chunk string) error {
			if chunk == "" {
				return nil
			}
			outputBuilder.WriteString(chunk)
			ctx.State.SetLastNodeOutput(outputBuilder.String())
			if ctx.EventBus != nil {
				ctx.EventBus.Publish(ctx.Config.TaskID, workflow.ProgressEvent{
					Type:   "sub_agent_update",
					Node:   "coordinator",
					Status: "running",
					Latest: chunk,
				})
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		response, err := c.llmProvider.ChatWithHistory(ctx.Context, messages)
		if err != nil {
			return err
		}
		outputBuilder.WriteString(response)
		if ctx.EventBus != nil && response != "" {
			ctx.EventBus.Publish(ctx.Config.TaskID, workflow.ProgressEvent{
				Type:   "sub_agent_update",
				Node:   "coordinator",
				Status: "running",
				Latest: response,
			})
		}
	}

	finalOutput := strings.TrimSpace(outputBuilder.String())
	ctx.State.SetIntention(ctx.UserInput)
	ctx.State.SetLastNodeOutput(finalOutput)
	ctx.State.SetFinalOutput(finalOutput)
	logger.Info("Coordinator flash completed", "task_id", ctx.Config.TaskID, "output_length", len(finalOutput))
	return nil
}

// parseIntention 从响应中解析意图
func (c *Coordinator) parseIntention(response string) string {
	// 简单提取 intention 部分
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.Contains(line, "intention") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	// 如果无法解析，返回原始响应
	return strings.TrimSpace(response)
}

// updateConfigBasedOnResponse 根据响应更新配置
func (c *Coordinator) updateConfigBasedOnResponse(ctx *workflow.WukongContext, response string) {
	// 检查是否需要规划
	if strings.Contains(strings.ToLower(response), "needs_planning") &&
		strings.Contains(strings.ToLower(response), "true") {
		ctx.Config.PlanEnabled = true
		ctx.State.PlanEnabled = true
	}

	// 检查是否需要子代理
	if strings.Contains(strings.ToLower(response), "needs_subagents") &&
		strings.Contains(strings.ToLower(response), "true") {
		ctx.Config.SubAgentEnabled = true
		ctx.State.SubAgentEnabled = true
	}
}
