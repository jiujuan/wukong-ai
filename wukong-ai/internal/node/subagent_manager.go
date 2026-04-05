package node

import (
	"encoding/json"
	"fmt"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/subagent"
	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// SubAgentManager 子代理管理器节点 - 调度并行子任务，将 ToolRegistry 透传给每个 SubAgent
type SubAgentManager struct {
	scheduler    *subagent.Scheduler
	maxAgents    int
	toolRegistry *tools.ToolRegistry
}

// NewSubAgentManager 创建子代理管理器节点
func NewSubAgentManager(scheduler *subagent.Scheduler, maxAgents int, toolRegistry *tools.ToolRegistry) *SubAgentManager {
	return &SubAgentManager{
		scheduler:    scheduler,
		maxAgents:    maxAgents,
		toolRegistry: toolRegistry,
	}
}

// Name 返回节点名称
func (s *SubAgentManager) Name() string {
	return "subagentmanager"
}

// Run 执行子代理管理逻辑
func (s *SubAgentManager) Run(ctx *workflow.WukongContext) error {
	logger.Info(
		"SubAgentManager running",
		"task_id", ctx.Config.TaskID,
		"tasks", len(ctx.State.Tasks),
		"max_agents", s.maxAgents,
	)

	if s.scheduler == nil {
		err := fmt.Errorf("subagent scheduler is nil")
		logger.Error("SubAgentManager invalid scheduler", "task_id", ctx.Config.TaskID, "err", err)
		return err
	}

	if len(ctx.State.Tasks) == 0 {
		logger.Info("No tasks to execute, skipping SubAgentManager", "task_id", ctx.Config.TaskID)
		return nil
	}

	tasks := ctx.State.Tasks
	if len(tasks) > s.maxAgents {
		tasks = tasks[:s.maxAgents]
		logger.Info("Limited tasks to max agents", "task_id", ctx.Config.TaskID, "limited_to", s.maxAgents)
	}

	subTasks := make([]*subagent.SubTask, len(tasks))
	for i, task := range tasks {
		record := &repository.SubAgentRecord{
			TaskID:     ctx.Config.TaskID,
			AgentIndex: i,
			SubTask:    task,
			Status:     "pending",
		}
		recordID, err := repository.CreateSubAgentRecord(record)
		if err != nil {
			logger.Warn("failed to create sub agent record", "task_id", ctx.Config.TaskID, "agent_index", i, "err", err)
		}
		subTasks[i] = &subagent.SubTask{
			ID:     recordID,
			TaskID: ctx.Config.TaskID,
			Index:  i,
			Input:  task,
		}
	}

	// // 执行所有子任务，并且将 toolRegistry 传给 Scheduler，由 Scheduler 传给每个 SubAgent
	results, err := s.scheduler.RunAll(ctx, subTasks, s.toolRegistry)
	if err != nil {
		logger.Error("sub agents execution failed", "task_id", ctx.Config.TaskID, "err", err)
		return err
	}

	for i, result := range results {
		if i >= len(subTasks) {
			logger.Warn("sub agent result index out of range", "task_id", ctx.Config.TaskID, "result_index", i, "task_count", len(subTasks))
			continue
		}
		ctx.State.AddSubAgentResult(result)
		if ctx.EventBus != nil && result != "" {
			ctx.EventBus.Publish(ctx.Config.TaskID, workflow.ProgressEvent{
				Type:   "sub_agent_update",
				Node:   "subagentmanager",
				Status: "running",
				Done:   i + 1,
				Total:  len(results),
				Latest: result,
			})
		}
		if err := repository.UpdateSubAgentRecord(subTasks[i].ID, "success", result, ""); err != nil {
			logger.Warn("failed to update sub agent record", "task_id", ctx.Config.TaskID, "agent_index", i, "record_id", subTasks[i].ID, "err", err)
		}
		logger.Info("sub agent completed", "task_id", ctx.Config.TaskID, "agent_index", i, "result_length", len(result))
	}

	if len(results) > 0 {
		subResultsJSON, _ := json.Marshal(results)
		if err := repository.UpdateSubResults(ctx.Config.TaskID, subResultsJSON); err != nil {
			logger.Warn("failed to save sub results", "task_id", ctx.Config.TaskID, "err", err)
		}
	}

	logger.Info("SubAgentManager completed", "task_id", ctx.Config.TaskID, "completed", len(results))
	return nil
}
