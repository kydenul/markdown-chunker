package markdownchunker

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"runtime"
	"slices"
	"strings"
	"time"

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

	Position ChunkPosition `json:"position"` // 在文档中的位置
	Links    []Link        `json:"links"`    // 包含的链接
	Images   []Image       `json:"images"`   // 包含的图片
	Hash     string        `json:"hash"`     // 内容哈希，用于去重
}

// LogContext 表示日志上下文信息
type LogContext struct {
	FunctionName string                 `json:"function_name"` // 函数名
	FileName     string                 `json:"file_name"`     // 文件名
	LineNumber   int                    `json:"line_number"`   // 行号
	NodeType     string                 `json:"node_type"`     // 节点类型
	NodeID       int                    `json:"node_id"`       // 节点ID
	ChunkCount   int                    `json:"chunk_count"`   // 块数量
	DocumentSize int                    `json:"document_size"` // 文档大小
	ProcessTime  time.Duration          `json:"process_time"`  // 处理时间
	Metadata     map[string]interface{} `json:"metadata"`      // 额外元数据
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

// NewLogContext 创建新的日志上下文
func NewLogContext(functionName string) *LogContext {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(1)
	fileName := "unknown"
	lineNumber := 0

	if ok {
		// 提取文件名（不包含路径）
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			fileName = parts[len(parts)-1]
		}
		lineNumber = line
	}

	return &LogContext{
		FunctionName: functionName,
		FileName:     fileName,
		LineNumber:   lineNumber,
		Metadata:     make(map[string]interface{}),
	}
}

// WithNodeInfo 添加节点信息到日志上下文
func (lc *LogContext) WithNodeInfo(nodeType string, nodeID int) *LogContext {
	lc.NodeType = nodeType
	lc.NodeID = nodeID
	return lc
}

// WithDocumentInfo 添加文档信息到日志上下文
func (lc *LogContext) WithDocumentInfo(documentSize int, chunkCount int) *LogContext {
	lc.DocumentSize = documentSize
	lc.ChunkCount = chunkCount
	return lc
}

// WithProcessTime 添加处理时间到日志上下文
func (lc *LogContext) WithProcessTime(duration time.Duration) *LogContext {
	lc.ProcessTime = duration
	return lc
}

// WithMetadata 添加自定义元数据到日志上下文
func (lc *LogContext) WithMetadata(key string, value interface{}) *LogContext {
	lc.Metadata[key] = value
	return lc
}

// WithTableInfo 添加表格特定信息到日志上下文
func (lc *LogContext) WithTableInfo(rowCount, columnCount int, isWellFormed bool) *LogContext {
	lc.Metadata["table_row_count"] = rowCount
	lc.Metadata["table_column_count"] = columnCount
	lc.Metadata["table_well_formed"] = isWellFormed
	return lc
}

// WithListInfo 添加列表特定信息到日志上下文
func (lc *LogContext) WithListInfo(listType string, itemCount int) *LogContext {
	lc.Metadata["list_type"] = listType
	lc.Metadata["list_item_count"] = itemCount
	return lc
}

// WithCodeInfo 添加代码块特定信息到日志上下文
func (lc *LogContext) WithCodeInfo(language string, lineCount int, codeBlockType string) *LogContext {
	lc.Metadata["code_language"] = language
	lc.Metadata["code_line_count"] = lineCount
	lc.Metadata["code_block_type"] = codeBlockType
	return lc
}

// WithHeadingInfo 添加标题特定信息到日志上下文
func (lc *LogContext) WithHeadingInfo(level int, wordCount int) *LogContext {
	lc.Metadata["heading_level"] = level
	lc.Metadata["heading_word_count"] = wordCount
	return lc
}

// WithContentInfo 添加内容统计信息到日志上下文
func (lc *LogContext) WithContentInfo(contentLength, textLength, wordCount int) *LogContext {
	lc.Metadata["content_length"] = contentLength
	lc.Metadata["text_length"] = textLength
	lc.Metadata["word_count"] = wordCount
	return lc
}

// WithPositionInfo 添加位置信息到日志上下文
func (lc *LogContext) WithPositionInfo(startLine, endLine, startCol, endCol int) *LogContext {
	lc.Metadata["start_line"] = startLine
	lc.Metadata["end_line"] = endLine
	lc.Metadata["start_col"] = startCol
	lc.Metadata["end_col"] = endCol
	return lc
}

// WithLinksAndImages 添加链接和图片信息到日志上下文
func (lc *LogContext) WithLinksAndImages(linksCount, imagesCount int) *LogContext {
	lc.Metadata["links_count"] = linksCount
	lc.Metadata["images_count"] = imagesCount
	return lc
}

// ToLogFields 将日志上下文转换为日志字段
func (lc *LogContext) ToLogFields() []interface{} {
	fields := []interface{}{
		"function", lc.FunctionName,
		"file", lc.FileName,
		"line", lc.LineNumber,
	}

	if lc.NodeType != "" {
		fields = append(fields, "node_type", lc.NodeType)
	}
	if lc.NodeID != 0 {
		fields = append(fields, "node_id", lc.NodeID)
	}
	if lc.DocumentSize != 0 {
		fields = append(fields, "document_size", lc.DocumentSize)
	}
	if lc.ChunkCount != 0 {
		fields = append(fields, "chunk_count", lc.ChunkCount)
	}
	if lc.ProcessTime != 0 {
		fields = append(fields, "process_time_ms", lc.ProcessTime.Milliseconds())
	}

	// 添加自定义元数据
	for key, value := range lc.Metadata {
		fields = append(fields, key, value)
	}

	return fields
}

