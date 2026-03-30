package subagent

import (
	"fmt"
	"sync"

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

// RunAll 并行执行所有子任务
func (s *Scheduler) RunAll(ctx *workflow.WukongContext, tasks []*SubTask) ([]string, error) {
	if len(tasks) == 0 {
		return []string{}, nil
	}

	logger.Info("Scheduler starting", "task_id", ctx.Config.TaskID, "task_count", len(tasks))

	results := make([]string, len(tasks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t *SubTask) {
			defer wg.Done()

			// 创建子代理
			agent := NewSubAgent(ctx.LLMProvider, ctx.Config.Mode.String())

			// 执行任务
			result, err := agent.Execute(t)
			mu.Lock()
			if err != nil {
				results[idx] = fmt.Sprintf("Error: %v", err)
				logger.Error("sub agent failed", "task_id", t.TaskID, "index", t.Index, "err", err)
			} else {
				results[idx] = result
			}
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()

	// 检查是否有错误
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
