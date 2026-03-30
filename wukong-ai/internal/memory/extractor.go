package memory

import (
	"regexp"
	"strings"
)

// Extractor 结构化记忆提取器
type Extractor struct{}

// NewExtractor 创建提取器
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractKeyPoints 提取关键点
func (e *Extractor) ExtractKeyPoints(content string) []string {
	var keyPoints []string

	// 提取以数字开头的点
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			// 匹配数字或项目符号开头的行
			patterns := []string{
				`^\d+[\.、\)]`,    // 1. 2. 或 1、2、
				`^[a-zA-Z][\.\)]`, // a. b. 或 a) b)
				`^[-•*]\s`,        // - • *
			}
			for _, pattern := range patterns {
				if matched, _ := regexp.MatchString(pattern, line); matched {
					keyPoints = append(keyPoints, line)
					break
				}
			}
		}
	}

	return keyPoints
}

// ExtractFacts 提取事实
func (e *Extractor) ExtractFacts(content string) []string {
	var facts []string

	// 简单的事实提取（以特定词汇开头）
	factKeywords := []string{"根据", "数据显示", "研究表明", "发现", "表明", "证明"}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, keyword := range factKeywords {
			if strings.HasPrefix(line, keyword) {
				facts = append(facts, line)
				break
			}
		}
	}

	return facts
}

// ExtractEntities 提取实体
func (e *Extractor) ExtractEntities(content string) []string {
	var entities []string

	// 提取引号中的内容
	quotePattern := regexp.MustCompile(`"([^"]+)"|'([^']+)'`)
	matches := quotePattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			entities = append(entities, match[1])
		}
		if len(match) > 2 && match[2] != "" {
			entities = append(entities, match[2])
		}
	}

	return entities
}

// Summarize 总结内容
func (e *Extractor) Summarize(content string, maxLength int) string {
	// 简单截取
	if len(content) <= maxLength {
		return content
	}

	// 在句号处截断
	truncated := content[:maxLength]
	lastPeriod := strings.LastIndex(truncated, "。")
	if lastPeriod > maxLength/2 {
		return truncated[:lastPeriod+1]
	}

	lastNewline := strings.LastIndex(truncated, "\n")
	if lastNewline > maxLength/2 {
		return truncated[:lastNewline]
	}

	return truncated + "..."
}
