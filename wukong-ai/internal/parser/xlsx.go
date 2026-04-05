package parser

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// XLSXParser 处理 .xlsx 文件
// 纯 Go：ZIP → xl/worksheets/sheet*.xml + xl/sharedStrings.xml 解析，转 Markdown 表格
type XLSXParser struct{}

func (p *XLSXParser) Parse(_ context.Context, filePath string) (*ParseResult, error) {
	text, err := extractXLSXText(filePath)
	if err != nil {
		return nil, err
	}
	return &ParseResult{
		Text:     text,
		Metadata: map[string]string{"type": "xlsx"},
	}, nil
}

func (p *XLSXParser) SupportedMIME() []string {
	return []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}
}

// extractXLSXText 从 .xlsx 中读取所有 Sheet，每个 Sheet 转为 Markdown 表格
func extractXLSXText(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	// 1. 读取 sharedStrings.xml（字符串池）
	sharedStrings, _ := readSharedStrings(r)

	// 2. 读取所有 worksheet
	sheetRe := regexp.MustCompile(`xl/worksheets/sheet\d+\.xml`)
	var sb strings.Builder
	sheetIdx := 1
	for _, f := range r.File {
		if !sheetRe.MatchString(f.Name) {
			continue
		}
		rows, err := readSheetRows(f, sharedStrings)
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n### Sheet %d\n\n", sheetIdx))
		sb.WriteString(csvToMarkdownTable(rows))
		sheetIdx++
	}
	return strings.TrimSpace(sb.String()), nil
}

// readSharedStrings 读取共享字符串池
func readSharedStrings(r *zip.ReadCloser) ([]string, error) {
	for _, f := range r.File {
		if f.Name != "xl/sharedStrings.xml" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, 5*1024*1024))
		rc.Close()
		if err != nil {
			return nil, err
		}
		return parseSharedStrings(string(data)), nil
	}
	return nil, nil
}

func parseSharedStrings(xmlStr string) []string {
	re := regexp.MustCompile(`<t[^>]*>([^<]*)</t>`)
	var result []string
	for _, m := range re.FindAllStringSubmatch(xmlStr, -1) {
		if len(m) > 1 {
			result = append(result, m[1])
		}
	}
	return result
}

// xlsxCell XLSX 单元格 XML 结构
type xlsxCell struct {
	R string `xml:"r,attr"`
	T string `xml:"t,attr"`
	V string `xml:"v"`
}

// xlsxRow XLSX 行 XML 结构
type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

// xlsxSheet XLSX Sheet XML 结构
type xlsxSheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

// readSheetRows 读取一个 Sheet 的所有行，返回 [][]string
func readSheetRows(f *zip.File, sharedStrings []string) ([][]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, 10*1024*1024))
	if err != nil {
		return nil, err
	}

	var sheet xlsxSheet
	if err := xml.Unmarshal(data, &sheet); err != nil {
		return nil, err
	}

	var rows [][]string
	for _, row := range sheet.Rows {
		var cells []string
		for _, cell := range row.Cells {
			val := cell.V
			// t="s" 表示共享字符串索引
			if cell.T == "s" && sharedStrings != nil {
				idx := 0
				fmt.Sscanf(cell.V, "%d", &idx)
				if idx >= 0 && idx < len(sharedStrings) {
					val = sharedStrings[idx]
				}
			}
			cells = append(cells, val)
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}
	return rows, nil
}
