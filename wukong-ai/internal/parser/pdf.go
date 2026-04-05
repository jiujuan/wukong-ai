package parser

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// PDFParser 处理 .pdf 文件
// 纯 Go 实现：从 PDF 内容流提取 BT...ET 块中的字符串，无需 cgo。
// 对复杂/扫描 PDF 返回空文本，上层可降级到 Vision 路径。
type PDFParser struct{}

func (p *PDFParser) Parse(_ context.Context, filePath string) (*ParseResult, error) {
	text, err := extractPDFText(filePath)
	if err != nil {
		// 提取失败不报错，返回空文本让上层走 Vision 降级
		return &ParseResult{
			Text:     "",
			Metadata: map[string]string{"type": "pdf", "parse_error": err.Error()},
		}, nil
	}
	return &ParseResult{
		Text:     text,
		Metadata: map[string]string{"type": "pdf"},
	}, nil
}

func (p *PDFParser) SupportedMIME() []string {
	return []string{"application/pdf"}
}

// extractPDFText 从 PDF 字节流提取可读文本（BT...ET 块方式）
func extractPDFText(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 最多读 10MB
	data, err := io.ReadAll(io.LimitReader(f, 10*1024*1024))
	if err != nil {
		return "", err
	}

	content := string(data)
	btEtRe := regexp.MustCompile(`(?s)BT.*?ET`)
	strRe := regexp.MustCompile(`\(([^)\\]*(?:\\.[^)\\]*)*)\)`)

	var sb strings.Builder
	for _, block := range btEtRe.FindAllString(content, -1) {
		for _, m := range strRe.FindAllStringSubmatch(block, -1) {
			if len(m) > 1 {
				text := decodePDFString(m[1])
				if text != "" {
					sb.WriteString(text)
					sb.WriteByte(' ')
				}
			}
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return "", fmt.Errorf("no extractable text (scanned or encrypted PDF)")
	}
	return result, nil
}

func decodePDFString(s string) string {
	s = strings.NewReplacer(
		`\n`, "\n", `\r`, "\r", `\t`, "\t",
		`\\`, "\\", `\(`, "(", `\)`, ")",
	).Replace(s)
	var b strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
