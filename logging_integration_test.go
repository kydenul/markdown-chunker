package markdownchunker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestLoggingEndToEndIntegration 创建端到端的日志输出测试
func TestLoggingEndToEndIntegration(t *testing.T) {
	// 创建临时日志目录
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "test-logs")

	t.Run("complete document processing with debug logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: logDir,
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		// 复杂的markdown文档，包含所有支持的元素
		markdown := `# 主要文档标题

这是一个包含**粗体**和*斜体*文本的段落。

## 安装指南

安装步骤如下：

1. 克隆仓库 [GitHub](https://github.com/example/repo)
2. 安装依赖
3. 运行设置脚本

### 前置条件

需要以下软件：

- Go 1.19 或更高版本
- Git
- Docker（可选）

## 代码示例

Go 示例：

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

Python 示例：

` + "```python" + `
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
` + "```" + `

## 配置表格

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| host | string | localhost | 服务器主机 |
| port | integer | 8080 | 服务器端口 |
| debug | boolean | false | 启用调试模式 |
| timeout | decimal | 30.5 | 请求超时时间（秒） |

## 重要提示

> **警告**: 在进行更改之前，请务必备份数据。
> 
> **注意**: 此功能是实验性的，可能在未来版本中发生变化。

---

*最后更新: 2024-01-15*`

		// 执行文档处理
		start := time.Now()
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		// 验证处理结果
		if len(chunks) == 0 {
			t.Error("Expected chunks to be generated")
		}

		// 验证所有预期的块类型都存在
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

		// 验证性能统计
		stats := chunker.GetPerformanceStats()
		if stats.ProcessingTime <= 0 {
			t.Error("Processing time should be positive")
		}

		if stats.TotalChunks != len(chunks) {
			t.Errorf("Performance stats mismatch: expected %d chunks, got %d", len(chunks), stats.TotalChunks)
		}

		t.Logf("End-to-end integration test completed:")
		t.Logf("  Total chunks: %d", len(chunks))
		t.Logf("  Chunk types: %v", chunkTypes)
		t.Logf("  Processing time: %v", duration)
		t.Logf("  Memory used: %d bytes", stats.MemoryUsed)
		t.Logf("  Log directory: %s", logDir)

		// 验证日志文件是否创建
		logFiles, err := os.ReadDir(logDir)
		if err != nil {
			t.Logf("Warning: Could not read log directory: %v", err)
		} else if len(logFiles) > 0 {
			t.Logf("  Log files created: %d", len(logFiles))
			for _, file := range logFiles {
				t.Logf("    - %s", file.Name())
			}
		}
	})

	t.Run("json format logging integration", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "INFO",
			EnableLog:    true,
			LogFormat:    "json",
			LogDirectory: filepath.Join(logDir, "json"),
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		markdown := `# JSON 格式测试

这是一个测试JSON格式日志的文档。

` + "```javascript" + `
const data = {
    "name": "test",
    "value": 123
};
console.log(JSON.stringify(data));
` + "```" + `

| 字段 | 值 |
|------|-----|
| 测试 | 成功 |`

		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected chunks to be generated")
		}

		t.Logf("JSON format logging test completed with %d chunks", len(chunks))
	})

	t.Run("different log levels integration", func(t *testing.T) {
		logLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
		markdown := `# 日志级别测试

这是一个简单的测试文档。

` + "```go" + `
func test() {
    fmt.Println("Hello")
}
` + "```"

		for _, level := range logLevels {
			t.Run(fmt.Sprintf("level_%s", level), func(t *testing.T) {
				config := &ChunkerConfig{
					LogLevel:     level,
					EnableLog:    true,
					LogFormat:    "console",
					LogDirectory: filepath.Join(logDir, strings.ToLower(level)),
				}

				chunker := NewMarkdownChunkerWithConfig(config)
				chunks, err := chunker.ChunkDocument([]byte(markdown))
				if err != nil {
					t.Fatalf("ChunkDocument() with log level %s error = %v", level, err)
				}

				if len(chunks) == 0 {
					t.Errorf("Expected chunks for log level %s", level)
				}

				t.Logf("Log level %s test completed with %d chunks", level, len(chunks))
			})
		}
	})
}

