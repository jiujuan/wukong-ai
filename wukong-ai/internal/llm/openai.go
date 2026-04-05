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
	"github.com/jiujuan/wukong-ai/pkg/llmstream"
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
	logger.Info("openai chat start", "model", o.model, "message_count", len(messages))
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
		logger.Error("openai chat marshal request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("openai chat create request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Error("openai chat http call failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("openai chat read response failed", "model", o.model, "status", resp.StatusCode, "err", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("openai chat api error", "model", o.model, "status", resp.StatusCode, "response_length", len(respBody))
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
		logger.Error("openai chat unmarshal response failed", "model", o.model, "err", err)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		logger.Error("openai chat empty choices", "model", o.model)
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	logger.Debug("OpenAI chat completed", "model", o.model, "content_length", len(result.Choices[0].Message.Content))
	return result.Choices[0].Message.Content, nil
}

func (o *OpenAILLM) ChatWithHistoryStream(ctx context.Context, messages []Message, onChunk func(chunk string) error) error {
	logger.Info("openai stream chat start", "model", o.model, "message_count", len(messages))
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
		"stream":      true,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("openai stream marshal request failed", "model", o.model, "err", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("openai stream create request failed", "model", o.model, "err", err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	if err := llmstream.Stream(o.client, req, onChunk, llmstream.ParseOpenAICompatibleChunk); err != nil {
		logger.Error("openai stream failed", "model", o.model, "err", err)
		return err
	}

	logger.Info("openai stream chat completed", "model", o.model)
	return nil
}

// Embed 生成文本向量
func (o *OpenAILLM) Embed(ctx context.Context, text string) ([]float32, error) {
	logger.Info("openai embedding start", "model", o.embeddingModel, "text_length", len(text))
	reqBody := map[string]any{
		"model": o.embeddingModel,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("openai embedding marshal request failed", "model", o.embeddingModel, "err", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("openai embedding create request failed", "model", o.embeddingModel, "err", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Error("openai embedding http call failed", "model", o.embeddingModel, "err", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("openai embedding read response failed", "model", o.embeddingModel, "status", resp.StatusCode, "err", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("openai embedding api error", "model", o.embeddingModel, "status", resp.StatusCode, "response_length", len(respBody))
		return nil, fmt.Errorf("OpenAI Embedding API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error("openai embedding unmarshal response failed", "model", o.embeddingModel, "err", err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Data) == 0 {
		logger.Error("openai embedding empty data", "model", o.embeddingModel)
		return nil, fmt.Errorf("no embeddings returned from OpenAI")
	}

	logger.Info("openai embedding completed", "model", o.embeddingModel, "dimension", len(result.Data[0].Embedding))
	return result.Data[0].Embedding, nil
}

// SupportsVision OpenAI 支持 Vision
func (o *OpenAILLM) SupportsVision() bool {
	return true
}

// ChatWithImages 发送携带图片的请求（OpenAI Vision API）
func (o *OpenAILLM) ChatWithImages(ctx context.Context, prompt string, images []string) (string, error) {
	logger.Info("openai vision chat start", "model", o.model, "prompt_length", len(prompt), "image_count", len(images))
	// 构造 multimodal content
	content := []map[string]any{
		{"type": "text", "text": prompt},
	}
	for _, b64 := range images {
		// 自动推断 MIME（简化处理，默认 jpeg）
		mime := "image/jpeg"
		if len(b64) > 10 {
			switch b64[:8] {
			case "iVBORw0K":
				mime = "image/png"
			case "R0lGODlh", "R0lGODdh":
				mime = "image/gif"
			}
		}
		dataURI := fmt.Sprintf("data:%s;base64,%s", mime, b64)
		content = append(content, map[string]any{
			"type":      "image_url",
			"image_url": map[string]string{"url": dataURI},
		})
	}

	reqBody := map[string]any{
		"model":       o.model,
		"messages":    []map[string]any{{"role": "user", "content": content}},
		"temperature": 0.7,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		logger.Error("openai vision marshal request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("ChatWithImages marshal: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		logger.Error("openai vision create request failed", "model", o.model, "err", err)
		return "", fmt.Errorf("ChatWithImages request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))

	resp, err := o.client.Do(req)
	if err != nil {
		logger.Error("openai vision http call failed", "model", o.model, "err", err)
		return "", fmt.Errorf("ChatWithImages send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logger.Error("openai vision api error", "model", o.model, "status", resp.StatusCode, "response_length", len(respBody))
		return "", fmt.Errorf("OpenAI Vision error: status=%d body=%s", resp.StatusCode, respBody)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		logger.Error("openai vision unmarshal response failed", "model", o.model, "err", err)
		return "", fmt.Errorf("ChatWithImages unmarshal: %w", err)
	}
	if len(result.Choices) == 0 {
		logger.Error("openai vision empty choices", "model", o.model)
		return "", fmt.Errorf("no choices from Vision API")
	}
	logger.Info("openai vision chat completed", "model", o.model, "content_length", len(result.Choices[0].Message.Content))
	return result.Choices[0].Message.Content, nil
}
