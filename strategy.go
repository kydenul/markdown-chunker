package markdownchunker

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// ChunkingStrategy 定义分块策略的核心接口
type ChunkingStrategy interface {
	// GetName 返回策略名称
	GetName() string

	// GetDescription 返回策略描述
	GetDescription() string

	// ChunkDocument 使用该策略对文档进行分块
	ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error)

	// ValidateConfig 验证策略特定的配置
	ValidateConfig(config *StrategyConfig) error

	// Clone 创建策略的副本（用于并发安全）
	Clone() ChunkingStrategy
}

// StrategyConfig 策略配置结构
type StrategyConfig struct {
	// 通用配置
	Name       string         `json:"name"`       // 策略名称
	Parameters map[string]any `json:"parameters"` // 策略参数

	// 层级策略特定配置
	MaxDepth   int  `json:"max_depth,omitempty"`   // 最大层级深度
	MinDepth   int  `json:"min_depth,omitempty"`   // 最小层级深度
	MergeEmpty bool `json:"merge_empty,omitempty"` // 是否合并空章节

	// 大小限制配置
	MinChunkSize int `json:"min_chunk_size,omitempty"` // 最小块大小
	MaxChunkSize int `json:"max_chunk_size,omitempty"` // 最大块大小

	// 内容过滤配置
	IncludeTypes []string `json:"include_types,omitempty"` // 包含的内容类型
	ExcludeTypes []string `json:"exclude_types,omitempty"` // 排除的内容类型
}

// StrategyRegistry 策略注册器
type StrategyRegistry struct {
	strategies map[string]ChunkingStrategy
	mutex      sync.RWMutex
}

// NewStrategyRegistry 创建策略注册器
func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[string]ChunkingStrategy),
	}
}

// Register 注册策略
func (sr *StrategyRegistry) Register(strategy ChunkingStrategy) error {
	if strategy == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略不能为空", nil).
			WithContext("function", "Register")
	}

	name := strategy.GetName()
	if name == "" {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略名称不能为空", nil).
			WithContext("function", "Register")
	}

	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if _, exists := sr.strategies[name]; exists {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略已存在", nil).
			WithContext("function", "Register").
			WithContext("strategy_name", name)
	}

	sr.strategies[name] = strategy
	return nil
}

// Get 获取策略
func (sr *StrategyRegistry) Get(name string) (ChunkingStrategy, error) {
	if name == "" {
		return nil, NewChunkerError(ErrorTypeStrategyNotFound, "策略名称不能为空", nil).
			WithContext("function", "Get")
	}

	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	strategy, exists := sr.strategies[name]
	if !exists {
		return nil, NewChunkerError(ErrorTypeStrategyNotFound, "策略未找到", nil).
			WithContext("function", "Get").
			WithContext("strategy_name", name)
	}

	return strategy, nil
}

// List 列出所有可用策略
func (sr *StrategyRegistry) List() []string {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	names := make([]string, 0, len(sr.strategies))
	for name := range sr.strategies {
		names = append(names, name)
	}
	return names
}

// Unregister 注销策略
func (sr *StrategyRegistry) Unregister(name string) error {
	if name == "" {
		return NewChunkerError(ErrorTypeStrategyNotFound, "策略名称不能为空", nil).
			WithContext("function", "Unregister")
	}

	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if _, exists := sr.strategies[name]; !exists {
		return NewChunkerError(ErrorTypeStrategyNotFound, "策略未找到", nil).
			WithContext("function", "Unregister").
			WithContext("strategy_name", name)
	}

	delete(sr.strategies, name)
	return nil
}

// GetStrategyCount 获取已注册策略数量
func (sr *StrategyRegistry) GetStrategyCount() int {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()
	return len(sr.strategies)
}

// HasStrategy 检查策略是否存在
func (sr *StrategyRegistry) HasStrategy(name string) bool {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()
	_, exists := sr.strategies[name]
	return exists
}

// ValidateConfig 验证策略配置
func (sc *StrategyConfig) ValidateConfig() error {
	if sc == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置不能为空", nil).
			WithContext("function", "ValidateConfig")
	}

	if sc.Name == "" {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略名称不能为空", nil).
			WithContext("function", "ValidateConfig")
	}

	// 验证层级配置
	if sc.MaxDepth < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大层级深度不能为负数", nil).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MaxDepth").
			WithContext("value", sc.MaxDepth)
	}

	if sc.MinDepth < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小层级深度不能为负数", nil).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MinDepth").
			WithContext("value", sc.MinDepth)
	}

	if sc.MaxDepth > 0 && sc.MinDepth > 0 && sc.MinDepth > sc.MaxDepth {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小层级深度不能大于最大层级深度", nil).
			WithContext("function", "ValidateConfig").
			WithContext("min_depth", sc.MinDepth).
			WithContext("max_depth", sc.MaxDepth)
	}

	// 验证大小限制配置
	if sc.MinChunkSize < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小块大小不能为负数", nil).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MinChunkSize").
			WithContext("value", sc.MinChunkSize)
	}

	if sc.MaxChunkSize < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大块大小不能为负数", nil).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MaxChunkSize").
			WithContext("value", sc.MaxChunkSize)
	}

	if sc.MaxChunkSize > 0 && sc.MinChunkSize > 0 && sc.MinChunkSize > sc.MaxChunkSize {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小块大小不能大于最大块大小", nil).
			WithContext("function", "ValidateConfig").
			WithContext("min_chunk_size", sc.MinChunkSize).
			WithContext("max_chunk_size", sc.MaxChunkSize)
	}

	// 验证内容类型配置
	validTypes := map[string]bool{
		"heading": true, "paragraph": true, "code": true,
		"table": true, "list": true, "blockquote": true,
		"thematic_break": true,
	}

	for _, includeType := range sc.IncludeTypes {
		if !validTypes[includeType] {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "无效的包含内容类型", nil).
				WithContext("function", "ValidateConfig").
				WithContext("field", "IncludeTypes").
				WithContext("invalid_type", includeType)
		}
	}

	for _, excludeType := range sc.ExcludeTypes {
		if !validTypes[excludeType] {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "无效的排除内容类型", nil).
				WithContext("function", "ValidateConfig").
				WithContext("field", "ExcludeTypes").
				WithContext("invalid_type", excludeType)
		}
	}

	return nil
}

// Clone 创建策略配置的副本
func (sc *StrategyConfig) Clone() *StrategyConfig {
	if sc == nil {
		return nil
	}

	clone := &StrategyConfig{
		Name:         sc.Name,
		MaxDepth:     sc.MaxDepth,
		MinDepth:     sc.MinDepth,
		MergeEmpty:   sc.MergeEmpty,
		MinChunkSize: sc.MinChunkSize,
		MaxChunkSize: sc.MaxChunkSize,
	}

	// 深拷贝参数映射
	if sc.Parameters != nil {
		clone.Parameters = make(map[string]any)
		maps.Copy(clone.Parameters, sc.Parameters)
	}

	// 深拷贝切片
	if sc.IncludeTypes != nil {
		clone.IncludeTypes = make([]string, len(sc.IncludeTypes))
		copy(clone.IncludeTypes, sc.IncludeTypes)
	}

	if sc.ExcludeTypes != nil {
		clone.ExcludeTypes = make([]string, len(sc.ExcludeTypes))
		copy(clone.ExcludeTypes, sc.ExcludeTypes)
	}

	return clone
}

// String 返回策略配置的字符串表示
func (sc *StrategyConfig) String() string {
	if sc == nil {
		return "<nil>"
	}
	return fmt.Sprintf("StrategyConfig{Name: %s, MaxDepth: %d, MinDepth: %d, MergeEmpty: %t, MinChunkSize: %d, MaxChunkSize: %d}",
		sc.Name, sc.MaxDepth, sc.MinDepth, sc.MergeEmpty, sc.MinChunkSize, sc.MaxChunkSize)
}

// DefaultStrategyConfig 创建默认策略配置
func DefaultStrategyConfig(name string) *StrategyConfig {
	return &StrategyConfig{
		Name:         name,
		Parameters:   make(map[string]any),
		MaxDepth:     0, // 无限制
		MinDepth:     0, // 无限制
		MergeEmpty:   true,
		MinChunkSize: 0, // 无限制
		MaxChunkSize: 0, // 无限制
		IncludeTypes: nil,
		ExcludeTypes: nil,
	}
}

// HierarchicalConfig 创建层级策略配置
func HierarchicalConfig(maxDepth int) *StrategyConfig {
	config := DefaultStrategyConfig("hierarchical")
	config.MaxDepth = maxDepth
	config.Parameters["max_depth"] = maxDepth
	config.Parameters["merge_empty"] = true
	return config
}

// DocumentLevelConfig 创建文档级策略配置
func DocumentLevelConfig() *StrategyConfig {
	return DefaultStrategyConfig("document-level")
}

// ElementLevelConfig 创建元素级策略配置
func ElementLevelConfig() *StrategyConfig {
	return DefaultStrategyConfig("element-level")
}

// ElementLevelConfigWithTypes 创建带内容类型过滤的元素级策略配置
func ElementLevelConfigWithTypes(includeTypes, excludeTypes []string) *StrategyConfig {
	config := DefaultStrategyConfig("element-level")
	config.IncludeTypes = includeTypes
	config.ExcludeTypes = excludeTypes

	// 添加到参数映射中
	if len(includeTypes) > 0 {
		config.Parameters["include_types"] = includeTypes
	}
	if len(excludeTypes) > 0 {
		config.Parameters["exclude_types"] = excludeTypes
	}

	return config
}

// ElementLevelConfigWithSize 创建带大小限制的元素级策略配置
func ElementLevelConfigWithSize(minSize, maxSize int) *StrategyConfig {
	config := DefaultStrategyConfig("element-level")
	config.MinChunkSize = minSize
	config.MaxChunkSize = maxSize

	// 添加到参数映射中
	config.Parameters["min_chunk_size"] = minSize
	config.Parameters["max_chunk_size"] = maxSize

	return config
}

// HierarchicalConfigAdvanced 创建高级层级策略配置
func HierarchicalConfigAdvanced(maxDepth, minDepth int, mergeEmpty bool) *StrategyConfig {
	config := DefaultStrategyConfig("hierarchical")
	config.MaxDepth = maxDepth
	config.MinDepth = minDepth
	config.MergeEmpty = mergeEmpty

	// 添加到参数映射中
	config.Parameters["max_depth"] = maxDepth
	config.Parameters["min_depth"] = minDepth
	config.Parameters["merge_empty"] = mergeEmpty

	return config
}

// HierarchicalConfigWithSize 创建带大小限制的层级策略配置
func HierarchicalConfigWithSize(maxDepth, minSize, maxSize int) *StrategyConfig {
	config := HierarchicalConfig(maxDepth)
	config.MinChunkSize = minSize
	config.MaxChunkSize = maxSize

	// 添加到参数映射中
	config.Parameters["min_chunk_size"] = minSize
	config.Parameters["max_chunk_size"] = maxSize

	return config
}

// DocumentLevelConfigWithSize 创建带大小限制的文档级策略配置
func DocumentLevelConfigWithSize(minSize, maxSize int) *StrategyConfig {
	config := DefaultStrategyConfig("document-level")
	config.MinChunkSize = minSize
	config.MaxChunkSize = maxSize

	// 添加到参数映射中
	config.Parameters["min_chunk_size"] = minSize
	config.Parameters["max_chunk_size"] = maxSize

	return config
}

// ValidateAndFillDefaults 验证策略配置并填充默认值
func ValidateAndFillDefaults(config *StrategyConfig) error {
	if config == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置不能为空", nil).
			WithContext("function", "ValidateAndFillDefaults")
	}

	// 验证策略名称
	if config.Name == "" {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略名称不能为空", nil).
			WithContext("function", "ValidateAndFillDefaults")
	}

	// 初始化参数映射
	if config.Parameters == nil {
		config.Parameters = make(map[string]any)
	}

	// 根据策略类型填充默认值和验证
	switch config.Name {
	case "hierarchical":
		return validateAndFillHierarchicalDefaults(config)
	case "element-level":
		return validateAndFillElementLevelDefaults(config)
	case "document-level":
		return validateAndFillDocumentLevelDefaults(config)
	default:
		// 对于自定义策略，只进行基本验证
		return config.ValidateConfig()
	}
}

// validateAndFillHierarchicalDefaults 验证并填充层级策略的默认值
func validateAndFillHierarchicalDefaults(config *StrategyConfig) error {
	// 验证层级深度
	if config.MaxDepth < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大层级深度不能为负数", nil).
			WithContext("function", "validateAndFillHierarchicalDefaults").
			WithContext("max_depth", config.MaxDepth)
	}

	if config.MinDepth < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小层级深度不能为负数", nil).
			WithContext("function", "validateAndFillHierarchicalDefaults").
			WithContext("min_depth", config.MinDepth)
	}

	if config.MaxDepth > 6 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大层级深度不能超过6", nil).
			WithContext("function", "validateAndFillHierarchicalDefaults").
			WithContext("max_depth", config.MaxDepth)
	}

	if config.MaxDepth > 0 && config.MinDepth > 0 && config.MinDepth > config.MaxDepth {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小层级深度不能大于最大层级深度", nil).
			WithContext("function", "validateAndFillHierarchicalDefaults").
			WithContext("min_depth", config.MinDepth).
			WithContext("max_depth", config.MaxDepth)
	}

	// 填充参数映射
	config.Parameters["max_depth"] = config.MaxDepth
	config.Parameters["min_depth"] = config.MinDepth
	config.Parameters["merge_empty"] = config.MergeEmpty

	return config.ValidateConfig()
}

