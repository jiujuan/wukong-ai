package subagent

import (
	"sync"
)

// Pool 协程池
type Pool struct {
	maxConcurrent int
	semaphore     chan struct{}
}

// NewPool 创建新的协程池
func NewPool(maxConcurrent int) *Pool {
	return &Pool{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

// Acquire 获取信号量
func (p *Pool) Acquire() {
	p.semaphore <- struct{}{}
}

// Release 释放信号量
func (p *Pool) Release() {
	<-p.semaphore
}

// GetMaxConcurrent 获取最大并发数
func (p *Pool) GetMaxConcurrent() int {
	return p.maxConcurrent
}

// PoolWithWaitGroup 带 WaitGroup 的协程池
type PoolWithWaitGroup struct {
	pool   *Pool
	wg     sync.WaitGroup
	mu     sync.Mutex
	errors []error
}

// NewPoolWithWaitGroup 创建带 WaitGroup 的协程池
func NewPoolWithWaitGroup(maxConcurrent int) *PoolWithWaitGroup {
	return &PoolWithWaitGroup{
		pool:   NewPool(maxConcurrent),
		errors: make([]error, 0),
	}
}

// Run 在协程中执行任务
func (p *PoolWithWaitGroup) Run(fn func() error) {
	p.pool.Acquire()
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.pool.Release()

		if err := fn(); err != nil {
			p.mu.Lock()
			p.errors = append(p.errors, err)
			p.mu.Unlock()
		}
	}()
}

// Wait 等待所有任务完成
func (p *PoolWithWaitGroup) Wait() []error {
	p.wg.Wait()
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.errors
}
