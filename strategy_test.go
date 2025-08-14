package markdownchunker

import (
	"fmt"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MockStrategy 用于测试的模拟策略
type MockStrategy struct {
	name        string
	description string
	shouldError bool
}

func (m *MockStrategy) GetName() string {
	return m.name
}

func (m *MockStrategy) GetDescription() string {
	return m.description
}

func (m *MockStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	if m.shouldError {
		return nil, NewChunkerError(ErrorTypeStrategyExecutionFailed, "模拟策略执行失败", nil)
	}
	return []Chunk{}, nil
}

func (m *MockStrategy) ValidateConfig(config *StrategyConfig) error {
	if m.shouldError {
		return NewChunkerError(ErrorTypeStrategyConfigInvalid, "模拟配置验证失败", nil)
	}
	return nil
}

func (m *MockStrategy) Clone() ChunkingStrategy {
	return &MockStrategy{
		name:        m.name,
		description: m.description,
		shouldError: m.shouldError,
	}
}

// TestStrategyRegistry 测试策略注册器
func TestStrategyRegistry(t *testing.T) {
	t.Run("创建新的策略注册器", func(t *testing.T) {
		registry := NewStrategyRegistry()
		if registry == nil {
			t.Fatal("策略注册器创建失败")
		}
		if registry.GetStrategyCount() != 0 {
			t.Errorf("新创建的注册器应该没有策略，实际有 %d 个", registry.GetStrategyCount())
		}
	})

	t.Run("注册策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		strategy := &MockStrategy{name: "test-strategy", description: "测试策略"}

		err := registry.Register(strategy)
		if err != nil {
			t.Fatalf("注册策略失败: %v", err)
		}

		if registry.GetStrategyCount() != 1 {
			t.Errorf("注册后应该有1个策略，实际有 %d 个", registry.GetStrategyCount())
		}

		if !registry.HasStrategy("test-strategy") {
			t.Error("注册的策略应该存在")
		}
	})

	t.Run("注册空策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		err := registry.Register(nil)
		if err == nil {
			t.Error("注册空策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("注册空名称策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		strategy := &MockStrategy{name: "", description: "无名策略"}

		err := registry.Register(strategy)
		if err == nil {
			t.Error("注册空名称策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("重复注册策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		strategy1 := &MockStrategy{name: "duplicate-strategy", description: "第一个策略"}
		strategy2 := &MockStrategy{name: "duplicate-strategy", description: "第二个策略"}

		err := registry.Register(strategy1)
		if err != nil {
			t.Fatalf("第一次注册失败: %v", err)
		}

		err = registry.Register(strategy2)
		if err == nil {
			t.Error("重复注册策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("获取策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		originalStrategy := &MockStrategy{name: "get-test", description: "获取测试策略"}

		err := registry.Register(originalStrategy)
		if err != nil {
			t.Fatalf("注册策略失败: %v", err)
		}

		retrievedStrategy, err := registry.Get("get-test")
		if err != nil {
			t.Fatalf("获取策略失败: %v", err)
		}

		if retrievedStrategy.GetName() != "get-test" {
			t.Errorf("获取的策略名称不正确，期望 'get-test'，实际 '%s'", retrievedStrategy.GetName())
		}
	})

	t.Run("获取不存在的策略", func(t *testing.T) {
		registry := NewStrategyRegistry()

		_, err := registry.Get("non-existent")
		if err == nil {
			t.Error("获取不存在的策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyNotFound {
			t.Errorf("错误类型应该是 ErrorTypeStrategyNotFound，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("获取空名称策略", func(t *testing.T) {
		registry := NewStrategyRegistry()

		_, err := registry.Get("")
		if err == nil {
			t.Error("获取空名称策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyNotFound {
			t.Errorf("错误类型应该是 ErrorTypeStrategyNotFound，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("列出策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		strategies := []*MockStrategy{
			{name: "strategy1", description: "策略1"},
			{name: "strategy2", description: "策略2"},
			{name: "strategy3", description: "策略3"},
		}

		for _, strategy := range strategies {
			err := registry.Register(strategy)
			if err != nil {
				t.Fatalf("注册策略 %s 失败: %v", strategy.name, err)
			}
		}

		names := registry.List()
		if len(names) != 3 {
			t.Errorf("应该有3个策略，实际有 %d 个", len(names))
		}

		// 检查所有策略名称都在列表中
		nameMap := make(map[string]bool)
		for _, name := range names {
			nameMap[name] = true
		}

		for _, strategy := range strategies {
			if !nameMap[strategy.name] {
				t.Errorf("策略 %s 不在列表中", strategy.name)
			}
		}
	})

	t.Run("注销策略", func(t *testing.T) {
		registry := NewStrategyRegistry()
		strategy := &MockStrategy{name: "unregister-test", description: "注销测试策略"}

		err := registry.Register(strategy)
		if err != nil {
			t.Fatalf("注册策略失败: %v", err)
		}

		if !registry.HasStrategy("unregister-test") {
			t.Error("策略应该存在")
		}

		err = registry.Unregister("unregister-test")
		if err != nil {
			t.Fatalf("注销策略失败: %v", err)
		}

		if registry.HasStrategy("unregister-test") {
			t.Error("策略应该已被注销")
		}

		if registry.GetStrategyCount() != 0 {
			t.Errorf("注销后应该没有策略，实际有 %d 个", registry.GetStrategyCount())
		}
	})

	t.Run("注销不存在的策略", func(t *testing.T) {
		registry := NewStrategyRegistry()

		err := registry.Unregister("non-existent")
		if err == nil {
			t.Error("注销不存在的策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyNotFound {
			t.Errorf("错误类型应该是 ErrorTypeStrategyNotFound，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("注销空名称策略", func(t *testing.T) {
		registry := NewStrategyRegistry()

		err := registry.Unregister("")
		if err == nil {
			t.Error("注销空名称策略应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyNotFound {
			t.Errorf("错误类型应该是 ErrorTypeStrategyNotFound，实际是 %v", chunkerErr.Type)
		}
	})
}

// TestStrategyConfig 测试策略配置
func TestStrategyConfig(t *testing.T) {
	t.Run("验证有效配置", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "test-strategy",
			MaxDepth:     3,
			MinDepth:     1,
			MergeEmpty:   true,
			MinChunkSize: 100,
			MaxChunkSize: 1000,
			IncludeTypes: []string{"heading", "paragraph"},
			ExcludeTypes: []string{"code"},
		}

		err := config.ValidateConfig()
		if err != nil {
			t.Errorf("有效配置验证失败: %v", err)
		}
	})

	t.Run("验证空配置", func(t *testing.T) {
		var config *StrategyConfig = nil

		err := config.ValidateConfig()
		if err == nil {
			t.Error("空配置应该验证失败")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证空名称配置", func(t *testing.T) {
		config := &StrategyConfig{
			Name: "",
		}

		err := config.ValidateConfig()
		if err == nil {
			t.Error("空名称配置应该验证失败")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证负数层级深度", func(t *testing.T) {
		config := &StrategyConfig{
			Name:     "test",
			MaxDepth: -1,
		}

		err := config.ValidateConfig()
		if err == nil {
			t.Error("负数最大层级深度应该验证失败")
		}

		config.MaxDepth = 0
		config.MinDepth = -1

		err = config.ValidateConfig()
		if err == nil {
			t.Error("负数最小层级深度应该验证失败")
		}
	})

	t.Run("验证层级深度逻辑", func(t *testing.T) {
		config := &StrategyConfig{
			Name:     "test",
			MaxDepth: 2,
			MinDepth: 3,
		}

		err := config.ValidateConfig()
		if err == nil {
			t.Error("最小层级深度大于最大层级深度应该验证失败")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证负数块大小", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "test",
			MinChunkSize: -1,
		}

		err := config.ValidateConfig()
		if err == nil {
			t.Error("负数最小块大小应该验证失败")
		}

		config.MinChunkSize = 0
		config.MaxChunkSize = -1

		err = config.ValidateConfig()
		if err == nil {
			t.Error("负数最大块大小应该验证失败")
		}
	})

	t.Run("验证块大小逻辑", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "test",
			MinChunkSize: 1000,
			MaxChunkSize: 500,
		}

		err := config.ValidateConfig()
		if err == nil {
			t.Error("最小块大小大于最大块大小应该验证失败")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证无效内容类型", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "test",
			IncludeTypes: []string{"invalid_type"},
		}

		err := config.ValidateConfig()
		if err == nil {
			t.Error("无效的包含内容类型应该验证失败")
		}

		config.IncludeTypes = nil
		config.ExcludeTypes = []string{"invalid_type"}

		err = config.ValidateConfig()
		if err == nil {
			t.Error("无效的排除内容类型应该验证失败")
		}
	})

	t.Run("克隆配置", func(t *testing.T) {
		original := &StrategyConfig{
			Name:         "original",
			MaxDepth:     3,
			MinDepth:     1,
			MergeEmpty:   true,
			MinChunkSize: 100,
			MaxChunkSize: 1000,
			Parameters:   map[string]any{"key": "value"},
			IncludeTypes: []string{"heading", "paragraph"},
			ExcludeTypes: []string{"code"},
		}

		clone := original.Clone()
		if clone == nil {
			t.Fatal("克隆配置失败")
		}

		// 验证基本字段
		if clone.Name != original.Name {
			t.Errorf("克隆的名称不匹配，期望 '%s'，实际 '%s'", original.Name, clone.Name)
		}
		if clone.MaxDepth != original.MaxDepth {
			t.Errorf("克隆的最大深度不匹配，期望 %d，实际 %d", original.MaxDepth, clone.MaxDepth)
		}

		// 验证深拷贝
		clone.Parameters["key"] = "modified"
		if original.Parameters["key"] == "modified" {
			t.Error("参数映射应该是深拷贝")
		}

		clone.IncludeTypes[0] = "modified"
		if original.IncludeTypes[0] == "modified" {
			t.Error("包含类型切片应该是深拷贝")
		}

		clone.ExcludeTypes[0] = "modified"
		if original.ExcludeTypes[0] == "modified" {
			t.Error("排除类型切片应该是深拷贝")
		}
	})

	t.Run("克隆空配置", func(t *testing.T) {
		var original *StrategyConfig = nil
		clone := original.Clone()
		if clone != nil {
			t.Error("克隆空配置应该返回 nil")
		}
	})

	t.Run("配置字符串表示", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "test-strategy",
			MaxDepth:     3,
			MinDepth:     1,
			MergeEmpty:   true,
			MinChunkSize: 100,
			MaxChunkSize: 1000,
		}

		str := config.String()
		if str == "" {
			t.Error("配置字符串表示不应该为空")
		}

		// 验证包含关键信息
		if !contains(str, "test-strategy") {
			t.Error("字符串表示应该包含策略名称")
		}
		if !contains(str, "MaxDepth: 3") {
			t.Error("字符串表示应该包含最大深度")
		}
	})

	t.Run("空配置字符串表示", func(t *testing.T) {
		var config *StrategyConfig = nil
		str := config.String()
		if str != "<nil>" {
			t.Errorf("空配置字符串表示应该是 '<nil>'，实际是 '%s'", str)
		}
	})
}

// TestDefaultConfigs 测试默认配置函数
func TestDefaultConfigs(t *testing.T) {
	t.Run("默认策略配置", func(t *testing.T) {
		config := DefaultStrategyConfig("test-strategy")
		if config == nil {
			t.Fatal("默认策略配置创建失败")
		}

		if config.Name != "test-strategy" {
			t.Errorf("默认配置名称不正确，期望 'test-strategy'，实际 '%s'", config.Name)
		}

		if config.Parameters == nil {
			t.Error("默认配置参数映射不应该为 nil")
		}

		if !config.MergeEmpty {
			t.Error("默认配置应该启用合并空章节")
		}

		err := config.ValidateConfig()
		if err != nil {
			t.Errorf("默认配置验证失败: %v", err)
		}
	})

	t.Run("层级配置", func(t *testing.T) {
		config := HierarchicalConfig(3)
		if config == nil {
			t.Fatal("层级配置创建失败")
		}

		if config.Name != "hierarchical" {
			t.Errorf("层级配置名称不正确，期望 'hierarchical'，实际 '%s'", config.Name)
		}

		if config.MaxDepth != 3 {
			t.Errorf("层级配置最大深度不正确，期望 3，实际 %d", config.MaxDepth)
		}

		if config.Parameters["max_depth"] != 3 {
			t.Error("层级配置参数中应该包含 max_depth")
		}

		err := config.ValidateConfig()
		if err != nil {
			t.Errorf("层级配置验证失败: %v", err)
		}
	})

	t.Run("文档级配置", func(t *testing.T) {
		config := DocumentLevelConfig()
		if config == nil {
			t.Fatal("文档级配置创建失败")
		}

		if config.Name != "document-level" {
			t.Errorf("文档级配置名称不正确，期望 'document-level'，实际 '%s'", config.Name)
		}

		err := config.ValidateConfig()
		if err != nil {
			t.Errorf("文档级配置验证失败: %v", err)
		}
	})

	t.Run("元素级配置", func(t *testing.T) {
		config := ElementLevelConfig()
		if config == nil {
			t.Fatal("元素级配置创建失败")
		}

		if config.Name != "element-level" {
			t.Errorf("元素级配置名称不正确，期望 'element-level'，实际 '%s'", config.Name)
		}

		err := config.ValidateConfig()
		if err != nil {
			t.Errorf("元素级配置验证失败: %v", err)
		}
	})
}

// TestMockStrategy 测试模拟策略
func TestMockStrategy(t *testing.T) {
	t.Run("正常模拟策略", func(t *testing.T) {
		strategy := &MockStrategy{
			name:        "mock-strategy",
			description: "模拟策略用于测试",
			shouldError: false,
		}

		if strategy.GetName() != "mock-strategy" {
			t.Errorf("策略名称不正确，期望 'mock-strategy'，实际 '%s'", strategy.GetName())
		}

		if strategy.GetDescription() != "模拟策略用于测试" {
			t.Errorf("策略描述不正确")
		}

		config := DefaultStrategyConfig("mock-strategy")
		err := strategy.ValidateConfig(config)
		if err != nil {
			t.Errorf("配置验证失败: %v", err)
		}

		chunks, err := strategy.ChunkDocument(nil, nil, nil)
		if err != nil {
			t.Errorf("文档分块失败: %v", err)
		}
		if chunks == nil {
			t.Error("分块结果不应该为 nil")
		}

		clone := strategy.Clone()
		if clone == nil {
			t.Error("策略克隆失败")
		}
		if clone.GetName() != strategy.GetName() {
			t.Error("克隆的策略名称不匹配")
		}
	})

	t.Run("错误模拟策略", func(t *testing.T) {
		strategy := &MockStrategy{
			name:        "error-strategy",
			description: "错误策略用于测试",
			shouldError: true,
		}

		config := DefaultStrategyConfig("error-strategy")
		err := strategy.ValidateConfig(config)
		if err == nil {
			t.Error("错误策略配置验证应该失败")
		}

		_, err = strategy.ChunkDocument(nil, nil, nil)
		if err == nil {
			t.Error("错误策略文档分块应该失败")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyExecutionFailed {
			t.Errorf("错误类型应该是 ErrorTypeStrategyExecutionFailed，实际是 %v", chunkerErr.Type)
		}
	})
}

// TestStrategyCache 测试策略缓存
func TestStrategyCache(t *testing.T) {
	t.Run("创建新的策略缓存", func(t *testing.T) {
		cache := NewStrategyCache()
		if cache == nil {
			t.Fatal("策略缓存创建失败")
		}
		if cache.Size() != 0 {
			t.Errorf("新创建的缓存应该为空，实际大小 %d", cache.Size())
		}
	})

	t.Run("缓存策略", func(t *testing.T) {
		cache := NewStrategyCache()
		strategy := &MockStrategy{name: "cache-test", description: "缓存测试策略"}

		cache.Put("cache-test", strategy)
		if cache.Size() != 1 {
			t.Errorf("缓存后大小应该为1，实际为 %d", cache.Size())
		}

		retrieved, exists := cache.Get("cache-test")
		if !exists {
			t.Error("缓存的策略应该存在")
		}
		if retrieved.GetName() != "cache-test" {
			t.Errorf("获取的策略名称不正确，期望 'cache-test'，实际 '%s'", retrieved.GetName())
		}
	})

	t.Run("缓存空策略", func(t *testing.T) {
		cache := NewStrategyCache()
		cache.Put("empty", nil)
		if cache.Size() != 0 {
			t.Error("缓存空策略不应该增加缓存大小")
		}
	})

	t.Run("缓存空名称策略", func(t *testing.T) {
		cache := NewStrategyCache()
		strategy := &MockStrategy{name: "test", description: "测试"}
		cache.Put("", strategy)
		if cache.Size() != 0 {
			t.Error("缓存空名称策略不应该增加缓存大小")
		}
	})

	t.Run("获取不存在的策略", func(t *testing.T) {
		cache := NewStrategyCache()
		_, exists := cache.Get("non-existent")
		if exists {
			t.Error("不存在的策略不应该在缓存中")
		}
	})

	t.Run("移除缓存策略", func(t *testing.T) {
		cache := NewStrategyCache()
		strategy := &MockStrategy{name: "remove-test", description: "移除测试策略"}

		cache.Put("remove-test", strategy)
		if cache.Size() != 1 {
			t.Error("缓存后大小应该为1")
		}

		cache.Remove("remove-test")
		if cache.Size() != 0 {
			t.Error("移除后缓存应该为空")
		}

		_, exists := cache.Get("remove-test")
		if exists {
			t.Error("移除的策略不应该存在")
		}
	})

	t.Run("清空缓存", func(t *testing.T) {
		cache := NewStrategyCache()
		strategies := []*MockStrategy{
			{name: "strategy1", description: "策略1"},
			{name: "strategy2", description: "策略2"},
			{name: "strategy3", description: "策略3"},
		}

		for _, strategy := range strategies {
			cache.Put(strategy.name, strategy)
		}

		if cache.Size() != 3 {
			t.Errorf("缓存大小应该为3，实际为 %d", cache.Size())
		}

		cache.Clear()
		if cache.Size() != 0 {
			t.Errorf("清空后缓存大小应该为0，实际为 %d", cache.Size())
		}
	})

	t.Run("获取缓存键列表", func(t *testing.T) {
		cache := NewStrategyCache()
		strategies := []*MockStrategy{
			{name: "strategy1", description: "策略1"},
			{name: "strategy2", description: "策略2"},
			{name: "strategy3", description: "策略3"},
		}

		for _, strategy := range strategies {
			cache.Put(strategy.name, strategy)
		}

		keys := cache.Keys()
		if len(keys) != 3 {
			t.Errorf("键列表长度应该为3，实际为 %d", len(keys))
		}

		// 检查所有键都存在
		keyMap := make(map[string]bool)
		for _, key := range keys {
			keyMap[key] = true
		}

		for _, strategy := range strategies {
			if !keyMap[strategy.name] {
				t.Errorf("键 %s 不在列表中", strategy.name)
			}
		}
	})
}

// TestStrategyPool 测试策略池
func TestStrategyPool(t *testing.T) {
	t.Run("创建新的策略池", func(t *testing.T) {
		pool := NewStrategyPool()
		if pool == nil {
			t.Fatal("策略池创建失败")
		}
		if pool.GetPoolCount() != 0 {
			t.Errorf("新创建的池应该没有子池，实际有 %d 个", pool.GetPoolCount())
		}
	})

	t.Run("从池获取策略实例", func(t *testing.T) {
		pool := NewStrategyPool()
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "pool-test", description: "池测试策略"}
		}

		strategy := pool.Get("pool-test", factory)
		if strategy == nil {
			t.Fatal("从池获取策略失败")
		}
		if strategy.GetName() != "pool-test" {
			t.Errorf("获取的策略名称不正确，期望 'pool-test'，实际 '%s'", strategy.GetName())
		}

		if !pool.HasPool("pool-test") {
			t.Error("池应该存在")
		}
		if pool.GetPoolCount() != 1 {
			t.Errorf("池数量应该为1，实际为 %d", pool.GetPoolCount())
		}
	})

	t.Run("将策略实例放回池", func(t *testing.T) {
		pool := NewStrategyPool()
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "put-test", description: "放回测试策略"}
		}

		// 先获取一个实例
		strategy1 := pool.Get("put-test", factory)
		if strategy1 == nil {
			t.Fatal("获取策略失败")
		}

		// 放回池中
		pool.Put(strategy1)

		// 再次获取，应该得到同一个实例（或者至少是相同类型的实例）
		strategy2 := pool.Get("put-test", factory)
		if strategy2 == nil {
			t.Fatal("再次获取策略失败")
		}
		if strategy2.GetName() != "put-test" {
			t.Error("获取的策略名称不正确")
		}
	})

	t.Run("放回空策略", func(t *testing.T) {
		pool := NewStrategyPool()
		pool.Put(nil) // 不应该崩溃
	})

	t.Run("放回无名称策略", func(t *testing.T) {
		pool := NewStrategyPool()
		strategy := &MockStrategy{name: "", description: "无名策略"}
		pool.Put(strategy) // 不应该崩溃
	})

	t.Run("创建指定策略的池", func(t *testing.T) {
		pool := NewStrategyPool()
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "create-pool-test", description: "创建池测试策略"}
		}

		pool.CreatePool("create-pool-test", factory)
		if !pool.HasPool("create-pool-test") {
			t.Error("创建的池应该存在")
		}
		if pool.GetPoolCount() != 1 {
			t.Errorf("池数量应该为1，实际为 %d", pool.GetPoolCount())
		}

		// 获取实例
		strategy := pool.Get("create-pool-test", nil)
		if strategy == nil {
			t.Fatal("从预创建的池获取策略失败")
		}
		if strategy.GetName() != "create-pool-test" {
			t.Error("获取的策略名称不正确")
		}
	})

	t.Run("创建池时传入空参数", func(t *testing.T) {
		pool := NewStrategyPool()

		// 空名称
		pool.CreatePool("", func() ChunkingStrategy { return nil })
		if pool.GetPoolCount() != 0 {
			t.Error("空名称不应该创建池")
		}

		// 空工厂函数
		pool.CreatePool("test", nil)
		if pool.GetPoolCount() != 0 {
			t.Error("空工厂函数不应该创建池")
		}
	})

	t.Run("移除池", func(t *testing.T) {
		pool := NewStrategyPool()
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "remove-pool-test", description: "移除池测试策略"}
		}

		pool.CreatePool("remove-pool-test", factory)
		if pool.GetPoolCount() != 1 {
			t.Error("创建后池数量应该为1")
		}

		pool.RemovePool("remove-pool-test")
		if pool.HasPool("remove-pool-test") {
			t.Error("移除的池不应该存在")
		}
		if pool.GetPoolCount() != 0 {
			t.Errorf("移除后池数量应该为0，实际为 %d", pool.GetPoolCount())
		}
	})

	t.Run("清空所有池", func(t *testing.T) {
		pool := NewStrategyPool()
		factories := []struct {
			name    string
			factory func() ChunkingStrategy
		}{
			{"pool1", func() ChunkingStrategy { return &MockStrategy{name: "pool1", description: "池1"} }},
			{"pool2", func() ChunkingStrategy { return &MockStrategy{name: "pool2", description: "池2"} }},
			{"pool3", func() ChunkingStrategy { return &MockStrategy{name: "pool3", description: "池3"} }},
		}

		for _, f := range factories {
			pool.CreatePool(f.name, f.factory)
		}

		if pool.GetPoolCount() != 3 {
			t.Errorf("创建后池数量应该为3，实际为 %d", pool.GetPoolCount())
		}

		pool.Clear()
		if pool.GetPoolCount() != 0 {
			t.Errorf("清空后池数量应该为0，实际为 %d", pool.GetPoolCount())
		}

		for _, f := range factories {
			if pool.HasPool(f.name) {
				t.Errorf("清空后池 %s 不应该存在", f.name)
			}
		}
	})

	t.Run("并发访问池", func(t *testing.T) {
		pool := NewStrategyPool()
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "concurrent-test", description: "并发测试策略"}
		}

		const numGoroutines = 10
		const numOperations = 100

		// 并发获取和放回策略
		done := make(chan bool, numGoroutines)
		for range numGoroutines {
			go func() {
				defer func() { done <- true }()
				for range numOperations {
					strategy := pool.Get("concurrent-test", factory)
					if strategy != nil {
						pool.Put(strategy)
					}
				}
			}()
		}

		// 等待所有goroutine完成
		for range numGoroutines {
			<-done
		}

		// 验证池仍然正常工作
		strategy := pool.Get("concurrent-test", factory)
		if strategy == nil {
			t.Error("并发测试后池应该仍然可用")
		}
	})
}

// TestCacheAndPoolIntegration 测试缓存和池的集成
func TestCacheAndPoolIntegration(t *testing.T) {
	t.Run("缓存和池协同工作", func(t *testing.T) {
		cache := NewStrategyCache()
		pool := NewStrategyPool()

		// 创建工厂函数
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "integration-test", description: "集成测试策略"}
		}

		// 从池获取策略
		strategy1 := pool.Get("integration-test", factory)
		if strategy1 == nil {
			t.Fatal("从池获取策略失败")
		}

		// 将策略放入缓存
		cache.Put("integration-test", strategy1)

		// 从缓存获取策略
		cachedStrategy, exists := cache.Get("integration-test")
		if !exists {
			t.Fatal("从缓存获取策略失败")
		}

		if cachedStrategy.GetName() != strategy1.GetName() {
			t.Error("缓存的策略与原策略不匹配")
		}

		// 将策略放回池
		pool.Put(strategy1)

		// 再次从池获取
		strategy2 := pool.Get("integration-test", factory)
		if strategy2 == nil {
			t.Fatal("再次从池获取策略失败")
		}

		if strategy2.GetName() != "integration-test" {
			t.Error("池中的策略名称不正确")
		}
	})
}

