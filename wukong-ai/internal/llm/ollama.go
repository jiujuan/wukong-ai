package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/llmstream"
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
	llm := &OllamaLLM{
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 300 * time.Second, // Ollama 可能需要更长时间
		},
	}
	logger.Info("ollama llm initialized", "base_url", llm.baseURL, "model", llm.model)
	return llm
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
	logger.Info("ollama chat start", "model", o.model, "message_count", len(messages))
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
		logger.Error("ollama chat marshal request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("ollama chat create request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Error("ollama chat http call failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ollama chat read response failed", "model", o.model, "status", resp.StatusCode, "err", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("ollama chat api error", "model", o.model, "status", resp.StatusCode, "response_length", len(respBody))
		return "", formatOllamaAPIError("Ollama API error", resp.StatusCode, respBody)
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error("ollama chat unmarshal response failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.Debug("Ollama chat completed", "model", o.model)
	return result.Message.Content, nil
}

func (o *OllamaLLM) ChatWithHistoryStream(ctx context.Context, messages []Message, onChunk func(chunk string) error) error {
	logger.Info("ollama stream chat start", "model", o.model, "message_count", len(messages))
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
		logger.Error("ollama stream marshal request failed", "model", o.model, "err", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("ollama stream create request failed", "model", o.model, "err", err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if err := llmstream.Stream(o.client, req, onChunk, llmstream.ParseOllamaChunk); err != nil {
		logger.Error("ollama stream failed", "model", o.model, "err", err)
		return err
	}

	logger.Info("ollama stream chat completed", "model", o.model)
	return nil
}

// Embed 生成文本向量 (Ollama 支持 embedding)
func (o *OllamaLLM) Embed(ctx context.Context, text string) ([]float32, error) {
	logger.Info("ollama embedding start", "model", o.model, "text_length", len(text))
	reqBody := map[string]any{
		"model": o.model,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("ollama embedding marshal request failed", "model", o.model, "err", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/embeddings", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("ollama embedding create request failed", "model", o.model, "err", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Error("ollama embedding http call failed", "model", o.model, "err", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ollama embedding read response failed", "model", o.model, "status", resp.StatusCode, "err", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("ollama embedding api error", "model", o.model, "status", resp.StatusCode, "response_length", len(respBody))
		return nil, formatOllamaAPIError("Ollama Embedding API error", resp.StatusCode, respBody)
	}

	var result struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error("ollama embedding unmarshal response failed", "model", o.model, "err", err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	logger.Info("ollama embedding completed", "model", o.model, "dimension", len(result.Embedding))
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

// SupportsVision Ollama 根据模型名判断是否支持 Vision
// 已知支持 Vision 的模型：llava, llava-llama3, bakllava, moondream, minicpm-v
func (o *OllamaLLM) SupportsVision() bool {
	modelLower := strings.ToLower(o.model)
	visionModels := []string{"llava", "bakllava", "moondream", "minicpm-v", "cogvlm"}
	for _, vm := range visionModels {
		if strings.Contains(modelLower, vm) {
			return true
		}
	}
	return false
}

// ChatWithImages Ollama Vision 调用（使用 /api/generate 接口的 images 字段）
func (o *OllamaLLM) ChatWithImages(ctx context.Context, prompt string, images []string) (string, error) {
	logger.Info("ollama vision chat start", "model", o.model, "prompt_length", len(prompt), "image_count", len(images))
	if !o.SupportsVision() {
		logger.Warn("ollama vision unsupported model, fallback to text chat", "model", o.model)
		return o.Chat(ctx, prompt)
	}

	reqBody := map[string]any{
		"model":  o.model,
		"prompt": prompt,
		"images": images, // Ollama 原生支持 base64 images 字段
		"stream": false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("ollama vision marshal request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("OllamaLLM ChatWithImages marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("ollama vision create request failed", "model", o.model, "err", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Error("ollama vision http call failed", "model", o.model, "err", err)
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Error("ollama vision api error", "model", o.model, "status", resp.StatusCode, "response_length", len(respBody))
		return "", fmt.Errorf("Ollama Vision error: status=%d", resp.StatusCode)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error("ollama vision parse response failed", "model", o.model, "err", err)
		return "", fmt.Errorf("Ollama Vision parse error: %w", err)
	}
	logger.Info("ollama vision chat completed", "model", o.model, "content_length", len(result.Response))
	return result.Response, nil
}
