package markdownchunker

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestAdvancedTableProcessor_WellFormedTable(t *testing.T) {
	markdown := `| Name | Age | City |
|------|-----|------|
| John | 25  | NYC  |
| Jane | 30  | LA   |`

	table := parseTableFromMarkdown(t, markdown)
	processor := NewAdvancedTableProcessor([]byte(markdown))
	info := processor.ProcessTable(table)

	if !info.IsWellFormed {
		t.Errorf("Expected well-formed table, got errors: %v", info.Errors)
	}

	if info.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", info.Rows)
	}

	if info.Columns != 3 {
		t.Errorf("Expected 3 columns, got %d", info.Columns)
	}

	if !info.HasHeader {
		t.Error("Expected table to have header")
	}

	expectedHeaders := []string{"Name", "Age", "City"}
	if len(info.HeaderCells) != len(expectedHeaders) {
		t.Errorf("Expected %d header cells, got %d", len(expectedHeaders), len(info.HeaderCells))
	}

	for i, expected := range expectedHeaders {
		if i < len(info.HeaderCells) && info.HeaderCells[i] != expected {
			t.Errorf("Expected header cell %d to be '%s', got '%s'", i, expected, info.HeaderCells[i])
		}
	}

	if len(info.DataRows) != 2 {
		t.Errorf("Expected 2 data rows, got %d", len(info.DataRows))
	}
}

func TestAdvancedTableProcessor_TableWithAlignments(t *testing.T) {
	markdown := `| Left | Center | Right |
|:-----|:------:|------:|
| L1   | C1     | R1    |
| L2   | C2     | R2    |`

	table := parseTableFromMarkdown(t, markdown)
	processor := NewAdvancedTableProcessor([]byte(markdown))
	info := processor.ProcessTable(table)

	expectedAlignments := []string{"left", "center", "right"}
	if len(info.Alignments) != len(expectedAlignments) {
		t.Errorf("Expected %d alignments, got %d", len(expectedAlignments), len(info.Alignments))
	}

	for i, expected := range expectedAlignments {
		if i < len(info.Alignments) && info.Alignments[i] != expected {
			t.Errorf("Expected alignment %d to be '%s', got '%s'", i, expected, info.Alignments[i])
		}
	}
}

func TestAdvancedTableProcessor_MalformedTable(t *testing.T) {
	markdown := `| Name | Age |
|------|-----|
| John | 25  | Extra |
| Jane |`

	table := parseTableFromMarkdown(t, markdown)
	processor := NewAdvancedTableProcessor([]byte(markdown))
	info := processor.ProcessTable(table)

	if info.IsWellFormed {
		t.Error("Expected malformed table to be detected")
	}

	if len(info.Errors) == 0 {
		t.Error("Expected errors to be reported for malformed table")
	}
}

func TestAdvancedTableProcessor_CellTypeDetection(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{"", "empty"},
		{"123", "integer"},
		{"12.34", "decimal"},
		{"2023-12-25", "date"},
		{"https://example.com", "url"},
		{"user@example.com", "email"},
		{"true", "boolean"},
		{"false", "boolean"},
		{"yes", "boolean"},
		{"no", "boolean"},
		{"Hello World", "text"},
	}

	processor := NewAdvancedTableProcessor([]byte(""))

	for _, test := range tests {
		result := processor.detectCellType(test.content)
		if result != test.expected {
			t.Errorf("detectCellType(%q) = %q, expected %q", test.content, result, test.expected)
		}
	}
}

