package state

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jiujuan/wukong-ai/pkg/logger"
)

const defaultTaskDir = "tasks"

// Store 状态持久化存储
type Store struct {
	taskDir string
	mu      sync.RWMutex
	cache   map[string]*RunState
}

// NewStore 创建状态存储
func NewStore(taskDir string) *Store {
	if taskDir == "" {
		taskDir = defaultTaskDir
	}
	return &Store{
		taskDir: taskDir,
		cache:   make(map[string]*RunState),
	}
}

// getFilePath 获取文件路径
func (s *Store) getFilePath(taskID string) string {
	return filepath.Join(s.taskDir, taskID+".json")
}

// Save 保存状态
func (s *Store) Save(state *RunState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := state.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// 确保目录存在
	if err := os.MkdirAll(s.taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task dir: %w", err)
	}

	path := s.getFilePath(state.TaskID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	// 更新缓存
	s.cache[state.TaskID] = state

	logger.Debug("state saved", "task_id", state.TaskID)
	return nil
}

// Load 加载状态
func (s *Store) Load(taskID string) (*RunState, error) {
	s.mu.RLock()
	if state, ok := s.cache[taskID]; ok {
		s.mu.RUnlock()
		return state, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 再次检查缓存
	if state, ok := s.cache[taskID]; ok {
		return state, nil
	}

	path := s.getFilePath(taskID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	state := &RunState{}
	if err := state.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	s.cache[taskID] = state
	return state, nil
}

// Delete 删除状态
func (s *Store) Delete(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.getFilePath(taskID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	delete(s.cache, taskID)
	return nil
}

// List 列出所有任务状态文件
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.taskDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read task dir: %w", err)
	}

	var taskIDs []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			taskID := entry.Name()[:len(entry.Name())-5]
			taskIDs = append(taskIDs, taskID)
		}
	}

	return taskIDs, nil
}

// ClearCache 清除缓存
func (s *Store) ClearCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]*RunState)
}

// GetCached 获取缓存的状态
func (s *Store) GetCached(taskID string) (*RunState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.cache[taskID]
	return state, ok
}
