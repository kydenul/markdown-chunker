package markdownchunker

import (
	"strings"
	"testing"
	"time"
)

func TestPerformanceMonitor_BasicFunctionality(t *testing.T) {
	pm := NewPerformanceMonitor()

	// 测试初始状态
	if pm.IsRunning() {
		t.Error("新创建的监控器不应该处于运行状态")
	}

	// 测试开始监控
	pm.Start()
	if !pm.IsRunning() {
		t.Error("开始监控后应该处于运行状态")
	}

	// 模拟处理一些块
	chunk1 := &Chunk{
		ID:      1,
		Type:    "paragraph",
		Content: "This is a test paragraph with some content.",
		Text:    "This is a test paragraph with some content.",
	}

	chunk2 := &Chunk{
		ID:      2,
		Type:    "heading",
		Content: "# Test Heading",
		Text:    "Test Heading",
	}

	// 记录输入字节数
	pm.RecordBytes(100)

	pm.RecordChunk(chunk1)
	pm.RecordChunk(chunk2)

	// 等待一小段时间以确保有处理时间
	time.Sleep(10 * time.Millisecond)

	// 停止监控
	pm.Stop()
	if pm.IsRunning() {
		t.Error("停止监控后不应该处于运行状态")
	}

	// 获取统计信息
	stats := pm.GetStats()

	// 验证统计信息
	if stats.TotalChunks != 2 {
		t.Errorf("期望处理2个块，实际处理了%d个", stats.TotalChunks)
	}

	if stats.ProcessingTime <= 0 {
		t.Error("处理时间应该大于0")
	}

	if stats.TotalBytes <= 0 {
		t.Error("输入字节数应该大于0")
	}

	if stats.ChunkBytes <= 0 {
		t.Error("块字节数应该大于0")
	}

	if stats.ChunksPerSecond <= 0 {
		t.Error("每秒处理块数应该大于0")
	}

	if stats.BytesPerSecond <= 0 {
		t.Error("每秒处理字节数应该大于0")
	}
}

func TestPerformanceMonitor_Reset(t *testing.T) {
	pm := NewPerformanceMonitor()

	// 开始监控并记录一些数据
	pm.Start()
	chunk := &Chunk{
		ID:      1,
		Type:    "paragraph",
		Content: "Test content",
		Text:    "Test content",
	}
	pm.RecordChunk(chunk)
	pm.Stop()

	// 获取统计信息
	stats := pm.GetStats()
	if stats.TotalChunks == 0 {
		t.Error("应该有处理的块")
	}

	// 重置监控器
	pm.Reset()

	// 验证重置后的状态
	if pm.IsRunning() {
		t.Error("重置后不应该处于运行状态")
	}

	stats = pm.GetStats()
	if stats.TotalChunks != 0 {
		t.Error("重置后块数应该为0")
	}

	if stats.TotalBytes != 0 {
		t.Error("重置后输入字节数应该为0")
	}

	if stats.ChunkBytes != 0 {
		t.Error("重置后块字节数应该为0")
	}

	if stats.ProcessingTime != 0 {
		t.Error("重置后处理时间应该为0")
	}
}

func TestPerformanceMonitor_RecordBytes(t *testing.T) {
	pm := NewPerformanceMonitor()

	pm.Start()

	// 记录字节数
	pm.RecordBytes(1000)
	pm.RecordBytes(500)

	pm.Stop()

	stats := pm.GetStats()
	if stats.TotalBytes != 1500 {
		t.Errorf("期望输入字节数为1500，实际为%d", stats.TotalBytes)
	}
}

func TestPerformanceMonitor_NotRunning(t *testing.T) {
	pm := NewPerformanceMonitor()

	// 在未开始监控时记录数据
	chunk := &Chunk{
		ID:      1,
		Type:    "paragraph",
		Content: "Test content",
		Text:    "Test content",
	}
	pm.RecordChunk(chunk)
	pm.RecordBytes(100)

	stats := pm.GetStats()

	// 验证未运行时不记录数据
	if stats.TotalChunks != 0 {
		t.Error("未运行时不应该记录块数")
	}

	if stats.TotalBytes != 0 {
		t.Error("未运行时不应该记录输入字节数")
	}

	if stats.ChunkBytes != 0 {
		t.Error("未运行时不应该记录块字节数")
	}
}

