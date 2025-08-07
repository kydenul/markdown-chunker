package markdownchunker

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestFullIntegration tests the complete workflow with all features enabled
func TestFullIntegration(t *testing.T) {
	// Complex markdown document with all supported elements
	markdown := `# Project Documentation

This is the main documentation for our project with **bold** text and *italic* text.

## Installation Guide

To install the project, follow these steps:

1. Clone the repository from [GitHub](https://github.com/example/repo)
2. Install dependencies
3. Run the setup script

### Prerequisites

You need the following software installed:

- Go 1.19 or later
- Git
- Docker (optional)

## Code Examples

Here's a simple Go example:

` + "```go" + `
package main

import (
    "fmt"
    "log"
)

func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}

func run() error {
    for i := 0; i < 10; i++ {
        fmt.Printf("Hello %d\n", i)
    }
    return nil
}
` + "```" + `

And here's a Python example:

` + "```python" + `
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# Calculate first 10 fibonacci numbers
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
` + "```" + `

## Configuration

The application can be configured using the following table:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| host | string | localhost | Server host |
| port | integer | 8080 | Server port |
| debug | boolean | false | Enable debug mode |
| timeout | decimal | 30.5 | Request timeout in seconds |
| admin_email | email | admin@example.com | Administrator email |
| docs_url | url | https://docs.example.com | Documentation URL |

## Important Notes

> **Warning**: Always backup your data before making changes.
> 
> **Note**: This feature is experimental and may change in future versions.

## API Reference

The following endpoints are available:

- GET /api/users - List all users
- POST /api/users - Create a new user
- PUT /api/users/{id} - Update a user
- DELETE /api/users/{id} - Delete a user

### Response Format

All API responses follow this format:

` + "```json" + `
{
    "status": "success",
    "data": {},
    "message": "Operation completed successfully"
}
` + "```" + `

---

## Troubleshooting

If you encounter issues, check the following:

1. Verify your configuration
2. Check the logs for errors
3. Ensure all dependencies are installed

For more help, visit our [support page](https://support.example.com) or contact us at support@example.com.

![Architecture Diagram](https://example.com/diagram.png "System Architecture")

---

*Last updated: 2024-01-15*`

	// Test with full configuration
	config := &ChunkerConfig{
		MaxChunkSize:        0,   // No limit
		EnabledTypes:        nil, // All types enabled
		CustomExtractors:    []MetadataExtractor{&LinkExtractor{}, &ImageExtractor{}, &CodeComplexityExtractor{}},
		ErrorHandling:       ErrorModePermissive,
		PerformanceMode:     PerformanceModeDefault,
		FilterEmptyChunks:   true,
		PreserveWhitespace:  false,
		MemoryLimit:         0, // No limit
		EnableObjectPooling: false,
	}

	chunker := NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	// Verify we got all expected chunk types
	chunkTypes := make(map[string]int)
	for _, chunk := range chunks {
		chunkTypes[chunk.Type]++
	}

	expectedTypes := []string{"heading", "paragraph", "list", "code", "table", "blockquote", "thematic_break"}
	for _, expectedType := range expectedTypes {
		if chunkTypes[expectedType] == 0 {
			t.Errorf("Expected to find chunks of type '%s', but found none", expectedType)
		}
	}

	// Verify metadata extraction worked
	foundLinks := false
	foundImages := false
	foundCodeComplexity := false

	for _, chunk := range chunks {
		if _, exists := chunk.Metadata["link_count"]; exists {
			foundLinks = true
		}
		if _, exists := chunk.Metadata["image_count"]; exists {
			foundImages = true
		}
		if _, exists := chunk.Metadata["code_complexity"]; exists {
			foundCodeComplexity = true
		}
	}

	if !foundLinks {
		t.Error("Expected to find link metadata")
	}
	if !foundImages {
		t.Error("Expected to find image metadata")
	}
	if !foundCodeComplexity {
		t.Error("Expected to find code complexity metadata")
	}

	// Verify performance monitoring
	stats := chunker.GetPerformanceStats()
	if stats.TotalChunks != len(chunks) {
		t.Errorf("Performance stats mismatch: expected %d chunks, got %d", len(chunks), stats.TotalChunks)
	}

	if stats.ProcessingTime <= 0 {
		t.Error("Processing time should be positive")
	}

	// Verify table processing
	tableFound := false
	for _, chunk := range chunks {
		if chunk.Type == "table" {
			tableFound = true
			// Check advanced table metadata
			if chunk.Metadata["has_header"] != "true" {
				t.Error("Table should have header")
			}
			if chunk.Metadata["is_well_formed"] != "true" {
				t.Errorf("Table should be well-formed, errors: %s", chunk.Metadata["errors"])
			}
			if !strings.Contains(chunk.Metadata["cell_types"], "string") && !strings.Contains(chunk.Metadata["cell_types"], "text") {
				t.Error("Table should contain text cell types")
			}
			break
		}
	}

	if !tableFound {
		t.Error("Expected to find a table chunk")
	}

	t.Logf("Integration test completed successfully:")
	t.Logf("  Total chunks: %d", len(chunks))
	t.Logf("  Chunk types: %v", chunkTypes)
	t.Logf("  Processing time: %v", stats.ProcessingTime)
	t.Logf("  Memory used: %d bytes", stats.MemoryUsed)
}