// BenchmarkStrategyCache 缓存性能基准测试
func BenchmarkStrategyCache(b *testing.B) {
	cache := NewStrategyCache()
	strategy := &MockStrategy{name: "benchmark-test", description: "基准测试策略"}
	cache.Put("benchmark-test", strategy)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get("benchmark-test")
		}
	})
}

// BenchmarkStrategyPool 池性能基准测试
func BenchmarkStrategyPool(b *testing.B) {
	pool := NewStrategyPool()
	factory := func() ChunkingStrategy {
		return &MockStrategy{name: "benchmark-test", description: "基准测试策略"}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			strategy := pool.Get("benchmark-test", factory)
			if strategy != nil {
				pool.Put(strategy)
			}
		}
	})
}

// BenchmarkDirectStrategyCreation 直接创建策略的基准测试（对比）
func BenchmarkDirectStrategyCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = &MockStrategy{name: "benchmark-test", description: "基准测试策略"}
		}
	})
}

// BenchmarkCacheVsDirectAccess 缓存访问与直接访问的性能对比
func BenchmarkCacheVsDirectAccess(b *testing.B) {
	b.Run("Cache", func(b *testing.B) {
		cache := NewStrategyCache()
		strategy := &MockStrategy{name: "compare-test", description: "对比测试策略"}
		cache.Put("compare-test", strategy)

		b.ResetTimer()
		for b.Loop() {
			_, _ = cache.Get("compare-test")
		}
	})

	b.Run("Direct", func(b *testing.B) {
		strategy := &MockStrategy{name: "compare-test", description: "对比测试策略"}

		b.ResetTimer()
		for b.Loop() {
			_ = strategy
		}
	})
}

