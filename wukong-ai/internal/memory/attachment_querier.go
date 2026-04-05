package memory

import (
	"context"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/pgvector/pgvector-go"
)

// AttachmentQuerier 实现 workflow.AttachmentMemoryQuerier 接口
// 通过向量相似度检索 memories 表中 memory_type='attachment' 的记录
type AttachmentQuerier struct {
	llmProvider llm.LLM
	topK        int
}

// NewAttachmentQuerier 创建附件查询器
func NewAttachmentQuerier(llmProvider llm.LLM, topK int) *AttachmentQuerier {
	if topK <= 0 {
		topK = 5
	}
	return &AttachmentQuerier{llmProvider: llmProvider, topK: topK}
}

// QueryAttachments 用 query 文本向量检索最相关的附件 Chunk
func (q *AttachmentQuerier) QueryAttachments(ctx context.Context, taskID, query string, topK int) ([]string, error) {
	if topK <= 0 {
		topK = q.topK
	}

	// 生成查询向量
	embedding, err := q.llmProvider.Embed(ctx, query)
	if err != nil {
		logger.Warn("attachment query embedding failed, returning empty", "err", err)
		return nil, nil
	}

	mems, err := repository.SearchAttachmentMemories(pgvector.NewVector(embedding), taskID, topK)
	if err != nil {
		return nil, err
	}

	results := make([]string, 0, len(mems))
	for _, m := range mems {
		if m.Content != "" {
			results = append(results, m.Content)
		}
	}
	return results, nil
}
