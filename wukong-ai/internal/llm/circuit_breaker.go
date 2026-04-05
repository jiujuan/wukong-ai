package llm

import (
	"fmt"
	"sync"
	"time"

	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// State 熔断器状态
type State int

const (
	StateClosed   State = iota // 正常放行
	StateOpen                  // 熔断中，快速失败
	StateHalfOpen              // 半开，放行一次试探
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker 熔断器（三态状态机）
//
//	Closed  ──(连续失败 >= threshold)──>  Open
//	Open    ──(超过 timeout)──>           HalfOpen
//	HalfOpen──(调用成功)──>               Closed
//	HalfOpen──(调用失败)──>               Open（重新计时）
type CircuitBreaker struct {
	providerName string
	state        State
	failures     int           // 当前连续失败次数
	threshold    int           // 触发熔断的连续失败阈值（默认 5）
	timeout      time.Duration // 熔断持续时间（默认 60s）
	lastFailTime time.Time     // 最近一次失败时间
	mu           sync.Mutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(providerName string, threshold int, timeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 5
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	cb := &CircuitBreaker{
		providerName: providerName,
		state:        StateClosed,
		threshold:    threshold,
		timeout:      timeout,
	}
	logger.Info("circuit breaker initialized", "provider", providerName, "threshold", threshold, "timeout_s", timeout.Seconds())
	return cb
}

// Allow 判断是否允许本次调用
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true

	case StateOpen:
		// 超过 timeout 进入半开，放行一次试探
		if time.Since(cb.lastFailTime) >= cb.timeout {
			cb.state = StateHalfOpen
			logger.Info("circuit breaker half-open, probing",
				"provider", cb.providerName)
			return true
		}
		return false // 熔断中，快速失败

	case StateHalfOpen:
		return true // 半开状态放行试探请求
	}
	return false
}

// RecordSuccess 记录成功，重置熔断器到 Closed 状态
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		logger.Info("circuit breaker closed after successful probe",
			"provider", cb.providerName)
	}
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure 记录失败，达到阈值时触发熔断
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.threshold {
			cb.state = StateOpen
			logger.Warn("circuit breaker opened",
				"provider", cb.providerName,
				"failures", cb.failures,
				"threshold", cb.threshold,
			)
		} else {
			logger.Warn("circuit breaker failure recorded",
				"provider", cb.providerName,
				"failures", cb.failures,
				"threshold", cb.threshold,
			)
		}
	case StateHalfOpen:
		// 试探失败，重新熔断
		cb.state = StateOpen
		cb.failures = cb.threshold // 保持在阈值，避免立即再次进入半开
		logger.Warn("circuit breaker re-opened after failed probe",
			"provider", cb.providerName)
	}
}

// State 获取当前状态（用于监控/日志）
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Stats 返回熔断器当前统计信息（用于健康检查接口）
func (cb *CircuitBreaker) Stats() map[string]any {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return map[string]any{
		"provider":   cb.providerName,
		"state":      cb.state.String(),
		"failures":   cb.failures,
		"threshold":  cb.threshold,
		"timeout_s":  cb.timeout.Seconds(),
		"last_fail":  fmt.Sprintf("%v", cb.lastFailTime),
	}
}