// BenchmarkPoolVsDirectCreation 池创建与直接创建的性能对比
func BenchmarkPoolVsDirectCreation(b *testing.B) {
	b.Run("Pool", func(b *testing.B) {
		pool := NewStrategyPool()
		factory := func() ChunkingStrategy {
			return &MockStrategy{name: "compare-test", description: "对比测试策略"}
		}

		b.ResetTimer()
		for b.Loop() {
			strategy := pool.Get("compare-test", factory)
			if strategy != nil {
				pool.Put(strategy)
			}
		}
	})

	b.Run("Direct", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = &MockStrategy{name: "compare-test", description: "对比测试策略"}
		}
	})
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestElementLevelStrategy 测试元素级分块策略
func TestElementLevelStrategy(t *testing.T) {
	t.Run("创建策略", func(t *testing.T) {
		strategy := NewElementLevelStrategy()

		if strategy == nil {
			t.Fatal("策略创建失败")
		}

		if strategy.GetName() != "element-level" {
			t.Errorf("策略名称错误，期望 'element-level'，实际 '%s'", strategy.GetName())
		}

		if strategy.GetDescription() == "" {
			t.Error("策略描述不能为空")
		}

		if strategy.config == nil {
			t.Error("策略配置不能为空")
		}
	})

	t.Run("使用配置创建策略", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "element-level",
			MinChunkSize: 10,
			MaxChunkSize: 1000,
		}

		strategy := NewElementLevelStrategyWithConfig(config)

		if strategy == nil {
			t.Fatal("策略创建失败")
		}

		if strategy.config.MinChunkSize != 10 {
			t.Errorf("最小块大小配置错误，期望 10，实际 %d", strategy.config.MinChunkSize)
		}

		if strategy.config.MaxChunkSize != 1000 {
			t.Errorf("最大块大小配置错误，期望 1000，实际 %d", strategy.config.MaxChunkSize)
		}
	})

	t.Run("使用空配置创建策略", func(t *testing.T) {
		strategy := NewElementLevelStrategyWithConfig(nil)

		if strategy == nil {
			t.Fatal("策略创建失败")
		}

		if strategy.config == nil {
			t.Error("应该使用默认配置")
		}

		if strategy.config.Name != "element-level" {
			t.Errorf("默认配置名称错误，期望 'element-level'，实际 '%s'", strategy.config.Name)
		}
	})
}

