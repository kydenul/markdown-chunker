package markdownchunker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/kydenul/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// ChunkPosition 表示块在文档中的位置
type ChunkPosition struct {
	StartLine int `json:"start_line"` // 起始行号（从1开始）
	EndLine   int `json:"end_line"`   // 结束行号（从1开始）
	StartCol  int `json:"start_col"`  // 起始列号（从1开始）
	EndCol    int `json:"end_col"`    // 结束列号（从1开始）
}

// Link 表示链接信息
type Link struct {
	Text string `json:"text"` // 链接文本
	URL  string `json:"url"`  // 链接地址
	Type string `json:"type"` // 链接类型：internal, external, anchor
}

// Image 表示图片信息
type Image struct {
	Alt    string `json:"alt"`    // 替代文本
	URL    string `json:"url"`    // 图片地址
	Title  string `json:"title"`  // 图片标题
	Width  string `json:"width"`  // 图片宽度
	Height string `json:"height"` // 图片高度
}

// Chunk 表示分块后的内容
type Chunk struct {
	ID       int               `json:"id"`
	Type     string            `json:"type"`    // heading, paragraph, table, code, list
	Content  string            `json:"content"` // 原始 markdown 内容
	Text     string            `json:"text"`    // 纯文本内容，用于向量化
	Level    int               `json:"level"`   // 标题层级 (仅对 heading 有效)
	Metadata map[string]string `json:"metadata"`

	// 新增字段
	Position ChunkPosition `json:"position"` // 在文档中的位置
	Links    []Link        `json:"links"`    // 包含的链接
	Images   []Image       `json:"images"`   // 包含的图片
	Hash     string        `json:"hash"`     // 内容哈希，用于去重
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

	// 日志配置
	LogLevel     string `json:"log_level"`     // DEBUG, INFO, WARN, ERROR
	EnableLog    bool   `json:"enable_log"`    // 是否启用日志
	LogFormat    string `json:"log_format"`    // 日志格式 (json, console)
	LogDirectory string `json:"log_directory"` // 日志文件目录
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
	logger             log.Logger // 日志器实例
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
		LogLevel:           "INFO",
		EnableLog:          true,
		LogFormat:          "console",
		LogDirectory:       "./logs", // 默认日志目录
	}
}

// parseLogLevel 解析日志级别字符串为 kydenul/log 支持的级别
func parseLogLevel(level string) string {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return "debug"
	case "INFO":
		return "info"
	case "WARN", "WARNING":
		return "warn"
	case "ERROR":
		return "error"
	default:
		return "info" // 默认为info级别
	}
}

