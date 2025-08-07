package markdownchunker

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestConcurrentChunker_BasicFunctionality(t *testing.T) {
	chunker := NewConcurrentChunker(DefaultConfig())

	content := []byte("# Test Heading\n\nThis is a test paragraph.")

	chunks, err := chunker.ChunkDocument(content)
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
}

func TestConcurrentChunker_ThreadSafety(t *testing.T) {
	chunker := NewConcurrentChunker(DefaultConfig())

	const numGoroutines = 10
	const numIterations = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				content := []byte(fmt.Sprintf("# Heading %d-%d\n\nParagraph %d-%d", goroutineID, j, goroutineID, j))

				chunks, err := chunker.ChunkDocument(content)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d: %v", goroutineID, j, err)
					return
				}

				if len(chunks) != 2 {
					errors <- fmt.Errorf("goroutine %d, iteration %d: expected 2 chunks, got %d", goroutineID, j, len(chunks))
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

func TestConcurrentChunker_ChunkDocumentConcurrent(t *testing.T) {
	chunker := NewConcurrentChunker(DefaultConfig())

	// Prepare test documents
	contents := make([][]byte, 10)
	for i := range contents {
		contents[i] = []byte(fmt.Sprintf("# Document %d\n\nThis is document %d content.", i, i))
	}

	results, errors := chunker.ChunkDocumentConcurrent(contents)

	if len(results) != len(contents) {
		t.Errorf("Expected %d results, got %d", len(contents), len(results))
	}

	if len(errors) != len(contents) {
		t.Errorf("Expected %d error entries, got %d", len(contents), len(errors))
	}

	// Verify results
	for i, chunks := range results {
		if errors[i] != nil {
			t.Errorf("Document %d processing failed: %v", i, errors[i])
			continue
		}

		if len(chunks) != 2 {
			t.Errorf("Document %d: expected 2 chunks, got %d", i, len(chunks))
		}
	}
}

func TestConcurrentChunker_ChunkDocumentBatch(t *testing.T) {
	chunker := NewConcurrentChunker(DefaultConfig())

	// Prepare test documents
	contents := make([][]byte, 20)
	for i := range contents {
		contents[i] = []byte(fmt.Sprintf("# Document %d\n\nThis is document %d content.\n\n```go\nfunc test%d() {}\n```", i, i, i))
	}

	maxConcurrency := 5
	results, errors := chunker.ChunkDocumentBatch(contents, maxConcurrency)

	if len(results) != len(contents) {
		t.Errorf("Expected %d results, got %d", len(contents), len(results))
	}

	if len(errors) != len(contents) {
		t.Errorf("Expected %d error entries, got %d", len(contents), len(errors))
	}

	// Verify results
	for i, chunks := range results {
		if errors[i] != nil {
			t.Errorf("Document %d processing failed: %v", i, errors[i])
			continue
		}

		if len(chunks) != 3 { // heading + paragraph + code
			t.Errorf("Document %d: expected 3 chunks, got %d", i, len(chunks))
		}
	}
}

func TestConcurrentChunker_ProcessDocumentsConcurrently(t *testing.T) {
	chunker := NewConcurrentChunker(DefaultConfig())

	// Prepare test documents
	contents := make([][]byte, 15)
	for i := range contents {
		contents[i] = []byte(fmt.Sprintf("# Document %d\n\nThis is document %d content.", i, i))
	}

	maxConcurrency := 3
	stats, results, errors := chunker.ProcessDocumentsConcurrently(contents, maxConcurrency)

	// Verify stats
	if stats.TotalDocuments != len(contents) {
		t.Errorf("Expected %d total documents, got %d", len(contents), stats.TotalDocuments)
	}

	if stats.ProcessedDocuments != len(contents) {
		t.Errorf("Expected %d processed documents, got %d", len(contents), stats.ProcessedDocuments)
	}

	if stats.FailedDocuments != 0 {
		t.Errorf("Expected 0 failed documents, got %d", stats.FailedDocuments)
	}

	if stats.TotalChunks != len(contents)*2 { // Each document has 2 chunks
		t.Errorf("Expected %d total chunks, got %d", len(contents)*2, stats.TotalChunks)
	}

	if stats.Concurrency != maxConcurrency {
		t.Errorf("Expected concurrency %d, got %d", maxConcurrency, stats.Concurrency)
	}

	if stats.ProcessingTime <= 0 {
		t.Error("Expected positive processing time")
	}

	if stats.ThroughputDocs <= 0 {
		t.Error("Expected positive document throughput")
	}

	if stats.ThroughputChunks <= 0 {
		t.Error("Expected positive chunk throughput")
	}

	// Verify results
	if len(results) != len(contents) {
		t.Errorf("Expected %d results, got %d", len(contents), len(results))
	}

	if len(errors) != len(contents) {
		t.Errorf("Expected %d error entries, got %d", len(contents), len(errors))
	}

	t.Logf("Concurrent processing stats:")
	t.Logf("  Total documents: %d", stats.TotalDocuments)
	t.Logf("  Processed documents: %d", stats.ProcessedDocuments)
	t.Logf("  Total chunks: %d", stats.TotalChunks)
	t.Logf("  Processing time: %v", stats.ProcessingTime)
	t.Logf("  Average time per document: %v", stats.AverageTime)
	t.Logf("  Document throughput: %.2f docs/sec", stats.ThroughputDocs)
	t.Logf("  Chunk throughput: %.2f chunks/sec", stats.ThroughputChunks)
}

func TestChunkerPool_BasicFunctionality(t *testing.T) {
	config := DefaultConfig()
	pool := NewChunkerPool(config)

	// Get a chunker from the pool
	chunker1 := pool.Get()
	if chunker1 == nil {
		t.Fatal("Expected to get a chunker from pool")
	}

	// Use the chunker
	content := []byte("# Test\n\nContent")
	chunks, err := chunker1.ChunkDocument(content)
	if err != nil {
		t.Fatalf("ChunkDocument() error = %v", err)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}

	// Put it back in the pool
	pool.Put(chunker1)

	// Get another chunker (should be the same instance, reset)
	chunker2 := pool.Get()
	if chunker2 == nil {
		t.Fatal("Expected to get a chunker from pool")
	}

	// Verify chunker is reset
	if chunker2.HasErrors() {
		t.Error("Expected chunker to be reset (no errors)")
	}

	stats := chunker2.GetPerformanceStats()
	if stats.TotalChunks != 0 {
		t.Error("Expected chunker performance stats to be reset")
	}
}

func TestWorkerPool_BasicFunctionality(t *testing.T) {
	config := DefaultConfig()
	pool := NewWorkerPool(3, config)

	// Prepare test documents
	contents := make([][]byte, 10)
	for i := range contents {
		contents[i] = []byte(fmt.Sprintf("# Document %d\n\nContent %d", i, i))
	}

	results, errors := pool.ProcessBatch(contents)

	if len(results) != len(contents) {
		t.Errorf("Expected %d results, got %d", len(contents), len(results))
	}

	if len(errors) != len(contents) {
		t.Errorf("Expected %d error entries, got %d", len(contents), len(errors))
	}

	// Verify results
	for i, chunks := range results {
		if errors[i] != nil {
			t.Errorf("Document %d processing failed: %v", i, errors[i])
			continue
		}

		if len(chunks) != 2 {
			t.Errorf("Document %d: expected 2 chunks, got %d", i, len(chunks))
		}
	}
}

func TestWorkerPool_ManualJobSubmission(t *testing.T) {
	config := DefaultConfig()
	pool := NewWorkerPool(2, config)

	// Start the pool
	pool.Start()

	// Submit jobs manually
	numJobs := 5
	for i := 0; i < numJobs; i++ {
		content := []byte(fmt.Sprintf("# Job %d\n\nContent %d", i, i))
		pool.Submit(ProcessingJob{
			ID:      i,
			Content: content,
		})
	}

	// Collect results
	results := make(map[int]ProcessingResult)
	for i := 0; i < numJobs; i++ {
		result := pool.GetResult()
		results[result.ID] = result
	}

	// Stop the pool
	pool.Stop()

	// Verify results
	for i := 0; i < numJobs; i++ {
		result, exists := results[i]
		if !exists {
			t.Errorf("Missing result for job %d", i)
			continue
		}

		if result.Error != nil {
			t.Errorf("Job %d failed: %v", i, result.Error)
			continue
		}

		if len(result.Chunks) != 2 {
			t.Errorf("Job %d: expected 2 chunks, got %d", i, len(result.Chunks))
		}
	}
}

func TestConcurrentChunker_PerformanceComparison(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	config := DefaultConfig()

	// Prepare test documents
	numDocs := 100
	contents := make([][]byte, numDocs)
	for i := range contents {
		var doc strings.Builder
		doc.WriteString(fmt.Sprintf("# Document %d\n\n", i))
		for j := 0; j < 10; j++ {
			doc.WriteString(fmt.Sprintf("This is paragraph %d in document %d.\n\n", j, i))
		}
		doc.WriteString("```go\nfunc example() {\n    fmt.Println(\"Hello\")\n}\n```\n\n")
		doc.WriteString("| Column 1 | Column 2 |\n|----------|----------|\n| Value 1  | Value 2  |\n\n")
		contents[i] = []byte(doc.String())
	}

	// Test sequential processing
	sequentialChunker := NewMarkdownChunker()
	startTime := time.Now()

	for _, content := range contents {
		_, err := sequentialChunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("Sequential processing failed: %v", err)
		}
	}

	sequentialTime := time.Since(startTime)

	// Test concurrent processing
	concurrentChunker := NewConcurrentChunker(config)
	startTime = time.Now()

	maxConcurrency := runtime.NumCPU()
	stats, _, errors := concurrentChunker.ProcessDocumentsConcurrently(contents, maxConcurrency)

	concurrentTime := time.Since(startTime)

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Concurrent processing failed for document %d: %v", i, err)
		}
	}

	// Log performance comparison
	t.Logf("Performance comparison:")
	t.Logf("  Documents processed: %d", numDocs)
	t.Logf("  Sequential time: %v", sequentialTime)
	t.Logf("  Concurrent time: %v", concurrentTime)
	t.Logf("  Concurrency level: %d", maxConcurrency)
	t.Logf("  Speedup: %.2fx", float64(sequentialTime)/float64(concurrentTime))
	t.Logf("  Concurrent throughput: %.2f docs/sec", stats.ThroughputDocs)
	t.Logf("  Total chunks produced: %d", stats.TotalChunks)

	// Concurrent processing should be faster (or at least not significantly slower)
	// Allow some overhead for small datasets
	if concurrentTime > sequentialTime*2 {
		t.Logf("Warning: Concurrent processing is significantly slower than sequential")
		t.Logf("This might be expected for small datasets due to goroutine overhead")
	}
}

