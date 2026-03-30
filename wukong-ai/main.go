package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/server"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

func main() {
	// 1. 加载配置（pkg/config）
	cfg := config.Load()

	// 2. 初始化日志（pkg/logger）
	logger.Init(cfg.Server.Env)

	logger.Info("starting Wukong-AI server", "env", cfg.Server.Env, "port", cfg.Server.Port)

	// 3. 初始化数据库
	db.Init(cfg.Database.DSN)

	// 执行数据库迁移
	if err := db.RunMigrations(); err != nil {
		logger.Warn("failed to run migrations", "err", err)
	}

	// 4. 初始化 LLM
	llmFactory := llm.NewFactory()
	llmProvider, err := llmFactory.CreateLLM(&cfg.LLM)
	if err != nil {
		logger.Error("failed to create LLM", "err", err)
		os.Exit(1)
	}
	logger.Info("LLM initialized", "provider", cfg.LLM.Provider, "model", cfg.LLM.Model)

	// 5. 创建服务器
	srv := server.New(cfg, llmProvider)

	// 6. 创建上下文用于取消
	ctx, cancel := context.WithCancel(context.Background())

	// 7. 启动服务器
	go srv.Start(ctx, llmProvider)

	// 8. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// 9. 优雅关闭
	cancel()
	srv.Stop()

	// 10. 关闭数据库连接
	if err := db.Close(); err != nil {
		logger.Error("failed to close database", "err", err)
	}

	logger.Info("server exited")
}
