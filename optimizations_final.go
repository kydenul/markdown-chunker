package markdownchunker

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kydenul/log"
)

// OptimizedLogContextManager provides efficient log context management
type OptimizedLogContextManager struct {
	fieldPool   sync.Pool
	contextPool sync.Pool
	logger      log.Logger
}

// NewOptimizedLogContextManager creates a new log context manager
func NewOptimizedLogContextManager(logger log.Logger) *OptimizedLogContextManager {
	return &OptimizedLogContextManager{
		logger: logger,
		fieldPool: sync.Pool{
			New: func() interface{} {
				return make([]interface{}, 0, 20)
			},
		},
		contextPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, 10)
			},
		},
	}
}

// CreateContext creates an optimized log context
func (lm *OptimizedLogContextManager) CreateContext(functionName string) *OptimizedLogContextFinal {
	fields := lm.contextPool.Get().(map[string]interface{})

	// Clear existing fields efficiently
	for k := range fields {
		delete(fields, k)
	}

	fields["function"] = functionName
	fields["timestamp"] = time.Now().Unix()

	return &OptimizedLogContextFinal{
		fields:  fields,
		manager: lm,
	}
}

// OptimizedLogContextFinal provides memory-efficient log context
type OptimizedLogContextFinal struct {
	fields  map[string]interface{}
	manager *OptimizedLogContextManager
}

// WithField adds a field efficiently
func (ctx *OptimizedLogContextFinal) WithField(key string, value interface{}) *OptimizedLogContextFinal {
	ctx.fields[key] = value
	return ctx
}

// WithError adds error information efficiently
func (ctx *OptimizedLogContextFinal) WithError(err error) *OptimizedLogContextFinal {
	if err != nil {
		ctx.fields["error"] = err.Error()
		if chunkerErr, ok := err.(*ChunkerError); ok {
			ctx.fields["error_type"] = chunkerErr.Type.String()
			for k, v := range chunkerErr.Context {
				ctx.fields["error_"+k] = v
			}
		}
	}
	return ctx
}

// ToLogFields converts to log fields efficiently
func (ctx *OptimizedLogContextFinal) ToLogFields() []interface{} {
	fields := ctx.manager.fieldPool.Get().([]interface{})
	fields = fields[:0] // Reset length but keep capacity

	for k, v := range ctx.fields {
		fields = append(fields, k, v)
	}

	return fields
}

// Release returns the context to the pool
func (ctx *OptimizedLogContextFinal) Release() {
	if ctx.manager != nil {
		ctx.manager.contextPool.Put(ctx.fields)
	}
}

// ReleaseFields returns fields to the pool
func (ctx *OptimizedLogContextFinal) ReleaseFields(fields []interface{}) {
	if ctx.manager != nil {
		ctx.manager.fieldPool.Put(fields)
	}
}

// OptimizedErrorHandlerFinal provides memory-efficient error handling
type OptimizedErrorHandlerFinal struct {
	errors     []ChunkerError
	mode       ErrorHandlingMode
	mutex      sync.RWMutex
	logger     log.Logger
	logManager *OptimizedLogContextManager
	maxErrors  int
}

// NewOptimizedErrorHandlerFinal creates an optimized error handler
func NewOptimizedErrorHandlerFinal(mode ErrorHandlingMode, logger log.Logger, maxErrors int) *OptimizedErrorHandlerFinal {
	return &OptimizedErrorHandlerFinal{
		errors:     make([]ChunkerError, 0, 16),
		mode:       mode,
		logger:     logger,
		logManager: NewOptimizedLogContextManager(logger),
		maxErrors:  maxErrors,
	}
}

