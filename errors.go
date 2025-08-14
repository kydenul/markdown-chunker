package markdownchunker

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kydenul/log"
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
	// ErrorTypeStrategyNotFound 策略未找到错误
	ErrorTypeStrategyNotFound
	// ErrorTypeStrategyConfigInvalid 策略配置无效错误
	ErrorTypeStrategyConfigInvalid
	// ErrorTypeStrategyExecutionFailed 策略执行失败错误
	ErrorTypeStrategyExecutionFailed
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
	case ErrorTypeStrategyNotFound:
		return "StrategyNotFound"
	case ErrorTypeStrategyConfigInvalid:
		return "StrategyConfigInvalid"
	case ErrorTypeStrategyExecutionFailed:
		return "StrategyExecutionFailed"
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
	mutex  sync.RWMutex
	logger log.Logger // 日志器实例
}

// NewDefaultErrorHandler 创建默认错误处理器
func NewDefaultErrorHandler(mode ErrorHandlingMode) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		errors: make([]ChunkerError, 0),
		mode:   mode,
		logger: nil, // 将在 SetLogger 中设置
	}
}

// NewDefaultErrorHandlerWithLogger 创建带日志器的默认错误处理器
func NewDefaultErrorHandlerWithLogger(mode ErrorHandlingMode, logger log.Logger) *DefaultErrorHandler {
	return &DefaultErrorHandler{
		errors: make([]ChunkerError, 0),
		mode:   mode,
		logger: logger,
	}
}

// SetLogger 设置日志器
func (h *DefaultErrorHandler) SetLogger(logger log.Logger) {
	h.logger = logger
}

// getStackTrace 获取堆栈跟踪信息
func getStackTrace(skip int) []string {
	var traces []string
	for i := skip; i < skip+10; i++ { // 获取最多10层堆栈
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// 只保留相对路径和函数名
		funcName := fn.Name()
		if idx := strings.LastIndex(funcName, "/"); idx >= 0 {
			funcName = funcName[idx+1:]
		}

		if idx := strings.LastIndex(file, "/"); idx >= 0 {
			file = file[idx+1:]
		}

		traces = append(traces, fmt.Sprintf("%s:%d %s", file, line, funcName))
	}
	return traces
}

// getLogLevelForErrorType 根据错误类型确定日志级别
func getLogLevelForErrorType(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeInvalidInput:
		return "warn" // 输入错误通常是警告级别
	case ErrorTypeParsingFailed:
		return "error" // 解析失败是错误级别
	case ErrorTypeMemoryExhausted:
		return "error" // 内存不足是严重错误
	case ErrorTypeTimeout:
		return "error" // 超时是错误级别
	case ErrorTypeConfigInvalid:
		return "warn" // 配置错误通常是警告级别
	case ErrorTypeChunkTooLarge:
		return "warn" // 块过大通常是警告级别
	case ErrorTypeStrategyNotFound:
		return "error" // 策略未找到是错误级别
	case ErrorTypeStrategyConfigInvalid:
		return "warn" // 策略配置错误通常是警告级别
	case ErrorTypeStrategyExecutionFailed:
		return "error" // 策略执行失败是错误级别
	default:
		return "error" // 未知错误默认为错误级别
	}
}

// HandleError 处理错误
func (h *DefaultErrorHandler) HandleError(err *ChunkerError) error {
	// 记录错误 (线程安全)
	h.mutex.Lock()
	h.errors = append(h.errors, *err)
	h.mutex.Unlock()

	// 如果有日志器，记录详细的错误日志
	if h.logger != nil {
		h.logError(err)
	}

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

// logError 记录详细的错误日志
func (h *DefaultErrorHandler) logError(err *ChunkerError) {
	// 获取堆栈跟踪信息
	stackTrace := getStackTrace(3) // 跳过当前函数和调用链

	// 确定日志级别
	logLevel := getLogLevelForErrorType(err.Type)

	// 构建日志参数
	logArgs := []interface{}{
		"error_type", err.Type.String(),
		"error_message", err.Message,
		"error_timestamp", err.Timestamp,
		"function", "HandleError",
	}

	// 添加上下文信息
	if len(err.Context) > 0 {
		for key, value := range err.Context {
			logArgs = append(logArgs, fmt.Sprintf("context_%s", key), value)
		}
	}

	// 添加原因错误信息
	if err.Cause != nil {
		logArgs = append(logArgs, "cause_error", err.Cause.Error())
	}

	// 添加堆栈跟踪信息（仅在DEBUG级别）
	if len(stackTrace) > 0 {
		logArgs = append(logArgs, "stack_trace", stackTrace)
	}

	// 根据错误类型和日志级别记录日志
	switch logLevel {
	case "warn":
		h.logger.Warnw("错误处理器捕获警告", logArgs...)
	case "error":
		h.logger.Errorw("错误处理器捕获错误", logArgs...)
	default:
		h.logger.Errorw("错误处理器捕获未知级别错误", logArgs...)
	}

	// 如果是严重错误，额外记录错误统计信息
	if err.Type == ErrorTypeMemoryExhausted || err.Type == ErrorTypeTimeout {
		h.mutex.RLock()
		errorCount := len(h.errors)
		h.mutex.RUnlock()

		h.logger.Errorw("检测到严重错误",
			"error_type", err.Type.String(),
			"total_error_count", errorCount,
			"handling_mode", h.getHandlingModeString(),
			"function", "HandleError")
	}
}

// getHandlingModeString 获取错误处理模式的字符串表示
func (h *DefaultErrorHandler) getHandlingModeString() string {
	switch h.mode {
	case ErrorModeStrict:
		return "strict"
	case ErrorModePermissive:
		return "permissive"
	case ErrorModeSilent:
		return "silent"
	default:
		return "unknown"
	}
}

// GetErrors 获取所有错误
func (h *DefaultErrorHandler) GetErrors() []*ChunkerError {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	errors := make([]*ChunkerError, len(h.errors))
	for i := range h.errors {
		errors[i] = &h.errors[i]
	}
	return errors
}

// ClearErrors 清除所有错误
func (h *DefaultErrorHandler) ClearErrors() {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.errors = h.errors[:0]
}

// HasErrors 检查是否有错误
func (h *DefaultErrorHandler) HasErrors() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.errors) > 0
}

// GetErrorsByType 按类型获取错误
func (h *DefaultErrorHandler) GetErrorsByType(errorType ErrorType) []*ChunkerError {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

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
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.errors)
}

// GetErrorCountByType 按类型获取错误数量
func (h *DefaultErrorHandler) GetErrorCountByType(errorType ErrorType) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	count := 0
	for _, err := range h.errors {
		if err.Type == errorType {
			count++
		}
	}
	return count
}
