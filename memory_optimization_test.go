package markdownchunker

import (
	"strings"
	"testing"
	"time"
)

func TestChunkPool_BasicFunctionality(t *testing.T) {
	pool := NewChunkPool()

	// Get a chunk from the pool
	chunk1 := pool.Get()
	if chunk1 == nil {
		t.Fatal("Expected to get a chunk from pool")
	}

	// Verify chunk is properly initialized
	if chunk1.Metadata == nil {
		t.Error("Expected chunk to have initialized metadata")
	}

	// Modify the chunk
	chunk1.ID = 1
	chunk1.Type = "test"
	chunk1.Content = "test content"
	chunk1.Metadata["key"] = "value"

	// Put it back in the pool
	pool.Put(chunk1)

	// Get another chunk
	chunk2 := pool.Get()
	if chunk2 == nil {
		t.Fatal("Expected to get a chunk from pool")
	}

	// Verify chunk is reset
	if chunk2.ID != 0 {
		t.Errorf("Expected chunk ID to be reset to 0, got %d", chunk2.ID)
	}
	if chunk2.Type != "" {
		t.Errorf("Expected chunk type to be reset to empty, got %s", chunk2.Type)
	}
	if chunk2.Content != "" {
		t.Errorf("Expected chunk content to be reset to empty, got %s", chunk2.Content)
	}
	if len(chunk2.Metadata) != 0 {
		t.Errorf("Expected chunk metadata to be reset to empty, got %v", chunk2.Metadata)
	}
}

func TestStringBuilderPool_BasicFunctionality(t *testing.T) {
	pool := NewStringBuilderPool()

	// Get a string builder from the pool
	sb1 := pool.Get()
	if sb1 == nil {
		t.Fatal("Expected to get a string builder from pool")
	}

	// Use the string builder
	sb1.WriteString("test content")
	if sb1.String() != "test content" {
		t.Errorf("Expected 'test content', got %s", sb1.String())
	}

	// Put it back in the pool
	pool.Put(sb1)

	// Get another string builder
	sb2 := pool.Get()
	if sb2 == nil {
		t.Fatal("Expected to get a string builder from pool")
	}

	// Verify string builder is reset
	if sb2.Len() != 0 {
		t.Errorf("Expected string builder to be reset, got length %d", sb2.Len())
	}
	if sb2.String() != "" {
		t.Errorf("Expected empty string builder, got %s", sb2.String())
	}
}

func TestStringBuilderPool_LargeCapacityHandling(t *testing.T) {
	pool := NewStringBuilderPool()

	// Get a string builder and make it very large
	sb := pool.Get()
	largeContent := strings.Repeat("x", 100*1024) // 100KB
	sb.WriteString(largeContent)

	// Put it back - should not be returned to pool due to large capacity
	pool.Put(sb)

	// Get another string builder - should be a new one
	sb2 := pool.Get()
	if sb2.Cap() > 1024 { // Should be a fresh builder with small capacity
		t.Errorf("Expected fresh string builder with small capacity, got capacity %d", sb2.Cap())
	}
}

func TestMemoryLimiter_BasicFunctionality(t *testing.T) {
	// Set a reasonable memory limit (100MB)
	limiter := NewMemoryLimiter(100 * 1024 * 1024)

	// Check current memory usage (should not exceed limit)
	err := limiter.CheckMemoryLimit()
	if err != nil {
		t.Errorf("Unexpected memory limit error: %v", err)
	}

	// Get current memory usage
	currentMemory := limiter.GetCurrentMemoryUsage()
	if currentMemory <= 0 {
		t.Error("Expected positive current memory usage")
	}

	// Test memory limit getter/setter
	if limiter.GetMemoryLimit() != 100*1024*1024 {
		t.Errorf("Expected memory limit 100MB, got %d", limiter.GetMemoryLimit())
	}

	limiter.SetMemoryLimit(200 * 1024 * 1024)
	if limiter.GetMemoryLimit() != 200*1024*1024 {
		t.Errorf("Expected memory limit 200MB, got %d", limiter.GetMemoryLimit())
	}
}