// logWithContext 使用上下文信息记录日志的辅助函数
func (c *MarkdownChunker) logWithContext(level string, message string, context *LogContext) {
	fields := context.ToLogFields()

	switch level {
	case "debug":
		c.logger.Debugw(message, fields...)
	case "info":
		c.logger.Infow(message, fields...)
	case "warn":
		c.logger.Warnw(message, fields...)
	case "error":
		c.logger.Errorw(message, fields...)
	default:
		c.logger.Infow(message, fields...)
	}
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
	// 创建临时日志器用于日志配置验证过程的日志记录
	tempLogger := log.NewLogger(&log.Options{
		Level:      "debug",
		Format:     "console",
		Directory:  "./logs",
		TimeLayout: "2006-01-02 15:04:05.000",
	})

	tempLogger.Debugw("开始日志配置验证",
		"function", "validateLogConfig")

	if config == nil {
		tempLogger.Errorw("日志配置验证失败：配置对象为空",
			"function", "validateLogConfig",
			"error_type", "config_null")

		return NewChunkerError(ErrorTypeConfigInvalid, "配置对象不能为空", nil).
			WithContext("function", "validateLogConfig")
	}

	// 验证日志级别
	tempLogger.Debugw("验证日志级别配置",
		"function", "validateLogConfig",
		"log_level", config.LogLevel)

	validLevels := map[string]bool{
		"DEBUG": true, "INFO": true, "WARN": true, "WARNING": true, "ERROR": true,
	}
	if config.LogLevel != "" && !validLevels[strings.ToUpper(config.LogLevel)] {
		var validLevelsList []string
		for level := range validLevels {
			validLevelsList = append(validLevelsList, level)
		}

		tempLogger.Errorw("日志配置验证失败：无效的日志级别",
			"function", "validateLogConfig",
			"field", "LogLevel",
			"invalid_level", config.LogLevel,
			"valid_levels", validLevelsList,
			"error_type", "invalid_log_level")

		return NewChunkerError(ErrorTypeConfigInvalid, "invalid log level configuration", nil).
			WithContext("function", "validateLogConfig").
			WithContext("field", "LogLevel").
			WithContext("invalid_level", config.LogLevel).
			WithContext("valid_levels", validLevelsList).
			WithContext("recommendation", "请使用有效的日志级别")
	}

	// 验证日志格式
	tempLogger.Debugw("验证日志格式配置",
		"function", "validateLogConfig",
		"log_format", config.LogFormat)

	validFormats := map[string]bool{
		"json": true, "console": true,
	}
	if config.LogFormat != "" && !validFormats[strings.ToLower(config.LogFormat)] {
		var validFormatsList []string
		for format := range validFormats {
			validFormatsList = append(validFormatsList, format)
		}

		tempLogger.Errorw("日志配置验证失败：无效的日志格式",
			"function", "validateLogConfig",
			"field", "LogFormat",
			"invalid_format", config.LogFormat,
			"valid_formats", validFormatsList,
			"error_type", "invalid_log_format")

		return NewChunkerError(ErrorTypeConfigInvalid, "无效的日志格式配置", nil).
			WithContext("function", "validateLogConfig").
			WithContext("field", "LogFormat").
			WithContext("invalid_format", config.LogFormat).
			WithContext("valid_formats", validFormatsList).
			WithContext("recommendation", "请使用json或console格式")
	}

	// 验证日志目录（如果为空，将使用默认值）
	tempLogger.Debugw("验证日志目录配置",
		"function", "validateLogConfig",
		"log_directory", config.LogDirectory)

	if config.LogDirectory != "" {
		if strings.TrimSpace(config.LogDirectory) == "" {
			tempLogger.Errorw("日志配置验证失败：日志目录为空白字符",
				"function", "validateLogConfig",
				"field", "LogDirectory",
				"value", config.LogDirectory,
				"error_type", "invalid_log_directory")

			return NewChunkerError(ErrorTypeConfigInvalid, "日志目录不能为空白字符", nil).
				WithContext("function", "validateLogConfig").
				WithContext("field", "LogDirectory").
				WithContext("value", config.LogDirectory).
				WithContext("recommendation", "请提供有效的目录路径或留空使用默认值")
		}
	}

	// 记录日志配置验证成功
	tempLogger.Debugw("日志配置验证成功",
		"function", "validateLogConfig",
		"log_level", config.LogLevel,
		"enable_log", config.EnableLog,
		"log_format", config.LogFormat,
		"log_directory", config.LogDirectory,
		"validation_result", "passed")

	return nil
}

