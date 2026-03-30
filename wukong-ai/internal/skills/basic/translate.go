package basic

import (
	"context"
	"strings"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
)

// Translate 翻译技能
type Translate struct {
	llmProvider llm.LLM
}

// NewTranslate 创建翻译技能
func NewTranslate(llmProvider llm.LLM) *Translate {
	return &Translate{
		llmProvider: llmProvider,
	}
}

// Name 返回技能名称
func (t *Translate) Name() string {
	return "translate"
}

// Description 返回技能描述
func (t *Translate) Description() string {
	return "Translate text between languages"
}

// Execute 执行技能
func (t *Translate) Execute(ctx context.Context, input string) (string, error) {
	// 解析输入：<target_lang>\n<text>
	parts := strings.SplitN(input, "\n", 2)
	if len(parts) < 2 {
		return "", &skills.SkillError{
			SkillName: t.Name(),
			Message:   "invalid input format, expected: <target_lang>\\n<text>",
		}
	}

	targetLang := strings.TrimSpace(parts[0])
	text := parts[1]

	prompt := `Translate the following text to ` + targetLang + `:

"""` + text + `"""

Provide only the translation without explanations.`

	messages := []llm.Message{
		{Role: "system", Content: "You are a professional translator."},
		{Role: "user", Content: prompt},
	}

	return t.llmProvider.ChatWithHistory(ctx, messages)
}

// GetPrompt 获取系统提示词
func (t *Translate) GetPrompt() string {
	return "You are a professional translator. Translate text accurately."
}

// Ensure Translate 实现 Skill 接口
var _ skills.Skill = (*Translate)(nil)

// SkillError 技能错误
type SkillError struct {
	SkillName string
	Message   string
}

func (e *SkillError) Error() string {
	return e.SkillName + ": " + e.Message
}
