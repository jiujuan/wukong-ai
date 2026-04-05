package parser

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// CSVParser 处理 .csv 文件，转为 Markdown 表格
type CSVParser struct{}

func (p *CSVParser) Parse(_ context.Context, filePath string) (*ParseResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	return &ParseResult{
		Text:     csvToMarkdownTable(records),
		Metadata: map[string]string{"type": "csv", "rows": fmt.Sprintf("%d", len(records))},
	}, nil
}

func (p *CSVParser) SupportedMIME() []string {
	return []string{"text/csv"}
}

// csvToMarkdownTable 将二维记录转为 Markdown 表格字符串
func csvToMarkdownTable(records [][]string) string {
	if len(records) == 0 {
		return ""
	}
	var sb strings.Builder

	// 表头
	sb.WriteString("| ")
	sb.WriteString(strings.Join(records[0], " | "))
	sb.WriteString(" |\n")

	// 分隔行
	sb.WriteString("|")
	for range records[0] {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// 数据行
	for _, row := range records[1:] {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(row, " | "))
		sb.WriteString(" |\n")
	}
	return sb.String()
}