// validateAndFillElementLevelDefaults 验证并填充元素级策略的默认值
func validateAndFillElementLevelDefaults(config *StrategyConfig) error {
	// 元素级策略不应该有层级配置
	if config.MaxDepth > 0 || config.MinDepth > 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "元素级策略不支持层级深度配置", nil).
			WithContext("function", "validateAndFillElementLevelDefaults").
			WithContext("max_depth", config.MaxDepth).
			WithContext("min_depth", config.MinDepth)
	}

	// 填充参数映射
	if len(config.IncludeTypes) > 0 {
		config.Parameters["include_types"] = config.IncludeTypes
	}
	if len(config.ExcludeTypes) > 0 {
		config.Parameters["exclude_types"] = config.ExcludeTypes
	}

	return config.ValidateConfig()
}

// validateAndFillDocumentLevelDefaults 验证并填充文档级策略的默认值
func validateAndFillDocumentLevelDefaults(config *StrategyConfig) error {
	// 文档级策略不应该有层级配置
	if config.MaxDepth > 0 || config.MinDepth > 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "文档级策略不支持层级深度配置", nil).
			WithContext("function", "validateAndFillDocumentLevelDefaults").
			WithContext("max_depth", config.MaxDepth).
			WithContext("min_depth", config.MinDepth)
	}

	// 文档级策略不应该有内容类型过滤
	if len(config.IncludeTypes) > 0 || len(config.ExcludeTypes) > 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "文档级策略不支持内容类型过滤", nil).
			WithContext("function", "validateAndFillDocumentLevelDefaults").
			WithContext("include_types", config.IncludeTypes).
			WithContext("exclude_types", config.ExcludeTypes)
	}

	return config.ValidateConfig()
}

// CreateConfigFromParameters 从参数映射创建策略配置
func CreateConfigFromParameters(strategyName string, params map[string]any) (*StrategyConfig, error) {
	if strategyName == "" {
		return nil, NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略名称不能为空", nil).
			WithContext("function", "CreateConfigFromParameters")
	}

	config := DefaultStrategyConfig(strategyName)

	// 从参数映射中提取配置
	if maxDepth, ok := params["max_depth"]; ok {
		if depth, ok := maxDepth.(int); ok {
			config.MaxDepth = depth
		}
	}

	if minDepth, ok := params["min_depth"]; ok {
		if depth, ok := minDepth.(int); ok {
			config.MinDepth = depth
		}
	}

	if mergeEmpty, ok := params["merge_empty"]; ok {
		if merge, ok := mergeEmpty.(bool); ok {
			config.MergeEmpty = merge
		}
	}

	if minSize, ok := params["min_chunk_size"]; ok {
		if size, ok := minSize.(int); ok {
			config.MinChunkSize = size
		}
	}

	if maxSize, ok := params["max_chunk_size"]; ok {
		if size, ok := maxSize.(int); ok {
			config.MaxChunkSize = size
		}
	}

	if includeTypes, ok := params["include_types"]; ok {
		if types, ok := includeTypes.([]string); ok {
			config.IncludeTypes = types
		}
	}

	if excludeTypes, ok := params["exclude_types"]; ok {
		if types, ok := excludeTypes.([]string); ok {
			config.ExcludeTypes = types
		}
	}

	// 复制参数映射
	maps.Copy(config.Parameters, params)

	// 验证并填充默认值
	if err := ValidateAndFillDefaults(config); err != nil {
		return nil, err
	}

	return config, nil
}

// MergeConfigs 合并两个策略配置
func MergeConfigs(base, override *StrategyConfig) (*StrategyConfig, error) {
	if base == nil {
		return nil, NewChunkerError(ErrorTypeStrategyConfigInvalid, "基础配置不能为空", nil).
			WithContext("function", "MergeConfigs")
	}

	// 如果覆盖配置为空，返回基础配置的副本
	if override == nil {
		return base.Clone(), nil
	}

	// 策略名称必须匹配
	if base.Name != override.Name {
		return nil, NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略名称不匹配", nil).
			WithContext("function", "MergeConfigs").
			WithContext("base_name", base.Name).
			WithContext("override_name", override.Name)
	}

	// 创建合并后的配置
	merged := base.Clone()

	// 合并字段（override 优先）
	if override.MaxDepth != 0 {
		merged.MaxDepth = override.MaxDepth
	}
	if override.MinDepth != 0 {
		merged.MinDepth = override.MinDepth
	}
	if override.MinChunkSize != 0 {
		merged.MinChunkSize = override.MinChunkSize
	}
	if override.MaxChunkSize != 0 {
		merged.MaxChunkSize = override.MaxChunkSize
	}
	// 对于布尔值，我们需要检查参数映射来确定是否应该覆盖
	if mergeEmptyParam, exists := override.Parameters["merge_empty"]; exists {
		if mergeEmpty, ok := mergeEmptyParam.(bool); ok {
			merged.MergeEmpty = mergeEmpty
		}
	}
	if override.IncludeTypes != nil {
		merged.IncludeTypes = make([]string, len(override.IncludeTypes))
		copy(merged.IncludeTypes, override.IncludeTypes)
	}
	if override.ExcludeTypes != nil {
		merged.ExcludeTypes = make([]string, len(override.ExcludeTypes))
		copy(merged.ExcludeTypes, override.ExcludeTypes)
	}

	// 合并参数映射
	maps.Copy(merged.Parameters, override.Parameters)

	// 验证合并后的配置
	if err := ValidateAndFillDefaults(merged); err != nil {
		return nil, err
	}

	return merged, nil
}

// StrategyCache 策略缓存
type StrategyCache struct {
	cache map[string]ChunkingStrategy
	mutex sync.RWMutex
}

// NewStrategyCache 创建策略缓存
func NewStrategyCache() *StrategyCache {
	return &StrategyCache{
		cache: make(map[string]ChunkingStrategy),
	}
}

// Get 从缓存获取策略
func (sc *StrategyCache) Get(name string) (ChunkingStrategy, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	strategy, exists := sc.cache[name]
	return strategy, exists
}

// Put 将策略放入缓存
func (sc *StrategyCache) Put(name string, strategy ChunkingStrategy) {
	if name == "" || strategy == nil {
		return
	}

	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.cache[name] = strategy
}

// Remove 从缓存移除策略
func (sc *StrategyCache) Remove(name string) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	delete(sc.cache, name)
}

// Clear 清空缓存
func (sc *StrategyCache) Clear() {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	sc.cache = make(map[string]ChunkingStrategy)
}

// Size 获取缓存大小
func (sc *StrategyCache) Size() int {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	return len(sc.cache)
}

// Keys 获取所有缓存的策略名称
func (sc *StrategyCache) Keys() []string {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()

	keys := make([]string, 0, len(sc.cache))
	for key := range sc.cache {
		keys = append(keys, key)
	}
	return keys
}

// StrategyPool 策略实例池
type StrategyPool struct {
	pools map[string]*sync.Pool
	mutex sync.RWMutex
}

// NewStrategyPool 创建策略池
func NewStrategyPool() *StrategyPool {
	return &StrategyPool{
		pools: make(map[string]*sync.Pool),
	}
}

// Get 从池中获取策略实例
func (sp *StrategyPool) Get(strategyName string, factory func() ChunkingStrategy) ChunkingStrategy {
	sp.mutex.RLock()
	pool, exists := sp.pools[strategyName]
	sp.mutex.RUnlock()

	if !exists {
		// 创建新的池
		sp.mutex.Lock()
		// 双重检查锁定
		if pool, exists = sp.pools[strategyName]; !exists {
			pool = &sync.Pool{
				New: func() any {
					if factory != nil {
						return factory()
					}
					return nil
				},
			}
			sp.pools[strategyName] = pool
		}
		sp.mutex.Unlock()
	}

	if strategy := pool.Get(); strategy != nil {
		if cs, ok := strategy.(ChunkingStrategy); ok {
			return cs
		}
	}

	// 如果池中没有可用实例，使用工厂函数创建
	if factory != nil {
		return factory()
	}

	return nil
}

// Put 将策略实例放回池中
func (sp *StrategyPool) Put(strategy ChunkingStrategy) {
	if strategy == nil {
		return
	}

	strategyName := strategy.GetName()
	if strategyName == "" {
		return
	}

	sp.mutex.RLock()
	pool, exists := sp.pools[strategyName]
	sp.mutex.RUnlock()

	if exists {
		pool.Put(strategy)
	}
}

// CreatePool 为指定策略创建池
func (sp *StrategyPool) CreatePool(strategyName string, factory func() ChunkingStrategy) {
	if strategyName == "" || factory == nil {
		return
	}

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	if _, exists := sp.pools[strategyName]; !exists {
		sp.pools[strategyName] = &sync.Pool{
			New: func() any {
				return factory()
			},
		}
	}
}

// RemovePool 移除指定策略的池
func (sp *StrategyPool) RemovePool(strategyName string) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	delete(sp.pools, strategyName)
}

// Clear 清空所有池
func (sp *StrategyPool) Clear() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.pools = make(map[string]*sync.Pool)
}

// GetPoolCount 获取池的数量
func (sp *StrategyPool) GetPoolCount() int {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	return len(sp.pools)
}

// HasPool 检查是否存在指定策略的池
func (sp *StrategyPool) HasPool(strategyName string) bool {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	_, exists := sp.pools[strategyName]
	return exists
}

// ElementLevelStrategy 元素级分块策略（默认策略）
// 按 Markdown 元素类型逐个分块，保持与当前行为完全一致
type ElementLevelStrategy struct {
	config *StrategyConfig
}

// NewElementLevelStrategy 创建新的元素级分块策略
func NewElementLevelStrategy() *ElementLevelStrategy {
	return &ElementLevelStrategy{
		config: ElementLevelConfig(),
	}
}

// NewElementLevelStrategyWithConfig 使用指定配置创建元素级分块策略
func NewElementLevelStrategyWithConfig(config *StrategyConfig) *ElementLevelStrategy {
	if config == nil {
		config = ElementLevelConfig()
	}
	return &ElementLevelStrategy{
		config: config,
	}
}

// GetName 返回策略名称
func (s *ElementLevelStrategy) GetName() string {
	return "element-level"
}

// GetDescription 返回策略描述
func (s *ElementLevelStrategy) GetDescription() string {
	return "按 Markdown 元素类型逐个分块，保持原有行为"
}

// ChunkDocument 使用元素级策略对文档进行分块
func (s *ElementLevelStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	if doc == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "文档节点不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	if chunker == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "分块器实例不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	// 设置分块器的源内容，这是processNode方法正常工作所必需的
	originalSource := chunker.source
	chunker.source = source
	defer func() {
		// 恢复原始源内容
		chunker.source = originalSource
	}()

	var chunks []Chunk
	chunkID := 0

	// 遍历文档的所有顶层子节点
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		// 使用分块器的 processNode 方法处理每个节点
		// 这保持了与当前分块行为完全一致的逻辑
		if chunk := chunker.processNode(child, chunkID); chunk != nil {
			// 应用策略特定的过滤和处理
			if s.shouldIncludeChunk(chunk) {
				chunks = append(chunks, *chunk)
				chunkID++
			}
		}
	}

	return chunks, nil
}

// ValidateConfig 验证策略特定的配置
func (s *ElementLevelStrategy) ValidateConfig(config *StrategyConfig) error {
	if config == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ValidateConfig")
	}

	// 元素级策略不需要特殊的配置验证，使用通用验证
	return config.ValidateConfig()
}

// Clone 创建策略的副本（用于并发安全）
func (s *ElementLevelStrategy) Clone() ChunkingStrategy {
	var configClone *StrategyConfig
	if s.config != nil {
		configClone = s.config.Clone()
	}

	return &ElementLevelStrategy{
		config: configClone,
	}
}

// shouldIncludeChunk 判断是否应该包含指定的块
func (s *ElementLevelStrategy) shouldIncludeChunk(chunk *Chunk) bool {
	if chunk == nil {
		return false
	}

	// 如果配置了包含类型，检查块类型是否在包含列表中
	if s.config != nil && len(s.config.IncludeTypes) > 0 {
		included := slices.Contains(s.config.IncludeTypes, chunk.Type)
		if !included {
			return false
		}
	}

	// 如果配置了排除类型，检查块类型是否在排除列表中
	if s.config != nil && len(s.config.ExcludeTypes) > 0 {
		if slices.Contains(s.config.ExcludeTypes, chunk.Type) {
			return false
		}
	}

	// 检查块大小限制
	if s.config != nil {
		contentLength := len(chunk.Content)

		// 检查最小块大小
		if s.config.MinChunkSize > 0 && contentLength < s.config.MinChunkSize {
			return false
		}

		// 检查最大块大小
		if s.config.MaxChunkSize > 0 && contentLength > s.config.MaxChunkSize {
			return false
		}
	}

	return true
}

// GetConfig 获取策略配置
func (s *ElementLevelStrategy) GetConfig() *StrategyConfig {
	if s.config == nil {
		return nil
	}
	return s.config.Clone()
}

// SetConfig 设置策略配置
func (s *ElementLevelStrategy) SetConfig(config *StrategyConfig) error {
	if config == nil {
		s.config = ElementLevelConfig()
		return nil
	}

	if err := s.ValidateConfig(config); err != nil {
		return err
	}

	s.config = config.Clone()
	return nil
}