// ValidateConfig 验证配置的有效性
func ValidateConfig(config *ChunkerConfig) error {
	// 创建临时日志器用于配置验证日志记录
	tempLogger := log.NewLogger(&log.Options{
		Level:      "info",
		Format:     "console",
		Directory:  "./logs",
		TimeLayout: "2006-01-02 15:04:05.000",
	})

	// 记录配置验证开始日志
	tempLogger.Infow("开始配置验证",
		"function", "ValidateConfig")

	if config == nil {
		tempLogger.Errorw("配置验证失败：配置对象为空",
			"function", "ValidateConfig",
			"error_type", "config_null",
			"expected", "non-nil ChunkerConfig",
			"actual", "nil")

		return NewChunkerError(ErrorTypeConfigInvalid, "配置对象不能为空", nil).
			WithContext("function", "ValidateConfig").
			WithContext("expected", "non-nil ChunkerConfig").
			WithContext("actual", "nil")
	}

	// 记录开始验证的配置参数
	tempLogger.Infow("验证配置参数",
		"function", "ValidateConfig",
		"max_chunk_size", config.MaxChunkSize,
		"enabled_types_count", len(config.EnabledTypes),
		"memory_limit_bytes", config.MemoryLimit,
		"memory_limit_mb", config.MemoryLimit/(1024*1024),
		"log_level", config.LogLevel,
		"enable_log", config.EnableLog,
		"log_format", config.LogFormat,
		"log_directory", config.LogDirectory,
		"error_handling_mode", config.ErrorHandling,
		"performance_mode", config.PerformanceMode,
		"filter_empty_chunks", config.FilterEmptyChunks,
		"preserve_whitespace", config.PreserveWhitespace,
		"enable_object_pooling", config.EnableObjectPooling)

	if config.MaxChunkSize < 0 {
		tempLogger.Errorw("配置验证失败：最大块大小无效",
			"function", "ValidateConfig",
			"field", "MaxChunkSize",
			"value", config.MaxChunkSize,
			"minimum_allowed", 0,
			"error_type", "invalid_chunk_size")

		return NewChunkerError(ErrorTypeConfigInvalid, "最大块大小不能为负数", nil).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MaxChunkSize").
			WithContext("value", config.MaxChunkSize).
			WithContext("minimum_allowed", 0)
	}

	// 验证启用的类型
	if config.EnabledTypes != nil {
		tempLogger.Debugw("验证启用的内容类型",
			"function", "ValidateConfig",
			"enabled_types_count", len(config.EnabledTypes))

		validTypes := map[string]bool{
			"heading": true, "paragraph": true, "code": true,
			"table": true, "list": true, "blockquote": true,
			"thematic_break": true,
		}

		for typeName := range config.EnabledTypes {
			if !validTypes[typeName] {
				var validTypesList []string
				for vt := range validTypes {
					validTypesList = append(validTypesList, vt)
				}

				tempLogger.Errorw("配置验证失败：无效的内容类型",
					"function", "ValidateConfig",
					"field", "EnabledTypes",
					"invalid_type", typeName,
					"valid_types", validTypesList,
					"error_type", "invalid_content_type")

				return NewChunkerError(ErrorTypeConfigInvalid, "无效的内容类型配置", nil).
					WithContext("function", "ValidateConfig").
					WithContext("field", "EnabledTypes").
					WithContext("invalid_type", typeName).
					WithContext("valid_types", validTypesList).
					WithContext("recommendation", "请使用有效的内容类型名称")
			}
		}

		tempLogger.Debugw("内容类型验证通过",
			"function", "ValidateConfig",
			"validated_types", len(config.EnabledTypes))
	}

	// 验证内存限制
	if config.MemoryLimit < 0 {
		tempLogger.Errorw("配置验证失败：内存限制无效",
			"function", "ValidateConfig",
			"field", "MemoryLimit",
			"value", config.MemoryLimit,
			"minimum_allowed", 0,
			"error_type", "invalid_memory_limit")

		return NewChunkerError(ErrorTypeConfigInvalid, "内存限制不能为负数", nil).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MemoryLimit").
			WithContext("value", config.MemoryLimit).
			WithContext("minimum_allowed", 0)
	}

	// 验证日志配置
	tempLogger.Debugw("验证日志配置",
		"function", "ValidateConfig",
		"log_level", config.LogLevel,
		"enable_log", config.EnableLog,
		"log_format", config.LogFormat,
		"log_directory", config.LogDirectory)

	if err := validateLogConfig(config); err != nil {
		tempLogger.Errorw("配置验证失败：日志配置无效",
			"function", "ValidateConfig",
			"validation_step", "log_config",
			"error", err.Error(),
			"error_type", "invalid_log_config")

		if chunkerErr, ok := err.(*ChunkerError); ok {
			return chunkerErr
		}
		return NewChunkerError(ErrorTypeConfigInvalid, "日志配置验证失败", err).
			WithContext("function", "ValidateConfig").
			WithContext("validation_step", "log_config")
	}

	// 记录配置验证成功日志
	tempLogger.Infow("配置验证成功",
		"function", "ValidateConfig",
		"max_chunk_size", config.MaxChunkSize,
		"memory_limit_mb", config.MemoryLimit/(1024*1024),
		"log_level", config.LogLevel,
		"log_format", config.LogFormat,
		"log_directory", config.LogDirectory,
		"validation_result", "passed")

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
		// 创建临时日志器来记录配置错误
		tempLogger := log.NewLogger(&log.Options{
			Level:      "error",
			Format:     "console",
			Directory:  "./logs",
			TimeLayout: "2006-01-02 15:04:05.000",
		})

		tempLogger.Errorw("配置验证失败，使用默认配置",
			"validation_error", err.Error(),
			"function", "NewMarkdownChunkerWithConfig")

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

	// 记录系统初始化开始日志
	logger.Infow("开始初始化 MarkdownChunker 系统",
		"function", "NewMarkdownChunkerWithConfig",
		"initialization_phase", "start")

	// 记录系统信息和配置参数
	logger.Infow("系统初始化信息",
		"function", "NewMarkdownChunkerWithConfig",
		"go_version", runtime.Version(),
		"go_arch", runtime.GOARCH,
		"go_os", runtime.GOOS,
		"num_cpu", runtime.NumCPU(),
		"max_chunk_size", config.MaxChunkSize,
		"memory_limit_bytes", config.MemoryLimit,
		"memory_limit_mb", config.MemoryLimit/(1024*1024),
		"error_handling_mode", config.ErrorHandling,
		"performance_mode", config.PerformanceMode,
		"filter_empty_chunks", config.FilterEmptyChunks,
		"preserve_whitespace", config.PreserveWhitespace,
		"enable_object_pooling", config.EnableObjectPooling)

	// 记录日志配置信息
	logger.Infow("日志系统配置",
		"function", "NewMarkdownChunkerWithConfig",
		"log_level", config.LogLevel,
		"parsed_log_level", parseLogLevel(config.LogLevel),
		"enable_log", config.EnableLog,
		"log_format", config.LogFormat,
		"log_directory", config.LogDirectory,
		"max_file_size_mb", opts.MaxSize,
		"max_backups", opts.MaxBackups,
		"time_layout", opts.TimeLayout)

	// 记录启用的内容类型
	if config.EnabledTypes != nil {
		var enabledTypesList []string
		for typeName, enabled := range config.EnabledTypes {
			if enabled {
				enabledTypesList = append(enabledTypesList, typeName)
			}
		}
		logger.Infow("内容类型配置",
			"function", "NewMarkdownChunkerWithConfig",
			"enabled_types", enabledTypesList,
			"enabled_types_count", len(enabledTypesList))
	} else {
		logger.Infow("内容类型配置",
			"function", "NewMarkdownChunkerWithConfig",
			"enabled_types", "all",
			"enabled_types_count", "unlimited")
	}

	// 创建错误处理器并设置日志器
	logger.Debugw("初始化错误处理器",
		"function", "NewMarkdownChunkerWithConfig",
		"error_handling_mode", config.ErrorHandling)

	errorHandler := NewDefaultErrorHandler(config.ErrorHandling)
	errorHandler.SetLogger(logger)

	// 创建性能监控器并设置日志器
	logger.Debugw("初始化性能监控器",
		"function", "NewMarkdownChunkerWithConfig",
		"performance_mode", config.PerformanceMode)

	performanceMonitor := NewPerformanceMonitor()
	performanceMonitor.SetLogger(logger)

	// 创建内存优化器并设置日志器（如果启用）
	var memoryOptimizer *MemoryOptimizer
	if config.EnableObjectPooling || config.MemoryLimit > 0 {
		logger.Debugw("初始化内存优化器",
			"function", "NewMarkdownChunkerWithConfig",
			"memory_limit_bytes", config.MemoryLimit,
			"memory_limit_mb", config.MemoryLimit/(1024*1024),
			"object_pooling_enabled", config.EnableObjectPooling)

		memoryOptimizer = NewMemoryOptimizer(config.MemoryLimit)
		memoryOptimizer.SetLogger(logger)

		logger.Infow("内存优化器已启用",
			"memory_limit_bytes", config.MemoryLimit,
			"memory_limit_mb", config.MemoryLimit/(1024*1024),
			"object_pooling_enabled", config.EnableObjectPooling,
			"function", "NewMarkdownChunkerWithConfig")
	} else {
		logger.Debugw("内存优化器未启用",
			"function", "NewMarkdownChunkerWithConfig",
			"reason", "memory_limit_and_object_pooling_disabled")
	}

	// 创建优化的字符串操作
	logger.Debugw("初始化字符串操作优化器",
		"function", "NewMarkdownChunkerWithConfig")

	stringOps := NewOptimizedStringOperations()

	// 记录自定义元数据提取器信息
	if len(config.CustomExtractors) > 0 {
		logger.Infow("自定义元数据提取器配置",
			"function", "NewMarkdownChunkerWithConfig",
			"extractors_count", len(config.CustomExtractors))
	}

	// 记录系统初始化完成日志
	logger.Infow("MarkdownChunker 系统初始化完成",
		"function", "NewMarkdownChunkerWithConfig",
		"initialization_phase", "complete",
		"components_initialized", []string{
			"goldmark_parser", "logger", "error_handler",
			"performance_monitor", "string_operations",
		})

	return &MarkdownChunker{
		md:                 md,
		config:             config,
		errorHandler:       errorHandler,
		performanceMonitor: performanceMonitor,
		memoryOptimizer:    memoryOptimizer,
		stringOps:          stringOps,
		chunks:             []Chunk{},
		logger:             logger,
	}
}

