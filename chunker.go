package markdownchunker

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// Chunk 表示分块后的内容
type Chunk struct {
	ID       int               `json:"id"`
	Type     string            `json:"type"`    // heading, paragraph, table, code, list
	Content  string            `json:"content"` // 原始 markdown 内容
	Text     string            `json:"text"`    // 纯文本内容，用于向量化
	Level    int               `json:"level"`   // 标题层级 (仅对 heading 有效)
	Metadata map[string]string `json:"metadata"`
}

// ErrorHandlingMode 错误处理模式
type ErrorHandlingMode int

const (
	// ErrorModeStrict 严格模式，遇到错误立即返回
	ErrorModeStrict ErrorHandlingMode = iota
	// ErrorModePermissive 宽松模式，记录错误但继续处理
	ErrorModePermissive
	// ErrorModeSilent 静默模式，忽略错误
	ErrorModeSilent
)

// PerformanceMode 性能模式
type PerformanceMode int

const (
	// PerformanceModeDefault 默认性能模式
	PerformanceModeDefault PerformanceMode = iota
	// PerformanceModeMemoryOptimized 内存优化模式
	PerformanceModeMemoryOptimized
	// PerformanceModeSpeedOptimized 速度优化模式
	PerformanceModeSpeedOptimized
)

// MetadataExtractor 元数据提取器接口
type MetadataExtractor interface {
	// Extract 从AST节点中提取元数据
	Extract(node ast.Node, source []byte) map[string]string
	// SupportedTypes 返回支持的内容类型
	SupportedTypes() []string
}

// ChunkerConfig 分块器配置
type ChunkerConfig struct {
	// MaxChunkSize 最大块大小（字符数），0表示无限制
	MaxChunkSize int

	// EnabledTypes 启用的内容类型，nil表示启用所有类型
	EnabledTypes map[string]bool

	// CustomExtractors 自定义元数据提取器
	CustomExtractors []MetadataExtractor

	// ErrorHandling 错误处理模式
	ErrorHandling ErrorHandlingMode

	// PerformanceMode 性能模式
	PerformanceMode PerformanceMode

	// FilterEmptyChunks 是否过滤空块
	FilterEmptyChunks bool

	// PreserveWhitespace 是否保留空白字符
	PreserveWhitespace bool

	// MemoryLimit 内存使用限制（字节），0表示无限制
	MemoryLimit int64

	// EnableObjectPooling 是否启用对象池化
	EnableObjectPooling bool
}

// MarkdownChunker Markdown 分块器
type MarkdownChunker struct {
	md                 goldmark.Markdown
	config             *ChunkerConfig
	errorHandler       ErrorHandler
	performanceMonitor *PerformanceMonitor
	memoryOptimizer    *MemoryOptimizer
	stringOps          *OptimizedStringOperations
	chunks             []Chunk
	source             []byte
}

// DefaultConfig 返回默认配置
func DefaultConfig() *ChunkerConfig {
	return &ChunkerConfig{
		MaxChunkSize:       0,   // 无限制
		EnabledTypes:       nil, // 启用所有类型
		CustomExtractors:   []MetadataExtractor{},
		ErrorHandling:      ErrorModePermissive,
		PerformanceMode:    PerformanceModeDefault,
		FilterEmptyChunks:  true,
		PreserveWhitespace: false,
	}
}

// ValidateConfig 验证配置的有效性
func ValidateConfig(config *ChunkerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.MaxChunkSize < 0 {
		return fmt.Errorf("MaxChunkSize cannot be negative")
	}

	// 验证启用的类型
	if config.EnabledTypes != nil {
		validTypes := map[string]bool{
			"heading": true, "paragraph": true, "code": true,
			"table": true, "list": true, "blockquote": true,
			"thematic_break": true,
		}

		for typeName := range config.EnabledTypes {
			if !validTypes[typeName] {
				return fmt.Errorf("invalid content type: %s", typeName)
			}
		}
	}

	return nil
}

