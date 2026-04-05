package parser

import "strings"

const (
	DefaultChunkSize    = 512  // 每块目标 Token 数（粗估：1 Token ≈ 1 词）
	DefaultChunkOverlap = 64   // 相邻块重叠 Token 数
	SmallFileThreshold  = 2000 // 小于此 Token 数不分块，直接全文注入
)

// Chunk 文本分块
type Chunk struct {
	Index   int    // 块序号（从 0 开始）
	Content string // 块内容
	Tokens  int    // 粗估 Token 数
}

// ChunkText 按词数分块（空格分词，适用中英文混合）
// - 小文件（词数 ≤ SmallFileThreshold）：返回单个 Chunk
// - 大文件：滑动窗口分块，相邻块有 overlap 个词重叠
func ChunkText(text string, chunkSize, overlap int) []Chunk {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if overlap < 0 {
		overlap = DefaultChunkOverlap
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 4 // 防止 overlap >= size 导致死循环
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	// 小文件：不分块
	if len(words) <= SmallFileThreshold {
		return []Chunk{{Index: 0, Content: text, Tokens: len(words)}}
	}

	var chunks []Chunk
	step := chunkSize - overlap
	if step <= 0 {
		step = 1
	}

	for i := 0; i < len(words); i += step {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, Chunk{
			Index:   len(chunks),
			Content: strings.Join(words[i:end], " "),
			Tokens:  end - i,
		})
		if end == len(words) {
			break
		}
	}
	return chunks
}

// ChunkTextDefault 使用默认参数分块
func ChunkTextDefault(text string) []Chunk {
	return ChunkText(text, DefaultChunkSize, DefaultChunkOverlap)
}
