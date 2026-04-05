package node

import (
	"encoding/json"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// Planner 计划节点 - 生成执行计划
type Planner struct {
	llmProvider llm.LLM
	promptDir   string
}

// NewPlanner 创建计划节点
func NewPlanner(llmProvider llm.LLM, promptDir string) *Planner {
	return &Planner{
		llmProvider: llmProvider,
		promptDir:   promptDir,
	}
}

// Name 返回节点名称
func (p *Planner) Name() string {
	return "planner"
}

// Run 执行计划逻辑
func (p *Planner) Run(ctx *workflow.WukongContext) error {
	logger.Info("Planner running", "task_id", ctx.Config.TaskID, "intention_length", len(ctx.State.Intention))

	// 加载系统提示词
	systemPrompt := prompts.LoadPrompt(p.promptDir, "planner.txt")
	if systemPrompt == "" {
		systemPrompt = `You are a Planner node for an AI task execution system.
Your role is to:
1. Break down the user's intention into actionable tasks
2. Determine dependencies between tasks
3. Estimate task complexity
4. Generate a clear execution plan

Respond with a JSON object containing:
{
  "plan": "A detailed execution plan with numbered steps",
  "tasks": ["Task 1 description", "Task 2 description", ...]
}`
	}

	// 构建消息
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "User Intention: " + ctx.State.Intention},
	}

	response, err := p.llmProvider.ChatWithHistory(ctx.Context, messages)
	if err != nil {
		logger.Error("Planner llm call failed", "task_id", ctx.Config.TaskID, "err", err)
		return err
	}

	// 解析计划和任务
	plan, tasks := p.parsePlanResponse(response)
	ctx.State.SetPlan(plan)
	ctx.State.SetTasks(tasks)
	ctx.State.SetLastNodeOutput(plan)
	if ctx.EventBus != nil && strings.TrimSpace(plan) != "" {
		ctx.EventBus.Publish(ctx.Config.TaskID, workflow.ProgressEvent{
			Type:   "sub_agent_update",
			Node:   "planner",
			Status: "running",
			Latest: plan,
		})
	}
	logger.Info("Planner state updated", "task_id", ctx.Config.TaskID, "plan_length", len(plan), "task_count", len(tasks))

	// 保存任务列表到数据库
	if len(tasks) > 0 {
		tasksJSON, marshalErr := json.Marshal(tasks)
		if marshalErr != nil {
			logger.Warn("failed to marshal tasks list", "task_id", ctx.Config.TaskID, "err", marshalErr)
		}
		if err := repository.UpdateTasksList(ctx.Config.TaskID, tasksJSON); err != nil {
			logger.Warn("failed to save tasks list", "task_id", ctx.Config.TaskID, "err", err)
		}
	}

	logger.Info("Planner completed", "task_id", ctx.Config.TaskID, "task_count", len(tasks))
	return nil
}

// parsePlanResponse 解析计划响应
func (p *Planner) parsePlanResponse(response string) (string, []string) {
	var plan, tasksStr string

	lines := strings.Split(response, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "plan") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				plan = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(lower, "tasks") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				tasksStr = strings.TrimSpace(parts[1])
			}
		}
	}

	// 解析任务列表
	var tasks []string
	if tasksStr != "" {
		// 尝试 JSON 解析
		var parsed struct {
			Tasks []string `json:"tasks"`
		}
		if err := json.Unmarshal([]byte(response), &parsed); err == nil && len(parsed.Tasks) > 0 {
			tasks = parsed.Tasks
		} else {
			// 尝试从文本中提取
			tasks = p.extractTasksFromText(response)
		}
	}

	return plan, tasks
}

// extractTasksFromText 从文本中提取任务
func (p *Planner) extractTasksFromText(response string) []string {
	var tasks []string
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 匹配数字开头的任务
		if len(line) > 0 && (line[0] >= '1' && line[0] <= '9') {
			// 去除数字和点号前缀
			parts := strings.SplitN(line, ".", 2)
			if len(parts) >= 2 {
				task := strings.TrimSpace(parts[1])
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}