// TestConfigurationVariations tests different configuration combinations
func TestConfigurationVariations(t *testing.T) {
	markdown := `# Test Document

This is a paragraph with [a link](https://example.com).

` + "```go" + `
func test() {
    fmt.Println("Hello")
}
` + "```" + `

| Col1 | Col2 |
|------|------|
| A    | B    |

> Quote

---`

	tests := []struct {
		name   string
		config *ChunkerConfig
		verify func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker)
	}{
		{
			name: "Only headings enabled",
			config: &ChunkerConfig{
				EnabledTypes: map[string]bool{
					"heading": true,
				},
				ErrorHandling:     ErrorModePermissive,
				FilterEmptyChunks: true,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				for _, chunk := range chunks {
					if chunk.Type != "heading" {
						t.Errorf("Expected only heading chunks, got %s", chunk.Type)
					}
				}
			},
		},
		{
			name: "Strict error handling with size limit",
			config: &ChunkerConfig{
				MaxChunkSize:      50,
				ErrorHandling:     ErrorModeStrict,
				FilterEmptyChunks: true,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Should have failed due to size limit in strict mode
				t.Error("Should have failed in strict mode with size limit")
			},
		},
		{
			name: "Silent error handling",
			config: &ChunkerConfig{
				MaxChunkSize:      10, // Very small limit
				ErrorHandling:     ErrorModeSilent,
				FilterEmptyChunks: true,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Should succeed but have errors recorded
				if !chunker.HasErrors() {
					t.Error("Expected errors to be recorded in silent mode")
				}
				// All chunks should be truncated
				for _, chunk := range chunks {
					if len(chunk.Content) > 10 {
						t.Errorf("Chunk content should be truncated to 10 chars, got %d", len(chunk.Content))
					}
				}
			},
		},
		{
			name: "Custom extractors only",
			config: &ChunkerConfig{
				CustomExtractors: []MetadataExtractor{&LinkExtractor{}},
				ErrorHandling:    ErrorModePermissive,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				foundLinkMetadata := false
				for _, chunk := range chunks {
					if _, exists := chunk.Metadata["link_count"]; exists {
						foundLinkMetadata = true
						break
					}
				}
				if !foundLinkMetadata {
					t.Error("Expected to find link metadata from custom extractor")
				}
			},
		},
		{
			name: "Empty chunks not filtered",
			config: &ChunkerConfig{
				FilterEmptyChunks: false,
				ErrorHandling:     ErrorModePermissive,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Just verify it doesn't crash
				if len(chunks) == 0 {
					t.Error("Expected some chunks")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunkerWithConfig(tt.config)
			chunks, err := chunker.ChunkDocument([]byte(markdown))

			if tt.name == "Strict error handling with size limit" {
				// This test expects an error
				if err == nil {
					tt.verify(t, chunks, chunker)
				}
				return
			}

			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}

			tt.verify(t, chunks, chunker)
		})
	}
}

