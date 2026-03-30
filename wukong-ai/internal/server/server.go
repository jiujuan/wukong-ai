package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jiujuan/wukong-ai/internal/db"
	"github.com/jiujuan/wukong-ai/internal/event"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/state"
	"github.com/jiujuan/wukong-ai/internal/worker"
	"github.com/jiujuan/wukong-ai/pkg/config"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// Server HTTP 服务器
type Server struct {
	cfg         *config.AppConfig
	httpServer  *http.Server
	eventBus    *event.EventBus
	stateMgr    *state.Manager
	queue       *queue.PersistentQueue
	workerPool  *worker.Pool
	llmProvider llm.LLM
	restartCh   chan struct{}
	serverMu    sync.Mutex
}

// New 创建新的服务器
func New(cfg *config.AppConfig, llmProvider llm.LLM) *Server {
	// 初始化数据库
	db.Init(cfg.Database.DSN)

	// 执行迁移
	if err := db.RunMigrations(); err != nil {
		logger.Warn("failed to run migrations", "err", err)
	}

	// 创建组件
	eventBus := event.NewEventBus()
	stateMgr := state.NewManager("tasks")
	taskQueue := queue.NewPersistentQueue()

	// 恢复僵尸任务
	if err := taskQueue.RecoverStaleJobs(cfg.Agent.StaleJobTimeout); err != nil {
		logger.Warn("failed to recover stale jobs", "err", err)
	}

	// 创建 Worker Pool
	workerPool := worker.NewPool(cfg.Agent.MaxWorkers, taskQueue, worker.NewWorkflowEventBusAdapter(eventBus))

	s := &Server{
		cfg:         cfg,
		eventBus:    eventBus,
		stateMgr:    stateMgr,
		queue:       taskQueue,
		workerPool:  workerPool,
		llmProvider: llmProvider,
		restartCh:   make(chan struct{}, 1),
	}
	s.httpServer = s.newHTTPServer()
	return s
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context, llmProvider llm.LLM) {
	s.workerPool.Start(ctx, llmProvider, &s.cfg.Agent, s.cfg.Prompts.Dir)
	go s.watchAndRestart(ctx)

	for {
		server := s.currentHTTPServer()
		errCh := make(chan error, 1)
		go func(httpServer *http.Server) {
			logger.Info("starting HTTP server", "port", s.cfg.Server.Port)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- err
				return
			}
			errCh <- nil
		}(server)

		select {
		case <-ctx.Done():
			s.shutdownHTTPServer(server)
			return
		case <-s.restartCh:
			logger.Warn("file changed, restarting HTTP server")
			s.shutdownHTTPServer(server)
			s.replaceHTTPServer()
			<-errCh
		case err := <-errCh:
			if err != nil {
				logger.Error("HTTP server error", "err", err)
			}
			return
		}
	}
}

// Stop 停止服务器
func (s *Server) Stop() {
	logger.Info("stopping server")

	s.shutdownHTTPServer(s.currentHTTPServer())
	s.workerPool.GracefulStop()

	if err := db.Close(); err != nil {
		logger.Error("failed to close database", "err", err)
	}

	logger.Info("server stopped")
}

// GetEventBus 获取事件总线
func (s *Server) GetEventBus() *event.EventBus {
	return s.eventBus
}

// GetStateManager 获取状态管理器
func (s *Server) GetStateManager() *state.Manager {
	return s.stateMgr
}

// GetQueue 获取任务队列
func (s *Server) GetQueue() *queue.PersistentQueue {
	return s.queue
}

func (s *Server) newHTTPServer() *http.Server {
	router := NewRouter(s.cfg, s.eventBus, s.stateMgr, s.queue, s.llmProvider)
	router.Setup()
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", s.cfg.Server.Port),
		Handler:      router.GetEngine(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func (s *Server) currentHTTPServer() *http.Server {
	s.serverMu.Lock()
	defer s.serverMu.Unlock()
	return s.httpServer
}

func (s *Server) replaceHTTPServer() {
	s.serverMu.Lock()
	s.httpServer = s.newHTTPServer()
	s.serverMu.Unlock()
}

func (s *Server) shutdownHTTPServer(httpServer *http.Server) {
	if httpServer == nil {
		return
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown HTTP server", "err", err)
	}
}

func (s *Server) watchAndRestart(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("failed to create watcher", "err", err)
		return
	}
	defer watcher.Close()

	root, err := os.Getwd()
	if err != nil {
		logger.Error("failed to get working directory", "err", err)
		return
	}

	if err := addRecursiveWatch(watcher, root); err != nil {
		logger.Error("failed to watch directories", "err", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case err := <-watcher.Errors:
			if err != nil {
				logger.Warn("file watcher error", "err", err)
			}
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && info.IsDir() {
					_ = addRecursiveWatch(watcher, event.Name)
				}
			}

			if !shouldTriggerRestart(event.Name, event.Op) {
				continue
			}

			select {
			case s.restartCh <- struct{}{}:
				logger.Info("detected file change", "file", event.Name, "operation", event.Op.String())
			default:
			}
		}
	}
}

func addRecursiveWatch(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if name == ".git" || name == "node_modules" || name == ".idea" || name == ".vscode" || name == "dist" || name == "build" {
			return filepath.SkipDir
		}
		return watcher.Add(path)
	})
}

func shouldTriggerRestart(path string, op fsnotify.Op) bool {
	if op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return false
	}
	name := strings.ToLower(path)
	return strings.HasSuffix(name, ".go") ||
		strings.HasSuffix(name, ".yaml") ||
		strings.HasSuffix(name, ".yml") ||
		strings.HasSuffix(name, ".toml") ||
		strings.HasSuffix(name, ".env")
}
