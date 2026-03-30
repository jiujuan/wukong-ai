package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// TavilySearch Tavily 搜索工具
type TavilySearch struct {
	apiKey string
	client *http.Client
}

// NewTavilySearch 创建 Tavily 搜索工具
func NewTavilySearch(apiKey string) *TavilySearch {
	return &TavilySearch{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回工具名称
func (t *TavilySearch) Name() string {
	return "tavily_search"
}

// Description 返回工具描述
func (t *TavilySearch) Description() string {
	return "Search the web using Tavily API for up-to-date information"
}

// Execute 执行搜索
func (t *TavilySearch) Execute(ctx context.Context, input string) (string, error) {
	if t.apiKey == "" {
		return "", fmt.Errorf("Tavily API key not configured")
	}

	reqBody := map[string]any{
		"query": input,
		"search_depth": "basic",
		"max_results": 5,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := "https://api.tavily.com/search"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("tv-api-key", t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Tavily API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Content     string `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 格式化结果
	var output string
	for _, r := range result.Results {
		output += fmt.Sprintf("## %s\n%s\nURL: %s\n\n", r.Title, r.Content, r.URL)
	}

	logger.Debug("Tavily search completed", "query", input, "results", len(result.Results))
	return output, nil
}

// Ensure TavilySearch 实现 Tool 接口
var _ tools.Tool = (*TavilySearch)(nil)