func TestMemoryLimiter_ExceedsLimit(t *testing.T) {
	// Set a very low memory limit (1 byte) to trigger the limit
	limiter := NewMemoryLimiter(1)

	// Check memory limit - should exceed
	err := limiter.CheckMemoryLimit()
	if err == nil {
		t.Error("Expected memory limit error with 1 byte limit")
	}

	// Verify error type
	if chunkerErr, ok := err.(*ChunkerError); ok {
		if chunkerErr.Type != ErrorTypeMemoryExhausted {
			t.Errorf("Expected ErrorTypeMemoryExhausted, got %v", chunkerErr.Type)
		}
	} else {
		t.Errorf("Expected ChunkerError, got %T", err)
	}
}

func TestMemoryOptimizer_BasicFunctionality(t *testing.T) {
	optimizer := NewMemoryOptimizer(100 * 1024 * 1024) // 100MB limit

	// Test chunk pool functionality
	chunk := optimizer.GetChunk()
	if chunk == nil {
		t.Fatal("Expected to get a chunk from optimizer")
	}

	chunk.ID = 1
	chunk.Type = "test"
	optimizer.PutChunk(chunk)

	// Test string builder pool functionality
	sb := optimizer.GetStringBuilder()
	if sb == nil {
		t.Fatal("Expected to get a string builder from optimizer")
	}

	sb.WriteString("test")
	optimizer.PutStringBuilder(sb)

	// Test memory limit checking
	err := optimizer.CheckMemoryLimit()
	if err != nil {
		t.Errorf("Unexpected memory limit error: %v", err)
	}
}

func TestMemoryOptimizer_GCThreshold(t *testing.T) {
	optimizer := NewMemoryOptimizer(0) // No memory limit

	// Test GC threshold getter/setter
	if optimizer.GetGCThreshold() != 10*1024*1024 { // Default 10MB
		t.Errorf("Expected default GC threshold 10MB, got %d", optimizer.GetGCThreshold())
	}

	optimizer.SetGCThreshold(5 * 1024 * 1024) // 5MB
	if optimizer.GetGCThreshold() != 5*1024*1024 {
		t.Errorf("Expected GC threshold 5MB, got %d", optimizer.GetGCThreshold())
	}

	// Test processed bytes recording
	optimizer.RecordProcessedBytes(1024)
	optimizer.RecordProcessedBytes(2048)

	// Force GC
	optimizer.ForceGC()

	// Test reset
	optimizer.Reset()
}

func TestMemoryOptimizer_Stats(t *testing.T) {
	optimizer := NewMemoryOptimizer(100 * 1024 * 1024)

	// Get memory stats
	stats := optimizer.GetMemoryStats()

	// Verify stats structure
	if stats.CurrentMemory <= 0 {
		t.Error("Expected positive current memory")
	}
	if stats.MemoryLimit != 100*1024*1024 {
		t.Errorf("Expected memory limit 100MB, got %d", stats.MemoryLimit)
	}
	if stats.GCThreshold != 10*1024*1024 {
		t.Errorf("Expected GC threshold 10MB, got %d", stats.GCThreshold)
	}
	if stats.TotalAllocations <= 0 {
		t.Error("Expected positive total allocations")
	}
	if stats.GCCycles < 0 {
		t.Error("Expected non-negative GC cycles")
	}
}

func TestOptimizedStringOperations_JoinStrings(t *testing.T) {
	ops := NewOptimizedStringOperations()

	// Test empty slice
	result := ops.JoinStrings([]string{}, ",")
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}

	// Test single string
	result = ops.JoinStrings([]string{"hello"}, ",")
	if result != "hello" {
		t.Errorf("Expected 'hello', got %s", result)
	}

	// Test multiple strings
	result = ops.JoinStrings([]string{"hello", "world", "test"}, ", ")
	if result != "hello, world, test" {
		t.Errorf("Expected 'hello, world, test', got %s", result)
	}
}

func TestOptimizedStringOperations_TrimAndClean(t *testing.T) {
	ops := NewOptimizedStringOperations()

	// Test empty string
	result := ops.TrimAndClean("")
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}

	// Test string with extra whitespace
	result = ops.TrimAndClean("  hello   world  \n\t  ")
	if result != "hello world" {
		t.Errorf("Expected 'hello world', got %s", result)
	}

	// Test short string (uses standard library path)
	result = ops.TrimAndClean("  short  ")
	if result != "short" {
		t.Errorf("Expected 'short', got %s", result)
	}

	// Test long string (uses optimized path)
	longText := strings.Repeat("word ", 50) // 250 characters
	result = ops.TrimAndClean("  " + longText + "  ")
	expected := strings.TrimSpace(strings.Join(strings.Fields(longText), " "))
	if result != expected {
		t.Errorf("Long string optimization failed")
	}
}

