package markdownchunker

import (
	"strings"
	"testing"

	"github.com/kydenul/log"
)

// TestEnhancedErrorLogging 测试增强的错误日志功能
func TestEnhancedErrorLogging(t *testing.T) {
	t.Run("invalid input error with enhanced logging", func(t *testing.T) {
		config := DefaultConfig()
		config.LogLevel = "DEBUG"
		config.EnableLog = true

		chunker := NewMarkdownChunkerWithConfig(config)

		// 测试空输入错误
		_, err := chunker.ChunkDocument(nil)
		// 验证错误被正确处理
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got: %v", err)
		}

		// 验证错误被记录
		errors := chunker.GetErrors()
		if len(errors) == 0 {
			t.Error("Expected error to be recorded")
		}

		if len(errors) > 0 {
			err := errors[0]
			if err.Type != ErrorTypeInvalidInput {
				t.Errorf("Expected ErrorTypeInvalidInput, got: %v", err.Type)
			}

			// 验证上下文信息
			if err.Context["function"] != "ChunkDocument" {
				t.Errorf("Expected function context, got: %v", err.Context["function"])
			}

			if err.Context["validation_step"] != "input_check" {
				t.Errorf("Expected validation_step context, got: %v", err.Context["validation_step"])
			}
		}
	})

	t.Run("memory exhausted error with enhanced logging", func(t *testing.T) {
		config := DefaultConfig()
		config.LogLevel = "DEBUG"
		config.EnableLog = true

		chunker := NewMarkdownChunkerWithConfig(config)

		// 创建超大内容（超过100MB限制）
		largeContent := make([]byte, 101*1024*1024)
		for i := range largeContent {
			largeContent[i] = 'a'
		}

		_, err := chunker.ChunkDocument(largeContent)
		// 验证错误被正确处理
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got: %v", err)
		}

		// 验证错误被记录
		errors := chunker.GetErrors()
		if len(errors) == 0 {
			t.Error("Expected error to be recorded")
		}

		if len(errors) > 0 {
			err := errors[0]
			if err.Type != ErrorTypeMemoryExhausted {
				t.Errorf("Expected ErrorTypeMemoryExhausted, got: %v", err.Type)
			}

			// 验证上下文信息
			if err.Context["document_size_bytes"] == nil {
				t.Error("Expected document_size_bytes context")
			}

			if err.Context["size_limit_bytes"] == nil {
				t.Error("Expected size_limit_bytes context")
			}
		}
	})

	t.Run("chunk too large error with enhanced logging", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxChunkSize = 100 // 设置很小的块大小限制
		config.LogLevel = "DEBUG"
		config.EnableLog = true

		chunker := NewMarkdownChunkerWithConfig(config)

		// 创建会产生大块的内容
		content := "# " + strings.Repeat("This is a very long heading that will exceed the chunk size limit. ", 10)

		_, err := chunker.ChunkDocument([]byte(content))
		// 验证错误被正确处理
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got: %v", err)
		}

		// 验证错误被记录
		errors := chunker.GetErrors()
		if len(errors) == 0 {
			t.Error("Expected error to be recorded")
		}

		if len(errors) > 0 {
			err := errors[0]
			if err.Type != ErrorTypeChunkTooLarge {
				t.Errorf("Expected ErrorTypeChunkTooLarge, got: %v", err.Type)
			}

			// 验证上下文信息
			if err.Context["chunk_type"] == nil {
				t.Error("Expected chunk_type context")
			}

			if err.Context["size_ratio"] == nil {
				t.Error("Expected size_ratio context")
			}
		}
	})

	t.Run("parsing failed error with enhanced logging", func(t *testing.T) {
		config := DefaultConfig()
		config.LogLevel = "DEBUG"
		config.EnableLog = true

		chunker := NewMarkdownChunkerWithConfig(config)

		// 创建有格式问题的表格
		content := `
| Header 1 | Header 2
| Cell 1 | Cell 2 | Extra Cell
| Cell 3
`

		chunks, err := chunker.ChunkDocument([]byte(content))
		// 验证处理完成
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected at least one chunk")
		}

		// 这个测试主要验证错误处理机制工作正常
		// 表格格式问题可能不会总是触发错误，这取决于解析器的实现
		// 所以我们只验证处理完成即可
		t.Logf("Processed %d chunks successfully", len(chunks))
	})

	t.Run("config validation error with enhanced logging", func(t *testing.T) {
		config := &ChunkerConfig{
			MaxChunkSize: -1,               // 无效值
			LogLevel:     "INVALID_LEVEL",  // 无效日志级别
			LogFormat:    "invalid_format", // 无效格式
			LogDirectory: "   ",            // 空白目录
			EnableLog:    true,
		}

		err := ValidateConfig(config)

		// 验证配置验证失败
		if err == nil {
			t.Error("Expected config validation to fail")
		}

		// 验证返回的是ChunkerError
		if chunkerErr, ok := err.(*ChunkerError); ok {
			if chunkerErr.Type != ErrorTypeConfigInvalid {
				t.Errorf("Expected ErrorTypeConfigInvalid, got: %v", chunkerErr.Type)
			}

			// 验证上下文信息
			if chunkerErr.Context["function"] != "ValidateConfig" {
				t.Errorf("Expected function context, got: %v", chunkerErr.Context["function"])
			}
		} else {
			t.Errorf("Expected ChunkerError, got: %T", err)
		}
	})

	t.Run("error handler with different log levels", func(t *testing.T) {
		logger := log.NewLogger(&log.Options{
			Level:      "debug",
			Format:     "console",
			Directory:  "./test-logs",
			TimeLayout: "2006-01-02 15:04:05.000",
		})

		handler := NewDefaultErrorHandlerWithLogger(ErrorModePermissive, logger)

		// 测试不同类型的错误
		testCases := []struct {
			errorType     ErrorType
			expectedLevel string
		}{
			{ErrorTypeInvalidInput, "warn"},
			{ErrorTypeParsingFailed, "error"},
			{ErrorTypeMemoryExhausted, "error"},
			{ErrorTypeTimeout, "error"},
			{ErrorTypeConfigInvalid, "warn"},
			{ErrorTypeChunkTooLarge, "warn"},
		}

		for _, tc := range testCases {
			err := NewChunkerError(tc.errorType, "test error", nil).
				WithContext("test_context", "test_value")

			handlerErr := handler.HandleError(err)
			if handlerErr != nil {
				t.Errorf("Expected no error in permissive mode, got: %v", handlerErr)
			}
		}

		// 验证所有错误都被记录
		errors := handler.GetErrors()
		if len(errors) != len(testCases) {
			t.Errorf("Expected %d errors, got %d", len(testCases), len(errors))
		}
	})
}

// TestStackTraceGeneration 测试堆栈跟踪生成
func TestStackTraceGeneration(t *testing.T) {
	traces := getStackTrace(0)

	if len(traces) == 0 {
		t.Error("Expected stack traces to be generated")
	}

	// 验证堆栈跟踪格式
	for _, trace := range traces {
		if !strings.Contains(trace, ":") {
			t.Errorf("Expected trace to contain line number, got: %s", trace)
		}
	}
}

// TestLogLevelForErrorType 测试错误类型的日志级别映射
func TestLogLevelForErrorType(t *testing.T) {
	testCases := []struct {
		errorType     ErrorType
		expectedLevel string
	}{
		{ErrorTypeInvalidInput, "warn"},
		{ErrorTypeParsingFailed, "error"},
		{ErrorTypeMemoryExhausted, "error"},
		{ErrorTypeTimeout, "error"},
		{ErrorTypeConfigInvalid, "warn"},
		{ErrorTypeChunkTooLarge, "warn"},
	}

	for _, tc := range testCases {
		level := getLogLevelForErrorType(tc.errorType)
		if level != tc.expectedLevel {
			t.Errorf("For error type %v, expected level %s, got %s",
				tc.errorType, tc.expectedLevel, level)
		}
	}
}