// ChunkDocument 对整个文档进行分块
func (c *MarkdownChunker) ChunkDocument(content []byte) ([]Chunk, error) {
	// 创建日志上下文
	logCtx := NewLogContext("ChunkDocument").WithDocumentInfo(len(content), 0)

	// 记录文档处理开始日志
	c.logWithContext("info", "开始处理 Markdown 文档", logCtx)

	// 开始性能监控
	c.performanceMonitor.Start()
	defer func() {
		c.performanceMonitor.Stop()

		// 获取性能统计信息
		stats := c.performanceMonitor.GetStats()

		// 更新日志上下文并记录文档处理结束日志
		endLogCtx := NewLogContext("ChunkDocument").
			WithDocumentInfo(len(content), len(c.chunks)).
			WithProcessTime(stats.ProcessingTime).
			WithMetadata("memory_used_bytes", stats.MemoryUsed)

		c.logWithContext("info", "完成 Markdown 文档处理", endLogCtx)
	}()

	// 记录输入文档大小
	c.performanceMonitor.RecordBytes(int64(len(content)))

	// 清除之前的错误
	c.errorHandler.ClearErrors()

	// 输入验证
	if content == nil {
		errorLogCtx := NewLogContext("ChunkDocument").WithMetadata("error", "content cannot be nil")
		c.logWithContext("error", "输入内容为空", errorLogCtx)

		err := NewChunkerError(ErrorTypeInvalidInput, "输入内容不能为空", nil).
			WithContext("function", "ChunkDocument").
			WithContext("validation_step", "input_check").
			WithContext("expected", "non-nil byte slice").
			WithContext("actual", "nil")
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
		return []Chunk{}, nil
	}

	// 记录空文档处理
	if len(content) == 0 {
		emptyLogCtx := NewLogContext("ChunkDocument").WithDocumentInfo(0, 0)
		c.logWithContext("info", "处理空文档", emptyLogCtx)
		return []Chunk{}, nil
	}

	// 检查内容大小
	if len(content) > 100*1024*1024 { // 100MB 限制
		sizeErrorLogCtx := NewLogContext("ChunkDocument").
			WithDocumentInfo(len(content), 0).
			WithMetadata("size_limit_bytes", 100*1024*1024).
			WithMetadata("size_limit_mb", 100).
			WithMetadata("document_size_mb", len(content)/(1024*1024))
		c.logWithContext("error", "文档大小超过限制", sizeErrorLogCtx)

		err := NewChunkerError(ErrorTypeMemoryExhausted, "文档大小超过系统限制", nil).
			WithContext("function", "ChunkDocument").
			WithContext("document_size_bytes", len(content)).
			WithContext("size_limit_bytes", 100*1024*1024).
			WithContext("size_limit_mb", 100).
			WithContext("document_size_mb", len(content)/(1024*1024)).
			WithContext("recommendation", "请考虑分割文档或增加内存限制")
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
	}

	// 记录大型文档处理警告
	if len(content) > 10*1024*1024 { // 10MB 警告阈值
		largeDocLogCtx := NewLogContext("ChunkDocument").
			WithDocumentInfo(len(content), 0).
			WithMetadata("recommendation", "考虑分批处理以优化性能")
		c.logWithContext("warn", "处理大型文档", largeDocLogCtx)
	}

	c.source = content
	c.chunks = []Chunk{}

	// 解析 Markdown
	parseLogCtx := NewLogContext("ChunkDocument").WithDocumentInfo(len(content), 0)
	c.logWithContext("debug", "开始解析 Markdown AST", parseLogCtx)

	reader := text.NewReader(content)
	doc := c.md.Parser().Parse(reader)

	// 检查解析结果
	if doc == nil {
		parseFailLogCtx := NewLogContext("ChunkDocument").WithDocumentInfo(len(content), 0)
		c.logWithContext("error", "Markdown AST 解析失败", parseFailLogCtx)

		err := NewChunkerError(ErrorTypeParsingFailed, "Markdown文档解析失败", nil).
			WithContext("function", "ChunkDocument").
			WithContext("document_size_bytes", len(content)).
			WithContext("parser_type", "goldmark").
			WithContext("content_preview", string(content[:min(500, len(content))])).
			WithContext("recommendation", "检查Markdown语法是否正确")
		if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
			return nil, handlerErr
		}
		return []Chunk{}, nil
	}

	parseCompleteLogCtx := NewLogContext("ChunkDocument").WithDocumentInfo(len(content), 0)
	c.logWithContext("debug", "Markdown AST 解析完成", parseCompleteLogCtx)

	// 遍历顶层节点进行分块
	traverseLogCtx := NewLogContext("ChunkDocument").WithDocumentInfo(len(content), 0)
	c.logWithContext("debug", "开始遍历 AST 节点进行分块", traverseLogCtx)

	chunkID := 0
	processedNodes := 0

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		processedNodes++

		// 使用defer和recover来捕获节点处理中的panic
		var chunk *Chunk
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicLogCtx := NewLogContext("ChunkDocument").
						WithNodeInfo(child.Kind().String(), chunkID).
						WithMetadata("processed_nodes", processedNodes).
						WithMetadata("panic_value", fmt.Sprintf("%v", r))
					c.logWithContext("error", "节点处理发生panic", panicLogCtx)

					err := NewChunkerError(ErrorTypeParsingFailed, "节点处理过程中发生严重错误", fmt.Errorf("panic: %v", r)).
						WithContext("function", "ChunkDocument").
						WithContext("node_id", chunkID).
						WithContext("processed_nodes", processedNodes).
						WithContext("panic_value", fmt.Sprintf("%v", r)).
						WithContext("node_type", child.Kind().String()).
						WithContext("recovery_action", "跳过当前节点继续处理")
					c.errorHandler.HandleError(err)
					chunk = nil // 确保不处理这个块
				}
			}()

			chunk = c.processNode(child, chunkID)
		}()

		if chunk != nil {
			// 检查类型是否启用
			if !c.isTypeEnabled(chunk.Type) {
				skipTypeLogCtx := NewLogContext("ChunkDocument").
					WithNodeInfo(chunk.Type, chunkID).
					WithMetadata("chunk_type", chunk.Type)
				c.logWithContext("debug", "跳过未启用的块类型", skipTypeLogCtx)
				continue
			}

			// 检查是否过滤空块
			if c.config.FilterEmptyChunks && strings.TrimSpace(chunk.Text) == "" {
				filterEmptyLogCtx := NewLogContext("ChunkDocument").
					WithNodeInfo(chunk.Type, chunkID).
					WithMetadata("chunk_type", chunk.Type)
				c.logWithContext("debug", "过滤空块", filterEmptyLogCtx)
				continue
			}

			// 检查块大小限制
			if c.config.MaxChunkSize > 0 && len(chunk.Content) > c.config.MaxChunkSize {
				oversizeLogCtx := NewLogContext("ChunkDocument").
					WithNodeInfo(chunk.Type, chunk.ID).
					WithMetadata("chunk_size", len(chunk.Content)).
					WithMetadata("max_size", c.config.MaxChunkSize).
					WithMetadata("size_ratio", float64(len(chunk.Content))/float64(c.config.MaxChunkSize))
				c.logWithContext("warn", "块大小超过限制", oversizeLogCtx)

				err := NewChunkerError(ErrorTypeChunkTooLarge, "生成的块大小超过配置限制", nil).
					WithContext("function", "ChunkDocument").
					WithContext("chunk_id", chunk.ID).
					WithContext("chunk_type", chunk.Type).
					WithContext("chunk_size_bytes", len(chunk.Content)).
					WithContext("max_size_bytes", c.config.MaxChunkSize).
					WithContext("size_ratio", float64(len(chunk.Content))/float64(c.config.MaxChunkSize)).
					WithContext("content_preview", chunk.Content[:min(100, len(chunk.Content))]).
					WithContext("handling_mode", c.config.ErrorHandling).
					WithContext("recommendation", "考虑增加MaxChunkSize或启用内容截断")

				if handlerErr := c.errorHandler.HandleError(err); handlerErr != nil {
					return nil, handlerErr
				}

				// 在宽松模式下截断内容
				if c.config.ErrorHandling != ErrorModeStrict {
					truncateLogCtx := NewLogContext("ChunkDocument").
						WithNodeInfo(chunk.Type, chunk.ID).
						WithMetadata("original_size", len(chunk.Content)).
						WithMetadata("truncated_size", c.config.MaxChunkSize)
					c.logWithContext("info", "截断超大块内容", truncateLogCtx)

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
			successChunkLogCtx := NewLogContext("ChunkDocument").
				WithNodeInfo(chunk.Type, chunk.ID).
				WithContentInfo(len(chunk.Content), len(chunk.Text), len(strings.Fields(chunk.Text)))
			c.logWithContext("debug", "成功处理块", successChunkLogCtx)

			chunkID++
		}

		// 每处理100个节点记录一次进度（用于大型文档）
		if processedNodes%100 == 0 && len(content) > 1024*1024 { // 只对大于1MB的文档记录进度
			// 检查内存使用情况
			c.performanceMonitor.CheckMemoryThresholds()

			// 如果启用了内存优化器，检查内存限制
			if c.memoryOptimizer != nil {
				if err := c.memoryOptimizer.CheckMemoryLimit(); err != nil {
					memoryErrorLogCtx := NewLogContext("ChunkDocument").
						WithDocumentInfo(len(content), len(c.chunks)).
						WithMetadata("processed_nodes", processedNodes).
						WithMetadata("error", err.Error())
					c.logWithContext("warn", "内存限制检查失败", memoryErrorLogCtx)

					// 在宽松模式下，尝试强制GC并继续
					if c.config.ErrorHandling != ErrorModeStrict {
						c.memoryOptimizer.ForceGC()
					} else {
						return nil, err
					}
				}

				// 记录已处理的字节数（用于GC触发）
				c.memoryOptimizer.RecordProcessedBytes(int64(len(content)) / 100) // 分摊到每100个节点
			}

			progressLogCtx := NewLogContext("ChunkDocument").
				WithDocumentInfo(len(content), len(c.chunks)).
				WithMetadata("processed_nodes", processedNodes).
				WithMetadata("document_size_mb", len(content)/(1024*1024)).
				WithMetadata("progress_percentage", float64(processedNodes*100)/float64(len(content)/1000))
			c.logWithContext("info", "文档处理进度", progressLogCtx)
		}
	}

	completeTraverseLogCtx := NewLogContext("ChunkDocument").
		WithDocumentInfo(len(content), len(c.chunks)).
		WithMetadata("total_processed_nodes", processedNodes)
	c.logWithContext("debug", "完成 AST 节点遍历", completeTraverseLogCtx)

	// 最终的资源使用情况检查和报告
	if c.memoryOptimizer != nil {
		// 检查最终内存状态
		if err := c.memoryOptimizer.CheckMemoryLimit(); err != nil {
			finalMemoryLogCtx := NewLogContext("ChunkDocument").
				WithDocumentInfo(len(content), len(c.chunks)).
				WithMetadata("error", err.Error()).
				WithMetadata("recommendation", "考虑在后续处理中释放资源")
			c.logWithContext("warn", "处理完成时内存使用仍然较高", finalMemoryLogCtx)
		}

		// 获取内存优化器统计信息
		memStats := c.memoryOptimizer.GetMemoryStats()
		memStatsLogCtx := NewLogContext("ChunkDocument").
			WithDocumentInfo(len(content), len(c.chunks)).
			WithMetadata("current_memory_mb", memStats.CurrentMemory/(1024*1024)).
			WithMetadata("memory_limit_mb", memStats.MemoryLimit/(1024*1024)).
			WithMetadata("processed_bytes_mb", memStats.ProcessedBytes/(1024*1024)).
			WithMetadata("gc_threshold_mb", memStats.GCThreshold/(1024*1024)).
			WithMetadata("total_allocations_mb", memStats.TotalAllocations/(1024*1024)).
			WithMetadata("gc_cycles", memStats.GCCycles)
		c.logWithContext("info", "内存优化器统计", memStatsLogCtx)
	}

	// 最终性能检查
	c.performanceMonitor.CheckMemoryThresholds()

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
	// 获取节点类型名称
	nodeType := node.Kind().String()

	// 创建节点处理日志上下文
	nodeLogCtx := NewLogContext("processNode").WithNodeInfo(nodeType, id)

	// 记录节点处理开始日志
	c.logWithContext("debug", "开始处理 AST 节点", nodeLogCtx)

	var chunk *Chunk

	switch n := node.(type) {
	case *ast.Heading:
		chunk = c.processHeading(n, id)
	case *ast.Paragraph:
		chunk = c.processParagraph(n, id)
	case *ast.FencedCodeBlock:
		chunk = c.processCodeBlock(n, id)
	case *ast.CodeBlock:
		chunk = c.processCodeBlock(n, id)
	case *extast.Table:
		chunk = c.processTable(n, id)
	case *ast.List:
		chunk = c.processList(n, id)
	case *ast.Blockquote:
		chunk = c.processBlockquote(n, id)
	case *ast.ThematicBreak:
		chunk = c.processThematicBreak(n, id)
	default:
		skipLogCtx := NewLogContext("processNode").WithNodeInfo(nodeType, id)
		c.logWithContext("debug", "跳过不支持的节点类型", skipLogCtx)
		return nil
	}

	// 记录节点处理结果日志
	if chunk != nil {
		successLogCtx := NewLogContext("processNode").
			WithNodeInfo(nodeType, id).
			WithMetadata("chunk_type", chunk.Type).
			WithContentInfo(len(chunk.Content), len(chunk.Text), len(strings.Fields(chunk.Text)))
		c.logWithContext("debug", "成功处理 AST 节点", successLogCtx)
	} else {
		nullLogCtx := NewLogContext("processNode").WithNodeInfo(nodeType, id)
		c.logWithContext("debug", "节点处理返回空块", nullLogCtx)
	}

	return chunk
}

