package markdownchunker

import (
	"testing"
)

func TestSetStrategy(t *testing.T) {
	tests := []struct {
		name         string
		strategyName string
		config       *StrategyConfig
		wantErr      bool
		expectedName string
	}{
		{
			name:         "set element-level strategy",
			strategyName: "element-level",
			config:       ElementLevelConfig(),
			wantErr:      false,
			expectedName: "element-level",
		},
		{
			name:         "set hierarchical strategy",
			strategyName: "hierarchical",
			config:       HierarchicalConfig(3),
			wantErr:      false,
			expectedName: "hierarchical",
		},
		{
			name:         "set document-level strategy",
			strategyName: "document-level",
			config:       DocumentLevelConfig(),
			wantErr:      false,
			expectedName: "document-level",
		},
		{
			name:         "set strategy without config",
			strategyName: "element-level",
			config:       nil,
			wantErr:      false,
			expectedName: "element-level",
		},
		{
			name:         "set non-existent strategy",
			strategyName: "non-existent",
			config:       nil,
			wantErr:      true,
			expectedName: "",
		},
		{
			name:         "set strategy with empty name",
			strategyName: "",
			config:       nil,
			wantErr:      true,
			expectedName: "",
		},
		{
			name:         "set strategy with invalid config",
			strategyName: "hierarchical",
			config: &StrategyConfig{
				Name:     "hierarchical",
				MaxDepth: 10, // 超过限制
			},
			wantErr:      true,
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()
			originalStrategy, _ := chunker.GetCurrentStrategy()

			err := chunker.SetStrategy(tt.strategyName, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				currentStrategy, _ := chunker.GetCurrentStrategy()
				if currentStrategy != tt.expectedName {
					t.Errorf("Expected strategy %s, got %s", tt.expectedName, currentStrategy)
				}

				// 验证配置是否正确设置
				if tt.config != nil {
					strategyConfig := chunker.GetStrategyConfig()
					if strategyConfig == nil {
						t.Error("Strategy config should not be nil")
					} else if strategyConfig.Name != tt.expectedName {
						t.Errorf("Expected config name %s, got %s", tt.expectedName, strategyConfig.Name)
					}
				}
			} else {
				// 错误情况下，策略应该保持不变
				currentStrategy, _ := chunker.GetCurrentStrategy()
				if currentStrategy != originalStrategy {
					t.Errorf("Strategy should not change on error, expected %s, got %s", originalStrategy, currentStrategy)
				}
			}
		})
	}
}

func TestGetCurrentStrategy(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 测试默认策略
	name, description := chunker.GetCurrentStrategy()
	if name != "element-level" {
		t.Errorf("Expected default strategy to be 'element-level', got '%s'", name)
	}
	if description == "" {
		t.Error("Strategy description should not be empty")
	}

	// 测试切换策略后的获取
	err := chunker.SetStrategy("hierarchical", HierarchicalConfig(2))
	if err != nil {
		t.Fatalf("SetStrategy() error = %v", err)
	}

	name, description = chunker.GetCurrentStrategy()
	if name != "hierarchical" {
		t.Errorf("Expected strategy to be 'hierarchical', got '%s'", name)
	}
	if description == "" {
		t.Error("Strategy description should not be empty")
	}
}

func TestGetAvailableStrategies(t *testing.T) {
	chunker := NewMarkdownChunker()

	strategies := chunker.GetAvailableStrategies()
	if len(strategies) == 0 {
		t.Error("Should have available strategies")
	}

	// 验证包含预期的策略
	expectedStrategies := []string{"element-level", "hierarchical", "document-level"}
	for _, expected := range expectedStrategies {
		found := false
		for _, available := range strategies {
			if available == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected strategy %s to be available, got: %v", expected, strategies)
		}
	}
}

func TestGetStrategyConfig(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 测试默认配置
	config := chunker.GetStrategyConfig()
	if config == nil {
		t.Fatal("Strategy config should not be nil")
	}
	if config.Name != "element-level" {
		t.Errorf("Expected config name to be 'element-level', got '%s'", config.Name)
	}

	// 测试设置配置后的获取
	hierarchicalConfig := HierarchicalConfig(3)
	err := chunker.SetStrategy("hierarchical", hierarchicalConfig)
	if err != nil {
		t.Fatalf("SetStrategy() error = %v", err)
	}

	config = chunker.GetStrategyConfig()
	if config == nil {
		t.Fatal("Strategy config should not be nil")
	}
	if config.Name != "hierarchical" {
		t.Errorf("Expected config name to be 'hierarchical', got '%s'", config.Name)
	}
	if config.MaxDepth != 3 {
		t.Errorf("Expected max depth to be 3, got %d", config.MaxDepth)
	}
}

