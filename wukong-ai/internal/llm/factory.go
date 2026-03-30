package llm

import (
	"fmt"

	"github.com/jiujuan/wukong-ai/pkg/config"
)

// Factory LLM 工厂
type Factory struct{}

// NewFactory 创建 LLM 工厂
func NewFactory() *Factory {
	return &Factory{}
}

// CreateLLM 根据配置创建 LLM 实例
func (f *Factory) CreateLLM(cfg *config.LLMConfig) (LLM, error) {
	switch cfg.Provider {
	case "openai":
		return NewOpenAILLM(cfg), nil
	case "deepseek":
		return NewDeepSeekLLM(cfg), nil
	case "ollama":
		return NewOllamaLLM(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

// CreateEmbeddingLLM 创建专用于 embedding 的 LLM（使用 OpenAI 作为默认）
func (f *Factory) CreateEmbeddingLLM(cfg *config.LLMConfig) (LLM, error) {
	// 始终使用 OpenAI 进行 embedding
	openaiCfg := &config.LLMConfig{
		APIKey:         cfg.APIKey,
		BaseURL:        "https://api.openai.com/v1",
		EmbeddingModel: cfg.EmbeddingModel,
		EmbeddingDim:   cfg.EmbeddingDim,
	}
	return NewOpenAILLM(openaiCfg), nil
}
