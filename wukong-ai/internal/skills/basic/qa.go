package basic

import (
	"context"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
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
	prompt := `Answer the following question based on the provided context:

"""` + input + `"""

Provide a clear and accurate answer. If the answer cannot be determined from the context, say so.`

	messages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant that answers questions based on the provided context."},
		{Role: "user", Content: prompt},
	}

	return q.llmProvider.ChatWithHistory(ctx, messages)
}

// GetPrompt 获取系统提示词
func (q *QA) GetPrompt() string {
	return "You are a helpful assistant that answers questions based on the provided context."
}

// Ensure QA 实现 Skill 接口
var _ skills.Skill = (*QA)(nil)
