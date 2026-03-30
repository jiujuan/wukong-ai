package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// DuckDuckGoSearch DuckDuckGo 搜索工具
type DuckDuckGoSearch struct {
	client *http.Client
}

// NewDuckDuckGoSearch 创建 DuckDuckGo 搜索工具
func NewDuckDuckGoSearch() *DuckDuckGoSearch {
	return &DuckDuckGoSearch{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name 返回工具名称
func (d *DuckDuckGoSearch) Name() string {
	return "duckduckgo_search"
}

// Description 返回工具描述
func (d *DuckDuckGoSearch) Description() string {
	return "Search the web using DuckDuckGo for up-to-date information"
}

// Execute 执行搜索
func (d *DuckDuckGoSearch) Execute(ctx context.Context, input string) (string, error) {
	// 使用 DuckDuckGo HTML 搜索（不需要 API key）
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(input))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Wukong-AI/1.0)")

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 解析 HTML 结果
	results := d.parseHTMLResults(string(respBody))

	if len(results) == 0 {
		return "No search results found.", nil
	}

	// 格式化结果
	var output strings.Builder
	for i, r := range results {
		if i >= 10 { // 限制结果数量
			break
		}
		output.WriteString(fmt.Sprintf("## Result %d\n%s\n%s\n\n", i+1, r.Title, r.Snippet))
	}

	logger.Debug("DuckDuckGo search completed", "query", input, "results", len(results))
	return output.String(), nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Title   string
	Snippet string
	URL     string
}

// parseHTMLResults 解析 HTML 结果
func (d *DuckDuckGoSearch) parseHTMLResults(html string) []SearchResult {
	var results []SearchResult

	// 简单的正则匹配
	lines := strings.Split(html, "\n")
	var currentResult *SearchResult

	for _, line := range lines {
		if strings.Contains(line, `class="result__title"`) || strings.Contains(line, `class="result__a"`) {
			// 开始新结果
			if currentResult != nil {
				results = append(results, *currentResult)
			}
			currentResult = &SearchResult{}
		}

		if currentResult != nil {
			// 提取标题
			if strings.Contains(line, `class="result__a"`) {
				start := strings.Index(line, ">")
				end := strings.LastIndex(line, "<")
				if start >= 0 && end > start {
					currentResult.Title = strings.TrimSpace(line[start+1 : end])
				}
			}

			// 提取摘要
			if strings.Contains(line, `class="result__snippet"`) {
				start := strings.Index(line, ">")
				end := strings.LastIndex(line, "<")
				if start >= 0 && end > start {
					currentResult.Snippet = strings.TrimSpace(line[start+1 : end])
				}
			}
		}
	}

	if currentResult != nil {
		results = append(results, *currentResult)
	}

	// 如果解析失败，返回 JSON 格式尝试解析
	if len(results) == 0 {
		var jsonResults []struct {
			Text string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		}
		// 尝试找 JSON 数据
		start := strings.Index(html, `DDG.pageLayout.load('`)
		if start >= 0 {
			end := strings.Index(html[start:], `');`)
			if end > 0 {
				jsonStr := html[start+20 : start+end]
				if err := json.Unmarshal([]byte(jsonStr), &jsonResults); err == nil {
					for _, r := range jsonResults {
						if r.Text != "" {
							results = append(results, SearchResult{
								Title:   r.Text,
								Snippet: r.Text,
								URL:     r.FirstURL,
							})
						}
					}
				}
			}
		}
	}

	return results
}

// Ensure DuckDuckGoSearch 实现 Tool 接口
var _ tools.Tool = (*DuckDuckGoSearch)(nil)
