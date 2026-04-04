package llm

import (
	"sync"
	"time"
)

// ProviderMetrics 单个 provider 的调用统计
type ProviderMetrics struct {
	TotalCalls   int64         `json:"total_calls"`
	Successes    int64         `json:"successes"`
	Failures     int64         `json:"failures"`
	Skipped      int64         `json:"skipped"`     // 熔断跳过次数
	TotalLatency time.Duration `json:"total_latency_ms"`
	AvgLatencyMs int64         `json:"avg_latency_ms"`
}

// CallMetrics 线程安全的调用统计表
type CallMetrics struct {
	mu      sync.RWMutex
	metrics map[string]*ProviderMetrics
}

// NewCallMetrics 创建调用统计
func NewCallMetrics() *CallMetrics {
	return &CallMetrics{
		metrics: make(map[string]*ProviderMetrics),
	}
}

// Record 记录一次调用结果
func (m *CallMetrics) Record(provider, status string, elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pm, ok := m.metrics[provider]
	if !ok {
		pm = &ProviderMetrics{}
		m.metrics[provider] = pm
	}

	pm.TotalCalls++
	switch status {
	case "success", "stream_success":
		pm.Successes++
		pm.TotalLatency += elapsed
		if pm.Successes > 0 {
			pm.AvgLatencyMs = pm.TotalLatency.Milliseconds() / pm.Successes
		}
	case "failed", "stream_failed":
		pm.Failures++
	case "skipped":
		pm.Skipped++
		pm.TotalCalls-- // 跳过不算调用
	}
}

// Get 获取指定 provider 的统计（返回副本）
func (m *CallMetrics) Get(provider string) ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if pm, ok := m.metrics[provider]; ok {
		return *pm
	}
	return ProviderMetrics{}
}

// GetAll 获取所有 provider 的统计
func (m *CallMetrics) GetAll() map[string]ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]ProviderMetrics, len(m.metrics))
	for k, v := range m.metrics {
		result[k] = *v
	}
	return result
}
