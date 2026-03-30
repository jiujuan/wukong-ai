package workflow

import (
	"context"
	"time"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/uuid"
)

// WukongContext 全局执行上下文，贯穿整个 DAG 执行生命周期
type WukongContext struct {
	Context     context.Context
	UserInput   string
	Config      *RunConfig
	State       *RunState
	LLMProvider llm.LLM
	EventBus    interface{ Publish(taskID string, event ProgressEvent) }
}

// RunConfig 单次执行配置
type RunConfig struct {
	TaskID          string        // 由 pkg/uuid.NewTaskID() 生成
	Mode            Mode
	ThinkingEnabled bool
	PlanEnabled     bool
	SubAgentEnabled bool
	MaxSubAgents    int           // 默认值来自 pkg/config.AgentConfig.MaxSubAgents
	Timeout         time.Duration // 默认值来自 pkg/config.AgentConfig.DefaultTimeout
	RetryCount      int           // 默认值来自 pkg/config.AgentConfig.RetryCount
}

// NewRunConfig 创建新的执行配置
func NewRunConfig(agentCfg *config.AgentConfig) *RunConfig {
	return &RunConfig{
		TaskID:          uuid.NewTaskID(),
		Mode:            ModeFlash,
		ThinkingEnabled: false,
		PlanEnabled:     false,
		SubAgentEnabled: false,
		MaxSubAgents:    agentCfg.MaxSubAgents,
		Timeout:         agentCfg.DefaultTimeout,
		RetryCount:      agentCfg.RetryCount,
	}
}

// NewWukongContext 创建新的执行上下文
func NewWukongContext(cfg *RunConfig, llmProvider llm.LLM, userInput string) *WukongContext {
	return &WukongContext{
		Context:     context.Background(),
		UserInput:   userInput,
		Config:      cfg,
		State:       NewRunState(cfg.TaskID, userInput),
		LLMProvider: llmProvider,
	}
}

// NewWukongContextWithCancel 创建带取消上下文的执行上下文
func NewWukongContextWithCancel(cfg *RunConfig, llmProvider llm.LLM, userInput string) (*WukongContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	return &WukongContext{
		Context:     ctx,
		UserInput:   userInput,
		Config:      cfg,
		State:       NewRunState(cfg.TaskID, userInput),
		LLMProvider: llmProvider,
	}, cancel
}

// ProgressEvent 进度事件
type ProgressEvent struct {
	Type      string `json:"type"`
	Node      string `json:"node,omitempty"`
	Status    string `json:"status,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	Total     int    `json:"total,omitempty"`
	Done      int    `json:"done,omitempty"`
	Latest    string `json:"latest,omitempty"`
	Output    string `json:"final_output,omitempty"`
	Timestamp string `json:"timestamp"`
}
