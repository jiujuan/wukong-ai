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

// OllamaLLM Ollama LLM 实现 (本地模型)
type OllamaLLM struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaLLM 创建 Ollama LLM 实例
func NewOllamaLLM(cfg *config.LLMConfig) *OllamaLLM {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/v1")
	model := cfg.Model
	if model == "" {
		model = "llama2"
	}
	return &OllamaLLM{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 300 * time.Second, // Ollama 可能需要更长时间
		},
	}
}

// Name 返回提供者名称
func (o *OllamaLLM) Name() string {
	return "ollama"
}

// Chat 发送对话请求
func (o *OllamaLLM) Chat(ctx context.Context, prompt string) (string, error) {
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	return o.ChatWithHistory(ctx, messages)
}

// ChatWithHistory 发送带历史的对话请求
func (o *OllamaLLM) ChatWithHistory(ctx context.Context, messages []Message) (string, error) {
	reqBody := map[string]any{
		"model":  o.model,
		"stream": false,
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
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return "", formatOllamaAPIError("Ollama API error", resp.StatusCode, respBody)
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.Debug("Ollama chat completed", "model", o.model)
	return result.Message.Content, nil
}

func (o *OllamaLLM) ChatWithHistoryStream(ctx context.Context, messages []Message, onChunk func(chunk string) error) error {
	reqBody := map[string]any{
		"model":  o.model,
		"stream": true,
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
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("ollama stream api error: status=%d", resp.StatusCode)
		}
		return formatOllamaAPIError("Ollama API error", resp.StatusCode, respBody)
	}

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var chunk struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			return fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		if chunk.Message.Content != "" && onChunk != nil {
			if err := onChunk(chunk.Message.Content); err != nil {
				return err
			}
		}

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream read error: %w", err)
	}

	return nil
}

// Embed 生成文本向量 (Ollama 支持 embedding)
func (o *OllamaLLM) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]any{
		"model": o.model,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return nil, formatOllamaAPIError("Ollama Embedding API error", resp.StatusCode, respBody)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Embedding, nil
}

func formatOllamaAPIError(prefix string, statusCode int, body []byte) error {
	var parsed struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error != "" {
		return fmt.Errorf("%s: status=%d, error=%s", prefix, statusCode, parsed.Error)
	}
	return fmt.Errorf("%s: status=%d, body=%s", prefix, statusCode, string(body))
}