// isTypeEnabled 检查指定类型是否启用
func (c *MarkdownChunker) isTypeEnabled(chunkType string) bool {
	if c.config.EnabledTypes == nil {
		return true // 如果没有指定，则启用所有类型
	}

	enabled, exists := c.config.EnabledTypes[chunkType]
	return exists && enabled
}

// NewMarkdownChunker 创建新的分块器，使用默认配置
func NewMarkdownChunker() *MarkdownChunker {
	return NewMarkdownChunkerWithConfig(DefaultConfig())
}

// NewMarkdownChunkerWithConfig 使用指定配置创建新的分块器
func NewMarkdownChunkerWithConfig(config *ChunkerConfig) *MarkdownChunker {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证配置
	if err := ValidateConfig(config); err != nil {
		// 如果配置无效，使用默认配置
		config = DefaultConfig()
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown (包含表格支持)
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	return &MarkdownChunker{
		md:                 md,
		config:             config,
		errorHandler:       NewDefaultErrorHandler(config.ErrorHandling),
		performanceMonitor: NewPerformanceMonitor(),
		chunks:             []Chunk{},
	}
}

// ChunkDocument 对整个文档进行分块
func (c *MarkdownChunker) ChunkDocument(content []byte) ([]Chunk, error) {
	// 开始性能监控
	c.performanceMonitor.Start()
	defer c.performanceMonitor.Stop()

	// 记录输入文档大小
	c.performanceMonitor.RecordBytes(int64(len(content)))

	// 清除之前的错误
	c.errorHandler.ClearErrors()

	// 输入验证
	if content == nil {
		err := NewChunkerError(ErrorTypeInvalidInput, "content cannot be nil", nil)
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
		return []Chunk{}, nil
	}

	// 检查内容大小
	if len(content) > 100*1024*1024 { // 100MB 限制
		err := NewChunkerError(ErrorTypeMemoryExhausted, "content too large", nil).
			WithContext("size", len(content)).
			WithContext("limit", 100*1024*1024)
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
	}

	c.source = content
	c.chunks = []Chunk{}

	// 解析 Markdown
	reader := text.NewReader(content)
	doc := c.md.Parser().Parse(reader)

	// 遍历顶层节点进行分块
	chunkID := 0
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		chunk := c.processNode(child, chunkID)
		if chunk != nil {
			// 检查类型是否启用
			if !c.isTypeEnabled(chunk.Type) {
				continue
			}

			// 检查是否过滤空块
			if c.config.FilterEmptyChunks && strings.TrimSpace(chunk.Text) == "" {
				continue
			}

			// 检查块大小限制
			if c.config.MaxChunkSize > 0 && len(chunk.Content) > c.config.MaxChunkSize {
				err := NewChunkerError(ErrorTypeChunkTooLarge, "chunk size exceeds maximum", nil).
					WithContext("chunk_size", len(chunk.Content)).
					WithContext("max_size", c.config.MaxChunkSize).
					WithContext("chunk_type", chunk.Type).
					WithContext("chunk_id", chunk.ID)

				if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
					return nil, handlerErr
				}

				// 在宽松模式下截断内容
				if c.config.ErrorHandling != ErrorModeStrict {
					chunk.Content = chunk.Content[:c.config.MaxChunkSize]
					chunk.Text = chunk.Text[:min(len(chunk.Text), c.config.MaxChunkSize)]
				}
			}

			// 应用自定义元数据提取器
			for _, extractor := range c.config.CustomExtractors {
				supportedTypes := extractor.SupportedTypes()
				if len(supportedTypes) == 0 || slices.Contains(supportedTypes, chunk.Type) {
					maps.Copy(chunk.Metadata, extractor.Extract(child, c.source))
				}
			}

			c.chunks = append(c.chunks, *chunk)

			// 记录处理的块
			c.performanceMonitor.RecordChunk(chunk)

			chunkID++
		}
	}

	return c.chunks, nil
}

