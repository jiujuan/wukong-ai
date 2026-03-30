package config

import (
	"time"
)

// AppConfig 顶层配置结构体
type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Sandbox  SandboxConfig  `mapstructure:"sandbox"`
	Tools    ToolsConfig    `mapstructure:"tools"`
	Memory   MemoryConfig   `mapstructure:"memory"`
	Prompts  PromptsConfig  `mapstructure:"prompts"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"` // dev / prod
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

// LLMConfig LLM 模型配置
type LLMConfig struct {
	Provider       string `mapstructure:"provider"`        // openai / deepseek / ollama
	APIKey         string `mapstructure:"api_key"`         // 可通过 ${ENV_VAR} 使用环境变量
	BaseURL        string `mapstructure:"base_url"`        // API 基础 URL
	Model          string `mapstructure:"model"`           // 聊天模型
	EmbeddingModel string `mapstructure:"embedding_model"` // 向量化模型
	EmbeddingDim   int    `mapstructure:"embedding_dim"`   // 向量维度 (OpenAI text-embedding-3-small = 1536)
}

// AgentConfig Agent 配置
type AgentConfig struct {
	MaxWorkers      int           `mapstructure:"max_workers"`       // Worker Pool 大小
	MaxSubAgents    int           `mapstructure:"max_sub_agents"`    // 单个 DAG 内最大子 Agent 并发数
	DefaultTimeout  time.Duration `mapstructure:"default_timeout"`   // 单任务最大执行时长
	RetryCount      int           `mapstructure:"retry_count"`       // 失败重试次数
	StaleJobTimeout time.Duration `mapstructure:"stale_job_timeout"` // 判定僵尸任务的超时阈值
}

// SandboxConfig 沙箱执行环境配置
type SandboxConfig struct {
	BaseDir           string `mapstructure:"base_dir"`            // 沙箱根目录
	PythonReplEnabled bool   `mapstructure:"python_repl_enabled"` // 是否启用 Python REPL
	BashEnabled       bool   `mapstructure:"bash_enabled"`        // 是否启用 Bash 执行
}

// ToolsConfig 工具系统配置
type ToolsConfig struct {
	Search SearchConfig `mapstructure:"search"`
	File   FileConfig   `mapstructure:"file"`
}

// SearchConfig 搜索工具配置
type SearchConfig struct {
	TavilyAPIKey      string `mapstructure:"tavily_api_key"`
	DuckDuckGoEnabled bool   `mapstructure:"duckduckgo_enabled"`
}

// FileConfig 文件工具配置
type FileConfig struct {
	AllowedPaths []string `mapstructure:"allowed_paths"` // 允许访问的目录列表
}

// MemoryConfig 记忆系统配置
type MemoryConfig struct {
	ShortTermMaxMessages int `mapstructure:"short_term_max_messages"` // 短期记忆最大消息数
	LongTermTopK         int `mapstructure:"long_term_top_k"`         // 长期记忆检索返回条数
}

// PromptsConfig 提示词配置
type PromptsConfig struct {
	Dir string `mapstructure:"dir"` // 提示词文件目录
}
