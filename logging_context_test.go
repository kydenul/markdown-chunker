package markdownchunker

import (
	"testing"
	"time"
)

func TestLogContext(t *testing.T) {
	t.Run("NewLogContext creates context with function info", func(t *testing.T) {
		ctx := NewLogContext("TestFunction")

		if ctx.FunctionName != "TestFunction" {
			t.Errorf("Expected function name 'TestFunction', got '%s'", ctx.FunctionName)
		}

		if ctx.FileName == "" {
			t.Error("Expected file name to be set")
		}

		if ctx.LineNumber == 0 {
			t.Error("Expected line number to be set")
		}

		if ctx.Metadata == nil {
			t.Error("Expected metadata map to be initialized")
		}
	})

	t.Run("WithNodeInfo adds node information", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").WithNodeInfo("Heading", 42)

		if ctx.NodeType != "Heading" {
			t.Errorf("Expected node type 'Heading', got '%s'", ctx.NodeType)
		}

		if ctx.NodeID != 42 {
			t.Errorf("Expected node ID 42, got %d", ctx.NodeID)
		}
	})

	t.Run("WithDocumentInfo adds document information", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").WithDocumentInfo(1024, 5)

		if ctx.DocumentSize != 1024 {
			t.Errorf("Expected document size 1024, got %d", ctx.DocumentSize)
		}

		if ctx.ChunkCount != 5 {
			t.Errorf("Expected chunk count 5, got %d", ctx.ChunkCount)
		}
	})

	t.Run("WithProcessTime adds timing information", func(t *testing.T) {
		duration := 100 * time.Millisecond
		ctx := NewLogContext("TestFunction").WithProcessTime(duration)

		if ctx.ProcessTime != duration {
			t.Errorf("Expected process time %v, got %v", duration, ctx.ProcessTime)
		}
	})

	t.Run("WithMetadata adds custom metadata", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").
			WithMetadata("key1", "value1").
			WithMetadata("key2", 42)

		if ctx.Metadata["key1"] != "value1" {
			t.Errorf("Expected metadata key1 to be 'value1', got %v", ctx.Metadata["key1"])
		}

		if ctx.Metadata["key2"] != 42 {
			t.Errorf("Expected metadata key2 to be 42, got %v", ctx.Metadata["key2"])
		}
	})

	t.Run("WithTableInfo adds table-specific metadata", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").WithTableInfo(5, 3, true)

		if ctx.Metadata["table_row_count"] != 5 {
			t.Errorf("Expected table_row_count to be 5, got %v", ctx.Metadata["table_row_count"])
		}

		if ctx.Metadata["table_column_count"] != 3 {
			t.Errorf("Expected table_column_count to be 3, got %v", ctx.Metadata["table_column_count"])
		}

		if ctx.Metadata["table_well_formed"] != true {
			t.Errorf("Expected table_well_formed to be true, got %v", ctx.Metadata["table_well_formed"])
		}
	})

	t.Run("WithListInfo adds list-specific metadata", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").WithListInfo("ordered", 3)

		if ctx.Metadata["list_type"] != "ordered" {
			t.Errorf("Expected list_type to be 'ordered', got %v", ctx.Metadata["list_type"])
		}

		if ctx.Metadata["list_item_count"] != 3 {
			t.Errorf("Expected list_item_count to be 3, got %v", ctx.Metadata["list_item_count"])
		}
	})

	t.Run("WithCodeInfo adds code-specific metadata", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").WithCodeInfo("go", 10, "fenced")

		if ctx.Metadata["code_language"] != "go" {
			t.Errorf("Expected code_language to be 'go', got %v", ctx.Metadata["code_language"])
		}

		if ctx.Metadata["code_line_count"] != 10 {
			t.Errorf("Expected code_line_count to be 10, got %v", ctx.Metadata["code_line_count"])
		}

		if ctx.Metadata["code_block_type"] != "fenced" {
			t.Errorf("Expected code_block_type to be 'fenced', got %v", ctx.Metadata["code_block_type"])
		}
	})

	t.Run("ToLogFields converts context to log fields", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").
			WithNodeInfo("Heading", 1).
			WithDocumentInfo(500, 3).
			WithProcessTime(50*time.Millisecond).
			WithMetadata("custom_key", "custom_value")

		fields := ctx.ToLogFields()

		// Convert fields slice to map for easier testing
		fieldMap := make(map[string]interface{})
		for i := 0; i < len(fields); i += 2 {
			key := fields[i].(string)
			value := fields[i+1]
			fieldMap[key] = value
		}

		// Check required fields
		if fieldMap["function"] != "TestFunction" {
			t.Errorf("Expected function field to be 'TestFunction', got %v", fieldMap["function"])
		}

		if fieldMap["node_type"] != "Heading" {
			t.Errorf("Expected node_type field to be 'Heading', got %v", fieldMap["node_type"])
		}

		if fieldMap["node_id"] != 1 {
			t.Errorf("Expected node_id field to be 1, got %v", fieldMap["node_id"])
		}

		if fieldMap["document_size"] != 500 {
			t.Errorf("Expected document_size field to be 500, got %v", fieldMap["document_size"])
		}

		if fieldMap["chunk_count"] != 3 {
			t.Errorf("Expected chunk_count field to be 3, got %v", fieldMap["chunk_count"])
		}

		if fieldMap["process_time_ms"] != int64(50) {
			t.Errorf("Expected process_time_ms field to be 50, got %v", fieldMap["process_time_ms"])
		}

		if fieldMap["custom_key"] != "custom_value" {
			t.Errorf("Expected custom_key field to be 'custom_value', got %v", fieldMap["custom_key"])
		}
	})

	t.Run("Method chaining works correctly", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").
			WithNodeInfo("Paragraph", 2).
			WithDocumentInfo(1000, 10).
			WithProcessTime(100*time.Millisecond).
			WithContentInfo(200, 180, 25).
			WithPositionInfo(1, 5, 0, 50).
			WithLinksAndImages(2, 1).
			WithMetadata("test", "value")

		// Verify all information is set correctly
		if ctx.FunctionName != "TestFunction" {
			t.Error("Function name not set correctly in chain")
		}
		if ctx.NodeType != "Paragraph" {
			t.Error("Node type not set correctly in chain")
		}
		if ctx.NodeID != 2 {
			t.Error("Node ID not set correctly in chain")
		}
		if ctx.DocumentSize != 1000 {
			t.Error("Document size not set correctly in chain")
		}
		if ctx.ChunkCount != 10 {
			t.Error("Chunk count not set correctly in chain")
		}
		if ctx.ProcessTime != 100*time.Millisecond {
			t.Error("Process time not set correctly in chain")
		}
		if ctx.Metadata["content_length"] != 200 {
			t.Error("Content length not set correctly in chain")
		}
		if ctx.Metadata["text_length"] != 180 {
			t.Error("Text length not set correctly in chain")
		}
		if ctx.Metadata["word_count"] != 25 {
			t.Error("Word count not set correctly in chain")
		}
		if ctx.Metadata["start_line"] != 1 {
			t.Error("Start line not set correctly in chain")
		}
		if ctx.Metadata["end_line"] != 5 {
			t.Error("End line not set correctly in chain")
		}
		if ctx.Metadata["links_count"] != 2 {
			t.Error("Links count not set correctly in chain")
		}
		if ctx.Metadata["images_count"] != 1 {
			t.Error("Images count not set correctly in chain")
		}
		if ctx.Metadata["test"] != "value" {
			t.Error("Custom metadata not set correctly in chain")
		}
	})
}

