package markdownchunker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// TableInfo 表格信息结构
type TableInfo struct {
	Rows         int
	Columns      int
	HasHeader    bool
	HeaderCells  []string
	DataRows     [][]string
	Alignments   []string          // left, center, right
	CellTypes    map[string]string // 单元格内容类型分析
	IsWellFormed bool
	Errors       []string
}

// AdvancedTableProcessor 高级表格处理器
type AdvancedTableProcessor struct {
	source []byte
}

// NewAdvancedTableProcessor 创建高级表格处理器
func NewAdvancedTableProcessor(source []byte) *AdvancedTableProcessor {
	return &AdvancedTableProcessor{
		source: source,
	}
}

// ProcessTable 处理表格并返回详细信息
func (p *AdvancedTableProcessor) ProcessTable(table *extast.Table) *TableInfo {
	info := &TableInfo{
		HeaderCells:  make([]string, 0),
		DataRows:     make([][]string, 0),
		Alignments:   make([]string, 0),
		CellTypes:    make(map[string]string),
		Errors:       make([]string, 0),
		IsWellFormed: true,
	}

	// 从AST分析表格结构
	p.analyzeTableStructure(table, info)

	// 从原始内容分析表格格式
	p.analyzeRawTableContent(info)

	// 分析单元格内容类型
	p.analyzeCellTypes(info)

	// 验证表格完整性
	p.validateTableIntegrity(info)

	return info
}

// analyzeTableStructure 分析表格AST结构
func (p *AdvancedTableProcessor) analyzeTableStructure(table *extast.Table, info *TableInfo) {
	dataRowCount := 0

	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch node := child.(type) {
		case *extast.TableHeader:
			// 处理表头
			var headerCells []string
			cellCount := 0

			for cell := node.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tableCell, ok := cell.(*extast.TableCell); ok {
					cellText := p.extractCellText(tableCell)
					headerCells = append(headerCells, cellText)
					cellCount++
				}
			}

			info.HeaderCells = headerCells
			info.HasHeader = true
			info.Columns = cellCount

		case *extast.TableRow:
			// 处理数据行
			var rowCells []string
			cellCount := 0

			for cell := node.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tableCell, ok := cell.(*extast.TableCell); ok {
					cellText := p.extractCellText(tableCell)
					rowCells = append(rowCells, cellText)
					cellCount++
				}
			}

			info.DataRows = append(info.DataRows, rowCells)
			dataRowCount++

			// 更新列数（以防表头列数不准确）
			if cellCount > info.Columns {
				info.Columns = cellCount
			}
		}
	}

	// 总行数 = 表头行数(1) + 数据行数
	info.Rows = 1 + dataRowCount
}

// analyzeRawTableContent 分析原始表格内容
func (p *AdvancedTableProcessor) analyzeRawTableContent(info *TableInfo) {
	// 从原始源码中提取表格内容
	lines := strings.Split(string(p.source), "\n")
	var tableLines []string
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			tableLines = append(tableLines, line)
			inTable = true
		} else if inTable {
			break
		}
	}

	if len(tableLines) == 0 {
		info.Errors = append(info.Errors, "No table content found in source")
		info.IsWellFormed = false
		return
	}

	// 分析对齐信息
	p.extractAlignments(tableLines, info)

	// 验证表格行的一致性
	p.validateRowConsistency(tableLines, info)
}

// extractAlignments 提取表格对齐信息
func (p *AdvancedTableProcessor) extractAlignments(tableLines []string, info *TableInfo) {
	// 查找分隔行（包含 --- 的行）
	for _, line := range tableLines {
		if strings.Contains(line, "---") {
			// 解析对齐信息
			cells := strings.Split(line, "|")
			for _, cell := range cells {
				cell = strings.TrimSpace(cell)
				if cell == "" {
					continue
				}

				if strings.HasPrefix(cell, ":") && strings.HasSuffix(cell, ":") {
					info.Alignments = append(info.Alignments, "center")
				} else if strings.HasSuffix(cell, ":") {
					info.Alignments = append(info.Alignments, "right")
				} else {
					info.Alignments = append(info.Alignments, "left")
				}
			}
			break
		}
	}
}

// validateRowConsistency 验证表格行的一致性
func (p *AdvancedTableProcessor) validateRowConsistency(tableLines []string, info *TableInfo) {
	expectedColumns := info.Columns

	for i, line := range tableLines {
		if strings.Contains(line, "---") {
			continue // 跳过分隔行
		}

		// 正确计算表格列数：去除首尾空白后按 | 分割，然后减去首尾的空元素
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
			continue // 不是标准表格行格式
		}

		// 去除首尾的 |
		line = strings.Trim(line, "|")
		cells := strings.Split(line, "|")
		actualColumns := len(cells)

		if actualColumns != expectedColumns {
			info.Errors = append(info.Errors,
				fmt.Sprintf("Row %d has %d columns, expected %d", i+1, actualColumns, expectedColumns))
			info.IsWellFormed = false
		}
	}
}

