package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/node"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/subagent"
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
func (p *Pool) runWorker(ctx context.Context, workerID string, llmProvider llm.LLM, agentCfg *config.AgentConfig, promptDir string) {
	defer p.wg.Done()

	logger.Info("worker started", "worker_id", workerID)

	// 创建节点
	scheduler := subagent.NewScheduler(agentCfg.MaxSubAgents)

	// 循环获取任务
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

		// 尝试获取任务
		job, err := p.queue.Dequeue(ctx, workerID)
		if err != nil {
			logger.Error("dequeue failed", "worker_id", workerID, "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if job == nil {
			// 队列为空，等待
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Info("worker picked job", "worker_id", workerID, "task_id", job.TaskID)

		// 执行任务
		p.executeJob(ctx, job, llmProvider, agentCfg, promptDir, scheduler)
	}
}

// executeJob 执行任务
func (p *Pool) executeJob(ctx context.Context, job *queue.TaskJob, llmProvider llm.LLM, agentCfg *config.AgentConfig, promptDir string, scheduler *subagent.Scheduler) {
	// 解析请求
	var req RunRequest
	if err := json.Unmarshal(job.Payload, &req); err != nil {
		logger.Error("failed to unmarshal job payload", "task_id", job.TaskID, "err", err)
		p.queue.MarkFailed(job.TaskID, err.Error())
		return
	}

	// 创建执行上下文
	cfg := workflow.NewRunConfig(agentCfg)
	cfg.TaskID = job.TaskID
	cfg.Mode = workflow.AutoSelectMode(cfg)
	cfg.ThinkingEnabled = req.ThinkingEnabled
	cfg.PlanEnabled = req.PlanEnabled
	cfg.SubAgentEnabled = req.SubAgentEnabled
	if req.MaxSubAgents > 0 {
		cfg.MaxSubAgents = req.MaxSubAgents
	}
	if req.TimeoutSeconds > 0 {
		cfg.Timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}

	wCtx := workflow.NewWukongContext(cfg, llmProvider, req.UserInput)
	wCtx.State.Mode = cfg.Mode.String()
	wCtx.EventBus = p.eventBus

	// 创建工作流
	nodes := createNodes(llmProvider, scheduler, promptDir)
	wf := workflow.BuildWorkflow(cfg.Mode, nodes)

	// 创建引擎
	engine := workflow.NewEngine(wf, p.eventBus)

	// 检查是否需要从断点恢复
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
		return
	}

	// 完成任务
	engine.Complete(wCtx)
	p.queue.MarkSuccess(job.TaskID)
	logger.Info("job completed", "task_id", job.TaskID)
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

// createNodes 创建工作流节点
func createNodes(llmProvider llm.LLM, scheduler *subagent.Scheduler, promptDir string) *workflow.NodeSet {
	return &workflow.NodeSet{
		Coordinator:     node.NewCoordinator(llmProvider, promptDir),
		Background:      node.NewBackground(llmProvider, promptDir),
		Planner:         node.NewPlanner(llmProvider, promptDir),
		Researcher:      node.NewResearcher(llmProvider, promptDir),
		SubAgentManager: node.NewSubAgentManager(scheduler, 3),
		Reporter:        node.NewReporter(llmProvider, promptDir),
	}
}

// RunRequest 运行请求
type RunRequest struct {
	UserInput       string `json:"user_input"`
	ThinkingEnabled bool   `json:"thinking_enabled"`
	PlanEnabled     bool   `json:"plan_enabled"`
	SubAgentEnabled bool   `json:"subagent_enabled"`
	MaxSubAgents    int    `json:"max_sub_agents"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	ResumeFrom      string `json:"resume_from,omitempty"`
}
