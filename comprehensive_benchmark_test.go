package markdownchunker

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// BenchmarkMemoryUsage_SmallDocuments 测试小文档的内存使用
func BenchmarkMemoryUsage_SmallDocuments(b *testing.B) {
	chunker := NewMarkdownChunker()
	content := []byte(`# Small Document

This is a small paragraph with some content.

## Section

Another paragraph here.

- List item 1
- List item 2`)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
		chunker.ResetPerformanceMonitor()
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkMemoryUsage_MediumDocuments 测试中等文档的内存使用
func BenchmarkMemoryUsage_MediumDocuments(b *testing.B) {
	chunker := NewMarkdownChunker()

	// 创建中等大小的文档
	var builder strings.Builder
	for i := 0; i < 50; i++ {
		builder.WriteString(fmt.Sprintf("# Heading %d\n\n", i))
		builder.WriteString("This is a paragraph with some content. ")
		builder.WriteString("It contains multiple sentences to make it more realistic. ")
		builder.WriteString("We want to test memory usage with medium-sized documents.\n\n")

		if i%5 == 0 {
			builder.WriteString("```go\n")
			builder.WriteString("func example() {\n")
			builder.WriteString("    fmt.Println(\"Hello, World!\")\n")
			builder.WriteString("}\n```\n\n")
		}

		if i%7 == 0 {
			builder.WriteString("| Column 1 | Column 2 |\n")
			builder.WriteString("|----------|----------|\n")
			builder.WriteString("| Cell 1   | Cell 2   |\n")
			builder.WriteString("| Cell 3   | Cell 4   |\n\n")
		}
	}
	content := []byte(builder.String())

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
		chunker.ResetPerformanceMonitor()
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkMemoryUsage_LargeDocuments 测试大文档的内存使用
func BenchmarkMemoryUsage_LargeDocuments(b *testing.B) {
	chunker := NewMarkdownChunker()

	// 创建大文档
	var builder strings.Builder
	for i := 0; i < 500; i++ {
		builder.WriteString(fmt.Sprintf("# Heading %d\n\n", i))
		builder.WriteString("This is a paragraph with substantial content. ")
		builder.WriteString("It contains multiple sentences to simulate real-world documents. ")
		builder.WriteString("We want to test memory usage and performance with large documents. ")
		builder.WriteString("This paragraph should be long enough to trigger various processing paths.\n\n")

		if i%10 == 0 {
			builder.WriteString("```python\n")
			builder.WriteString("def example_function():\n")
			builder.WriteString("    print('This is a code block')\n")
			builder.WriteString("    for i in range(10):\n")
			builder.WriteString("        print(f'Item {i}')\n")
			builder.WriteString("```\n\n")
		}

		if i%15 == 0 {
			builder.WriteString("| Column A | Column B | Column C |\n")
			builder.WriteString("|----------|----------|----------|\n")
			for j := 0; j < 5; j++ {
				builder.WriteString(fmt.Sprintf("| Data %d   | Value %d  | Result %d |\n", j, j*2, j*3))
			}
			builder.WriteString("\n")
		}

		if i%20 == 0 {
			builder.WriteString("- List item with detailed description\n")
			builder.WriteString("- Another list item with more content\n")
			builder.WriteString("- Third item with even more detailed information\n")
			builder.WriteString("  - Nested item 1\n")
			builder.WriteString("  - Nested item 2\n\n")
		}
	}
	content := []byte(builder.String())

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
		chunker.ResetPerformanceMonitor()
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkProcessingSpeed_SmallDocuments 测试小文档的处理速度
func BenchmarkProcessingSpeed_SmallDocuments(b *testing.B) {
	chunker := NewMarkdownChunker()
	content := []byte(`# Small Document

This is a small paragraph.

## Section

Another paragraph.`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
		if len(chunks) == 0 {
			b.Fatal("Expected chunks")
		}
		chunker.ResetPerformanceMonitor()
	}
}

// BenchmarkProcessingSpeed_MediumDocuments 测试中等文档的处理速度
func BenchmarkProcessingSpeed_MediumDocuments(b *testing.B) {
	chunker := NewMarkdownChunker()

	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("# Heading %d\n\n", i))
		builder.WriteString("This is a paragraph with content. ")
		builder.WriteString("It has multiple sentences for testing.\n\n")

		if i%5 == 0 {
			builder.WriteString("```go\nfunc test() {}\n```\n\n")
		}
	}
	content := []byte(builder.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
		if len(chunks) == 0 {
			b.Fatal("Expected chunks")
		}
		chunker.ResetPerformanceMonitor()
	}
}

// BenchmarkProcessingSpeed_LargeDocuments 测试大文档的处理速度
func BenchmarkProcessingSpeed_LargeDocuments(b *testing.B) {
	chunker := NewMarkdownChunker()

	var builder strings.Builder
	for i := 0; i < 1000; i++ {
		builder.WriteString(fmt.Sprintf("# Heading %d\n\n", i))
		builder.WriteString("This is a substantial paragraph with meaningful content. ")
		builder.WriteString("It contains multiple sentences to simulate real documents.\n\n")

		if i%10 == 0 {
			builder.WriteString("```javascript\n")
			builder.WriteString("function example() {\n")
			builder.WriteString("  console.log('Hello');\n")
			builder.WriteString("}\n```\n\n")
		}
	}
	content := []byte(builder.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatal(err)
		}
		if len(chunks) == 0 {
			b.Fatal("Expected chunks")
		}
		chunker.ResetPerformanceMonitor()
	}
}