// analyzeCellTypes 分析单元格内容类型
func (p *AdvancedTableProcessor) analyzeCellTypes(info *TableInfo) {
	// 分析表头类型
	for i, header := range info.HeaderCells {
		cellType := p.detectCellType(header)
		info.CellTypes[fmt.Sprintf("header_%d", i)] = cellType
	}

	// 分析数据行类型
	for rowIdx, row := range info.DataRows {
		for colIdx, cell := range row {
			cellType := p.detectCellType(cell)
			info.CellTypes[fmt.Sprintf("data_%d_%d", rowIdx, colIdx)] = cellType
		}
	}
}

// detectCellType 检测单元格内容类型
func (p *AdvancedTableProcessor) detectCellType(content string) string {
	content = strings.TrimSpace(content)

	if content == "" {
		return "empty"
	}

	// 检测数字
	if matched, _ := regexp.MatchString(`^\d+$`, content); matched {
		return "integer"
	}

	if matched, _ := regexp.MatchString(`^\d+\.\d+$`, content); matched {
		return "decimal"
	}

	// 检测日期
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, content); matched {
		return "date"
	}

	// 检测URL
	if matched, _ := regexp.MatchString(`^https?://`, content); matched {
		return "url"
	}

	// 检测邮箱
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, content); matched {
		return "email"
	}

	// 检测布尔值
	lower := strings.ToLower(content)
	if lower == "true" || lower == "false" || lower == "yes" || lower == "no" {
		return "boolean"
	}

	return "text"
}

// validateTableIntegrity 验证表格完整性
func (p *AdvancedTableProcessor) validateTableIntegrity(info *TableInfo) {
	// 检查是否有表头
	if !info.HasHeader {
		info.Errors = append(info.Errors, "Table appears to have no header")
	}

	// 检查是否有数据行
	if len(info.DataRows) == 0 {
		info.Errors = append(info.Errors, "Table has no data rows")
		info.IsWellFormed = false
	}

	// 检查列数一致性
	if len(info.HeaderCells) != info.Columns {
		info.Errors = append(info.Errors,
			fmt.Sprintf("Header has %d cells but table has %d columns", len(info.HeaderCells), info.Columns))
		info.IsWellFormed = false
	}

	// 检查对齐信息
	if len(info.Alignments) > 0 && len(info.Alignments) != info.Columns {
		info.Errors = append(info.Errors,
			fmt.Sprintf("Alignment specification has %d entries but table has %d columns", len(info.Alignments), info.Columns))
	}
}

// extractCellText 提取单元格文本内容
func (p *AdvancedTableProcessor) extractCellText(cell *extast.TableCell) string {
	var text strings.Builder

	// 使用AST遍历提取所有文本节点，包括链接等
	err := ast.Walk(cell, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch node := n.(type) {
			case *ast.Text:
				text.Write(node.Segment.Value(p.source))
			case *ast.AutoLink:
				// 处理自动链接
				text.Write(node.URL(p.source))
			case *ast.Link:
				// 对于普通链接，我们需要提取URL
				text.Write(node.Destination)
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return ""
	}

	return strings.TrimSpace(text.String())
}

// GetTableMetadata 获取表格元数据
func (info *TableInfo) GetTableMetadata() map[string]string {
	metadata := map[string]string{
		"rows":           fmt.Sprintf("%d", info.Rows),
		"columns":        fmt.Sprintf("%d", info.Columns),
		"has_header":     fmt.Sprintf("%t", info.HasHeader),
		"is_well_formed": fmt.Sprintf("%t", info.IsWellFormed),
		"error_count":    fmt.Sprintf("%d", len(info.Errors)),
	}

	if len(info.Alignments) > 0 {
		metadata["alignments"] = strings.Join(info.Alignments, ",")
	}

	if len(info.Errors) > 0 {
		metadata["errors"] = strings.Join(info.Errors, "; ")
	}

	// 添加列类型统计
	typeCount := make(map[string]int)
	for _, cellType := range info.CellTypes {
		typeCount[cellType]++
	}

	var types []string
	for cellType, count := range typeCount {
		types = append(types, fmt.Sprintf("%s:%d", cellType, count))
	}
	if len(types) > 0 {
		metadata["cell_types"] = strings.Join(types, ",")
	}

	return metadata
}