func TestOptimizedStringOperations_BuildContent(t *testing.T) {
	ops := NewOptimizedStringOperations()

	// Test empty parts
	result := ops.BuildContent()
	if result != "" {
		t.Errorf("Expected empty string, got %s", result)
	}

	// Test with some empty parts
	result = ops.BuildContent("hello", "", "world", "", "test")
	if result != "helloworldtest" {
		t.Errorf("Expected 'helloworldtest', got %s", result)
	}

	// Test normal parts
	result = ops.BuildContent("part1", "part2", "part3")
	if result != "part1part2part3" {
		t.Errorf("Expected 'part1part2part3', got %s", result)
	}
}

func BenchmarkChunkPool_GetPut(b *testing.B) {
	pool := NewChunkPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunk := pool.Get()
		chunk.ID = i
		chunk.Type = "test"
		pool.Put(chunk)
	}
}

func BenchmarkStringBuilderPool_GetPut(b *testing.B) {
	pool := NewStringBuilderPool()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb := pool.Get()
		sb.WriteString("test content")
		pool.Put(sb)
	}
}

func BenchmarkOptimizedStringOperations_JoinStrings(b *testing.B) {
	ops := NewOptimizedStringOperations()
	strs := []string{"hello", "world", "test", "benchmark"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ops.JoinStrings(strs, ", ")
	}
}

func BenchmarkOptimizedStringOperations_TrimAndClean(b *testing.B) {
	ops := NewOptimizedStringOperations()
	text := "  " + strings.Repeat("word ", 100) + "  "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ops.TrimAndClean(text)
	}
}

func TestMemoryOptimizer_Integration(t *testing.T) {
	// Test memory optimizer integration with actual markdown processing
	optimizer := NewMemoryOptimizer(50 * 1024 * 1024) // 50MB limit

	// Create a large markdown document
	var content strings.Builder
	for i := 0; i < 1000; i++ {
		content.WriteString("# Heading ")
		content.WriteString(strings.Repeat("A", i%10+1))
		content.WriteString("\n\n")
		content.WriteString("This is paragraph ")
		content.WriteString(strings.Repeat("content ", i%5+1))
		content.WriteString("\n\n")
		if i%10 == 0 {
			content.WriteString("```go\nfunc test() {\n    // code\n}\n```\n\n")
		}
	}

	markdown := content.String()

	// Process with memory monitoring
	start := time.Now()
	initialStats := optimizer.GetMemoryStats()

	// Simulate chunk processing using the optimizer
	chunks := make([]*Chunk, 0, 1000)
	for i := 0; i < 1000; i++ {
		chunk := optimizer.GetChunk()
		chunk.ID = i
		chunk.Type = "paragraph"
		chunk.Content = markdown[i%len(markdown) : min(len(markdown), i%len(markdown)+100)]
		chunk.Text = chunk.Content
		chunks = append(chunks, chunk)

		// Record processed bytes
		optimizer.RecordProcessedBytes(int64(len(chunk.Content)))

		// Check memory limit periodically
		if i%100 == 0 {
			if err := optimizer.CheckMemoryLimit(); err != nil {
				t.Errorf("Memory limit exceeded at iteration %d: %v", i, err)
			}
		}
	}

	// Return chunks to pool
	for _, chunk := range chunks {
		optimizer.PutChunk(chunk)
	}

	processingTime := time.Since(start)
	finalStats := optimizer.GetMemoryStats()

	// Verify performance
	if processingTime > 5*time.Second {
		t.Errorf("Processing took too long: %v", processingTime)
	}

	// Verify memory stats
	if finalStats.CurrentMemory <= 0 {
		t.Error("Expected positive current memory")
	}

	if finalStats.TotalAllocations <= initialStats.TotalAllocations {
		t.Error("Expected total allocations to increase")
	}

	t.Logf("Memory optimization integration test:")
	t.Logf("  Processing time: %v", processingTime)
	t.Logf("  Initial memory: %d bytes", initialStats.CurrentMemory)
	t.Logf("  Final memory: %d bytes", finalStats.CurrentMemory)
	t.Logf("  Total allocations: %d", finalStats.TotalAllocations)
	t.Logf("  GC cycles: %d", finalStats.GCCycles)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