func TestMarkdownChunker_PerformanceIntegration(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 测试文档
	content := `# Test Document

This is a paragraph with some content.

## Section 2

Another paragraph here.

` + "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```" + `

- List item 1
- List item 2
- List item 3

> This is a blockquote
> with multiple lines

| Column 1 | Column 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |
`

	// 处理文档
	chunks, err := chunker.ChunkDocument([]byte(content))
	if err != nil {
		t.Fatalf("处理文档时出错: %v", err)
	}

	// 获取性能统计
	stats := chunker.GetPerformanceStats()

	// 验证统计信息
	if stats.TotalChunks != len(chunks) {
		t.Errorf("统计的块数(%d)与实际块数(%d)不匹配", stats.TotalChunks, len(chunks))
	}

	if stats.TotalBytes != int64(len(content)) {
		t.Errorf("统计的输入字节数(%d)与实际字节数(%d)不匹配", stats.TotalBytes, len(content))
	}

	if stats.ChunkBytes <= 0 {
		t.Error("块内容字节数应该大于0")
	}

	if stats.ProcessingTime <= 0 {
		t.Error("处理时间应该大于0")
	}

	if stats.ChunksPerSecond <= 0 {
		t.Error("每秒处理块数应该大于0")
	}

	if stats.BytesPerSecond <= 0 {
		t.Error("每秒处理字节数应该大于0")
	}

	// 测试性能监控器访问方法
	monitor := chunker.GetPerformanceMonitor()
	if monitor == nil {
		t.Error("应该能够获取性能监控器")
	}

	// 测试重置
	chunker.ResetPerformanceMonitor()
	resetStats := chunker.GetPerformanceStats()
	if resetStats.TotalChunks != 0 {
		t.Error("重置后块数应该为0")
	}
}

func TestPerformanceMonitor_LargeDocument(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 创建一个大文档
	var builder strings.Builder
	for i := 0; i < 1000; i++ {
		builder.WriteString("# Heading ")
		builder.WriteString(string(rune('A' + i%26)))
		builder.WriteString("\n\n")
		builder.WriteString("This is paragraph number ")
		builder.WriteString(string(rune('0' + i%10)))
		builder.WriteString(" with some content to make it longer and more realistic.")
		builder.WriteString("\n\n")
	}

	content := builder.String()

	// 处理大文档
	startTime := time.Now()
	chunks, err := chunker.ChunkDocument([]byte(content))
	processingTime := time.Since(startTime)

	if err != nil {
		t.Fatalf("处理大文档时出错: %v", err)
	}

	// 获取性能统计
	stats := chunker.GetPerformanceStats()

	// 验证统计信息
	if stats.TotalChunks != len(chunks) {
		t.Errorf("统计的块数(%d)与实际块数(%d)不匹配", stats.TotalChunks, len(chunks))
	}

	if len(chunks) < 1000 { // 至少应该有1000个标题
		t.Errorf("期望至少1000个块，实际得到%d个", len(chunks))
	}

	// 验证处理时间合理性（应该与实际处理时间接近）
	timeDiff := stats.ProcessingTime - processingTime
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// 允许10%的误差
	if float64(timeDiff) > float64(processingTime)*0.1 {
		t.Errorf("统计的处理时间(%v)与实际处理时间(%v)差异过大", stats.ProcessingTime, processingTime)
	}

	t.Logf("大文档性能统计:")
	t.Logf("  处理时间: %v", stats.ProcessingTime)
	t.Logf("  总块数: %d", stats.TotalChunks)
	t.Logf("  总字节数: %d", stats.TotalBytes)
	t.Logf("  每秒处理块数: %.2f", stats.ChunksPerSecond)
	t.Logf("  每秒处理字节数: %.2f", stats.BytesPerSecond)
	t.Logf("  内存使用: %d bytes", stats.MemoryUsed)
	t.Logf("  峰值内存: %d bytes", stats.PeakMemory)
}

func TestPerformanceMonitor_MemoryTracking(t *testing.T) {
	pm := NewPerformanceMonitor()

	pm.Start()

	// 创建一些大的块来测试内存跟踪
	for i := 0; i < 100; i++ {
		chunk := &Chunk{
			ID:      i,
			Type:    "paragraph",
			Content: strings.Repeat("This is a large content block. ", 100),
			Text:    strings.Repeat("This is a large content block. ", 100),
		}
		pm.RecordChunk(chunk)
	}

	pm.Stop()

	stats := pm.GetStats()

	// 验证内存统计
	if stats.PeakMemory <= 0 {
		t.Error("峰值内存应该大于0")
	}

	// 内存使用可能为负（如果GC运行），但峰值内存应该是正数
	if stats.PeakMemory < stats.MemoryUsed {
		t.Error("峰值内存应该大于等于当前内存使用")
	}
}

func BenchmarkPerformanceMonitor_RecordChunk(b *testing.B) {
	pm := NewPerformanceMonitor()
	pm.Start()

	chunk := &Chunk{
		ID:      1,
		Type:    "paragraph",
		Content: "Test content for benchmarking",
		Text:    "Test content for benchmarking",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.RecordChunk(chunk)
	}
}

func BenchmarkPerformanceMonitor_GetStats(b *testing.B) {
	pm := NewPerformanceMonitor()
	pm.Start()

	// 记录一些数据
	chunk := &Chunk{
		ID:      1,
		Type:    "paragraph",
		Content: "Test content",
		Text:    "Test content",
	}
	pm.RecordChunk(chunk)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.GetStats()
	}
}