// HierarchicalChunk 表示层级结构中的块
type HierarchicalChunk struct {
	Chunk    Chunk                `json:"chunk"`    // 基础块信息
	Children []*HierarchicalChunk `json:"children"` // 子块列表
	Parent   *HierarchicalChunk   `json:"-"`        // 父块引用（不序列化）
	Level    int                  `json:"level"`    // 层级深度
}

// HierarchicalStrategy 层级分块策略
// 按文档层级结构分块，将标题及其下属内容作为一个块
type HierarchicalStrategy struct {
	config *StrategyConfig
}

// NewHierarchicalStrategy 创建新的层级分块策略
func NewHierarchicalStrategy() *HierarchicalStrategy {
	return &HierarchicalStrategy{
		config: HierarchicalConfig(0), // 默认无层级限制
	}
}

// NewHierarchicalStrategyWithConfig 使用指定配置创建层级分块策略
func NewHierarchicalStrategyWithConfig(config *StrategyConfig) *HierarchicalStrategy {
	if config == nil {
		config = HierarchicalConfig(0)
	}
	return &HierarchicalStrategy{
		config: config,
	}
}

// GetName 返回策略名称
func (s *HierarchicalStrategy) GetName() string {
	return "hierarchical"
}

// GetDescription 返回策略描述
func (s *HierarchicalStrategy) GetDescription() string {
	return "按文档层级结构分块，将标题及其下属内容作为一个块"
}

// ChunkDocument 使用层级策略对文档进行分块
func (s *HierarchicalStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	if doc == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "文档节点不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	if chunker == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "分块器实例不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	// 1. 先使用元素级策略获取所有基础块
	elementStrategy := NewElementLevelStrategy()
	baseChunks, err := elementStrategy.ChunkDocument(doc, source, chunker)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "获取基础块失败", err).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	// 2. 构建层级结构
	hierarchicalChunks := s.buildHierarchy(baseChunks)

	// 3. 根据配置扁平化为目标层级的块
	return s.flattenToTargetLevel(hierarchicalChunks), nil
}

// ValidateConfig 验证策略特定的配置
func (s *HierarchicalStrategy) ValidateConfig(config *StrategyConfig) error {
	if config == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ValidateConfig")
	}

	// 使用通用配置验证
	if err := config.ValidateConfig(); err != nil {
		return err
	}

	// 层级策略特定验证
	if config.MaxDepth > 6 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大层级深度不能超过6", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ValidateConfig").
			WithContext("field", "MaxDepth").
			WithContext("value", config.MaxDepth).
			WithContext("max_allowed", 6)
	}

	return nil
}

// Clone 创建策略的副本（用于并发安全）
func (s *HierarchicalStrategy) Clone() ChunkingStrategy {
	var configClone *StrategyConfig
	if s.config != nil {
		configClone = s.config.Clone()
	}

	return &HierarchicalStrategy{
		config: configClone,
	}
}

// buildHierarchy 构建层级结构
func (s *HierarchicalStrategy) buildHierarchy(chunks []Chunk) []*HierarchicalChunk {
	if len(chunks) == 0 {
		return []*HierarchicalChunk{}
	}

	var root []*HierarchicalChunk
	var stack []*HierarchicalChunk // 用于跟踪当前层级的父节点

	for _, chunk := range chunks {
		hierarchicalChunk := &HierarchicalChunk{
			Chunk:    chunk,
			Children: []*HierarchicalChunk{},
			Level:    0, // 将在下面设置
		}

		if chunk.Type == "heading" {
			headingLevel := chunk.Level

			// 处理无效的标题层级（应该在1-6之间）
			if headingLevel < 1 {
				headingLevel = 1
			} else if headingLevel > 6 {
				headingLevel = 6
			}

			// 清理栈，移除层级更深或相等的节点
			for len(stack) > 0 && stack[len(stack)-1].Chunk.Level >= headingLevel {
				stack = stack[:len(stack)-1]
			}

			// 处理跳跃的标题层级（例如从H1直接跳到H3）
			// 如果当前标题层级比栈顶标题层级高出超过1，需要创建中间层级
			if len(stack) > 0 {
				topLevel := stack[len(stack)-1].Chunk.Level
				if headingLevel > topLevel+1 {
					// 创建中间层级的虚拟标题
					for level := topLevel + 1; level < headingLevel; level++ {
						virtualChunk := &HierarchicalChunk{
							Chunk: Chunk{
								ID:      -1, // 使用负数ID标识虚拟块
								Type:    "heading",
								Content: "",
								Text:    "",
								Level:   level,
								Metadata: map[string]string{
									"virtual":       "true",
									"heading_level": fmt.Sprintf("%d", level),
								},
							},
							Children: []*HierarchicalChunk{},
							Level:    len(stack),
						}

						// 添加到父节点
						if len(stack) > 0 {
							parent := stack[len(stack)-1]
							virtualChunk.Parent = parent
							parent.Children = append(parent.Children, virtualChunk)
						} else {
							root = append(root, virtualChunk)
						}

						// 推入栈
						stack = append(stack, virtualChunk)
					}
				}
			}

			// 设置层级深度（基于栈的深度）
			hierarchicalChunk.Level = len(stack)

			// 如果有父节点，添加到父节点的子列表中
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				hierarchicalChunk.Parent = parent
				parent.Children = append(parent.Children, hierarchicalChunk)
			} else {
				// 顶级标题，添加到根列表
				root = append(root, hierarchicalChunk)
			}

			// 将当前标题推入栈
			stack = append(stack, hierarchicalChunk)
		} else {
			// 非标题内容
			if len(stack) > 0 {
				// 有当前标题，添加到最近的标题下
				parent := stack[len(stack)-1]
				hierarchicalChunk.Parent = parent
				hierarchicalChunk.Level = parent.Level + 1
				parent.Children = append(parent.Children, hierarchicalChunk)
			} else {
				// 没有标题，作为顶级内容
				// 创建一个虚拟的根节点来包含这些内容
				if len(root) == 0 || root[len(root)-1].Chunk.Type != "root" {
					virtualRoot := &HierarchicalChunk{
						Chunk: Chunk{
							ID:      -2, // 使用-2标识根虚拟块
							Type:    "root",
							Content: "",
							Text:    "",
							Level:   0,
							Metadata: map[string]string{
								"virtual": "true",
								"type":    "root",
							},
						},
						Children: []*HierarchicalChunk{},
						Level:    0,
					}
					root = append(root, virtualRoot)
				}

				// 添加到虚拟根节点
				virtualRoot := root[len(root)-1]
				hierarchicalChunk.Parent = virtualRoot
				hierarchicalChunk.Level = 1
				virtualRoot.Children = append(virtualRoot.Children, hierarchicalChunk)
			}
		}
	}

	return root
}

// flattenToTargetLevel 扁平化到目标层级
func (s *HierarchicalStrategy) flattenToTargetLevel(hierarchicalChunks []*HierarchicalChunk) []Chunk {
	var result []Chunk
	chunkID := 0

	// 递归遍历层级结构
	var traverse func([]*HierarchicalChunk, int)
	traverse = func(chunks []*HierarchicalChunk, currentDepth int) {
		for _, hChunk := range chunks {
			// 跳过虚拟块，除非它们包含实际内容
			if s.isVirtualChunk(hChunk) && !s.hasNonVirtualContent(hChunk) {
				// 直接处理子节点
				traverse(hChunk.Children, currentDepth)
				continue
			}

			// 检查是否应该在当前层级创建块
			shouldCreateChunk := s.shouldCreateChunkAtLevel(hChunk, currentDepth)

			if shouldCreateChunk {
				// 创建合并的块
				mergedChunk := s.createMergedChunk(hChunk, chunkID)
				if mergedChunk != nil {
					result = append(result, *mergedChunk)
					chunkID++
				}
			} else {
				// 继续遍历子节点
				traverse(hChunk.Children, currentDepth+1)
			}
		}
	}

	traverse(hierarchicalChunks, 0)
	return result
}

// isVirtualChunk 检查是否为虚拟块
func (s *HierarchicalStrategy) isVirtualChunk(hChunk *HierarchicalChunk) bool {
	if hChunk == nil {
		return false
	}

	virtual, exists := hChunk.Chunk.Metadata["virtual"]
	return exists && virtual == "true"
}

// hasNonVirtualContent 检查层级块是否包含非虚拟内容
func (s *HierarchicalStrategy) hasNonVirtualContent(hChunk *HierarchicalChunk) bool {
	if hChunk == nil {
		return false
	}

	// 如果当前块不是虚拟的且有内容，返回true
	if !s.isVirtualChunk(hChunk) && strings.TrimSpace(hChunk.Chunk.Text) != "" {
		return true
	}

	// 递归检查子块
	return slices.ContainsFunc(hChunk.Children, s.hasNonVirtualContent)
}

// shouldCreateChunkAtLevel 判断是否应该在指定层级创建块
func (s *HierarchicalStrategy) shouldCreateChunkAtLevel(hChunk *HierarchicalChunk, currentDepth int) bool {
	if hChunk == nil {
		return false
	}

	// 如果配置了最大深度，检查是否达到最大深度
	if s.config.MaxDepth > 0 && currentDepth >= s.config.MaxDepth {
		return true
	}

	// 如果配置了最小深度，检查是否达到最小深度
	if s.config.MinDepth > 0 && currentDepth < s.config.MinDepth {
		return false
	}

	// 对于虚拟块，只有在包含实际内容时才创建
	if s.isVirtualChunk(hChunk) {
		return s.hasNonVirtualContent(hChunk)
	}

	// 如果是标题且有子内容，考虑合并
	if hChunk.Chunk.Type == "heading" && len(hChunk.Children) > 0 {
		// 如果启用了空章节合并，检查是否有实际内容
		if s.config.MergeEmpty {
			return s.hasNonEmptyContent(hChunk)
		}
		return true
	}

	// 对于叶子节点（没有子节点的节点），总是创建块
	if len(hChunk.Children) == 0 {
		return true
	}

	// 对于有子节点的非标题内容，检查是否应该合并
	// 这通常发生在根级别的内容组织中
	if hChunk.Chunk.Type != "heading" {
		// 如果内容很少，可能需要与子内容合并
		contentLength := len(strings.TrimSpace(hChunk.Chunk.Text))
		if contentLength < 50 { // 少于50个字符的内容考虑合并
			return true
		}
	}

	// 默认情况下，在当前层级创建块
	return true
}

// hasNonEmptyContent 检查层级块是否包含非空内容
func (s *HierarchicalStrategy) hasNonEmptyContent(hChunk *HierarchicalChunk) bool {
	if hChunk == nil {
		return false
	}

	// 检查当前块是否有内容
	if strings.TrimSpace(hChunk.Chunk.Text) != "" {
		return true
	}

	// 递归检查子块
	return slices.ContainsFunc(hChunk.Children, s.hasNonEmptyContent)
}

// createMergedChunk 创建合并的块
func (s *HierarchicalStrategy) createMergedChunk(hChunk *HierarchicalChunk, id int) *Chunk {
	if hChunk == nil {
		return nil
	}

	// 收集所有内容
	var contentParts []string
	var textParts []string
	var allLinks []Link
	var allImages []Image
	var firstPosition, lastPosition ChunkPosition
	var hasPosition bool

	// 递归收集内容
	var collectContent func(*HierarchicalChunk, bool)
	collectContent = func(chunk *HierarchicalChunk, isFirst bool) {
		if chunk == nil {
			return
		}

		// 跳过虚拟块的内容，但处理其子块
		if s.isVirtualChunk(chunk) {
			for i, child := range chunk.Children {
				collectContent(child, isFirst && i == 0)
			}
			return
		}

		// 添加当前块的内容
		if strings.TrimSpace(chunk.Chunk.Content) != "" {
			contentParts = append(contentParts, chunk.Chunk.Content)
		}
		if strings.TrimSpace(chunk.Chunk.Text) != "" {
			textParts = append(textParts, chunk.Chunk.Text)
		}

		// 收集链接和图片
		allLinks = append(allLinks, chunk.Chunk.Links...)
		allImages = append(allImages, chunk.Chunk.Images...)

		// 记录位置信息
		if !hasPosition {
			firstPosition = chunk.Chunk.Position
			hasPosition = true
		}
		lastPosition = chunk.Chunk.Position

		// 递归处理子块
		for _, child := range chunk.Children {
			collectContent(child, false)
		}
	}

	collectContent(hChunk, true)

	// 如果没有内容，返回nil
	if len(contentParts) == 0 && len(textParts) == 0 {
		return nil
	}

	// 合并内容，保持适当的间距
	mergedContent := strings.Join(contentParts, "\n\n")
	mergedText := strings.Join(textParts, " ")

	// 清理合并后的文本
	mergedText = strings.Join(strings.Fields(mergedText), " ")

	// 计算位置信息
	position := firstPosition
	if hasPosition {
		position.EndLine = lastPosition.EndLine
		position.EndCol = lastPosition.EndCol
	}

	// 计算哈希
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(mergedContent)))

	// 创建元数据
	metadata := make(map[string]string)

	// 如果不是虚拟块，复制原始元数据
	if !s.isVirtualChunk(hChunk) {
		maps.Copy(metadata, hChunk.Chunk.Metadata)
	}

	metadata["strategy"] = "hierarchical"
	metadata["merged_chunks"] = fmt.Sprintf("%d", s.countTotalChunks(hChunk))
	metadata["hierarchy_level"] = fmt.Sprintf("%d", hChunk.Level)
	metadata["content_length"] = fmt.Sprintf("%d", len(mergedContent))
	metadata["text_length"] = fmt.Sprintf("%d", len(mergedText))
	metadata["word_count"] = fmt.Sprintf("%d", len(strings.Fields(mergedText)))

	// 确定块类型和层级
	chunkType := hChunk.Chunk.Type
	level := hChunk.Chunk.Level

	// 处理虚拟块
	if s.isVirtualChunk(hChunk) {
		if hChunk.Chunk.Type == "root" {
			chunkType = "document"
			level = 0
		} else {
			chunkType = "hierarchical"
			level = 0
		}
	} else if chunkType == "heading" {
		metadata["original_heading_level"] = fmt.Sprintf("%d", level)
		metadata["heading_level"] = fmt.Sprintf("%d", level)
	} else if len(hChunk.Children) > 0 {
		// 对于包含子内容的非标题块，使用混合类型
		chunkType = "hierarchical"
		level = 0
	}

	// 添加子块类型统计
	childTypes := make(map[string]int)
	var countChildTypes func(*HierarchicalChunk)
	countChildTypes = func(chunk *HierarchicalChunk) {
		if chunk == nil || s.isVirtualChunk(chunk) {
			return
		}
		childTypes[chunk.Chunk.Type]++
		for _, child := range chunk.Children {
			countChildTypes(child)
		}
	}

	for _, child := range hChunk.Children {
		countChildTypes(child)
	}

	if len(childTypes) > 0 {
		var typeList []string
		for chunkType, count := range childTypes {
			typeList = append(typeList, fmt.Sprintf("%s:%d", chunkType, count))
		}
		metadata["child_types"] = strings.Join(typeList, ",")
	}

	return &Chunk{
		ID:       id,
		Type:     chunkType,
		Content:  mergedContent,
		Text:     mergedText,
		Level:    level,
		Metadata: metadata,
		Position: position,
		Links:    allLinks,
		Images:   allImages,
		Hash:     hash,
	}
}

