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
	"github.com/jiujuan/wukong-ai/internal/memory"
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

func (p *Pool) runWorker(ctx context.Context, workerID string, llmProvider llm.LLM,
	agentCfg *config.AgentConfig, promptDir string) {

	defer p.wg.Done()
	logger.Info("worker started", "worker_id", workerID)

	scheduler := subagent.NewScheduler(agentCfg.MaxSubAgents)

	for {
		select {
		case <-p.stopCh:
			return
		case <-ctx.Done():
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

	// ── 加载对话历史 ──────────────────────────────────────────
	if req.ConversationID != "" {
		p.loadConversationHistory(wCtx, req.ConversationID)
	}

	// ── 加载附件列表 + 注入 AttachmentQuerier（v1.1）──────────
	p.loadAttachments(wCtx, job.TaskID, llmProvider, 5)

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
		if req.ConversationID != "" {
			p.saveTurns(wCtx, req.UserInput, "", req.ConversationID)
		}
		return
	}

	engine.Complete(wCtx)
	p.queue.MarkSuccess(job.TaskID)
	logger.Info("job completed", "task_id", job.TaskID)

	if req.ConversationID != "" {
		p.saveTurns(wCtx, req.UserInput, wCtx.State.FinalOutput, req.ConversationID)
	}
}

// loadAttachments 加载任务附件列表并注入 AttachmentQuerier
func (p *Pool) loadAttachments(wCtx *workflow.WukongContext, taskID string, llmProvider llm.LLM, topK int) {
	atts, err := repository.GetAttachmentsByTaskID(taskID)
	if err != nil {
		logger.Warn("failed to load attachments", "task_id", taskID, "err", err)
		return
	}
	if len(atts) == 0 {
		return
	}

	// 过滤出已提取完成的附件
	var doneAtts []*repository.TaskAttachment
	for _, a := range atts {
		if a.ExtractStatus == "done" {
			doneAtts = append(doneAtts, a)
		}
	}

	if len(doneAtts) == 0 {
		logger.Info("attachments not ready yet, skipping RAG injection",
			"task_id", taskID, "total", len(atts))
		return
	}

	wCtx.Attachments = doneAtts
	if topK <= 0 {
		topK = 5
	}
	wCtx.AttachmentQuerier = memory.NewAttachmentQuerier(llmProvider, topK)
	logger.Info("attachments loaded for RAG",
		"task_id", taskID, "count", len(doneAtts))
}

// loadConversationHistory 查询历史轮次并构建注入字符串
func (p *Pool) loadConversationHistory(wCtx *workflow.WukongContext, convID string) {
	conv, err := repository.GetConversation(convID)
	if err != nil || conv == nil {
		logger.Warn("conversation not found", "conversation_id", convID, "err", err)
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
		"conversation_id", convID, "turns_loaded", len(turns))
}

// saveTurns 保存对话轮次
func (p *Pool) saveTurns(wCtx *workflow.WukongContext, userInput, finalOutput, convID string) {
	nextIdx, err := repository.NextTurnIndex(convID)
	if err != nil {
		return
	}
	now := time.Now()

	userTurn := &conversation.Turn{
		ConversationID: convID,
		TurnIndex:      nextIdx,
		Role:           "user",
		Content:        userInput,
		CreateTime:     now,
	}
	repository.AddTurn(userTurn)

	if finalOutput != "" {
		summary := conversation.TruncateOutput(finalOutput, 200)
		assistantTurn := &conversation.Turn{
			ConversationID: convID,
			TaskID:         wCtx.Config.TaskID,
			TurnIndex:      nextIdx + 1,
			Role:           "assistant",
			Content:        summary,
			FullOutput:     finalOutput,
			CreateTime:     now,
		}
		repository.AddTurn(assistantTurn)
	}

	repository.IncrTurnCount(convID)

	if wCtx.Conv != nil && (wCtx.Conv.Title == "新对话" || wCtx.Conv.Title == "") {
		title := conversation.TruncateOutput(userInput, 40)
		repository.UpdateConversationTitle(convID, title)
	}
}

// ── Tool / Skill 注册 ─────────────────────────────────────────────────────────

func registerTools(registry *tools.ToolRegistry, agentCfg *config.AgentConfig, req RunRequest) {
	if req.TavilyAPIKey != "" {
		registry.Register(toolsearch.NewTavilySearch(req.TavilyAPIKey))
	} else if req.DuckDuckGoEnabled {
		registry.Register(toolsearch.NewDuckDuckGoSearch())
	} else {
		registry.Register(toolsearch.NewMockSearch())
	}
	if len(req.FileAllowedPaths) > 0 {
		registry.Register(toolfile.NewReader(req.FileAllowedPaths))
		registry.Register(toolfile.NewWriter(req.FileAllowedPaths))
	}
	if req.SandboxDir != "" {
		if req.PythonReplEnabled {
			registry.Register(toolcode.NewPythonREPL(req.SandboxDir, true))
		}
		if req.BashEnabled {
			registry.Register(toolcode.NewBashExec(req.SandboxDir, true))
		}
	}
}

func registerSkills(registry *skills.SkillRegistry, llmProvider llm.LLM) {
	registry.Register(basic.NewSummarize(llmProvider))
	registry.Register(basic.NewTranslate(llmProvider))
	registry.Register(basic.NewQA(llmProvider))
}

func createNodes(llmProvider llm.LLM, scheduler *subagent.Scheduler, promptDir string,
	toolRegistry *tools.ToolRegistry, skillRegistry *skills.SkillRegistry) *workflow.NodeSet {
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
	close(p.stopCh)
	p.wg.Wait()
}

// GetSize 获取 Pool 大小
func (p *Pool) GetSize() int { return p.size }

// RunRequest 运行请求
type RunRequest struct {
	UserInput       string   `json:"user_input"`
	ThinkingEnabled bool     `json:"thinking_enabled"`
	PlanEnabled     bool     `json:"plan_enabled"`
	SubAgentEnabled bool     `json:"subagent_enabled"`
	MaxSubAgents    int      `json:"max_sub_agents"`
	TimeoutSeconds  int      `json:"timeout_seconds"`
	ResumeFrom      string   `json:"resume_from,omitempty"`
	ConversationID  string   `json:"conversation_id,omitempty"`
	TavilyAPIKey      string   `json:"tavily_api_key,omitempty"`
	DuckDuckGoEnabled bool     `json:"duckduckgo_enabled"`
	FileAllowedPaths  []string `json:"file_allowed_paths,omitempty"`
	SandboxDir        string   `json:"sandbox_dir,omitempty"`
	PythonReplEnabled bool     `json:"python_repl_enabled"`
	BashEnabled       bool     `json:"bash_enabled"`
}
