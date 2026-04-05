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

// SupportsVision DeepSeek 支持 Vision（deepseek-vl 系列）
// 对于 deepseek-chat 等文本模型返回 false；vl 模型返回 true
func (d *DeepSeekLLM) SupportsVision() bool {
	return strings.Contains(strings.ToLower(d.model), "vl")
}

// ChatWithImages DeepSeek Vision 调用（格式与 OpenAI 兼容）
func (d *DeepSeekLLM) ChatWithImages(ctx context.Context, prompt string, images []string) (string, error) {
	if !d.SupportsVision() {
		// 非 Vision 模型：降级为纯文本
		return d.Chat(ctx, prompt)
	}

	content := []map[string]any{{"type": "text", "text": prompt}}
	for _, b64 := range images {
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": "data:image/jpeg;base64," + b64},
		})
	}

	reqBody := map[string]any{
		"model":    d.model,
		"messages": []map[string]any{{"role": "user", "content": content}},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("DeepSeek ChatWithImages marshal: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", d.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("DeepSeek ChatWithImages request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("DeepSeek ChatWithImages send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek Vision error: status=%d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct{ Content string `json:"content"` } `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || len(result.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek Vision parse error")
	}
	return result.Choices[0].Message.Content, nil
}
