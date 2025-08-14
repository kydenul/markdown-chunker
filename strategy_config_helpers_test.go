package markdownchunker

import (
	"reflect"
	"testing"
)

// TestElementLevelConfigWithTypes 测试带内容类型过滤的元素级策略配置
func TestElementLevelConfigWithTypes(t *testing.T) {
	includeTypes := []string{"heading", "paragraph"}
	excludeTypes := []string{"code", "table"}

	config := ElementLevelConfigWithTypes(includeTypes, excludeTypes)

	if config == nil {
		t.Fatal("配置创建失败")
	}

	if config.Name != "element-level" {
		t.Errorf("期望策略名称为 'element-level'，实际为 '%s'", config.Name)
	}

	if !reflect.DeepEqual(config.IncludeTypes, includeTypes) {
		t.Errorf("期望包含类型为 %v，实际为 %v", includeTypes, config.IncludeTypes)
	}

	if !reflect.DeepEqual(config.ExcludeTypes, excludeTypes) {
		t.Errorf("期望排除类型为 %v，实际为 %v", excludeTypes, config.ExcludeTypes)
	}

	// 检查参数映射
	if paramInclude, ok := config.Parameters["include_types"]; !ok {
		t.Error("参数映射中缺少 include_types")
	} else if !reflect.DeepEqual(paramInclude, includeTypes) {
		t.Errorf("参数映射中的 include_types 不匹配")
	}

	if paramExclude, ok := config.Parameters["exclude_types"]; !ok {
		t.Error("参数映射中缺少 exclude_types")
	} else if !reflect.DeepEqual(paramExclude, excludeTypes) {
		t.Errorf("参数映射中的 exclude_types 不匹配")
	}
}

// TestElementLevelConfigWithSize 测试带大小限制的元素级策略配置
func TestElementLevelConfigWithSize(t *testing.T) {
	minSize := 100
	maxSize := 1000

	config := ElementLevelConfigWithSize(minSize, maxSize)

	if config == nil {
		t.Fatal("配置创建失败")
	}

	if config.Name != "element-level" {
		t.Errorf("期望策略名称为 'element-level'，实际为 '%s'", config.Name)
	}

	if config.MinChunkSize != minSize {
		t.Errorf("期望最小块大小为 %d，实际为 %d", minSize, config.MinChunkSize)
	}

	if config.MaxChunkSize != maxSize {
		t.Errorf("期望最大块大小为 %d，实际为 %d", maxSize, config.MaxChunkSize)
	}

	// 检查参数映射
	if paramMin, ok := config.Parameters["min_chunk_size"]; !ok {
		t.Error("参数映射中缺少 min_chunk_size")
	} else if paramMin != minSize {
		t.Errorf("参数映射中的 min_chunk_size 不匹配")
	}

	if paramMax, ok := config.Parameters["max_chunk_size"]; !ok {
		t.Error("参数映射中缺少 max_chunk_size")
	} else if paramMax != maxSize {
		t.Errorf("参数映射中的 max_chunk_size 不匹配")
	}
}

// TestHierarchicalConfigAdvanced 测试高级层级策略配置
func TestHierarchicalConfigAdvanced(t *testing.T) {
	maxDepth := 3
	minDepth := 1
	mergeEmpty := false

	config := HierarchicalConfigAdvanced(maxDepth, minDepth, mergeEmpty)

	if config == nil {
		t.Fatal("配置创建失败")
	}

	if config.Name != "hierarchical" {
		t.Errorf("期望策略名称为 'hierarchical'，实际为 '%s'", config.Name)
	}

	if config.MaxDepth != maxDepth {
		t.Errorf("期望最大深度为 %d，实际为 %d", maxDepth, config.MaxDepth)
	}

	if config.MinDepth != minDepth {
		t.Errorf("期望最小深度为 %d，实际为 %d", minDepth, config.MinDepth)
	}

	if config.MergeEmpty != mergeEmpty {
		t.Errorf("期望合并空章节为 %t，实际为 %t", mergeEmpty, config.MergeEmpty)
	}

	// 检查参数映射
	if paramMaxDepth, ok := config.Parameters["max_depth"]; !ok {
		t.Error("参数映射中缺少 max_depth")
	} else if paramMaxDepth != maxDepth {
		t.Errorf("参数映射中的 max_depth 不匹配")
	}

	if paramMinDepth, ok := config.Parameters["min_depth"]; !ok {
		t.Error("参数映射中缺少 min_depth")
	} else if paramMinDepth != minDepth {
		t.Errorf("参数映射中的 min_depth 不匹配")
	}

	if paramMergeEmpty, ok := config.Parameters["merge_empty"]; !ok {
		t.Error("参数映射中缺少 merge_empty")
	} else if paramMergeEmpty != mergeEmpty {
		t.Errorf("参数映射中的 merge_empty 不匹配")
	}
}

