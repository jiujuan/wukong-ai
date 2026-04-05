package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/event"
	"github.com/jiujuan/wukong-ai/internal/handler"
	"github.com/jiujuan/wukong-ai/internal/llm"
	"github.com/jiujuan/wukong-ai/internal/middleware"
	"github.com/jiujuan/wukong-ai/internal/parser"
	"github.com/jiujuan/wukong-ai/internal/queue"
	"github.com/jiujuan/wukong-ai/internal/state"
	"github.com/jiujuan/wukong-ai/internal/upload"
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
	extractor   *parser.Extractor   // v1.1 附件提取器
	uploadSvc   *upload.UploadService // v1.1 上传服务
}

// NewRouter 创建路由
func NewRouter(cfg *config.AppConfig, eventBus *event.EventBus, stateMgr *state.Manager,
	q *queue.PersistentQueue, llmProvider llm.LLM,
	extractor *parser.Extractor, uploadSvc *upload.UploadService) *Router {

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
		extractor:   extractor,
		uploadSvc:   uploadSvc,
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

// Setup 注册所有路由
func (r *Router) Setup() {
	runHandler          := handler.NewRunHandler(r.queue, r.cfg)
	resumeHandler       := handler.NewResumeHandler(r.queue, r.cfg)
	taskHandler         := handler.NewTaskHandler()
	listHandler         := handler.NewListHandler()
	dagHandler          := handler.NewDagHandler()
	streamHandler       := handler.NewStreamHandler(r.eventBus)
	cancelHandler       := handler.NewCancelHandler(r.queue)
	conversationHandler := handler.NewConversationHandler()
	healthHandler       := handler.NewHealthHandler(r.llmProvider)
	uploadHandler       := handler.NewUploadHandler(r.uploadSvc, r.extractor) // v1.1

	// 健康检查
	r.engine.GET("/health",     healthHandler.Handle)
	r.engine.GET("/health/llm", healthHandler.HandleLLM)

	api := r.engine.Group("/api")
	{
		// 任务相关
		api.POST("/run",         runHandler.Handle)
		api.POST("/resume",      resumeHandler.Handle)
		api.GET("/task",         taskHandler.Handle)
		api.GET("/task/dag",     dagHandler.Handle)
		api.GET("/task/stream",  streamHandler.Handle)
		api.POST("/task/cancel", cancelHandler.Handle)
		api.GET("/list",         listHandler.Handle)

		// 多轮对话（v0.9）
		conv := api.Group("/conversation")
		{
			conv.POST("",         conversationHandler.Create)
			conv.GET("/list",     conversationHandler.List)
			conv.GET("/:id",      conversationHandler.GetDetail)
			conv.DELETE("/:id",   conversationHandler.Delete)
			conv.POST("/:id/run", r.convRunHandler(runHandler))
		}

		// 文件上传（v1.1）
		api.POST("/upload",        uploadHandler.Handle)
		api.GET("/upload/status",  uploadHandler.HandleStatus)
	}

	ui := r.engine.Group("/ui")
	{
		ui.GET("/list", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Frontend UI should be served separately"})
		})
	}
}

// convRunHandler POST /api/conversation/:id/run
func (r *Router) convRunHandler(runHandler *handler.RunHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		convID := c.Param("id")
		c.Set("override_conversation_id", convID)
		runHandler.HandleWithConvID(c, convID)
	}
}

// GetEngine 获取 Gin 引擎
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
