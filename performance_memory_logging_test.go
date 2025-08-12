package markdownchunker

import (
	"strings"
	"testing"
	"time"
)

// TestPerformanceAndMemoryLogging 测试性能监控和内存优化的日志集成
func TestPerformanceAndMemoryLogging(t *testing.T) {
	// 创建配置，启用日志和内存优化
	config := &ChunkerConfig{
		MaxChunkSize:        0,
		EnabledTypes:        nil,
		CustomExtractors:    []MetadataExtractor{},
		ErrorHandling:       ErrorModePermissive,
		PerformanceMode:     PerformanceModeMemoryOptimized,
		FilterEmptyChunks:   true,
		PreserveWhitespace:  false,
		MemoryLimit:         50 * 1024 * 1024, // 50MB限制
		EnableObjectPooling: true,
		LogLevel:            "DEBUG",
		EnableLog:           true,
		LogFormat:           "console",
		LogDirectory:        "./test-logs",
	}

	chunker := NewMarkdownChunkerWithConfig(config)

	// 创建一个中等大小的测试文档
	var contentBuilder strings.Builder
	contentBuilder.WriteString("# 性能和内存监控测试文档\n\n")

	// 添加多种类型的内容来触发不同的处理路径
	for i := 0; i < 100; i++ {
		contentBuilder.WriteString("## 标题 ")
		contentBuilder.WriteString(strings.Repeat("A", i))
		contentBuilder.WriteString("\n\n")

		contentBuilder.WriteString("这是一个段落，包含一些文本内容。")
		contentBuilder.WriteString(strings.Repeat("内容", i%10))
		contentBuilder.WriteString("\n\n")

		if i%10 == 0 {
			contentBuilder.WriteString("```go\n")
			contentBuilder.WriteString("func example() {\n")
			contentBuilder.WriteString("    // 代码示例\n")
			contentBuilder.WriteString("    return nil\n")
			contentBuilder.WriteString("}\n")
			contentBuilder.WriteString("```\n\n")
		}

		if i%15 == 0 {
			contentBuilder.WriteString("| 列1 | 列2 | 列3 |\n")
			contentBuilder.WriteString("|-----|-----|-----|\n")
			contentBuilder.WriteString("| 值1 | 值2 | 值3 |\n")
			contentBuilder.WriteString("| 值4 | 值5 | 值6 |\n\n")
		}

		if i%20 == 0 {
			contentBuilder.WriteString("- 列表项 1\n")
			contentBuilder.WriteString("- 列表项 2\n")
			contentBuilder.WriteString("- 列表项 3\n\n")
		}
	}

	content := []byte(contentBuilder.String())
	t.Logf("测试文档大小: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))

	// 处理文档
	startTime := time.Now()
	chunks, err := chunker.ChunkDocument(content)
	processingTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("文档处理失败: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("应该生成一些块")
	}

	// 获取性能统计信息
	perfStats := chunker.GetPerformanceStats()

	t.Logf("性能和内存监控测试结果:")
	t.Logf("  处理时间: %v", processingTime)
	t.Logf("  生成块数: %d", len(chunks))
	t.Logf("  文档大小: %d bytes", len(content))
	t.Logf("  性能统计:")
	t.Logf("    处理时间: %v", perfStats.ProcessingTime)
	t.Logf("    内存使用: %d bytes (%.2f MB)", perfStats.MemoryUsed, float64(perfStats.MemoryUsed)/(1024*1024))
	t.Logf("    峰值内存: %d bytes (%.2f MB)", perfStats.PeakMemory, float64(perfStats.PeakMemory)/(1024*1024))
	t.Logf("    每秒处理块数: %.2f", perfStats.ChunksPerSecond)
	t.Logf("    每秒处理字节数: %.2f", perfStats.BytesPerSecond)

	// 如果启用了内存优化器，获取其统计信息
	if chunker.memoryOptimizer != nil {
		memStats := chunker.memoryOptimizer.GetMemoryStats()
		t.Logf("  内存优化器统计:")
		t.Logf("    当前内存: %d bytes (%.2f MB)", memStats.CurrentMemory, float64(memStats.CurrentMemory)/(1024*1024))
		t.Logf("    内存限制: %d bytes (%.2f MB)", memStats.MemoryLimit, float64(memStats.MemoryLimit)/(1024*1024))
		t.Logf("    已处理字节数: %d bytes (%.2f MB)", memStats.ProcessedBytes, float64(memStats.ProcessedBytes)/(1024*1024))
		t.Logf("    GC阈值: %d bytes (%.2f MB)", memStats.GCThreshold, float64(memStats.GCThreshold)/(1024*1024))
		t.Logf("    总分配: %d bytes (%.2f MB)", memStats.TotalAllocations, float64(memStats.TotalAllocations)/(1024*1024))
		t.Logf("    GC周期数: %d", memStats.GCCycles)
	}

	// 验证基本功能
	if perfStats.TotalChunks != len(chunks) {
		t.Errorf("性能统计中的块数不匹配: 期望 %d, 实际 %d", len(chunks), perfStats.TotalChunks)
	}

	if perfStats.TotalBytes != int64(len(content)) {
		t.Errorf("性能统计中的字节数不匹配: 期望 %d, 实际 %d", len(content), perfStats.TotalBytes)
	}

	// 检查是否有错误
	if chunker.HasErrors() {
		errors := chunker.GetErrors()
		t.Logf("处理过程中的错误数量: %d", len(errors))
		for i, err := range errors {
			t.Logf("  错误 %d: %s", i+1, err.Error())
		}
	}
}

// TestMemoryThresholdWarnings 测试内存阈值警告
func TestMemoryThresholdWarnings(t *testing.T) {
	config := &ChunkerConfig{
		LogLevel:     "DEBUG",
		EnableLog:    true,
		LogFormat:    "console",
		LogDirectory: "./test-logs",
		MemoryLimit:  1024 * 1024, // 1MB限制，很容易触发
	}

	chunker := NewMarkdownChunkerWithConfig(config)

	// 创建一个相对较大的文档来触发内存警告
	var contentBuilder strings.Builder
	for i := 0; i < 1000; i++ {
		contentBuilder.WriteString("# 大标题 ")
		contentBuilder.WriteString(strings.Repeat("X", 100))
		contentBuilder.WriteString("\n\n")
		contentBuilder.WriteString("这是一个很长的段落，")
		contentBuilder.WriteString(strings.Repeat("包含大量重复内容", 20))
		contentBuilder.WriteString("\n\n")
	}

	content := []byte(contentBuilder.String())
	t.Logf("大文档大小: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))

	// 处理文档（可能会触发内存警告）
	chunks, err := chunker.ChunkDocument(content)
	// 在宽松模式下，即使内存超限也应该能继续处理
	if err != nil {
		t.Logf("处理过程中出现错误（预期的）: %v", err)
	}

	t.Logf("生成的块数: %d", len(chunks))

	// 检查错误
	if chunker.HasErrors() {
		errors := chunker.GetErrors()
		t.Logf("捕获的错误数量: %d", len(errors))

		// 检查是否有内存相关的错误
		memoryErrors := chunker.GetErrorsByType(ErrorTypeMemoryExhausted)
		if len(memoryErrors) > 0 {
			t.Logf("内存耗尽错误数量: %d", len(memoryErrors))
			for i, err := range memoryErrors {
				t.Logf("  内存错误 %d: %s", i+1, err.Error())
			}
		}
	}

	// 获取最终的性能统计
	perfStats := chunker.GetPerformanceStats()
	t.Logf("最终性能统计:")
	t.Logf("  峰值内存: %d bytes (%.2f MB)", perfStats.PeakMemory, float64(perfStats.PeakMemory)/(1024*1024))
	t.Logf("  内存使用: %d bytes (%.2f MB)", perfStats.MemoryUsed, float64(perfStats.MemoryUsed)/(1024*1024))
}

// TestProgressLoggingForLargeDocuments 测试大型文档的进度日志
func TestProgressLoggingForLargeDocuments(t *testing.T) {
	config := &ChunkerConfig{
		LogLevel:     "INFO",
		EnableLog:    true,
		LogFormat:    "console",
		LogDirectory: "./test-logs",
	}

	chunker := NewMarkdownChunkerWithConfig(config)

	// 创建一个大于1MB的文档来触发进度日志
	var contentBuilder strings.Builder
	for i := 0; i < 2000; i++ {
		contentBuilder.WriteString("## 进度测试标题 ")
		contentBuilder.WriteString(strings.Repeat("P", i%50))
		contentBuilder.WriteString("\n\n")
		contentBuilder.WriteString("进度测试段落内容，")
		contentBuilder.WriteString(strings.Repeat("用于测试进度日志功能", 10))
		contentBuilder.WriteString("\n\n")
	}

	content := []byte(contentBuilder.String())
	t.Logf("大文档大小: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))

	// 确保文档大于1MB以触发进度日志
	if len(content) <= 1024*1024 {
		// 如果还不够大，继续添加内容
		for len(content) <= 1024*1024 {
			contentBuilder.WriteString("额外内容用于达到1MB阈值。")
			contentBuilder.WriteString(strings.Repeat("X", 1000))
			contentBuilder.WriteString("\n\n")
			content = []byte(contentBuilder.String())
		}
		t.Logf("调整后的文档大小: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))
	}

	// 处理文档
	startTime := time.Now()
	chunks, err := chunker.ChunkDocument(content)
	processingTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("大文档处理失败: %v", err)
	}

	t.Logf("大文档处理结果:")
	t.Logf("  处理时间: %v", processingTime)
	t.Logf("  生成块数: %d", len(chunks))
	t.Logf("  文档大小: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))

	// 获取性能统计
	perfStats := chunker.GetPerformanceStats()
	t.Logf("  性能指标:")
	t.Logf("    每秒处理块数: %.2f", perfStats.ChunksPerSecond)
	t.Logf("    每秒处理MB数: %.2f", perfStats.BytesPerSecond/(1024*1024))
	t.Logf("    峰值内存: %.2f MB", float64(perfStats.PeakMemory)/(1024*1024))
}