func TestAdvancedTableProcessor_ComplexTable(t *testing.T) {
	markdown := `| Product | Price | Available | URL |
|---------|-------|-----------|-----|
| Widget  | 19.99 | true      | https://shop.com/widget |
| Gadget  | 29.50 | false     | https://shop.com/gadget |
| Tool    | 15.00 | yes       | https://shop.com/tool   |`

	table := parseTableFromMarkdown(t, markdown)
	processor := NewAdvancedTableProcessor([]byte(markdown))
	info := processor.ProcessTable(table)

	if !info.IsWellFormed {
		t.Errorf("Expected well-formed table, got errors: %v", info.Errors)
	}

	// 检查元数据
	metadata := info.GetTableMetadata()

	if metadata["rows"] != "4" {
		t.Errorf("Expected rows metadata to be '4', got '%s'", metadata["rows"])
	}

	if metadata["columns"] != "4" {
		t.Errorf("Expected columns metadata to be '4', got '%s'", metadata["columns"])
	}

	if metadata["has_header"] != "true" {
		t.Errorf("Expected has_header metadata to be 'true', got '%s'", metadata["has_header"])
	}

	if metadata["is_well_formed"] != "true" {
		t.Errorf("Expected is_well_formed metadata to be 'true', got '%s'", metadata["is_well_formed"])
	}

	// 检查单元格类型

	if !strings.Contains(metadata["cell_types"], "decimal") {
		t.Error("Expected cell_types to contain 'decimal'")
	}

	if !strings.Contains(metadata["cell_types"], "boolean") {
		t.Error("Expected cell_types to contain 'boolean'")
	}

	if !strings.Contains(metadata["cell_types"], "url") {
		t.Error("Expected cell_types to contain 'url'")
	}
}

func TestAdvancedTableProcessor_EmptyTable(t *testing.T) {
	markdown := `| Header |
|--------|`

	table := parseTableFromMarkdown(t, markdown)
	processor := NewAdvancedTableProcessor([]byte(markdown))
	info := processor.ProcessTable(table)

	if info.IsWellFormed {
		t.Error("Expected empty table to be marked as not well-formed")
	}

	if len(info.Errors) == 0 {
		t.Error("Expected errors for empty table")
	}

	// 应该包含"no data rows"错误
	hasDataRowError := false
	for _, err := range info.Errors {
		if strings.Contains(err, "no data rows") {
			hasDataRowError = true
			break
		}
	}

	if !hasDataRowError {
		t.Error("Expected 'no data rows' error")
	}
}

func TestTableProcessorIntegration(t *testing.T) {
	markdown := `| Name | Age | Email |
|------|-----|-------|
| John | 25  | john@example.com |
| Jane | 30  | jane@example.com |
| Bob  | invalid | not-email |`

	chunker := NewMarkdownChunker()
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if chunk.Type != "table" {
		t.Errorf("Expected chunk type 'table', got '%s'", chunk.Type)
	}

	// 检查增强的元数据
	if chunk.Metadata["rows"] != "4" {
		t.Errorf("Expected 4 rows, got %s", chunk.Metadata["rows"])
	}

	if chunk.Metadata["columns"] != "3" {
		t.Errorf("Expected 3 columns, got %s", chunk.Metadata["columns"])
	}

	if chunk.Metadata["has_header"] != "true" {
		t.Errorf("Expected has_header to be true, got %s", chunk.Metadata["has_header"])
	}

	// 检查单元格类型分析
	if cellTypes, exists := chunk.Metadata["cell_types"]; exists {
		if !strings.Contains(cellTypes, "email") {
			t.Error("Expected cell_types to contain 'email'")
		}
		if !strings.Contains(cellTypes, "integer") {
			t.Error("Expected cell_types to contain 'integer'")
		}
		if !strings.Contains(cellTypes, "text") {
			t.Error("Expected cell_types to contain 'text'")
		}
	} else {
		t.Error("Expected cell_types metadata to exist")
	}
}

// 辅助函数：从markdown解析表格
func parseTableFromMarkdown(t *testing.T, markdown string) *extast.Table {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	reader := text.NewReader([]byte(markdown))
	doc := md.Parser().Parse(reader)

	// 查找表格节点
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if table, ok := child.(*extast.Table); ok {
			return table
		}
	}

	t.Fatal("No table found in markdown")
	return nil
}
