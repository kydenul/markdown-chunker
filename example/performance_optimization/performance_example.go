package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	mc "github.com/kydenul/markdown-chunker"
)

func main() {
	fmt.Println("=== 性能优化使用示例 ===")

	// 生成大型测试文档
	largeMarkdown := generateLargeMarkdown()
	fmt.Printf("生成的测试文档大小: %d 字节\n", len(largeMarkdown))

	// 示例 1: 默认性能模式
	fmt.Println("\n1. 默认性能模式")
	runPerformanceTest("默认模式", mc.PerformanceModeDefault, largeMarkdown)

	// 示例 2: 内存优化模式
	fmt.Println("\n2. 内存优化模式")
	runPerformanceTest("内存优化", mc.PerformanceModeMemoryOptimized, largeMarkdown)

	// 示例 3: 速度优化模式
	fmt.Println("\n3. 速度优化模式")
	runPerformanceTest("速度优化", mc.PerformanceModeSpeedOptimized, largeMarkdown)

	// 示例 4: 内存限制配置
	fmt.Println("\n4. 内存限制配置")
	testMemoryLimit(largeMarkdown)

	// 示例 5: 对象池化
	fmt.Println("\n5. 对象池化性能测试")
	testObjectPooling(largeMarkdown)

	// 示例 6: 并发处理性能
	fmt.Println("\n6. 并发处理性能测试")
	testConcurrentProcessing(largeMarkdown)

	// 示例 7: 性能监控详细信息
	fmt.Println("\n7. 详细性能监控")
	demonstrateDetailedMonitoring(largeMarkdown)
}

// runPerformanceTest 运行性能测试
func runPerformanceTest(name string, mode mc.PerformanceMode, content string) {
	config := mc.DefaultConfig()
	config.PerformanceMode = mode
	config.EnableObjectPooling = true

	chunker := mc.NewMarkdownChunkerWithConfig(config)

	// 记录系统内存使用
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	start := time.Now()
	chunks, err := chunker.ChunkDocument([]byte(content))
	elapsed := time.Since(start)

	runtime.ReadMemStats(&m2)

	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	// 获取性能统计
	stats := chunker.GetPerformanceStats()

	fmt.Printf("%s结果:\n", name)
	fmt.Printf("  处理时间: %v\n", elapsed)
	fmt.Printf("  块数量: %d\n", len(chunks))
	fmt.Printf("  处理速度: %.2f 块/秒\n", stats.ChunksPerSecond)
	fmt.Printf("  字节处理速度: %.2f KB/秒\n", stats.BytesPerSecond/1024)
	fmt.Printf("  内存使用: %d KB\n", stats.MemoryUsed/1024)
	fmt.Printf("  峰值内存: %d KB\n", stats.PeakMemory/1024)
	fmt.Printf("  系统内存增长: %d KB\n", (m2.Alloc-m1.Alloc)/1024)

	// 检查错误
	if chunker.HasErrors() {
		fmt.Printf("  处理错误: %d\n", len(chunker.GetErrors()))
	}
}

// testMemoryLimit 测试内存限制功能
func testMemoryLimit(content string) {
	config := mc.DefaultConfig()
	config.MemoryLimit = 10 * 1024 * 1024 // 10MB 限制
	config.ErrorHandling = mc.ErrorModePermissive

	chunker := mc.NewMarkdownChunkerWithConfig(config)
	chunks, err := chunker.ChunkDocument([]byte(content))

	fmt.Printf("内存限制测试结果:\n")
	fmt.Printf("  内存限制: %d MB\n", config.MemoryLimit/(1024*1024))
	fmt.Printf("  处理结果: %v\n", err)
	fmt.Printf("  块数量: %d\n", len(chunks))

	stats := chunker.GetPerformanceStats()
	fmt.Printf("  实际内存使用: %d KB\n", stats.MemoryUsed/1024)

	if chunker.HasErrors() {
		memoryErrors := chunker.GetErrorsByType(mc.ErrorTypeMemoryExhausted)
		fmt.Printf("  内存不足错误: %d\n", len(memoryErrors))
	}
}

// testObjectPooling 测试对象池化性能
func testObjectPooling(content string) {
	// 不使用对象池化
	config1 := mc.DefaultConfig()
	config1.EnableObjectPooling = false

	chunker1 := mc.NewMarkdownChunkerWithConfig(config1)

	start1 := time.Now()
	for i := 0; i < 5; i++ {
		chunker1.ChunkDocument([]byte(content))
		chunker1.ResetPerformanceMonitor()
	}
	elapsed1 := time.Since(start1)

	// 使用对象池化
	config2 := mc.DefaultConfig()
	config2.EnableObjectPooling = true

	chunker2 := mc.NewMarkdownChunkerWithConfig(config2)

	start2 := time.Now()
	for i := 0; i < 5; i++ {
		chunker2.ChunkDocument([]byte(content))
		chunker2.ResetPerformanceMonitor()
	}
	elapsed2 := time.Since(start2)

	fmt.Printf("对象池化性能对比 (5次处理):\n")
	fmt.Printf("  不使用池化: %v\n", elapsed1)
	fmt.Printf("  使用池化: %v\n", elapsed2)
	fmt.Printf("  性能提升: %.2f%%\n", float64(elapsed1-elapsed2)/float64(elapsed1)*100)
}