// TestErrorHandlingScenarios tests various error conditions
func TestErrorHandlingScenarios(t *testing.T) {
	tests := []struct {
		name          string
		content       []byte
		config        *ChunkerConfig
		expectError   bool
		expectedType  ErrorType
		verifyHandler func(t *testing.T, chunker *MarkdownChunker)
	}{
		{
			name:    "Nil content strict mode",
			content: nil,
			config: &ChunkerConfig{
				ErrorHandling: ErrorModeStrict,
			},
			expectError:  true,
			expectedType: ErrorTypeInvalidInput,
		},
		{
			name:    "Nil content permissive mode",
			content: nil,
			config: &ChunkerConfig{
				ErrorHandling: ErrorModePermissive,
			},
			expectError: false,
			verifyHandler: func(t *testing.T, chunker *MarkdownChunker) {
				if !chunker.HasErrors() {
					t.Error("Expected errors to be recorded")
				}
				errors := chunker.GetErrorsByType(ErrorTypeInvalidInput)
				if len(errors) != 1 {
					t.Errorf("Expected 1 InvalidInput error, got %d", len(errors))
				}
			},
		},
		{
			name:    "Very large content",
			content: []byte(strings.Repeat("# Heading\n\nParagraph\n\n", 1000)), // ~10KB (more reasonable)
			config: &ChunkerConfig{
				ErrorHandling: ErrorModePermissive,
			},
			expectError: false,
			verifyHandler: func(t *testing.T, chunker *MarkdownChunker) {
				stats := chunker.GetPerformanceStats()
				if stats.TotalBytes <= 0 {
					t.Error("Should have processed bytes")
				}
				// Verify we can handle large content efficiently
				if stats.ProcessingTime > 5*time.Second {
					t.Errorf("Processing took too long: %v", stats.ProcessingTime)
				}
			},
		},
		{
			name:    "Chunk size limit exceeded",
			content: []byte("# Very Long Heading That Definitely Exceeds The Small Limit We Set"),
			config: &ChunkerConfig{
				MaxChunkSize:  20,
				ErrorHandling: ErrorModeStrict,
			},
			expectError:  true,
			expectedType: ErrorTypeChunkTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunkerWithConfig(tt.config)
			chunks, err := chunker.ChunkDocument(tt.content)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				var chunkerErr *ChunkerError
				if !errors.As(err, &chunkerErr) {
					t.Error("Expected ChunkerError")
					return
				}
				if chunkerErr.Type != tt.expectedType {
					t.Errorf("Expected error type %v, got %v", tt.expectedType, chunkerErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if tt.verifyHandler != nil {
					tt.verifyHandler(t, chunker)
				}
			}

			_ = chunks // Use chunks to avoid unused variable warning
		})
	}
}

// TestPerformanceUnderLoad tests performance with various load conditions
func TestPerformanceUnderLoad(t *testing.T) {
	tests := []struct {
		name        string
		generateDoc func() []byte
		maxTime     time.Duration
		verify      func(t *testing.T, stats PerformanceStats, chunks []Chunk)
	}{
		{
			name: "Many small chunks",
			generateDoc: func() []byte {
				var doc strings.Builder
				for i := 0; i < 1000; i++ {
					doc.WriteString(fmt.Sprintf("# Heading %d\n\nParagraph %d\n\n", i, i))
				}
				return []byte(doc.String())
			},
			maxTime: 5 * time.Second,
			verify: func(t *testing.T, stats PerformanceStats, chunks []Chunk) {
				if len(chunks) < 2000 { // Should have ~2000 chunks (1000 headings + 1000 paragraphs)
					t.Errorf("Expected at least 2000 chunks, got %d", len(chunks))
				}
				if stats.ChunksPerSecond < 100 {
					t.Errorf("Performance too slow: %.2f chunks/sec", stats.ChunksPerSecond)
				}
			},
		},
		{
			name: "Large code blocks",
			generateDoc: func() []byte {
				var doc strings.Builder
				for i := 0; i < 100; i++ {
					doc.WriteString(fmt.Sprintf("# Code Example %d\n\n", i))
					doc.WriteString("```go\n")
					for j := 0; j < 100; j++ {
						doc.WriteString(fmt.Sprintf("// Line %d\nfmt.Println(\"Hello %d\")\n", j, j))
					}
					doc.WriteString("```\n\n")
				}
				return []byte(doc.String())
			},
			maxTime: 10 * time.Second,
			verify: func(t *testing.T, stats PerformanceStats, chunks []Chunk) {
				codeChunks := 0
				for _, chunk := range chunks {
					if chunk.Type == "code" {
						codeChunks++
					}
				}
				if codeChunks < 100 {
					t.Errorf("Expected at least 100 code chunks, got %d", codeChunks)
				}
			},
		},
		{
			name: "Complex tables",
			generateDoc: func() []byte {
				var doc strings.Builder
				for i := 0; i < 50; i++ {
					doc.WriteString(fmt.Sprintf("# Table %d\n\n", i))
					doc.WriteString("| Col1 | Col2 | Col3 | Col4 | Col5 |\n")
					doc.WriteString("|------|------|------|------|------|\n")
					for j := 0; j < 20; j++ {
						doc.WriteString(fmt.Sprintf("| Data%d | %d | %.2f | https://example%d.com | user%d@test.com |\n",
							j, j*10, float64(j)*1.5, j, j))
					}
					doc.WriteString("\n")
				}
				return []byte(doc.String())
			},
			maxTime: 15 * time.Second,
			verify: func(t *testing.T, stats PerformanceStats, chunks []Chunk) {
				tableChunks := 0
				for _, chunk := range chunks {
					if chunk.Type == "table" {
						tableChunks++
						// Verify table metadata
						if chunk.Metadata["rows"] == "" {
							t.Error("Table chunk missing rows metadata")
						}
						if chunk.Metadata["columns"] != "5" {
							t.Errorf("Expected 5 columns, got %s", chunk.Metadata["columns"])
						}
					}
				}
				if tableChunks < 50 {
					t.Errorf("Expected at least 50 table chunks, got %d", tableChunks)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := tt.generateDoc()
			chunker := NewMarkdownChunker()

			start := time.Now()
			chunks, err := chunker.ChunkDocument(doc)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}

			if elapsed > tt.maxTime {
				t.Errorf("Processing took too long: %v (max: %v)", elapsed, tt.maxTime)
			}

			stats := chunker.GetPerformanceStats()
			tt.verify(t, stats, chunks)

			t.Logf("Performance test '%s' completed:", tt.name)
			t.Logf("  Document size: %d bytes", len(doc))
			t.Logf("  Chunks produced: %d", len(chunks))
			t.Logf("  Processing time: %v", elapsed)
			t.Logf("  Chunks per second: %.2f", stats.ChunksPerSecond)
			t.Logf("  Bytes per second: %.2f", stats.BytesPerSecond)
		})
	}
}

