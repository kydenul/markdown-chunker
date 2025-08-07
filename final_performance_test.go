package markdownchunker

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestFinalPerformanceDemo 最终性能演示测试
func TestFinalPerformanceDemo(t *testing.T) {
	t.Log("=== 最终性能演示测试 ===")

	// 1. 性能监控演示
	t.Log("\n1. 性能监控功能演示:")
	chunker := NewMarkdownChunker()

	// 创建测试文档
	content := createLargeTestDocument(1000)

	// 处理文档并获取性能统计
	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("处理文档失败: %v", err)
	}

	stats := chunker.GetPerformanceStats()
	t.Logf("   处理时间: %v", stats.ProcessingTime)
	t.Logf("   总块数: %d", stats.TotalChunks)
	t.Logf("   输入字节数: %d", stats.TotalBytes)
	t.Logf("   块字节数: %d", stats.ChunkBytes)
	t.Logf("   处理速度: %.2f 块/秒", stats.ChunksPerSecond)
	t.Logf("   吞吐量: %.2f 字节/秒", stats.BytesPerSecond)
	t.Logf("   内存使用: %d 字节", stats.MemoryUsed)
	t.Logf("   峰值内存: %d 字节", stats.PeakMemory)

	// 2. 内存优化演示
	t.Log("\n2. 内存优化功能演示:")
	optimizer := NewMemoryOptimizer(50 * 1024 * 1024) // 50MB限制

	// 测试对象池
	t.Log("   对象池测试:")
	poolStart := time.Now()
	for i := 0; i < 1000; i++ {
		chunk := optimizer.GetChunk()
		chunk.ID = i
		chunk.Type = "test"
		chunk.Content = fmt.Sprintf("Content %d", i)
		optimizer.PutChunk(chunk)
	}
	poolTime := time.Since(poolStart)
	t.Logf("   对象池操作时间: %v (1000次获取/归还)", poolTime)

	// 测试字符串优化
	ops := NewOptimizedStringOperations()
	strs := []string{"hello", "world", "test", "performance"}

	optimizedStart := time.Now()
	for i := 0; i < 10000; i++ {
		ops.JoinStrings(strs, ", ")
	}
	optimizedTime := time.Since(optimizedStart)
	t.Logf("   优化字符串操作时间: %v (10000次连接)", optimizedTime)

	// 获取内存统计
	memStats := optimizer.GetMemoryStats()
	t.Logf("   当前内存: %d 字节", memStats.CurrentMemory)
	t.Logf("   内存限制: %d 字节", memStats.MemoryLimit)
	t.Logf("   GC周期: %d", memStats.GCCycles)

	// 3. 并发安全性演示
	t.Log("\n3. 并发安全性功能演示:")
	concurrentChunker := NewConcurrentChunker(DefaultConfig())

	// 准备多个文档
	documents := make([][]byte, 50)
	for i := range documents {
		documents[i] = []byte(fmt.Sprintf("# Document %d\n\nContent for document %d", i, i))
	}

	// 并发处理
	concurrentStart := time.Now()
	concurrentStats, results, errors := concurrentChunker.ProcessDocumentsConcurrently(documents, runtime.NumCPU())
	concurrentTime := time.Since(concurrentStart)

	t.Logf("   并发处理时间: %v", concurrentTime)
	t.Logf("   处理文档数: %d", concurrentStats.ProcessedDocuments)
	t.Logf("   失败文档数: %d", concurrentStats.FailedDocuments)
	t.Logf("   总块数: %d", concurrentStats.TotalChunks)
	t.Logf("   并发度: %d", concurrentStats.Concurrency)
	t.Logf("   文档吞吐量: %.2f 文档/秒", concurrentStats.ThroughputDocs)
	t.Logf("   块吞吐量: %.2f 块/秒", concurrentStats.ThroughputChunks)

	// 验证结果
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}
	t.Logf("   成功处理: %d/%d 文档", successCount, len(documents))

	// 4. 工作池演示
	t.Log("\n4. 工作池功能演示:")
	pool := NewWorkerPool(4, DefaultConfig())

	poolStart = time.Now()
	poolResults, poolErrors := pool.ProcessBatch(documents)
	poolTime = time.Since(poolStart)

	t.Logf("   工作池处理时间: %v", poolTime)
	t.Logf("   工作协程数: 4")

	poolSuccessCount := 0
	for _, err := range poolErrors {
		if err == nil {
			poolSuccessCount++
		}
	}
	t.Logf("   成功处理: %d/%d 文档", poolSuccessCount, len(documents))

	// 5. 性能对比总结
	t.Log("\n5. 性能对比总结:")

	// 顺序处理时间估算
	sequentialTime := stats.ProcessingTime * time.Duration(len(documents))

	t.Logf("   估算顺序处理时间: %v", sequentialTime)
	t.Logf("   实际并发处理时间: %v", concurrentTime)
	t.Logf("   工作池处理时间: %v", poolTime)

	if concurrentTime < sequentialTime {
		speedup := float64(sequentialTime) / float64(concurrentTime)
		t.Logf("   并发加速比: %.2fx", speedup)
	}

	if poolTime < sequentialTime {
		speedup := float64(sequentialTime) / float64(poolTime)
		t.Logf("   工作池加速比: %.2fx", speedup)
	}

	// 验证所有功能正常工作
	if len(chunks) == 0 {
		t.Error("性能监控测试失败：没有生成块")
	}

	if stats.ProcessingTime <= 0 {
		t.Error("性能监控测试失败：处理时间无效")
	}

	if len(results) != len(documents) {
		t.Error("并发处理测试失败：结果数量不匹配")
	}

	if len(poolResults) != len(documents) {
		t.Error("工作池测试失败：结果数量不匹配")
	}

	t.Log("\n=== 所有功能测试完成 ===")
}

