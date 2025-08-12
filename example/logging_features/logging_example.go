package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 日志功能演示示例 ===")

	// 创建测试文档
	testMarkdown := createTestDocument()

	// 演示各种日志功能
	demonstrateBasicLogging(testMarkdown)
	demonstrateLogLevels(testMarkdown)
	demonstrateLogFormats(testMarkdown)
	demonstrateErrorLogging(testMarkdown)
	demonstratePerformanceLogging(testMarkdown)
	demonstrateCustomLogDirectory(testMarkdown)
	demonstrateLoggingWithConfiguration(testMarkdown)

	fmt.Println("\n=== 日志功能演示完成 ===")
	fmt.Println("请查看生成的日志文件以了解详细的日志输出。")
}

// createTestDocument 创建测试文档
func createTestDocument() string {
	return `# 日志功能测试文档

这是一个用于测试日志功能的示例文档。

## 基本内容

这是一个普通段落，包含一些文本内容。

### 代码示例

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
    for i := 0; i < 10; i++ {
        if i%2 == 0 {
            fmt.Printf("Even: %d\n", i)
        }
    }
}
` + "```" + `

### 表格示例

| 名称 | 值 | 状态 |
|------|----|----- |
| 项目A | 100 | 活跃 |
| 项目B | 200 | 非活跃 |

### 列表示例

1. 第一项
2. 第二项
   - 子项 A
   - 子项 B
3. 第三项

> 这是一个引用块，包含重要信息。

---

*文档结束*`
}

// demonstrateBasicLogging 演示基本日志功能
func demonstrateBasicLogging(markdown string) {
	fmt.Println("\n1. 基本日志功能演示")

	// 创建临时日志目录
	logDir := filepath.Join(".", "demo-logs", "basic")
	os.MkdirAll(logDir, 0o755)

	// 配置基本日志
	config := mc.DefaultConfig()
	config.EnableLog = true
	config.LogLevel = "INFO"
	config.LogFormat = "console"
	config.LogDirectory = logDir

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("  处理完成: %d 个块\n", len(chunks))
	fmt.Printf("  日志文件位置: %s\n", logDir)
	fmt.Printf("  日志级别: %s\n", config.LogLevel)
	fmt.Printf("  日志格式: %s\n", config.LogFormat)

	// 显示日志文件
	showLogFiles(logDir)
}

// demonstrateLogLevels 演示不同日志级别
func demonstrateLogLevels(markdown string) {
	fmt.Println("\n2. 不同日志级别演示")

	levels := []string{"DEBUG", "INFO", "WARN", "ERROR"}

	for _, level := range levels {
		fmt.Printf("  测试日志级别: %s\n", level)

		logDir := filepath.Join(".", "demo-logs", "levels", level)
		os.MkdirAll(logDir, 0o755)

		config := mc.DefaultConfig()
		config.EnableLog = true
		config.LogLevel = level
		config.LogFormat = "console"
		config.LogDirectory = logDir

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("    处理错误: %v", err)
			continue
		}

		fmt.Printf("    处理结果: %d 个块\n", len(chunks))
		fmt.Printf("    日志目录: %s\n", logDir)

		// 显示日志文件大小（不同级别会产生不同数量的日志）
		showLogFileSize(logDir)
	}
}

// demonstrateLogFormats 演示不同日志格式
func demonstrateLogFormats(markdown string) {
	fmt.Println("\n3. 不同日志格式演示")

	formats := []string{"console", "json"}

	for _, format := range formats {
		fmt.Printf("  测试日志格式: %s\n", format)

		logDir := filepath.Join(".", "demo-logs", "formats", format)
		os.MkdirAll(logDir, 0o755)

		config := mc.DefaultConfig()
		config.EnableLog = true
		config.LogLevel = "INFO"
		config.LogFormat = format
		config.LogDirectory = logDir

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("    处理错误: %v", err)
			continue
		}

		fmt.Printf("    处理结果: %d 个块\n", len(chunks))
		fmt.Printf("    日志目录: %s\n", logDir)

		// 显示日志文件内容示例
		showLogSample(logDir, format)
	}
}

// demonstrateErrorLogging 演示错误日志记录
func demonstrateErrorLogging(markdown string) {
	fmt.Println("\n4. 错误日志记录演示")

	logDir := filepath.Join(".", "demo-logs", "errors")
	os.MkdirAll(logDir, 0o755)

	// 创建会产生错误的配置
	config := mc.DefaultConfig()
	config.EnableLog = true
	config.LogLevel = "DEBUG"
	config.LogFormat = "console"
	config.LogDirectory = logDir
	config.MaxChunkSize = 50 // 很小的限制，会产生错误
	config.ErrorHandling = mc.ErrorModePermissive

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))

	fmt.Printf("  处理结果: %d 个块\n", len(chunks))
	fmt.Printf("  返回错误: %v\n", err)
	fmt.Printf("  记录的错误数: %d\n", len(chunker.GetErrors()))
	fmt.Printf("  日志目录: %s\n", logDir)

	if chunker.HasErrors() {
		fmt.Printf("  错误类型:\n")
		for _, err := range chunker.GetErrors() {
			fmt.Printf("    - %s: %s\n", err.Type.String(), err.Message)
		}
	}

	showLogFiles(logDir)
}

