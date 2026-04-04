package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/llm"
)

// HealthHandler 健康检查处理器（含 LLM 熔断器状态）
type HealthHandler struct {
	// llmProvider 若是 FallbackChain，可提供详细熔断状态；否则返回基础 ok
	llmProvider llm.LLM
}

func NewHealthHandler(llmProvider llm.LLM) *HealthHandler {
	return &HealthHandler{llmProvider: llmProvider}
}

// Handle GET /health  基础健康检查
func (h *HealthHandler) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "v1.0",
	})
}

// HandleLLM GET /health/llm  LLM 熔断器 + 调用统计详情
func (h *HealthHandler) HandleLLM(c *gin.Context) {
	// 若是 FallbackChain 则返回详细状态
	if chain, ok := h.llmProvider.(*llm.FallbackChain); ok {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"providers": chain.HealthStatus(),
		})
		return
	}

	// 普通单 provider，返回简化响应
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"provider": h.llmProvider.Name(),
		"mode":     "single",
	})
}
