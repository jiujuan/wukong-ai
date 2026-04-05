package parser

import (
	"context"
	"os"
)

// PlainTextParser 处理 .txt / .md 文件（纯文本直读）
type PlainTextParser struct{}

func (p *PlainTextParser) Parse(_ context.Context, filePath string) (*ParseResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &ParseResult{
		Text:     string(data),
		Metadata: map[string]string{"type": "plaintext"},
	}, nil
}

func (p *PlainTextParser) SupportedMIME() []string {
	return []string{"text/plain", "text/markdown"}
}