// GetErrors 获取处理过程中的所有错误
func (c *MarkdownChunker) GetErrors() []*ChunkerError {
	return c.errorHandler.GetErrors()
}

// HasErrors 检查是否有错误
func (c *MarkdownChunker) HasErrors() bool {
	return c.errorHandler.HasErrors()
}

// ClearErrors 清除所有错误
func (c *MarkdownChunker) ClearErrors() {
	c.errorHandler.ClearErrors()
}

// GetErrorsByType 按类型获取错误
func (c *MarkdownChunker) GetErrorsByType(errorType ErrorType) []*ChunkerError {
	if handler, ok := c.errorHandler.(*DefaultErrorHandler); ok {
		return handler.GetErrorsByType(errorType)
	}
	// 如果不是默认处理器，遍历所有错误
	var filtered []*ChunkerError
	for _, err := range c.errorHandler.GetErrors() {
		if err.Type == errorType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetPerformanceStats 获取性能统计信息
func (c *MarkdownChunker) GetPerformanceStats() PerformanceStats {
	return c.performanceMonitor.GetStats()
}

// GetPerformanceMonitor 获取性能监控器（用于高级用法）
func (c *MarkdownChunker) GetPerformanceMonitor() *PerformanceMonitor {
	return c.performanceMonitor
}

// ResetPerformanceMonitor 重置性能监控器
func (c *MarkdownChunker) ResetPerformanceMonitor() {
	c.performanceMonitor.Reset()
}

// processNode 处理单个 AST 节点
func (c *MarkdownChunker) processNode(node ast.Node, id int) *Chunk {
	switch n := node.(type) {
	case *ast.Heading:
		return c.processHeading(n, id)
	case *ast.Paragraph:
		return c.processParagraph(n, id)
	case *ast.FencedCodeBlock:
		return c.processCodeBlock(n, id)
	case *ast.CodeBlock:
		return c.processCodeBlock(n, id)
	case *extast.Table:
		return c.processTable(n, id)
	case *ast.List:
		return c.processList(n, id)
	case *ast.Blockquote:
		return c.processBlockquote(n, id)
	case *ast.ThematicBreak:
		return c.processThematicBreak(n, id)
	default:
		return nil
	}
}

// processHeading 处理标题
func (c *MarkdownChunker) processHeading(heading *ast.Heading, id int) *Chunk {
	content := c.getNodeRawContent(heading)
	text := c.getNodeText(heading)

	return &Chunk{
		ID:      id,
		Type:    "heading",
		Content: content,
		Text:    text,
		Level:   heading.Level,
		Metadata: map[string]string{
			"heading_level": fmt.Sprintf("%d", heading.Level),
		},
	}
}

// processParagraph 处理段落
func (c *MarkdownChunker) processParagraph(para *ast.Paragraph, id int) *Chunk {
	content := c.getNodeRawContent(para)
	text := c.getNodeText(para)

	// 过滤掉空段落
	if strings.TrimSpace(text) == "" {
		return nil
	}

	return &Chunk{
		ID:      id,
		Type:    "paragraph",
		Content: content,
		Text:    text,
		Level:   0,
		Metadata: map[string]string{
			"word_count": fmt.Sprintf("%d", len(strings.Fields(text))),
		},
	}
}

// processCodeBlock 处理代码块
func (c *MarkdownChunker) processCodeBlock(code ast.Node, id int) *Chunk {
	var language string
	content := c.getNodeRawContent(code)

	// 提取代码块的纯文本内容，去除尾部空行
	var codeLines []string
	for i := 0; i < code.Lines().Len(); i++ {
		line := code.Lines().At(i)
		codeLines = append(codeLines, string(line.Value(c.source)))
	}

	// 去除尾部的空行
	for len(codeLines) > 0 && strings.TrimSpace(codeLines[len(codeLines)-1]) == "" {
		codeLines = codeLines[:len(codeLines)-1]
	}

	var textBuf bytes.Buffer
	for i, line := range codeLines {
		textBuf.WriteString(strings.TrimRight(line, "\n"))
		if i < len(codeLines)-1 {
			textBuf.WriteByte('\n')
		}
	}
	text := textBuf.String()

	// 获取代码语言
	if fenced, ok := code.(*ast.FencedCodeBlock); ok {
		if fenced.Info != nil {
			language = strings.TrimSpace(string(fenced.Info.Segment.Value(c.source)))
		}
	}

	// 计算行数（使用清理后的行数）
	lineCount := len(codeLines)

	return &Chunk{
		ID:      id,
		Type:    "code",
		Content: content,
		Text:    text,
		Level:   0,
		Metadata: map[string]string{
			"language":   language,
			"line_count": fmt.Sprintf("%d", lineCount),
		},
	}
}

// processTable 处理表格
func (c *MarkdownChunker) processTable(table *extast.Table, id int) *Chunk {
	content := c.getNodeRawContent(table)
	text := c.getNodeText(table)

	// 使用高级表格处理器分析表格
	processor := NewAdvancedTableProcessor(c.source)
	tableInfo := processor.ProcessTable(table)

	// 获取基础元数据
	metadata := tableInfo.GetTableMetadata()

	// 如果表格格式有问题，记录错误
	if !tableInfo.IsWellFormed && len(tableInfo.Errors) > 0 {
		err := NewChunkerError(ErrorTypeParsingFailed, "table format issues detected", nil).
			WithContext("table_errors", strings.Join(tableInfo.Errors, "; ")).
			WithContext("chunk_id", id)
		c.errorHandler.HandleError(err)
	}

	return &Chunk{
		ID:       id,
		Type:     "table",
		Content:  content,
		Text:     text,
		Level:    0,
		Metadata: metadata,
	}
}

// processList 处理列表
func (c *MarkdownChunker) processList(list *ast.List, id int) *Chunk {
	content := c.getNodeRawContent(list)
	text := c.getListText(list)

	listType := "unordered"
	if list.IsOrdered() {
		listType = "ordered"
	}

	// 计算列表项数量
	itemCount := 0
	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*ast.ListItem); ok {
			itemCount++
		}
	}

	return &Chunk{
		ID:      id,
		Type:    "list",
		Content: content,
		Text:    text,
		Level:   0,
		Metadata: map[string]string{
			"list_type":  listType,
			"item_count": fmt.Sprintf("%d", itemCount),
		},
	}
}

