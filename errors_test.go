package markdownchunker

import (
	"errors"
	"testing"
	"time"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeInvalidInput, "InvalidInput"},
		{ErrorTypeParsingFailed, "ParsingFailed"},
		{ErrorTypeMemoryExhausted, "MemoryExhausted"},
		{ErrorTypeTimeout, "Timeout"},
		{ErrorTypeConfigInvalid, "ConfigInvalid"},
		{ErrorTypeChunkTooLarge, "ChunkTooLarge"},
		{ErrorType(999), "Unknown"},
	}

	for _, test := range tests {
		result := test.errorType.String()
		if result != test.expected {
			t.Errorf("ErrorType(%d).String() = %s, expected %s", test.errorType, result, test.expected)
		}
	}
}

func TestNewChunkerError(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewChunkerError(ErrorTypeInvalidInput, "test message", cause)

	if err.Type != ErrorTypeInvalidInput {
		t.Errorf("Expected Type to be ErrorTypeInvalidInput, got %v", err.Type)
	}

	if err.Message != "test message" {
		t.Errorf("Expected Message to be 'test message', got %s", err.Message)
	}

	if err.Cause != cause {
		t.Errorf("Expected Cause to be the provided error, got %v", err.Cause)
	}

	if err.Context == nil {
		t.Error("Expected Context to be initialized")
	}

	if time.Since(err.Timestamp) > time.Second {
		t.Error("Expected Timestamp to be recent")
	}
}

func TestChunkerError_Error(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := NewChunkerError(ErrorTypeInvalidInput, "test message", cause)

		expected := "[InvalidInput] test message: underlying error"
		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("without cause", func(t *testing.T) {
		err := NewChunkerError(ErrorTypeInvalidInput, "test message", nil)

		expected := "[InvalidInput] test message"
		if err.Error() != expected {
			t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
		}
	})
}

func TestChunkerError_WithContext(t *testing.T) {
	err := NewChunkerError(ErrorTypeInvalidInput, "test message", nil)
	err.WithContext("key1", "value1").WithContext("key2", 42)

	if err.Context["key1"] != "value1" {
		t.Errorf("Expected Context['key1'] to be 'value1', got %v", err.Context["key1"])
	}

	if err.Context["key2"] != 42 {
		t.Errorf("Expected Context['key2'] to be 42, got %v", err.Context["key2"])
	}
}

func TestChunkerError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewChunkerError(ErrorTypeInvalidInput, "test message", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected Unwrap() to return the cause, got %v", unwrapped)
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	t.Run("strict mode", func(t *testing.T) {
		handler := NewDefaultErrorHandler(ErrorModeStrict)
		err := NewChunkerError(ErrorTypeInvalidInput, "test error", nil)

		result := handler.HandleError(err)
		if result == nil {
			t.Error("Expected HandleError to return error in strict mode")
		}

		if !handler.HasErrors() {
			t.Error("Expected handler to have errors")
		}

		errors := handler.GetErrors()
		if len(errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(errors))
		}
	})

	t.Run("permissive mode", func(t *testing.T) {
		handler := NewDefaultErrorHandler(ErrorModePermissive)
		err := NewChunkerError(ErrorTypeInvalidInput, "test error", nil)

		result := handler.HandleError(err)
		if result != nil {
			t.Errorf("Expected HandleError to return nil in permissive mode, got %v", result)
		}

		if !handler.HasErrors() {
			t.Error("Expected handler to have errors")
		}
	})

	t.Run("silent mode", func(t *testing.T) {
		handler := NewDefaultErrorHandler(ErrorModeSilent)
		err := NewChunkerError(ErrorTypeInvalidInput, "test error", nil)

		result := handler.HandleError(err)
		if result != nil {
			t.Errorf("Expected HandleError to return nil in silent mode, got %v", result)
		}

		if !handler.HasErrors() {
			t.Error("Expected handler to have errors even in silent mode")
		}
	})
}