// HandleError handles errors with optimized performance
func (h *OptimizedErrorHandlerFinal) HandleError(err *ChunkerError) error {
	if err == nil {
		return nil
	}

	// Fast path for silent mode
	if h.mode == ErrorModeSilent {
		return nil
	}

	// Use optimized logging
	h.logErrorOptimized(err)

	// Handle error storage with memory limits
	h.mutex.Lock()
	if len(h.errors) >= h.maxErrors {
		// Remove oldest errors to prevent memory bloat
		copy(h.errors, h.errors[1:])
		h.errors = h.errors[:len(h.errors)-1]
	}
	h.errors = append(h.errors, *err)
	h.mutex.Unlock()

	// Return based on mode
	switch h.mode {
	case ErrorModeStrict:
		return err
	case ErrorModePermissive:
		return nil
	default:
		return err
	}
}

// logErrorOptimized provides optimized error logging
func (h *OptimizedErrorHandlerFinal) logErrorOptimized(err *ChunkerError) {
	if h.logger == nil {
		return
	}

	ctx := h.logManager.CreateContext("HandleError")
	defer ctx.Release()

	ctx.WithField("error_type", err.Type.String()).
		WithField("error_message", err.Message).
		WithField("error_timestamp", err.Timestamp.Format(time.RFC3339))

	// Add context fields efficiently
	for k, v := range err.Context {
		ctx.WithField(k, v)
	}

	if err.Cause != nil {
		ctx.WithField("cause_error", err.Cause.Error())
	}

	fields := ctx.ToLogFields()
	defer ctx.ReleaseFields(fields)

	// Use appropriate log level
	switch err.Type {
	case ErrorTypeMemoryExhausted, ErrorTypeTimeout:
		h.logger.Errorw("Critical error occurred", fields...)
	case ErrorTypeStrategyExecutionFailed, ErrorTypeParsingFailed:
		h.logger.Errorw("Execution error occurred", fields...)
	case ErrorTypeInvalidInput, ErrorTypeConfigInvalid:
		h.logger.Warnw("Configuration error occurred", fields...)
	default:
		h.logger.Infow("General error occurred", fields...)
	}
}

// GetErrorSummary returns a summary of errors
func (h *OptimizedErrorHandlerFinal) GetErrorSummary() map[ErrorType]int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	summary := make(map[ErrorType]int)
	for _, err := range h.errors {
		summary[err.Type]++
	}
	return summary
}

// OptimizedPerformanceMonitorFinal provides efficient performance monitoring
type OptimizedPerformanceMonitorFinal struct {
	mu            sync.RWMutex
	startTime     time.Time
	endTime       time.Time
	initialMemory uint64
	peakMemory    uint64
	totalBytes    int64
	chunkBytes    int64
	chunkCount    int64
	isRunning     bool
	logger        log.Logger
	logManager    *OptimizedLogContextManager
	lastLogTime   time.Time
	logInterval   time.Duration
	memoryChecks  int64
}

// NewOptimizedPerformanceMonitorFinal creates an optimized performance monitor
func NewOptimizedPerformanceMonitorFinal(logger log.Logger) *OptimizedPerformanceMonitorFinal {
	return &OptimizedPerformanceMonitorFinal{
		logger:      logger,
		logManager:  NewOptimizedLogContextManager(logger),
		logInterval: 10 * time.Second,
	}
}

// Start begins performance monitoring
func (pm *OptimizedPerformanceMonitorFinal) Start() {
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
	pm.lastLogTime = pm.startTime
	pm.memoryChecks = 0

	if pm.logger != nil {
		ctx := pm.logManager.CreateContext("Start")
		defer ctx.Release()

		ctx.WithField("start_time", pm.startTime.Format(time.RFC3339)).
			WithField("initial_memory_mb", pm.initialMemory/(1024*1024))

		fields := ctx.ToLogFields()
		defer ctx.ReleaseFields(fields)

		pm.logger.Infow("Performance monitoring started", fields...)
	}
}