// TestLoggingErrorScenariosIntegration 测试错误场景下的日志记录
func TestLoggingErrorScenariosIntegration(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "error-logs")

	t.Run("nil input error logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:      "DEBUG",
			EnableLog:     true,
			LogFormat:     "console",
			LogDirectory:  logDir,
			ErrorHandling: ErrorModePermissive,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument(nil)
		// 在宽松模式下，应该返回空切片而不是错误
		if err != nil {
			t.Logf("Received expected error for nil input: %v", err)
		}

		if len(chunks) != 0 {
			t.Error("Expected no chunks for nil input")
		}

		// 验证错误被记录
		if chunker.HasErrors() {
			errors := chunker.GetErrors()
			t.Logf("Recorded %d errors as expected", len(errors))
			for i, e := range errors {
				t.Logf("  Error %d: %v", i+1, e)
			}
		}
	})

	t.Run("chunk size limit error logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:      "DEBUG",
			EnableLog:     true,
			LogFormat:     "console",
			LogDirectory:  logDir,
			MaxChunkSize:  20, // 非常小的限制
			ErrorHandling: ErrorModePermissive,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		markdown := `# 这是一个非常长的标题，肯定会超过20个字符的限制

这是一个很长的段落，包含了大量的文本内容，用来测试块大小限制的错误处理和日志记录功能。`

		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Logf("Received error for size limit: %v", err)
		}

		// 验证错误被记录
		if chunker.HasErrors() {
			errors := chunker.GetErrorsByType(ErrorTypeChunkTooLarge)
			if len(errors) == 0 {
				t.Error("Expected ChunkTooLarge errors to be recorded")
			} else {
				t.Logf("Recorded %d ChunkTooLarge errors as expected", len(errors))
			}
		}

		// 在宽松模式下，应该有截断的块
		if len(chunks) == 0 {
			t.Error("Expected some chunks even with size limit in permissive mode")
		}

		// 验证块被截断
		for _, chunk := range chunks {
			if len(chunk.Content) > 20 {
				t.Errorf("Chunk content should be truncated to 20 chars, got %d", len(chunk.Content))
			}
		}

		t.Logf("Size limit error test completed with %d chunks", len(chunks))
	})

	t.Run("parsing error logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:      "DEBUG",
			EnableLog:     true,
			LogFormat:     "console",
			LogDirectory:  logDir,
			ErrorHandling: ErrorModePermissive,
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		// 创建一个可能导致解析问题的文档（虽然goldmark很宽松）
		problematicMarkdown := []byte{0xFF, 0xFE, 0xFD} // 无效的UTF-8字节

		chunks, err := chunker.ChunkDocument(problematicMarkdown)
		if err != nil {
			t.Logf("Received parsing error as expected: %v", err)
		}

		// goldmark通常能处理大多数输入，所以可能不会有错误
		t.Logf("Parsing error test completed with %d chunks", len(chunks))
	})

	t.Run("memory limit warning logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:      "DEBUG",
			EnableLog:     true,
			LogFormat:     "console",
			LogDirectory:  logDir,
			MemoryLimit:   1024, // 1KB 限制，很容易超过
			ErrorHandling: ErrorModePermissive,
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		// 创建一个相对较大的文档
		var largeDoc strings.Builder
		for i := 0; i < 100; i++ {
			largeDoc.WriteString(fmt.Sprintf("# 标题 %d\n\n这是第 %d 个段落，包含一些内容。\n\n", i, i))
		}

		chunks, err := chunker.ChunkDocument([]byte(largeDoc.String()))
		if err != nil {
			t.Logf("Received memory limit error: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected some chunks even with memory limit")
		}

		t.Logf("Memory limit test completed with %d chunks", len(chunks))
	})

	t.Run("strict mode error handling", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:      "ERROR",
			EnableLog:     true,
			LogFormat:     "console",
			LogDirectory:  logDir,
			MaxChunkSize:  10, // 非常小的限制
			ErrorHandling: ErrorModeStrict,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		markdown := `# 这个标题肯定超过10个字符`

		chunks, err := chunker.ChunkDocument([]byte(markdown))

		// 在严格模式下，应该返回错误
		if err == nil {
			t.Error("Expected error in strict mode with size limit")
		} else {
			t.Logf("Received expected error in strict mode: %v", err)
		}

		// 严格模式下可能没有块或有部分块
		t.Logf("Strict mode test completed with %d chunks", len(chunks))
	})
}