// processHeading 处理标题
func (c *MarkdownChunker) processHeading(heading *ast.Heading, id int) *Chunk {
	// 创建标题处理日志上下文
	headingLogCtx := NewLogContext("processHeading").
		WithNodeInfo("Heading", id).
		WithHeadingInfo(heading.Level, 0) // wordCount will be updated later

	c.logWithContext("debug", "处理标题节点", headingLogCtx)

	content := c.getNodeRawContent(heading)
	text := c.getNodeText(heading)
	position := c.calculatePosition(heading)
	links := c.extractLinks(heading)
	images := c.extractImages(heading)
	hash := c.calculateContentHash(content)

	// 记录提取的内容统计信息
	wordCount := len(strings.Fields(text))
	completeLogCtx := NewLogContext("processHeading").
		WithNodeInfo("Heading", id).
		WithHeadingInfo(heading.Level, wordCount).
		WithContentInfo(len(content), len(text), wordCount).
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images))

	c.logWithContext("debug", "标题内容提取完成", completeLogCtx)

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
	// 创建段落处理日志上下文
	paraLogCtx := NewLogContext("processParagraph").WithNodeInfo("Paragraph", id)
	c.logWithContext("debug", "处理段落节点", paraLogCtx)

	content := c.getNodeRawContent(para)
	text := c.getNodeText(para)

	// 过滤掉空段落
	if strings.TrimSpace(text) == "" {
		emptyParaLogCtx := NewLogContext("processParagraph").WithNodeInfo("Paragraph", id)
		c.logWithContext("debug", "跳过空段落", emptyParaLogCtx)
		return nil
	}

	position := c.calculatePosition(para)
	links := c.extractLinks(para)
	images := c.extractImages(para)
	hash := c.calculateContentHash(content)

	// 记录提取的内容统计信息
	wordCount := len(strings.Fields(text))
	completeParaLogCtx := NewLogContext("processParagraph").
		WithNodeInfo("Paragraph", id).
		WithContentInfo(len(content), len(text), wordCount).
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images)).
		WithMetadata("char_count", len(text))

	c.logWithContext("debug", "段落内容提取完成", completeParaLogCtx)

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
	// 确定代码块类型
	codeBlockType := "indented"
	if _, ok := code.(*ast.FencedCodeBlock); ok {
		codeBlockType = "fenced"
	}

	// 创建代码块处理日志上下文
	codeLogCtx := NewLogContext("processCodeBlock").
		WithNodeInfo("CodeBlock", id).
		WithCodeInfo("", 0, codeBlockType) // language and lineCount will be updated later

	c.logWithContext("debug", "处理代码块节点", codeLogCtx)

	var language string
	content := c.getNodeRawContent(code)

	// 提取代码块的纯文本内容，去除尾部空行
	var codeLines []string
	for i := 0; i < code.Lines().Len(); i++ {
		line := code.Lines().At(i)
		codeLines = append(codeLines, string(line.Value(c.source)))
	}

	// 去除尾部的空行
	originalLineCount := len(codeLines)
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

	// 记录提取的内容统计信息
	completeCodeLogCtx := NewLogContext("processCodeBlock").
		WithNodeInfo("CodeBlock", id).
		WithCodeInfo(language, lineCount, codeBlockType).
		WithContentInfo(len(content), len(text), 0). // code blocks don't have meaningful word count
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images)).
		WithMetadata("original_line_count", originalLineCount).
		WithMetadata("cleaned_line_count", lineCount)

	c.logWithContext("debug", "代码块内容提取完成", completeCodeLogCtx)

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
	// 创建表格处理日志上下文
	tableLogCtx := NewLogContext("processTable").WithNodeInfo("Table", id)
	c.logWithContext("debug", "处理表格节点", tableLogCtx)

	content := c.getNodeRawContent(table)
	text := c.getNodeText(table)

	// 使用高级表格处理器分析表格
	processor := NewAdvancedTableProcessor(c.source)
	tableInfo := processor.ProcessTable(table)

	// 获取基础元数据
	metadata := tableInfo.GetTableMetadata()

	// 如果表格格式有问题，记录错误
	if !tableInfo.IsWellFormed && len(tableInfo.Errors) > 0 {
		errorTableLogCtx := NewLogContext("processTable").
			WithNodeInfo("Table", id).
			WithMetadata("table_errors", strings.Join(tableInfo.Errors, "; "))
		c.logWithContext("debug", "检测到表格格式问题", errorTableLogCtx)

		err := NewChunkerError(ErrorTypeParsingFailed, "表格格式解析存在问题", nil).
			WithContext("function", "processTable").
			WithContext("chunk_id", id).
			WithContext("chunk_type", "table").
			WithContext("table_errors", strings.Join(tableInfo.Errors, "; ")).
			WithContext("error_count", len(tableInfo.Errors)).
			WithContext("is_well_formed", tableInfo.IsWellFormed).
			WithContext("table_metadata", tableInfo.GetTableMetadata()).
			WithContext("content_preview", content[:min(200, len(content))]).
			WithContext("recommendation", "检查表格的Markdown语法格式")
		c.errorHandler.HandleError(err)
	}

	position := c.calculatePosition(table)
	links := c.extractLinks(table)
	images := c.extractImages(table)
	hash := c.calculateContentHash(content)

	// 记录提取的内容统计信息
	// 解析行数和列数
	rowCount := 0
	columnCount := 0
	if rowCountStr, ok := metadata["row_count"]; ok {
		if rc, err := fmt.Sscanf(rowCountStr, "%d", &rowCount); err == nil && rc == 1 {
			// rowCount parsed successfully
		}
	}
	if columnCountStr, ok := metadata["column_count"]; ok {
		if cc, err := fmt.Sscanf(columnCountStr, "%d", &columnCount); err == nil && cc == 1 {
			// columnCount parsed successfully
		}
	}

	completeTableLogCtx := NewLogContext("processTable").
		WithNodeInfo("Table", id).
		WithTableInfo(rowCount, columnCount, tableInfo.IsWellFormed).
		WithContentInfo(len(content), len(text), len(strings.Fields(text))).
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images))

	c.logWithContext("debug", "表格内容提取完成", completeTableLogCtx)

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
	listType := "unordered"
	if list.IsOrdered() {
		listType = "ordered"
	}

	// 创建列表处理日志上下文
	listLogCtx := NewLogContext("processList").
		WithNodeInfo("List", id).
		WithListInfo(listType, 0) // itemCount will be updated later

	c.logWithContext("debug", "处理列表节点", listLogCtx)

	content := c.getNodeRawContent(list)
	text := c.getListText(list)

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

	// 记录提取的内容统计信息
	completeListLogCtx := NewLogContext("processList").
		WithNodeInfo("List", id).
		WithListInfo(listType, itemCount).
		WithContentInfo(len(content), len(text), len(strings.Fields(text))).
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images))

	c.logWithContext("debug", "列表内容提取完成", completeListLogCtx)

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
	// 创建引用块处理日志上下文
	quoteLogCtx := NewLogContext("processBlockquote").WithNodeInfo("Blockquote", id)
	c.logWithContext("debug", "处理引用块节点", quoteLogCtx)

	content := c.getNodeRawContent(quote)
	text := c.getNodeText(quote)

	position := c.calculatePosition(quote)
	links := c.extractLinks(quote)
	images := c.extractImages(quote)
	hash := c.calculateContentHash(content)

	// 记录提取的内容统计信息
	wordCount := len(strings.Fields(text))
	completeQuoteLogCtx := NewLogContext("processBlockquote").
		WithNodeInfo("Blockquote", id).
		WithContentInfo(len(content), len(text), wordCount).
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images))

	c.logWithContext("debug", "引用块内容提取完成", completeQuoteLogCtx)

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
	// 创建分隔线处理日志上下文
	hrLogCtx := NewLogContext("processThematicBreak").WithNodeInfo("ThematicBreak", id)
	c.logWithContext("debug", "处理分隔线节点", hrLogCtx)

	content := c.getNodeRawContent(hr)

	position := c.calculatePosition(hr)
	links := c.extractLinks(hr)
	images := c.extractImages(hr)
	hash := c.calculateContentHash(content)

	// 记录提取的内容统计信息
	completeHrLogCtx := NewLogContext("processThematicBreak").
		WithNodeInfo("ThematicBreak", id).
		WithContentInfo(len(content), 3, 0). // ThematicBreak has fixed text "---"
		WithPositionInfo(position.StartLine, position.EndLine, position.StartCol, position.EndCol).
		WithLinksAndImages(len(links), len(images))

	c.logWithContext("debug", "分隔线内容提取完成", completeHrLogCtx)

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