// TestElementLevelStrategyValidateConfig 测试元素级策略配置验证
func TestElementLevelStrategyValidateConfig(t *testing.T) {
	strategy := NewElementLevelStrategy()

	t.Run("验证有效配置", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "element-level",
			MinChunkSize: 10,
			MaxChunkSize: 1000,
			IncludeTypes: []string{"heading", "paragraph"},
		}

		err := strategy.ValidateConfig(config)
		if err != nil {
			t.Errorf("有效配置验证失败: %v", err)
		}
	})

	t.Run("验证空配置", func(t *testing.T) {
		err := strategy.ValidateConfig(nil)
		if err == nil {
			t.Error("空配置应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证无效配置", func(t *testing.T) {
		config := &StrategyConfig{
			Name: "", // 空名称
		}

		err := strategy.ValidateConfig(config)
		if err == nil {
			t.Error("无效配置应该返回错误")
		}
	})
}

// TestElementLevelStrategyClone 测试元素级策略克隆
func TestElementLevelStrategyClone(t *testing.T) {
	t.Run("克隆策略", func(t *testing.T) {
		original := NewElementLevelStrategy()
		original.config.MinChunkSize = 100
		original.config.MaxChunkSize = 2000

		cloned := original.Clone()

		if cloned == nil {
			t.Fatal("克隆策略失败")
		}

		elementCloned, ok := cloned.(*ElementLevelStrategy)
		if !ok {
			t.Fatal("克隆的策略类型错误")
		}

		if elementCloned.GetName() != original.GetName() {
			t.Error("克隆策略名称不匹配")
		}

		if elementCloned.config.MinChunkSize != original.config.MinChunkSize {
			t.Error("克隆策略配置不匹配")
		}

		// 验证是深拷贝
		elementCloned.config.MinChunkSize = 200
		if original.config.MinChunkSize == 200 {
			t.Error("克隆应该是深拷贝，不应该影响原始对象")
		}
	})

	t.Run("克隆空配置策略", func(t *testing.T) {
		original := &ElementLevelStrategy{config: nil}
		cloned := original.Clone()

		if cloned == nil {
			t.Fatal("克隆策略失败")
		}

		elementCloned, ok := cloned.(*ElementLevelStrategy)
		if !ok {
			t.Fatal("克隆的策略类型错误")
		}

		if elementCloned.config != nil {
			t.Error("克隆的空配置策略应该保持配置为空")
		}
	})
}

// TestElementLevelStrategyShouldIncludeChunk 测试块包含逻辑
func TestElementLevelStrategyShouldIncludeChunk(t *testing.T) {
	t.Run("包含所有类型", func(t *testing.T) {
		strategy := NewElementLevelStrategy()

		chunk := &Chunk{
			Type:    "heading",
			Content: "# Test Heading",
		}

		if !strategy.shouldIncludeChunk(chunk) {
			t.Error("应该包含所有类型的块")
		}
	})

	t.Run("包含指定类型", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "element-level",
			IncludeTypes: []string{"heading", "paragraph"},
		}
		strategy := NewElementLevelStrategyWithConfig(config)

		headingChunk := &Chunk{Type: "heading", Content: "# Test"}
		paragraphChunk := &Chunk{Type: "paragraph", Content: "Test paragraph"}
		codeChunk := &Chunk{Type: "code", Content: "```\ncode\n```"}

		if !strategy.shouldIncludeChunk(headingChunk) {
			t.Error("应该包含标题块")
		}

		if !strategy.shouldIncludeChunk(paragraphChunk) {
			t.Error("应该包含段落块")
		}

		if strategy.shouldIncludeChunk(codeChunk) {
			t.Error("不应该包含代码块")
		}
	})

	t.Run("排除指定类型", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "element-level",
			ExcludeTypes: []string{"code"},
		}
		strategy := NewElementLevelStrategyWithConfig(config)

		headingChunk := &Chunk{Type: "heading", Content: "# Test"}
		codeChunk := &Chunk{Type: "code", Content: "```\ncode\n```"}

		if !strategy.shouldIncludeChunk(headingChunk) {
			t.Error("应该包含标题块")
		}

		if strategy.shouldIncludeChunk(codeChunk) {
			t.Error("不应该包含代码块")
		}
	})

	t.Run("块大小限制", func(t *testing.T) {
		config := &StrategyConfig{
			Name:         "element-level",
			MinChunkSize: 10,
			MaxChunkSize: 50,
		}
		strategy := NewElementLevelStrategyWithConfig(config)

		smallChunk := &Chunk{Type: "paragraph", Content: "Small"}
		normalChunk := &Chunk{Type: "paragraph", Content: "This is a normal sized chunk"}
		largeChunk := &Chunk{Type: "paragraph", Content: "This is a very large chunk that exceeds the maximum size limit"}

		if strategy.shouldIncludeChunk(smallChunk) {
			t.Error("不应该包含太小的块")
		}

		if !strategy.shouldIncludeChunk(normalChunk) {
			t.Error("应该包含正常大小的块")
		}

		if strategy.shouldIncludeChunk(largeChunk) {
			t.Error("不应该包含太大的块")
		}
	})

	t.Run("空块处理", func(t *testing.T) {
		strategy := NewElementLevelStrategy()

		if strategy.shouldIncludeChunk(nil) {
			t.Error("不应该包含空块")
		}
	})
}

