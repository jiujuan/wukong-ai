package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// OpenAILLM OpenAI LLM 实现
type OpenAILLM struct {
	apiKey         string
	baseURL        string
	model          string
	embeddingModel string
	embeddingDim   int
	client         *http.Client
}

// NewOpenAILLM 创建 OpenAI LLM 实例
func NewOpenAILLM(cfg *config.LLMConfig) *OpenAILLM {
	return &OpenAILLM{
		apiKey:         cfg.APIKey,
		baseURL:        cfg.BaseURL,
		model:          cfg.Model,
		embeddingModel: cfg.EmbeddingModel,
		embeddingDim:   cfg.EmbeddingDim,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name 返回提供者名称
func (o *OpenAILLM) Name() string {
	return "openai"
}

// Chat 发送对话请求
func (o *OpenAILLM) Chat(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return o.ChatWithHistory(ctx, messages)
}

// ChatWithHistory 发送带历史的对话请求
func (o *OpenAILLM) ChatWithHistory(ctx context.Context, messages []Message) (string, error) {
	reqBody := map[string]any{
		"model": o.model,
		"messages": func() []map[string]string {
			result := make([]map[string]string, len(messages))
			for i, m := range messages {
				result[i] = map[string]string{
					"role":    m.Role,
					"content": m.Content,
				}
			}
			return result
		}(),
		"temperature": 0.7,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	logger.Debug("OpenAI chat completed", "model", o.model, "content_length", len(result.Choices[0].Message.Content))
	return result.Choices[0].Message.Content, nil
}

// Embed 生成文本向量
func (o *OpenAILLM) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]any{
		"model": o.embeddingModel,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI Embedding API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned from OpenAI")
	}

	return result.Data[0].Embedding, nil
}