// BenchmarkLargeDocumentProcessing 测试超大文档处理
func BenchmarkLargeDocumentProcessing(b *testing.B) {
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			chunker := NewMarkdownChunker()

			var builder strings.Builder
			for i := 0; i < size; i++ {
				builder.WriteString(fmt.Sprintf("# Section %d\n\n", i))
				builder.WriteString("This is a comprehensive paragraph with detailed content. ")
				builder.WriteString("It simulates real-world documentation with substantial text. ")
				builder.WriteString("The content is designed to test performance under load.\n\n")

				if i%20 == 0 {
					builder.WriteString("```python\n")
					builder.WriteString("def process_data(data):\n")
					builder.WriteString("    result = []\n")
					builder.WriteString("    for item in data:\n")
					builder.WriteString("        result.append(transform(item))\n")
					builder.WriteString("    return result\n")
					builder.WriteString("```\n\n")
				}

				if i%30 == 0 {
					builder.WriteString("| Feature | Description | Status |\n")
					builder.WriteString("|---------|-------------|--------|\n")
					builder.WriteString("| Feature A | Detailed description | Active |\n")
					builder.WriteString("| Feature B | Another description | Pending |\n")
					builder.WriteString("| Feature C | Third description | Complete |\n\n")
				}
			}
			content := []byte(builder.String())

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				chunks, err := chunker.ChunkDocument(content)
				if err != nil {
					b.Fatal(err)
				}
				if len(chunks) == 0 {
					b.Fatal("Expected chunks")
				}
				chunker.ResetPerformanceMonitor()
			}
		})
	}
}

// BenchmarkConcurrentProcessing 测试并发处理性能
func BenchmarkConcurrentProcessing(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8}

	// 创建测试文档
	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(fmt.Sprintf("# Document Section %d\n\n", i))
		builder.WriteString("This is content for concurrent processing testing. ")
		builder.WriteString("Each document should be processed independently.\n\n")
	}
	content := []byte(builder.String())

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			config := DefaultConfig()
			concurrentChunker := NewConcurrentChunker(config)

			// 创建多个文档副本
			documents := make([][]byte, concurrency)
			for i := 0; i < concurrency; i++ {
				documents[i] = content
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, errors := concurrentChunker.ChunkDocumentConcurrent(documents)

				// 验证结果
				for j, err := range errors {
					if err != nil {
						b.Fatalf("Document %d failed: %v", j, err)
					}
				}

				for j, chunks := range results {
					if len(chunks) == 0 {
						b.Fatalf("Document %d produced no chunks", j)
					}
				}
			}
		})
	}
}

