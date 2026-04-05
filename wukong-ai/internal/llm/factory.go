package llm

import (
	"fmt"

	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Factory LLM 工厂
type Factory struct{}

// NewFactory 创建 LLM 工厂
func NewFactory() *Factory {
	return &Factory{}
}

// CreateLLM 根据单个 LLMConfig 创建 LLM 实例（向前兼容）
func (f *Factory) CreateLLM(cfg *config.LLMConfig) (LLM, error) {
	logger.Info("create llm requested", "provider", cfg.Provider, "model", cfg.Model)
	switch cfg.Provider {
	case "openai":
		llm := NewOpenAILLM(cfg)
		logger.Info("create llm success", "provider", cfg.Provider, "model", cfg.Model)
		return llm, nil
	case "deepseek":
		llm := NewDeepSeekLLM(cfg)
		logger.Info("create llm success", "provider", cfg.Provider, "model", cfg.Model)
		return llm, nil
	case "ollama":
		llm := NewOllamaLLM(cfg)
		logger.Info("create llm success", "provider", cfg.Provider, "model", cfg.Model)
		return llm, nil
	default:
		logger.Error("create llm failed: unsupported provider", "provider", cfg.Provider)
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

// BuildFallbackChain 根据完整 LLMConfig 构建带熔断+限流+降级的 FallbackChain。
//
// 构建逻辑：
//  1. 主模型（cfg.Provider）作为降级链首位
//  2. cfg.Fallbacks 中的模型依次追加
//  3. 每个 provider 各自持有独立的 CircuitBreaker 和 RateLimiter
//
// 若 cfg.Fallbacks 为空，降级链只含主模型，行为等价于单 provider。
// 调用方无需感知降级链存在，直接把返回值当 LLM 使用。
func (f *Factory) BuildFallbackChain(cfg *config.LLMConfig) (*FallbackChain, error) {
	// 熔断参数
	cbThreshold := cfg.CircuitBreaker.Threshold
	if cbThreshold <= 0 {
		cbThreshold = config.DefaultCircuitBreakerThreshold
	}
	cbTimeout := cfg.CircuitBreaker.Timeout
	if cbTimeout <= 0 {
		cbTimeout = config.DefaultCircuitBreakerTimeout
	}

	var entries []*ProviderEntry

	// ── 主模型 ────────────────────────────────────────────────
	primaryProvider, err := f.CreateLLM(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary LLM (%s): %w", cfg.Provider, err)
	}
	entries = append(entries, &ProviderEntry{
		provider:    primaryProvider,
		breaker:     NewCircuitBreaker(cfg.Provider, cbThreshold, cbTimeout),
		rateLimiter: NewRateLimiter(cfg.Provider, cfg.MaxRPM, cfg.MaxTPM),
	})
	logger.Info("fallback chain: primary provider added",
		"provider", cfg.Provider,
		"model", cfg.Model,
		"max_rpm", cfg.MaxRPM,
		"max_tpm", cfg.MaxTPM,
		"cb_threshold", cbThreshold,
		"cb_timeout_s", cbTimeout.Seconds(),
	)

	// ── 降级模型 ──────────────────────────────────────────────
	for i, fb := range cfg.Fallbacks {
		fallbackCfg := &config.LLMConfig{
			Provider: fb.Provider,
			APIKey:   fb.APIKey,
			BaseURL:  fb.BaseURL,
			Model:    fb.Model,
		}
		fbProvider, err := f.CreateLLM(fallbackCfg)
		if err != nil {
			logger.Warn("fallback chain: failed to create fallback provider, skipping",
				"index", i,
				"provider", fb.Provider,
				"err", err,
			)
			continue
		}
		entries = append(entries, &ProviderEntry{
			provider:    fbProvider,
			breaker:     NewCircuitBreaker(fb.Provider, cbThreshold, cbTimeout),
			rateLimiter: NewRateLimiter(fb.Provider, fb.MaxRPM, fb.MaxTPM),
		})
		logger.Info("fallback chain: fallback provider added",
			"index", i+1,
			"provider", fb.Provider,
			"model", fb.Model,
		)
	}

	logger.Info("fallback chain built", "total_providers", len(entries))
	return NewFallbackChain(entries), nil
}

// CreateEmbeddingLLM 创建专用于 embedding 的 LLM
func (f *Factory) CreateEmbeddingLLM(cfg *config.LLMConfig) (LLM, error) {
	logger.Info("create embedding llm requested", "provider", "openai", "embedding_model", cfg.EmbeddingModel, "embedding_dim", cfg.EmbeddingDim)
	openaiCfg := &config.LLMConfig{
		APIKey:         cfg.APIKey,
		BaseURL:        "https://api.openai.com/v1",
		EmbeddingModel: cfg.EmbeddingModel,
		EmbeddingDim:   cfg.EmbeddingDim,
	}
	llm := NewOpenAILLM(openaiCfg)
	logger.Info("create embedding llm success", "provider", "openai", "embedding_model", cfg.EmbeddingModel)
	return llm, nil
}