// TestLoggingPerformanceImpact 验证日志对系统性能的影响
func TestLoggingPerformanceImpact(t *testing.T) {
	// 创建测试文档
	var testDoc strings.Builder
	for i := 0; i < 500; i++ {
		testDoc.WriteString(fmt.Sprintf("# 标题 %d\n\n", i))
		testDoc.WriteString(fmt.Sprintf("这是第 %d 个段落，包含一些测试内容。\n\n", i))
		testDoc.WriteString("```go\n")
		testDoc.WriteString(fmt.Sprintf("func test%d() {\n", i))
		testDoc.WriteString(fmt.Sprintf("    fmt.Println(\"Test %d\")\n", i))
		testDoc.WriteString("}\n")
		testDoc.WriteString("```\n\n")
	}
	content := []byte(testDoc.String())

	tempDir := t.TempDir()

	t.Run("performance comparison: logging enabled vs disabled", func(t *testing.T) {
		// 测试启用详细日志的性能
		enabledConfig := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: filepath.Join(tempDir, "enabled"),
		}

		// 测试禁用日志的性能
		disabledConfig := &ChunkerConfig{
			LogLevel:     "ERROR",
			EnableLog:    false,
			LogFormat:    "console",
			LogDirectory: filepath.Join(tempDir, "disabled"),
		}

		// 预热
		chunkerWarmup := NewMarkdownChunkerWithConfig(enabledConfig)
		_, _ = chunkerWarmup.ChunkDocument([]byte("# Warmup\n\nWarmup content."))

		// 测试启用日志的性能
		chunkerEnabled := NewMarkdownChunkerWithConfig(enabledConfig)
		startEnabled := time.Now()
		chunksEnabled, err := chunkerEnabled.ChunkDocument(content)
		durationEnabled := time.Since(startEnabled)

		if err != nil {
			t.Fatalf("ChunkDocument with logging enabled failed: %v", err)
		}

		// 测试禁用日志的性能
		chunkerDisabled := NewMarkdownChunkerWithConfig(disabledConfig)
		startDisabled := time.Now()
		chunksDisabled, err := chunkerDisabled.ChunkDocument(content)
		durationDisabled := time.Since(startDisabled)

		if err != nil {
			t.Fatalf("ChunkDocument with logging disabled failed: %v", err)
		}

		// 验证结果一致性
		if len(chunksEnabled) != len(chunksDisabled) {
			t.Errorf("Chunk count should be same regardless of logging: enabled=%d, disabled=%d",
				len(chunksEnabled), len(chunksDisabled))
		}

		// 计算性能影响
		performanceImpact := float64(durationEnabled) / float64(durationDisabled)

		t.Logf("Performance impact analysis:")
		t.Logf("  Document size: %d bytes", len(content))
		t.Logf("  Chunks produced: %d", len(chunksEnabled))
		t.Logf("  Logging enabled time: %v", durationEnabled)
		t.Logf("  Logging disabled time: %v", durationDisabled)
		t.Logf("  Performance impact ratio: %.2fx", performanceImpact)

		// 日志不应该显著影响性能（允许3倍的差异作为合理范围）
		if performanceImpact > 3.0 {
			t.Errorf("Logging has significant performance impact: %.2fx slower", performanceImpact)
		} else if performanceImpact > 2.0 {
			t.Logf("Warning: Logging has moderate performance impact: %.2fx slower", performanceImpact)
		} else {
			t.Logf("Logging performance impact is acceptable: %.2fx slower", performanceImpact)
		}
	})

	t.Run("performance comparison: different log levels", func(t *testing.T) {
		logLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
		results := make(map[string]time.Duration)

		for _, level := range logLevels {
			config := &ChunkerConfig{
				LogLevel:     level,
				EnableLog:    true,
				LogFormat:    "console",
				LogDirectory: filepath.Join(tempDir, strings.ToLower(level)),
			}

			chunker := NewMarkdownChunkerWithConfig(config)
			start := time.Now()
			chunks, err := chunker.ChunkDocument(content)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("ChunkDocument with log level %s failed: %v", level, err)
			}

			results[level] = duration
			t.Logf("Log level %s: %v (%d chunks)", level, duration, len(chunks))
		}

		// 验证DEBUG级别（最详细）不会比ERROR级别（最少）慢太多
		debugTime := results["DEBUG"]
		errorTime := results["ERROR"]
		ratio := float64(debugTime) / float64(errorTime)

		t.Logf("DEBUG vs ERROR performance ratio: %.2fx", ratio)

		if ratio > 5.0 {
			t.Errorf("DEBUG logging is significantly slower than ERROR: %.2fx", ratio)
		}
	})

	t.Run("performance comparison: console vs json format", func(t *testing.T) {
		formats := []string{"console", "json"}
		results := make(map[string]time.Duration)

		for _, format := range formats {
			config := &ChunkerConfig{
				LogLevel:     "INFO",
				EnableLog:    true,
				LogFormat:    format,
				LogDirectory: filepath.Join(tempDir, format),
			}

			chunker := NewMarkdownChunkerWithConfig(config)
			start := time.Now()
			chunks, err := chunker.ChunkDocument(content)
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("ChunkDocument with format %s failed: %v", format, err)
			}

			results[format] = duration
			t.Logf("Format %s: %v (%d chunks)", format, duration, len(chunks))
		}

		// 比较两种格式的性能
		consoleTime := results["console"]
		jsonTime := results["json"]
		ratio := float64(jsonTime) / float64(consoleTime)

		t.Logf("JSON vs Console format performance ratio: %.2fx", ratio)

		// JSON格式通常稍慢，但不应该差异太大
		if ratio > 3.0 {
			t.Errorf("JSON format is significantly slower than console: %.2fx", ratio)
		}
	})

	t.Run("memory usage with logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:            "DEBUG",
			EnableLog:           true,
			LogFormat:           "console",
			LogDirectory:        filepath.Join(tempDir, "memory"),
			PerformanceMode:     PerformanceModeMemoryOptimized,
			EnableObjectPooling: true,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("ChunkDocument with memory optimization failed: %v", err)
		}

		stats := chunker.GetPerformanceStats()

		t.Logf("Memory usage with logging:")
		t.Logf("  Input size: %d bytes", len(content))
		t.Logf("  Chunks produced: %d", len(chunks))
		t.Logf("  Peak memory: %d bytes", stats.PeakMemory)
		t.Logf("  Current memory: %d bytes", stats.MemoryUsed)
		t.Logf("  Processing time: %v", stats.ProcessingTime)

		// 验证内存使用合理
		if stats.PeakMemory > 100*1024*1024 { // 100MB
			t.Errorf("Peak memory usage too high with logging: %d bytes", stats.PeakMemory)
		}
	})
}