// TestHierarchicalStrategy 测试层级分块策略
func TestHierarchicalStrategy(t *testing.T) {
	t.Run("创建策略", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()

		if strategy == nil {
			t.Fatal("策略创建失败")
		}

		if strategy.GetName() != "hierarchical" {
			t.Errorf("策略名称错误，期望 'hierarchical'，实际 '%s'", strategy.GetName())
		}

		if strategy.GetDescription() == "" {
			t.Error("策略描述不能为空")
		}

		if strategy.config == nil {
			t.Error("策略配置不能为空")
		}
	})

	t.Run("使用配置创建策略", func(t *testing.T) {
		config := HierarchicalConfig(3)
		strategy := NewHierarchicalStrategyWithConfig(config)

		if strategy == nil {
			t.Fatal("策略创建失败")
		}

		if strategy.config.MaxDepth != 3 {
			t.Errorf("最大深度配置错误，期望 3，实际 %d", strategy.config.MaxDepth)
		}

		if !strategy.config.MergeEmpty {
			t.Error("应该启用空章节合并")
		}
	})

	t.Run("使用空配置创建策略", func(t *testing.T) {
		strategy := NewHierarchicalStrategyWithConfig(nil)

		if strategy == nil {
			t.Fatal("策略创建失败")
		}

		if strategy.config == nil {
			t.Error("应该使用默认配置")
		}

		if strategy.config.Name != "hierarchical" {
			t.Errorf("默认配置名称错误，期望 'hierarchical'，实际 '%s'", strategy.config.Name)
		}
	})
}

