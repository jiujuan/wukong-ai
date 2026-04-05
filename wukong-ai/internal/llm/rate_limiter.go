package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// TokenBucket 令牌桶（线程安全）
// 用于同时支持 RPM（请求/分钟）和 TPM（Token/分钟）两个维度的限流。
type TokenBucket struct {
	mu       sync.Mutex
	capacity int64   // 桶容量（= maxRPM 或 maxTPM）
	tokens   float64 // 当前令牌数
	rate     float64 // 每纳秒补充的令牌数
	lastTime time.Time
}

func newTokenBucket(capacity int64, perMinute int64) *TokenBucket {
	if perMinute <= 0 {
		perMinute = capacity
	}
	// 每纳秒补充速率
	rate := float64(perMinute) / float64(time.Minute)
	return &TokenBucket{
		capacity: capacity,
		tokens:   float64(capacity), // 初始满桶
		rate:     rate,
		lastTime: time.Now(),
	}
}

// take 消耗 n 个令牌，返回需要等待的时间（0 表示立即可用）
func (b *TokenBucket) take(n int64) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastTime)
	b.lastTime = now

	// 补充令牌（不超过桶容量）
	b.tokens += float64(elapsed) * b.rate
	if b.tokens > float64(b.capacity) {
		b.tokens = float64(b.capacity)
	}

	if b.tokens >= float64(n) {
		b.tokens -= float64(n)
		return 0
	}

	// 计算需要等待多久
	deficit := float64(n) - b.tokens
	waitNs := time.Duration(deficit / b.rate)
	b.tokens = 0
	return waitNs
}

// LLMRateLimiter 双维度限流器（RPM + TPM）
type LLMRateLimiter struct {
	providerName string
	rpmBucket    *TokenBucket // 请求频率桶
	tpmBucket    *TokenBucket // Token 消耗桶
	enabled      bool
}

// NewRateLimiter 创建限流器
// maxRPM=0 或 maxTPM=0 表示对应维度不限流
func NewRateLimiter(providerName string, maxRPM, maxTPM int) *LLMRateLimiter {
	rl := &LLMRateLimiter{
		providerName: providerName,
		enabled:      maxRPM > 0 || maxTPM > 0,
	}
	if maxRPM > 0 {
		rl.rpmBucket = newTokenBucket(int64(maxRPM), int64(maxRPM))
	}
	if maxTPM > 0 {
		rl.tpmBucket = newTokenBucket(int64(maxTPM), int64(maxTPM))
	}
	logger.Info("rate limiter initialized", "provider", providerName, "enabled", rl.enabled, "max_rpm", maxRPM, "max_tpm", maxTPM)
	return rl
}

// Wait 等待直到可以发起调用，或 context 取消
// estimatedTokens：预估本次调用的 Token 数（用于 TPM 桶）；传 0 表示不计 Token
func (r *LLMRateLimiter) Wait(ctx context.Context, estimatedTokens int) error {
	if !r.enabled {
		return nil
	}
	logger.Debug("rate limiter wait start", "provider", r.providerName, "estimated_tokens", estimatedTokens)

	// RPM 限流
	if r.rpmBucket != nil {
		if wait := r.rpmBucket.take(1); wait > 0 {
			logger.Warn("rate limiter rpm throttled", "provider", r.providerName, "wait_ms", wait.Milliseconds())
			select {
			case <-ctx.Done():
				logger.Warn("rate limiter rpm wait cancelled", "provider", r.providerName, "err", ctx.Err())
				return fmt.Errorf("rate limit wait cancelled (RPM): %w", ctx.Err())
			case <-time.After(wait):
			}
		}
	}

	// TPM 限流
	if r.tpmBucket != nil && estimatedTokens > 0 {
		if wait := r.tpmBucket.take(int64(estimatedTokens)); wait > 0 {
			logger.Warn("rate limiter tpm throttled", "provider", r.providerName, "wait_ms", wait.Milliseconds(), "estimated_tokens", estimatedTokens)
			select {
			case <-ctx.Done():
				logger.Warn("rate limiter tpm wait cancelled", "provider", r.providerName, "err", ctx.Err())
				return fmt.Errorf("rate limit wait cancelled (TPM): %w", ctx.Err())
			case <-time.After(wait):
			}
		}
	}

	logger.Debug("rate limiter wait done", "provider", r.providerName)
	return nil
}

// IsEnabled 是否启用了限流
func (r *LLMRateLimiter) IsEnabled() bool {
	return r.enabled
}
