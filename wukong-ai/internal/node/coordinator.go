package node

import (
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Coordinator 协调器节点 - 分析意图，按需调用 Skills，注入对话历史
type Coordinator struct {
	llmProvider   llm.LLM
	promptDir     string
	skillRegistry *skills.SkillRegistry
}

// NewCoordinator 创建协调器节点
func NewCoordinator(llmProvider llm.LLM, promptDir string, skillRegistry *skills.SkillRegistry) *Coordinator {
	return &Coordinator{
		llmProvider:   llmProvider,
		promptDir:     promptDir,
		skillRegistry: skillRegistry,
	}
}

// Name 返回节点名称
func (c *Coordinator) Name() string {
	return "coordinator"
}

// Run 执行协调器逻辑
func (c *Coordinator) Run(ctx *workflow.WukongContext) error {
	logger.Info("Coordinator running", "task_id", ctx.Config.TaskID,
		"conversation_id", ctx.ConversationID)

	if ctx.Config.Mode == workflow.ModeFlash {
		return c.runFlashMode(ctx)
	}

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

	// ── 检测是否为翻译/问答类任务，优先用 Skill 直接处理 ──────────
	if skillResult, handled := c.tryHandleWithSkill(ctx); handled {
		ctx.State.SetIntention(ctx.UserInput)
		ctx.State.SetLastNodeOutput(skillResult)
		ctx.State.SetFinalOutput(skillResult)
		logger.Info("Coordinator handled by skill", "task_id", ctx.Config.TaskID)
		return nil
	}

	// ── 构建用户消息，如有对话历史则前置注入 ─────────────────────
	userMessage := c.buildUserMessage(ctx)

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	response, err := c.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		return err
	}

	intention := c.parseIntention(response)
	ctx.State.SetIntention(intention)
	c.updateConfigBasedOnResponse(ctx, response)

	logger.Info("Coordinator completed", "task_id", ctx.Config.TaskID, "intention", intention)
	return nil
}

// buildUserMessage 构建携带对话历史的用户消息
func (c *Coordinator) buildUserMessage(ctx *workflow.WukongContext) string {
	if ctx.ConversationHistory == "" {
		return ctx.UserInput
	}
	// 将历史前置在当前问题前
	return ctx.ConversationHistory + ctx.UserInput
}

// tryHandleWithSkill 识别简单任务并直接用 Skill 处理
func (c *Coordinator) tryHandleWithSkill(ctx *workflow.WukongContext) (string, bool) {
	if c.skillRegistry == nil {
		return "", false
	}

	input := strings.ToLower(ctx.UserInput)

	// 翻译任务识别
	if strings.Contains(input, "翻译") || strings.Contains(input, "translate") ||
		strings.Contains(input, "译成") || strings.Contains(input, "转换为") {
		if skill, ok := c.skillRegistry.Get("translate"); ok {
			result, err := skill.Execute(ctx.Context, ctx.UserInput)
			if err == nil && result != "" {
				logger.Info("Skill translate applied", "task_id", ctx.Config.TaskID)
				return result, true
			}
		}
	}

	// Flash 模式下简单问答
	if ctx.Config.Mode == workflow.ModeFlash && len(ctx.UserInput) < 100 {
		if skill, ok := c.skillRegistry.Get("qa"); ok {
			result, err := skill.Execute(ctx.Context, ctx.UserInput)
			if err == nil && result != "" {
				logger.Info("Skill qa applied", "task_id", ctx.Config.TaskID)
				return result, true
			}
		}
	}

	return "", false
}

func (c *Coordinator) runFlashMode(ctx *workflow.WukongContext) error {
	// Flash 模式先尝试 skill，再降级到 LLM 直接调用
	if skillResult, handled := c.tryHandleWithSkill(ctx); handled {
		ctx.State.SetIntention(ctx.UserInput)
		ctx.State.SetLastNodeOutput(skillResult)
		ctx.State.SetFinalOutput(skillResult)
		return nil
	}

	// 携带对话历史构建完整消息
	userContent := c.buildUserMessage(ctx)
	messages := []llm.Message{
		{Role: "user", Content: userContent},
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
	logger.Info("Coordinator flash completed", "task_id", ctx.Config.TaskID,
		"output_length", len(finalOutput))
	return nil
}

func (c *Coordinator) parseIntention(response string) string {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.Contains(line, "intention") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return strings.TrimSpace(response)
}

func (c *Coordinator) updateConfigBasedOnResponse(ctx *workflow.WukongContext, response string) {
	if strings.Contains(strings.ToLower(response), "needs_planning") &&
		strings.Contains(strings.ToLower(response), "true") {
		ctx.Config.PlanEnabled = true
		ctx.State.PlanEnabled = true
	}
	if strings.Contains(strings.ToLower(response), "needs_subagents") &&
		strings.Contains(strings.ToLower(response), "true") {
		ctx.Config.SubAgentEnabled = true
		ctx.State.SubAgentEnabled = true
	}
}