// TestHierarchicalStrategyValidateConfig 测试层级策略配置验证
func TestHierarchicalStrategyValidateConfig(t *testing.T) {
	strategy := NewHierarchicalStrategy()

	t.Run("验证有效配置", func(t *testing.T) {
		config := &StrategyConfig{
			Name:       "hierarchical",
			MaxDepth:   3,
			MinDepth:   1,
			MergeEmpty: true,
		}

		err := strategy.ValidateConfig(config)
		if err != nil {
			t.Errorf("有效配置验证失败: %v", err)
		}
	})

	t.Run("验证空配置", func(t *testing.T) {
		err := strategy.ValidateConfig(nil)
		if err == nil {
			t.Error("空配置应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证层级深度超限", func(t *testing.T) {
		config := &StrategyConfig{
			Name:     "hierarchical",
			MaxDepth: 7, // 超过6的限制
		}

		err := strategy.ValidateConfig(config)
		if err == nil {
			t.Error("超过最大层级深度限制应该返回错误")
		}

		chunkerErr, ok := err.(*ChunkerError)
		if !ok {
			t.Error("应该返回 ChunkerError 类型的错误")
		} else if chunkerErr.Type != ErrorTypeStrategyConfigInvalid {
			t.Errorf("错误类型应该是 ErrorTypeStrategyConfigInvalid，实际是 %v", chunkerErr.Type)
		}
	})

	t.Run("验证其他无效配置", func(t *testing.T) {
		config := &StrategyConfig{
			Name: "", // 空名称
		}

		err := strategy.ValidateConfig(config)
		if err == nil {
			t.Error("无效配置应该返回错误")
		}
	})
}

// TestHierarchicalStrategyClone 测试层级策略克隆
func TestHierarchicalStrategyClone(t *testing.T) {
	t.Run("克隆策略", func(t *testing.T) {
		original := NewHierarchicalStrategy()
		original.config.MaxDepth = 3
		original.config.MergeEmpty = false

		cloned := original.Clone()

		if cloned == nil {
			t.Fatal("克隆策略失败")
		}

		hierarchicalCloned, ok := cloned.(*HierarchicalStrategy)
		if !ok {
			t.Fatal("克隆的策略类型错误")
		}

		if hierarchicalCloned.GetName() != original.GetName() {
			t.Error("克隆策略名称不匹配")
		}

		if hierarchicalCloned.config.MaxDepth != original.config.MaxDepth {
			t.Error("克隆策略配置不匹配")
		}

		// 验证是深拷贝
		hierarchicalCloned.config.MaxDepth = 5
		if original.config.MaxDepth == 5 {
			t.Error("克隆应该是深拷贝，不应该影响原始对象")
		}
	})

	t.Run("克隆空配置策略", func(t *testing.T) {
		original := &HierarchicalStrategy{config: nil}
		cloned := original.Clone()

		if cloned == nil {
			t.Fatal("克隆策略失败")
		}

		hierarchicalCloned, ok := cloned.(*HierarchicalStrategy)
		if !ok {
			t.Fatal("克隆的策略类型错误")
		}

		if hierarchicalCloned.config != nil {
			t.Error("克隆的空配置策略应该保持配置为空")
		}
	})
}

// TestHierarchicalStrategyBuildHierarchy 测试层级结构构建
func TestHierarchicalStrategyBuildHierarchy(t *testing.T) {
	strategy := NewHierarchicalStrategy()

	t.Run("空块列表", func(t *testing.T) {
		result := strategy.buildHierarchy([]Chunk{})
		if len(result) != 0 {
			t.Errorf("空块列表应该返回空结果，实际返回 %d 个块", len(result))
		}
	})

	t.Run("简单层级结构", func(t *testing.T) {
		chunks := []Chunk{
			{ID: 0, Type: "heading", Level: 1, Text: "Chapter 1", Content: "# Chapter 1"},
			{ID: 1, Type: "paragraph", Text: "Introduction", Content: "Introduction"},
			{ID: 2, Type: "heading", Level: 2, Text: "Section 1.1", Content: "## Section 1.1"},
			{ID: 3, Type: "paragraph", Text: "Section content", Content: "Section content"},
		}

		result := strategy.buildHierarchy(chunks)

		if len(result) != 1 {
			t.Errorf("应该有1个根级块，实际有 %d 个", len(result))
		}

		root := result[0]
		if root.Chunk.Type != "heading" || root.Chunk.Level != 1 {
			t.Error("根块应该是1级标题")
		}

		if len(root.Children) != 2 {
			t.Errorf("根块应该有2个子块，实际有 %d 个", len(root.Children))
		}

		// 检查第一个子块（段落）
		if root.Children[0].Chunk.Type != "paragraph" {
			t.Error("第一个子块应该是段落")
		}

		// 检查第二个子块（2级标题）
		if root.Children[1].Chunk.Type != "heading" || root.Children[1].Chunk.Level != 2 {
			t.Error("第二个子块应该是2级标题")
		}

		// 检查2级标题的子块
		if len(root.Children[1].Children) != 1 {
			t.Errorf("2级标题应该有1个子块，实际有 %d 个", len(root.Children[1].Children))
		}

		if root.Children[1].Children[0].Chunk.Type != "paragraph" {
			t.Error("2级标题的子块应该是段落")
		}
	})

	t.Run("跳跃层级处理", func(t *testing.T) {
		chunks := []Chunk{
			{ID: 0, Type: "heading", Level: 1, Text: "Chapter 1", Content: "# Chapter 1"},
			{ID: 1, Type: "heading", Level: 3, Text: "Subsection", Content: "### Subsection"}, // 跳过了2级
		}

		result := strategy.buildHierarchy(chunks)

		if len(result) != 1 {
			t.Errorf("应该有1个根级块，实际有 %d 个", len(result))
		}

		root := result[0]
		if len(root.Children) != 1 {
			t.Errorf("根块应该有1个子块（虚拟2级标题），实际有 %d 个", len(root.Children))
		}

		// 检查虚拟2级标题
		virtualH2 := root.Children[0]
		if virtualH2.Chunk.Level != 2 {
			t.Errorf("虚拟块应该是2级标题，实际是 %d 级", virtualH2.Chunk.Level)
		}

		if virtualH2.Chunk.Metadata["virtual"] != "true" {
			t.Error("应该标记为虚拟块")
		}

		// 检查3级标题
		if len(virtualH2.Children) != 1 {
			t.Errorf("虚拟2级标题应该有1个子块，实际有 %d 个", len(virtualH2.Children))
		}

		if virtualH2.Children[0].Chunk.Level != 3 {
			t.Error("子块应该是3级标题")
		}
	})

	t.Run("无标题内容处理", func(t *testing.T) {
		chunks := []Chunk{
			{ID: 0, Type: "paragraph", Text: "Opening paragraph", Content: "Opening paragraph"},
			{ID: 1, Type: "heading", Level: 1, Text: "Chapter 1", Content: "# Chapter 1"},
			{ID: 2, Type: "paragraph", Text: "Chapter content", Content: "Chapter content"},
		}

		result := strategy.buildHierarchy(chunks)

		if len(result) != 2 {
			t.Errorf("应该有2个根级块，实际有 %d 个", len(result))
		}

		// 第一个应该是虚拟根块，包含开头的段落
		virtualRoot := result[0]
		if virtualRoot.Chunk.Type != "root" {
			t.Error("第一个块应该是虚拟根块")
		}

		if len(virtualRoot.Children) != 1 {
			t.Errorf("虚拟根块应该有1个子块，实际有 %d 个", len(virtualRoot.Children))
		}

		if virtualRoot.Children[0].Chunk.Type != "paragraph" {
			t.Error("虚拟根块的子块应该是段落")
		}

		// 第二个应该是1级标题
		chapter := result[1]
		if chapter.Chunk.Type != "heading" || chapter.Chunk.Level != 1 {
			t.Error("第二个块应该是1级标题")
		}

		if len(chapter.Children) != 1 {
			t.Errorf("1级标题应该有1个子块，实际有 %d 个", len(chapter.Children))
		}
	})

	t.Run("层级深度计算", func(t *testing.T) {
		chunks := []Chunk{
			{ID: 0, Type: "heading", Level: 1, Text: "H1", Content: "# H1"},
			{ID: 1, Type: "heading", Level: 2, Text: "H2", Content: "## H2"},
			{ID: 2, Type: "heading", Level: 3, Text: "H3", Content: "### H3"},
			{ID: 3, Type: "paragraph", Text: "Content", Content: "Content"},
		}

		result := strategy.buildHierarchy(chunks)

		// 验证层级深度
		if result[0].Level != 0 {
			t.Errorf("H1的层级深度应该是0，实际是 %d", result[0].Level)
		}

		h2 := result[0].Children[0]
		if h2.Level != 1 {
			t.Errorf("H2的层级深度应该是1，实际是 %d", h2.Level)
		}

		h3 := h2.Children[0]
		if h3.Level != 2 {
			t.Errorf("H3的层级深度应该是2，实际是 %d", h3.Level)
		}

		paragraph := h3.Children[0]
		if paragraph.Level != 3 {
			t.Errorf("段落的层级深度应该是3，实际是 %d", paragraph.Level)
		}
	})
}

// TestHierarchicalStrategyFlattenToTargetLevel 测试扁平化到目标层级
func TestHierarchicalStrategyFlattenToTargetLevel(t *testing.T) {
	strategy := NewHierarchicalStrategy()

	t.Run("无层级限制扁平化", func(t *testing.T) {
		// 构建简单的层级结构
		hierarchicalChunks := []*HierarchicalChunk{
			{
				Chunk: Chunk{ID: 0, Type: "heading", Level: 1, Text: "Chapter 1", Content: "# Chapter 1"},
				Level: 0,
				Children: []*HierarchicalChunk{
					{
						Chunk: Chunk{ID: 1, Type: "paragraph", Text: "Introduction", Content: "Introduction"},
						Level: 1,
					},
				},
			},
		}

		result := strategy.flattenToTargetLevel(hierarchicalChunks)

		if len(result) != 1 {
			t.Errorf("应该生成1个块，实际生成 %d 个", len(result))
		}

		chunk := result[0]
		if chunk.Type != "heading" {
			t.Errorf("块类型应该是heading，实际是 %s", chunk.Type)
		}

		if chunk.Metadata["strategy"] != "hierarchical" {
			t.Error("应该标记为层级策略生成")
		}

		if chunk.Metadata["merged_chunks"] != "2" {
			t.Errorf("应该合并2个块，实际合并 %s 个", chunk.Metadata["merged_chunks"])
		}
	})

	t.Run("最大深度限制", func(t *testing.T) {
		strategy.config.MaxDepth = 1

		// 构建多层级结构
		hierarchicalChunks := []*HierarchicalChunk{
			{
				Chunk: Chunk{ID: 0, Type: "heading", Level: 1, Text: "Chapter 1", Content: "# Chapter 1"},
				Level: 0,
				Children: []*HierarchicalChunk{
					{
						Chunk: Chunk{ID: 1, Type: "heading", Level: 2, Text: "Section 1.1", Content: "## Section 1.1"},
						Level: 1,
						Children: []*HierarchicalChunk{
							{
								Chunk: Chunk{ID: 2, Type: "paragraph", Text: "Content", Content: "Content"},
								Level: 2,
							},
						},
					},
				},
			},
		}

		result := strategy.flattenToTargetLevel(hierarchicalChunks)

		// 由于最大深度为1，应该在深度1处创建块
		if len(result) != 1 {
			t.Errorf("应该生成1个块，实际生成 %d 个", len(result))
		}

		// 重置配置
		strategy.config.MaxDepth = 0
	})

	t.Run("虚拟块处理", func(t *testing.T) {
		// 构建包含虚拟块的结构
		hierarchicalChunks := []*HierarchicalChunk{
			{
				Chunk: Chunk{
					ID:       -1,
					Type:     "heading",
					Level:    2,
					Metadata: map[string]string{"virtual": "true"},
				},
				Level: 0,
				Children: []*HierarchicalChunk{
					{
						Chunk: Chunk{ID: 1, Type: "paragraph", Text: "Real content", Content: "Real content"},
						Level: 1,
					},
				},
			},
		}

		result := strategy.flattenToTargetLevel(hierarchicalChunks)

		if len(result) != 1 {
			t.Errorf("应该生成1个块，实际生成 %d 个", len(result))
		}

		// 虚拟块应该被跳过，直接处理实际内容
		chunk := result[0]
		if chunk.Type != "hierarchical" {
			t.Errorf("块类型应该是hierarchical，实际是 %s", chunk.Type)
		}
	})
}

// TestHierarchicalStrategyHelperMethods 测试辅助方法
func TestHierarchicalStrategyHelperMethods(t *testing.T) {
	strategy := NewHierarchicalStrategy()

	t.Run("isVirtualChunk", func(t *testing.T) {
		virtualChunk := &HierarchicalChunk{
			Chunk: Chunk{
				Metadata: map[string]string{"virtual": "true"},
			},
		}

		realChunk := &HierarchicalChunk{
			Chunk: Chunk{
				Metadata: map[string]string{},
			},
		}

		if !strategy.isVirtualChunk(virtualChunk) {
			t.Error("应该识别为虚拟块")
		}

		if strategy.isVirtualChunk(realChunk) {
			t.Error("不应该识别为虚拟块")
		}

		if strategy.isVirtualChunk(nil) {
			t.Error("空块不应该识别为虚拟块")
		}
	})

	t.Run("hasNonEmptyContent", func(t *testing.T) {
		emptyChunk := &HierarchicalChunk{
			Chunk: Chunk{Text: ""},
		}

		nonEmptyChunk := &HierarchicalChunk{
			Chunk: Chunk{Text: "Some content"},
		}

		chunkWithNonEmptyChild := &HierarchicalChunk{
			Chunk: Chunk{Text: ""},
			Children: []*HierarchicalChunk{
				{Chunk: Chunk{Text: "Child content"}},
			},
		}

		if strategy.hasNonEmptyContent(emptyChunk) {
			t.Error("空块不应该有内容")
		}

		if !strategy.hasNonEmptyContent(nonEmptyChunk) {
			t.Error("非空块应该有内容")
		}

		if !strategy.hasNonEmptyContent(chunkWithNonEmptyChild) {
			t.Error("有非空子块的块应该有内容")
		}

		if strategy.hasNonEmptyContent(nil) {
			t.Error("空块不应该有内容")
		}
	})

	t.Run("countTotalChunks", func(t *testing.T) {
		singleChunk := &HierarchicalChunk{
			Chunk: Chunk{ID: 1},
		}

		chunkWithChildren := &HierarchicalChunk{
			Chunk: Chunk{ID: 1},
			Children: []*HierarchicalChunk{
				{Chunk: Chunk{ID: 2}},
				{
					Chunk: Chunk{ID: 3},
					Children: []*HierarchicalChunk{
						{Chunk: Chunk{ID: 4}},
					},
				},
			},
		}

		if strategy.countTotalChunks(singleChunk) != 1 {
			t.Errorf("单个块应该计数为1，实际为 %d", strategy.countTotalChunks(singleChunk))
		}

		if strategy.countTotalChunks(chunkWithChildren) != 4 {
			t.Errorf("有子块的块应该计数为4，实际为 %d", strategy.countTotalChunks(chunkWithChildren))
		}

		if strategy.countTotalChunks(nil) != 0 {
			t.Errorf("空块应该计数为0，实际为 %d", strategy.countTotalChunks(nil))
		}
	})

	t.Run("getMaxDepthInHierarchy", func(t *testing.T) {
		hierarchicalChunks := []*HierarchicalChunk{
			{
				Level: 0,
				Children: []*HierarchicalChunk{
					{
						Level: 1,
						Children: []*HierarchicalChunk{
							{Level: 2},
						},
					},
				},
			},
		}

		maxDepth := strategy.getMaxDepthInHierarchy(hierarchicalChunks)
		if maxDepth != 2 {
			t.Errorf("最大深度应该是2，实际是 %d", maxDepth)
		}

		emptyHierarchy := []*HierarchicalChunk{}
		maxDepth = strategy.getMaxDepthInHierarchy(emptyHierarchy)
		if maxDepth != 0 {
			t.Errorf("空层级的最大深度应该是0，实际是 %d", maxDepth)
		}
	})
}

// TestHierarchicalStrategyChunkDocument 测试文档分块
func TestHierarchicalStrategyChunkDocument(t *testing.T) {
	t.Run("空文档处理", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()
		chunker := NewMarkdownChunker()

		_, err := strategy.ChunkDocument(nil, []byte{}, chunker)
		if err == nil {
			t.Error("空文档节点应该返回错误")
		}
	})

	t.Run("空分块器处理", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()

		// 创建一个简单的AST节点用于测试
		chunker := NewMarkdownChunker()
		reader := text.NewReader([]byte("# Test"))
		doc := chunker.md.Parser().Parse(reader)

		_, err := strategy.ChunkDocument(doc, []byte("# Test"), nil)
		if err == nil {
			t.Error("空分块器应该返回错误")
		}
	})

	t.Run("正常文档分块", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()
		chunker := NewMarkdownChunker()

		content := []byte(`# Chapter 1

Introduction paragraph.

## Section 1.1

Section content.`)

		reader := text.NewReader(content)
		doc := chunker.md.Parser().Parse(reader)

		chunks, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			t.Fatalf("文档分块失败: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("应该生成至少一个块")
		}

		// 验证块的基本属性
		for i, chunk := range chunks {
			if chunk.Metadata["strategy"] != "hierarchical" {
				t.Errorf("块 %d 应该标记为层级策略生成", i)
			}

			if chunk.Hash == "" {
				t.Errorf("块 %d 应该有哈希值", i)
			}
		}
	})
}

