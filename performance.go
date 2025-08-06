package markdownchunker

import (
	"runtime"
	"sync"
	"time"
)

// PerformanceStats 性能统计信息
type PerformanceStats struct {
	ProcessingTime  time.Duration `json:"processing_time"`   // 处理时间
	MemoryUsed      int64         `json:"memory_used"`       // 使用的内存（字节）
	ChunksPerSecond float64       `json:"chunks_per_second"` // 每秒处理的块数
	BytesPerSecond  float64       `json:"bytes_per_second"`  // 每秒处理的字节数
	TotalChunks     int           `json:"total_chunks"`      // 总块数
	TotalBytes      int64         `json:"total_bytes"`       // 总字节数（输入文档大小）
	ChunkBytes      int64         `json:"chunk_bytes"`       // 块内容总字节数
	PeakMemory      int64         `json:"peak_memory"`       // 峰值内存使用
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	mu            sync.RWMutex
	startTime     time.Time
	endTime       time.Time
	initialMemory int64
	peakMemory    int64
	totalBytes    int64 // 输入文档总字节数
	chunkBytes    int64 // 块内容总字节数
	chunkCount    int
	isRunning     bool
}

// NewPerformanceMonitor 创建新的性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		peakMemory: 0,
		totalBytes: 0,
		chunkBytes: 0,
		chunkCount: 0,
		isRunning:  false,
	}
}

// Start 开始性能监控
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.startTime = time.Now()
	pm.endTime = time.Time{}
	pm.initialMemory = pm.getCurrentMemoryUsage()
	pm.peakMemory = pm.initialMemory
	pm.totalBytes = 0
	pm.chunkBytes = 0
	pm.chunkCount = 0
	pm.isRunning = true
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning {
		pm.endTime = time.Now()
		pm.isRunning = false
	}
}

// RecordChunk 记录处理的块信息
func (pm *PerformanceMonitor) RecordChunk(chunk *Chunk) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return
	}

	pm.chunkCount++
	pm.chunkBytes += int64(len(chunk.Content))

	// 更新峰值内存使用
	currentMemory := pm.getCurrentMemoryUsage()
	if currentMemory > pm.peakMemory {
		pm.peakMemory = currentMemory
	}
}

// RecordBytes 记录处理的字节数（用于输入文档大小）
func (pm *PerformanceMonitor) RecordBytes(bytes int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return
	}

	pm.totalBytes += bytes
}

// GetStats 获取性能统计信息
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var processingTime time.Duration
	if pm.isRunning {
		processingTime = time.Since(pm.startTime)
	} else if !pm.endTime.IsZero() {
		processingTime = pm.endTime.Sub(pm.startTime)
	}

	var chunksPerSecond, bytesPerSecond float64
	if processingTime > 0 {
		seconds := processingTime.Seconds()
		chunksPerSecond = float64(pm.chunkCount) / seconds
		bytesPerSecond = float64(pm.totalBytes) / seconds
	}

	memoryUsed := max(pm.peakMemory-pm.initialMemory, 0)

	return PerformanceStats{
		ProcessingTime:  processingTime,
		MemoryUsed:      memoryUsed,
		ChunksPerSecond: chunksPerSecond,
		BytesPerSecond:  bytesPerSecond,
		TotalChunks:     pm.chunkCount,
		TotalBytes:      pm.totalBytes,
		ChunkBytes:      pm.chunkBytes,
		PeakMemory:      pm.peakMemory,
	}
}

// IsRunning 检查监控器是否正在运行
func (pm *PerformanceMonitor) IsRunning() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.isRunning
}

// Reset 重置监控器状态
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.startTime = time.Time{}
	pm.endTime = time.Time{}
	pm.initialMemory = 0
	pm.peakMemory = 0
	pm.totalBytes = 0
	pm.chunkBytes = 0
	pm.chunkCount = 0
	pm.isRunning = false
}

// getCurrentMemoryUsage 获取当前内存使用量
func (pm *PerformanceMonitor) getCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// ForceGC 强制垃圾回收（用于测试和内存优化）
func (pm *PerformanceMonitor) ForceGC() {
	runtime.GC()
}

// GetMemoryStats 获取详细的内存统计信息
func (pm *PerformanceMonitor) GetMemoryStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}
