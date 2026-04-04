package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jiujuan/wukong-ai/internal/conversation"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/node"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/skills"
	"github.com/jiujuan/wukong-ai/internal/skills/basic"
	"github.com/jiujuan/wukong-ai/internal/subagent"
	"github.com/jiujuan/wukong-ai/internal/tools"
	toolcode "github.com/jiujuan/wukong-ai/internal/tools/code"
	toolfile "github.com/jiujuan/wukong-ai/internal/tools/file"
	toolsearch "github.com/jiujuan/wukong-ai/internal/tools/search"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Pool 全局 Worker Pool
type Pool struct {
	size     int
	queue    *queue.PersistentQueue
	eventBus workflow.EventBus
	stopCh   chan struct{}
	wg       sync.WaitGroup
	workerID int
	mu       sync.Mutex
}

// NewPool 创建新的 Worker Pool
func NewPool(size int, q *queue.PersistentQueue, eventBus workflow.EventBus) *Pool {
	return &Pool{
		size:     size,
		queue:    q,
		eventBus: eventBus,
		stopCh:   make(chan struct{}),
	}
}

// SetEngine 设置工作流引擎
func (p *Pool) SetEngine(eventBus workflow.EventBus) {
	p.eventBus = eventBus
}

// Start 启动 Worker Pool
func (p *Pool) Start(ctx context.Context, llmProvider llm.LLM, agentCfg *config.AgentConfig, promptDir string) {
	logger.Info("starting worker pool", "size", p.size)
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		workerID := fmt.Sprintf("worker-%d", i)
		go p.runWorker(ctx, workerID, llmProvider, agentCfg, promptDir)
	}
}

