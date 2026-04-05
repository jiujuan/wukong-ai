package llm

import (
	"context"
)

// LLM LLM 统一接口
type LLM interface {
	// Name 返回 LLM 提供者名称
	Name() string
	// Chat 发送对话请求并返回响应
	Chat(ctx context.Context, prompt string) (string, error)
	// ChatWithHistory 发送带历史的对话请求
	ChatWithHistory(ctx context.Context, messages []Message) (string, error)
	// Embed 生成文本向量
	Embed(ctx context.Context, text string) ([]float32, error)
	// SupportsVision 是否支持图片理解（Vision）
	SupportsVision() bool
	// ChatWithImages 发送携带图片的消息（Vision 专用）
	// images: base64 编码的图片列表
	ChatWithImages(ctx context.Context, prompt string, images []string) (string, error)
}

// StreamLLM 流式输出扩展接口
type StreamLLM interface {
	ChatWithHistoryStream(ctx context.Context, messages []Message, onChunk func(chunk string) error) error
}

// Message 对话消息
type Message struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage,omitempty"`
}

// Usage 使用量统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

// EmbedResponse 向量响应
type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

// StreamingChunk 流式响应块
type StreamingChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}
