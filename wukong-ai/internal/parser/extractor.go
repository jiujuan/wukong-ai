package parser

import (
	"context"
	"sync"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/pgvector/pgvector-go"
)

// EmbedFunc 向量化函数类型（由 llm.LLM.Embed 提供）
type EmbedFunc func(ctx context.Context, text string) ([]float32, error)

// ExtractJob 异步提取任务
type ExtractJob struct {
	AttachmentID int64
	TaskID       string
}

// Extractor 异步内容提取调度器
// 负责：解析文件 → 分块 → 向量化 → 存入 memories 表
type Extractor struct {
	jobs     chan ExtractJob
	embedFn  EmbedFunc
	workers  int
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewExtractor 创建提取器
// workers: 并发提取 goroutine 数；embedFn: 向量化函数
func NewExtractor(workers int, embedFn EmbedFunc) *Extractor {
	if workers <= 0 {
		workers = 2
	}
	return &Extractor{
		jobs:    make(chan ExtractJob, 100),
		embedFn: embedFn,
		workers: workers,
		stopCh:  make(chan struct{}),
	}
}

// Start 启动后台 Worker goroutine
func (e *Extractor) Start(ctx context.Context) {
	for i := 0; i < e.workers; i++ {
		e.wg.Add(1)
		go e.runWorker(ctx)
	}
	logger.Info("attachment extractor started", "workers", e.workers)
}

// Enqueue 投递提取任务（非阻塞）
func (e *Extractor) Enqueue(job ExtractJob) {
	select {
	case e.jobs <- job:
	default:
		logger.Warn("extractor queue full, dropping job",
			"attachment_id", job.AttachmentID)
	}
}

// Stop 优雅关闭
func (e *Extractor) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

func (e *Extractor) runWorker(ctx context.Context) {
	defer e.wg.Done()
	for {
		select {
		case <-e.stopCh:
			return
		case job, ok := <-e.jobs:
			if !ok {
				return
			}
			e.processJob(ctx, job)
		}
	}
}

// processJob 执行单个附件的完整提取流程
func (e *Extractor) processJob(ctx context.Context, job ExtractJob) {
	att, err := repository.GetAttachment(job.AttachmentID)
	if err != nil || att == nil {
		logger.Warn("attachment not found", "id", job.AttachmentID, "err", err)
		return
	}

	// 标记为 extracting
	_ = repository.UpdateAttachmentStatus(att.ID, "extracting", "", 0)

	// ── Step 3：解析文件内容 ─────────────────────────────────
	parser, err := NewParser(att.MimeType)
	if err != nil {
		logger.Warn("no parser for mime type",
			"mime", att.MimeType, "attachment_id", att.ID)
		_ = repository.UpdateAttachmentStatus(att.ID, "failed", err.Error(), 0)
		return
	}

	result, err := parser.Parse(ctx, att.FilePath)
	if err != nil {
		logger.Error("parse failed",
			"attachment_id", att.ID, "file", att.FileName, "err", err)
		_ = repository.UpdateAttachmentStatus(att.ID, "failed", err.Error(), 0)
		return
	}

	// 图片文件：base64 存入单个 memory（不分块）
	if result.IsImage {
		e.saveImageMemory(ctx, att, result)
		return
	}

	// 文本为空（扫描件 PDF 等）也记为 done，chunk_count=0
	if result.Text == "" {
		logger.Info("no text extracted, marking done",
			"attachment_id", att.ID, "file", att.FileName)
		_ = repository.UpdateAttachmentStatus(att.ID, "done", "", 0)
		return
	}

	// ── Step 4：分块 ─────────────────────────────────────────
	chunks := ChunkTextDefault(result.Text)

	// ── Step 5：向量化并存入 memories ─────────────────────────
	saved := 0
	for _, chunk := range chunks {
		if err := e.saveChunkMemory(ctx, att, chunk); err != nil {
			logger.Warn("failed to save chunk",
				"attachment_id", att.ID, "chunk_index", chunk.Index, "err", err)
			continue
		}
		saved++
	}

	_ = repository.UpdateAttachmentStatus(att.ID, "done", "", saved)
	logger.Info("attachment extracted",
		"attachment_id", att.ID,
		"file", att.FileName,
		"chunks", saved,
	)
}

// saveChunkMemory 向量化单个 Chunk 并存入 memories 表
func (e *Extractor) saveChunkMemory(ctx context.Context, att *repository.TaskAttachment, chunk Chunk) error {
	var embedding []float32

	if e.embedFn != nil {
		emb, err := e.embedFn(ctx, chunk.Content)
		if err != nil {
			logger.Warn("embedding failed, using zero vector",
				"attachment_id", att.ID, "err", err)
			embedding = make([]float32, 1536)
		} else {
			embedding = emb
		}
	} else {
		embedding = make([]float32, 1536) // embedFn 未配置时用零向量占位
	}

	mem := &repository.Memory{
		Content:    chunk.Content,
		Embedding:  pgvector.NewVector(embedding),
		MemoryType: "attachment",
	}
	if att.TaskID != "" {
		mem.TaskID.String = att.TaskID
		mem.TaskID.Valid = true
	}

	_, err := repository.SaveMemoryWithAttachment(mem, att.ID, chunk.Index)
	return err
}

// saveImageMemory 存储图片的 base64 作为 attachment 类型 memory
func (e *Extractor) saveImageMemory(ctx context.Context, att *repository.TaskAttachment, result *ParseResult) {
	// 图片不向量化，Content 存路径描述，ImageB64 通过 metadata 标注
	content := "[图片附件: " + att.FileName + "]"
	if result.ImageB64 != "" {
		// 存 base64（可能较大），截断前 200 字符作为摘要
		if len(result.ImageB64) > 200 {
			content = "[图片附件: " + att.FileName + "] base64_preview=" + result.ImageB64[:200]
		} else {
			content = "[图片附件: " + att.FileName + "] base64=" + result.ImageB64
		}
	}

	mem := &repository.Memory{
		Content:    content,
		Embedding:  pgvector.NewVector(make([]float32, 1536)),
		MemoryType: "attachment",
	}
	if att.TaskID != "" {
		mem.TaskID.String = att.TaskID
		mem.TaskID.Valid = true
	}

	_, err := repository.SaveMemoryWithAttachment(mem, att.ID, 0)
	if err != nil {
		logger.Warn("failed to save image memory",
			"attachment_id", att.ID, "err", err)
	}

	_ = repository.UpdateAttachmentStatus(att.ID, "done", "", 1)
	logger.Info("image attachment saved",
		"attachment_id", att.ID, "file", att.FileName)
}