// RecordChunk records chunk processing
func (pm *OptimizedPerformanceMonitorFinal) RecordChunk(chunk *Chunk) {
	pm.mu.Lock()
	if !pm.isRunning {
		pm.mu.Unlock()
		return
	}

	pm.chunkCount++
	pm.chunkBytes += int64(len(chunk.Content))

	// Check memory periodically
	shouldCheckMemory := pm.memoryChecks%100 == 0
	pm.memoryChecks++
	pm.mu.Unlock()

	if shouldCheckMemory {
		currentMemory := pm.getCurrentMemoryUsage()

		pm.mu.Lock()
		if currentMemory > pm.peakMemory {
			pm.peakMemory = currentMemory
		}
		pm.mu.Unlock()

		// Log progress periodically
		now := time.Now()
		pm.mu.RLock()
		shouldLog := now.Sub(pm.lastLogTime) >= pm.logInterval
		chunkCount := pm.chunkCount
		pm.mu.RUnlock()

		if shouldLog && pm.logger != nil {
			pm.mu.Lock()
			pm.lastLogTime = now
			pm.mu.Unlock()

			ctx := pm.logManager.CreateContext("RecordChunk")
			defer ctx.Release()

			ctx.WithField("chunks_processed", chunkCount).
				WithField("current_memory_mb", currentMemory/(1024*1024)).
				WithField("peak_memory_mb", pm.peakMemory/(1024*1024)).
				WithField("elapsed_seconds", now.Sub(pm.startTime).Seconds())

			fields := ctx.ToLogFields()
			defer ctx.ReleaseFields(fields)

			pm.logger.Debugw("Chunk processing progress", fields...)
		}
	}
}

// getCurrentMemoryUsage gets current memory usage
func (pm *OptimizedPerformanceMonitorFinal) getCurrentMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// Stop ends monitoring
func (pm *OptimizedPerformanceMonitorFinal) Stop() {
	pm.mu.Lock()
	if !pm.isRunning {
		pm.mu.Unlock()
		return
	}

	pm.endTime = time.Now()
	pm.isRunning = false

	// Capture final state
	processingTime := pm.endTime.Sub(pm.startTime)
	currentMemory := pm.getCurrentMemoryUsage()
	memoryUsed := int64(currentMemory) - int64(pm.initialMemory)
	chunkCount := pm.chunkCount
	totalBytes := pm.totalBytes
	chunkBytes := pm.chunkBytes
	peakMemory := pm.peakMemory
	pm.mu.Unlock()

	if pm.logger != nil {
		ctx := pm.logManager.CreateContext("Stop")
		defer ctx.Release()

		ctx.WithField("end_time", pm.endTime.Format(time.RFC3339)).
			WithField("processing_time_ms", processingTime.Milliseconds()).
			WithField("processing_time_seconds", processingTime.Seconds()).
			WithField("total_chunks", chunkCount).
			WithField("total_bytes", totalBytes).
			WithField("chunk_bytes", chunkBytes).
			WithField("memory_used_bytes", memoryUsed).
			WithField("memory_used_mb", memoryUsed/(1024*1024)).
			WithField("peak_memory_bytes", peakMemory).
			WithField("peak_memory_mb", peakMemory/(1024*1024))

		// Calculate performance metrics
		if processingTime.Seconds() > 0 {
			chunksPerSecond := float64(chunkCount) / processingTime.Seconds()
			bytesPerSecond := float64(totalBytes) / processingTime.Seconds()

			ctx.WithField("chunks_per_second", chunksPerSecond).
				WithField("bytes_per_second", bytesPerSecond).
				WithField("mb_per_second", bytesPerSecond/(1024*1024))
		}

		fields := ctx.ToLogFields()
		defer ctx.ReleaseFields(fields)

		pm.logger.Infow("Performance monitoring completed", fields...)
	}
}

// OptimizedMemoryLimiterFinal provides efficient memory limit checking
type OptimizedMemoryLimiterFinal struct {
	maxMemoryBytes uint64
	mu             sync.RWMutex
	logger         log.Logger
	logManager     *OptimizedLogContextManager
	lastCheck      time.Time
	checkInterval  time.Duration
}