// testConcurrentProcessing 测试并发处理性能
func testConcurrentProcessing(content string) {
	documents := []string{content, content, content, content}

	// 串行处理
	start1 := time.Now()
	for _, doc := range documents {
		chunker := mc.NewMarkdownChunker()
		chunker.ChunkDocument([]byte(doc))
	}
	serialTime := time.Since(start1)

	// 并发处理
	start2 := time.Now()
	done := make(chan bool, len(documents))

	for _, doc := range documents {
		go func(d string) {
			chunker := mc.NewMarkdownChunker()
			chunker.ChunkDocument([]byte(d))
			done <- true
		}(doc)
	}

	for i := 0; i < len(documents); i++ {
		<-done
	}
	concurrentTime := time.Since(start2)

	fmt.Printf("并发处理性能对比 (%d个文档):\n", len(documents))
	fmt.Printf("  串行处理: %v\n", serialTime)
	fmt.Printf("  并发处理: %v\n", concurrentTime)
	fmt.Printf("  性能提升: %.2f%%\n", float64(serialTime-concurrentTime)/float64(serialTime)*100)
}

// demonstrateDetailedMonitoring 演示详细的性能监控
func demonstrateDetailedMonitoring(content string) {
	config := mc.DefaultConfig()
	config.PerformanceMode = mc.PerformanceModeDefault

	chunker := mc.NewMarkdownChunkerWithConfig(config)

	// 获取性能监控器
	monitor := chunker.GetPerformanceMonitor()

	fmt.Printf("处理前监控器状态: 运行中=%t\n", monitor.IsRunning())

	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		log.Printf("错误: %v", err)
		return
	}

	fmt.Printf("处理后监控器状态: 运行中=%t\n", monitor.IsRunning())

	// 获取详细统计信息
	stats := monitor.GetStats()
	fmt.Printf("详细性能统计:\n")
	fmt.Printf("  处理时间: %v\n", stats.ProcessingTime)
	fmt.Printf("  总块数: %d\n", stats.TotalChunks)
	fmt.Printf("  总字节数: %d\n", stats.TotalBytes)
	fmt.Printf("  块内容字节数: %d\n", stats.ChunkBytes)
	fmt.Printf("  处理效率: %.2f 块/秒\n", stats.ChunksPerSecond)
	fmt.Printf("  字节处理效率: %.2f 字节/秒\n", stats.BytesPerSecond)
	fmt.Printf("  内存使用: %d 字节\n", stats.MemoryUsed)
	fmt.Printf("  峰值内存: %d 字节\n", stats.PeakMemory)

	// 获取系统内存统计
	memStats := monitor.GetMemoryStats()
	fmt.Printf("系统内存统计:\n")
	fmt.Printf("  当前分配: %d KB\n", memStats.Alloc/1024)
	fmt.Printf("  总分配: %d KB\n", memStats.TotalAlloc/1024)
	fmt.Printf("  系统内存: %d KB\n", memStats.Sys/1024)
	fmt.Printf("  GC次数: %d\n", memStats.NumGC)

	fmt.Printf("处理结果: %d 个块\n", len(chunks))
}

// generateLargeMarkdown 生成大型测试文档
func generateLargeMarkdown() string {
	var builder strings.Builder

	// 添加标题层次结构
	for i := 1; i <= 6; i++ {
		builder.WriteString(fmt.Sprintf("%s 标题级别 %d\n\n", strings.Repeat("#", i), i))

		// 添加段落
		for j := 0; j < 3; j++ {
			builder.WriteString(fmt.Sprintf("这是标题级别 %d 下的第 %d 个段落。", i, j+1))
			builder.WriteString("它包含了一些测试内容，用于性能测试。")
			builder.WriteString("段落中可能包含[链接](https://example.com)和![图片](image.jpg)。\n\n")
		}

		// 添加代码块
		builder.WriteString("```go\n")
		builder.WriteString("func exampleFunction() {\n")
		for k := 0; k < 10; k++ {
			builder.WriteString(fmt.Sprintf("    fmt.Println(\"Line %d\")\n", k+1))
		}
		builder.WriteString("}\n")
		builder.WriteString("```\n\n")

		// 添加表格
		builder.WriteString("| 列1 | 列2 | 列3 | 列4 |\n")
		builder.WriteString("|-----|-----|-----|-----|\n")
		for l := 0; l < 5; l++ {
			builder.WriteString(fmt.Sprintf("| 数据%d | 值%d | 信息%d | 内容%d |\n", l+1, l+1, l+1, l+1))
		}
		builder.WriteString("\n")

		// 添加列表
		builder.WriteString("有序列表:\n")
		for m := 1; m <= 5; m++ {
			builder.WriteString(fmt.Sprintf("%d. 列表项 %d\n", m, m))
		}
		builder.WriteString("\n无序列表:\n")
		for n := 0; n < 5; n++ {
			builder.WriteString(fmt.Sprintf("- 项目 %d\n", n+1))
		}
		builder.WriteString("\n")

		// 添加引用块
		builder.WriteString("> 这是一个引用块，包含一些重要信息。\n")
		builder.WriteString("> 它可能跨越多行，用于测试引用块的处理性能。\n\n")

		// 添加分隔线
		builder.WriteString("---\n\n")
	}

	return builder.String()
}
