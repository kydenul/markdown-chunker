package markdownchunker

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.MaxChunkSize != 0 {
		t.Errorf("Expected MaxChunkSize to be 0, got %d", config.MaxChunkSize)
	}

	if config.EnabledTypes != nil {
		t.Errorf("Expected EnabledTypes to be nil, got %v", config.EnabledTypes)
	}

	if config.ErrorHandling != ErrorModePermissive {
		t.Errorf("Expected ErrorHandling to be ErrorModePermissive, got %v", config.ErrorHandling)
	}

	if config.PerformanceMode != PerformanceModeDefault {
		t.Errorf("Expected PerformanceMode to be PerformanceModeDefault, got %v", config.PerformanceMode)
	}

	if !config.FilterEmptyChunks {
		t.Error("Expected FilterEmptyChunks to be true")
	}

	if config.PreserveWhitespace {
		t.Error("Expected PreserveWhitespace to be false")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ChunkerConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "negative MaxChunkSize",
			config: &ChunkerConfig{
				MaxChunkSize: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid content type",
			config: &ChunkerConfig{
				EnabledTypes: map[string]bool{
					"invalid_type": true,
				},
			},
			wantErr: true,
		},
		{
			name: "valid content types",
			config: &ChunkerConfig{
				EnabledTypes: map[string]bool{
					"heading":   true,
					"paragraph": true,
					"code":      true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMarkdownChunkerWithConfig(t *testing.T) {
	config := &ChunkerConfig{
		MaxChunkSize: 1000,
		EnabledTypes: map[string]bool{
			"heading":   true,
			"paragraph": true,
		},
		ErrorHandling: ErrorModeStrict,
	}

	chunker := NewMarkdownChunkerWithConfig(config)

	if chunker == nil {
		t.Fatal("NewMarkdownChunkerWithConfig() returned nil")
	}

	if chunker.config != config {
		t.Error("Config was not set correctly")
	}

	if chunker.md == nil {
		t.Error("Markdown parser was not initialized")
	}
}

func TestNewMarkdownChunkerWithNilConfig(t *testing.T) {
	chunker := NewMarkdownChunkerWithConfig(nil)

	if chunker == nil {
		t.Fatal("NewMarkdownChunkerWithConfig(nil) returned nil")
	}

	// 应该使用默认配置
	if chunker.config == nil {
		t.Error("Config should not be nil when passing nil config")
	}

	if chunker.config.ErrorHandling != ErrorModePermissive {
		t.Error("Should use default config when nil is passed")
	}
}

func TestIsTypeEnabled(t *testing.T) {
	tests := []struct {
		name         string
		enabledTypes map[string]bool
		chunkType    string
		expected     bool
	}{
		{
			name:         "nil enabled types - all enabled",
			enabledTypes: nil,
			chunkType:    "heading",
			expected:     true,
		},
		{
			name: "type explicitly enabled",
			enabledTypes: map[string]bool{
				"heading":   true,
				"paragraph": false,
			},
			chunkType: "heading",
			expected:  true,
		},
		{
			name: "type explicitly disabled",
			enabledTypes: map[string]bool{
				"heading":   true,
				"paragraph": false,
			},
			chunkType: "paragraph",
			expected:  false,
		},
		{
			name: "type not in map",
			enabledTypes: map[string]bool{
				"heading": true,
			},
			chunkType: "paragraph",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.EnabledTypes = tt.enabledTypes

			chunker := NewMarkdownChunkerWithConfig(config)
			result := chunker.isTypeEnabled(tt.chunkType)

			if result != tt.expected {
				t.Errorf("isTypeEnabled() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestChunkDocumentWithConfig(t *testing.T) {
	markdown := `# Heading 1

This is a paragraph.

## Heading 2

Another paragraph.

- List item 1
- List item 2`

	t.Run("filter by type", func(t *testing.T) {
		config := DefaultConfig()
		config.EnabledTypes = map[string]bool{
			"heading":   true,
			"paragraph": false,
			"list":      false,
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		// 应该只有标题块
		for _, chunk := range chunks {
			if chunk.Type != "heading" {
				t.Errorf("Expected only heading chunks, got %s", chunk.Type)
			}
		}
	})

	t.Run("max chunk size", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxChunkSize = 10 // 很小的限制
		config.ErrorHandling = ErrorModePermissive

		chunker := NewMarkdownChunkerWithConfig(config)
		chunks, err := chunker.ChunkDocument([]byte(markdown))
		if err != nil {
			t.Fatalf("ChunkDocument() error = %v", err)
		}

		// 检查所有块的大小
		for _, chunk := range chunks {
			if len(chunk.Content) > 10 {
				t.Errorf("Chunk content size %d exceeds limit 10", len(chunk.Content))
			}
		}
	})

	t.Run("strict mode with size limit", func(t *testing.T) {
		config := DefaultConfig()
		config.MaxChunkSize = 5 // 非常小的限制
		config.ErrorHandling = ErrorModeStrict

		chunker := NewMarkdownChunkerWithConfig(config)
		_, err := chunker.ChunkDocument([]byte(markdown))

		if err == nil {
			t.Error("Expected error in strict mode with small chunk size limit")
		}
	})
}

func TestValidateConfigWithStrategy(t *testing.T) {
	tests := []struct {
		name    string
		config  *ChunkerConfig
		wantErr bool
	}{
		{
			name: "valid element-level strategy",
			config: &ChunkerConfig{
				ChunkingStrategy: ElementLevelConfig(),
			},
			wantErr: false,
		},
		{
			name: "valid hierarchical strategy",
			config: &ChunkerConfig{
				ChunkingStrategy: HierarchicalConfig(3),
			},
			wantErr: false,
		},
		{
			name: "valid document-level strategy",
			config: &ChunkerConfig{
				ChunkingStrategy: DocumentLevelConfig(),
			},
			wantErr: false,
		},
		{
			name: "invalid strategy - empty name",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name: "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid hierarchical strategy - max depth too high",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name:     "hierarchical",
					MaxDepth: 10, // 超过6的限制
				},
			},
			wantErr: true,
		},
		{
			name: "invalid element-level strategy - has depth config",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name:     "element-level",
					MaxDepth: 3, // 元素级策略不应该有层级配置
				},
			},
			wantErr: true,
		},
		{
			name: "invalid document-level strategy - has type filters",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name:         "document-level",
					IncludeTypes: []string{"heading"}, // 文档级策略不应该有类型过滤
				},
			},
			wantErr: true,
		},
		{
			name: "nil strategy config",
			config: &ChunkerConfig{
				ChunkingStrategy: nil, // 应该使用默认策略
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsureDefaultStrategyConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         *ChunkerConfig
		expectedName   string
		expectedParams bool
	}{
		{
			name:           "nil config",
			config:         nil,
			expectedName:   "",
			expectedParams: false,
		},
		{
			name: "nil strategy config",
			config: &ChunkerConfig{
				ChunkingStrategy: nil,
			},
			expectedName:   "element-level",
			expectedParams: true,
		},
		{
			name: "empty strategy name",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name: "",
				},
			},
			expectedName:   "element-level",
			expectedParams: true,
		},
		{
			name: "valid strategy with nil parameters",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name:       "hierarchical",
					Parameters: nil,
				},
			},
			expectedName:   "hierarchical",
			expectedParams: true,
		},
		{
			name: "hierarchical strategy without depth config",
			config: &ChunkerConfig{
				ChunkingStrategy: &StrategyConfig{
					Name:       "hierarchical",
					Parameters: make(map[string]interface{}),
				},
			},
			expectedName:   "hierarchical",
			expectedParams: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnsureDefaultStrategyConfig(tt.config)

			if tt.config == nil {
				return // nil config case
			}

			if tt.config.ChunkingStrategy == nil {
				t.Error("Expected strategy config to be set")
				return
			}

			if tt.config.ChunkingStrategy.Name != tt.expectedName {
				t.Errorf("Expected strategy name %s, got %s", tt.expectedName, tt.config.ChunkingStrategy.Name)
			}

			if tt.expectedParams && tt.config.ChunkingStrategy.Parameters == nil {
				t.Error("Expected parameters map to be initialized")
			}

			// 检查层级策略的特殊处理
			if tt.config.ChunkingStrategy.Name == "hierarchical" {
				if _, hasMaxDepth := tt.config.ChunkingStrategy.Parameters["max_depth"]; !hasMaxDepth {
					t.Error("Expected hierarchical strategy to have max_depth parameter")
				}
				if _, hasMergeEmpty := tt.config.ChunkingStrategy.Parameters["merge_empty"]; !hasMergeEmpty {
					t.Error("Expected hierarchical strategy to have merge_empty parameter")
				}
			}
		})
	}
}

