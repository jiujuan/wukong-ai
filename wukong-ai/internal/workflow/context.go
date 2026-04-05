package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jiujuan/wukong-ai/internal/conversation"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/skills"
	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/uuid"
)

// AttachmentMemoryQuerier 附件 RAG 检索接口（避免循环导入）
type AttachmentMemoryQuerier interface {
	QueryAttachments(ctx context.Context, taskID, query string, topK int) ([]string, error)
}

// WukongContext 全局执行上下文，贯穿整个 DAG 执行生命周期
type WukongContext struct {
	Context       context.Context
	UserInput     string
	Config        *RunConfig
	State         *RunState
	LLMProvider   llm.LLM
	EventBus      interface{ Publish(taskID string, event ProgressEvent) }
	ToolRegistry  *tools.ToolRegistry
	SkillRegistry *skills.SkillRegistry

	// ── 多轮对话（v0.9）──────────────────────────────────────
	ConversationID      string
	ConversationHistory string
	Conv                *conversation.Conversation

	// ── 文件附件（v1.1）──────────────────────────────────────
	Attachments       []*repository.TaskAttachment // 附件元信息列表
	AttachmentQuerier AttachmentMemoryQuerier       // 附件 RAG 检索
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
	ConversationID  string
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

// HasConversation 是否关联了多轮对话
func (w *WukongContext) HasConversation() bool {
	return w.ConversationID != ""
}

// HasAttachments 是否有文件附件
func (w *WukongContext) HasAttachments() bool {
	return len(w.Attachments) > 0
}

// QueryAttachmentContext 通过 RAG 检索附件相关内容，返回注入 Prompt 的文本块
func (w *WukongContext) QueryAttachmentContext(topK int) string {
	if !w.HasAttachments() || w.AttachmentQuerier == nil {
		return ""
	}
	chunks, err := w.AttachmentQuerier.QueryAttachments(
		w.Context, w.Config.TaskID, w.UserInput, topK,
	)
	if err != nil || len(chunks) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("[Reference Context — 来自用户上传文件]\n---\n")
	for i, chunk := range chunks {
		sb.WriteString(fmt.Sprintf("[片段 %d]\n%s\n\n", i+1, chunk))
	}
	sb.WriteString("---\n\n")
	return sb.String()
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
