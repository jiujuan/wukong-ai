package workflow

import (
	"context"
	"time"

	"github.com/jiujuan/wukong-ai/internal/conversation"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/uuid"
)

// WukongContext 全局执行上下文，贯穿整个 DAG 执行生命周期
type WukongContext struct {
	Context       context.Context
	UserInput     string
	Config        *RunConfig
	State         *RunState
	LLMProvider   llm.LLM
	EventBus      interface{ Publish(taskID string, event ProgressEvent) }
	ToolRegistry  *tools.ToolRegistry   // 工具注册表，节点按需调用
	SkillRegistry *skills.SkillRegistry // 技能注册表，节点按需调用

	// ── 多轮对话 ──────────────────────────────────────────────
	ConversationID      string             // 所属对话 ID（空 = 单次任务，不关联对话）
	ConversationHistory string             // 由 BuildConversationContext 生成，注入 Prompt
	Conv                *conversation.Conversation // 对话元信息（用于摘要更新）
}

// RunConfig 单次执行配置
type RunConfig struct {
	TaskID          string
	Mode            Mode
	ThinkingEnabled bool
	PlanEnabled     bool
	SubAgentEnabled bool
	MaxSubAgents    int
	Timeout         time.Duration
	RetryCount      int
	ConversationID  string // 透传到 WukongContext
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
		Context:        context.Background(),
		UserInput:      userInput,
		Config:         cfg,
		State:          NewRunState(cfg.TaskID, userInput),
		LLMProvider:    llmProvider,
		ToolRegistry:   tools.NewToolRegistry(),
		SkillRegistry:  skills.NewSkillRegistry(),
		ConversationID: cfg.ConversationID,
	}
}

// NewWukongContextWithCancel 创建带取消上下文的执行上下文
func NewWukongContextWithCancel(cfg *RunConfig, llmProvider llm.LLM, userInput string) (*WukongContext, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	return &WukongContext{
		Context:        ctx,
		UserInput:      userInput,
		Config:         cfg,
		State:          NewRunState(cfg.TaskID, userInput),
		LLMProvider:    llmProvider,
		ToolRegistry:   tools.NewToolRegistry(),
		SkillRegistry:  skills.NewSkillRegistry(),
		ConversationID: cfg.ConversationID,
	}, cancel
}

// HasConversation 判断当前任务是否关联了多轮对话
func (w *WukongContext) HasConversation() bool {
	return w.ConversationID != ""
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
