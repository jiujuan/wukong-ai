package state

import (
	"encoding/json"
)

// RunState 运行状态
type RunState struct {
	TaskID          string   `json:"task_id"`
	Status          string   `json:"status"`          // pending/running/success/failed
	LastNode        string   `json:"last_node"`        // 断点续跑核心字段
	UserInput       string   `json:"user_input"`
	Intention       string   `json:"intention"`
	Plan            string   `json:"plan"`
	Tasks           []string `json:"tasks"`
	SubAgentResults []string `json:"sub_results"`
	FinalOutput     string   `json:"final_output"`
	CreateTime      string   `json:"create_time"`
	FinishTime      string   `json:"finish_time"`
	RetryCount      int      `json:"retry_count"`
	ErrorMsg        string   `json:"error_msg,omitempty"`
	Mode            string   `json:"mode"`
}

// NewRunState 创建新的运行状态
func NewRunState(taskID, userInput string) *RunState {
	return &RunState{
		TaskID:     taskID,
		Status:     "pending",
		UserInput:  userInput,
		CreateTime: "",
		Tasks:      []string{},
		SubAgentResults: []string{},
	}
}

// NewRunStateWithMode 创建带模式的状态
func NewRunStateWithMode(taskID, userInput, mode string) *RunState {
	state := NewRunState(taskID, userInput)
	state.Mode = mode
	return state
}

// SetStatus 设置状态
func (s *RunState) SetStatus(status string) {
	s.Status = status
}

// SetIntention 设置意图
func (s *RunState) SetIntention(intention string) {
	s.Intention = intention
}

// SetPlan 设置计划
func (s *RunState) SetPlan(plan string) {
	s.Plan = plan
}

// SetTasks 设置任务列表
func (s *RunState) SetTasks(tasks []string) {
	s.Tasks = tasks
}

// AddSubAgentResult 添加子代理结果
func (s *RunState) AddSubAgentResult(result string) {
	s.SubAgentResults = append(s.SubAgentResults, result)
}

// SetFinalOutput 设置最终输出
func (s *RunState) SetFinalOutput(output string) {
	s.FinalOutput = output
}

// SetLastNode 设置最后节点
func (s *RunState) SetLastNode(node string) {
	s.LastNode = node
}

// SetError 设置错误
func (s *RunState) SetError(errMsg string) {
	s.ErrorMsg = errMsg
}

// IncrementRetry 增加重试计数
func (s *RunState) IncrementRetry() {
	s.RetryCount++
}

// ToJSON 序列化为 JSON
func (s *RunState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON 从 JSON 反序列化
func (s *RunState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}
