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
func NewRouter(cfg *config.AppConfig, eventBus *event.EventBus, stateMgr *state.Manager,
	q *queue.PersistentQueue, llmProvider llm.LLM) *Router {

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
	runHandler          := handler.NewRunHandler(r.queue, r.cfg)
	resumeHandler       := handler.NewResumeHandler(r.queue, r.cfg)
	taskHandler         := handler.NewTaskHandler()
	listHandler         := handler.NewListHandler()
	dagHandler          := handler.NewDagHandler()
	streamHandler       := handler.NewStreamHandler(r.eventBus)
	cancelHandler       := handler.NewCancelHandler(r.queue)
	conversationHandler := handler.NewConversationHandler()

	// 健康检查
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": "v0.9"})
	})

	api := r.engine.Group("/api")
	{
		// ── 任务相关 ────────────────────────────────────────────
		api.POST("/run",         runHandler.Handle)
		api.POST("/resume",      resumeHandler.Handle)
		api.GET("/task",         taskHandler.Handle)
		api.GET("/task/dag",     dagHandler.Handle)
		api.GET("/task/stream",  streamHandler.Handle)
		api.POST("/task/cancel", cancelHandler.Handle)
		api.GET("/list",         listHandler.Handle)

		// ── 多轮对话 ────────────────────────────────────────────
		// POST /api/conversation          创建对话
		// POST /api/conversation/:id/run  在对话中提交新一轮任务（复用 runHandler）
		// GET  /api/conversation/list     对话列表
		// GET  /api/conversation/:id      对话详情 + 所有轮次
		// DELETE /api/conversation/:id    删除对话
		conv := api.Group("/conversation")
		{
			conv.POST("",          conversationHandler.Create)
			conv.GET("/list",      conversationHandler.List)
			conv.GET("/:id",       conversationHandler.GetDetail)
			conv.DELETE("/:id",    conversationHandler.Delete)
			// 在指定对话中提交新轮任务：前端传 conversation_id 到 /api/run 即可，
			// 此路由作为语义化快捷入口，内部委托给 runHandler
			conv.POST("/:id/run",  r.convRunHandler(runHandler))
		}
	}

	// 前端 UI 路由
	ui := r.engine.Group("/ui")
	{
		ui.GET("/list", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Frontend UI should be served separately"})
		})
	}
}

// convRunHandler 语义化路由 POST /api/conversation/:id/run
// 自动将路径中的 :id 填入请求体的 conversation_id 字段
func (r *Router) convRunHandler(runHandler *handler.RunHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		convID := c.Param("id")
		// 将 conversation_id 注入到请求体前，先在 context 里透传
		c.Set("override_conversation_id", convID)
		runHandler.HandleWithConvID(c, convID)
	}
}

// GetEngine 获取 Gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