// countTotalChunks 计算层级块中包含的总块数
func (s *HierarchicalStrategy) countTotalChunks(hChunk *HierarchicalChunk) int {
	if hChunk == nil {
		return 0
	}

	count := 1 // 当前块
	for _, child := range hChunk.Children {
		count += s.countTotalChunks(child)
	}
	return count
}

// GetConfig 获取策略配置
func (s *HierarchicalStrategy) GetConfig() *StrategyConfig {
	if s.config == nil {
		return nil
	}
	return s.config.Clone()
}

// SetConfig 设置策略配置
func (s *HierarchicalStrategy) SetConfig(config *StrategyConfig) error {
	if config == nil {
		s.config = HierarchicalConfig(0)
		return nil
	}

	if err := s.ValidateConfig(config); err != nil {
		return err
	}

	s.config = config.Clone()
	return nil
}

// getMaxDepthInHierarchy 获取层级结构中的最大深度
func (s *HierarchicalStrategy) getMaxDepthInHierarchy(chunks []*HierarchicalChunk) int {
	maxDepth := 0

	var traverse func([]*HierarchicalChunk, int)
	traverse = func(chunks []*HierarchicalChunk, currentDepth int) {
		for _, chunk := range chunks {
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
			traverse(chunk.Children, currentDepth+1)
		}
	}

	traverse(chunks, 0)
	return maxDepth
}

// ChunkingContext 分块上下文，提供分块过程中的状态信息
type ChunkingContext struct {
	CurrentChunk   *Chunk           // 当前正在处理的块
	PreviousChunk  *Chunk           // 前一个块
	ParentNode     ast.Node         // 父节点
	Depth          int              // 当前深度
	ChunkCount     int              // 已处理的块数量
	TotalNodes     int              // 总节点数
	Source         []byte           // 源文档内容
	Chunker        *MarkdownChunker // 分块器实例
	CustomData     map[string]any   // 自定义数据
	ProcessingTime time.Duration    // 处理时间
}

// NewChunkingContext 创建新的分块上下文
func NewChunkingContext(chunker *MarkdownChunker, source []byte) *ChunkingContext {
	return &ChunkingContext{
		Chunker:    chunker,
		Source:     source,
		CustomData: make(map[string]any),
	}
}

// Clone 创建上下文的副本
func (ctx *ChunkingContext) Clone() *ChunkingContext {
	if ctx == nil {
		return nil
	}

	clone := &ChunkingContext{
		CurrentChunk:   ctx.CurrentChunk,
		PreviousChunk:  ctx.PreviousChunk,
		ParentNode:     ctx.ParentNode,
		Depth:          ctx.Depth,
		ChunkCount:     ctx.ChunkCount,
		TotalNodes:     ctx.TotalNodes,
		Source:         ctx.Source,
		Chunker:        ctx.Chunker,
		ProcessingTime: ctx.ProcessingTime,
	}

	// 深拷贝自定义数据
	if ctx.CustomData != nil {
		clone.CustomData = make(map[string]any)
		maps.Copy(clone.CustomData, ctx.CustomData)
	}

	return clone
}

// SetCustomData 设置自定义数据
func (ctx *ChunkingContext) SetCustomData(key string, value any) {
	if ctx.CustomData == nil {
		ctx.CustomData = make(map[string]any)
	}
	ctx.CustomData[key] = value
}

// GetCustomData 获取自定义数据
func (ctx *ChunkingContext) GetCustomData(key string) (any, bool) {
	if ctx.CustomData == nil {
		return nil, false
	}
	value, exists := ctx.CustomData[key]
	return value, exists
}

// RuleCondition 规则条件接口
// 定义分块规则的匹配条件
type RuleCondition interface {
	// Match 检查节点是否匹配条件
	Match(node ast.Node, context *ChunkingContext) bool

	// GetName 返回条件名称
	GetName() string

	// GetDescription 返回条件描述
	GetDescription() string

	// Validate 验证条件配置是否有效
	Validate() error

	// Clone 创建条件的副本
	Clone() RuleCondition
}

// RuleAction 规则动作接口
// 定义匹配条件后执行的动作
type RuleAction interface {
	// Execute 执行动作，返回处理后的块
	Execute(node ast.Node, context *ChunkingContext) (*Chunk, error)

	// GetName 返回动作名称
	GetName() string

	// GetDescription 返回动作描述
	GetDescription() string

	// Validate 验证动作配置是否有效
	Validate() error

	// Clone 创建动作的副本
	Clone() RuleAction
}

// ChunkingRule 分块规则
// 将条件和动作组合成完整的规则
type ChunkingRule struct {
	Name        string        `json:"name"`        // 规则名称
	Description string        `json:"description"` // 规则描述
	Condition   RuleCondition `json:"-"`           // 规则条件（不序列化）
	Action      RuleAction    `json:"-"`           // 规则动作（不序列化）
	Priority    int           `json:"priority"`    // 规则优先级（数值越大优先级越高）
	Enabled     bool          `json:"enabled"`     // 是否启用
}

// NewChunkingRule 创建新的分块规则
func NewChunkingRule(name, description string, condition RuleCondition, action RuleAction, priority int) *ChunkingRule {
	return &ChunkingRule{
		Name:        name,
		Description: description,
		Condition:   condition,
		Action:      action,
		Priority:    priority,
		Enabled:     true,
	}
}

// Match 检查规则是否匹配
func (r *ChunkingRule) Match(node ast.Node, context *ChunkingContext) bool {
	if !r.Enabled || r.Condition == nil {
		return false
	}
	return r.Condition.Match(node, context)
}

// Execute 执行规则动作
func (r *ChunkingRule) Execute(node ast.Node, context *ChunkingContext) (*Chunk, error) {
	if !r.Enabled || r.Action == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "规则未启用或动作为空", nil).
			WithContext("rule_name", r.Name)
	}
	return r.Action.Execute(node, context)
}

// Validate 验证规则配置
func (r *ChunkingRule) Validate() error {
	if r.Name == "" {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则名称不能为空", nil)
	}

	if r.Condition == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则条件不能为空", nil).
			WithContext("rule_name", r.Name)
	}

	if r.Action == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则动作不能为空", nil).
			WithContext("rule_name", r.Name)
	}

	if err := r.Condition.Validate(); err != nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则条件验证失败", err).
			WithContext("rule_name", r.Name)
	}

	if err := r.Action.Validate(); err != nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则动作验证失败", err).
			WithContext("rule_name", r.Name)
	}

	return nil
}

// Clone 创建规则的副本
func (r *ChunkingRule) Clone() *ChunkingRule {
	if r == nil {
		return nil
	}

	var conditionClone RuleCondition
	if r.Condition != nil {
		conditionClone = r.Condition.Clone()
	}

	var actionClone RuleAction
	if r.Action != nil {
		actionClone = r.Action.Clone()
	}

	return &ChunkingRule{
		Name:        r.Name,
		Description: r.Description,
		Condition:   conditionClone,
		Action:      actionClone,
		Priority:    r.Priority,
		Enabled:     r.Enabled,
	}
}

// String 返回规则的字符串表示
func (r *ChunkingRule) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ChunkingRule{Name: %s, Priority: %d, Enabled: %t}", r.Name, r.Priority, r.Enabled)
}

// 预定义条件实现

// HeadingLevelCondition 标题层级条件
// 匹配指定层级范围内的标题
type HeadingLevelCondition struct {
	MinLevel int `json:"min_level"` // 最小层级（包含）
	MaxLevel int `json:"max_level"` // 最大层级（包含）
}

// NewHeadingLevelCondition 创建标题层级条件
func NewHeadingLevelCondition(minLevel, maxLevel int) *HeadingLevelCondition {
	return &HeadingLevelCondition{
		MinLevel: minLevel,
		MaxLevel: maxLevel,
	}
}

// Match 检查节点是否匹配标题层级条件
func (c *HeadingLevelCondition) Match(node ast.Node, context *ChunkingContext) bool {
	if node == nil {
		return false
	}

	heading, ok := node.(*ast.Heading)
	if !ok {
		return false
	}

	level := heading.Level
	return level >= c.MinLevel && level <= c.MaxLevel
}

// GetName 返回条件名称
func (c *HeadingLevelCondition) GetName() string {
	return "heading-level"
}

// GetDescription 返回条件描述
func (c *HeadingLevelCondition) GetDescription() string {
	return fmt.Sprintf("匹配层级在 %d-%d 之间的标题", c.MinLevel, c.MaxLevel)
}

// Validate 验证条件配置
func (c *HeadingLevelCondition) Validate() error {
	if c.MinLevel < 1 || c.MinLevel > 6 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小标题层级必须在1-6之间", nil).
			WithContext("min_level", c.MinLevel)
	}

	if c.MaxLevel < 1 || c.MaxLevel > 6 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大标题层级必须在1-6之间", nil).
			WithContext("max_level", c.MaxLevel)
	}

	if c.MinLevel > c.MaxLevel {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小层级不能大于最大层级", nil).
			WithContext("min_level", c.MinLevel).
			WithContext("max_level", c.MaxLevel)
	}

	return nil
}

// Clone 创建条件的副本
func (c *HeadingLevelCondition) Clone() RuleCondition {
	if c == nil {
		return nil
	}
	return &HeadingLevelCondition{
		MinLevel: c.MinLevel,
		MaxLevel: c.MaxLevel,
	}
}

// ContentTypeCondition 内容类型条件
// 匹配指定类型的内容节点
type ContentTypeCondition struct {
	Types []string `json:"types"` // 允许的内容类型列表
}

// NewContentTypeCondition 创建内容类型条件
func NewContentTypeCondition(types ...string) *ContentTypeCondition {
	return &ContentTypeCondition{
		Types: types,
	}
}

// Match 检查节点是否匹配内容类型条件
func (c *ContentTypeCondition) Match(node ast.Node, context *ChunkingContext) bool {
	if node == nil || len(c.Types) == 0 {
		return false
	}

	nodeType := getNodeType(node)
	return slices.Contains(c.Types, nodeType)
}

// GetName 返回条件名称
func (c *ContentTypeCondition) GetName() string {
	return "content-type"
}

// GetDescription 返回条件描述
func (c *ContentTypeCondition) GetDescription() string {
	return fmt.Sprintf("匹配类型为 %v 的内容", c.Types)
}

// Validate 验证条件配置
func (c *ContentTypeCondition) Validate() error {
	if len(c.Types) == 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "内容类型列表不能为空", nil)
	}

	validTypes := map[string]bool{
		"heading": true, "paragraph": true, "code": true,
		"table": true, "list": true, "blockquote": true,
		"thematic_break": true, "text": true, "emphasis": true,
		"link": true, "image": true,
	}

	for _, contentType := range c.Types {
		if !validTypes[contentType] {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "无效的内容类型", nil).
				WithContext("invalid_type", contentType)
		}
	}

	return nil
}

// Clone 创建条件的副本
func (c *ContentTypeCondition) Clone() RuleCondition {
	if c == nil {
		return nil
	}

	clone := &ContentTypeCondition{
		Types: make([]string, len(c.Types)),
	}
	copy(clone.Types, c.Types)
	return clone
}

// ContentSizeCondition 内容大小条件
// 匹配指定大小范围内的内容
type ContentSizeCondition struct {
	MinSize int `json:"min_size"` // 最小内容大小（字符数）
	MaxSize int `json:"max_size"` // 最大内容大小（字符数，0表示无限制）
}

// NewContentSizeCondition 创建内容大小条件
func NewContentSizeCondition(minSize, maxSize int) *ContentSizeCondition {
	return &ContentSizeCondition{
		MinSize: minSize,
		MaxSize: maxSize,
	}
}

