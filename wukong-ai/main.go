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
	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化日志
	logger.Init(cfg.Server.Env)
	logger.Info("starting Wukong-AI server",
		"version", "v1.0",
		"env", cfg.Server.Env,
		"port", cfg.Server.Port,
	)

	// 3. 初始化数据库
	db.Init(cfg.Database.DSN)
	if err := db.RunMigrations(); err != nil {
		logger.Warn("failed to run migrations", "err", err)
	}

	// 4. 构建 LLM 降级链（FallbackChain 实现 LLM 接口，对 server 层完全透明）
	llmFactory := llm.NewFactory()
	fallbackChain, err := llmFactory.BuildFallbackChain(&cfg.LLM)
	if err != nil {
		logger.Error("failed to build LLM fallback chain", "err", err)
		os.Exit(1)
	}

	// FallbackChain 实现了 llm.LLM 接口，直接传给 server
	var llmProvider llm.LLM = fallbackChain

	logger.Info("LLM fallback chain initialized",
		"primary_provider", cfg.LLM.Provider,
		"primary_model", cfg.LLM.Model,
		"fallback_count", len(cfg.LLM.Fallbacks),
	)

	// 5. 创建并启动服务器
	srv := server.New(cfg, llmProvider)

	ctx, cancel := context.WithCancel(context.Background())

	go srv.Start(ctx, llmProvider)

	// 6. 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	cancel()
	srv.Stop()

	if err := db.Close(); err != nil {
		logger.Error("failed to close database", "err", err)
	}

	logger.Info("server exited")
}
