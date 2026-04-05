package basic

import (
	"context"
	"fmt"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// QA 问答技能
type QA struct {
	llmProvider llm.LLM
}

// NewQA 创建问答技能
func NewQA(llmProvider llm.LLM) *QA {
	return &QA{
		llmProvider: llmProvider,
	}
}

// Name 返回技能名称
func (q *QA) Name() string {
	return "qa"
}

// Description 返回技能描述
func (q *QA) Description() string {
	return "Answer questions based on provided context"
}

// Execute 执行技能
func (q *QA) Execute(ctx context.Context, input string) (string, error) {
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		return "", fmt.Errorf("qa: empty input")
	}
	logger.Info("qa skill start", "input_length", len(cleanInput))

	prompt := `Please answer the user's question directly and clearly.

If the input includes explicit context, prioritize that context.
If no explicit context is provided, answer with your general knowledge.

User input:

"""` + cleanInput + `"""

Return only the final answer.`

	messages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant. Provide accurate, concise, and directly useful answers."},
		{Role: "user", Content: prompt},
	}

	result, err := q.llmProvider.ChatWithHistory(ctx, messages)
	if err != nil {
		logger.Warn("qa skill llm call failed", "err", err)
		return "", err
	}
	result = strings.TrimSpace(result)
	if result == "" {
		logger.Warn("qa skill empty output")
		return "", fmt.Errorf("qa: empty output")
	}
	logger.Info("qa skill completed", "output_length", len(result))
	return result, nil
}

// GetPrompt 获取系统提示词
func (q *QA) GetPrompt() string {
	return "You are a helpful assistant that answers questions based on the provided context."
}

// Ensure QA 实现 Skill 接口
var _ skills.Skill = (*QA)(nil)
