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
	db         *repository.MemoryRepository
	llmProvider llm.LLM
	topK       int
}

// MemoryRepository 数据库仓库
type MemoryRepository = repository.Memory

// NewLongTermMemory 创建长期记忆
func NewLongTermMemory(db repository.MemoryRepository, llmProvider llm.LLM, topK int) *LongTermMemory {
	return &LongTermMemory{
		db:         db,
		llmProvider: llmProvider,
		topK:       topK,
	}
}

// Save 保存记忆
func (m *LongTermMemory) Save(ctx context.Context, taskID, content string, memoryType MemoryType) error {
	// 生成向量
	embedding, err := m.llmProvider.Embed(ctx, content)
	if err != nil {
		logger.Warn("failed to generate embedding", "err", err)
		// 即使 embedding 失败也保存文本
		embedding = make([]float32, 1536)
	}

	// 保存到数据库
	memory := &repository.Memory{
		Content: content,
		Embedding: pgvector.NewVector(embedding),
		MemoryType: string(memoryType),
	}
	if taskID != "" {
		memory.TaskID.String = taskID
		memory.TaskID.Valid = true
	}

	_, err = repository.SaveMemory(memory)
	if err != nil {
		return err
	}

	logger.Debug("long term memory saved", "task_id", taskID, "type", memoryType)
	return nil
}

// Query 查询记忆
func (m *LongTermMemory) Query(ctx context.Context, taskID, query string) ([]string, error) {
	// 生成查询向量
	embedding, err := m.llmProvider.Embed(ctx, query)
	if err != nil {
		logger.Error("failed to generate query embedding", "err", err)
		return []string{}, nil
	}

	// 搜索相似记忆
	memories, err := repository.SearchMemoriesByEmbedding(
		pgvector.NewVector(embedding),
		taskID,
		m.topK,
	)
	if err != nil {
		return []string{}, err
	}

	// 提取内容
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