func TestDefaultConfigHasValidStrategy(t *testing.T) {
	config := DefaultConfig()

	if config.ChunkingStrategy == nil {
		t.Fatal("Default config should have a strategy configuration")
	}

	if config.ChunkingStrategy.Name == "" {
		t.Error("Default strategy should have a name")
	}

	if config.ChunkingStrategy.Name != "element-level" {
		t.Errorf("Expected default strategy to be 'element-level', got '%s'", config.ChunkingStrategy.Name)
	}

	// 验证默认配置是有效的
	if err := ValidateConfig(config); err != nil {
		t.Errorf("Default config should be valid, got error: %v", err)
	}
}

func TestNewMarkdownChunkerWithStrategy(t *testing.T) {
	tests := []struct {
		name             string
		strategyName     string
		expectedStrategy string
	}{
		{
			name:             "element-level strategy",
			strategyName:     "element-level",
			expectedStrategy: "element-level",
		},
		{
			name:             "hierarchical strategy",
			strategyName:     "hierarchical",
			expectedStrategy: "hierarchical",
		},
		{
			name:             "document-level strategy",
			strategyName:     "document-level",
			expectedStrategy: "document-level",
		},
		{
			name:             "unknown strategy defaults to element-level",
			strategyName:     "unknown-strategy",
			expectedStrategy: "element-level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunkerWithStrategy(tt.strategyName)

			if chunker == nil {
				t.Fatal("NewMarkdownChunkerWithStrategy() returned nil")
			}

			if chunker.strategy == nil {
				t.Fatal("Chunker strategy should not be nil")
			}

			if chunker.strategy.GetName() != tt.expectedStrategy {
				t.Errorf("Expected strategy %s, got %s", tt.expectedStrategy, chunker.strategy.GetName())
			}

			if chunker.strategyRegistry == nil {
				t.Error("Strategy registry should not be nil")
			}

			// 验证策略注册器包含预期的策略
			availableStrategies := chunker.strategyRegistry.List()
			if len(availableStrategies) == 0 {
				t.Error("Strategy registry should contain registered strategies")
			}

			// 验证至少包含基本策略
			expectedStrategies := []string{"element-level", "hierarchical", "document-level"}
			for _, expected := range expectedStrategies {
				found := false
				for _, available := range availableStrategies {
					if available == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected strategy %s to be registered, available: %v", expected, availableStrategies)
				}
			}
		})
	}
}

