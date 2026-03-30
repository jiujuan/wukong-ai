package prompts

import (
	"os"
	"path/filepath"
	"strings"
)

var promptDir = "./prompts"

// SetPromptDir 设置提示词目录
func SetPromptDir(dir string) {
	promptDir = dir
}

// LoadPrompt 加载提示词
func LoadPrompt(dir, filename string) string {
	if dir != "" {
		promptDir = dir
	}

	path := filepath.Join(promptDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// LoadAllPrompts 加载所有提示词
func LoadAllPrompts(dir string) map[string]string {
	if dir != "" {
		promptDir = dir
	}

	prompts := make(map[string]string)
	files := []string{
		"coordinator.txt",
		"background.txt",
		"planner.txt",
		"researcher.txt",
		"subagent.txt",
		"reporter.txt",
	}

	for _, file := range files {
		name := strings.TrimSuffix(file, ".txt")
		prompts[name] = LoadPrompt(promptDir, file)
	}

	return prompts
}

// GetDefaultCoordinatorPrompt 获取默认协调器提示词
func GetDefaultCoordinatorPrompt() string {
	return `You are a Coordinator for an AI task execution system.
Your role is to:
1. Analyze the user's input to understand their intention
2. Determine if planning is needed based on task complexity
3. Decide if sub-agents should be used for parallel execution
4. Provide a clear intention summary

Respond with a clear summary of the user's intention.`
}

// GetDefaultPlannerPrompt 获取默认计划器提示词
func GetDefaultPlannerPrompt() string {
	return `You are a Planner node for an AI task execution system.
Your role is to:
1. Break down the user's intention into actionable tasks
2. Determine dependencies between tasks
3. Estimate task complexity
4. Generate a clear execution plan

Provide a numbered list of tasks.`
}

// GetDefaultReporterPrompt 获取默认报告器提示词
func GetDefaultReporterPrompt() string {
	return `You are a Reporter node for an AI task execution system.
Your role is to:
1. Synthesize all findings and results
2. Generate a comprehensive final report
3. Present information in a clear, structured format
4. Ensure the report addresses the user's original intent

Format the output as a well-structured markdown report.`
}
