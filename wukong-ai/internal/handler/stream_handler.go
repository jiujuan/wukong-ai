package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/event"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// StreamHandler SSE 流处理器
type StreamHandler struct {
	eventBus *event.EventBus
}

// NewStreamHandler 创建流处理器
func NewStreamHandler(eventBus *event.EventBus) *StreamHandler {
	return &StreamHandler{
		eventBus: eventBus,
	}
}

// Handle 处理 SSE 流请求
func (h *StreamHandler) Handle(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	// 设置 SSE 头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 订阅事件
	sub := h.eventBus.Subscribe(taskID)
	defer h.eventBus.Unsubscribe(taskID, sub)

	clientGone := c.Request.Context().Done()

	// 流式响应
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			logger.Debug("client disconnected", "task_id", taskID)
			return false
		case event, ok := <-sub:
			if !ok {
				return false
			}

			data, err := json.Marshal(event)
			if err != nil {
				logger.Error("failed to marshal event", "err", err)
				return true
			}

			c.SSEvent(event.Type, string(data))

			// 终态事件关闭流
			if event.Type == "task_done" || event.Type == "task_failed" {
				return false
			}

			return true
		}
	})
}
