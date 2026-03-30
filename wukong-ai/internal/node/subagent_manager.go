package node

import (
	"encoding/json"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/subagent"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// SubAgentManager 子代理管理器节点 - 管理多个子代理并行执行
type SubAgentManager struct {
	scheduler *subagent.Scheduler
	maxAgents int
}

// NewSubAgentManager 创建子代理管理器节点
func NewSubAgentManager(scheduler *subagent.Scheduler, maxAgents int) *SubAgentManager {
	return &SubAgentManager{
		scheduler: scheduler,
		maxAgents: maxAgents,
	}
}

// Name 返回节点名称
func (s *SubAgentManager) Name() string {
	return "subagentmanager"
}

// Run 执行子代理管理逻辑
func (s *SubAgentManager) Run(ctx *workflow.WukongContext) error {
	logger.Info("SubAgentManager running", "task_id", ctx.Config.TaskID, "tasks", len(ctx.State.Tasks))

	if len(ctx.State.Tasks) == 0 {
		logger.Info("No tasks to execute, skipping SubAgentManager", "task_id", ctx.Config.TaskID)
		return nil
	}

	// 限制并发数
	tasks := ctx.State.Tasks
	if len(tasks) > s.maxAgents {
		tasks = tasks[:s.maxAgents]
		logger.Info("Limited tasks to max agents", "task_id", ctx.Config.TaskID, "limited_to", s.maxAgents)
	}

	// 创建子代理任务
	subTasks := make([]*subagent.SubTask, len(tasks))
	for i, task := range tasks {
		// 创建子代理记录
		record := &repository.SubAgentRecord{
			TaskID:     ctx.Config.TaskID,
			AgentIndex: i,
			SubTask:    task,
			Status:     "pending",
		}
		recordID, err := repository.CreateSubAgentRecord(record)
		if err != nil {
			logger.Warn("failed to create sub agent record", "err", err)
		}

		subTasks[i] = &subagent.SubTask{
			ID:     recordID,
			TaskID: ctx.Config.TaskID,
			Index:  i,
			Input:  task,
		}
	}

	// 执行所有子任务
	results, err := s.scheduler.RunAll(ctx, subTasks)
	if err != nil {
		logger.Error("sub agents execution failed", "task_id", ctx.Config.TaskID, "err", err)
		return err
	}

	// 保存结果
	for i, result := range results {
		ctx.State.AddSubAgentResult(result)
		if err := repository.UpdateSubAgentRecord(subTasks[i].ID, "success", result, ""); err != nil {
			logger.Warn("failed to update sub agent record", "err", err)
		}
	}

	// 更新数据库中的子结果
	if len(results) > 0 {
		subResultsJSON, _ := json.Marshal(results)
		if err := repository.UpdateSubResults(ctx.Config.TaskID, subResultsJSON); err != nil {
			logger.Warn("failed to save sub results", "err", err)
		}
	}

	logger.Info("SubAgentManager completed", "task_id", ctx.Config.TaskID, "completed", len(results))
	return nil
}
