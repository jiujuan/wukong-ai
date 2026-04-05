package memory

import (
	"context"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/pgvector/pgvector-go"
)

// LongTermMemory 长期记忆（pgvector 向量存储）
type LongTermMemory struct {
	llmProvider llm.LLM
	topK        int
}

// NewLongTermMemory 创建长期记忆
func NewLongTermMemory(llmProvider llm.LLM, topK int) *LongTermMemory {
	return &LongTermMemory{
		llmProvider: llmProvider,
		topK:        topK,
	}
}

// Save 保存记忆（生成向量 + 写 DB）
func (m *LongTermMemory) Save(ctx context.Context, taskID, content string, memoryType MemoryType) error {
	embedding, err := m.llmProvider.Embed(ctx, content)
	if err != nil {
		logger.Warn("failed to generate embedding, using zero vector", "err", err)
		embedding = make([]float32, 1536)
	}

	mem := &repository.Memory{
		Content:    content,
		Embedding:  pgvector.NewVector(embedding),
		MemoryType: string(memoryType),
	}
	if taskID != "" {
		mem.TaskID.String = taskID
		mem.TaskID.Valid = true
	}

	_, err = repository.SaveMemory(mem)
	if err != nil {
		return err
	}

	logger.Debug("long term memory saved", "task_id", taskID, "type", memoryType)
	return nil
}

// Query 向量检索记忆
func (m *LongTermMemory) Query(ctx context.Context, taskID, query string) ([]string, error) {
	embedding, err := m.llmProvider.Embed(ctx, query)
	if err != nil {
		logger.Error("failed to generate query embedding", "err", err)
		return []string{}, nil
	}

	memories, err := repository.SearchMemoriesByEmbedding(
		pgvector.NewVector(embedding), taskID, m.topK,
	)
	if err != nil {
		return []string{}, err
	}

	results := make([]string, len(memories))
	for i, mem := range memories {
		results[i] = mem.Content
	}
	return results, nil
}

// GetByTaskID 获取任务的所有记忆
func (m *LongTermMemory) GetByTaskID(taskID string) ([]string, error) {
	memories, err := repository.GetMemoriesByTaskID(taskID)
	if err != nil {
		return []string{}, err
	}

	results := make([]string, len(memories))
	for i, mem := range memories {
		results[i] = mem.Content
	}
	return results, nil
}