// Match 检查节点是否匹配内容大小条件
func (c *ContentSizeCondition) Match(node ast.Node, context *ChunkingContext) bool {
	if node == nil || context == nil || context.Source == nil {
		return false
	}

	// 提取节点的文本内容
	var buf bytes.Buffer
	extractNodeText(node, context.Source, &buf)
	contentSize := buf.Len()

	// 检查大小范围
	if contentSize < c.MinSize {
		return false
	}

	if c.MaxSize > 0 && contentSize > c.MaxSize {
		return false
	}

	return true
}

// GetName 返回条件名称
func (c *ContentSizeCondition) GetName() string {
	return "content-size"
}

// GetDescription 返回条件描述
func (c *ContentSizeCondition) GetDescription() string {
	if c.MaxSize > 0 {
		return fmt.Sprintf("匹配内容大小在 %d-%d 字符之间的内容", c.MinSize, c.MaxSize)
	}
	return fmt.Sprintf("匹配内容大小大于等于 %d 字符的内容", c.MinSize)
}

// Validate 验证条件配置
func (c *ContentSizeCondition) Validate() error {
	if c.MinSize < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小内容大小不能为负数", nil).
			WithContext("min_size", c.MinSize)
	}

	if c.MaxSize < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大内容大小不能为负数", nil).
			WithContext("max_size", c.MaxSize)
	}

	if c.MaxSize > 0 && c.MinSize > c.MaxSize {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小大小不能大于最大大小", nil).
			WithContext("min_size", c.MinSize).
			WithContext("max_size", c.MaxSize)
	}

	return nil
}

// Clone 创建条件的副本
func (c *ContentSizeCondition) Clone() RuleCondition {
	if c == nil {
		return nil
	}
	return &ContentSizeCondition{
		MinSize: c.MinSize,
		MaxSize: c.MaxSize,
	}
}

// DepthCondition 深度条件
// 匹配指定深度范围内的节点
type DepthCondition struct {
	MinDepth int `json:"min_depth"` // 最小深度
	MaxDepth int `json:"max_depth"` // 最大深度（0表示无限制）
}

// NewDepthCondition 创建深度条件
func NewDepthCondition(minDepth, maxDepth int) *DepthCondition {
	return &DepthCondition{
		MinDepth: minDepth,
		MaxDepth: maxDepth,
	}
}

// Match 检查节点是否匹配深度条件
func (c *DepthCondition) Match(node ast.Node, context *ChunkingContext) bool {
	if node == nil || context == nil {
		return false
	}

	depth := context.Depth
	if depth < c.MinDepth {
		return false
	}

	if c.MaxDepth > 0 && depth > c.MaxDepth {
		return false
	}

	return true
}

// GetName 返回条件名称
func (c *DepthCondition) GetName() string {
	return "depth"
}

// GetDescription 返回条件描述
func (c *DepthCondition) GetDescription() string {
	if c.MaxDepth > 0 {
		return fmt.Sprintf("匹配深度在 %d-%d 之间的节点", c.MinDepth, c.MaxDepth)
	}
	return fmt.Sprintf("匹配深度大于等于 %d 的节点", c.MinDepth)
}

// Validate 验证条件配置
func (c *DepthCondition) Validate() error {
	if c.MinDepth < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小深度不能为负数", nil).
			WithContext("min_depth", c.MinDepth)
	}

	if c.MaxDepth < 0 {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最大深度不能为负数", nil).
			WithContext("max_depth", c.MaxDepth)
	}

	if c.MaxDepth > 0 && c.MinDepth > c.MaxDepth {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "最小深度不能大于最大深度", nil).
			WithContext("min_depth", c.MinDepth).
			WithContext("max_depth", c.MaxDepth)
	}

	return nil
}

// Clone 创建条件的副本
func (c *DepthCondition) Clone() RuleCondition {
	if c == nil {
		return nil
	}
	return &DepthCondition{
		MinDepth: c.MinDepth,
		MaxDepth: c.MaxDepth,
	}
}

// 预定义动作实现

// CreateSeparateChunkAction 创建独立块动作
// 将匹配的节点创建为独立的块
type CreateSeparateChunkAction struct {
	ChunkType string            `json:"chunk_type"` // 块类型（可选，为空时使用节点类型）
	Metadata  map[string]string `json:"metadata"`   // 附加元数据
}

// NewCreateSeparateChunkAction 创建独立块动作
func NewCreateSeparateChunkAction(chunkType string, metadata map[string]string) *CreateSeparateChunkAction {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	return &CreateSeparateChunkAction{
		ChunkType: chunkType,
		Metadata:  metadata,
	}
}

// Execute 执行创建独立块动作
func (a *CreateSeparateChunkAction) Execute(node ast.Node, context *ChunkingContext) (*Chunk, error) {
	if node == nil || context == nil || context.Chunker == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "节点、上下文或分块器不能为空", nil).
			WithContext("action", a.GetName())
	}

	// 确保分块器有源内容
	originalSource := context.Chunker.source
	if context.Source != nil {
		context.Chunker.source = context.Source
	}
	defer func() {
		context.Chunker.source = originalSource
	}()

	// 使用分块器的 processNode 方法创建块
	chunk := context.Chunker.processNode(node, context.ChunkCount)
	if chunk == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "无法创建块", nil).
			WithContext("action", a.GetName()).
			WithContext("node_type", getNodeType(node))
	}

	// 应用自定义块类型
	if a.ChunkType != "" {
		chunk.Type = a.ChunkType
	}

	// 添加自定义元数据
	if chunk.Metadata == nil {
		chunk.Metadata = make(map[string]string)
	}
	maps.Copy(chunk.Metadata, a.Metadata)

	// 添加动作标识
	chunk.Metadata["action"] = a.GetName()
	chunk.Metadata["custom_rule"] = "true"

	return chunk, nil
}

// GetName 返回动作名称
func (a *CreateSeparateChunkAction) GetName() string {
	return "create-separate-chunk"
}

// GetDescription 返回动作描述
func (a *CreateSeparateChunkAction) GetDescription() string {
	return "将匹配的节点创建为独立的块"
}

// Validate 验证动作配置
func (a *CreateSeparateChunkAction) Validate() error {
	// 创建独立块动作没有特殊的验证要求
	return nil
}

// Clone 创建动作的副本
func (a *CreateSeparateChunkAction) Clone() RuleAction {
	if a == nil {
		return nil
	}

	clone := &CreateSeparateChunkAction{
		ChunkType: a.ChunkType,
	}

	// 深拷贝元数据
	if a.Metadata != nil {
		clone.Metadata = make(map[string]string)
		maps.Copy(clone.Metadata, a.Metadata)
	}

	return clone
}

// MergeWithParentAction 与父块合并动作
// 将匹配的节点合并到父块中
type MergeWithParentAction struct {
	Separator string `json:"separator"` // 合并时使用的分隔符
}

// NewMergeWithParentAction 创建与父块合并动作
func NewMergeWithParentAction(separator string) *MergeWithParentAction {
	if separator == "" {
		separator = "\n\n" // 默认分隔符
	}
	return &MergeWithParentAction{
		Separator: separator,
	}
}

// Execute 执行与父块合并动作
func (a *MergeWithParentAction) Execute(node ast.Node, context *ChunkingContext) (*Chunk, error) {
	if node == nil || context == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "节点或上下文不能为空", nil).
			WithContext("action", a.GetName())
	}

	// 如果没有前一个块，创建新块
	if context.PreviousChunk == nil {
		return a.createNewChunk(node, context)
	}

	// 提取当前节点的内容
	var contentBuf, textBuf bytes.Buffer
	extractNodeContent(node, context.Source, &contentBuf)
	extractNodeText(node, context.Source, &textBuf)

	currentContent := contentBuf.String()
	currentText := textBuf.String()

	if strings.TrimSpace(currentContent) == "" && strings.TrimSpace(currentText) == "" {
		// 空内容，不进行合并
		return nil, nil
	}

	// 创建合并后的块
	mergedChunk := &Chunk{
		ID:       context.PreviousChunk.ID,
		Type:     context.PreviousChunk.Type,
		Content:  context.PreviousChunk.Content + a.Separator + currentContent,
		Text:     context.PreviousChunk.Text + " " + currentText,
		Level:    context.PreviousChunk.Level,
		Metadata: make(map[string]string),
		Position: context.PreviousChunk.Position,
		Links:    append(context.PreviousChunk.Links, extractLinks(node, context.Source)...),
		Images:   append(context.PreviousChunk.Images, extractImages(node, context.Source)...),
	}

	// 复制原有元数据
	maps.Copy(mergedChunk.Metadata, context.PreviousChunk.Metadata)

	// 添加合并标识
	mergedChunk.Metadata["action"] = a.GetName()
	mergedChunk.Metadata["merged"] = "true"
	mergedChunk.Metadata["merged_node_type"] = getNodeType(node)

	// 更新位置信息
	nodePos := getNodePosition(node, context.Source)
	if nodePos.EndLine > mergedChunk.Position.EndLine {
		mergedChunk.Position.EndLine = nodePos.EndLine
		mergedChunk.Position.EndCol = nodePos.EndCol
	}

	// 重新计算哈希
	mergedChunk.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(mergedChunk.Content)))

	return mergedChunk, nil
}

// createNewChunk 创建新块（当没有父块时）
func (a *MergeWithParentAction) createNewChunk(node ast.Node, context *ChunkingContext) (*Chunk, error) {
	if context.Chunker == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "分块器不能为空", nil).
			WithContext("action", a.GetName())
	}

	// 确保分块器有源内容
	originalSource := context.Chunker.source
	if context.Source != nil {
		context.Chunker.source = context.Source
	}
	defer func() {
		context.Chunker.source = originalSource
	}()

	chunk := context.Chunker.processNode(node, context.ChunkCount)
	if chunk == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "无法创建块", nil).
			WithContext("action", a.GetName()).
			WithContext("node_type", getNodeType(node))
	}

	// 添加动作标识
	if chunk.Metadata == nil {
		chunk.Metadata = make(map[string]string)
	}
	chunk.Metadata["action"] = a.GetName()
	chunk.Metadata["fallback_to_new_chunk"] = "true"

	return chunk, nil
}

// GetName 返回动作名称
func (a *MergeWithParentAction) GetName() string {
	return "merge-with-parent"
}

// GetDescription 返回动作描述
func (a *MergeWithParentAction) GetDescription() string {
	return "将匹配的节点合并到父块中"
}

// Validate 验证动作配置
func (a *MergeWithParentAction) Validate() error {
	// 合并动作没有特殊的验证要求
	return nil
}

// Clone 创建动作的副本
func (a *MergeWithParentAction) Clone() RuleAction {
	if a == nil {
		return nil
	}
	return &MergeWithParentAction{
		Separator: a.Separator,
	}
}

// SkipNodeAction 跳过节点动作
// 跳过匹配的节点，不创建块
type SkipNodeAction struct {
	Reason string `json:"reason"` // 跳过原因
}

// NewSkipNodeAction 创建跳过节点动作
func NewSkipNodeAction(reason string) *SkipNodeAction {
	return &SkipNodeAction{
		Reason: reason,
	}
}

// Execute 执行跳过节点动作
func (a *SkipNodeAction) Execute(node ast.Node, context *ChunkingContext) (*Chunk, error) {
	// 跳过节点，返回nil表示不创建块
	return nil, nil
}

// GetName 返回动作名称
func (a *SkipNodeAction) GetName() string {
	return "skip-node"
}

// GetDescription 返回动作描述
func (a *SkipNodeAction) GetDescription() string {
	if a.Reason != "" {
		return fmt.Sprintf("跳过匹配的节点：%s", a.Reason)
	}
	return "跳过匹配的节点"
}

// Validate 验证动作配置
func (a *SkipNodeAction) Validate() error {
	// 跳过动作没有特殊的验证要求
	return nil
}

// Clone 创建动作的副本
func (a *SkipNodeAction) Clone() RuleAction {
	if a == nil {
		return nil
	}
	return &SkipNodeAction{
		Reason: a.Reason,
	}
}

// 辅助函数

// getNodeType 获取节点类型字符串
func getNodeType(node ast.Node) string {
	if node == nil {
		return "unknown"
	}

	switch node.(type) {
	case *ast.Heading:
		return "heading"
	case *ast.Paragraph:
		return "paragraph"
	case *ast.CodeBlock, *ast.FencedCodeBlock:
		return "code"
	case *ast.List:
		return "list"
	case *ast.Blockquote:
		return "blockquote"
	case *ast.ThematicBreak:
		return "thematic_break"
	case *ast.Text:
		return "text"
	case *ast.Emphasis:
		return "emphasis"
	case *ast.Link:
		return "link"
	case *ast.Image:
		return "image"
	case *extast.Table:
		return "table"
	default:
		return "unknown"
	}
}

// extractNodeContent 提取节点的原始内容
func extractNodeContent(node ast.Node, source []byte, buf *bytes.Buffer) {
	if node == nil || source == nil || buf == nil {
		return
	}

	segment := node.Lines()
	if segment.Len() == 0 {
		return
	}

	for i := 0; i < segment.Len(); i++ {
		line := segment.At(i)
		buf.Write(line.Value(source))
	}
}

// extractNodeText 提取节点的纯文本内容
func extractNodeText(node ast.Node, source []byte, buf *bytes.Buffer) {
	if node == nil || source == nil || buf == nil {
		return
	}

	// 递归提取文本内容
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if textNode, ok := child.(*ast.Text); ok {
			segment := textNode.Segment
			buf.Write(segment.Value(source))
		} else {
			extractNodeText(child, source, buf)
		}
	}
}

