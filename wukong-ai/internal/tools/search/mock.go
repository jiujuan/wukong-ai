package search

import (
	"context"
	"fmt"

	"github.com/jiujuan/wukong-ai/internal/tools"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// MockSearch Mock 搜索工具（用于测试）
type MockSearch struct{}

// NewMockSearch 创建 Mock 搜索工具
func NewMockSearch() *MockSearch {
	return &MockSearch{}
}

// Name 返回工具名称
func (m *MockSearch) Name() string {
	return "mock_search"
}

// Description 返回工具描述
func (m *MockSearch) Description() string {
	return "Mock search tool for testing purposes"
}

// Execute 执行搜索（返回模拟数据）
func (m *MockSearch) Execute(ctx context.Context, input string) (string, error) {
	logger.Info("Mock search executing", "query", input)

	result := fmt.Sprintf(`## Mock Search Results for: %s

### Result 1: Mock Document Title
This is a mock search result for testing purposes. The actual search functionality
is not available in this mock implementation.

Source: mock://example.com/1

### Result 2: Another Mock Document
Additional mock content for testing the search integration.

Source: mock://example.com/2

### Result 3: Third Mock Result
More mock data to simulate search results.

Source: mock://example.com/3

---
Note: This is a mock search tool. Configure Tavily or DuckDuckGo for real search functionality.`, input)

	return result, nil
}

// Ensure MockSearch 实现 Tool 接口
var _ tools.Tool = (*MockSearch)(nil)
