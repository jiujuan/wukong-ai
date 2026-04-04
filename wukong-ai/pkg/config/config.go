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

// LLMConfig 主模型 + 降级链配置
type LLMConfig struct {
	// ── 主模型（向前兼容旧配置）──────────────────────────────────
	Provider       string `mapstructure:"provider"`        // openai / deepseek / ollama
	APIKey         string `mapstructure:"api_key"`
	BaseURL        string `mapstructure:"base_url"`
	Model          string `mapstructure:"model"`
	EmbeddingModel string `mapstructure:"embedding_model"`
	EmbeddingDim   int    `mapstructure:"embedding_dim"`

	// ── 限流（主模型）─────────────────────────────────────────────
	MaxRPM int `mapstructure:"max_rpm"` // 每分钟最大请求数，0 = 不限
	MaxTPM int `mapstructure:"max_tpm"` // 每分钟最大 Token 数，0 = 不限

	// ── 降级链（可选）────────────────────────────────────────────
	// 配置后主模型失败时依次尝试 Fallbacks 中的模型
	Fallbacks []FallbackLLMConfig `mapstructure:"fallbacks"`

	// ── 熔断器全局参数──────────────────────────────────────────
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

// FallbackLLMConfig 降级链中单个备用模型配置
type FallbackLLMConfig struct {
	Provider string `mapstructure:"provider"` // openai / deepseek / ollama
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
	MaxRPM   int    `mapstructure:"max_rpm"`
	MaxTPM   int    `mapstructure:"max_tpm"`
}

// CircuitBreakerConfig 熔断器参数
type CircuitBreakerConfig struct {
	Threshold int           `mapstructure:"threshold"` // 连续失败次数阈值（默认 5）
	Timeout   time.Duration `mapstructure:"timeout"`   // 熔断持续时间（默认 60s）
}

// AgentConfig Agent 配置
type AgentConfig struct {
	MaxWorkers      int           `mapstructure:"max_workers"`
	MaxSubAgents    int           `mapstructure:"max_sub_agents"`
	DefaultTimeout  time.Duration `mapstructure:"default_timeout"`
	RetryCount      int           `mapstructure:"retry_count"`
	StaleJobTimeout time.Duration `mapstructure:"stale_job_timeout"`
}

// SandboxConfig 沙箱执行环境配置
type SandboxConfig struct {
	BaseDir           string `mapstructure:"base_dir"`
	PythonReplEnabled bool   `mapstructure:"python_repl_enabled"`
	BashEnabled       bool   `mapstructure:"bash_enabled"`
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
	AllowedPaths []string `mapstructure:"allowed_paths"`
}

// MemoryConfig 记忆系统配置
type MemoryConfig struct {
	ShortTermMaxMessages int `mapstructure:"short_term_max_messages"`
	LongTermTopK         int `mapstructure:"long_term_top_k"`
}

// PromptsConfig 提示词配置
type PromptsConfig struct {
	Dir string `mapstructure:"dir"`
}

// DefaultCircuitBreakerThreshold 熔断器默认阈值
const DefaultCircuitBreakerThreshold = 5

// DefaultCircuitBreakerTimeout 熔断器默认超时
const DefaultCircuitBreakerTimeout = 60 * time.Second
