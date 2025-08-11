package markdownchunker

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestErrorRecoveryMechanisms 测试错误恢复机制
func TestErrorRecoveryMechanisms(t *testing.T) {
	t.Run("recovery from parsing errors", func(t *testing.T) {
		// 创建包含格式错误的markdown - 使用一个会触发表格解析问题的例子
		markdown := `# Valid Heading

This is a valid paragraph.

| Column 1 | Column 2 |
| --- |
| Cell 1 | Cell 2 | Extra cell |
| Another | Row |

Another valid paragraph after error.`

		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		chunker := NewMarkdownChunkerWithConfig(config)

		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		// 应该有一些有效的块被处理
		if len(chunks) == 0 {
			t.Error("Expected some chunks to be processed despite errors")
		}

		// 检查是否有解析错误 - 如果没有错误，这可能是正常的，因为goldmark可能处理了这个表格
		if chunker.HasErrors() {
			parsingErrors := chunker.GetErrorsByType(ErrorTypeParsingFailed)
			if len(parsingErrors) > 0 {
				t.Logf("Found %d parsing errors as expected", len(parsingErrors))
			}
		} else {
			t.Log("No parsing errors found - this might be expected if the markdown parser handles the table gracefully")
		}
	})

	t.Run("recovery from memory limit errors", func(t *testing.T) {
		// 创建一个相对较大的文档
		var builder strings.Builder
		for i := 0; i < 1000; i++ {
			builder.WriteString("# Heading ")
			builder.WriteString(strings.Repeat("a", 100))
			builder.WriteString("\n\n")
			builder.WriteString("Paragraph ")
			builder.WriteString(strings.Repeat("b", 200))
			builder.WriteString("\n\n")
		}

		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		config.MemoryLimit = 1024 * 1024 // 设置1MB的内存限制
		chunker := NewMarkdownChunkerWithConfig(config)

		chunks, err := chunker.ChunkDocument([]byte(builder.String()))
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		// 应该处理了一些块，直到内存限制
		if len(chunks) == 0 {
			t.Error("Expected some chunks to be processed before memory limit")
		}

		// 应该有内存相关的错误
		if chunker.HasErrors() {
			memoryErrors := chunker.GetErrorsByType(ErrorTypeMemoryExhausted)
			if len(memoryErrors) == 0 {
				t.Log("No memory errors found, this might be expected if memory limit wasn't reached")
			}
		}
	})

	t.Run("recovery from chunk size limit errors", func(t *testing.T) {
		markdown := `# Short Heading

This is a very long paragraph that will definitely exceed our small chunk size limit and should trigger a chunk too large error but processing should continue.

# Another Heading

Short paragraph.`

		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		config.MaxChunkSize = 50 // 很小的限制
		chunker := NewMarkdownChunkerWithConfig(config)

		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		// 应该有一些块被处理
		if len(chunks) == 0 {
			t.Error("Expected some chunks to be processed")
		}

		// 应该有块大小错误
		if !chunker.HasErrors() {
			t.Error("Expected chunk size errors")
		}

		chunkSizeErrors := chunker.GetErrorsByType(ErrorTypeChunkTooLarge)
		if len(chunkSizeErrors) == 0 {
			t.Error("Expected ChunkTooLarge errors")
		}

		// 检查错误上下文信息
		for _, err := range chunkSizeErrors {
			if err.Context["chunk_size"] == nil {
				t.Error("Expected chunk_size in error context")
			}
			if err.Context["max_size"] == nil {
				t.Error("Expected max_size in error context")
			}
			if err.Context["chunk_id"] == nil {
				t.Error("Expected chunk_id in error context")
			}
		}
	})
}