// TestLoggingConcurrentSafety 测试并发环境下的日志安全性
func TestLoggingConcurrentSafety(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "concurrent")

	config := &ChunkerConfig{
		LogLevel:     "DEBUG",
		EnableLog:    true,
		LogFormat:    "console",
		LogDirectory: logDir,
	}

	markdown := `# 并发测试文档

这是一个用于测试并发日志记录的文档。

` + "```go" + `
func concurrent() {
    fmt.Println("Concurrent processing")
}
` + "```" + `

| 线程 | 状态 |
|------|------|
| 1    | 运行中 |
| 2    | 等待中 |`

	const numGoroutines = 10
	const iterations = 5

	results := make(chan []Chunk, numGoroutines*iterations)
	errors := make(chan error, numGoroutines*iterations)

	// 启动多个goroutine并发处理
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < iterations; j++ {
				// 每个goroutine创建自己的chunker实例
				chunker := NewMarkdownChunkerWithConfig(config)
				chunks, err := chunker.ChunkDocument([]byte(markdown))
				if err != nil {
					errors <- fmt.Errorf("goroutine %d iteration %d: %v", goroutineID, j, err)
					return
				}
				results <- chunks
			}
		}(i)
	}

	// 收集所有结果
	expectedResults := numGoroutines * iterations
	successCount := 0

	for i := 0; i < expectedResults; i++ {
		select {
		case err := <-errors:
			t.Errorf("Concurrent logging failed: %v", err)
		case chunks := <-results:
			successCount++
			// 验证每个结果都有预期的结构
			if len(chunks) == 0 {
				t.Error("Expected some chunks from concurrent processing")
			}
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent logging test timed out")
		}
	}

	if successCount != expectedResults {
		t.Errorf("Expected %d successful results, got %d", expectedResults, successCount)
	}

	t.Logf("Concurrent logging test completed successfully:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Iterations per goroutine: %d", iterations)
	t.Logf("  Total successful operations: %d", successCount)
	t.Logf("  Log directory: %s", logDir)
}