func TestConcurrentChunker_ErrorHandling(t *testing.T) {
	config := DefaultConfig()
	config.ErrorHandling = ErrorModeStrict
	chunker := NewConcurrentChunker(config)

	// Prepare test documents with some invalid content
	contents := [][]byte{
		[]byte("# Valid Document\n\nValid content"),
		nil, // This should cause an error
		[]byte("# Another Valid Document\n\nMore valid content"),
		[]byte(""), // Empty content
	}

	results, errors := chunker.ChunkDocumentConcurrent(contents)

	if len(results) != len(contents) {
		t.Errorf("Expected %d results, got %d", len(contents), len(results))
	}

	if len(errors) != len(contents) {
		t.Errorf("Expected %d error entries, got %d", len(contents), len(errors))
	}

	// Check specific results
	if errors[0] != nil {
		t.Errorf("Document 0 should not have error: %v", errors[0])
	}

	if errors[1] == nil {
		t.Error("Document 1 (nil content) should have error")
	}

	if errors[2] != nil {
		t.Errorf("Document 2 should not have error: %v", errors[2])
	}

	// Empty content handling depends on configuration
	// In strict mode, it might be treated as an error or produce empty results
}

func BenchmarkConcurrentChunker_Sequential(b *testing.B) {
	chunker := NewConcurrentChunker(DefaultConfig())
	content := []byte("# Benchmark Document\n\nThis is benchmark content with some text.\n\n```go\nfunc benchmark() {}\n```")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := chunker.ChunkDocument(content)
		if err != nil {
			b.Fatalf("ChunkDocument() error = %v", err)
		}
	}
}

func BenchmarkConcurrentChunker_Concurrent(b *testing.B) {
	chunker := NewConcurrentChunker(DefaultConfig())
	content := []byte("# Benchmark Document\n\nThis is benchmark content with some text.\n\n```go\nfunc benchmark() {}\n```")

	// Prepare multiple copies of the same content
	contents := make([][]byte, b.N)
	for i := range contents {
		contents[i] = content
	}

	b.ResetTimer()
	_, errors := chunker.ChunkDocumentConcurrent(contents)

	// Check for errors
	for _, err := range errors {
		if err != nil {
			b.Fatalf("Concurrent processing error: %v", err)
		}
	}
}

func BenchmarkWorkerPool_ProcessBatch(b *testing.B) {
	config := DefaultConfig()
	content := []byte("# Benchmark Document\n\nThis is benchmark content with some text.\n\n```go\nfunc benchmark() {}\n```")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Use a small batch size for benchmarking to avoid deadlock
		contents := [][]byte{content}

		pool := NewWorkerPool(runtime.NumCPU(), config)
		_, errors := pool.ProcessBatch(contents)

		// Check for errors
		for _, err := range errors {
			if err != nil {
				b.Fatalf("Worker pool processing error: %v", err)
			}
		}
	}
}