// getListText 获取列表的纯文本内容，确保项目之间有空格
func (c *MarkdownChunker) getListText(list *ast.List) string {
	var items []string

	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if listItem, ok := child.(*ast.ListItem); ok {
			text := c.getNodeText(listItem)
			items = append(items, text)
		}
	}

	return strings.Join(items, " ")
}

// processBlockquote 处理引用块
func (c *MarkdownChunker) processBlockquote(quote *ast.Blockquote, id int) *Chunk {
	content := c.getNodeRawContent(quote)
	text := c.getNodeText(quote)

	return &Chunk{
		ID:      id,
		Type:    "blockquote",
		Content: content,
		Text:    text,
		Level:   0,
		Metadata: map[string]string{
			"word_count": fmt.Sprintf("%d", len(strings.Fields(text))),
		},
	}
}

// processThematicBreak 处理分隔线
func (c *MarkdownChunker) processThematicBreak(hr *ast.ThematicBreak, id int) *Chunk {
	content := c.getNodeRawContent(hr)

	return &Chunk{
		ID:      id,
		Type:    "thematic_break",
		Content: content,
		Text:    "---",
		Level:   0,
		Metadata: map[string]string{
			"type": "horizontal_rule",
		},
	}
}

// getNodeRawContent 获取节点的原始 markdown 内容
func (c *MarkdownChunker) getNodeRawContent(node ast.Node) string {
	// 特殊处理某些节点类型
	switch n := node.(type) {
	case *ast.ThematicBreak:
		return "---"
	case *ast.Heading:
		// 对于标题，确保包含 # 符号
		text := c.getNodeText(n)
		prefix := strings.Repeat("#", n.Level) + " "
		return prefix + text
	case *ast.FencedCodeBlock:
		// 处理围栏代码块
		var buf bytes.Buffer

		// 添加开始的围栏
		buf.WriteString("```")
		if n.Info != nil {
			buf.Write(n.Info.Segment.Value(c.source))
		}
		buf.WriteByte('\n')

		// 添加代码内容，去除尾部空行
		var codeLines []string
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			codeLines = append(codeLines, string(line.Value(c.source)))
		}

		// 去除尾部的空行
		for len(codeLines) > 0 && strings.TrimSpace(codeLines[len(codeLines)-1]) == "" {
			codeLines = codeLines[:len(codeLines)-1]
		}

		for i, line := range codeLines {
			buf.WriteString(strings.TrimRight(line, "\n"))
			if i < len(codeLines)-1 {
				buf.WriteByte('\n')
			}
		}

		// 添加结束的围栏
		buf.WriteString("\n```")
		return buf.String()
	case *ast.CodeBlock:
		// 处理缩进代码块
		var buf bytes.Buffer
		var codeLines []string
		for i := 0; i < n.Lines().Len(); i++ {
			line := n.Lines().At(i)
			codeLines = append(codeLines, string(line.Value(c.source)))
		}

		// 去除尾部的空行
		for len(codeLines) > 0 && strings.TrimSpace(codeLines[len(codeLines)-1]) == "" {
			codeLines = codeLines[:len(codeLines)-1]
		}

		for i, line := range codeLines {
			buf.WriteString("    ") // 添加4个空格的缩进
			buf.WriteString(strings.TrimRight(line, "\n"))
			if i < len(codeLines)-1 {
				buf.WriteByte('\n')
			}
		}
		return buf.String()
	case *ast.List:
		// 处理列表
		return c.reconstructList(n)
	case *ast.Blockquote:
		// 处理引用块，需要从子节点重构
		return c.reconstructBlockquote(n)
	case *extast.Table:
		// 处理表格，需要从子节点重构
		return c.reconstructTable(n)
	}

	// 对于有Lines的节点，直接提取原始内容
	if node.Lines().Len() > 0 {
		var buf bytes.Buffer
		for i := 0; i < node.Lines().Len(); i++ {
			line := node.Lines().At(i)
			buf.Write(line.Value(c.source))
			// 保持原始换行符，除了最后一行
			if i < node.Lines().Len()-1 {
				buf.WriteByte('\n')
			}
		}
		return strings.TrimRight(buf.String(), "\n")
	}

	return ""
}