// TestHierarchicalStrategyConfigMethods 测试配置方法
func TestHierarchicalStrategyConfigMethods(t *testing.T) {
	t.Run("GetConfig", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()
		strategy.config.MaxDepth = 3

		config := strategy.GetConfig()
		if config == nil {
			t.Fatal("获取配置失败")
		}

		if config.MaxDepth != 3 {
			t.Errorf("配置最大深度不匹配，期望 3，实际 %d", config.MaxDepth)
		}

		// 验证是深拷贝
		config.MaxDepth = 5
		if strategy.config.MaxDepth == 5 {
			t.Error("获取的配置应该是深拷贝")
		}
	})

	t.Run("GetConfig空配置", func(t *testing.T) {
		strategy := &HierarchicalStrategy{config: nil}
		config := strategy.GetConfig()
		if config != nil {
			t.Error("空配置策略应该返回nil")
		}
	})

	t.Run("SetConfig", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()

		newConfig := &StrategyConfig{
			Name:       "hierarchical",
			MaxDepth:   4,
			MergeEmpty: false,
		}

		err := strategy.SetConfig(newConfig)
		if err != nil {
			t.Errorf("设置配置失败: %v", err)
		}

		if strategy.config.MaxDepth != 4 {
			t.Errorf("配置设置失败，期望 4，实际 %d", strategy.config.MaxDepth)
		}

		if strategy.config.MergeEmpty {
			t.Error("配置设置失败，MergeEmpty应该为false")
		}

		// 验证是深拷贝
		newConfig.MaxDepth = 6
		if strategy.config.MaxDepth == 6 {
			t.Error("设置的配置应该是深拷贝")
		}
	})

	t.Run("SetConfig空配置", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()

		err := strategy.SetConfig(nil)
		if err != nil {
			t.Errorf("设置空配置失败: %v", err)
		}

		if strategy.config == nil {
			t.Error("设置空配置后应该使用默认配置")
		}

		if strategy.config.Name != "hierarchical" {
			t.Error("默认配置名称应该是hierarchical")
		}
	})

	t.Run("SetConfig无效配置", func(t *testing.T) {
		strategy := NewHierarchicalStrategy()

		invalidConfig := &StrategyConfig{
			Name:     "hierarchical",
			MaxDepth: 10, // 超过限制
		}

		err := strategy.SetConfig(invalidConfig)
		if err == nil {
			t.Error("设置无效配置应该返回错误")
		}
	})
}