// BenchmarkMemoryOptimization 测试内存优化功能
func BenchmarkMemoryOptimization(b *testing.B) {
	configs := []struct {
		name     string
		pooling  bool
		memLimit int64
	}{
		{"NoOptimization", false, 0},
		{"WithPooling", true, 0},
		{"WithMemoryLimit", false, 10 * 1024 * 1024}, // 10MB
		{"FullOptimization", true, 10 * 1024 * 1024},
	}

	// 创建测试内容
	var builder strings.Builder
	for i := 0; i < 200; i++ {
		builder.WriteString(fmt.Sprintf("# Section %d\n\n", i))
		builder.WriteString("This is content for memory optimization testing. ")
		builder.WriteString("We want to see how different optimization settings affect performance.\n\n")

		if i%10 == 0 {
			builder.WriteString("```go\n")
			builder.WriteString("func optimizationTest() {\n")
			builder.WriteString("    // Test memory optimization\n")
			builder.WriteString("}\n```\n\n")
		}
	}
	content := []byte(builder.String())

	for _, config := range configs {
		b.Run(config.name, func(b *testing.B) {
			chunkerConfig := DefaultConfig()
			chunkerConfig.EnableObjectPooling = config.pooling
			chunkerConfig.MemoryLimit = config.memLimit

			chunker := NewMarkdownChunkerWithConfig(chunkerConfig)

			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				chunks, err := chunker.ChunkDocument(content)
				if err != nil {
					b.Fatal(err)
				}
				if len(chunks) == 0 {
					b.Fatal("Expected chunks")
				}
				chunker.ResetPerformanceMonitor()
			}

			runtime.GC()
			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
		})
	}
}

// BenchmarkDifferentContentTypes 测试不同内容类型的处理性能
func BenchmarkDifferentContentTypes(b *testing.B) {
	contentTypes := map[string]string{
		"Headings":   generateHeadingContent(100),
		"Paragraphs": generateParagraphContent(100),
		"CodeBlocks": generateCodeBlockContent(50),
		"Tables":     generateTableContent(20),
		"Lists":      generateListContent(50),
		"Mixed":      generateMixedContent(100),
	}

	for contentType, content := range contentTypes {
		b.Run(contentType, func(b *testing.B) {
			chunker := NewMarkdownChunker()
			contentBytes := []byte(content)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				chunks, err := chunker.ChunkDocument(contentBytes)
				if err != nil {
					b.Fatal(err)
				}
				if len(chunks) == 0 {
					b.Fatal("Expected chunks")
				}
				chunker.ResetPerformanceMonitor()
			}
		})
	}
}

// BenchmarkPerformanceMonitoring 测试性能监控本身的开销
func BenchmarkPerformanceMonitoring(b *testing.B) {
	content := []byte(`# Test Document

This is a test paragraph for performance monitoring benchmarks.

## Section

Another paragraph here.`)

	b.Run("WithMonitoring", func(b *testing.B) {
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatal(err)
			}
			if len(chunks) == 0 {
				b.Fatal("Expected chunks")
			}
			// 获取性能统计（模拟实际使用）
			_ = chunker.GetPerformanceStats()
			chunker.ResetPerformanceMonitor()
		}
	})

	b.Run("WithoutMonitoring", func(b *testing.B) {
		// 创建一个没有性能监控的简化版本
		chunker := NewMarkdownChunker()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				b.Fatal(err)
			}
			if len(chunks) == 0 {
				b.Fatal("Expected chunks")
			}
			// 不获取性能统计
			chunker.ResetPerformanceMonitor()
		}
	})
}

// Helper functions to generate different types of content

func generateHeadingContent(count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		level := (i % 6) + 1
		builder.WriteString(strings.Repeat("#", level))
		builder.WriteString(fmt.Sprintf(" Heading %d Level %d\n\n", i, level))
	}
	return builder.String()
}