// runWorker 运行单个 Worker
func (p *Pool) runWorker(ctx context.Context, workerID string, llmProvider llm.LLM,
	agentCfg *config.AgentConfig, promptDir string) {

	defer p.wg.Done()
	logger.Info("worker started", "worker_id", workerID)

	scheduler := subagent.NewScheduler(agentCfg.MaxSubAgents)

	for {
		select {
		case <-p.stopCh:
			logger.Info("worker stopping", "worker_id", workerID)
			return
		case <-ctx.Done():
			logger.Info("worker context cancelled", "worker_id", workerID)
			return
		default:
		}

		job, err := p.queue.Dequeue(ctx, workerID)
		if err != nil {
			logger.Error("dequeue failed", "worker_id", workerID, "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if job == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Info("worker picked job", "worker_id", workerID, "task_id", job.TaskID)
		p.executeJob(ctx, job, llmProvider, agentCfg, promptDir, scheduler)
	}
}

// executeJob 执行单个任务
func (p *Pool) executeJob(ctx context.Context, job *queue.TaskJob, llmProvider llm.LLM,
	agentCfg *config.AgentConfig, promptDir string, scheduler *subagent.Scheduler) {

	var req RunRequest
	if err := json.Unmarshal(job.Payload, &req); err != nil {
		logger.Error("failed to unmarshal job payload", "task_id", job.TaskID, "err", err)
		p.queue.MarkFailed(job.TaskID, err.Error())
		return
	}

	cfg := workflow.NewRunConfig(agentCfg)
	cfg.TaskID = job.TaskID
	cfg.ThinkingEnabled = req.ThinkingEnabled
	cfg.PlanEnabled = req.PlanEnabled
	cfg.SubAgentEnabled = req.SubAgentEnabled
	cfg.ConversationID = req.ConversationID
	if req.MaxSubAgents > 0 {
		cfg.MaxSubAgents = req.MaxSubAgents
	}
	if req.TimeoutSeconds > 0 {
		cfg.Timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}
	cfg.Mode = workflow.AutoSelectMode(cfg)

	wCtx := workflow.NewWukongContext(cfg, llmProvider, req.UserInput)
	wCtx.State.Mode = cfg.Mode.String()
	wCtx.EventBus = p.eventBus

	// ── 注册 Tools + Skills ───────────────────────────────────
	registerTools(wCtx.ToolRegistry, agentCfg, req)
	registerSkills(wCtx.SkillRegistry, llmProvider)
	logger.Info("registered tools and skills",
		"task_id", job.TaskID,
		"tools", wCtx.ToolRegistry.GetNames(),
		"skills", wCtx.SkillRegistry.GetNames(),
	)

	// ── 加载对话历史，注入上下文 ──────────────────────────────
	if req.ConversationID != "" {
		p.loadConversationHistory(wCtx, req.ConversationID)
	}

	// ── 创建 DAG 并执行 ───────────────────────────────────────
	nodes := createNodes(llmProvider, scheduler, promptDir, wCtx.ToolRegistry, wCtx.SkillRegistry)
	wf := workflow.BuildWorkflow(cfg.Mode, nodes)
	engine := workflow.NewEngine(wf, p.eventBus)

	var startErr error
	if req.ResumeFrom != "" {
		startErr = engine.RunFromBreakpoint(wCtx)
	} else {
		startErr = engine.RunFromStart(wCtx)
	}

	if startErr != nil {
		logger.Error("job execution failed", "task_id", job.TaskID, "err", startErr)
		engine.Fail(wCtx, startErr.Error())
		p.queue.MarkFailed(job.TaskID, startErr.Error())
		// 即使失败也保存用户轮（AI 轮不保存）
		if req.ConversationID != "" {
			p.saveTurns(wCtx, req.UserInput, "", req.ConversationID)
		}
		return
	}

	engine.Complete(wCtx)
	p.queue.MarkSuccess(job.TaskID)
	logger.Info("job completed", "task_id", job.TaskID)

	// ── 任务成功后保存对话轮次 ────────────────────────────────
	if req.ConversationID != "" {
		p.saveTurns(wCtx, req.UserInput, wCtx.State.FinalOutput, req.ConversationID)
	}
}

// loadConversationHistory 查询历史轮次并构建注入字符串
func (p *Pool) loadConversationHistory(wCtx *workflow.WukongContext, convID string) {
	conv, err := repository.GetConversation(convID)
	if err != nil || conv == nil {
		logger.Warn("conversation not found, running without history",
			"conversation_id", convID, "err", err)
		return
	}
	wCtx.Conv = conv

	turns, err := repository.GetRecentTurns(convID, conversation.MaxRecentTurns)
	if err != nil {
		logger.Warn("failed to load conversation turns", "conversation_id", convID, "err", err)
		return
	}

	wCtx.ConversationHistory = conversation.BuildConversationContext(conv, turns, conversation.MaxRecentTurns)
	logger.Info("conversation history loaded",
		"conversation_id", convID,
		"turns_loaded", len(turns),
		"history_length", len(wCtx.ConversationHistory),
	)
}

// saveTurns 任务执行完成后，将本轮用户输入和 AI 输出追加到对话轮次表
func (p *Pool) saveTurns(wCtx *workflow.WukongContext, userInput, finalOutput, convID string) {
	nextIdx, err := repository.NextTurnIndex(convID)
	if err != nil {
		logger.Warn("failed to get next turn index", "conversation_id", convID, "err", err)
		return
	}

	now := time.Now()

	// 保存用户轮
	userTurn := &conversation.Turn{
		ConversationID: convID,
		TurnIndex:      nextIdx,
		Role:           "user",
		Content:        userInput,
		CreateTime:     now,
	}
	if _, err := repository.AddTurn(userTurn); err != nil {
		logger.Warn("failed to save user turn", "conversation_id", convID, "err", err)
	}

	// 保存 AI 轮（有输出时才保存）
	if finalOutput != "" {
		summary := conversation.TruncateOutput(finalOutput, 200)
		assistantTurn := &conversation.Turn{
			ConversationID: convID,
			TaskID:         wCtx.Config.TaskID,
			TurnIndex:      nextIdx + 1,
			Role:           "assistant",
			Content:        summary,     // 摘要（注入 Prompt 用）
			FullOutput:     finalOutput, // 完整输出
			CreateTime:     now,
		}
		if _, err := repository.AddTurn(assistantTurn); err != nil {
			logger.Warn("failed to save assistant turn", "conversation_id", convID, "err", err)
		}
	}

	// 更新对话 turn_count
	if err := repository.IncrTurnCount(convID); err != nil {
		logger.Warn("failed to incr turn count", "conversation_id", convID, "err", err)
	}

	// 自动更新对话标题（如果还是默认标题）
	if wCtx.Conv != nil && (wCtx.Conv.Title == "新对话" || wCtx.Conv.Title == "") {
		title := conversation.TruncateOutput(userInput, 40)
		if err := repository.UpdateConversationTitle(convID, title); err != nil {
			logger.Warn("failed to update conversation title", "err", err)
		}
	}

	logger.Info("conversation turns saved",
		"conversation_id", convID,
		"task_id", wCtx.Config.TaskID,
		"turn_index", nextIdx,
	)
}

// ── Tool / Skill 注册 ─────────────────────────────────────────────────────────

func registerTools(registry *tools.ToolRegistry, agentCfg *config.AgentConfig, req RunRequest) {
	if req.TavilyAPIKey != "" {
		registry.Register(toolsearch.NewTavilySearch(req.TavilyAPIKey))
		logger.Debug("tool registered: tavily_search")
	} else if req.DuckDuckGoEnabled {
		registry.Register(toolsearch.NewDuckDuckGoSearch())
		logger.Debug("tool registered: duckduckgo_search")
	} else {
		registry.Register(toolsearch.NewMockSearch())
		logger.Debug("tool registered: mock_search (fallback)")
	}

	if len(req.FileAllowedPaths) > 0 {
		registry.Register(toolfile.NewReader(req.FileAllowedPaths))
		registry.Register(toolfile.NewWriter(req.FileAllowedPaths))
		logger.Debug("tool registered: file_reader, file_writer")
	}

	if req.SandboxDir != "" {
		if req.PythonReplEnabled {
			registry.Register(toolcode.NewPythonREPL(req.SandboxDir, true))
			logger.Debug("tool registered: python_repl")
		}
		if req.BashEnabled {
			registry.Register(toolcode.NewBashExec(req.SandboxDir, true))
			logger.Debug("tool registered: bash_exec")
		}
	}
}

func registerSkills(registry *skills.SkillRegistry, llmProvider llm.LLM) {
	registry.Register(basic.NewSummarize(llmProvider))
	registry.Register(basic.NewTranslate(llmProvider))
	registry.Register(basic.NewQA(llmProvider))
	logger.Debug("skills registered: summarize, translate, qa")
}

func createNodes(
	llmProvider llm.LLM,
	scheduler *subagent.Scheduler,
	promptDir string,
	toolRegistry *tools.ToolRegistry,
	skillRegistry *skills.SkillRegistry,
) *workflow.NodeSet {
	return &workflow.NodeSet{
		Coordinator:     node.NewCoordinator(llmProvider, promptDir, skillRegistry),
		Background:      node.NewBackground(llmProvider, promptDir, toolRegistry),
		Planner:         node.NewPlanner(llmProvider, promptDir),
		Researcher:      node.NewResearcher(llmProvider, promptDir, toolRegistry),
		SubAgentManager: node.NewSubAgentManager(scheduler, 3, toolRegistry),
		Reporter:        node.NewReporter(llmProvider, promptDir, skillRegistry),
	}
}

// GracefulStop 优雅关闭
func (p *Pool) GracefulStop() {
	logger.Info("stopping worker pool")
	close(p.stopCh)
	p.wg.Wait()
	logger.Info("worker pool stopped gracefully")
}

// GetSize 获取 Pool 大小
func (p *Pool) GetSize() int {
	return p.size
}

// RunRequest 运行请求（含工具开关和对话 ID）
type RunRequest struct {
	UserInput       string   `json:"user_input"`
	ThinkingEnabled bool     `json:"thinking_enabled"`
	PlanEnabled     bool     `json:"plan_enabled"`
	SubAgentEnabled bool     `json:"subagent_enabled"`
	MaxSubAgents    int      `json:"max_sub_agents"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	ResumeFrom      string   `json:"resume_from,omitempty"`
	ConversationID  string   `json:"conversation_id,omitempty"` // 多轮对话 ID
	// 工具配置
	TavilyAPIKey      string   `json:"tavily_api_key,omitempty"`
	DuckDuckGoEnabled bool     `json:"duckduckgo_enabled"`
	FileAllowedPaths  []string `json:"file_allowed_paths,omitempty"`
	SandboxDir        string   `json:"sandbox_dir,omitempty"`
	PythonReplEnabled bool     `json:"python_repl_enabled"`
	BashEnabled       bool     `json:"bash_enabled"`
}
