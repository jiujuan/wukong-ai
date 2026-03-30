package subagent

import (
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/prompts"
)

// SubAgent 子代理结构体
type SubAgent struct {
	llmProvider llm.LLM
	promptDir   string
}

// SubTask 子任务
type SubTask struct {
	ID     int64  `json:"id"`
	TaskID string `json:"task_id"`
	Index  int    `json:"index"`
	Input  string `json:"input"`
}

// NewSubAgent 创建新的子代理
func NewSubAgent(llmProvider llm.LLM, promptDir string) *SubAgent {
	return &SubAgent{
		llmProvider: llmProvider,
		promptDir:   promptDir,
	}
}

// Execute 执行子代理任务
func (s *SubAgent) Execute(task *SubTask) (string, error) {
	logger.Debug("SubAgent executing", "task_id", task.TaskID, "index", task.Index)

	// 加载系统提示词
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

	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task.Input},
	}

	response, err := s.llmProvider.ChatWithHistory(nil, messages)
	if err != nil {
		logger.Error("SubAgent execution failed", "task_id", task.TaskID, "index", task.Index, "err", err)
		return "", err
	}

	logger.Debug("SubAgent completed", "task_id", task.TaskID, "index", task.Index)
	return response, nil
}
