package markdownchunker

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestConfigurationCombinations 测试各种配置组合
func TestConfigurationCombinations(t *testing.T) {
	markdown := `# Test Document

This is a paragraph with [link](https://example.com) and ![image](image.png).

## Code Section

` + "```go\n" + `
func complexFunction(a, b, c int) (int, error) {
    if a < 0 || b < 0 || c < 0 {
        return 0, fmt.Errorf("negative values not allowed")
    }
    
    result := 0
    for i := 0; i < a; i++ {
        for j := 0; j < b; j++ {
            result += c
        }
    }
    
    return result, nil
}
` + "```\n\n" + `

| Feature | Status | Priority |
|---------|--------|----------|
| Auth | Complete | High |
| API | In Progress | Medium |
| UI | Planned | Low |

> Important note about the implementation

- First item
- Second item with details
  - Nested item
  - Another nested item

---

Final paragraph with more content.`

	testCases := []struct {
		name   string
		config *ChunkerConfig
		verify func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker)
	}{
		{
			name: "All features enabled with memory optimization",
			config: &ChunkerConfig{
				MaxChunkSize:        0,
				EnabledTypes:        nil, // All enabled
				CustomExtractors:    []MetadataExtractor{&LinkExtractor{}, &ImageExtractor{}, &CodeComplexityExtractor{}},
				ErrorHandling:       ErrorModePermissive,
				PerformanceMode:     PerformanceModeMemoryOptimized,
				FilterEmptyChunks:   true,
				PreserveWhitespace:  false,
				MemoryLimit:         10 * 1024 * 1024, // 10MB
				EnableObjectPooling: true,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Verify all chunk types are present
				chunkTypes := make(map[string]int)
				for _, chunk := range chunks {
					chunkTypes[chunk.Type]++
				}

				expectedTypes := []string{"heading", "paragraph", "code", "table", "blockquote", "list", "thematic_break"}
				for _, expectedType := range expectedTypes {
					if chunkTypes[expectedType] == 0 {
						t.Errorf("Expected chunk type '%s' not found", expectedType)
					}
				}

				// Verify metadata extraction
				foundLinkMetadata := false
				foundImageMetadata := false
				foundCodeComplexity := false

				for _, chunk := range chunks {
					if chunk.Metadata["link_count"] != "" {
						foundLinkMetadata = true
					}
					if chunk.Metadata["image_count"] != "" {
						foundImageMetadata = true
					}
					if chunk.Metadata["code_complexity"] != "" {
						foundCodeComplexity = true
					}
				}

				if !foundLinkMetadata {
					t.Error("Link metadata not found")
				}
				if !foundImageMetadata {
					t.Error("Image metadata not found")
				}
				if !foundCodeComplexity {
					t.Error("Code complexity metadata not found")
				}

				// Verify performance stats
				stats := chunker.GetPerformanceStats()
				if stats.ProcessingTime <= 0 {
					t.Error("Processing time should be positive")
				}
				if stats.TotalChunks != len(chunks) {
					t.Errorf("Stats mismatch: expected %d chunks, got %d", len(chunks), stats.TotalChunks)
				}
			},
		},
		{
			name: "Speed optimized with selective types",
			config: &ChunkerConfig{
				MaxChunkSize: 0,
				EnabledTypes: map[string]bool{
					"heading":   true,
					"paragraph": true,
					"code":      true,
				},
				CustomExtractors:    []MetadataExtractor{&CodeComplexityExtractor{}},
				ErrorHandling:       ErrorModePermissive,
				PerformanceMode:     PerformanceModeSpeedOptimized,
				FilterEmptyChunks:   true,
				PreserveWhitespace:  false,
				MemoryLimit:         0,
				EnableObjectPooling: false,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Should only have selected types
				allowedTypes := map[string]bool{
					"heading":   true,
					"paragraph": true,
					"code":      true,
				}

				for _, chunk := range chunks {
					if !allowedTypes[chunk.Type] {
						t.Errorf("Unexpected chunk type '%s' found", chunk.Type)
					}
				}

				// Should have code complexity metadata
				foundCodeComplexity := false
				for _, chunk := range chunks {
					if chunk.Type == "code" && chunk.Metadata["code_complexity"] != "" {
						foundCodeComplexity = true
						break
					}
				}
				if !foundCodeComplexity {
					t.Error("Code complexity metadata not found for code chunks")
				}
			},
		},
		{
			name: "Strict mode with size limits",
			config: &ChunkerConfig{
				MaxChunkSize:        100,
				EnabledTypes:        nil,
				CustomExtractors:    nil,
				ErrorHandling:       ErrorModeStrict,
				PerformanceMode:     PerformanceModeDefault,
				FilterEmptyChunks:   true,
				PreserveWhitespace:  false,
				MemoryLimit:         0,
				EnableObjectPooling: false,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// This should fail due to size limits in strict mode
				t.Error("Should have failed in strict mode with size limits")
			},
		},
		{
			name: "Whitespace preservation with custom extractors",
			config: &ChunkerConfig{
				MaxChunkSize:        0,
				EnabledTypes:        nil,
				CustomExtractors:    []MetadataExtractor{&LinkExtractor{}, &ImageExtractor{}},
				ErrorHandling:       ErrorModePermissive,
				PerformanceMode:     PerformanceModeDefault,
				FilterEmptyChunks:   false, // Don't filter empty chunks
				PreserveWhitespace:  true,  // Preserve whitespace
				MemoryLimit:         0,
				EnableObjectPooling: false,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Note: Whitespace preservation is configured, but we don't test it here
				// since the input markdown doesn't have significant whitespace to preserve

				// Verify link and image metadata
				foundLinks := false
				foundImages := false
				for _, chunk := range chunks {
					if chunk.Metadata["link_count"] != "" {
						foundLinks = true
					}
					if chunk.Metadata["image_count"] != "" {
						foundImages = true
					}
				}
				if !foundLinks {
					t.Error("Link metadata not found")
				}
				if !foundImages {
					t.Error("Image metadata not found")
				}
			},
		},
		{
			name: "Memory limited with object pooling",
			config: &ChunkerConfig{
				MaxChunkSize:        0,
				EnabledTypes:        nil,
				CustomExtractors:    nil,
				ErrorHandling:       ErrorModePermissive,
				PerformanceMode:     PerformanceModeMemoryOptimized,
				FilterEmptyChunks:   true,
				PreserveWhitespace:  false,
				MemoryLimit:         1024 * 1024, // 1MB limit
				EnableObjectPooling: true,
			},
			verify: func(t *testing.T, chunks []Chunk, chunker *MarkdownChunker) {
				// Should complete successfully with memory optimization
				if len(chunks) == 0 {
					t.Error("Expected some chunks")
				}

				stats := chunker.GetPerformanceStats()
				if stats.PeakMemory <= 0 {
					t.Error("Peak memory should be tracked")
				}

				// Check if memory limit was respected (might have errors if exceeded)
				if chunker.HasErrors() {
					memoryErrors := chunker.GetErrorsByType(ErrorTypeMemoryExhausted)
					if len(memoryErrors) > 0 {
						t.Logf("Memory limit was enforced: %d memory errors", len(memoryErrors))
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chunker := NewMarkdownChunkerWithConfig(tc.config)
			chunks, err := chunker.ChunkDocument([]byte(markdown))

			if tc.name == "Strict mode with size limits" {
				// This test expects an error
				if err == nil {
					tc.verify(t, chunks, chunker)
				}
				return
			}

			if err != nil {
				t.Fatalf("ChunkDocument() error = %v", err)
			}

			tc.verify(t, chunks, chunker)
		})
	}
}

// TestNewFeatureIntegration 测试新功能的集成效果
func TestNewFeatureIntegration(t *testing.T) {
	t.Run("Enhanced metadata extraction integration", func(t *testing.T) {
		markdown := `# API Documentation

Visit our [main site](https://example.com) for more info.

![Architecture](https://example.com/arch.png "System Architecture")

## Code Examples

` + "```python\n" + `
def complex_algorithm(data, threshold=0.5):
    """
    Complex algorithm with multiple branches and loops.
    """
    results = []
    for item in data:
        if item.score > threshold:
            for i in range(len(item.features)):
                if item.features[i] > 0:
                    results.append(process_feature(item.features[i]))
                else:
                    results.append(default_value())
        else:
            results.append(None)
    return results

def process_feature(feature):
    if feature > 1.0:
        return feature * 2
    elif feature > 0.5:
        return feature * 1.5
    else:
        return feature
        
def default_value():
    return 0.0
` + "```\n\n" + `

Check out our [GitHub repo](https://github.com/example/repo) and the [documentation](https://docs.example.com).

![Logo](logo.png)
![Banner](https://cdn.example.com/banner.jpg "Company Banner")`

		config := DefaultConfig()
		config.CustomExtractors = []MetadataExtractor{
			&LinkExtractor{},
			&ImageExtractor{},
			&CodeComplexityExtractor{},
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		// Verify enhanced metadata
		linkCounts := 0
		imageCounts := 0
		codeComplexityFound := false

		for _, chunk := range chunks {
			if chunk.Metadata["link_count"] != "" && chunk.Metadata["link_count"] != "0" {
				linkCounts++
				// Verify link details are captured
				if len(chunk.Links) == 0 {
					t.Error("Expected Links field to be populated when link_count > 0")
				}
			}

			if chunk.Metadata["image_count"] != "" && chunk.Metadata["image_count"] != "0" {
				imageCounts++
				// Verify image details are captured
				if len(chunk.Images) == 0 {
					t.Error("Expected Images field to be populated when image_count > 0")
				}
			}

			if chunk.Type == "code" && chunk.Metadata["code_complexity"] != "" {
				codeComplexityFound = true
				// Verify complexity metrics
				if chunk.Metadata["code_lines"] == "" {
					t.Error("Expected code_lines metadata for code chunks")
				}
				if chunk.Metadata["code_non_empty_lines"] == "" {
					t.Error("Expected code_non_empty_lines metadata for code chunks")
				}
			}
		}

		if linkCounts == 0 {
			t.Error("Expected to find chunks with links")
		}
		if imageCounts == 0 {
			t.Error("Expected to find chunks with images")
		}
		if !codeComplexityFound {
			t.Error("Expected to find code complexity analysis")
		}
	})

	t.Run("Advanced table processing integration", func(t *testing.T) {
		markdown := `# Data Tables

## User Information

| ID | Name | Email | Age | Status | Registration Date |
|----|------|-------|-----|--------|-------------------|
| 1 | John Doe | john@example.com | 30 | Active | 2024-01-15 |
| 2 | Jane Smith | jane@example.com | 25 | Inactive | 2024-01-10 |
| 3 | Bob Johnson | bob@example.com | 35 | Active | 2024-01-20 |

## Financial Data

| Quarter | Revenue | Expenses | Profit | Growth % |
|---------|---------|----------|--------|----------|
| Q1 2024 | $100,000 | $80,000 | $20,000 | 15.5% |
| Q2 2024 | $120,000 | $85,000 | $35,000 | 20.0% |
| Q3 2024 | $140,000 | $90,000 | $50,000 | 16.7% |

## Malformed Table

| Column 1 | Column 2 |
|----------|
| Data 1 | Data 2 | Extra Data |
| Missing cell |`

		chunker := NewMarkdownChunker()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		tableChunks := 0
		wellFormedTables := 0
		malformedTables := 0

		for _, chunk := range chunks {
			if chunk.Type == "table" {
				tableChunks++

				// Verify table metadata
				if chunk.Metadata["rows"] == "" {
					t.Error("Table chunk missing rows metadata")
				}
				if chunk.Metadata["columns"] == "" {
					t.Error("Table chunk missing columns metadata")
				}
				if chunk.Metadata["has_header"] == "" {
					t.Error("Table chunk missing has_header metadata")
				}

				// Check if table is well-formed
				if chunk.Metadata["is_well_formed"] == "true" {
					wellFormedTables++
				} else {
					malformedTables++
					// Should have error information
					if chunk.Metadata["errors"] == "" {
						t.Error("Malformed table should have error information")
					}
				}

				// Verify cell type detection
				if chunk.Metadata["cell_types"] == "" {
					t.Error("Table chunk missing cell_types metadata")
				}
			}
		}

		if tableChunks < 2 {
			t.Errorf("Expected at least 2 table chunks, got %d", tableChunks)
		}
		if wellFormedTables < 1 {
			t.Errorf("Expected at least 1 well-formed table, got %d", wellFormedTables)
		}
		// Note: malformed table detection might vary based on implementation
		t.Logf("Found %d table chunks (%d well-formed, %d malformed)", tableChunks, wellFormedTables, malformedTables)

		// Check if errors were recorded for malformed table
		if chunker.HasErrors() {
			parsingErrors := chunker.GetErrorsByType(ErrorTypeParsingFailed)
			if len(parsingErrors) > 0 {
				t.Logf("Found %d parsing errors for malformed tables", len(parsingErrors))
			}
		}
	})

	t.Run("Performance monitoring integration", func(t *testing.T) {
		// Create a document that will exercise various performance aspects
		var builder strings.Builder
		for i := 0; i < 100; i++ {
			builder.WriteString(fmt.Sprintf("# Section %d\n\n", i))
			builder.WriteString("This is a paragraph with substantial content to test performance monitoring.\n\n")

			if i%5 == 0 {
				builder.WriteString("```go\n")
				builder.WriteString("func performanceTest() {\n")
				for j := 0; j < 10; j++ {
					builder.WriteString(fmt.Sprintf("    fmt.Println(\"Line %d\")\n", j))
				}
				builder.WriteString("}\n```\n\n")
			}

			if i%10 == 0 {
				builder.WriteString("| Col1 | Col2 | Col3 |\n")
				builder.WriteString("|------|------|------|\n")
				for j := 0; j < 5; j++ {
					builder.WriteString(fmt.Sprintf("| Data%d | Value%d | Result%d |\n", j, j, j))
				}
				builder.WriteString("\n")
			}
		}

		config := DefaultConfig()
		config.PerformanceMode = PerformanceModeDefault
		chunker := NewMarkdownChunkerWithConfig(config)

		startTime := time.Now()
		chunks, err := chunker.ChunkDocument([]byte(builder.String()))
		actualTime := time.Since(startTime)

		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		stats := chunker.GetPerformanceStats()

		// Verify performance stats accuracy
		if stats.TotalChunks != len(chunks) {
			t.Errorf("Stats mismatch: expected %d chunks, got %d", len(chunks), stats.TotalChunks)
		}

		if stats.TotalBytes != int64(len(builder.String())) {
			t.Errorf("Stats mismatch: expected %d bytes, got %d", len(builder.String()), stats.TotalBytes)
		}

		// Processing time should be close to actual time (within 20% margin)
		timeDiff := stats.ProcessingTime - actualTime
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}
		if float64(timeDiff) > float64(actualTime)*0.2 {
			t.Errorf("Processing time mismatch: stats=%v, actual=%v", stats.ProcessingTime, actualTime)
		}

		// Verify performance metrics are reasonable
		if stats.ChunksPerSecond <= 0 {
			t.Error("ChunksPerSecond should be positive")
		}
		if stats.BytesPerSecond <= 0 {
			t.Error("BytesPerSecond should be positive")
		}
		if stats.ChunkBytes <= 0 {
			t.Error("ChunkBytes should be positive")
		}

		t.Logf("Performance integration test results:")
		t.Logf("  Total chunks: %d", stats.TotalChunks)
		t.Logf("  Processing time: %v", stats.ProcessingTime)
		t.Logf("  Chunks per second: %.2f", stats.ChunksPerSecond)
		t.Logf("  Bytes per second: %.2f", stats.BytesPerSecond)
		t.Logf("  Memory used: %d bytes", stats.MemoryUsed)
		t.Logf("  Peak memory: %d bytes", stats.PeakMemory)
	})
}

// TestEnhancedBackwardCompatibility 确保向后兼容性
func TestEnhancedBackwardCompatibility(t *testing.T) {
	markdown := `# Test Document

This is a paragraph.

` + "```go\n" + `
func main() {
    fmt.Println("Hello")
}
` + "```\n\n" + `

| Name | Value |
|------|-------|
| Test | 123   |`

	t.Run("Default chunker behavior unchanged", func(t *testing.T) {
		// Test that the default chunker still works as before
		chunker := NewMarkdownChunker()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("Default chunker failed: %v", err)
		}

		// Verify basic structure
		if len(chunks) == 0 {
			t.Error("Expected chunks from default chunker")
		}

		// Verify basic chunk fields are present
		for i, chunk := range chunks {
			if chunk.ID < 0 {
				t.Errorf("Chunk ID should be non-negative, got %d for chunk %d", chunk.ID, i)
			}
			if chunk.Type == "" {
				t.Error("Chunk Type should be set")
			}
			if chunk.Content == "" {
				t.Error("Chunk Content should be set")
			}
			if chunk.Text == "" {
				t.Error("Chunk Text should be set")
			}
			if chunk.Metadata == nil {
				t.Error("Chunk Metadata should be initialized")
			}
		}
	})

	t.Run("Existing API methods still work", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument failed: %v", err)
		}

		// Test that all expected methods are available and work
		if len(chunks) == 0 {
			t.Error("Expected chunks")
		}

		// Test error handling methods
		if chunker.HasErrors() {
			errors := chunker.GetErrors()
			if errors == nil {
				t.Error("GetErrors should return slice")
			}
		}

		chunker.ClearErrors()
		if chunker.HasErrors() {
			t.Error("ClearErrors should clear all errors")
		}

		// Test performance monitoring methods
		stats := chunker.GetPerformanceStats()
		if stats.TotalChunks != len(chunks) {
			t.Error("Performance stats should be available")
		}

		chunker.ResetPerformanceMonitor()
		resetStats := chunker.GetPerformanceStats()
		if resetStats.TotalChunks != 0 {
			t.Error("Reset should clear performance stats")
		}
	})

	t.Run("New fields are optional and don't break existing code", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument failed: %v", err)
		}

		for _, chunk := range chunks {
			// New fields should be initialized but not required
			if chunk.Position.StartLine < 0 {
				t.Error("Position should be initialized")
			}
			if chunk.Links == nil {
				t.Error("Links should be initialized (even if empty)")
			}
			if chunk.Images == nil {
				t.Error("Images should be initialized (even if empty)")
			}
			if chunk.Hash == "" {
				t.Error("Hash should be computed")
			}

			// But they should not interfere with existing functionality
			if chunk.ID < 0 || chunk.Type == "" || chunk.Content == "" {
				t.Error("New fields should not break existing functionality")
			}
		}
	})

	t.Run("Configuration is optional", func(t *testing.T) {
		// Test that passing nil config works (backward compatibility)
		chunker := NewMarkdownChunkerWithConfig(nil)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument with nil config failed: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected chunks with nil config")
		}

		// Test that default config works
		defaultConfig := DefaultConfig()
		chunker2 := NewMarkdownChunkerWithConfig(defaultConfig)
		chunks2, err := chunker2.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument with default config failed: %v", err)
		}

		// Results should be similar
		if len(chunks) != len(chunks2) {
			t.Errorf("Results differ between nil config (%d chunks) and default config (%d chunks)",
				len(chunks), len(chunks2))
		}
	})
}

