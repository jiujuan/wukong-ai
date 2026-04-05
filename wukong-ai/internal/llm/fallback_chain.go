package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// ProviderEntry 降级链中的单个提供者条目
type ProviderEntry struct {
	provider    LLM
	breaker     *CircuitBreaker
	rateLimiter *LLMRateLimiter
}

// FallbackChain 多模型降级链，实现 LLM 接口，对上层完全透明。
//
// 调用顺序：
//  1. 按 providers 顺序遍历
//  2. 跳过熔断中的 provider（breaker.Allow() == false）
//  3. 等待限流（rateLimiter.Wait()）
//  4. 发起调用；成功 → 记录成功，返回；失败 → 记录失败，尝试下一个
//  5. 全部失败 → 返回 ErrAllProvidersFailed
type FallbackChain struct {
	entries     []*ProviderEntry
	callMetrics *CallMetrics // 可选：调用统计
}

// NewFallbackChain 创建降级链
func NewFallbackChain(entries []*ProviderEntry) *FallbackChain {
	return &FallbackChain{
		entries:     entries,
		callMetrics: NewCallMetrics(),
	}
}

// Name 返回当前活跃 provider 名称（用于日志）
func (c *FallbackChain) Name() string {
	for _, e := range c.entries {
		if e.breaker.Allow() {
			return fmt.Sprintf("fallback_chain[primary=%s]", e.provider.Name())
		}
	}
	return "fallback_chain[all_open]"
}

// Chat 实现 LLM.Chat，自动熔断 + 限流 + 降级
func (c *FallbackChain) Chat(ctx context.Context, prompt string) (string, error) {
	msgs := []Message{{Role: "user", Content: prompt}}
	return c.ChatWithHistory(ctx, msgs)
}

// ChatWithHistory 实现 LLM.ChatWithHistory，自动熔断 + 限流 + 降级
func (c *FallbackChain) ChatWithHistory(ctx context.Context, messages []Message) (string, error) {
	var lastErr error
	openSkipped := 0
	logger.Info("fallback chain chat start", "provider_count", len(c.entries), "message_count", len(messages))

	for i, entry := range c.entries {
		providerName := entry.provider.Name()

		// ── 1. 熔断检查 ────────────────────────────────────────
		if !entry.breaker.Allow() {
			logger.Warn("circuit breaker open, skipping provider",
				"provider", providerName,
				"state", entry.breaker.GetState().String(),
			)
			c.callMetrics.Record(providerName, "skipped", 0)
			openSkipped++
			continue
		}

		// ── 2. 限流等待 ────────────────────────────────────────
		// 简单估算 token 数：字符总数 / 4（粗略），最少估 100
		estimatedTokens := estimateTokens(messages)
		logger.Debug("fallback chain estimated tokens", "provider", providerName, "estimated_tokens", estimatedTokens)
		if err := entry.rateLimiter.Wait(ctx, estimatedTokens); err != nil {
			logger.Warn("rate limiter wait failed", "provider", providerName, "err", err)
			lastErr = err
			continue
		}

		// ── 3. 发起调用 ────────────────────────────────────────
		start := time.Now()
		result, err := entry.provider.ChatWithHistory(ctx, messages)
		elapsed := time.Since(start)

		if err != nil {
			entry.breaker.RecordFailure()
			c.callMetrics.Record(providerName, "failed", elapsed)
			logger.Warn("provider call failed, trying next",
				"provider", providerName,
				"err", err,
				"elapsed_ms", elapsed.Milliseconds(),
				"next_available", i+1 < len(c.entries),
			)
			lastErr = err
			continue
		}

		// ── 4. 调用成功 ────────────────────────────────────────
		entry.breaker.RecordSuccess()
		c.callMetrics.Record(providerName, "success", elapsed)

		if i > 0 {
			// 使用的不是首选 provider，记录降级日志
			logger.Info("fallback provider used",
				"provider", providerName,
				"fallback_index", i,
				"elapsed_ms", elapsed.Milliseconds(),
			)
		} else {
			logger.Debug("primary provider call success",
				"provider", providerName,
				"elapsed_ms", elapsed.Milliseconds(),
			)
		}
		return result, nil
	}

	// 所有 provider 都因熔断被跳过时，强制探测首选 provider 一次，避免整个任务链路完全无法前进
	if lastErr == nil && openSkipped == len(c.entries) && len(c.entries) > 0 {
		entry := c.entries[0]
		providerName := entry.provider.Name()
		logger.Warn("all providers skipped by breaker, force probing primary provider", "provider", providerName)

		estimatedTokens := estimateTokens(messages)
		if err := entry.rateLimiter.Wait(ctx, estimatedTokens); err != nil {
			lastErr = err
		} else {
			start := time.Now()
			result, err := entry.provider.ChatWithHistory(ctx, messages)
			elapsed := time.Since(start)
			if err != nil {
				entry.breaker.RecordFailure()
				c.callMetrics.Record(providerName, "failed", elapsed)
				lastErr = err
			} else {
				entry.breaker.RecordSuccess()
				c.callMetrics.Record(providerName, "success", elapsed)
				logger.Info("force probe provider succeeded", "provider", providerName, "elapsed_ms", elapsed.Milliseconds())
				return result, nil
			}
		}
	}

	// 所有 provider 均失败或熔断
	logger.Error(
		"fallback chain chat all providers failed",
		"provider_count", len(c.entries),
		"open_skipped", openSkipped,
		"last_err", wrapLastErr(lastErr),
	)
	return "", fmt.Errorf("all LLM providers failed or circuit-opened: last_err=%w",
		wrapLastErr(lastErr))
}

