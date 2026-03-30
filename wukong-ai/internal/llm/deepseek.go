package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// DeepSeekLLM DeepSeek LLM 实现
type DeepSeekLLM struct {
	apiKey         string
	baseURL        string
	model          string
	embeddingModel string
	embeddingDim   int
	client         *http.Client
}

// NewDeepSeekLLM 创建 DeepSeek LLM 实例
func NewDeepSeekLLM(cfg *config.LLMConfig) *DeepSeekLLM {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" || strings.EqualFold(model, "deepseek") {
		model = "deepseek-chat"
	}
	return &DeepSeekLLM{
		apiKey:         cfg.APIKey,
		baseURL:        baseURL,
		model:          model,
		embeddingModel: cfg.EmbeddingModel,
		embeddingDim:   cfg.EmbeddingDim,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name 返回提供者名称
func (d *DeepSeekLLM) Name() string {
	return "deepseek"
}

// Chat 发送对话请求
func (d *DeepSeekLLM) Chat(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return d.ChatWithHistory(ctx, messages)
}

// ChatWithHistory 发送带历史的对话请求
func (d *DeepSeekLLM) ChatWithHistory(ctx context.Context, messages []Message) (string, error) {
	reqBody := map[string]any{
		"model": d.model,
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

	url := fmt.Sprintf("%s/chat/completions", d.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API error: status=%d, model=%s, body=%s", resp.StatusCode, d.model, string(respBody))
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
		return "", fmt.Errorf("no choices returned from DeepSeek")
	}

	logger.Debug("DeepSeek chat completed", "model", d.model)
	return result.Choices[0].Message.Content, nil
}

func (d *DeepSeekLLM) ChatWithHistoryStream(ctx context.Context, messages []Message, onChunk func(chunk string) error) error {
	reqBody := map[string]any{
		"model": d.model,
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
		"stream":      true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", d.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("DeepSeek API error: status=%d, model=%s", resp.StatusCode, d.model)
		}
		return fmt.Errorf("DeepSeek API error: status=%d, model=%s, body=%s", resp.StatusCode, d.model, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data:") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		if line == "" || line == "[DONE]" {
			continue
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("failed to parse deepseek stream chunk: %w", err)
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		content := chunk.Choices[0].Delta.Content
		if content == "" {
			continue
		}
		if onChunk != nil {
			if err := onChunk(content); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("deepseek stream read error: %w", err)
	}
	return nil
}

// Embed 生成文本向量 (DeepSeek 可能不支持 embedding API，使用备用方案)
func (d *DeepSeekLLM) Embed(ctx context.Context, text string) ([]float32, error) {
	// DeepSeek 目前没有公开的 embedding API，返回一个占位实现
	// 实际使用时可以通过 OpenAI API 或其他服务生成
	embedding := make([]float32, d.embeddingDim)
	for i := range embedding {
		embedding[i] = 0.0
	}
	logger.Warn("DeepSeek embedding not supported, returning placeholder")
	return embedding, nil
}
