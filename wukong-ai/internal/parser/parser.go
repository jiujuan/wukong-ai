// Package parser 提供统一的文件内容解析接口。
// 每种文件类型对应一个 Parser 实现，由 NewParser 工厂方法按 MIME 类型分发。
// 所有 Parser 均不引入需要 cgo 的外部依赖，保持纯 Go 实现。
// 需要 OCR / pdfcpu / excelize 的功能通过可选降级策略处理。
package parser

import (
	"context"
	"fmt"
	"strings"
)

// ParseResult 统一解析结果
type ParseResult struct {
	Text     string            // 提取的纯文本（文档/表格/OCR 结果）
	ImageB64 string            // base64 图片数据（Vision 模式使用）
	Metadata map[string]string // 元信息（页数、尺寸等）
	IsImage  bool              // 是否为图片，决定走 Vision 还是文本路径
}

// Parser 统一解析接口
type Parser interface {
	Parse(ctx context.Context, filePath string) (*ParseResult, error)
	SupportedMIME() []string
}

// NewParser 工厂方法：按 MIME 类型返回对应 Parser
func NewParser(mimeType string) (Parser, error) {
	switch {
	case mimeType == "text/plain" || mimeType == "text/markdown":
		return &PlainTextParser{}, nil
	case mimeType == "application/pdf":
		return &PDFParser{}, nil
	case strings.Contains(mimeType, "wordprocessingml"):
		return &DocxParser{}, nil
	case mimeType == "text/csv":
		return &CSVParser{}, nil
	case strings.Contains(mimeType, "spreadsheetml"):
		return &XLSXParser{}, nil
	case strings.HasPrefix(mimeType, "image/"):
		return &ImageParser{}, nil
	default:
		return nil, fmt.Errorf("no parser for MIME type: %s", mimeType)
	}
}
