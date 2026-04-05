package subagent

import (
	"context"
	"fmt"
	"sync"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/internal/workflow"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Scheduler 并行调度器
type Scheduler struct {
	maxConcurrent int
	subAgentPool  *PoolWithWaitGroup
}

// NewScheduler 创建新的调度器
func NewScheduler(maxConcurrent int) *Scheduler {
	return &Scheduler{
		maxConcurrent: maxConcurrent,
		subAgentPool:  NewPoolWithWaitGroup(maxConcurrent),
	}
}

// RunAll 并行执行所有子任务，toolRegistry 透传给每个 SubAgent
func (s *Scheduler) RunAll(ctx *workflow.WukongContext, tasks []*SubTask, toolRegistry *tools.ToolRegistry) ([]string, error) {
	if len(tasks) == 0 {
		return []string{}, nil
	}

	logger.Info("Scheduler starting", "task_id", ctx.Config.TaskID, "task_count", len(tasks))
	execCtx := ctx.Context
	if execCtx == nil {
		execCtx = context.Background()
	}

	results := make([]string, len(tasks))
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.maxConcurrent)

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t *SubTask) {
			defer wg.Done()

			// 信号量限流
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 每个 SubAgent 共享同一个 toolRegistry（只读，线程安全）
			agent := NewSubAgent(ctx.LLMProvider, ctx.Config.Mode.String(), toolRegistry)
			result, err := agent.Execute(execCtx, t)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[idx] = fmt.Sprintf("Error: %v", err)
				logger.Error("sub agent failed", "task_id", t.TaskID, "index", t.Index, "err", err)
			} else {
				results[idx] = result
			}
		}(i, task)
	}

	wg.Wait()

	var errors []error
	for i, result := range results {
		if len(result) > 6 && result[:6] == "Error:" {
			errors = append(errors, fmt.Errorf("task %d: %s", i, result[7:]))
		}
	}
	if len(errors) > 0 {
		logger.Warn("some sub agents failed", "task_id", ctx.Config.TaskID, "failed_count", len(errors))
	}

	logger.Info("Scheduler completed", "task_id", ctx.Config.TaskID, "completed", len(results))
	return results, nil
}