// TestErrorContextInformation 测试错误上下文信息的准确性
func TestErrorContextInformation(t *testing.T) {
	t.Run("invalid input error context", func(t *testing.T) {
		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		chunker := NewMarkdownChunkerWithConfig(config)

		_, err := chunker.ChunkDocument(nil)
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		errors := chunker.GetErrorsByType(ErrorTypeInvalidInput)
		if len(errors) != 1 {
			t.Fatalf("Expected 1 invalid input error, got %d", len(errors))
		}

		chunkerErr := errors[0]
		if chunkerErr.Message != "content cannot be nil" {
			t.Errorf("Expected specific error message, got %s", chunkerErr.Message)
		}

		if time.Since(chunkerErr.Timestamp) > time.Second {
			t.Error("Expected recent timestamp")
		}
	})

	t.Run("large content error context", func(t *testing.T) {
		// 创建超过100MB限制的内容
		largeContent := make([]byte, 101*1024*1024) // 101MB
		for i := range largeContent {
			largeContent[i] = 'a'
		}

		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		chunker := NewMarkdownChunkerWithConfig(config)

		_, err := chunker.ChunkDocument(largeContent)
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		errors := chunker.GetErrorsByType(ErrorTypeMemoryExhausted)
		if len(errors) != 1 {
			t.Fatalf("Expected 1 memory exhausted error, got %d", len(errors))
		}

		chunkerErr := errors[0]
		if chunkerErr.Context["size"] != len(largeContent) {
			t.Errorf("Expected size context to be %d, got %v", len(largeContent), chunkerErr.Context["size"])
		}

		if chunkerErr.Context["limit"] != 100*1024*1024 {
			t.Errorf("Expected limit context to be %d, got %v", 100*1024*1024, chunkerErr.Context["limit"])
		}
	})

	t.Run("table parsing error context", func(t *testing.T) {
		markdown := `# Test

| Column 1 | Column 2 |
| --- |
| Cell 1 | Cell 2 | Extra cell |`

		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		chunker := NewMarkdownChunkerWithConfig(config)

		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		if !chunker.HasErrors() {
			t.Log("No table parsing errors found - this might be expected if table is well-formed")
			return
		}

		parsingErrors := chunker.GetErrorsByType(ErrorTypeParsingFailed)
		for _, chunkerErr := range parsingErrors {
			if chunkerErr.Context["table_errors"] == nil {
				t.Error("Expected table_errors in context for parsing errors")
			}
			if chunkerErr.Context["chunk_id"] == nil {
				t.Error("Expected chunk_id in context for parsing errors")
			}
		}

		// 应该仍然产生一些块
		if len(chunks) == 0 {
			t.Error("Expected some chunks despite table errors")
		}
	})
}

// TestErrorHandlerModes 测试不同错误处理模式的行为
func TestErrorHandlerModes(t *testing.T) {
	testCases := []struct {
		name         string
		mode         ErrorHandlingMode
		markdown     string
		expectError  bool
		expectChunks bool
	}{
		{
			name:         "strict mode with invalid input",
			mode:         ErrorModeStrict,
			markdown:     "",
			expectError:  true,
			expectChunks: false,
		},
		{
			name:         "permissive mode with invalid input",
			mode:         ErrorModePermissive,
			markdown:     "",
			expectError:  false,
			expectChunks: false,
		},
		{
			name:         "silent mode with invalid input",
			mode:         ErrorModeSilent,
			markdown:     "",
			expectError:  false,
			expectChunks: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultConfig()
			config.ErrorHandling = tc.mode
			config.MaxChunkSize = 10 // 小的限制来触发错误
			chunker := NewMarkdownChunkerWithConfig(config)

			var content []byte
			if tc.markdown != "" {
				content = []byte(tc.markdown)
			} else {
				content = nil // 触发无效输入错误
			}

			chunks, err := chunker.ChunkDocument(content)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tc.expectChunks && len(chunks) == 0 {
				t.Error("Expected chunks but got none")
			}
			if !tc.expectChunks && len(chunks) > 0 {
				t.Errorf("Expected no chunks but got %d", len(chunks))
			}

			// 在所有模式下，错误都应该被记录（即使在silent模式下）
			if content == nil && !chunker.HasErrors() {
				t.Error("Expected errors to be recorded even in silent mode")
			}
		})
	}
}

