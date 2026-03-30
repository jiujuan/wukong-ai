package workflow

import (
	"sync"
	"time"
)

// RunState 运行状态
type RunState struct {
	TaskID          string   `json:"task_id"`
	Status          string   `json:"status"`            // pending/running/success/failed/queued
	Mode            string   `json:"mode"`              // flash/standard/pro/ultra
	UserInput       string   `json:"user_input"`
	Intention       string   `json:"intention"`         // 用户意图解析结果
	Plan            string   `json:"plan"`              // 执行计划
	Tasks           []string `json:"tasks"`             // Planner 生成的子任务列表
	SubAgentResults []string `json:"sub_results"`       // SubAgent 各子任务结果
	FinalOutput     string   `json:"final_output"`      // 最终输出
	LastNode        string   `json:"last_node"`          // 最后执行的节点名
	LastNodeOutput  string   `json:"last_node_output"`   // 最后节点输出
	CreateTime      string   `json:"create_time"`
	FinishTime      string   `json:"finish_time"`
	RetryCount      int      `json:"retry_count"`
	ErrorMsg        string   `json:"error_msg,omitempty"`
	ThinkingEnabled bool     `json:"thinking_enabled"`
	PlanEnabled     bool     `json:"plan_enabled"`
	SubAgentEnabled bool     `json:"subagent_enabled"`
	mu              sync.RWMutex
}

// NewRunState 创建新的运行状态
func NewRunState(taskID, userInput string) *RunState {
	return &RunState{
		TaskID:     taskID,
		Status:     "pending",
		UserInput:  userInput,
		CreateTime: time.Now().Format(time.RFC3339),
	}
}

// SetStatus 设置状态
func (s *RunState) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// GetStatus 获取状态
func (s *RunState) GetStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// SetIntention 设置意图
func (s *RunState) SetIntention(intention string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Intention = intention
}

// SetPlan 设置计划
func (s *RunState) SetPlan(plan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Plan = plan
}

// SetTasks 设置任务列表
func (s *RunState) SetTasks(tasks []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tasks = tasks
}

// AddSubAgentResult 添加子 Agent 结果
func (s *RunState) AddSubAgentResult(result string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SubAgentResults = append(s.SubAgentResults, result)
}

// SetFinalOutput 设置最终输出
func (s *RunState) SetFinalOutput(output string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FinalOutput = output
}

// SetLastNode 设置最后节点
func (s *RunState) SetLastNode(node string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastNode = node
}

// SetLastNodeOutput 设置最后节点输出
func (s *RunState) SetLastNodeOutput(output string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastNodeOutput = output
}

// SetError 设置错误
func (s *RunState) SetError(errMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ErrorMsg = errMsg
}

// IncrementRetry 增加重试计数
func (s *RunState) IncrementRetry() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RetryCount++
}