// TestMemoryManagement tests memory usage and optimization
func TestMemoryManagement(t *testing.T) {
	// Create a large document
	var largeDoc strings.Builder
	for i := 0; i < 5000; i++ {
		largeDoc.WriteString(fmt.Sprintf("# Section %d\n\n", i))
		largeDoc.WriteString("This is a paragraph with some content that takes up space.\n\n")
		largeDoc.WriteString("```go\n")
		largeDoc.WriteString("func example() {\n")
		largeDoc.WriteString("    return \"Hello World\"\n")
		largeDoc.WriteString("}\n")
		largeDoc.WriteString("```\n\n")
	}

	config := &ChunkerConfig{
		ErrorHandling:       ErrorModePermissive,
		PerformanceMode:     PerformanceModeMemoryOptimized,
		EnableObjectPooling: true,
		MemoryLimit:         50 * 1024 * 1024, // 50MB limit
	}

	chunker := NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(largeDoc.String()))
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	stats := chunker.GetPerformanceStats()

	t.Logf("Memory management test results:")
	t.Logf("  Input size: %d bytes", len(largeDoc.String()))
	t.Logf("  Chunks produced: %d", len(chunks))
	t.Logf("  Peak memory: %d bytes", stats.PeakMemory)
	t.Logf("  Current memory: %d bytes", stats.MemoryUsed)
	t.Logf("  Processing time: %v", stats.ProcessingTime)

	// Verify memory usage is reasonable
	if stats.PeakMemory > 100*1024*1024 { // 100MB
		t.Errorf("Peak memory usage too high: %d bytes", stats.PeakMemory)
	}

	// Verify we got expected number of chunks
	expectedChunks := 5000 * 3 // 5000 sections * (1 heading + 1 paragraph + 1 code block)
	if len(chunks) < expectedChunks-100 || len(chunks) > expectedChunks+100 {
		t.Errorf("Unexpected number of chunks: got %d, expected around %d", len(chunks), expectedChunks)
	}
}

// TestConcurrentProcessing tests thread safety with multiple goroutines
func TestConcurrentProcessing(t *testing.T) {
	markdown := `# Test Document

This is a test paragraph with some content.

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

| Name | Value |
|------|-------|
| Test | 123   |

> This is a quote

---`

	const numGoroutines = 20
	const iterations = 10

	results := make(chan []Chunk, numGoroutines*iterations)
	errors := make(chan error, numGoroutines*iterations)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < iterations; j++ {
				// Each goroutine creates its own chunker instance
				chunker := NewMarkdownChunker()
				chunks, err := chunker.ChunkDocument([]byte(markdown))
				if err != nil {
					errors <- fmt.Errorf("goroutine %d iteration %d: %v", goroutineID, j, err)
					return
				}
				results <- chunks
			}
		}(i)
	}

	// Collect all results
	expectedResults := numGoroutines * iterations
	for i := 0; i < expectedResults; i++ {
		select {
		case err := <-errors:
			t.Fatalf("Concurrent processing failed: %v", err)
		case chunks := <-results:
			// Verify each result has expected structure
			if len(chunks) == 0 {
				t.Error("Expected some chunks from concurrent processing")
			}
			// Verify chunk types are present
			hasHeading := false
			hasParagraph := false
			hasCode := false
			hasTable := false
			for _, chunk := range chunks {
				switch chunk.Type {
				case "heading":
					hasHeading = true
				case "paragraph":
					hasParagraph = true
				case "code":
					hasCode = true
				case "table":
					hasTable = true
				}
			}
			if !hasHeading || !hasParagraph || !hasCode || !hasTable {
				t.Error("Missing expected chunk types in concurrent processing")
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent processing test timed out")
		}
	}

	t.Logf("Concurrent processing test completed successfully:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Iterations per goroutine: %d", iterations)
	t.Logf("  Total operations: %d", expectedResults)
}