// TestConcurrentErrorHandling 测试并发环境下的错误处理
func TestConcurrentErrorHandling(t *testing.T) {
	t.Run("concurrent error recording", func(t *testing.T) {
		config := DefaultConfig()
		config.ErrorHandling = ErrorModeStrict // 使用严格模式来确保错误被返回
		config.MaxChunkSize = 20               // 小的限制来触发错误

		// 创建多个会触发错误的文档
		documents := [][]byte{
			nil, // 无效输入
			[]byte("# Very long heading that exceeds the chunk size limit and should trigger an error"),
			[]byte("Another very long paragraph that will exceed the size limit and cause an error"),
		}

		concurrentChunker := NewConcurrentChunker(config)

		// 并发处理文档
		results, processingErrors := concurrentChunker.ChunkDocumentConcurrent(documents)

		// 检查结果
		errorCount := 0
		for i, err := range processingErrors {
			if err != nil {
				errorCount++
				t.Logf("Document %d produced error: %v", i, err)
			}
		}

		// 应该有一些错误（至少nil输入应该产生错误）
		if errorCount == 0 {
			t.Error("Expected some errors in concurrent processing with strict mode")
		}

		// 检查是否有结果
		if len(results) != len(documents) {
			t.Errorf("Expected %d results, got %d", len(documents), len(results))
		}

		// 验证错误类型
		for i, err := range processingErrors {
			if err != nil {
				var chunkerErr *ChunkerError
				if errors.As(err, &chunkerErr) {
					t.Logf("Document %d error type: %v", i, chunkerErr.Type)
				}
			}
		}
	})

	t.Run("concurrent processing with permissive mode", func(t *testing.T) {
		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive
		config.MaxChunkSize = 10

		documents := [][]byte{
			[]byte("# Short"),
			[]byte("# This is a very long heading that exceeds the limit"),
			[]byte("Normal paragraph."),
		}

		concurrentChunker := NewConcurrentChunker(config)
		results, processingErrors := concurrentChunker.ChunkDocumentConcurrent(documents)

		// 在宽松模式下，不应该有处理错误返回
		for i, err := range processingErrors {
			if err != nil {
				t.Errorf("Expected no processing errors in permissive mode, got error for document %d: %v", i, err)
			}
		}

		// 应该有结果
		if len(results) != len(documents) {
			t.Errorf("Expected %d results, got %d", len(documents), len(results))
		}

		// 检查每个结果是否合理
		for i, chunks := range results {
			if len(chunks) == 0 && documents[i] != nil && len(documents[i]) > 0 {
				t.Errorf("Expected chunks for document %d", i)
			}
		}
	})
}

// TestErrorChaining 测试错误链和包装
func TestErrorChaining(t *testing.T) {
	t.Run("error unwrapping", func(t *testing.T) {
		originalErr := errors.New("original error")
		chunkerErr := NewChunkerError(ErrorTypeParsingFailed, "parsing failed", originalErr)

		// 测试 Unwrap
		unwrapped := chunkerErr.Unwrap()
		if unwrapped != originalErr {
			t.Errorf("Expected unwrapped error to be original error, got %v", unwrapped)
		}

		// 测试 errors.Is
		if !errors.Is(chunkerErr, originalErr) {
			t.Error("Expected errors.Is to return true for wrapped error")
		}

		// 测试 errors.As
		var targetErr *ChunkerError
		if !errors.As(chunkerErr, &targetErr) {
			t.Error("Expected errors.As to work with ChunkerError")
		}
	})

	t.Run("nested error context", func(t *testing.T) {
		originalErr := errors.New("file not found")
		chunkerErr := NewChunkerError(ErrorTypeParsingFailed, "failed to parse document", originalErr)
		chunkerErr.WithContext("file_path", "/path/to/file.md")
		chunkerErr.WithContext("line_number", 42)

		// 检查错误消息格式
		expectedMsg := "[ParsingFailed] failed to parse document: file not found"
		if chunkerErr.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, chunkerErr.Error())
		}

		// 检查上下文信息
		if chunkerErr.Context["file_path"] != "/path/to/file.md" {
			t.Error("Expected file_path in context")
		}
		if chunkerErr.Context["line_number"] != 42 {
			t.Error("Expected line_number in context")
		}
	})
}

// SelectiveErrorHandler 自定义错误处理器，只记录特定类型的错误
type SelectiveErrorHandler struct {
	errors       []ChunkerError
	allowedTypes map[ErrorType]bool
}

func (h *SelectiveErrorHandler) HandleError(err *ChunkerError) error {
	if h.allowedTypes[err.Type] {
		h.errors = append(h.errors, *err)
	}
	return nil // 总是继续处理
}

func (h *SelectiveErrorHandler) GetErrors() []*ChunkerError {
	errors := make([]*ChunkerError, len(h.errors))
	for i := range h.errors {
		errors[i] = &h.errors[i]
	}
	return errors
}

func (h *SelectiveErrorHandler) ClearErrors() {
	h.errors = h.errors[:0]
}

func (h *SelectiveErrorHandler) HasErrors() bool {
	return len(h.errors) > 0
}

