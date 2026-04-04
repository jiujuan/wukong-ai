package conversation

import (
	"fmt"
	"strings"
)

// BuildConversationContext 将对话历史构建成注入 Prompt 的上下文字符串。
// 策略：
//   - 历史轮数 ≤ MaxRecentTurns：全部注入
//   - 历史轮数 > MaxRecentTurns：最早的轮次用 summary 代替，只保留最近 MaxRecentTurns 轮
//   - role=assistant 只注入 content（摘要），不注入 full_output，避免 Token 爆炸
func BuildConversationContext(conv *Conversation, turns []Turn, maxTurns int) string {
	if len(turns) == 0 {
		return ""
	}

	if maxTurns <= 0 {
		maxTurns = MaxRecentTurns
	}

	var sb strings.Builder
	sb.WriteString("[对话历史]\n")

	// 超出 maxTurns 的早期轮次用会话级 Summary 替代
	if len(turns) > maxTurns {
		summary := conv.Summary
		if summary == "" {
			summary = fmt.Sprintf("（前 %d 轮对话已压缩）", len(turns)-maxTurns)
		}
		sb.WriteString(fmt.Sprintf("（早期摘要）%s\n\n", summary))
		turns = turns[len(turns)-maxTurns:]
	}

	for _, t := range turns {
		role := "用户"
		if t.Role == "assistant" {
			role = "AI"
		}
		// content 字段存的是摘要，不是全文；直接注入
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", role, t.Content))
	}

	sb.WriteString("[当前问题]\n")
	return sb.String()
}

// SummarizeTurns 将一批历史轮次压缩为一句简短摘要（由 LLM 调用方完成真正压缩，
// 这里提供纯文本拼接版兜底，供不需要 LLM 的场景使用）。
func SummarizeTurns(turns []Turn) string {
	if len(turns) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, t := range turns {
		role := "用户"
		if t.Role == "assistant" {
			role = "AI"
		}
		// 每轮只取前 80 字作为摘要片段
		content := t.Content
		if len([]rune(content)) > 80 {
			content = string([]rune(content)[:80]) + "..."
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", role, content))
	}
	return sb.String()
}

// TruncateOutput 将长输出截断为摘要内容（存入 Turn.Content）
func TruncateOutput(output string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = 200
	}
	runes := []rune(output)
	if len(runes) <= maxRunes {
		return output
	}
	return string(runes[:maxRunes]) + "..."
}