// validateLogConfig 验证日志配置的有效性
func validateLogConfig(config *ChunkerConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证日志级别
	validLevels := map[string]bool{
		"DEBUG": true, "INFO": true, "WARN": true, "WARNING": true, "ERROR": true,
	}
	if config.LogLevel != "" && !validLevels[strings.ToUpper(config.LogLevel)] {
		return fmt.Errorf("invalid log level: %s, must be one of DEBUG, INFO, WARN, ERROR", config.LogLevel)
	}

	// 验证日志格式
	validFormats := map[string]bool{
		"json": true, "console": true,
	}
	if config.LogFormat != "" && !validFormats[strings.ToLower(config.LogFormat)] {
		return fmt.Errorf("invalid log format: %s, must be one of json, console", config.LogFormat)
	}

	// 验证日志目录（如果为空，将使用默认值）
	if config.LogDirectory != "" {
		// 这里可以添加更多的目录验证逻辑，比如检查目录是否可写，但为了保持简单，只检查不为空
		if strings.TrimSpace(config.LogDirectory) == "" {
			return fmt.Errorf("log directory cannot be empty or whitespace only")
		}
	}

	return nil
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

	// 验证日志配置
	if err := validateLogConfig(config); err != nil {
		return err
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

	// 初始化日志器
	// 设置默认格式如果为空
	logFormat := strings.ToLower(config.LogFormat)
	if logFormat == "" {
		logFormat = "console"
	}

	// 设置默认日志目录如果为空
	logDirectory := config.LogDirectory
	if logDirectory == "" {
		logDirectory = "./logs"
	}

	opts := &log.Options{
		Level:      parseLogLevel(config.LogLevel),
		Format:     logFormat,
		Directory:  logDirectory,              // 使用配置中的日志目录
		TimeLayout: "2006-01-02 15:04:05.000", // 设置时间格式
		MaxSize:    100,                       // 设置最大文件大小 (MB)
		MaxBackups: 3,                         // 设置最大备份文件数
	}

	if !config.EnableLog {
		opts.Level = "error" // 只记录错误
	}

	logger := log.NewLogger(opts)

	return &MarkdownChunker{
		md:                 md,
		config:             config,
		errorHandler:       NewDefaultErrorHandler(config.ErrorHandling),
		performanceMonitor: NewPerformanceMonitor(),
		chunks:             []Chunk{},
		logger:             logger,
	}
}

// ChunkDocument 对整个文档进行分块
func (c *MarkdownChunker) ChunkDocument(content []byte) ([]Chunk, error) {
	// 记录文档处理开始日志
	c.logger.Infow("开始处理 Markdown 文档",
		"document_size_bytes", len(content),
		"function", "ChunkDocument")

	// 开始性能监控
	c.performanceMonitor.Start()
	defer func() {
		c.performanceMonitor.Stop()

		// 获取性能统计信息
		stats := c.performanceMonitor.GetStats()

		// 记录文档处理结束日志，包含性能信息
		c.logger.Infow("完成 Markdown 文档处理",
			"chunk_count", len(c.chunks),
			"document_size_bytes", len(content),
			"processing_time_ms", stats.ProcessingTime.Milliseconds(),
			"memory_used_bytes", stats.MemoryUsed,
			"function", "ChunkDocument")
	}()

	// 记录输入文档大小
	c.performanceMonitor.RecordBytes(int64(len(content)))

	// 清除之前的错误
	c.errorHandler.ClearErrors()

	// 输入验证
	if content == nil {
		c.logger.Errorw("输入内容为空",
			"error", "content cannot be nil",
			"function", "ChunkDocument")

		err := NewChunkerError(ErrorTypeInvalidInput, "content cannot be nil", nil)
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
		return []Chunk{}, nil
	}

	// 记录空文档处理
	if len(content) == 0 {
		c.logger.Infow("处理空文档",
			"function", "ChunkDocument")
		return []Chunk{}, nil
	}

	// 检查内容大小
	if len(content) > 100*1024*1024 { // 100MB 限制
		c.logger.Errorw("文档大小超过限制",
			"document_size_bytes", len(content),
			"size_limit_bytes", 100*1024*1024,
			"function", "ChunkDocument")

		err := NewChunkerError(ErrorTypeMemoryExhausted, "content too large", nil).
			WithContext("size", len(content)).
			WithContext("limit", 100*1024*1024)
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
	}

	// 记录大型文档处理警告
	if len(content) > 10*1024*1024 { // 10MB 警告阈值
		c.logger.Warnw("处理大型文档",
			"document_size_bytes", len(content),
			"recommendation", "考虑分批处理以优化性能",
			"function", "ChunkDocument")
	}

	c.source = content
	c.chunks = []Chunk{}

	// 解析 Markdown
	c.logger.Debugw("开始解析 Markdown AST",
		"function", "ChunkDocument")

	reader := text.NewReader(content)
	doc := c.md.Parser().Parse(reader)

	c.logger.Debugw("Markdown AST 解析完成",
		"function", "ChunkDocument")

	// 遍历顶层节点进行分块
	c.logger.Debugw("开始遍历 AST 节点进行分块",
		"function", "ChunkDocument")

	chunkID := 0
	processedNodes := 0

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		processedNodes++
		chunk := c.processNode(child, chunkID)
		if chunk != nil {
			// 检查类型是否启用
			if !c.isTypeEnabled(chunk.Type) {
				c.logger.Debugw("跳过未启用的块类型",
					"chunk_type", chunk.Type,
					"chunk_id", chunkID,
					"function", "ChunkDocument")
				continue
			}

			// 检查是否过滤空块
			if c.config.FilterEmptyChunks && strings.TrimSpace(chunk.Text) == "" {
				c.logger.Debugw("过滤空块",
					"chunk_type", chunk.Type,
					"chunk_id", chunkID,
					"function", "ChunkDocument")
				continue
			}

			// 检查块大小限制
			if c.config.MaxChunkSize > 0 && len(chunk.Content) > c.config.MaxChunkSize {
				c.logger.Warnw("块大小超过限制",
					"chunk_size", len(chunk.Content),
					"max_size", c.config.MaxChunkSize,
					"chunk_type", chunk.Type,
					"chunk_id", chunk.ID,
					"function", "ChunkDocument")

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
					c.logger.Infow("截断超大块内容",
						"original_size", len(chunk.Content),
						"truncated_size", c.config.MaxChunkSize,
						"chunk_type", chunk.Type,
						"chunk_id", chunk.ID,
						"function", "ChunkDocument")

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

			// 记录成功处理的块
			c.logger.Debugw("成功处理块",
				"chunk_type", chunk.Type,
				"chunk_id", chunk.ID,
				"content_size", len(chunk.Content),
				"text_size", len(chunk.Text),
				"function", "ChunkDocument")

			chunkID++
		}

		// 每处理100个节点记录一次进度（用于大型文档）
		if processedNodes%100 == 0 && len(content) > 1024*1024 { // 只对大于1MB的文档记录进度
			c.logger.Infow("文档处理进度",
				"processed_nodes", processedNodes,
				"generated_chunks", len(c.chunks),
				"function", "ChunkDocument")
		}
	}

	c.logger.Debugw("完成 AST 节点遍历",
		"total_processed_nodes", processedNodes,
		"generated_chunks", len(c.chunks),
		"function", "ChunkDocument")

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
	position := c.calculatePosition(heading)
	links := c.extractLinks(heading)
	images := c.extractImages(heading)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "heading",
		Content:  content,
		Text:     text,
		Level:    heading.Level,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
		Metadata: map[string]string{
			"heading_level": fmt.Sprintf("%d", heading.Level),
			"level":         fmt.Sprintf("%d", heading.Level), // 为了兼容性
			"word_count":    fmt.Sprintf("%d", len(strings.Fields(text))),
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

	position := c.calculatePosition(para)
	links := c.extractLinks(para)
	images := c.extractImages(para)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "paragraph",
		Content:  content,
		Text:     text,
		Level:    0,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
		Metadata: map[string]string{
			"word_count": fmt.Sprintf("%d", len(strings.Fields(text))),
			"char_count": fmt.Sprintf("%d", len(text)),
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

	position := c.calculatePosition(code)
	links := c.extractLinks(code)
	images := c.extractImages(code)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "code",
		Content:  content,
		Text:     text,
		Level:    0,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
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

	position := c.calculatePosition(table)
	links := c.extractLinks(table)
	images := c.extractImages(table)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "table",
		Content:  content,
		Text:     text,
		Level:    0,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
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

	position := c.calculatePosition(list)
	links := c.extractLinks(list)
	images := c.extractImages(list)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "list",
		Content:  content,
		Text:     text,
		Level:    0,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
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

	position := c.calculatePosition(quote)
	links := c.extractLinks(quote)
	images := c.extractImages(quote)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "blockquote",
		Content:  content,
		Text:     text,
		Level:    0,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
		Metadata: map[string]string{
			"word_count": fmt.Sprintf("%d", len(strings.Fields(text))),
		},
	}
}

// processThematicBreak 处理分隔线
func (c *MarkdownChunker) processThematicBreak(hr *ast.ThematicBreak, id int) *Chunk {
	content := c.getNodeRawContent(hr)

	position := c.calculatePosition(hr)
	links := c.extractLinks(hr)
	images := c.extractImages(hr)
	hash := c.calculateContentHash(content)

	return &Chunk{
		ID:       id,
		Type:     "thematic_break",
		Content:  content,
		Text:     "---",
		Level:    0,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
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

	// 遍历blockquote的子节点
	for child := quote.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Paragraph:
			// 处理段落
			for i := 0; i < n.Lines().Len(); i++ {
				line := n.Lines().At(i)
				buf.WriteString("> ")
				// 去除行尾的换行符，因为我们会自己添加
				lineContent := strings.TrimRight(string(line.Value(c.source)), "\n")
				buf.WriteString(lineContent)
				if i < n.Lines().Len()-1 {
					buf.WriteByte('\n')
				}
			}
		case *ast.Blockquote:
			// 处理嵌套的引用块
			nestedContent := c.reconstructBlockquote(n)
			lines := strings.Split(nestedContent, "\n")
			for i, line := range lines {
				if line != "" {
					buf.WriteString("> ")
					buf.WriteString(line)
					if i < len(lines)-1 {
						buf.WriteByte('\n')
					}
				}
			}
		default:
			// 处理其他类型的子节点
			if child.Lines().Len() > 0 {
				for i := 0; i < child.Lines().Len(); i++ {
					line := child.Lines().At(i)
					buf.WriteString("> ")
					lineContent := strings.TrimRight(string(line.Value(c.source)), "\n")
					buf.WriteString(lineContent)
					if i < child.Lines().Len()-1 {
						buf.WriteByte('\n')
					}
				}
			}
		}

		// 在子节点之间添加换行
		if child.NextSibling() != nil {
			buf.WriteByte('\n')
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

// calculateContentHash 计算内容的SHA256哈希值
func (c *MarkdownChunker) calculateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// calculatePosition 计算节点在文档中的位置
func (c *MarkdownChunker) calculatePosition(node ast.Node) ChunkPosition {
	// 默认位置
	defaultPos := ChunkPosition{
		StartLine: 1,
		EndLine:   1,
		StartCol:  1,
		EndCol:    1,
	}

	// 检查节点是否有Lines信息
	if node.Lines().Len() == 0 {
		return defaultPos
	}

	// 将源码按行分割来计算行号
	lines := strings.Split(string(c.source), "\n")

	// 获取节点的字节位置
	segment := node.Lines().At(0)
	startByte := segment.Start
	endByte := segment.Stop

	// 如果节点有多行，获取最后一行的结束位置
	if node.Lines().Len() > 1 {
		lastSegment := node.Lines().At(node.Lines().Len() - 1)
		endByte = lastSegment.Stop
	}

	// 计算起始行号和列号
	startLine := 1
	startCol := 1
	currentByte := 0

	for lineNum, line := range lines {
		lineLength := len(line) + 1 // +1 for newline character
		if currentByte+lineLength > startByte {
			startLine = lineNum + 1
			startCol = startByte - currentByte + 1
			break
		}
		currentByte += lineLength
	}

	// 计算结束行号和列号
	endLine := startLine
	endCol := startCol
	currentByte = 0

	for lineNum, line := range lines {
		lineLength := len(line) + 1 // +1 for newline character
		if currentByte+lineLength > endByte {
			endLine = lineNum + 1
			endCol = endByte - currentByte + 1
			break
		}
		currentByte += lineLength
	}

	return ChunkPosition{
		StartLine: startLine,
		EndLine:   endLine,
		StartCol:  startCol,
		EndCol:    endCol,
	}
}

// extractLinks 从节点中提取链接信息
func (c *MarkdownChunker) extractLinks(node ast.Node) []Link {
	links := make([]Link, 0) // 初始化为空切片而不是nil

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch link := n.(type) {
			case *ast.Link:
				linkText := c.getNodeText(link)
				linkURL := string(link.Destination)
				linkType := c.determineLinkType(linkURL)

				links = append(links, Link{
					Text: linkText,
					URL:  linkURL,
					Type: linkType,
				})
			case *ast.AutoLink:
				linkURL := string(link.URL(c.source))
				linkType := c.determineLinkType(linkURL)

				links = append(links, Link{
					Text: linkURL,
					URL:  linkURL,
					Type: linkType,
				})
			}
		}
		return ast.WalkContinue, nil
	})

	return links
}

// extractImages 从节点中提取图片信息
func (c *MarkdownChunker) extractImages(node ast.Node) []Image {
	images := make([]Image, 0) // 初始化为空切片而不是nil

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if img, ok := n.(*ast.Image); ok {
				// 直接从子节点提取alt文本，避免包含title
				alt := ""
				for child := img.FirstChild(); child != nil; child = child.NextSibling() {
					if textNode, ok := child.(*ast.Text); ok {
						alt += string(textNode.Segment.Value(c.source))
					}
				}

				url := string(img.Destination)
				title := ""
				if img.Title != nil {
					title = string(img.Title)
				}

				images = append(images, Image{
					Alt:    alt,
					URL:    url,
					Title:  title,
					Width:  "", // 这些属性在标准markdown中不可用，可以通过扩展获取
					Height: "",
				})
			}
		}
		return ast.WalkContinue, nil
	})

	return images
}

// determineLinkType 确定链接类型
func (c *MarkdownChunker) determineLinkType(url string) string {
	if strings.HasPrefix(url, "#") {
		return "anchor"
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return "external"
	}
	if strings.HasPrefix(url, "mailto:") {
		return "external"
	}
	return "internal"
}
