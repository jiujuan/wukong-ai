package state

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Manager 状态管理器
type Manager struct {
	store *Store
	mu    sync.RWMutex
}

// NewManager 创建状态管理器
func NewManager(taskDir string) *Manager {
	return &Manager{
		store: NewStore(taskDir),
	}
}

// Create 创建新任务状态
func (m *Manager) Create(taskID, userInput, mode string) (*RunState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := NewRunStateWithMode(taskID, userInput, mode)

	// 保存到文件
	if err := m.store.Save(state); err != nil {
		return nil, err
	}

	// 保存到数据库
	task := &repository.Task{
		ID:              taskID,
		Status:          "pending",
		Mode:            mode,
		UserInput:       userInput,
		ThinkingEnabled: false,
		PlanEnabled:     false,
		SubagentEnabled: false,
	}
	if err := repository.CreateTask(task); err != nil {
		logger.Warn("failed to create task in database", "err", err)
	}

	return state, nil
}

// Get 获取任务状态
func (m *Manager) Get(taskID string) (*RunState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.store.Load(taskID)
}

// Save 保存任务状态
func (m *Manager) Save(state *RunState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.store.Save(state)
}

// UpdateStatus 更新任务状态
func (m *Manager) UpdateStatus(taskID, status string) error {
	// 更新文件
	state, err := m.store.Load(taskID)
	if err != nil || state == nil {
		return fmt.Errorf("state not found: %s", taskID)
	}

	state.SetStatus(status)
	if err := m.store.Save(state); err != nil {
		return err
	}

	// 更新数据库
	return repository.UpdateTaskStatus(taskID, status)
}

// UpdateResult 更新任务结果
func (m *Manager) UpdateResult(taskID, intention, plan, finalOutput string) error {
	// 更新文件
	state, err := m.store.Load(taskID)
	if err != nil || state == nil {
		return fmt.Errorf("state not found: %s", taskID)
	}

	state.SetIntention(intention)
	state.SetPlan(plan)
	state.SetFinalOutput(finalOutput)

	if err := m.store.Save(state); err != nil {
		return err
	}

	// 更新数据库
	return repository.UpdateTaskResult(taskID, intention, plan, finalOutput)
}

// UpdateLastNode 更新最后节点（断点续跑）
func (m *Manager) UpdateLastNode(taskID, lastNode string) error {
	state, err := m.store.Load(taskID)
	if err != nil || state == nil {
		return fmt.Errorf("state not found: %s", taskID)
	}

	state.SetLastNode(lastNode)
	return m.store.Save(state)
}

// Complete 完成任务
func (m *Manager) Complete(taskID, finalOutput string) error {
	state, err := m.store.Load(taskID)
	if err != nil || state == nil {
		return fmt.Errorf("state not found: %s", taskID)
	}

	state.SetStatus("success")
	state.SetFinalOutput(finalOutput)

	if err := m.store.Save(state); err != nil {
		return err
	}

	// 更新数据库
	return repository.CompleteTask(taskID, finalOutput)
}

// Fail 标记任务失败
func (m *Manager) Fail(taskID, errorMsg string) error {
	state, err := m.store.Load(taskID)
	if err != nil || state == nil {
		return fmt.Errorf("state not found: %s", taskID)
	}

	state.SetStatus("failed")
	state.SetError(errorMsg)

	if err := m.store.Save(state); err != nil {
		return err
	}

	// 更新数据库
	return repository.FailTask(taskID, errorMsg)
}

// List 列出所有任务
func (m *Manager) List() ([]*RunState, error) {
	taskIDs, err := m.store.List()
	if err != nil {
		return nil, err
	}

	var states []*RunState
	for _, taskID := range taskIDs {
		state, err := m.store.Load(taskID)
		if err != nil {
			continue
		}
		if state != nil {
			states = append(states, state)
		}
	}

	return states, nil
}

// MarshalState 序列化状态为 JSON
func MarshalState(state *RunState) ([]byte, error) {
	return json.MarshalIndent(state, "", "  ")
}

// UnmarshalState 从 JSON 反序列化状态
func UnmarshalState(data []byte) (*RunState, error) {
	state := &RunState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}
	return state, nil
}