func generateParagraphContent(count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString(fmt.Sprintf("This is paragraph number %d. ", i))
		builder.WriteString("It contains multiple sentences to make it more realistic. ")
		builder.WriteString("Each paragraph should be processed as a separate chunk. ")
		builder.WriteString("This helps us test paragraph processing performance.\n\n")
	}
	return builder.String()
}

func generateCodeBlockContent(count int) string {
	var builder strings.Builder
	languages := []string{"go", "python", "javascript", "java", "cpp"}

	for i := 0; i < count; i++ {
		lang := languages[i%len(languages)]
		builder.WriteString(fmt.Sprintf("```%s\n", lang))
		builder.WriteString(fmt.Sprintf("// Code block %d in %s\n", i, lang))
		builder.WriteString("function example() {\n")
		builder.WriteString("    console.log('Hello, World!');\n")
		builder.WriteString("    return true;\n")
		builder.WriteString("}\n```\n\n")
	}
	return builder.String()
}

func generateTableContent(count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString(fmt.Sprintf("| Column A %d | Column B %d | Column C %d |\n", i, i, i))
		builder.WriteString("|-------------|-------------|-------------|\n")
		for j := 0; j < 5; j++ {
			builder.WriteString(fmt.Sprintf("| Data %d-%d   | Value %d-%d  | Result %d-%d |\n", i, j, i, j, i, j))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func generateListContent(count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString(fmt.Sprintf("- List item %d with detailed description\n", i))
		builder.WriteString(fmt.Sprintf("- Another item %d with more content\n", i))
		if i%5 == 0 {
			builder.WriteString("  - Nested item 1\n")
			builder.WriteString("  - Nested item 2\n")
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func generateMixedContent(count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		// Add heading
		builder.WriteString(fmt.Sprintf("# Section %d\n\n", i))

		// Add paragraph
		builder.WriteString(fmt.Sprintf("This is paragraph %d with mixed content testing. ", i))
		builder.WriteString("It combines different markdown elements.\n\n")

		// Add code block occasionally
		if i%5 == 0 {
			builder.WriteString("```go\n")
			builder.WriteString(fmt.Sprintf("func section%d() {\n", i))
			builder.WriteString("    fmt.Println(\"Mixed content\")\n")
			builder.WriteString("}\n```\n\n")
		}

		// Add table occasionally
		if i%10 == 0 {
			builder.WriteString("| Feature | Status |\n")
			builder.WriteString("|---------|--------|\n")
			builder.WriteString(fmt.Sprintf("| Item %d | Active |\n\n", i))
		}

		// Add list occasionally
		if i%7 == 0 {
			builder.WriteString(fmt.Sprintf("- List for section %d\n", i))
			builder.WriteString("- Another list item\n\n")
		}
	}
	return builder.String()
}

// BenchmarkRealWorldScenarios 测试真实世界场景
func BenchmarkRealWorldScenarios(b *testing.B) {
	scenarios := map[string]func() string{
		"TechnicalDocumentation": generateTechnicalDoc,
		"BlogPost":               generateBlogPost,
		"APIDocumentation":       generateAPIDoc,
		"Tutorial":               generateTutorial,
	}

	for scenario, generator := range scenarios {
		b.Run(scenario, func(b *testing.B) {
			chunker := NewMarkdownChunker()
			content := []byte(generator())

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				chunks, err := chunker.ChunkDocument(content)
				if err != nil {
					b.Fatal(err)
				}
				if len(chunks) == 0 {
					b.Fatal("Expected chunks")
				}
				chunker.ResetPerformanceMonitor()
			}
		})
	}
}

func generateTechnicalDoc() string {
	return `# Technical Documentation

## Overview

This document describes the technical implementation of our system.

### Architecture

The system consists of multiple components:

- **Frontend**: React-based user interface
- **Backend**: Go-based API server
- **Database**: PostgreSQL for data persistence

### Implementation Details

` + "```go\n" + `
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/api/health", healthCheck)
    http.ListenAndServe(":8080", nil)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "OK")
}
` + "```\n\n" + `

### Configuration

| Parameter | Type | Description |
|-----------|------|-------------|
| port | int | Server port |
| database_url | string | Database connection string |
| log_level | string | Logging level |

### Deployment

1. Build the application
2. Configure environment variables
3. Deploy to production

> **Note**: Always test in staging first.
`
}

func generateBlogPost() string {
	return `# Understanding Go Concurrency

## Introduction

Go's concurrency model is one of its most powerful features. In this post, we'll explore goroutines and channels.

## Goroutines

Goroutines are lightweight threads managed by the Go runtime.

` + "```go\n" + `
func main() {
    go func() {
        fmt.Println("Hello from goroutine")
    }()
    
    time.Sleep(time.Second)
}
` + "```\n\n" + `

## Channels

Channels provide a way for goroutines to communicate.

### Buffered vs Unbuffered

- **Unbuffered**: Synchronous communication
- **Buffered**: Asynchronous communication up to buffer size

## Best Practices

1. Don't communicate by sharing memory; share memory by communicating
2. Use channels to orchestrate goroutines
3. Always handle channel closing properly

## Conclusion

Go's concurrency primitives make it easy to write concurrent programs.
`
}

func generateAPIDoc() string {
	return `# API Documentation

## Authentication

All API requests require authentication using Bearer tokens.

` + "```bash\n" + `
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://api.example.com/v1/users
` + "```\n\n" + `

## Endpoints

### Users

#### GET /v1/users

Returns a list of users.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| limit | int | Maximum number of users to return |
| offset | int | Number of users to skip |

**Response:**

` + "```json\n" + `
{
  "users": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com"
    }
  ],
  "total": 100
}
` + "```\n\n" + `

#### POST /v1/users

Creates a new user.

**Request Body:**

` + "```json\n" + `
{
  "name": "Jane Doe",
  "email": "jane@example.com"
}
` + "```\n\n" + `

## Error Handling

The API returns standard HTTP status codes:

- 200: Success
- 400: Bad Request
- 401: Unauthorized
- 404: Not Found
- 500: Internal Server Error
`
}

func generateTutorial() string {
	return `# Getting Started with Markdown Chunker

## Installation

First, install the package:

` + "```bash\n" + `
go get github.com/example/markdown-chunker
` + "```\n\n" + `

## Basic Usage

Here's how to use the chunker:

` + "```go\n" + `
package main

import (
    "fmt"
    "github.com/example/markdown-chunker"
)

func main() {
    chunker := markdownchunker.NewMarkdownChunker()
    
    content := []byte("# Hello World\n\nThis is a paragraph.")
    chunks, err := chunker.ChunkDocument(content)
    if err != nil {
        panic(err)
    }
    
    for _, chunk := range chunks {
        fmt.Printf("Type: %s, Content: %s\n", chunk.Type, chunk.Content)
    }
}
` + "```\n\n" + `

## Configuration

You can customize the chunker behavior:

` + "```go\n" + `
config := markdownchunker.DefaultConfig()
config.MaxChunkSize = 1000
config.FilterEmptyChunks = true

chunker := markdownchunker.NewMarkdownChunkerWithConfig(config)
` + "```\n\n" + `

## Advanced Features

### Error Handling

The chunker supports different error handling modes:

- **Strict**: Stop on first error
- **Permissive**: Continue processing, collect errors
- **Silent**: Ignore errors

### Performance Monitoring

Track processing performance:

` + "```go\n" + `
chunks, err := chunker.ChunkDocument(content)
stats := chunker.GetPerformanceStats()

fmt.Printf("Processing time: %v\n", stats.ProcessingTime)
fmt.Printf("Memory used: %d bytes\n", stats.MemoryUsed)
` + "```\n\n" + `

## Tips and Tricks

1. Use appropriate chunk sizes for your use case
2. Enable object pooling for better performance
3. Monitor memory usage for large documents

## Conclusion

The markdown chunker provides a flexible way to process markdown documents efficiently.
`
}