// createLargeTestDocument 创建大型测试文档
func createLargeTestDocument(sections int) string {
	var doc strings.Builder

	doc.WriteString("# 性能测试文档\n\n")
	doc.WriteString("这是一个用于性能测试的大型Markdown文档。\n\n")

	for i := 0; i < sections; i++ {
		doc.WriteString(fmt.Sprintf("## 第%d节\n\n", i+1))
		doc.WriteString(fmt.Sprintf("这是第%d节的内容，包含一些文本来测试处理性能。", i+1))
		doc.WriteString("这里有更多的文本内容，用于增加文档的大小和复杂性。\n\n")

		if i%10 == 0 {
			doc.WriteString("```go\n")
			doc.WriteString(fmt.Sprintf("func section%d() {\n", i))
			doc.WriteString("    fmt.Println(\"Hello from section\")\n")
			doc.WriteString("    // 这是一些示例代码\n")
			doc.WriteString("    for i := 0; i < 10; i++ {\n")
			doc.WriteString("        process(i)\n")
			doc.WriteString("    }\n")
			doc.WriteString("}\n")
			doc.WriteString("```\n\n")
		}

		if i%15 == 0 {
			doc.WriteString("| 列1 | 列2 | 列3 |\n")
			doc.WriteString("|-----|-----|-----|\n")
			doc.WriteString(fmt.Sprintf("| 数据%d | 值%d | 结果%d |\n", i, i*2, i*3))
			doc.WriteString(fmt.Sprintf("| 数据%d | 值%d | 结果%d |\n\n", i+1, (i+1)*2, (i+1)*3))
		}

		if i%20 == 0 {
			doc.WriteString("- 列表项1\n")
			doc.WriteString("- 列表项2\n")
			doc.WriteString("- 列表项3\n\n")

			doc.WriteString("> 这是一个引用块\n")
			doc.WriteString("> 包含一些重要信息\n\n")
		}
	}

	return doc.String()
}