// reconstructList 重构列表的原始markdown
func (c *MarkdownChunker) reconstructList(list *ast.List) string {
	var buf bytes.Buffer
	itemIndex := 1

	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if listItem, ok := child.(*ast.ListItem); ok {
			// 添加列表标记
			if list.IsOrdered() {
				buf.WriteString(fmt.Sprintf("%d. ", itemIndex))
				itemIndex++
			} else {
				buf.WriteString("- ")
			}

			// 添加列表项内容
			text := c.getNodeText(listItem)
			buf.WriteString(text)

			// 如果不是最后一项，添加换行
			if child.NextSibling() != nil {
				buf.WriteByte('\n')
			}
		}
	}

	return buf.String()
}

// reconstructBlockquote 重构引用块的原始markdown
func (c *MarkdownChunker) reconstructBlockquote(quote *ast.Blockquote) string {
	var buf bytes.Buffer

	// 遍历blockquote的子节点（通常是段落）
	for child := quote.FirstChild(); child != nil; child = child.NextSibling() {
		if para, ok := child.(*ast.Paragraph); ok {
			// 从段落的每一行重构blockquote
			for i := 0; i < para.Lines().Len(); i++ {
				line := para.Lines().At(i)
				buf.WriteString("> ")
				// 去除行尾的换行符，因为我们会自己添加
				lineContent := strings.TrimRight(string(line.Value(c.source)), "\n")
				buf.WriteString(lineContent)
				if i < para.Lines().Len()-1 {
					buf.WriteByte('\n')
				}
			}
		}
	}

	return buf.String()
}

