package markdownchunker

import (
	"runtime"
	"strings"
	"sync"
)

// ObjectPool 对象池接口
type ObjectPool interface {
	Get() any
	Put(any)
	Reset()
}

// ChunkPool 块对象池
type ChunkPool struct {
	pool sync.Pool
}

// NewChunkPool 创建新的块对象池
func NewChunkPool() *ChunkPool {
	return &ChunkPool{
		pool: sync.Pool{
			New: func() any {
				return &Chunk{
					Metadata: make(map[string]string),
				}
			},
		},
	}
}

// Get 从池中获取一个块对象
func (cp *ChunkPool) Get() *Chunk {
	chunk := cp.pool.Get().(*Chunk)
	// 重置块对象
	chunk.ID = 0
	chunk.Type = ""
	chunk.Content = ""
	chunk.Text = ""
	chunk.Level = 0
	// 清空但保留map容量
	for k := range chunk.Metadata {
		delete(chunk.Metadata, k)
	}
	return chunk
}

// Put 将块对象放回池中
func (cp *ChunkPool) Put(chunk *Chunk) {
	if chunk != nil {
		cp.pool.Put(chunk)
	}
}

// Reset 重置池（清空所有对象）
func (cp *ChunkPool) Reset() {
	// 创建新的池来替换旧的
	cp.pool = sync.Pool{
		New: func() any {
			return &Chunk{
				Metadata: make(map[string]string),
			}
		},
	}
}

// StringBuilderPool 字符串构建器对象池
type StringBuilderPool struct {
	pool sync.Pool
}

// NewStringBuilderPool 创建新的字符串构建器对象池
func NewStringBuilderPool() *StringBuilderPool {
	return &StringBuilderPool{
		pool: sync.Pool{
			New: func() any {
				return &strings.Builder{}
			},
		},
	}
}

// Get 从池中获取一个字符串构建器
func (sbp *StringBuilderPool) Get() *strings.Builder {
	sb := sbp.pool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// Put 将字符串构建器放回池中
func (sbp *StringBuilderPool) Put(sb *strings.Builder) {
	if sb != nil {
		// 如果构建器太大，不放回池中以避免内存泄漏
		if sb.Cap() > 64*1024 { // 64KB
			return
		}
		sbp.pool.Put(sb)
	}
}

// MemoryLimiter 内存限制器
type MemoryLimiter struct {
	maxMemoryBytes int64
	mu             sync.RWMutex
}

// NewMemoryLimiter 创建新的内存限制器
func NewMemoryLimiter(maxMemoryBytes int64) *MemoryLimiter {
	return &MemoryLimiter{
		maxMemoryBytes: maxMemoryBytes,
	}
}

// CheckMemoryLimit 检查内存使用是否超过限制
func (ml *MemoryLimiter) CheckMemoryLimit() error {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	if ml.maxMemoryBytes <= 0 {
		return nil // 无限制
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if int64(m.Alloc) > ml.maxMemoryBytes {
		return NewChunkerError(ErrorTypeMemoryExhausted,
			"memory usage exceeds limit", nil).
			WithContext("current_memory", int64(m.Alloc)).
			WithContext("memory_limit", ml.maxMemoryBytes)
	}

	return nil
}

// SetMemoryLimit 设置内存限制
func (ml *MemoryLimiter) SetMemoryLimit(maxMemoryBytes int64) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.maxMemoryBytes = maxMemoryBytes
}

// GetMemoryLimit 获取内存限制
func (ml *MemoryLimiter) GetMemoryLimit() int64 {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return ml.maxMemoryBytes
}

// GetCurrentMemoryUsage 获取当前内存使用量
func (ml *MemoryLimiter) GetCurrentMemoryUsage() int64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int64(m.Alloc)
}

// MemoryOptimizer 内存优化器
type MemoryOptimizer struct {
	chunkPool         *ChunkPool
	stringBuilderPool *StringBuilderPool
	memoryLimiter     *MemoryLimiter
	gcThreshold       int64 // GC触发阈值
	processedBytes    int64 // 已处理字节数
	mu                sync.RWMutex
}

// NewMemoryOptimizer 创建新的内存优化器
func NewMemoryOptimizer(memoryLimit int64) *MemoryOptimizer {
	return &MemoryOptimizer{
		chunkPool:         NewChunkPool(),
		stringBuilderPool: NewStringBuilderPool(),
		memoryLimiter:     NewMemoryLimiter(memoryLimit),
		gcThreshold:       10 * 1024 * 1024, // 10MB
		processedBytes:    0,
	}
}

// GetChunk 获取一个优化的块对象
func (mo *MemoryOptimizer) GetChunk() *Chunk {
	return mo.chunkPool.Get()
}