// getNodePosition 获取节点在文档中的位置
func getNodePosition(node ast.Node, source []byte) ChunkPosition {
	if node == nil || source == nil {
		return ChunkPosition{}
	}

	segment := node.Lines()
	if segment.Len() == 0 {
		return ChunkPosition{}
	}

	firstLine := segment.At(0)
	lastLine := segment.At(segment.Len() - 1)

	// 计算行号（从1开始）
	startLine := 1
	endLine := 1
	startCol := 1
	endCol := 1

	// 计算起始位置
	for i := 0; i < firstLine.Start; i++ {
		if source[i] == '\n' {
			startLine++
			startCol = 1
		} else {
			startCol++
		}
	}

	// 计算结束位置
	endLine = startLine
	endCol = startCol
	for i := firstLine.Start; i < lastLine.Stop; i++ {
		if source[i] == '\n' {
			endLine++
			endCol = 1
		} else {
			endCol++
		}
	}

	return ChunkPosition{
		StartLine: startLine,
		EndLine:   endLine,
		StartCol:  startCol,
		EndCol:    endCol,
	}
}

// extractLinks 提取节点中的链接
func extractLinks(node ast.Node, source []byte) []Link {
	var links []Link

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if link, ok := n.(*ast.Link); ok {
			linkText := ""
			if link.FirstChild() != nil {
				if textNode, ok := link.FirstChild().(*ast.Text); ok {
					linkText = string(textNode.Segment.Value(source))
				}
			}

			linkURL := string(link.Destination)
			linkType := "external"
			if strings.HasPrefix(linkURL, "#") {
				linkType = "anchor"
			} else if !strings.Contains(linkURL, "://") {
				linkType = "internal"
			}

			links = append(links, Link{
				Text: linkText,
				URL:  linkURL,
				Type: linkType,
			})
		}

		return ast.WalkContinue, nil
	})

	return links
}

// extractImages 提取节点中的图片
func extractImages(node ast.Node, source []byte) []Image {
	var images []Image

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if img, ok := n.(*ast.Image); ok {
			imgAlt := ""
			if img.FirstChild() != nil {
				if textNode, ok := img.FirstChild().(*ast.Text); ok {
					imgAlt = string(textNode.Segment.Value(source))
				}
			}

			imgURL := string(img.Destination)
			imgTitle := string(img.Title)

			images = append(images, Image{
				Alt:   imgAlt,
				URL:   imgURL,
				Title: imgTitle,
			})
		}

		return ast.WalkContinue, nil
	})

	return images
}

// CustomStrategyBuilder 自定义策略构建器
// 用于构建基于规则的自定义分块策略
type CustomStrategyBuilder struct {
	Name        string          `json:"name"`        // 策略名称
	Description string          `json:"description"` // 策略描述
	Rules       []*ChunkingRule `json:"rules"`       // 规则列表
	Config      *StrategyConfig `json:"config"`      // 策略配置
	mutex       sync.RWMutex    `json:"-"`           // 读写锁
}

// NewCustomStrategyBuilder 创建新的自定义策略构建器
func NewCustomStrategyBuilder(name, description string) *CustomStrategyBuilder {
	return &CustomStrategyBuilder{
		Name:        name,
		Description: description,
		Rules:       make([]*ChunkingRule, 0),
		Config:      DefaultStrategyConfig(name),
	}
}

// SetName 设置策略名称
func (b *CustomStrategyBuilder) SetName(name string) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Name = name
	if b.Config != nil {
		b.Config.Name = name
	}
	return b
}

// SetDescription 设置策略描述
func (b *CustomStrategyBuilder) SetDescription(description string) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Description = description
	return b
}

// SetConfig 设置策略配置
func (b *CustomStrategyBuilder) SetConfig(config *StrategyConfig) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if config != nil {
		b.Config = config.Clone()
		b.Config.Name = b.Name // 确保名称一致
	}
	return b
}

// AddRule 添加分块规则
func (b *CustomStrategyBuilder) AddRule(name, description string, condition RuleCondition, action RuleAction, priority int) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	rule := NewChunkingRule(name, description, condition, action, priority)
	b.Rules = append(b.Rules, rule)

	// 按优先级排序（优先级高的在前）
	b.sortRulesByPriority()

	return b
}

// AddRuleObject 添加规则对象
func (b *CustomStrategyBuilder) AddRuleObject(rule *ChunkingRule) *CustomStrategyBuilder {
	if rule == nil {
		return b
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.Rules = append(b.Rules, rule.Clone())
	b.sortRulesByPriority()

	return b
}

// RemoveRule 移除指定名称的规则
func (b *CustomStrategyBuilder) RemoveRule(name string) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for i, rule := range b.Rules {
		if rule.Name == name {
			// 删除规则
			b.Rules = append(b.Rules[:i], b.Rules[i+1:]...)
			break
		}
	}

	return b
}

// EnableRule 启用指定名称的规则
func (b *CustomStrategyBuilder) EnableRule(name string) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, rule := range b.Rules {
		if rule.Name == name {
			rule.Enabled = true
			break
		}
	}

	return b
}

// DisableRule 禁用指定名称的规则
func (b *CustomStrategyBuilder) DisableRule(name string) *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, rule := range b.Rules {
		if rule.Name == name {
			rule.Enabled = false
			break
		}
	}

	return b
}

// GetRules 获取所有规则
func (b *CustomStrategyBuilder) GetRules() []*ChunkingRule {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	rules := make([]*ChunkingRule, len(b.Rules))
	for i, rule := range b.Rules {
		rules[i] = rule.Clone()
	}
	return rules
}

// GetRule 获取指定名称的规则
func (b *CustomStrategyBuilder) GetRule(name string) *ChunkingRule {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	for _, rule := range b.Rules {
		if rule.Name == name {
			return rule.Clone()
		}
	}
	return nil
}

// HasRule 检查是否存在指定名称的规则
func (b *CustomStrategyBuilder) HasRule(name string) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	for _, rule := range b.Rules {
		if rule.Name == name {
			return true
		}
	}
	return false
}

// GetRuleCount 获取规则数量
func (b *CustomStrategyBuilder) GetRuleCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.Rules)
}

// ClearRules 清空所有规则
func (b *CustomStrategyBuilder) ClearRules() *CustomStrategyBuilder {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Rules = make([]*ChunkingRule, 0)
	return b
}

// sortRulesByPriority 按优先级排序规则（优先级高的在前）
func (b *CustomStrategyBuilder) sortRulesByPriority() {
	// 使用稳定排序保持相同优先级规则的相对顺序
	for i := 1; i < len(b.Rules); i++ {
		key := b.Rules[i]
		j := i - 1
		for j >= 0 && b.Rules[j].Priority < key.Priority {
			b.Rules[j+1] = b.Rules[j]
			j--
		}
		b.Rules[j+1] = key
	}
}

// Validate 验证策略构建器配置
func (b *CustomStrategyBuilder) Validate() error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if b.Name == "" {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略名称不能为空", nil).
			WithContext("builder", "CustomStrategyBuilder")
	}

	if b.Config != nil {
		if err := b.Config.ValidateConfig(); err != nil {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置验证失败", err).
				WithContext("builder", "CustomStrategyBuilder").
				WithContext("strategy_name", b.Name)
		}
	}

	// 验证所有规则
	for i, rule := range b.Rules {
		if rule == nil {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则不能为空", nil).
				WithContext("builder", "CustomStrategyBuilder").
				WithContext("strategy_name", b.Name).
				WithContext("rule_index", i)
		}

		if err := rule.Validate(); err != nil {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则验证失败", err).
				WithContext("builder", "CustomStrategyBuilder").
				WithContext("strategy_name", b.Name).
				WithContext("rule_name", rule.Name).
				WithContext("rule_index", i)
		}
	}

	// 检查规则名称唯一性
	ruleNames := make(map[string]bool)
	for i, rule := range b.Rules {
		if ruleNames[rule.Name] {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则名称重复", nil).
				WithContext("builder", "CustomStrategyBuilder").
				WithContext("strategy_name", b.Name).
				WithContext("duplicate_rule_name", rule.Name).
				WithContext("rule_index", i)
		}
		ruleNames[rule.Name] = true
	}

	// 检查是否至少有一个启用的规则
	hasEnabledRule := false
	for _, rule := range b.Rules {
		if rule.Enabled {
			hasEnabledRule = true
			break
		}
	}

	if !hasEnabledRule {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "至少需要一个启用的规则", nil).
			WithContext("builder", "CustomStrategyBuilder").
			WithContext("strategy_name", b.Name)
	}

	return nil
}

// Build 构建自定义策略
func (b *CustomStrategyBuilder) Build() (ChunkingStrategy, error) {
	// 验证配置
	if err := b.Validate(); err != nil {
		return nil, err
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// 创建自定义策略实例
	strategy := &CustomStrategy{
		Name:        b.Name,
		Description: b.Description,
		Rules:       make([]*ChunkingRule, len(b.Rules)),
		Config:      b.Config.Clone(),
	}

	// 深拷贝规则
	for i, rule := range b.Rules {
		strategy.Rules[i] = rule.Clone()
	}

	return strategy, nil
}

// Clone 创建构建器的副本
func (b *CustomStrategyBuilder) Clone() *CustomStrategyBuilder {
	if b == nil {
		return nil
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	clone := &CustomStrategyBuilder{
		Name:        b.Name,
		Description: b.Description,
		Rules:       make([]*ChunkingRule, len(b.Rules)),
	}

	// 深拷贝规则
	for i, rule := range b.Rules {
		clone.Rules[i] = rule.Clone()
	}

	// 深拷贝配置
	if b.Config != nil {
		clone.Config = b.Config.Clone()
	}

	return clone
}

// String 返回构建器的字符串表示
func (b *CustomStrategyBuilder) String() string {
	if b == nil {
		return "<nil>"
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return fmt.Sprintf("CustomStrategyBuilder{Name: %s, Rules: %d, Description: %s}",
		b.Name, len(b.Rules), b.Description)
}

// CustomStrategy 自定义分块策略
// 基于规则的自定义分块策略实现
type CustomStrategy struct {
	Name        string          `json:"name"`        // 策略名称
	Description string          `json:"description"` // 策略描述
	Rules       []*ChunkingRule `json:"rules"`       // 规则列表（按优先级排序）
	Config      *StrategyConfig `json:"config"`      // 策略配置
	mutex       sync.RWMutex    `json:"-"`           // 读写锁
}

// GetName 返回策略名称
func (s *CustomStrategy) GetName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Name
}

// GetDescription 返回策略描述
func (s *CustomStrategy) GetDescription() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Description
}

// ChunkDocument 使用自定义策略对文档进行分块
func (s *CustomStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	if doc == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "文档节点不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	if chunker == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "分块器实例不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	s.mutex.RLock()
	rules := make([]*ChunkingRule, len(s.Rules))
	for i, rule := range s.Rules {
		rules[i] = rule.Clone()
	}
	s.mutex.RUnlock()

	// 创建分块上下文
	context := NewChunkingContext(chunker, source)
	var chunks []Chunk
	chunkID := 0

	// 设置分块器的源内容
	originalSource := chunker.source
	chunker.source = source
	defer func() {
		chunker.source = originalSource
	}()

	// 遍历文档的所有顶层子节点
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		// 更新上下文
		context.ChunkCount = chunkID
		if len(chunks) > 0 {
			context.PreviousChunk = &chunks[len(chunks)-1]
		}

		// 处理节点
		processedChunks, err := s.processNodeWithRules(child, context, rules, 0)
		if err != nil {
			return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "处理节点失败", err).
				WithContext("strategy", s.GetName()).
				WithContext("function", "ChunkDocument").
				WithContext("node_type", getNodeType(child))
		}

		// 添加处理后的块
		for _, chunk := range processedChunks {
			chunk.ID = chunkID
			chunks = append(chunks, chunk)
			chunkID++
		}
	}

	return chunks, nil
}

// processNodeWithRules 使用规则处理节点
func (s *CustomStrategy) processNodeWithRules(node ast.Node, context *ChunkingContext, rules []*ChunkingRule, depth int) ([]Chunk, error) {
	if node == nil {
		return nil, nil
	}

	// 更新上下文深度
	context.Depth = depth

	// 查找匹配的规则（按优先级顺序）
	var matchedRule *ChunkingRule
	for _, rule := range rules {
		if rule.Enabled && rule.Match(node, context) {
			matchedRule = rule
			break
		}
	}

	var chunks []Chunk

	if matchedRule != nil {
		// 执行匹配的规则
		chunk, err := matchedRule.Execute(node, context)
		if err != nil {
			return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "执行规则失败", err).
				WithContext("strategy", s.GetName()).
				WithContext("rule_name", matchedRule.Name).
				WithContext("node_type", getNodeType(node))
		}

		if chunk != nil {
			chunks = append(chunks, *chunk)
		}
	} else {
		// 没有匹配的规则，使用默认处理
		chunk := context.Chunker.processNode(node, context.ChunkCount)
		if chunk != nil {
			// 添加默认处理标识
			if chunk.Metadata == nil {
				chunk.Metadata = make(map[string]string)
			}
			chunk.Metadata["processed_by"] = "default"
			chunk.Metadata["strategy"] = s.GetName()
			chunks = append(chunks, *chunk)
		}
	}

	// 不递归处理子节点，让分块器的 processNode 方法处理
	// 这与其他策略保持一致

	return chunks, nil
}