func TestDefaultErrorHandler_Methods(t *testing.T) {
	handler := NewDefaultErrorHandler(ErrorModePermissive)

	// 添加不同类型的错误
	err1 := NewChunkerError(ErrorTypeInvalidInput, "error 1", nil)
	err2 := NewChunkerError(ErrorTypeParsingFailed, "error 2", nil)
	err3 := NewChunkerError(ErrorTypeInvalidInput, "error 3", nil)

	handler.HandleError(err1)
	handler.HandleError(err2)
	handler.HandleError(err3)

	// 测试 GetErrorCount
	if handler.GetErrorCount() != 3 {
		t.Errorf("Expected 3 errors, got %d", handler.GetErrorCount())
	}

	// 测试 GetErrorCountByType
	invalidInputCount := handler.GetErrorCountByType(ErrorTypeInvalidInput)
	if invalidInputCount != 2 {
		t.Errorf("Expected 2 InvalidInput errors, got %d", invalidInputCount)
	}

	parsingFailedCount := handler.GetErrorCountByType(ErrorTypeParsingFailed)
	if parsingFailedCount != 1 {
		t.Errorf("Expected 1 ParsingFailed error, got %d", parsingFailedCount)
	}

	// 测试 GetErrorsByType
	invalidInputErrors := handler.GetErrorsByType(ErrorTypeInvalidInput)
	if len(invalidInputErrors) != 2 {
		t.Errorf("Expected 2 InvalidInput errors, got %d", len(invalidInputErrors))
	}

	// 测试 ClearErrors
	handler.ClearErrors()
	if handler.HasErrors() {
		t.Error("Expected no errors after ClearErrors")
	}

	if handler.GetErrorCount() != 0 {
		t.Errorf("Expected 0 errors after ClearErrors, got %d", handler.GetErrorCount())
	}
}

func TestChunkDocumentWithErrorHandling(t *testing.T) {
	t.Run("nil content", func(t *testing.T) {
		config := DefaultConfig()
		config.ErrorHandling = ErrorModePermissive

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument(nil)
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		if len(chunks) != 0 {
			t.Errorf("Expected 0 chunks for nil content, got %d", len(chunks))
		}

		if !chunker.HasErrors() {
			t.Error("Expected chunker to have errors")
		}

		errors := chunker.GetErrorsByType(ErrorTypeInvalidInput)
		if len(errors) != 1 {
			t.Errorf("Expected 1 InvalidInput error, got %d", len(errors))
		}
	})

	t.Run("nil content strict mode", func(t *testing.T) {
		config := DefaultConfig()
		config.ErrorHandling = ErrorModeStrict

		chunker := NewMarkdownChunkerWithConfig(config)
		_, err := chunker.ChunkDocument(nil)

		if err == nil {
			t.Error("Expected error in strict mode")
		}

		var chunkerErr *ChunkerError
		if !errors.As(err, &chunkerErr) {
			t.Error("Expected error to be ChunkerError")
		}

		if chunkerErr.Type != ErrorTypeInvalidInput {
			t.Errorf("Expected ErrorTypeInvalidInput, got %v", chunkerErr.Type)
		}
	})

	t.Run("chunk size limit", func(t *testing.T) {
		markdown := `# Very Long Heading That Exceeds The Limit

This is a very long paragraph that definitely exceeds the small chunk size limit we're going to set for testing purposes.`

		config := DefaultConfig()
		config.MaxChunkSize = 20 // 很小的限制
		config.ErrorHandling = ErrorModePermissive

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Errorf("Expected no error in permissive mode, got %v", err)
		}

		if !chunker.HasErrors() {
			t.Error("Expected chunker to have errors due to size limit")
		}

		chunkTooLargeErrors := chunker.GetErrorsByType(ErrorTypeChunkTooLarge)
		if len(chunkTooLargeErrors) == 0 {
			t.Error("Expected ChunkTooLarge errors")
		}

		// 检查块是否被截断
		for _, chunk := range chunks {
			if len(chunk.Content) > 20 {
				t.Errorf("Expected chunk content to be truncated to 20 chars, got %d", len(chunk.Content))
			}
		}
	})

	t.Run("chunk size limit strict mode", func(t *testing.T) {
		markdown := `# Very Long Heading That Exceeds The Limit`

		config := DefaultConfig()
		config.MaxChunkSize = 10 // 非常小的限制
		config.ErrorHandling = ErrorModeStrict

		chunker := NewMarkdownChunkerWithConfig(config)
		_, err := chunker.ChunkDocument([]byte(markdown))

		if err == nil {
			t.Error("Expected error in strict mode with size limit")
		}

		var chunkerErr *ChunkerError
		if errors.As(err, &chunkerErr) {
			if chunkerErr.Type != ErrorTypeChunkTooLarge {
				t.Errorf("Expected ErrorTypeChunkTooLarge, got %v", chunkerErr.Type)
			}
		}
	})
}