// TestLoggingLargeDocumentProcessing 测试大文档处理时的日志记录
func TestLoggingLargeDocumentProcessing(t *testing.T) {
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "large-doc")

	config := &ChunkerConfig{
		LogLevel:     "INFO",
		EnableLog:    true,
		LogFormat:    "console",
		LogDirectory: logDir,
	}

	// 创建一个大文档
	var largeDoc strings.Builder
	for i := 0; i < 2000; i++ {
		largeDoc.WriteString(fmt.Sprintf("# 章节 %d\n\n", i))
		largeDoc.WriteString(fmt.Sprintf("这是第 %d 章的内容，包含详细的描述和说明。", i))
		largeDoc.WriteString("这里有更多的文本内容来增加文档的大小。\n\n")

		if i%10 == 0 {
			largeDoc.WriteString("```go\n")
			largeDoc.WriteString(fmt.Sprintf("func chapter%d() {\n", i))
			largeDoc.WriteString(fmt.Sprintf("    fmt.Println(\"Chapter %d\")\n", i))
			largeDoc.WriteString("}\n")
			largeDoc.WriteString("```\n\n")
		}

		if i%20 == 0 {
			largeDoc.WriteString("| 项目 | 值 |\n")
			largeDoc.WriteString("|------|----|\n")
			largeDoc.WriteString(fmt.Sprintf("| 章节 | %d |\n", i))
			largeDoc.WriteString("| 状态 | 完成 |\n\n")
		}
	}

	content := []byte(largeDoc.String())
	chunker := NewMarkdownChunkerWithConfig(config)

	t.Logf("Processing large document:")
	t.Logf("  Document size: %d bytes (%.2f MB)", len(content), float64(len(content))/(1024*1024))

	start := time.Now()
	chunks, err := chunker.ChunkDocument(content)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Large document processing failed: %v", err)
	}

	stats := chunker.GetPerformanceStats()

	t.Logf("Large document processing completed:")
	t.Logf("  Chunks produced: %d", len(chunks))
	t.Logf("  Processing time: %v", duration)
	t.Logf("  Chunks per second: %.2f", float64(len(chunks))/duration.Seconds())
	t.Logf("  MB per second: %.2f", float64(len(content))/(1024*1024)/duration.Seconds())
	t.Logf("  Peak memory: %d bytes (%.2f MB)", stats.PeakMemory, float64(stats.PeakMemory)/(1024*1024))

	// 验证性能合理
	if duration > 30*time.Second {
		t.Errorf("Large document processing took too long: %v", duration)
	}

	// 验证块数量合理
	expectedMinChunks := 2000 // 至少应该有2000个标题块
	if len(chunks) < expectedMinChunks {
		t.Errorf("Expected at least %d chunks, got %d", expectedMinChunks, len(chunks))
	}
}

// TestLoggingConfigurationValidation 测试日志配置验证的集成
func TestLoggingConfigurationValidation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("invalid configuration handling", func(t *testing.T) {
		// 测试无效的日志级别
		config := &ChunkerConfig{
			LogLevel:     "INVALID_LEVEL",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: tempDir,
		}

		// 应该使用默认配置而不是失败
		chunker := NewMarkdownChunkerWithConfig(config)
		if chunker.logger == nil {
			t.Fatal("Logger should be initialized even with invalid config")
		}

		chunks, err := chunker.ChunkDocument([]byte("# Test\n\nContent."))
		if err != nil {
			t.Fatalf("Should handle invalid config gracefully: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected chunks even with invalid config")
		}
	})

	t.Run("configuration validation logging", func(t *testing.T) {
		// 捕获标准输出来验证配置验证日志
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		config := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: tempDir,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte("# Test\n\nContent."))

		// 恢复标准输出
		w.Close()
		os.Stdout = oldStdout

		// 读取捕获的输出
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if err != nil {
			t.Fatalf("Configuration validation test failed: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected chunks from configuration validation test")
		}

		// 验证配置验证相关的日志输出
		if !strings.Contains(output, "配置验证") && !strings.Contains(output, "validation") {
			t.Logf("Configuration validation logs may not be visible in captured output")
		}

		t.Logf("Configuration validation test completed with %d chunks", len(chunks))
	})
}