func TestNewMarkdownChunkerWithHierarchicalStrategy(t *testing.T) {
	tests := []struct {
		name     string
		maxDepth int
	}{
		{
			name:     "no depth limit",
			maxDepth: 0,
		},
		{
			name:     "depth limit 3",
			maxDepth: 3,
		},
		{
			name:     "depth limit 6",
			maxDepth: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunkerWithHierarchicalStrategy(tt.maxDepth)

			if chunker == nil {
				t.Fatal("NewMarkdownChunkerWithHierarchicalStrategy() returned nil")
			}

			if chunker.strategy == nil {
				t.Fatal("Chunker strategy should not be nil")
			}

			if chunker.strategy.GetName() != "hierarchical" {
				t.Errorf("Expected hierarchical strategy, got %s", chunker.strategy.GetName())
			}

			// 验证配置中的最大深度设置
			if chunker.config.ChunkingStrategy == nil {
				t.Fatal("Chunker config should have strategy configuration")
			}

			if chunker.config.ChunkingStrategy.MaxDepth != tt.maxDepth {
				t.Errorf("Expected max depth %d, got %d", tt.maxDepth, chunker.config.ChunkingStrategy.MaxDepth)
			}
		})
	}
}

func TestConstructorBackwardCompatibility(t *testing.T) {
	// 测试现有的构造函数仍然工作
	t.Run("NewMarkdownChunker", func(t *testing.T) {
		chunker := NewMarkdownChunker()
		if chunker == nil {
			t.Fatal("NewMarkdownChunker() should not return nil")
		}

		// 应该使用默认的元素级策略
		if chunker.strategy.GetName() != "element-level" {
			t.Errorf("Expected default strategy to be element-level, got %s", chunker.strategy.GetName())
		}
	})

	t.Run("NewMarkdownChunkerWithConfig with nil", func(t *testing.T) {
		chunker := NewMarkdownChunkerWithConfig(nil)
		if chunker == nil {
			t.Fatal("NewMarkdownChunkerWithConfig(nil) should not return nil")
		}

		// 应该使用默认配置和策略
		if chunker.strategy.GetName() != "element-level" {
			t.Errorf("Expected default strategy to be element-level, got %s", chunker.strategy.GetName())
		}
	})

	t.Run("NewMarkdownChunkerWithConfig with empty config", func(t *testing.T) {
		config := &ChunkerConfig{}
		chunker := NewMarkdownChunkerWithConfig(config)
		if chunker == nil {
			t.Fatal("NewMarkdownChunkerWithConfig() should not return nil")
		}

		// 应该使用默认策略
		if chunker.strategy.GetName() != "element-level" {
			t.Errorf("Expected default strategy to be element-level, got %s", chunker.strategy.GetName())
		}
	})
}

func TestStrategySystemInitialization(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 验证策略注册器已正确初始化
	if chunker.strategyRegistry == nil {
		t.Fatal("Strategy registry should be initialized")
	}

	// 验证所有默认策略都已注册
	expectedStrategies := []string{"element-level", "hierarchical", "document-level"}
	availableStrategies := chunker.strategyRegistry.List()

	for _, expected := range expectedStrategies {
		found := false
		for _, available := range availableStrategies {
			if available == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected strategy %s to be registered, available: %v", expected, availableStrategies)
		}
	}

	// 验证当前策略已正确设置
	if chunker.strategy == nil {
		t.Fatal("Current strategy should be set")
	}

	// 验证策略注册器的功能
	strategy, err := chunker.strategyRegistry.Get("element-level")
	if err != nil {
		t.Errorf("Should be able to get element-level strategy: %v", err)
	}
	if strategy == nil {
		t.Error("Retrieved strategy should not be nil")
	}
}