// ValidateConfig 验证策略特定的配置
func (s *CustomStrategy) ValidateConfig(config *StrategyConfig) error {
	if config == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ValidateConfig")
	}

	// 使用通用配置验证
	if err := config.ValidateConfig(); err != nil {
		return err
	}

	// 自定义策略特定验证
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 验证所有规则
	for i, rule := range s.Rules {
		if rule == nil {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则不能为空", nil).
				WithContext("strategy", s.GetName()).
				WithContext("function", "ValidateConfig").
				WithContext("rule_index", i)
		}

		if err := rule.Validate(); err != nil {
			return NewChunkerError(ErrorTypeStrategyConfigInvalid, "规则验证失败", err).
				WithContext("strategy", s.GetName()).
				WithContext("function", "ValidateConfig").
				WithContext("rule_name", rule.Name).
				WithContext("rule_index", i)
		}
	}

	return nil
}

// Clone 创建策略的副本（用于并发安全）
func (s *CustomStrategy) Clone() ChunkingStrategy {
	if s == nil {
		return nil
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clone := &CustomStrategy{
		Name:        s.Name,
		Description: s.Description,
		Rules:       make([]*ChunkingRule, len(s.Rules)),
	}

	// 深拷贝规则
	for i, rule := range s.Rules {
		clone.Rules[i] = rule.Clone()
	}

	// 深拷贝配置
	if s.Config != nil {
		clone.Config = s.Config.Clone()
	}

	return clone
}

// GetConfig 获取策略配置
func (s *CustomStrategy) GetConfig() *StrategyConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.Config == nil {
		return nil
	}
	return s.Config.Clone()
}

// SetConfig 设置策略配置
func (s *CustomStrategy) SetConfig(config *StrategyConfig) error {
	if config == nil {
		s.mutex.Lock()
		s.Config = DefaultStrategyConfig(s.Name)
		s.mutex.Unlock()
		return nil
	}

	if err := s.ValidateConfig(config); err != nil {
		return err
	}

	s.mutex.Lock()
	s.Config = config.Clone()
	s.Config.Name = s.Name // 确保名称一致
	s.mutex.Unlock()

	return nil
}

// GetRules 获取所有规则
func (s *CustomStrategy) GetRules() []*ChunkingRule {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rules := make([]*ChunkingRule, len(s.Rules))
	for i, rule := range s.Rules {
		rules[i] = rule.Clone()
	}
	return rules
}

// GetRuleCount 获取规则数量
func (s *CustomStrategy) GetRuleCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.Rules)
}

// GetEnabledRuleCount 获取启用的规则数量
func (s *CustomStrategy) GetEnabledRuleCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	for _, rule := range s.Rules {
		if rule.Enabled {
			count++
		}
	}
	return count
}

// String 返回策略的字符串表示
func (s *CustomStrategy) String() string {
	if s == nil {
		return "<nil>"
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return fmt.Sprintf("CustomStrategy{Name: %s, Rules: %d, EnabledRules: %d, Description: %s}",
		s.Name, len(s.Rules), s.GetEnabledRuleCount(), s.Description)
}

// 预定义的自定义策略构建器工厂函数

// NewHeadingBasedStrategyBuilder 创建基于标题的策略构建器
func NewHeadingBasedStrategyBuilder(name string, maxLevel int) *CustomStrategyBuilder {
	builder := NewCustomStrategyBuilder(name, fmt.Sprintf("基于标题层级的分块策略，最大层级：%d", maxLevel))

	// 添加标题处理规则
	builder.AddRule(
		"heading-separate",
		fmt.Sprintf("将层级1-%d的标题创建为独立块", maxLevel),
		NewHeadingLevelCondition(1, maxLevel),
		NewCreateSeparateChunkAction("", map[string]string{
			"heading_based": "true",
		}),
		100, // 高优先级
	)

	// 添加内容合并规则
	builder.AddRule(
		"content-merge",
		"将段落和列表等内容合并到前一个块",
		NewContentTypeCondition("paragraph", "list", "blockquote"),
		NewMergeWithParentAction("\n\n"),
		50, // 中等优先级
	)

	return builder
}

// NewContentTypeBasedStrategyBuilder 创建基于内容类型的策略构建器
func NewContentTypeBasedStrategyBuilder(name string, separateTypes, mergeTypes []string) *CustomStrategyBuilder {
	builder := NewCustomStrategyBuilder(name, "基于内容类型的分块策略")

	// 添加独立块规则
	if len(separateTypes) > 0 {
		builder.AddRule(
			"separate-types",
			fmt.Sprintf("将 %v 类型的内容创建为独立块", separateTypes),
			NewContentTypeCondition(separateTypes...),
			NewCreateSeparateChunkAction("", map[string]string{
				"content_type_based": "true",
			}),
			100, // 高优先级
		)
	}

	// 添加合并规则
	if len(mergeTypes) > 0 {
		builder.AddRule(
			"merge-types",
			fmt.Sprintf("将 %v 类型的内容合并到前一个块", mergeTypes),
			NewContentTypeCondition(mergeTypes...),
			NewMergeWithParentAction("\n\n"),
			50, // 中等优先级
		)
	}

	return builder
}

// NewSizeBasedStrategyBuilder 创建基于大小的策略构建器
func NewSizeBasedStrategyBuilder(name string, minSize, maxSize int) *CustomStrategyBuilder {
	builder := NewCustomStrategyBuilder(name, fmt.Sprintf("基于内容大小的分块策略，大小范围：%d-%d", minSize, maxSize))

	// 添加大内容独立处理规则
	builder.AddRule(
		"large-content-separate",
		fmt.Sprintf("将大于 %d 字符的内容创建为独立块", maxSize),
		NewContentSizeCondition(maxSize+1, 0),
		NewCreateSeparateChunkAction("", map[string]string{
			"size_based":   "true",
			"content_size": "large",
		}),
		100, // 高优先级
	)

	// 添加小内容合并规则
	builder.AddRule(
		"small-content-merge",
		fmt.Sprintf("将小于 %d 字符的内容合并到前一个块", minSize),
		NewContentSizeCondition(0, minSize-1),
		NewMergeWithParentAction(" "),
		80, // 较高优先级
	)

	// 添加中等大小内容独立处理规则
	builder.AddRule(
		"medium-content-separate",
		fmt.Sprintf("将 %d-%d 字符的内容创建为独立块", minSize, maxSize),
		NewContentSizeCondition(minSize, maxSize),
		NewCreateSeparateChunkAction("", map[string]string{
			"size_based":   "true",
			"content_size": "medium",
		}),
		60, // 中等优先级
	)

	return builder
}

// DocumentLevelStrategy 文档级分块策略
// 将整个文档作为单个块处理
type DocumentLevelStrategy struct {
	config *StrategyConfig
}

// NewDocumentLevelStrategy 创建新的文档级分块策略
func NewDocumentLevelStrategy() *DocumentLevelStrategy {
	return &DocumentLevelStrategy{
		config: DocumentLevelConfig(),
	}
}

// NewDocumentLevelStrategyWithConfig 使用指定配置创建文档级分块策略
func NewDocumentLevelStrategyWithConfig(config *StrategyConfig) *DocumentLevelStrategy {
	if config == nil {
		config = DocumentLevelConfig()
	}
	return &DocumentLevelStrategy{
		config: config,
	}
}

// GetName 返回策略名称
func (s *DocumentLevelStrategy) GetName() string {
	return "document-level"
}

// GetDescription 返回策略描述
func (s *DocumentLevelStrategy) GetDescription() string {
	return "将整个文档作为单个块处理"
}

// ChunkDocument 使用文档级策略对文档进行分块
func (s *DocumentLevelStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	if doc == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "文档节点不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	if chunker == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "分块器实例不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	if source == nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "源内容不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ChunkDocument")
	}

	// 检查是否为大文档，如果是则使用流式处理
	if len(source) > s.getLargeDocumentThreshold() {
		return s.chunkLargeDocument(doc, source, chunker)
	}

	// 将整个文档内容作为一个块
	content := string(source)

	// 提取纯文本内容
	text := s.extractTextFromDocument(doc, source)

	// 提取文档级别的元数据
	metadata := s.extractDocumentMetadata(doc, source, chunker)

	// 提取链接和图片
	links, images := s.extractLinksAndImages(doc, source)

	// 计算位置信息
	position := s.calculateDocumentPosition(source)

	// 计算内容哈希
	hash := fmt.Sprintf("%x", sha256.Sum256(source))

	// 创建文档级块
	chunk := Chunk{
		ID:       0,
		Type:     "document",
		Content:  content,
		Text:     text,
		Level:    0,
		Metadata: metadata,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
	}

	return []Chunk{chunk}, nil
}

// ValidateConfig 验证策略特定的配置
func (s *DocumentLevelStrategy) ValidateConfig(config *StrategyConfig) error {
	if config == nil {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "策略配置不能为空", nil).
			WithContext("strategy", s.GetName()).
			WithContext("function", "ValidateConfig")
	}

	// 文档级策略使用通用配置验证
	return config.ValidateConfig()
}

// Clone 创建策略的副本（用于并发安全）
func (s *DocumentLevelStrategy) Clone() ChunkingStrategy {
	var configClone *StrategyConfig
	if s.config != nil {
		configClone = s.config.Clone()
	}

	return &DocumentLevelStrategy{
		config: configClone,
	}
}

// extractTextFromDocument 从文档中提取纯文本内容
func (s *DocumentLevelStrategy) extractTextFromDocument(doc ast.Node, source []byte) string {
	var textParts []string

	// 遍历所有节点提取文本
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Text:
			// 提取文本节点的内容
			text := string(n.Segment.Value(source))
			if strings.TrimSpace(text) != "" {
				textParts = append(textParts, text)
			}
		case *ast.CodeSpan:
			// 对于代码段，跳过其子节点以避免重复提取
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	// 合并文本并清理
	mergedText := strings.Join(textParts, " ")
	return strings.Join(strings.Fields(mergedText), " ")
}

// extractDocumentMetadata 提取文档级别的元数据
func (s *DocumentLevelStrategy) extractDocumentMetadata(doc ast.Node, source []byte, chunker *MarkdownChunker) map[string]string {
	metadata := make(map[string]string)

	// 基本元数据
	metadata["strategy"] = "document-level"
	metadata["total_size"] = fmt.Sprintf("%d", len(source))
	metadata["content_length"] = fmt.Sprintf("%d", len(source))

	// 统计各种元素数量
	var headingCount, paragraphCount, codeBlockCount, tableCount, listCount int
	var maxHeadingLevel int

	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Heading:
			headingCount++
			if n.Level > maxHeadingLevel {
				maxHeadingLevel = n.Level
			}
		case *ast.Paragraph:
			paragraphCount++
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			codeBlockCount++
		case *extast.Table:
			tableCount++
		case *ast.List:
			listCount++
		}

		return ast.WalkContinue, nil
	})

	// 添加统计信息到元数据
	metadata["heading_count"] = fmt.Sprintf("%d", headingCount)
	metadata["paragraph_count"] = fmt.Sprintf("%d", paragraphCount)
	metadata["code_block_count"] = fmt.Sprintf("%d", codeBlockCount)
	metadata["table_count"] = fmt.Sprintf("%d", tableCount)
	metadata["list_count"] = fmt.Sprintf("%d", listCount)
	metadata["max_heading_level"] = fmt.Sprintf("%d", maxHeadingLevel)

	// 计算文本统计
	text := s.extractTextFromDocument(doc, source)
	wordCount := len(strings.Fields(text))
	metadata["text_length"] = fmt.Sprintf("%d", len(text))
	metadata["word_count"] = fmt.Sprintf("%d", wordCount)

	// 估算阅读时间（假设每分钟200词）
	readingTimeMinutes := wordCount / 200
	if readingTimeMinutes == 0 && wordCount > 0 {
		readingTimeMinutes = 1
	}
	metadata["estimated_reading_time_minutes"] = fmt.Sprintf("%d", readingTimeMinutes)

	// 文档复杂度评估
	complexity := s.calculateDocumentComplexity(headingCount, codeBlockCount, tableCount, listCount)
	metadata["document_complexity"] = complexity

	return metadata
}

// extractLinksAndImages 提取文档中的所有链接和图片
func (s *DocumentLevelStrategy) extractLinksAndImages(doc ast.Node, source []byte) ([]Link, []Image) {
	var links []Link
	var images []Image

	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Link:
			// 提取链接信息
			linkText := string(n.Text(source))
			linkURL := string(n.Destination)

			// 确定链接类型
			linkType := "external"
			if strings.HasPrefix(linkURL, "#") {
				linkType = "anchor"
			} else if strings.HasPrefix(linkURL, "/") || !strings.Contains(linkURL, "://") {
				linkType = "internal"
			}

			links = append(links, Link{
				Text: linkText,
				URL:  linkURL,
				Type: linkType,
			})

		case *ast.Image:
			// 提取图片信息
			imageAlt := string(n.Text(source))
			imageURL := string(n.Destination)
			imageTitle := string(n.Title)

			images = append(images, Image{
				Alt:   imageAlt,
				URL:   imageURL,
				Title: imageTitle,
				// Width 和 Height 在标准 Markdown 中通常不可用
				Width:  "",
				Height: "",
			})
		}

		return ast.WalkContinue, nil
	})

	return links, images
}

// calculateDocumentPosition 计算文档的位置信息
func (s *DocumentLevelStrategy) calculateDocumentPosition(source []byte) ChunkPosition {
	if len(source) == 0 {
		return ChunkPosition{
			StartLine: 1,
			EndLine:   1,
			StartCol:  1,
			EndCol:    1,
		}
	}

	// 计算总行数
	lines := strings.Split(string(source), "\n")
	endLine := len(lines)

	// 计算最后一行的列数
	endCol := 1
	if len(lines) > 0 {
		endCol = len(lines[len(lines)-1]) + 1
	}

	return ChunkPosition{
		StartLine: 1,
		EndLine:   endLine,
		StartCol:  1,
		EndCol:    endCol,
	}
}

