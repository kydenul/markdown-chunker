package markdownchunker

import (
	"sync"
	"time"
)

// ConcurrentChunker 并发安全的分块器包装器
type ConcurrentChunker struct {
	chunker *MarkdownChunker
	mu      sync.RWMutex
	pool    *ChunkerPool
}

// ChunkerPool 分块器对象池，用于并发处理
type ChunkerPool struct {
	pool   sync.Pool
	config *ChunkerConfig
}

// NewChunkerPool 创建新的分块器对象池
func NewChunkerPool(config *ChunkerConfig) *ChunkerPool {
	if config == nil {
		config = DefaultConfig()
	}

	return &ChunkerPool{
		config: config,
		pool: sync.Pool{
			New: func() interface{} {
				return NewMarkdownChunkerWithConfig(config)
			},
		},
	}
}

// Get 从池中获取一个分块器实例
func (cp *ChunkerPool) Get() *MarkdownChunker {
	chunker := cp.pool.Get().(*MarkdownChunker)
	// 重置分块器状态
	chunker.ClearErrors()
	chunker.ResetPerformanceMonitor()
	return chunker
}

// Put 将分块器实例放回池中
func (cp *ChunkerPool) Put(chunker *MarkdownChunker) {
	if chunker != nil {
		cp.pool.Put(chunker)
	}
}

// NewConcurrentChunker 创建新的并发安全分块器
func NewConcurrentChunker(config *ChunkerConfig) *ConcurrentChunker {
	if config == nil {
		config = DefaultConfig()
	}

	return &ConcurrentChunker{
		chunker: NewMarkdownChunkerWithConfig(config),
		pool:    NewChunkerPool(config),
	}
}

// ChunkDocument 线程安全的文档分块方法
func (cc *ConcurrentChunker) ChunkDocument(content []byte) ([]Chunk, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	return cc.chunker.ChunkDocument(content)
}

// ChunkDocumentConcurrent 并发处理多个文档
func (cc *ConcurrentChunker) ChunkDocumentConcurrent(contents [][]byte) ([][]Chunk, []error) {
	if len(contents) == 0 {
		return nil, nil
	}

	results := make([][]Chunk, len(contents))
	errors := make([]error, len(contents))

	var wg sync.WaitGroup

	for i, content := range contents {
		wg.Add(1)
		go func(index int, data []byte) {
			defer wg.Done()

			// 从池中获取分块器实例
			chunker := cc.pool.Get()
			defer cc.pool.Put(chunker)

			// 处理文档
			chunks, err := chunker.ChunkDocument(data)
			results[index] = chunks
			errors[index] = err
		}(i, content)
	}

	wg.Wait()
	return results, errors
}

// ChunkDocumentBatch 批量处理文档，支持并发控制
func (cc *ConcurrentChunker) ChunkDocumentBatch(contents [][]byte, maxConcurrency int) ([][]Chunk, []error) {
	if len(contents) == 0 {
		return nil, nil
	}

	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	results := make([][]Chunk, len(contents))
	errors := make([]error, len(contents))

	// 创建工作通道
	jobs := make(chan int, len(contents))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for index := range jobs {
				// 从池中获取分块器实例
				chunker := cc.pool.Get()

				// 处理文档
				chunks, err := chunker.ChunkDocument(contents[index])
				results[index] = chunks
				errors[index] = err

				// 归还分块器实例
				cc.pool.Put(chunker)
			}
		}()
	}

	// 发送任务
	for i := range contents {
		jobs <- i
	}
	close(jobs)

	// 等待所有任务完成
	wg.Wait()

	return results, errors
}

// GetPerformanceStats 获取性能统计信息（线程安全）
func (cc *ConcurrentChunker) GetPerformanceStats() PerformanceStats {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.chunker.GetPerformanceStats()
}

// GetErrors 获取错误信息（线程安全）
func (cc *ConcurrentChunker) GetErrors() []*ChunkerError {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.chunker.GetErrors()
}

// ClearErrors 清除错误信息（线程安全）
func (cc *ConcurrentChunker) ClearErrors() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.chunker.ClearErrors()
}

