package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/event"
	"github.com/jiujuan/wukong-ai/internal/handler"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/middleware"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/state"
	"github.com/jiujuan/wukong-ai/pkg/config"
)

// Router 路由配置
type Router struct {
	engine      *gin.Engine
	cfg         *config.AppConfig
	eventBus    *event.EventBus
	stateMgr    *state.Manager
	queue       *queue.PersistentQueue
	llmProvider llm.LLM
}

// NewRouter 创建路由
func NewRouter(cfg *config.AppConfig, eventBus *event.EventBus, stateMgr *state.Manager, q *queue.PersistentQueue, llmProvider llm.LLM) *Router {
	if cfg.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.RecoveryLogger())
	engine.Use(corsMiddleware())

	return &Router{
		engine:      engine,
		cfg:         cfg,
		eventBus:    eventBus,
		stateMgr:    stateMgr,
		queue:       q,
		llmProvider: llmProvider,
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Setup 注册路由
func (r *Router) Setup() {
	// 创建处理器
	runHandler := handler.NewRunHandler(r.queue)
	taskHandler := handler.NewTaskHandler()
	listHandler := handler.NewListHandler()
	dagHandler := handler.NewDagHandler()
	streamHandler := handler.NewStreamHandler(r.eventBus)
	cancelHandler := handler.NewCancelHandler(r.queue)

	// 健康检查
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由
	api := r.engine.Group("/api")
	{
		// 任务相关
		api.POST("/run", runHandler.Handle)
		api.POST("/resume", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "resume endpoint, use /run with resume_from parameter"})
		})
		api.GET("/task", taskHandler.Handle)
		api.GET("/task/dag", dagHandler.Handle)
		api.GET("/task/stream", streamHandler.Handle)
		api.POST("/task/cancel", cancelHandler.Handle)
		api.GET("/list", listHandler.Handle)
	}

	// 前端 UI 路由 (如果需要)
	ui := r.engine.Group("/ui")
	{
		ui.GET("/list", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Frontend UI should be served separately"})
		})
	}
}

// GetEngine 获取 Gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
