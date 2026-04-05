package parser

import (
	"archive/zip"
	"context"
	"io"
	"regexp"
	"strings"
)

// DocxParser 处理 .docx 文件
// .docx 是 ZIP 格式，内含 word/document.xml；用正则提取 <w:t> 标签文本，纯 Go 无 cgo。
type DocxParser struct{}

func (p *DocxParser) Parse(_ context.Context, filePath string) (*ParseResult, error) {
	text, err := extractDocxText(filePath)
	if err != nil {
		return nil, err
	}
	return &ParseResult{
		Text:     text,
		Metadata: map[string]string{"type": "docx"},
	}, nil
}

func (p *DocxParser) SupportedMIME() []string {
	return []string{"application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
}

// extractDocxText 从 .docx ZIP 中读取 word/document.xml，提取 <w:t> 文本
func extractDocxText(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "word/document.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 10*1024*1024))
		rc.Close()
		if err != nil {
			return "", err
		}
		return parseWordXML(string(data)), nil
	}
	return "", nil
}

var wTextRe = regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
var parRe   = regexp.MustCompile(`<w:p[ />]`)

// parseWordXML 提取 XML 中的 <w:t> 文本，段落间加换行
func parseWordXML(xml string) string {
	// 将段落标记替换为换行标记
	xml = parRe.ReplaceAllString(xml, "\n<w:p ")

	var sb strings.Builder
	for _, m := range wTextRe.FindAllStringSubmatch(xml, -1) {
		if len(m) > 1 {
			sb.WriteString(m[1])
		}
	}
	// 合并多余空行
	result := strings.TrimSpace(sb.String())
	result = regexp.MustCompile(`\n{3,}`).ReplaceAllString(result, "\n\n")
	return result
}