func TestLogWithContext(t *testing.T) {
	t.Run("logWithContext processes different log levels", func(t *testing.T) {
		// Create a chunker with logging enabled
		config := DefaultConfig()
		config.LogLevel = "DEBUG"
		chunker := NewMarkdownChunkerWithConfig(config)

		ctx := NewLogContext("TestFunction").
			WithNodeInfo("TestNode", 1).
			WithMetadata("test_key", "test_value")

		// Test different log levels - these should not panic
		chunker.logWithContext("debug", "Debug message", ctx)
		chunker.logWithContext("info", "Info message", ctx)
		chunker.logWithContext("warn", "Warn message", ctx)
		chunker.logWithContext("error", "Error message", ctx)
		chunker.logWithContext("unknown", "Unknown level message", ctx) // Should default to info
	})
}

func TestLoggingContextIntegration(t *testing.T) {
	t.Run("ChunkDocument uses logging context", func(t *testing.T) {
		// Create a chunker with debug logging to see all context
		config := DefaultConfig()
		config.LogLevel = "DEBUG"
		chunker := NewMarkdownChunkerWithConfig(config)

		content := []byte("# Test Heading\n\nThis is a test paragraph.")

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("ChunkDocument failed: %v", err)
		}

		if len(chunks) != 2 {
			t.Errorf("Expected 2 chunks, got %d", len(chunks))
		}

		// Verify chunks were created correctly
		if chunks[0].Type != "heading" {
			t.Errorf("Expected first chunk to be heading, got %s", chunks[0].Type)
		}

		if chunks[1].Type != "paragraph" {
			t.Errorf("Expected second chunk to be paragraph, got %s", chunks[1].Type)
		}

		// The logging context functionality is tested implicitly through the log output
		// which we can see in the test output above
	})
}
