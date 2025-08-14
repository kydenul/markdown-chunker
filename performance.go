package markdownchunker

import (
	"runtime"
	"sync"
	"time"

	"github.com/kydenul/log"
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
	logger        log.Logger // 日志器实例
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

// SetLogger 设置日志器
func (pm *PerformanceMonitor) SetLogger(logger log.Logger) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.logger = logger
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

	if pm.logger != nil {
		pm.logger.Infow("性能监控开始",
			"start_time", pm.startTime.Format("2006-01-02 15:04:05.000"),
			"initial_memory_bytes", pm.initialMemory,
			"initial_memory_mb", pm.initialMemory/(1024*1024),
			"function", "PerformanceMonitor.Start")
	}
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning {
		pm.endTime = time.Now()
		pm.isRunning = false

		if pm.logger != nil {
			processingTime := pm.endTime.Sub(pm.startTime)
			currentMemory := pm.getCurrentMemoryUsage()
			memoryUsed := currentMemory - pm.initialMemory

			pm.logger.Infow("性能监控结束",
				"end_time", pm.endTime.Format("2006-01-02 15:04:05.000"),
				"processing_time_ms", processingTime.Milliseconds(),
				"processing_time_seconds", processingTime.Seconds(),
				"total_chunks", pm.chunkCount,
				"total_bytes", pm.totalBytes,
				"chunk_bytes", pm.chunkBytes,
				"memory_used_bytes", memoryUsed,
				"memory_used_mb", memoryUsed/(1024*1024),
				"peak_memory_bytes", pm.peakMemory,
				"peak_memory_mb", pm.peakMemory/(1024*1024),
				"function", "PerformanceMonitor.Stop")

			// 计算性能指标
			if processingTime > 0 {
				chunksPerSecond := float64(pm.chunkCount) / processingTime.Seconds()
				bytesPerSecond := float64(pm.totalBytes) / processingTime.Seconds()

				pm.logger.Infow("性能指标统计",
					"chunks_per_second", chunksPerSecond,
					"bytes_per_second", bytesPerSecond,
					"mb_per_second", bytesPerSecond/(1024*1024),
					"function", "PerformanceMonitor.Stop")
			}
		}
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
	previousPeak := pm.peakMemory
	if currentMemory > pm.peakMemory {
		pm.peakMemory = currentMemory

		// 记录新的内存峰值
		if pm.logger != nil {
			pm.logger.Debugw("检测到新的内存使用峰值",
				"previous_peak_bytes", previousPeak,
				"previous_peak_mb", previousPeak/(1024*1024),
				"new_peak_bytes", pm.peakMemory,
				"new_peak_mb", pm.peakMemory/(1024*1024),
				"memory_increase_bytes", pm.peakMemory-previousPeak,
				"memory_increase_mb", (pm.peakMemory-previousPeak)/(1024*1024),
				"chunk_id", chunk.ID,
				"chunk_type", chunk.Type,
				"chunk_size_bytes", len(chunk.Content),
				"total_chunks_processed", pm.chunkCount,
				"function", "PerformanceMonitor.RecordChunk")
		}
	}

	// 每处理100个块记录一次进度
	if pm.chunkCount%100 == 0 && pm.logger != nil {
		elapsedTime := time.Since(pm.startTime)
		chunksPerSecond := float64(pm.chunkCount) / elapsedTime.Seconds()

		pm.logger.Infow("块处理进度报告",
			"chunks_processed", pm.chunkCount,
			"elapsed_time_seconds", elapsedTime.Seconds(),
			"chunks_per_second", chunksPerSecond,
			"current_memory_bytes", currentMemory,
			"current_memory_mb", currentMemory/(1024*1024),
			"peak_memory_bytes", pm.peakMemory,
			"peak_memory_mb", pm.peakMemory/(1024*1024),
			"function", "PerformanceMonitor.RecordChunk")
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

	if pm.logger != nil {
		pm.logger.Debugw("记录处理字节数",
			"bytes_added", bytes,
			"bytes_added_mb", bytes/(1024*1024),
			"total_bytes", pm.totalBytes,
			"total_mb", pm.totalBytes/(1024*1024),
			"function", "PerformanceMonitor.RecordBytes")

		// 对大型文档记录警告
		if pm.totalBytes > 50*1024*1024 { // 50MB
			pm.logger.Warnw("处理大型文档",
				"total_bytes", pm.totalBytes,
				"total_mb", pm.totalBytes/(1024*1024),
				"recommendation", "考虑分批处理以优化内存使用",
				"function", "PerformanceMonitor.RecordBytes")
		}
	}
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

// CheckMemoryThresholds 检查内存使用阈值并记录警告
func (pm *PerformanceMonitor) CheckMemoryThresholds() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if pm.logger == nil {
		return
	}

	currentMemory := pm.getCurrentMemoryUsage()
	memoryIncrease := currentMemory - pm.initialMemory

	// 内存使用超过100MB时记录警告
	if memoryIncrease > 100*1024*1024 {
		pm.logger.Warnw("内存使用量较高",
			"current_memory_bytes", currentMemory,
			"current_memory_mb", currentMemory/(1024*1024),
			"memory_increase_bytes", memoryIncrease,
			"memory_increase_mb", memoryIncrease/(1024*1024),
			"initial_memory_mb", pm.initialMemory/(1024*1024),
			"recommendation", "考虑启用内存优化或增加内存限制",
			"function", "PerformanceMonitor.CheckMemoryThresholds")
	}

	// 内存使用超过500MB时记录错误
	if memoryIncrease > 500*1024*1024 {
		pm.logger.Errorw("内存使用量过高",
			"current_memory_bytes", currentMemory,
			"current_memory_mb", currentMemory/(1024*1024),
			"memory_increase_bytes", memoryIncrease,
			"memory_increase_mb", memoryIncrease/(1024*1024),
			"initial_memory_mb", pm.initialMemory/(1024*1024),
			"recommendation", "立即考虑优化内存使用或终止处理",
			"function", "PerformanceMonitor.CheckMemoryThresholds")
	}
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

// RecordStrategyExecution 记录策略执行信息
func (pm *PerformanceMonitor) RecordStrategyExecution(strategyName string, executionTime time.Duration, chunksGenerated int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning {
		return
	}

	if pm.logger != nil {
		pm.logger.Debugw("策略执行统计",
			"strategy_name", strategyName,
			"execution_time_ms", executionTime.Milliseconds(),
			"execution_time_seconds", executionTime.Seconds(),
			"chunks_generated", chunksGenerated,
			"chunks_per_second", func() float64 {
				if executionTime.Seconds() > 0 {
					return float64(chunksGenerated) / executionTime.Seconds()
				}
				return 0
			}(),
			"function", "PerformanceMonitor.RecordStrategyExecution")

		// 记录性能警告
		if executionTime > 5*time.Second {
			pm.logger.Warnw("策略执行时间较长",
				"strategy_name", strategyName,
				"execution_time_seconds", executionTime.Seconds(),
				"chunks_generated", chunksGenerated,
				"recommendation", "考虑优化策略实现或使用更高效的策略",
				"function", "PerformanceMonitor.RecordStrategyExecution")
		}

		// 记录低效率警告
		if executionTime.Seconds() > 0 {
			chunksPerSecond := float64(chunksGenerated) / executionTime.Seconds()
			if chunksPerSecond < 10 && chunksGenerated > 0 {
				pm.logger.Warnw("策略执行效率较低",
					"strategy_name", strategyName,
					"chunks_per_second", chunksPerSecond,
					"chunks_generated", chunksGenerated,
					"execution_time_seconds", executionTime.Seconds(),
					"recommendation", "考虑优化策略算法或检查文档复杂度",
					"function", "PerformanceMonitor.RecordStrategyExecution")
			}
		}
	}
}