// TestHierarchicalConfigWithSize 测试带大小限制的层级策略配置
func TestHierarchicalConfigWithSize(t *testing.T) {
	maxDepth := 2
	minSize := 50
	maxSize := 500

	config := HierarchicalConfigWithSize(maxDepth, minSize, maxSize)

	if config == nil {
		t.Fatal("配置创建失败")
	}

	if config.Name != "hierarchical" {
		t.Errorf("期望策略名称为 'hierarchical'，实际为 '%s'", config.Name)
	}

	if config.MaxDepth != maxDepth {
		t.Errorf("期望最大深度为 %d，实际为 %d", maxDepth, config.MaxDepth)
	}

	if config.MinChunkSize != minSize {
		t.Errorf("期望最小块大小为 %d，实际为 %d", minSize, config.MinChunkSize)
	}

	if config.MaxChunkSize != maxSize {
		t.Errorf("期望最大块大小为 %d，实际为 %d", maxSize, config.MaxChunkSize)
	}
}

// TestDocumentLevelConfigWithSize 测试带大小限制的文档级策略配置
func TestDocumentLevelConfigWithSize(t *testing.T) {
	minSize := 1000
	maxSize := 10000

	config := DocumentLevelConfigWithSize(minSize, maxSize)

	if config == nil {
		t.Fatal("配置创建失败")
	}

	if config.Name != "document-level" {
		t.Errorf("期望策略名称为 'document-level'，实际为 '%s'", config.Name)
	}

	if config.MinChunkSize != minSize {
		t.Errorf("期望最小块大小为 %d，实际为 %d", minSize, config.MinChunkSize)
	}

	if config.MaxChunkSize != maxSize {
		t.Errorf("期望最大块大小为 %d，实际为 %d", maxSize, config.MaxChunkSize)
	}
}

// TestValidateAndFillDefaults 测试配置验证和默认值填充
func TestValidateAndFillDefaults(t *testing.T) {
	tests := []struct {
		name        string
		config      *StrategyConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
			errorMsg:    "策略配置不能为空",
		},
		{
			name: "empty strategy name",
			config: &StrategyConfig{
				Name: "",
			},
			expectError: true,
			errorMsg:    "策略名称不能为空",
		},
		{
			name: "valid hierarchical config",
			config: &StrategyConfig{
				Name:     "hierarchical",
				MaxDepth: 3,
			},
			expectError: false,
		},
		{
			name: "invalid hierarchical config - negative max depth",
			config: &StrategyConfig{
				Name:     "hierarchical",
				MaxDepth: -1,
			},
			expectError: true,
			errorMsg:    "最大层级深度不能为负数",
		},
		{
			name: "invalid hierarchical config - max depth too large",
			config: &StrategyConfig{
				Name:     "hierarchical",
				MaxDepth: 7,
			},
			expectError: true,
			errorMsg:    "最大层级深度不能超过6",
		},
		{
			name: "invalid hierarchical config - min > max",
			config: &StrategyConfig{
				Name:     "hierarchical",
				MaxDepth: 2,
				MinDepth: 3,
			},
			expectError: true,
			errorMsg:    "最小层级深度不能大于最大层级深度",
		},
		{
			name: "invalid element-level config - has depth",
			config: &StrategyConfig{
				Name:     "element-level",
				MaxDepth: 2,
			},
			expectError: true,
			errorMsg:    "元素级策略不支持层级深度配置",
		},
		{
			name: "invalid document-level config - has depth",
			config: &StrategyConfig{
				Name:     "document-level",
				MaxDepth: 2,
			},
			expectError: true,
			errorMsg:    "文档级策略不支持层级深度配置",
		},
		{
			name: "invalid document-level config - has types",
			config: &StrategyConfig{
				Name:         "document-level",
				IncludeTypes: []string{"heading"},
			},
			expectError: true,
			errorMsg:    "文档级策略不支持内容类型过滤",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAndFillDefaults(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				} else if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("期望错误消息包含 '%s'，实际错误: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("不期望出现错误，但出现了错误: %v", err)
				} else {
					// 验证参数映射是否正确填充
					if tt.config.Parameters == nil {
						t.Error("参数映射应该被初始化")
					}
				}
			}
		})
	}
}

