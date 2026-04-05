package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/wukong-ai/internal/db/repository"
	"github.com/jiujuan/wukong-ai/internal/parser"
	"github.com/jiujuan/wukong-ai/internal/upload"
	"github.com/jiujuan/wukong-ai/pkg/logger"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	service   *upload.UploadService
	validator *upload.FileValidator
	extractor interface {
		Enqueue(parser.ExtractJob)
	}
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(
	service *upload.UploadService,
	extractor interface{ Enqueue(parser.ExtractJob) },
) *UploadHandler {
	return &UploadHandler{
		service:   service,
		validator: upload.NewFileValidator(),
		extractor: extractor,
	}
}

// UploadResult 单个文件上传结果
type UploadResult struct {
	FileName     string `json:"file_name"`
	AttachmentID int64  `json:"attachment_id"`
	Success      bool   `json:"success"`
	Status       string `json:"status,omitempty"`
	Error        string `json:"error,omitempty"`
}

// Handle POST /api/upload
func (h *UploadHandler) Handle(c *gin.Context) {
	taskID := c.PostForm("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	// 校验任务是否存在
	task, err := repository.GetTaskByID(taskID)
	if err != nil || task == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task not found: " + taskID})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}
	if len(files) > upload.MaxFileCount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "too many files, max " + string(rune('0'+upload.MaxFileCount)),
		})
		return
	}

	var results []UploadResult
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			results = append(results, UploadResult{
				FileName: header.Filename, Success: false,
				Error: "failed to open file: " + err.Error(),
			})
			continue
		}

		// 安全校验（大小 + MIME 白名单 + 魔数）
		if err := h.validator.Validate(header, file); err != nil {
			file.Close()
			results = append(results, UploadResult{
				FileName: header.Filename, Success: false, Error: err.Error(),
			})
			continue
		}

		// 持久化存储
		att, err := h.service.Save(c.Request.Context(), taskID, header, file)
		file.Close()
		if err != nil {
			logger.Error("failed to save upload", "file", header.Filename, "err", err)
			results = append(results, UploadResult{
				FileName: header.Filename, Success: false, Error: err.Error(),
			})
			continue
		}

		// 投递异步提取任务
		if h.extractor != nil {
			h.extractor.Enqueue(parser.ExtractJob{
				AttachmentID: att.ID,
				TaskID:       taskID,
			})
		}

		results = append(results, UploadResult{
			FileName:     header.Filename,
			AttachmentID: att.ID,
			Success:      true,
			Status:       "extracting",
		})
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// HandleStatus GET /api/upload/status?task_id={id}
func (h *UploadHandler) HandleStatus(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id is required"})
		return
	}

	atts, err := repository.GetAttachmentsByTaskID(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":     taskID,
		"attachments": atts,
	})
}