// ConcurrentProcessingStats 并发处理统计信息
type ConcurrentProcessingStats struct {
	TotalDocuments     int           `json:"total_documents"`     // 总文档数
	ProcessedDocuments int           `json:"processed_documents"` // 已处理文档数
	FailedDocuments    int           `json:"failed_documents"`    // 失败文档数
	TotalChunks        int           `json:"total_chunks"`        // 总块数
	ProcessingTime     time.Duration `json:"processing_time"`     // 总处理时间
	AverageTime        time.Duration `json:"average_time"`        // 平均处理时间
	Concurrency        int           `json:"concurrency"`         // 并发度
	ThroughputDocs     float64       `json:"throughput_docs"`     // 文档吞吐量（文档/秒）
	ThroughputChunks   float64       `json:"throughput_chunks"`   // 块吞吐量（块/秒）
}

// ProcessDocumentsConcurrently 并发处理文档并收集统计信息
func (cc *ConcurrentChunker) ProcessDocumentsConcurrently(contents [][]byte, maxConcurrency int) (*ConcurrentProcessingStats, [][]Chunk, []error) {
	startTime := time.Now()

	results, errors := cc.ChunkDocumentBatch(contents, maxConcurrency)

	processingTime := time.Since(startTime)

	// 计算统计信息
	stats := &ConcurrentProcessingStats{
		TotalDocuments: len(contents),
		ProcessingTime: processingTime,
		Concurrency:    maxConcurrency,
	}

	processedDocs := 0
	failedDocs := 0
	totalChunks := 0

	for i, err := range errors {
		if err != nil {
			failedDocs++
		} else {
			processedDocs++
			if results[i] != nil {
				totalChunks += len(results[i])
			}
		}
	}

	stats.ProcessedDocuments = processedDocs
	stats.FailedDocuments = failedDocs
	stats.TotalChunks = totalChunks

	if processedDocs > 0 {
		stats.AverageTime = processingTime / time.Duration(processedDocs)
	}

	if processingTime > 0 {
		seconds := processingTime.Seconds()
		stats.ThroughputDocs = float64(processedDocs) / seconds
		stats.ThroughputChunks = float64(totalChunks) / seconds
	}

	return stats, results, errors
}

// WorkerPool 工作池，用于更精细的并发控制
type WorkerPool struct {
	workers     int
	jobs        chan ProcessingJob
	results     chan ProcessingResult
	chunkerPool *ChunkerPool
	wg          sync.WaitGroup
}

// ProcessingJob 处理任务
type ProcessingJob struct {
	ID      int
	Content []byte
}

// ProcessingResult 处理结果
type ProcessingResult struct {
	ID     int
	Chunks []Chunk
	Error  error
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(workers int, config *ChunkerConfig) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	return &WorkerPool{
		workers:     workers,
		jobs:        make(chan ProcessingJob, workers*2),     // 缓冲队列
		results:     make(chan ProcessingResult, workers*10), // 更大的缓冲区
		chunkerPool: NewChunkerPool(config),
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// worker 工作协程
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for job := range wp.jobs {
		// 从池中获取分块器
		chunker := wp.chunkerPool.Get()

		// 处理任务
		chunks, err := chunker.ChunkDocument(job.Content)

		// 发送结果
		wp.results <- ProcessingResult{
			ID:     job.ID,
			Chunks: chunks,
			Error:  err,
		}

		// 归还分块器
		wp.chunkerPool.Put(chunker)
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(job ProcessingJob) {
	wp.jobs <- job
}

// GetResult 获取结果
func (wp *WorkerPool) GetResult() ProcessingResult {
	return <-wp.results
}

// ProcessBatch 批量处理任务
func (wp *WorkerPool) ProcessBatch(contents [][]byte) ([][]Chunk, []error) {
	if len(contents) == 0 {
		return nil, nil
	}

	// 启动工作池
	wp.Start()
	defer wp.Stop()

	// 提交所有任务
	for i, content := range contents {
		wp.Submit(ProcessingJob{
			ID:      i,
			Content: content,
		})
	}

	// 收集结果
	results := make([][]Chunk, len(contents))
	errors := make([]error, len(contents))

	for i := 0; i < len(contents); i++ {
		result := wp.GetResult()
		results[result.ID] = result.Chunks
		errors[result.ID] = result.Error
	}

	return results, errors
}