// PutChunk 归还块对象到池中
func (mo *MemoryOptimizer) PutChunk(chunk *Chunk) {
	mo.chunkPool.Put(chunk)
}

// GetStringBuilder 获取一个字符串构建器
func (mo *MemoryOptimizer) GetStringBuilder() *strings.Builder {
	return mo.stringBuilderPool.Get()
}

// PutStringBuilder 归还字符串构建器到池中
func (mo *MemoryOptimizer) PutStringBuilder(sb *strings.Builder) {
	mo.stringBuilderPool.Put(sb)
}

// CheckMemoryLimit 检查内存限制
func (mo *MemoryOptimizer) CheckMemoryLimit() error {
	return mo.memoryLimiter.CheckMemoryLimit()
}

// RecordProcessedBytes 记录已处理的字节数
func (mo *MemoryOptimizer) RecordProcessedBytes(bytes int64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.processedBytes += bytes

	// 如果处理的字节数超过阈值，触发GC
	if mo.processedBytes >= mo.gcThreshold {
		runtime.GC()
		mo.processedBytes = 0
	}
}

// SetGCThreshold 设置GC触发阈值
func (mo *MemoryOptimizer) SetGCThreshold(threshold int64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.gcThreshold = threshold
}

// GetGCThreshold 获取GC触发阈值
func (mo *MemoryOptimizer) GetGCThreshold() int64 {
	mo.mu.RLock()
	defer mo.mu.RUnlock()
	return mo.gcThreshold
}

// ForceGC 强制执行垃圾回收
func (mo *MemoryOptimizer) ForceGC() {
	runtime.GC()
	mo.mu.Lock()
	mo.processedBytes = 0
	mo.mu.Unlock()
}

// Reset 重置优化器状态
func (mo *MemoryOptimizer) Reset() {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.chunkPool.Reset()
	mo.processedBytes = 0
}

// GetMemoryStats 获取内存统计信息
func (mo *MemoryOptimizer) GetMemoryStats() MemoryOptimizerStats {
	mo.mu.RLock()
	defer mo.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryOptimizerStats{
		CurrentMemory:    int64(m.Alloc),
		MemoryLimit:      mo.memoryLimiter.GetMemoryLimit(),
		ProcessedBytes:   mo.processedBytes,
		GCThreshold:      mo.gcThreshold,
		TotalAllocations: int64(m.TotalAlloc),
		GCCycles:         int64(m.NumGC),
	}
}

// MemoryOptimizerStats 内存优化器统计信息
type MemoryOptimizerStats struct {
	CurrentMemory    int64 `json:"current_memory"`    // 当前内存使用
	MemoryLimit      int64 `json:"memory_limit"`      // 内存限制
	ProcessedBytes   int64 `json:"processed_bytes"`   // 已处理字节数
	GCThreshold      int64 `json:"gc_threshold"`      // GC阈值
	TotalAllocations int64 `json:"total_allocations"` // 总分配内存
	GCCycles         int64 `json:"gc_cycles"`         // GC周期数
}

// OptimizedStringOperations 优化的字符串操作
type OptimizedStringOperations struct {
	builderPool *StringBuilderPool
}

// NewOptimizedStringOperations 创建优化的字符串操作实例
func NewOptimizedStringOperations() *OptimizedStringOperations {
	return &OptimizedStringOperations{
		builderPool: NewStringBuilderPool(),
	}
}

// JoinStrings 优化的字符串连接
func (oso *OptimizedStringOperations) JoinStrings(strs []string, separator string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	sb := oso.builderPool.Get()
	defer oso.builderPool.Put(sb)

	for i, str := range strs {
		if i > 0 {
			sb.WriteString(separator)
		}
		sb.WriteString(str)
	}

	return sb.String()
}

// TrimAndClean 优化的字符串清理
func (oso *OptimizedStringOperations) TrimAndClean(text string) string {
	if text == "" {
		return ""
	}

	// 使用更高效的方式处理空白字符
	text = strings.TrimSpace(text)

	// 如果字符串很短，直接使用标准库
	if len(text) < 100 {
		return strings.Join(strings.Fields(text), " ")
	}

	// 对于长字符串，使用优化的方法
	sb := oso.builderPool.Get()
	defer oso.builderPool.Put(sb)

	fields := strings.Fields(text)
	for i, field := range fields {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(field)
	}

	return sb.String()
}

// BuildContent 优化的内容构建
func (oso *OptimizedStringOperations) BuildContent(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	sb := oso.builderPool.Get()
	defer oso.builderPool.Put(sb)

	for _, part := range parts {
		if part != "" {
			sb.WriteString(part)
		}
	}

	return sb.String()
}
