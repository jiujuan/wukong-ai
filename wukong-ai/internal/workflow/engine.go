package workflow

import (
	"fmt"
	"reflect"
	"time"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Engine DAG 执行引擎
type Engine struct {
	wf            *Workflow
	eventBus     EventBus
	currentStep  int
	totalSteps   int
}

// EventBus 事件总线接口
type EventBus interface {
	Publish(taskID string, event ProgressEvent)
}

// NewEngine 创建新的执行引擎
func NewEngine(wf *Workflow, eventBus EventBus) *Engine {
	totalSteps := wf.CalculateTotalSteps()
	return &Engine{
		wf:          wf,
		eventBus:    eventBus,
		currentStep: 0,
		totalSteps:  totalSteps,
	}
}

// RunFromStart 从头执行 DAG
func (e *Engine) RunFromStart(ctx *WukongContext) error {
	ctx.State.Status = "running"
	ctx.State.CreateTime = time.Now().Format(time.RFC3339)
	e.currentStep = 0

	// 保存初始状态
	if err := e.saveState(ctx); err != nil {
		logger.Warn("failed to save initial state", "err", err)
	}

	return e.runNode(ctx, e.wf.Start)
}

// RunFromBreakpoint 从断点恢复执行
func (e *Engine) RunFromBreakpoint(ctx *WukongContext) error {
	if ctx.State.LastNode == "" {
		return e.RunFromStart(ctx)
	}

	// 恢复运行状态
	ctx.State.Status = "running"

	node, ok := e.wf.Nodes[ctx.State.LastNode]
	if !ok {
		return fmt.Errorf("last node not found: %s", ctx.State.LastNode)
	}

	// 计算当前步骤
	e.currentStep = e.calculateCurrentStep(ctx.State.LastNode)

	return e.runNode(ctx, node)
}

// runNode 执行单个节点并推进到后继节点
func (e *Engine) runNode(ctx *WukongContext, node Node) error {
	if node == nil {
		return nil
	}

	nodeName := node.Name()
	logger.Info("executing node", "task_id", ctx.Config.TaskID, "node", nodeName)

	// 创建节点日志
	startTime := time.Now()
	logID, err := e.createNodeLog(ctx.Config.TaskID, nodeName, "running", "")
	if err != nil {
		logger.Warn("failed to create node log", "err", err)
	}

	// 发布开始事件
	e.publishEvent(ctx, ProgressEvent{
		Type:     "node_start",
		Node:     nodeName,
		Status:   "running",
		Progress: e.currentStep + 1,
		Total:    e.totalSteps,
	})

	// 执行节点
	var nodeErr error
	output := ""
	func() {
		defer func() {
			if p := recover(); p != nil {
				nodeErr = fmt.Errorf("node panic: %v", p)
			}
		}()
		nodeErr = node.Run(ctx)
		if nodeErr == nil {
			// 从 context 获取输出
			output = ctx.State.LastNodeOutput
		}
	}()

	// 更新节点日志
	durationMs := time.Since(startTime).Milliseconds()
	if nodeErr != nil {
		e.updateNodeLog(logID, "failed", "", nodeErr.Error(), durationMs)
		e.publishEvent(ctx, ProgressEvent{
			Type:   "node_done",
			Node:   nodeName,
			Status: "failed",
		})
		ctx.State.Status = "failed"
		ctx.State.ErrorMsg = nodeErr.Error()
		ctx.State.RetryCount++
		if err := e.saveState(ctx); err != nil {
			logger.Warn("failed to save state after node failure", "err", err)
		}
		return fmt.Errorf("node %s failed: %w", nodeName, nodeErr)
	}

	e.updateNodeLog(logID, "success", output, "", durationMs)
	e.publishEvent(ctx, ProgressEvent{
		Type:     "node_done",
		Node:     nodeName,
		Status:   "success",
		Progress: e.currentStep + 1,
		Total:    e.totalSteps,
	})

	// 更新上下文状态
	ctx.State.LastNode = nodeName
	if err := e.saveState(ctx); err != nil {
		logger.Warn("failed to save state after node completion", "err", err)
	}

	// 执行后继节点
	for _, next := range e.wf.Edges[nodeName] {
		e.currentStep++
		if err := e.runNode(ctx, next); err != nil {
			return err
		}
	}

	return nil
}

// publishEvent 发布事件到事件总线
func (e *Engine) publishEvent(ctx *WukongContext, event ProgressEvent) {
	if e.hasEventBus() {
		event.Timestamp = time.Now().Format(time.RFC3339)
		e.eventBus.Publish(ctx.Config.TaskID, event)
	}
}

// calculateCurrentStep 计算当前步骤数
func (e *Engine) calculateCurrentStep(lastNode string) int {
	visited := make(map[string]bool)
	step := 0
	e.countStepsToNode(e.wf.Start, lastNode, visited, &step)
	return step
}

func (e *Engine) countStepsToNode(node Node, target string, visited map[string]bool, step *int) bool {
	if node == nil || visited[node.Name()] {
		return false
	}
	visited[node.Name()] = true
	*step++

	if node.Name() == target {
		return true
	}

	for _, successor := range e.wf.Edges[node.Name()] {
		if e.countStepsToNode(successor, target, visited, step) {
			return true
		}
	}

	return false
}

// saveState 保存状态
func (e *Engine) saveState(ctx *WukongContext) error {
	// 更新数据库
	if err := repository.UpdateTaskStatus(ctx.Config.TaskID, ctx.State.Status); err != nil {
		return err
	}
	if ctx.State.LastNode != "" {
		if err := repository.UpdateTaskLastNode(ctx.Config.TaskID, ctx.State.LastNode); err != nil {
			return err
		}
	}
	if ctx.State.Intention != "" {
		if err := repository.UpdateTaskResult(ctx.Config.TaskID, ctx.State.Intention, ctx.State.Plan, ctx.State.FinalOutput); err != nil {
			return err
		}
	}
	if ctx.State.ErrorMsg != "" {
		if err := repository.UpdateTaskError(ctx.Config.TaskID, ctx.State.ErrorMsg, ctx.State.RetryCount); err != nil {
			return err
		}
	}
	return nil
}

// createNodeLog 创建节点日志
func (e *Engine) createNodeLog(taskID, nodeName, status, input string) (int64, error) {
	log := &repository.NodeExecutionLog{
		TaskID:    taskID,
		NodeName:  nodeName,
		Status:    status,
		StartTime: time.Now(),
	}
	if input != "" {
		log.Input.String = input
		log.Input.Valid = true
	}
	return repository.CreateNodeLog(log)
}

// updateNodeLog 更新节点日志
func (e *Engine) updateNodeLog(id int64, status, output, errorMsg string, durationMs int64) {
	if err := repository.UpdateNodeLog(id, status, output, errorMsg, durationMs); err != nil {
		logger.Warn("failed to update node log", "log_id", id, "err", err)
	}
}

// Complete 完成任务
func (e *Engine) Complete(ctx *WukongContext) {
	ctx.State.Status = "success"
	ctx.State.FinishTime = time.Now().Format(time.RFC3339)

	if err := repository.CompleteTask(ctx.Config.TaskID, ctx.State.FinalOutput); err != nil {
		logger.Error("failed to complete task", "task_id", ctx.Config.TaskID, "err", err)
	}

	e.publishEvent(ctx, ProgressEvent{
		Type:   "task_done",
		Status: "success",
		Output: ctx.State.FinalOutput,
	})
}

// Fail 标记任务失败
func (e *Engine) Fail(ctx *WukongContext, errMsg string) {
	ctx.State.Status = "failed"
	ctx.State.ErrorMsg = errMsg
	ctx.State.FinishTime = time.Now().Format(time.RFC3339)

	if err := repository.FailTask(ctx.Config.TaskID, errMsg); err != nil {
		logger.Error("failed to fail task", "task_id", ctx.Config.TaskID, "err", err)
	}

	e.publishEvent(ctx, ProgressEvent{
		Type:   "task_failed",
		Status: "failed",
	})
}

// Publish 实现 EventBus 接口
func (e *Engine) Publish(taskID string, event ProgressEvent) {
	if e == nil {
		return
	}
	if e.hasEventBus() {
		event.Timestamp = time.Now().Format(time.RFC3339)
		e.eventBus.Publish(taskID, event)
	}
}

func (e *Engine) hasEventBus() bool {
	if e == nil || e.eventBus == nil {
		return false
	}
	value := reflect.ValueOf(e.eventBus)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}