// calculateDocumentComplexity 计算文档复杂度
func (s *DocumentLevelStrategy) calculateDocumentComplexity(headingCount, codeBlockCount, tableCount, listCount int) string {
	// 简单的复杂度评估算法
	complexityScore := headingCount + codeBlockCount*2 + tableCount*3 + listCount

	if complexityScore <= 5 {
		return "simple"
	} else if complexityScore <= 15 {
		return "moderate"
	} else if complexityScore <= 30 {
		return "complex"
	} else {
		return "very_complex"
	}
}

// GetConfig 获取策略配置
func (s *DocumentLevelStrategy) GetConfig() *StrategyConfig {
	if s.config == nil {
		return nil
	}
	return s.config.Clone()
}

// SetConfig 设置策略配置
func (s *DocumentLevelStrategy) SetConfig(config *StrategyConfig) error {
	if config == nil {
		s.config = DocumentLevelConfig()
		return nil
	}

	if err := s.ValidateConfig(config); err != nil {
		return err
	}

	s.config = config.Clone()
	return nil
}

// getLargeDocumentThreshold 获取大文档阈值（字节）
func (s *DocumentLevelStrategy) getLargeDocumentThreshold() int {
	// 默认阈值为 5MB
	defaultThreshold := 5 * 1024 * 1024

	if s.config != nil && s.config.Parameters != nil {
		if threshold, ok := s.config.Parameters["large_document_threshold"].(int); ok && threshold > 0 {
			return threshold
		}
	}

	return defaultThreshold
}

// chunkLargeDocument 使用流式处理处理大文档
func (s *DocumentLevelStrategy) chunkLargeDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	// 创建进度回调函数
	progressCallback := s.createProgressCallback(chunker, len(source))

	// 报告开始处理
	progressCallback(0, "开始处理大文档")

	// 使用流式处理提取文本内容
	text, err := s.extractTextStreamingly(doc, source, progressCallback)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "流式文本提取失败", err).
			WithContext("strategy", s.GetName()).
			WithContext("function", "chunkLargeDocument").
			WithContext("document_size", len(source))
	}

	// 报告文本提取完成
	progressCallback(25, "文本提取完成")

	// 使用流式处理提取元数据
	metadata, err := s.extractMetadataStreamingly(doc, source, chunker, progressCallback)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "流式元数据提取失败", err).
			WithContext("strategy", s.GetName()).
			WithContext("function", "chunkLargeDocument").
			WithContext("document_size", len(source))
	}

	// 报告元数据提取完成
	progressCallback(50, "元数据提取完成")

	// 使用流式处理提取链接和图片
	links, images, err := s.extractLinksAndImagesStreamingly(doc, source, progressCallback)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "流式链接和图片提取失败", err).
			WithContext("strategy", s.GetName()).
			WithContext("function", "chunkLargeDocument").
			WithContext("document_size", len(source))
	}

	// 报告链接和图片提取完成
	progressCallback(75, "链接和图片提取完成")

	// 计算位置信息（轻量级操作）
	position := s.calculateDocumentPosition(source)

	// 使用流式哈希计算
	hash, err := s.calculateHashStreamingly(source, progressCallback)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "流式哈希计算失败", err).
			WithContext("strategy", s.GetName()).
			WithContext("function", "chunkLargeDocument").
			WithContext("document_size", len(source))
	}

	// 报告哈希计算完成
	progressCallback(90, "哈希计算完成")

	// 创建文档级块（避免复制大量内容）
	content := string(source) // 这里仍然需要转换，但我们已经优化了其他部分

	chunk := Chunk{
		ID:       0,
		Type:     "document",
		Content:  content,
		Text:     text,
		Level:    0,
		Metadata: metadata,
		Position: position,
		Links:    links,
		Images:   images,
		Hash:     hash,
	}

	// 报告处理完成
	progressCallback(100, "大文档处理完成")

	return []Chunk{chunk}, nil
}

// createProgressCallback 创建进度回调函数
func (s *DocumentLevelStrategy) createProgressCallback(chunker *MarkdownChunker, totalSize int) func(int, string) {
	return func(progress int, message string) {
		if chunker != nil && chunker.logger != nil {
			chunker.logger.Infow("大文档处理进度",
				"strategy", s.GetName(),
				"progress_percent", progress,
				"message", message,
				"document_size_bytes", totalSize,
				"document_size_mb", totalSize/(1024*1024),
				"function", "chunkLargeDocument")
		}
	}
}

// extractTextStreamingly 使用流式处理提取文本内容
func (s *DocumentLevelStrategy) extractTextStreamingly(doc ast.Node, source []byte, progressCallback func(int, string)) (string, error) {
	var textParts []string
	processedNodes := 0
	totalNodes := s.countNodes(doc)

	// 遍历所有节点提取文本
	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		processedNodes++

		// 定期报告进度
		if processedNodes%1000 == 0 {
			progress := int(float64(processedNodes) / float64(totalNodes) * 25) // 文本提取占总进度的25%
			progressCallback(progress, fmt.Sprintf("正在提取文本 (%d/%d 节点)", processedNodes, totalNodes))
		}

		switch n := node.(type) {
		case *ast.Text:
			// 提取文本节点的内容
			text := string(n.Segment.Value(source))
			if strings.TrimSpace(text) != "" {
				textParts = append(textParts, text)
			}
		case *ast.CodeSpan:
			// 对于代码段，跳过其子节点以避免重复提取
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return "", err
	}

	// 合并文本并清理
	mergedText := strings.Join(textParts, " ")
	return strings.Join(strings.Fields(mergedText), " "), nil
}

// extractMetadataStreamingly 使用流式处理提取元数据
func (s *DocumentLevelStrategy) extractMetadataStreamingly(doc ast.Node, source []byte, chunker *MarkdownChunker, progressCallback func(int, string)) (map[string]string, error) {
	metadata := make(map[string]string)

	// 基本元数据
	metadata["strategy"] = "document-level"
	metadata["total_size"] = fmt.Sprintf("%d", len(source))
	metadata["content_length"] = fmt.Sprintf("%d", len(source))
	metadata["processing_mode"] = "streaming"

	// 统计各种元素数量
	var headingCount, paragraphCount, codeBlockCount, tableCount, listCount int
	var maxHeadingLevel int
	processedNodes := 0
	totalNodes := s.countNodes(doc)

	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		processedNodes++

		// 定期报告进度
		if processedNodes%500 == 0 {
			progress := 25 + int(float64(processedNodes)/float64(totalNodes)*25) // 元数据提取占25%
			progressCallback(progress, fmt.Sprintf("正在提取元数据 (%d/%d 节点)", processedNodes, totalNodes))
		}

		switch n := node.(type) {
		case *ast.Heading:
			headingCount++
			if n.Level > maxHeadingLevel {
				maxHeadingLevel = n.Level
			}
		case *ast.Paragraph:
			paragraphCount++
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			codeBlockCount++
		case *extast.Table:
			tableCount++
		case *ast.List:
			listCount++
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	// 添加统计信息到元数据
	metadata["heading_count"] = fmt.Sprintf("%d", headingCount)
	metadata["paragraph_count"] = fmt.Sprintf("%d", paragraphCount)
	metadata["code_block_count"] = fmt.Sprintf("%d", codeBlockCount)
	metadata["table_count"] = fmt.Sprintf("%d", tableCount)
	metadata["list_count"] = fmt.Sprintf("%d", listCount)
	metadata["max_heading_level"] = fmt.Sprintf("%d", maxHeadingLevel)
	metadata["total_nodes"] = fmt.Sprintf("%d", totalNodes)

	// 计算文本统计（重用已提取的文本）
	text, err := s.extractTextStreamingly(doc, source, func(int, string) {}) // 空回调避免重复日志
	if err != nil {
		return nil, err
	}

	wordCount := len(strings.Fields(text))
	metadata["text_length"] = fmt.Sprintf("%d", len(text))
	metadata["word_count"] = fmt.Sprintf("%d", wordCount)

	// 估算阅读时间（假设每分钟200词）
	readingTimeMinutes := wordCount / 200
	if readingTimeMinutes == 0 && wordCount > 0 {
		readingTimeMinutes = 1
	}
	metadata["estimated_reading_time_minutes"] = fmt.Sprintf("%d", readingTimeMinutes)

	// 文档复杂度评估
	complexity := s.calculateDocumentComplexity(headingCount, codeBlockCount, tableCount, listCount)
	metadata["document_complexity"] = complexity

	return metadata, nil
}

// extractLinksAndImagesStreamingly 使用流式处理提取链接和图片
func (s *DocumentLevelStrategy) extractLinksAndImagesStreamingly(doc ast.Node, source []byte, progressCallback func(int, string)) ([]Link, []Image, error) {
	var links []Link
	var images []Image
	processedNodes := 0
	totalNodes := s.countNodes(doc)

	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		processedNodes++

		// 定期报告进度
		if processedNodes%500 == 0 {
			progress := 50 + int(float64(processedNodes)/float64(totalNodes)*25) // 链接和图片提取占25%
			progressCallback(progress, fmt.Sprintf("正在提取链接和图片 (%d/%d 节点)", processedNodes, totalNodes))
		}

		switch n := node.(type) {
		case *ast.Link:
			// 提取链接信息
			linkText := string(n.Text(source))
			linkURL := string(n.Destination)

			// 确定链接类型
			linkType := "external"
			if strings.HasPrefix(linkURL, "#") {
				linkType = "anchor"
			} else if strings.HasPrefix(linkURL, "/") || !strings.Contains(linkURL, "://") {
				linkType = "internal"
			}

			links = append(links, Link{
				Text: linkText,
				URL:  linkURL,
				Type: linkType,
			})

		case *ast.Image:
			// 提取图片信息
			imageAlt := string(n.Text(source))
			imageURL := string(n.Destination)
			imageTitle := string(n.Title)

			images = append(images, Image{
				Alt:   imageAlt,
				URL:   imageURL,
				Title: imageTitle,
				// Width 和 Height 在标准 Markdown 中通常不可用
				Width:  "",
				Height: "",
			})
		}

		return ast.WalkContinue, nil
	})

	return links, images, err
}

// calculateHashStreamingly 使用流式处理计算哈希
func (s *DocumentLevelStrategy) calculateHashStreamingly(source []byte, progressCallback func(int, string)) (string, error) {
	hasher := sha256.New()

	// 分块处理以减少内存占用
	chunkSize := 64 * 1024 // 64KB 块
	totalSize := len(source)
	processed := 0

	for processed < totalSize {
		end := processed + chunkSize
		if end > totalSize {
			end = totalSize
		}

		chunk := source[processed:end]
		hasher.Write(chunk)

		processed = end

		// 报告进度
		progress := 75 + int(float64(processed)/float64(totalSize)*15) // 哈希计算占15%
		progressCallback(progress, fmt.Sprintf("正在计算哈希 (%d/%d 字节)", processed, totalSize))
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// countNodes 计算 AST 中的节点总数
func (s *DocumentLevelStrategy) countNodes(doc ast.Node) int {
	count := 0
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			count++
		}
		return ast.WalkContinue, nil
	})
	return count
}

// optimizeHierarchy 优化层级结构，移除不必要的中间层级
func (s *HierarchicalStrategy) optimizeHierarchy(chunks []*HierarchicalChunk) []*HierarchicalChunk {
	var optimized []*HierarchicalChunk

	for _, chunk := range chunks {
		optimizedChunk := s.optimizeChunk(chunk)
		if optimizedChunk != nil {
			optimized = append(optimized, optimizedChunk)
		}
	}

	return optimized
}

// optimizeChunk 优化单个层级块
func (s *HierarchicalStrategy) optimizeChunk(chunk *HierarchicalChunk) *HierarchicalChunk {
	if chunk == nil {
		return nil
	}

	// 递归优化子块
	var optimizedChildren []*HierarchicalChunk
	for _, child := range chunk.Children {
		optimizedChild := s.optimizeChunk(child)
		if optimizedChild != nil {
			optimizedChildren = append(optimizedChildren, optimizedChild)
		}
	}

	// 如果是虚拟块且只有一个子块，考虑合并
	if s.isVirtualChunk(chunk) && len(optimizedChildren) == 1 {
		// 将子块提升到当前层级
		child := optimizedChildren[0]
		child.Parent = chunk.Parent
		child.Level = chunk.Level
		return child
	}

	// 如果是空的虚拟块，移除
	if s.isVirtualChunk(chunk) && len(optimizedChildren) == 0 {
		return nil
	}

	// 更新子块列表
	chunk.Children = optimizedChildren

	// 更新子块的父引用
	for _, child := range chunk.Children {
		child.Parent = chunk
	}

	return chunk
}

// printHierarchy 打印层级结构（用于调试）
func (s *HierarchicalStrategy) printHierarchy(chunks []*HierarchicalChunk, indent string) string {
	var result strings.Builder

	for _, chunk := range chunks {
		result.WriteString(fmt.Sprintf("%s- %s (Level: %d, Type: %s",
			indent,
			strings.TrimSpace(chunk.Chunk.Text[:min(50, len(chunk.Chunk.Text))]),
			chunk.Level,
			chunk.Chunk.Type))

		if s.isVirtualChunk(chunk) {
			result.WriteString(", Virtual")
		}

		result.WriteString(")\n")

		if len(chunk.Children) > 0 {
			result.WriteString(s.printHierarchy(chunk.Children, indent+"  "))
		}
	}

	return result.String()
}