// BenchmarkHierarchicalStrategy 层级策略性能基准测试
func BenchmarkHierarchicalStrategy(b *testing.B) {
	strategy := NewHierarchicalStrategy()
	chunker := NewMarkdownChunker()

	content := []byte(`# Chapter 1

Introduction to the chapter.

## Section 1.1

Content of section 1.1.

### Subsection 1.1.1

Detailed content here.

## Section 1.2

Content of section 1.2.

# Chapter 2

Second chapter content.

## Section 2.1

More content here.`)

	reader := text.NewReader(content)
	doc := chunker.md.Parser().Parse(reader)

	for b.Loop() {
		_, err := strategy.ChunkDocument(doc, content, chunker)
		if err != nil {
			b.Fatalf("分块失败: %v", err)
		}
	}
}

// BenchmarkHierarchicalStrategyBuildHierarchy 层级构建性能基准测试
func BenchmarkHierarchicalStrategyBuildHierarchy(b *testing.B) {
	strategy := NewHierarchicalStrategy()

	// 创建大量块用于测试
	chunks := make([]Chunk, 1000)
	for i := range 1000 {
		if i%10 == 0 {
			chunks[i] = Chunk{
				ID:      i,
				Type:    "heading",
				Level:   (i/10)%6 + 1,
				Text:    fmt.Sprintf("Heading %d", i),
				Content: fmt.Sprintf("# Heading %d", i),
			}
		} else {
			chunks[i] = Chunk{
				ID:      i,
				Type:    "paragraph",
				Text:    fmt.Sprintf("Paragraph %d", i),
				Content: fmt.Sprintf("Paragraph %d", i),
			}
		}
	}

	for b.Loop() {
		_ = strategy.buildHierarchy(chunks)
	}
}

// BenchmarkHierarchicalStrategyFlatten 扁平化性能基准测试
func BenchmarkHierarchicalStrategyFlatten(b *testing.B) {
	strategy := NewHierarchicalStrategy()

	// 创建复杂的层级结构
	var createHierarchy func(depth, maxDepth int) []*HierarchicalChunk
	createHierarchy = func(depth, maxDepth int) []*HierarchicalChunk {
		if depth >= maxDepth {
			return []*HierarchicalChunk{}
		}

		var chunks []*HierarchicalChunk
		for i := range 5 {
			chunk := &HierarchicalChunk{
				Chunk: Chunk{
					ID:      depth*10 + i,
					Type:    "heading",
					Level:   depth + 1,
					Text:    fmt.Sprintf("Heading %d-%d", depth, i),
					Content: fmt.Sprintf("# Heading %d-%d", depth, i),
				},
				Level:    depth,
				Children: createHierarchy(depth+1, maxDepth),
			}
			chunks = append(chunks, chunk)
		}
		return chunks
	}

	hierarchicalChunks := createHierarchy(0, 4)

	for b.Loop() {
		_ = strategy.flattenToTargetLevel(hierarchicalChunks)
	}
}
