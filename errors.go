package markdownchunker

import (
	"fmt"
	"time"
)

// ErrorType 错误类型
type ErrorType int

const (
	// ErrorTypeInvalidInput 无效输入错误
	ErrorTypeInvalidInput ErrorType = iota
	// ErrorTypeParsingFailed 解析失败错误
	ErrorTypeParsingFailed
	// ErrorTypeMemoryExhausted 内存不足错误
	ErrorTypeMemoryExhausted
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout
	// ErrorTypeConfigInvalid 配置无效错误
	ErrorTypeConfigInvalid
	// ErrorTypeChunkTooLarge 块过大错误
	ErrorTypeChunkTooLarge
)

// String 返回错误类型的字符串表示
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeInvalidInput:
		return "InvalidInput"
	case ErrorTypeParsingFailed:
		return "ParsingFailed"
	case ErrorTypeMemoryExhausted:
		return "MemoryExhausted"
	case ErrorTypeTimeout:
		return "Timeout"
	case ErrorTypeConfigInvalid:
		return "ConfigInvalid"
	case ErrorTypeChunkTooLarge:
		return "ChunkTooLarge"
	default:
		return "Unknown"
	}
}

// ChunkerError 分块器错误
type ChunkerError struct {
	Type      ErrorType      `json:"type"`
	Message   string         `json:"message"`
	Context   map[string]any `json:"context"`
	Cause     error          `json:"cause,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Error 实现 error 接口
func (e *ChunkerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *ChunkerError) Unwrap() error {
	return e.Cause
}

// NewChunkerError 创建新的分块器错误
func NewChunkerError(errorType ErrorType, message string, cause error) *ChunkerError {
	return &ChunkerError{
		Type:      errorType,
		Message:   message,
		Context:   make(map[string]any),
		Cause:     cause,
		Timestamp: time.Now(),
	}
}

// WithContext 添加上下文信息
func (e *ChunkerError) WithContext(key string, value any) *ChunkerError {
	e.Context[key] = value
	return e
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	// HandleError 处理错误
	HandleError(err *ChunkerError) error
	// GetErrors 获取所有错误
	GetErrors() []*ChunkerError
	// ClearErrors 清除所有错误
	ClearErrors()
	// HasErrors 检查是否有错误
	HasErrors() bool
}

// DefaultErrorHandler 默认错误处理器
type DefaultErrorHandler struct {
	errors []ChunkerError
	mode   ErrorHandlingMode
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(mode ErrorHandlingMode) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		errors: make([]ChunkerError, 0),
		mode:   mode,
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err *ChunkerError) error {
	// 记录错误
	h.errors = append(h.errors, *err)

	switch h.mode {
	case ErrorModeStrict:
		// 严格模式：立即返回错误
		return err
	case ErrorModePermissive:
		// 宽松模式：记录错误但继续处理
		return nil
	case ErrorModeSilent:
		// 静默模式：忽略错误
		return nil
	default:
		return err
	}
}

// GetErrors 获取所有错误
func (h *DefaultErrorHandler) GetErrors() []*ChunkerError {
	errors := make([]*ChunkerError, len(h.errors))
	for i := range h.errors {
		errors[i] = &h.errors[i]
	}
	return errors
}

// ClearErrors 清除所有错误
func (h *DefaultErrorHandler) ClearErrors() {
	h.errors = h.errors[:0]
}

// HasErrors 检查是否有错误
func (h *DefaultErrorHandler) HasErrors() bool {
	return len(h.errors) > 0
}

// GetErrorsByType 按类型获取错误
func (h *DefaultErrorHandler) GetErrorsByType(errorType ErrorType) []*ChunkerError {
	var filtered []*ChunkerError
	for i := range h.errors {
		if h.errors[i].Type == errorType {
			filtered = append(filtered, &h.errors[i])
		}
	}
	return filtered
}

// GetErrorCount 获取错误数量
func (h *DefaultErrorHandler) GetErrorCount() int {
	return len(h.errors)
}

// GetErrorCountByType 按类型获取错误数量
func (h *DefaultErrorHandler) GetErrorCountByType(errorType ErrorType) int {
	count := 0
	for _, err := range h.errors {
		if err.Type == errorType {
			count++
		}
	}
	return count
}
