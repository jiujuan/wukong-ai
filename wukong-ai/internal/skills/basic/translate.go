package basic

import (
	"context"
	"fmt"
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
	return &Translate{llmProvider: llmProvider}
}

// Name 返回技能名称
func (t *Translate) Name() string { return "translate" }

// Description 返回技能描述
func (t *Translate) Description() string { return "Translate text between languages" }

// Execute 执行翻译
// input 格式1：直接是待翻译文本（自动翻译为中文或英文）
// input 格式2：<target_lang>\n<text>（指定目标语言）
func (t *Translate) Execute(ctx context.Context, input string) (string, error) {
	var targetLang, text string

	parts := strings.SplitN(input, "\n", 2)
	if len(parts) == 2 && len(parts[0]) <= 20 {
		// 格式2：首行为目标语言
		targetLang = strings.TrimSpace(parts[0])
		text = parts[1]
	} else {
		// 格式1：自动检测目标语言
		targetLang = "English"
		if containsChinese(input) {
			targetLang = "English"
		} else {
			targetLang = "Chinese"
		}
		text = input
	}

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("translate: empty input text")
	}

	prompt := fmt.Sprintf("Translate the following text to %s:\n\n\"\"\"%s\"\"\"\n\nProvide only the translation without explanations.", targetLang, text)

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

// containsChinese 检测字符串是否包含中文字符
func containsChinese(s string) bool {
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}

// 编译期接口检查
var _ skills.Skill = (*Translate)(nil)
