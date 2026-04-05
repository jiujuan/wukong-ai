package parser

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

// ImageParser 处理图片文件（.png / .jpg / .webp / .gif）
// 将图片读取为 base64，供 Vision LLM 直接理解；
// 若 LLM 不支持 Vision，由 image_processor.go 负责降级到文本描述。
type ImageParser struct{}

func (p *ImageParser) Parse(_ context.Context, filePath string) (*ParseResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filePath), "."))

	return &ParseResult{
		ImageB64: b64,
		IsImage:  true,
		Metadata: map[string]string{
			"path": filePath,
			"ext":  ext,
			"size": strings.Join([]string{}, ""), // placeholder
		},
	}, nil
}

func (p *ImageParser) SupportedMIME() []string {
	return []string{"image/png", "image/jpeg", "image/webp", "image/gif"}
}