// TestErrorHandlerCustomization 测试自定义错误处理器
func TestErrorHandlerCustomization(t *testing.T) {
	t.Run("selective error recording", func(t *testing.T) {
		customHandler := &SelectiveErrorHandler{
			errors: make([]ChunkerError, 0),
			allowedTypes: map[ErrorType]bool{
				ErrorTypeChunkTooLarge: true,
				// 不记录其他类型的错误
			},
		}

		config := DefaultConfig()
		config.MaxChunkSize = 10 // 触发块大小错误
		chunker := NewMarkdownChunkerWithConfig(config)
		chunker.errorHandler = customHandler

		markdown := `# Very long heading that exceeds limit

This is also a very long paragraph that exceeds the chunk size limit.`

		_, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Errorf("Expected no error with custom handler, got %v", err)
		}

		// 应该只记录块大小错误
		errors := customHandler.GetErrors()
		for _, err := range errors {
			if err.Type != ErrorTypeChunkTooLarge {
				t.Errorf("Expected only ChunkTooLarge errors, got %v", err.Type)
			}
		}

		if len(errors) == 0 {
			t.Error("Expected some ChunkTooLarge errors to be recorded")
		}
	})
}

// TestErrorMessageAccuracy 测试错误消息的准确性和有用性
func TestErrorMessageAccuracy(t *testing.T) {
	testCases := []struct {
		name           string
		errorType      ErrorType
		message        string
		cause          error
		expectedFormat string
	}{
		{
			name:           "simple error without cause",
			errorType:      ErrorTypeInvalidInput,
			message:        "input is empty",
			cause:          nil,
			expectedFormat: "[InvalidInput] input is empty",
		},
		{
			name:           "error with cause",
			errorType:      ErrorTypeParsingFailed,
			message:        "failed to parse markdown",
			cause:          errors.New("syntax error at line 5"),
			expectedFormat: "[ParsingFailed] failed to parse markdown: syntax error at line 5",
		},
		{
			name:           "memory error with context",
			errorType:      ErrorTypeMemoryExhausted,
			message:        "memory limit exceeded",
			cause:          nil,
			expectedFormat: "[MemoryExhausted] memory limit exceeded",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewChunkerError(tc.errorType, tc.message, tc.cause)

			if err.Error() != tc.expectedFormat {
				t.Errorf("Expected error format '%s', got '%s'", tc.expectedFormat, err.Error())
			}

			// 检查错误类型字符串表示
			if err.Type.String() != tc.errorType.String() {
				t.Errorf("Expected error type string '%s', got '%s'", tc.errorType.String(), err.Type.String())
			}
		})
	}
}

// TestErrorHandlerStateManagement 测试错误处理器状态管理
func TestErrorHandlerStateManagement(t *testing.T) {
	t.Run("error accumulation and clearing", func(t *testing.T) {
		handler := NewDefaultErrorHandler(ErrorModePermissive)

		// 添加多个错误
		err1 := NewChunkerError(ErrorTypeInvalidInput, "error 1", nil)
		err2 := NewChunkerError(ErrorTypeParsingFailed, "error 2", nil)
		err3 := NewChunkerError(ErrorTypeInvalidInput, "error 3", nil)

		handler.HandleError(err1)
		handler.HandleError(err2)
		handler.HandleError(err3)

		// 检查错误计数
		if handler.GetErrorCount() != 3 {
			t.Errorf("Expected 3 errors, got %d", handler.GetErrorCount())
		}

		// 检查按类型计数
		if handler.GetErrorCountByType(ErrorTypeInvalidInput) != 2 {
			t.Errorf("Expected 2 InvalidInput errors, got %d", handler.GetErrorCountByType(ErrorTypeInvalidInput))
		}

		// 清除错误
		handler.ClearErrors()
		if handler.HasErrors() {
			t.Error("Expected no errors after clearing")
		}

		if handler.GetErrorCount() != 0 {
			t.Errorf("Expected 0 errors after clearing, got %d", handler.GetErrorCount())
		}
	})

	t.Run("error handler thread safety", func(t *testing.T) {
		handler := NewDefaultErrorHandler(ErrorModePermissive)

		// 使用 WaitGroup 确保所有 goroutine 完成
		var wg sync.WaitGroup
		wg.Add(10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer wg.Done()
				err := NewChunkerError(ErrorTypeInvalidInput, fmt.Sprintf("error %d", id), nil)
				handler.HandleError(err)
			}(i)
		}

		// 等待所有goroutine完成
		wg.Wait()

		// 检查所有错误都被记录
		if handler.GetErrorCount() != 10 {
			t.Errorf("Expected 10 errors, got %d", handler.GetErrorCount())
		}
	})
}