// demonstratePerformanceLogging 演示性能日志记录
func demonstratePerformanceLogging(markdown string) {
	fmt.Println("\n5. 性能日志记录演示")

	logDir := filepath.Join(".", "demo-logs", "performance")
	os.MkdirAll(logDir, 0o755)

	config := mc.DefaultConfig()
	config.EnableLog = true
	config.LogLevel = "INFO"
	config.LogFormat = "json" // JSON格式便于性能数据分析
	config.LogDirectory = logDir
	config.PerformanceMode = mc.PerformanceModeSpeedOptimized

	chunker := mc.NewMarkdownChunkerWithConfig(config)

	// 多次处理以生成性能日志
	fmt.Printf("  执行多次处理以生成性能日志...\n")
	for i := 0; i < 3; i++ {
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("    处理 %d 错误: %v", i+1, err)
			continue
		}
		fmt.Printf("    处理 %d: %d 个块\n", i+1, len(chunks))
		time.Sleep(100 * time.Millisecond) // 短暂延迟
	}

	// 获取性能统计
	stats := chunker.GetPerformanceStats()
	fmt.Printf("  性能统计:\n")
	fmt.Printf("    处理时间: %v\n", stats.ProcessingTime)
	fmt.Printf("    内存使用: %d KB\n", stats.MemoryUsed/1024)
	fmt.Printf("    处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)
	fmt.Printf("  日志目录: %s\n", logDir)

	showLogFiles(logDir)
}

// demonstrateCustomLogDirectory 演示自定义日志目录
func demonstrateCustomLogDirectory(markdown string) {
	fmt.Println("\n6. 自定义日志目录演示")

	// 创建多个自定义目录
	customDirs := []string{
		filepath.Join(".", "custom-logs", "project-a"),
		filepath.Join(".", "custom-logs", "project-b"),
		filepath.Join(".", "custom-logs", "project-c"),
	}

	for i, logDir := range customDirs {
		fmt.Printf("  测试自定义目录 %d: %s\n", i+1, logDir)

		os.MkdirAll(logDir, 0o755)

		config := mc.DefaultConfig()
		config.EnableLog = true
		config.LogLevel = "INFO"
		config.LogFormat = "console"
		config.LogDirectory = logDir

		chunker := mc.NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			log.Printf("    处理错误: %v", err)
			continue
		}

		fmt.Printf("    处理结果: %d 个块\n", len(chunks))
		fmt.Printf("    日志文件: ")
		showLogFiles(logDir)
	}
}

// demonstrateLoggingWithConfiguration 演示日志与配置的结合使用
func demonstrateLoggingWithConfiguration(markdown string) {
	fmt.Println("\n7. 日志与配置结合使用演示")

	logDir := filepath.Join(".", "demo-logs", "configuration")
	os.MkdirAll(logDir, 0o755)

	// 创建复杂配置
	config := mc.DefaultConfig()
	config.EnableLog = true
	config.LogLevel = "DEBUG"
	config.LogFormat = "json"
	config.LogDirectory = logDir

	// 配置其他功能
	config.MaxChunkSize = 1000
	config.ErrorHandling = mc.ErrorModePermissive
	config.PerformanceMode = mc.PerformanceModeSpeedOptimized
	config.FilterEmptyChunks = true
	config.EnableObjectPooling = true

	// 添加自定义元数据提取器
	config.CustomExtractors = []mc.MetadataExtractor{
		&mc.LinkExtractor{},
		&mc.ImageExtractor{},
		&mc.CodeComplexityExtractor{},
	}

	// 只启用特定内容类型
	config.EnabledTypes = map[string]bool{
		"heading":        true,
		"paragraph":      true,
		"code":           true,
		"table":          true,
		"list":           false, // 禁用列表处理
		"blockquote":     true,
		"thematic_break": false, // 禁用分隔线处理
	}

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(markdown))

	fmt.Printf("  复杂配置处理结果:\n")
	fmt.Printf("    块数量: %d\n", len(chunks))
	fmt.Printf("    返回错误: %v\n", err)
	fmt.Printf("    记录的错误数: %d\n", len(chunker.GetErrors()))

	// 统计块类型
	typeCount := make(map[string]int)
	for _, chunk := range chunks {
		typeCount[chunk.Type]++
	}

	fmt.Printf("    块类型分布:\n")
	for chunkType, count := range typeCount {
		fmt.Printf("      %s: %d\n", chunkType, count)
	}

	// 性能统计
	stats := chunker.GetPerformanceStats()
	fmt.Printf("    性能统计:\n")
	fmt.Printf("      处理时间: %v\n", stats.ProcessingTime)
	fmt.Printf("      内存使用: %d KB\n", stats.MemoryUsed/1024)
	fmt.Printf("      处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)

	fmt.Printf("  日志目录: %s\n", logDir)
	showLogFiles(logDir)
}

// showLogFiles 显示日志文件信息
func showLogFiles(logDir string) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		fmt.Printf("    无法读取日志目录: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Printf("    没有生成日志文件\n")
		return
	}

	fmt.Printf("    日志文件:\n")
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				continue
			}
			fmt.Printf("      - %s (%d 字节)\n", file.Name(), info.Size())
		}
	}
}

// showLogFileSize 显示日志文件大小
func showLogFileSize(logDir string) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	var totalSize int64
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
		}
	}

	fmt.Printf("    日志总大小: %d 字节\n", totalSize)
}

// showLogSample 显示日志文件内容示例
func showLogSample(logDir, format string) {
	files, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(logDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			// 只显示前200个字符
			sample := string(content)
			if len(sample) > 200 {
				sample = sample[:200] + "..."
			}

			fmt.Printf("    %s格式日志示例:\n", format)
			fmt.Printf("      %s\n", sample)
			break
		}
	}
}
