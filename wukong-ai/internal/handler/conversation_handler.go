package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/conversation"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/pkg/logger"
	"github.com/jiujuan/wukong-ai/pkg/uuid"
)

// ConversationHandler 对话相关 Handler
type ConversationHandler struct{}

func NewConversationHandler() *ConversationHandler {
	return &ConversationHandler{}
}

// ── POST /api/conversation ───────────────────────────────────────────────────

type CreateConvRequest struct {
	Title string `json:"title"` // 可选，空时由后端自动填 "新对话"
}

type CreateConvResponse struct {
	ConversationID string    `json:"conversation_id"`
	Title          string    `json:"title"`
	CreateTime     time.Time `json:"create_time"`
}

// Create 创建对话
func (h *ConversationHandler) Create(c *gin.Context) {
	var req CreateConvRequest
	_ = c.ShouldBindJSON(&req)

	title := req.Title
	if title == "" {
		title = "新对话"
	}

	now := time.Now()
	conv := &conversation.Conversation{
		ID:         uuid.NewConversationID(),
		Title:      title,
		TurnCount:  0,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := repository.CreateConversation(conv); err != nil {
		logger.Error("failed to create conversation", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, CreateConvResponse{
		ConversationID: conv.ID,
		Title:          conv.Title,
		CreateTime:     conv.CreateTime,
	})
}

// ── GET /api/conversation/:id ────────────────────────────────────────────────

type ConvDetailResponse struct {
	Conversation *conversation.Conversation `json:"conversation"`
	Turns        []conversation.Turn        `json:"turns"`
}

// GetDetail 获取对话详情及所有轮次
func (h *ConversationHandler) GetDetail(c *gin.Context) {
	convID := c.Param("id")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id is required"})
		return
	}

	conv, err := repository.GetConversation(convID)
	if err != nil {
		logger.Error("failed to get conversation", "id", convID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	turns, err := repository.GetTurns(convID)
	if err != nil {
		logger.Error("failed to get turns", "id", convID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get turns"})
		return
	}

	c.JSON(http.StatusOK, ConvDetailResponse{
		Conversation: conv,
		Turns:        turns,
	})
}

// ── GET /api/conversation/list ───────────────────────────────────────────────

type ConvListResponse struct {
	Total         int                          `json:"total"`
	Page          int                          `json:"page"`
	Size          int                          `json:"size"`
	Conversations []*conversation.Conversation `json:"conversations"`
}

// List 对话列表（分页）
func (h *ConversationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	convs, total, err := repository.ListConversations(page, size)
	if err != nil {
		logger.Error("failed to list conversations", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, ConvListResponse{
		Total:         total,
		Page:          page,
		Size:          size,
		Conversations: convs,
	})
}

// ── DELETE /api/conversation/:id ─────────────────────────────────────────────

// Delete 删除对话（级联删除所有轮次）
func (h *ConversationHandler) Delete(c *gin.Context) {
	convID := c.Param("id")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id is required"})
		return
	}

	conv, err := repository.GetConversation(convID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check conversation"})
		return
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	if err := repository.DeleteConversation(convID); err != nil {
		logger.Error("failed to delete conversation", "id", convID, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted", "id": convID})
}
