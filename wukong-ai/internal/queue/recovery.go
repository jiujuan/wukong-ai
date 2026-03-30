package queue

import (
	"time"

	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Recovery 任务恢复
type Recovery struct{}

// NewRecovery 创建恢复器
func NewRecovery() *Recovery {
	return &Recovery{}
}

// RecoverStaleJobs 恢复僵尸任务
// 服务启动时调用，将上次崩溃遗留的 running 状态任务重置为 queued
func (r *Recovery) RecoverStaleJobs(staleDuration time.Duration) error {
	logger.Info("recovering stale jobs", "stale_duration", staleDuration.String())

	err := repository.RecoverStaleJobs(staleDuration)
	if err != nil {
		logger.Error("failed to recover stale jobs", "err", err)
		return err
	}

	logger.Info("stale jobs recovered successfully")
	return nil
}

// CleanOrphanedRecords 清理孤立记录
// 清理没有对应任务队列记录的孤立数据库记录
func (r *Recovery) CleanOrphanedRecords() error {
	logger.Info("cleaning orphaned records")

	// 这个功能需要根据具体业务逻辑实现
	// 目前为空，后续可以根据需要扩展

	return nil
}