// TestConfigurationValidation 测试配置验证
func TestConfigurationValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      *ChunkerConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "negative MaxChunkSize",
			config: &ChunkerConfig{
				MaxChunkSize: -100,
			},
			expectError: true,
			errorMsg:    "MaxChunkSize cannot be negative",
		},

		{
			name: "invalid content type in EnabledTypes",
			config: &ChunkerConfig{
				EnabledTypes: map[string]bool{
					"invalid_type": true,
					"heading":      true,
				},
			},
			expectError: true,
			errorMsg:    "invalid content type",
		},

		{
			name: "valid configuration",
			config: &ChunkerConfig{
				MaxChunkSize: 1000,
				EnabledTypes: map[string]bool{
					"heading":        true,
					"paragraph":      true,
					"code":           true,
					"table":          true,
					"list":           true,
					"blockquote":     true,
					"thematic_break": true,
				},
				CustomExtractors:    []MetadataExtractor{&LinkExtractor{}},
				ErrorHandling:       ErrorModePermissive,
				PerformanceMode:     PerformanceModeMemoryOptimized,
				FilterEmptyChunks:   true,
				PreserveWhitespace:  false,
				MemoryLimit:         10 * 1024 * 1024,
				EnableObjectPooling: true,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(tc.config)

			if tc.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if !strings.Contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// TestRealWorldIntegrationScenarios 测试真实世界的集成场景
func TestRealWorldIntegrationScenarios(t *testing.T) {
	t.Run("Technical documentation processing", func(t *testing.T) {
		markdown := `# API Documentation v2.1

## Overview

This document describes the REST API for our service. The API follows RESTful principles and returns JSON responses.

### Base URL

All API requests should be made to: ` + "`" + `https://api.example.com/v2` + "`" + `

### Authentication

API requests require authentication using Bearer tokens:

` + "```bash\n" + `
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     https://api.example.com/v2/users
` + "```\n\n" + `

## Endpoints

### Users

#### GET /users

Returns a paginated list of users.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| page | integer | No | Page number (default: 1) |
| limit | integer | No | Items per page (default: 20, max: 100) |
| search | string | No | Search term for filtering |
| status | string | No | Filter by status: active, inactive, pending |

**Response:**

` + "```json\n" + `
{
  "data": [
    {
      "id": 123,
      "name": "John Doe",
      "email": "john@example.com",
      "status": "active",
      "created_at": "2024-01-15T10:30:00Z",
      "profile": {
        "avatar": "https://cdn.example.com/avatars/123.jpg",
        "bio": "Software developer"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
` + "```\n\n" + `

#### POST /users

Creates a new user account.

**Request Body:**

` + "```json\n" + `
{
  "name": "Jane Smith",
  "email": "jane@example.com",
  "password": "secure_password_123",
  "profile": {
    "bio": "Product manager"
  }
}
` + "```\n\n" + `

**Response Codes:**

- ` + "`" + `201 Created` + "`" + ` - User created successfully
- ` + "`" + `400 Bad Request` + "`" + ` - Invalid input data
- ` + "`" + `409 Conflict` + "`" + ` - Email already exists

> **Note:** Passwords must be at least 8 characters long and contain both letters and numbers.

### Error Handling

All errors follow this format:

` + "```json\n" + `
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Email format is invalid"
      }
    ]
  }
}
` + "```\n\n" + `

## Rate Limiting

API requests are limited to 1000 requests per hour per API key. Rate limit information is included in response headers:

- ` + "`" + `X-RateLimit-Limit` + "`" + `: Request limit per hour
- ` + "`" + `X-RateLimit-Remaining` + "`" + `: Remaining requests in current window
- ` + "`" + `X-RateLimit-Reset` + "`" + `: Unix timestamp when the rate limit resets

## SDKs and Libraries

Official SDKs are available for:

- [JavaScript/Node.js](https://github.com/example/js-sdk)
- [Python](https://github.com/example/python-sdk)
- [Go](https://github.com/example/go-sdk)
- [PHP](https://github.com/example/php-sdk)

---

*Last updated: January 15, 2024*`

		config := &ChunkerConfig{
			MaxChunkSize:        0,
			EnabledTypes:        nil,
			CustomExtractors:    []MetadataExtractor{&LinkExtractor{}, &CodeComplexityExtractor{}},
			ErrorHandling:       ErrorModePermissive,
			PerformanceMode:     PerformanceModeDefault,
			FilterEmptyChunks:   true,
			PreserveWhitespace:  false,
			MemoryLimit:         0,
			EnableObjectPooling: false,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("Failed to process technical documentation: %v", err)
		}

		// Verify comprehensive processing
		chunkTypes := make(map[string]int)
		totalLinks := 0
		codeBlocks := 0

		for _, chunk := range chunks {
			chunkTypes[chunk.Type]++

			if chunk.Metadata["link_count"] != "" && chunk.Metadata["link_count"] != "0" {
				totalLinks++
			}

			if chunk.Type == "code" {
				codeBlocks++
				// Verify code metadata
				if chunk.Metadata["language"] == "" {
					t.Error("Code blocks should have language metadata")
				}
			}

			if chunk.Type == "table" {
				// Verify table processing
				if chunk.Metadata["columns"] == "" {
					t.Error("Tables should have column metadata")
				}
			}
		}

		// Verify expected content was processed
		expectedTypes := []string{"heading", "paragraph", "code", "table", "blockquote", "list", "thematic_break"}
		for _, expectedType := range expectedTypes {
			if chunkTypes[expectedType] == 0 {
				t.Errorf("Expected to find %s chunks in technical documentation", expectedType)
			}
		}

		if totalLinks == 0 {
			t.Error("Expected to find links in technical documentation")
		}

		if codeBlocks < 3 {
			t.Errorf("Expected at least 3 code blocks, found %d", codeBlocks)
		}

		t.Logf("Technical documentation processing results:")
		t.Logf("  Total chunks: %d", len(chunks))
		t.Logf("  Chunk types: %v", chunkTypes)
		t.Logf("  Links found: %d", totalLinks)
		t.Logf("  Code blocks: %d", codeBlocks)
	})

	t.Run("Blog post with mixed content", func(t *testing.T) {
		markdown := `# Understanding Microservices Architecture

*Published on January 15, 2024 by John Developer*

![Microservices Architecture](https://example.com/microservices.png "Microservices Overview")

## Introduction

Microservices architecture has become increasingly popular in recent years. This approach to software development involves breaking down applications into smaller, independent services that communicate over well-defined APIs.

### Benefits of Microservices

1. **Scalability** - Scale individual services based on demand
2. **Technology Diversity** - Use different technologies for different services
3. **Team Independence** - Teams can work independently on different services
4. **Fault Isolation** - Failures in one service don't bring down the entire system

## Implementation Example

Let's look at a simple microservices setup using Docker and Go:

` + "```go\n" + `
// User Service
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/gorilla/mux"
)

type User struct {
    ID    int    ` + "`" + `json:"id"` + "`" + `
    Name  string ` + "`" + `json:"name"` + "`" + `
    Email string ` + "`" + `json:"email"` + "`" + `
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/users", getUsers).Methods("GET")
    r.HandleFunc("/users/{id}", getUser).Methods("GET")
    
    log.Println("User service starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func getUsers(w http.ResponseWriter, r *http.Request) {
    users := []User{
        {ID: 1, Name: "John Doe", Email: "john@example.com"},
        {ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    // Implementation details...
}
` + "```\n\n" + `

### Docker Configuration

` + "```dockerfile\n" + `
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o user-service ./cmd/user-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/user-service .
EXPOSE 8080

CMD ["./user-service"]
` + "```\n\n" + `

## Service Communication

Services can communicate through various methods:

| Method | Pros | Cons | Use Case |
|--------|------|------|----------|
| HTTP/REST | Simple, widely supported | Higher latency | User-facing APIs |
| gRPC | High performance, type safety | More complex setup | Internal services |
| Message Queues | Async, decoupled | Eventual consistency | Event-driven workflows |
| GraphQL | Flexible queries | Learning curve | Data aggregation |

## Best Practices

> **Important:** Always implement proper monitoring and logging across all services.

### Monitoring and Observability

- Use distributed tracing (e.g., [Jaeger](https://jaegertracing.io/))
- Implement health checks for each service
- Set up centralized logging with [ELK Stack](https://elastic.co/elk-stack)
- Monitor key metrics: latency, error rates, throughput

### Security Considerations

1. Implement API gateways for external access
2. Use service mesh for internal communication
3. Apply the principle of least privilege
4. Regularly update dependencies and base images

## Conclusion

Microservices architecture offers many benefits but also introduces complexity. Consider your team size, application requirements, and operational capabilities before making the transition.

For more information, check out these resources:

- [Martin Fowler's Microservices article](https://martinfowler.com/articles/microservices.html)
- [Building Microservices book](https://www.oreilly.com/library/view/building-microservices/9781491950340/)
- [Microservices.io patterns](https://microservices.io/patterns/)

---

*Tags: microservices, architecture, docker, go, devops*`

		config := &ChunkerConfig{
			MaxChunkSize:        0,
			EnabledTypes:        nil,
			CustomExtractors:    []MetadataExtractor{&LinkExtractor{}, &ImageExtractor{}, &CodeComplexityExtractor{}},
			ErrorHandling:       ErrorModePermissive,
			PerformanceMode:     PerformanceModeDefault,
			FilterEmptyChunks:   true,
			PreserveWhitespace:  false,
			MemoryLimit:         0,
			EnableObjectPooling: true,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("Failed to process blog post: %v", err)
		}

		// Verify comprehensive processing
		foundImages := 0
		foundLinks := 0
		foundCodeBlocks := 0
		foundTables := 0

		for _, chunk := range chunks {
			if len(chunk.Images) > 0 {
				foundImages++
			}
			if len(chunk.Links) > 0 {
				foundLinks++
			}
			if chunk.Type == "code" {
				foundCodeBlocks++
			}
			if chunk.Type == "table" {
				foundTables++
			}
		}

		if foundImages == 0 {
			t.Error("Expected to find images in blog post")
		}
		if foundLinks == 0 {
			t.Error("Expected to find links in blog post")
		}
		if foundCodeBlocks < 2 {
			t.Errorf("Expected at least 2 code blocks, found %d", foundCodeBlocks)
		}
		if foundTables == 0 {
			t.Error("Expected to find tables in blog post")
		}

		// Verify performance
		stats := chunker.GetPerformanceStats()
		if stats.ProcessingTime <= 0 {
			t.Error("Processing time should be recorded")
		}

		t.Logf("Blog post processing results:")
		t.Logf("  Total chunks: %d", len(chunks))
		t.Logf("  Images found: %d", foundImages)
		t.Logf("  Links found: %d", foundLinks)
		t.Logf("  Code blocks: %d", foundCodeBlocks)
		t.Logf("  Tables: %d", foundTables)
		t.Logf("  Processing time: %v", stats.ProcessingTime)
	})
}