func TestUpdateStrategyConfig(t *testing.T) {
	tests := []struct {
		name            string
		initialStrategy string
		initialConfig   *StrategyConfig
		updateConfig    *StrategyConfig
		wantErr         bool
	}{
		{
			name:            "update hierarchical config",
			initialStrategy: "hierarchical",
			initialConfig:   HierarchicalConfig(2),
			updateConfig:    HierarchicalConfig(4),
			wantErr:         false,
		},
		{
			name:            "update with nil config",
			initialStrategy: "element-level",
			initialConfig:   ElementLevelConfig(),
			updateConfig:    nil,
			wantErr:         false,
		},
		{
			name:            "update with invalid config",
			initialStrategy: "hierarchical",
			initialConfig:   HierarchicalConfig(2),
			updateConfig: &StrategyConfig{
				Name:     "hierarchical",
				MaxDepth: 10, // 超过限制
			},
			wantErr: true,
		},
		{
			name:            "update with mismatched strategy name",
			initialStrategy: "element-level",
			initialConfig:   ElementLevelConfig(),
			updateConfig: &StrategyConfig{
				Name: "hierarchical", // 与当前策略不匹配
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunker()

			// 设置初始策略
			err := chunker.SetStrategy(tt.initialStrategy, tt.initialConfig)
			if err != nil {
				t.Fatalf("SetStrategy() error = %v", err)
			}

			// 更新配置
			err = chunker.UpdateStrategyConfig(tt.updateConfig)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStrategyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证配置是否正确更新
				config := chunker.GetStrategyConfig()
				if config == nil {
					t.Error("Strategy config should not be nil")
					return
				}

				if config.Name != tt.initialStrategy {
					t.Errorf("Expected config name %s, got %s", tt.initialStrategy, config.Name)
				}

				if tt.updateConfig != nil && tt.updateConfig.MaxDepth > 0 {
					if config.MaxDepth != tt.updateConfig.MaxDepth {
						t.Errorf("Expected max depth %d, got %d", tt.updateConfig.MaxDepth, config.MaxDepth)
					}
				}
			}
		})
	}
}

func TestRegisterStrategy(t *testing.T) {
	chunker := NewMarkdownChunker()
	initialCount := chunker.GetStrategyCount()

	// 创建一个简单的测试策略
	testStrategy := &ElementLevelStrategy{
		config: &StrategyConfig{
			Name: "test-strategy",
		},
	}

	// 重写 GetName 方法以返回测试名称
	testStrategy.config.Name = "test-strategy"

	// 注册策略
	err := chunker.RegisterStrategy(testStrategy)
	if err != nil {
		t.Errorf("RegisterStrategy() error = %v", err)
	}

	// 验证策略数量增加
	newCount := chunker.GetStrategyCount()
	if newCount != initialCount+1 {
		t.Errorf("Expected strategy count to increase by 1, got %d -> %d", initialCount, newCount)
	}

	// 验证策略可用
	if !chunker.HasStrategy("test-strategy") {
		t.Error("Test strategy should be available")
	}

	// 测试注册 nil 策略
	err = chunker.RegisterStrategy(nil)
	if err == nil {
		t.Error("RegisterStrategy(nil) should return error")
	}
}

func TestUnregisterStrategy(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 创建并注册测试策略
	testStrategy := &ElementLevelStrategy{
		config: &StrategyConfig{
			Name: "test-strategy",
		},
	}
	testStrategy.config.Name = "test-strategy"

	err := chunker.RegisterStrategy(testStrategy)
	if err != nil {
		t.Fatalf("RegisterStrategy() error = %v", err)
	}

	initialCount := chunker.GetStrategyCount()

	// 注销策略
	err = chunker.UnregisterStrategy("test-strategy")
	if err != nil {
		t.Errorf("UnregisterStrategy() error = %v", err)
	}

	// 验证策略数量减少
	newCount := chunker.GetStrategyCount()
	if newCount != initialCount-1 {
		t.Errorf("Expected strategy count to decrease by 1, got %d -> %d", initialCount, newCount)
	}

	// 验证策略不再可用
	if chunker.HasStrategy("test-strategy") {
		t.Error("Test strategy should not be available after unregistering")
	}

	// 测试注销当前使用的策略
	currentStrategy, _ := chunker.GetCurrentStrategy()
	err = chunker.UnregisterStrategy(currentStrategy)
	if err == nil {
		t.Error("UnregisterStrategy() should return error when trying to unregister current strategy")
	}

	// 测试注销不存在的策略
	err = chunker.UnregisterStrategy("non-existent")
	if err == nil {
		t.Error("UnregisterStrategy() should return error for non-existent strategy")
	}

	// 测试注销空名称策略
	err = chunker.UnregisterStrategy("")
	if err == nil {
		t.Error("UnregisterStrategy() should return error for empty strategy name")
	}
}

func TestHasStrategy(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 测试默认策略
	if !chunker.HasStrategy("element-level") {
		t.Error("Should have element-level strategy")
	}
	if !chunker.HasStrategy("hierarchical") {
		t.Error("Should have hierarchical strategy")
	}
	if !chunker.HasStrategy("document-level") {
		t.Error("Should have document-level strategy")
	}

	// 测试不存在的策略
	if chunker.HasStrategy("non-existent") {
		t.Error("Should not have non-existent strategy")
	}
}

func TestGetStrategyCount(t *testing.T) {
	chunker := NewMarkdownChunker()

	count := chunker.GetStrategyCount()
	if count < 3 {
		t.Errorf("Expected at least 3 strategies, got %d", count)
	}

	// 注册新策略后验证数量增加
	testStrategy := &ElementLevelStrategy{
		config: &StrategyConfig{
			Name: "test-strategy",
		},
	}
	testStrategy.config.Name = "test-strategy"

	err := chunker.RegisterStrategy(testStrategy)
	if err != nil {
		t.Fatalf("RegisterStrategy() error = %v", err)
	}

	newCount := chunker.GetStrategyCount()
	if newCount != count+1 {
		t.Errorf("Expected count to increase by 1, got %d -> %d", count, newCount)
	}
}