// NewOptimizedMemoryLimiterFinal creates an optimized memory limiter
func NewOptimizedMemoryLimiterFinal(maxMemoryBytes uint64, logger log.Logger) *OptimizedMemoryLimiterFinal {
	return &OptimizedMemoryLimiterFinal{
		maxMemoryBytes: maxMemoryBytes,
		logger:         logger,
		logManager:     NewOptimizedLogContextManager(logger),
		checkInterval:  time.Second,
	}
}

// CheckMemoryLimit checks memory with throttling
func (ml *OptimizedMemoryLimiterFinal) CheckMemoryLimit() error {
	if ml.maxMemoryBytes == 0 {
		return nil
	}

	now := time.Now()
	ml.mu.RLock()
	shouldCheck := now.Sub(ml.lastCheck) >= ml.checkInterval
	ml.mu.RUnlock()

	if !shouldCheck {
		return nil
	}

	ml.mu.Lock()
	if now.Sub(ml.lastCheck) >= ml.checkInterval {
		ml.lastCheck = now
		ml.mu.Unlock()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		currentMemory := m.Alloc
		usageRatio := float64(currentMemory) / float64(ml.maxMemoryBytes)

		if ml.logger != nil && usageRatio > 0.8 {
			ctx := ml.logManager.CreateContext("CheckMemoryLimit")
			defer ctx.Release()

			ctx.WithField("current_memory_mb", currentMemory/(1024*1024)).
				WithField("memory_limit_mb", ml.maxMemoryBytes/(1024*1024)).
				WithField("usage_percentage", usageRatio*100).
				WithField("gc_count", m.NumGC)

			fields := ctx.ToLogFields()
			defer ctx.ReleaseFields(fields)

			if usageRatio > 0.95 {
				ml.logger.Errorw("Memory usage critical", fields...)
			} else {
				ml.logger.Warnw("Memory usage high", fields...)
			}
		}

		if currentMemory > ml.maxMemoryBytes {
			return NewChunkerError(ErrorTypeMemoryExhausted,
				fmt.Sprintf("Memory usage %d MB exceeds limit %d MB",
					currentMemory/(1024*1024), ml.maxMemoryBytes/(1024*1024)), nil).
				WithContext("current_memory_bytes", currentMemory).
				WithContext("memory_limit_bytes", ml.maxMemoryBytes).
				WithContext("usage_ratio", usageRatio)
		}

		return nil
	}
	ml.mu.Unlock()
	return nil
}

// OptimizedStringBuilderManager provides efficient string building
type OptimizedStringBuilderManager struct {
	pool sync.Pool
}

// NewOptimizedStringBuilderManager creates a new string builder manager
func NewOptimizedStringBuilderManager() *OptimizedStringBuilderManager {
	return &OptimizedStringBuilderManager{
		pool: sync.Pool{
			New: func() interface{} {
				sb := &strings.Builder{}
				sb.Grow(512)
				return sb
			},
		},
	}
}

// Get retrieves a string builder from the pool
func (sbm *OptimizedStringBuilderManager) Get() *strings.Builder {
	sb := sbm.pool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// Put returns a string builder to the pool
func (sbm *OptimizedStringBuilderManager) Put(sb *strings.Builder) {
	if sb.Cap() <= 4096 {
		sbm.pool.Put(sb)
	}
}

// JoinStrings efficiently joins strings
func (sbm *OptimizedStringBuilderManager) JoinStrings(strs []string, separator string) string {
	switch len(strs) {
	case 0:
		return ""
	case 1:
		return strs[0]
	case 2:
		if separator == "" {
			return strs[0] + strs[1]
		}
		return strs[0] + separator + strs[1]
	}

	sb := sbm.Get()
	defer sbm.Put(sb)

	// Calculate total length
	totalLen := 0
	sepLen := len(separator)
	for _, str := range strs {
		totalLen += len(str)
	}
	if sepLen > 0 {
		totalLen += sepLen * (len(strs) - 1)
	}

	sb.Grow(totalLen)
	sb.WriteString(strs[0])
	for i := 1; i < len(strs); i++ {
		if sepLen > 0 {
			sb.WriteString(separator)
		}
		sb.WriteString(strs[i])
	}

	return sb.String()
}