// TestCreateConfigFromParameters 测试从参数映射创建策略配置
func TestCreateConfigFromParameters(t *testing.T) {
	tests := []struct {
		name         string
		strategyName string
		params       map[string]interface{}
		expectError  bool
		validate     func(*testing.T, *StrategyConfig)
	}{
		{
			name:         "empty strategy name",
			strategyName: "",
			params:       map[string]interface{}{},
			expectError:  true,
		},
		{
			name:         "hierarchical with parameters",
			strategyName: "hierarchical",
			params: map[string]interface{}{
				"max_depth":   3,
				"min_depth":   1,
				"merge_empty": false,
			},
			expectError: false,
			validate: func(t *testing.T, config *StrategyConfig) {
				if config.MaxDepth != 3 {
					t.Errorf("期望最大深度为 3，实际为 %d", config.MaxDepth)
				}
				if config.MinDepth != 1 {
					t.Errorf("期望最小深度为 1，实际为 %d", config.MinDepth)
				}
				if config.MergeEmpty != false {
					t.Errorf("期望合并空章节为 false，实际为 %t", config.MergeEmpty)
				}
			},
		},
		{
			name:         "element-level with size parameters",
			strategyName: "element-level",
			params: map[string]interface{}{
				"min_chunk_size": 100,
				"max_chunk_size": 1000,
				"include_types":  []string{"heading", "paragraph"},
			},
			expectError: false,
			validate: func(t *testing.T, config *StrategyConfig) {
				if config.MinChunkSize != 100 {
					t.Errorf("期望最小块大小为 100，实际为 %d", config.MinChunkSize)
				}
				if config.MaxChunkSize != 1000 {
					t.Errorf("期望最大块大小为 1000，实际为 %d", config.MaxChunkSize)
				}
				expectedTypes := []string{"heading", "paragraph"}
				if !reflect.DeepEqual(config.IncludeTypes, expectedTypes) {
					t.Errorf("期望包含类型为 %v，实际为 %v", expectedTypes, config.IncludeTypes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := CreateConfigFromParameters(tt.strategyName, tt.params)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				}
			} else {
				if err != nil {
					t.Errorf("不期望出现错误，但出现了错误: %v", err)
				} else if config == nil {
					t.Error("配置不应该为空")
				} else if tt.validate != nil {
					tt.validate(t, config)
				}
			}
		})
	}
}

// TestMergeConfigs 测试配置合并
func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name        string
		base        *StrategyConfig
		override    *StrategyConfig
		expectError bool
		validate    func(*testing.T, *StrategyConfig)
	}{
		{
			name:        "nil base config",
			base:        nil,
			override:    ElementLevelConfig(),
			expectError: true,
		},
		{
			name:        "nil override config",
			base:        ElementLevelConfig(),
			override:    nil,
			expectError: false,
			validate: func(t *testing.T, config *StrategyConfig) {
				if config.Name != "element-level" {
					t.Errorf("期望策略名称为 'element-level'，实际为 '%s'", config.Name)
				}
			},
		},
		{
			name:        "mismatched strategy names",
			base:        ElementLevelConfig(),
			override:    HierarchicalConfig(3),
			expectError: true,
		},
		{
			name:        "merge hierarchical configs",
			base:        HierarchicalConfig(2),
			override:    HierarchicalConfigAdvanced(4, 1, false),
			expectError: false,
			validate: func(t *testing.T, config *StrategyConfig) {
				if config.MaxDepth != 4 {
					t.Errorf("期望最大深度为 4，实际为 %d", config.MaxDepth)
				}
				if config.MinDepth != 1 {
					t.Errorf("期望最小深度为 1，实际为 %d", config.MinDepth)
				}
				if config.MergeEmpty != false {
					t.Errorf("期望合并空章节为 false，实际为 %t", config.MergeEmpty)
				}
			},
		},
		{
			name: "merge with size limits",
			base: ElementLevelConfig(),
			override: &StrategyConfig{
				Name:         "element-level",
				MinChunkSize: 100,
				MaxChunkSize: 1000,
				Parameters:   map[string]interface{}{},
			},
			expectError: false,
			validate: func(t *testing.T, config *StrategyConfig) {
				if config.MinChunkSize != 100 {
					t.Errorf("期望最小块大小为 100，实际为 %d", config.MinChunkSize)
				}
				if config.MaxChunkSize != 1000 {
					t.Errorf("期望最大块大小为 1000，实际为 %d", config.MaxChunkSize)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := MergeConfigs(tt.base, tt.override)

			if tt.expectError {
				if err == nil {
					t.Errorf("期望出现错误，但没有错误")
				}
			} else {
				if err != nil {
					t.Errorf("不期望出现错误，但出现了错误: %v", err)
				} else if merged == nil {
					t.Error("合并后的配置不应该为空")
				} else if tt.validate != nil {
					tt.validate(t, merged)
				}
			}
		})
	}
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
