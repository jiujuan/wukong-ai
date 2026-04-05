package basic

import (
	"context"
	"fmt"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
)

// Summarize 摘要技能
type Summarize struct {
	llmProvider llm.LLM
}

// NewSummarize 创建摘要技能
func NewSummarize(llmProvider llm.LLM) *Summarize {
	return &Summarize{
		llmProvider: llmProvider,
	}
}

// Name 返回技能名称
func (s *Summarize) Name() string {
	return "summarize"
}

// Description 返回技能描述
func (s *Summarize) Description() string {
	return "Summarize long text into concise key points"
}

// Execute 执行技能
func (s *Summarize) Execute(ctx context.Context, input string) (string, error) {
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		return "", fmt.Errorf("summarize: empty input")
	}

	prompt := `Please summarize the following text into concise key points:

"""` + cleanInput + `"""

Provide a clear and concise summary with main points.`

	messages := []llm.Message{
		{Role: "system", Content: "You are a summarization expert."},
		{Role: "user", Content: prompt},
	}

	result, err := s.llmProvider.ChatWithHistory(ctx, messages)
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)
	if result == "" {
		return "", fmt.Errorf("summarize: empty output")
	}
	return result, nil
}

// GetPrompt 获取系统提示词
func (s *Summarize) GetPrompt() string {
	return "You are a summarization expert. Summarize text concisely."
}

// Ensure Summarize 实现 Skill 接口
var _ skills.Skill = (*Summarize)(nil)

// SummarizeText 简单的文本摘要
func SummarizeText(text string, maxLength int) string {
	// 简单实现：截取前 N 个字符
	if len(text) <= maxLength {
		return text
	}

	// 在句号处截断
	truncated := text[:maxLength]
	lastPeriod := strings.LastIndex(truncated, "。")
	if lastPeriod > maxLength/2 {
		return truncated[:lastPeriod+1]
	}

	return truncated + "..."
}