// reconstructTable 重构表格的原始markdown
func (c *MarkdownChunker) reconstructTable(table *extast.Table) string {
	var buf bytes.Buffer
	isFirstRow := true

	// 遍历表格的所有行
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		if tableRow, ok := child.(*extast.TableRow); ok {
			buf.WriteString("|")

			// 遍历行中的所有单元格
			for cell := tableRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tableCell, ok := cell.(*extast.TableCell); ok {
					buf.WriteString(" ")
					cellText := c.getNodeText(tableCell)
					buf.WriteString(cellText)
					buf.WriteString(" |")
				}
			}

			// 如果不是最后一行，添加换行
			if child.NextSibling() != nil {
				buf.WriteByte('\n')

				// 如果这是第一行（表头），添加分隔行
				if isFirstRow {
					buf.WriteString("|")
					for cell := tableRow.FirstChild(); cell != nil; cell = cell.NextSibling() {
						if _, ok := cell.(*extast.TableCell); ok {
							buf.WriteString("------|")
						}
					}
					if child.NextSibling() != nil {
						buf.WriteByte('\n')
					}
					isFirstRow = false
				}
			}
		}
	}

	// 总是从原始源码中提取表格内容，因为AST重构可能不完整
	// 从原始源码中提取表格内容
	lines := strings.Split(string(c.source), "\n")
	var tableLines []string
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			tableLines = append(tableLines, line)
			inTable = true
		} else if inTable {
			// 如果我们已经在表格中，但遇到了非表格行，表格结束
			break
		}
	}

	if len(tableLines) > 0 {
		return strings.Join(tableLines, "\n")
	}

	// 如果从源码提取失败，使用AST重构的结果
	result := buf.String()

	return result
}

// getNodeText 获取节点的纯文本内容
func (c *MarkdownChunker) getNodeText(node ast.Node) string {
	var buf bytes.Buffer

	// 遍历所有子节点提取文本
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindText:
				text := n.(*ast.Text)
				segment := text.Segment
				// 安全地获取文本内容，避免越界
				if segment.Start < len(c.source) && segment.Stop <= len(c.source) && segment.Start <= segment.Stop {
					buf.Write(segment.Value(c.source))
				}
				if text.HardLineBreak() || text.SoftLineBreak() {
					buf.WriteString(" ")
				}
			case ast.KindAutoLink:
				autolink := n.(*ast.AutoLink)
				segment := autolink.URL(c.source)
				buf.Write(segment)
			case ast.KindCodeSpan:
				codeSpan := n.(*ast.CodeSpan)
				// 处理代码段的文本内容
				if codeSpan.HasChildren() {
					// 递归处理子节点
				} else {
					// 直接从子节点获取文本
					for child := codeSpan.FirstChild(); child != nil; child = child.NextSibling() {
						if textNode, ok := child.(*ast.Text); ok {
							buf.Write(textNode.Segment.Value(c.source))
						}
					}
				}
			case ast.KindEmphasis:
				// 强调标记本身不添加文本，只处理其子节点
			case ast.KindLink:
				// 链接的文本由其子节点处理
			case ast.KindImage:
				// 图片显示alt文本
				img := n.(*ast.Image)
				if img.Title != nil {
					buf.Write(img.Title)
				}
			}
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return ""
	}

	// 清理多余的空格
	text := strings.TrimSpace(buf.String())
	// 将多个连续空格替换为单个空格
	text = strings.Join(strings.Fields(text), " ")

	return text
}