// ChatWithHistoryStream 流式调用，仅尝试第一个未熔断的 provider
// （降级链的流式支持：不跨 provider 降级，保持流式语义）
func (c *FallbackChain) ChatWithHistoryStream(
	ctx context.Context,
	messages []Message,
	onChunk func(chunk string) error,
) error {
	logger.Info("fallback chain stream start", "provider_count", len(c.entries), "message_count", len(messages))
	for i, entry := range c.entries {
		providerName := entry.provider.Name()

		if !entry.breaker.Allow() {
			logger.Warn("stream: circuit open, skipping", "provider", providerName)
			continue
		}

		if err := entry.rateLimiter.Wait(ctx, estimateTokens(messages)); err != nil {
			logger.Warn("stream rate limiter wait failed", "provider", providerName, "err", err)
			continue
		}

		streamProvider, ok := entry.provider.(StreamLLM)
		if !ok {
			// 该 provider 不支持流式，尝试下一个
			logger.Debug("stream: provider does not support streaming, skipping",
				"provider", providerName)
			continue
		}

		start := time.Now()
		err := streamProvider.ChatWithHistoryStream(ctx, messages, onChunk)
		elapsed := time.Since(start)

		if err != nil {
			entry.breaker.RecordFailure()
			c.callMetrics.Record(providerName, "stream_failed", elapsed)
			logger.Warn("stream provider failed, trying next",
				"provider", providerName, "err", err)
			continue
		}

		entry.breaker.RecordSuccess()
		c.callMetrics.Record(providerName, "stream_success", elapsed)
		if i > 0 {
			logger.Info("stream fallback provider used",
				"provider", providerName, "fallback_index", i)
		}
		return nil
	}
	logger.Error("fallback chain stream all providers failed", "provider_count", len(c.entries))
	return fmt.Errorf("all stream providers failed or circuit-opened")
}

// Embed 向量化（仅用主 provider，不跨 provider 降级）
func (c *FallbackChain) Embed(ctx context.Context, text string) ([]float32, error) {
	logger.Info("fallback chain embedding start", "provider_count", len(c.entries), "text_length", len(text))
	for _, entry := range c.entries {
		providerName := entry.provider.Name()
		if !entry.breaker.Allow() {
			logger.Warn("embedding breaker open, skipping provider", "provider", providerName)
			continue
		}
		result, err := entry.provider.Embed(ctx, text)
		if err != nil {
			entry.breaker.RecordFailure()
			logger.Warn("embedding provider failed", "provider", providerName, "err", err)
			continue
		}
		entry.breaker.RecordSuccess()
		logger.Info("embedding provider success", "provider", providerName, "dimension", len(result))
		return result, nil
	}
	logger.Error("fallback chain embedding all providers failed", "provider_count", len(c.entries))
	return nil, fmt.Errorf("all providers failed for embedding")
}

// HealthStatus 返回所有 provider 的熔断器状态（用于健康检查 API）
func (c *FallbackChain) HealthStatus() []map[string]any {
	status := make([]map[string]any, 0, len(c.entries))
	for _, e := range c.entries {
		stat := e.breaker.Stats()
		stat["metrics"] = c.callMetrics.Get(e.provider.Name())
		status = append(status, stat)
	}
	return status
}

// ── helpers ──────────────────────────────────────────────────────────────────

// estimateTokens 粗略估算消息 Token 数（字符总数 / 4，最小 100）
func estimateTokens(messages []Message) int {
	total := 0
	for _, m := range messages {
		total += len([]rune(m.Content))
	}
	estimated := total / 4
	if estimated < 100 {
		estimated = 100
	}
	return estimated
}

func wrapLastErr(err error) error {
	if err == nil {
		return fmt.Errorf("no error recorded")
	}
	return err
}

// SupportsVision 返回第一个未熔断 provider 的 Vision 支持状态
func (c *FallbackChain) SupportsVision() bool {
	for _, e := range c.entries {
		if e.breaker.Allow() {
			return e.provider.SupportsVision()
		}
	}
	return false
}

// ChatWithImages 降级链的 Vision 调用：优先选支持 Vision 的未熔断 provider
func (c *FallbackChain) ChatWithImages(ctx context.Context, prompt string, images []string) (string, error) {
	logger.Info("fallback chain vision start", "provider_count", len(c.entries), "prompt_length", len(prompt), "image_count", len(images))
	for i, entry := range c.entries {
		providerName := entry.provider.Name()
		if !entry.breaker.Allow() {
			logger.Warn("vision breaker open, skipping provider", "provider", providerName)
			continue
		}
		if !entry.provider.SupportsVision() {
			logger.Debug("vision unsupported provider skipped", "provider", providerName)
			continue // 跳过不支持 Vision 的 provider
		}
		result, err := entry.provider.ChatWithImages(ctx, prompt, images)
		if err != nil {
			entry.breaker.RecordFailure()
			logger.Warn("ChatWithImages failed, trying next vision provider",
				"provider", providerName, "err", err)
			continue
		}
		entry.breaker.RecordSuccess()
		if i > 0 {
			logger.Info("fallback vision provider used", "provider", providerName)
		}
		return result, nil
	}
	// 全部 Vision provider 失败，降级为纯文本
	logger.Warn("fallback chain vision fallback to text chat", "provider_count", len(c.entries))
	return c.Chat(ctx, prompt)
}
