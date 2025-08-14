package markdownchunker

import (
	"encoding/json"
	"testing"

	"github.com/kydenul/log"
)

func TestMigrateConfig(t *testing.T) {
	t.Run("空配置迁移", func(t *testing.T) {
		result, err := MigrateConfig(nil)
		if err != nil {
			t.Fatalf("迁移空配置失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config == nil {
			t.Fatal("期望迁移后的配置不为空")
		}

		if result.Config.ChunkingStrategy == nil {
			t.Error("期望迁移后的配置包含策略配置")
		}

		if result.Config.ChunkingStrategy.Name != "element-level" {
			t.Errorf("期望默认策略为element-level，实际为: %s", result.Config.ChunkingStrategy.Name)
		}

		if len(result.Warnings) == 0 {
			t.Error("期望有迁移警告")
		}

		if len(result.Notes) == 0 {
			t.Error("期望有迁移说明")
		}
	})

	t.Run("V2配置无需迁移", func(t *testing.T) {
		originalConfig := DefaultConfig()
		originalConfig.MaxChunkSize = 1000
		originalConfig.ChunkingStrategy = HierarchicalConfig(3)

		result, err := MigrateConfig(originalConfig)
		if err != nil {
			t.Fatalf("迁移V2配置失败: %v", err)
		}

		if result.Migrated {
			t.Error("期望V2配置不需要迁移")
		}

		if result.OriginalVersion != ConfigVersionV2 {
			t.Errorf("期望原始版本为V2，实际为: %s", result.OriginalVersion)
		}

		if result.TargetVersion != ConfigVersionV2 {
			t.Errorf("期望目标版本为V2，实际为: %s", result.TargetVersion)
		}

		if result.Config.MaxChunkSize != 1000 {
			t.Errorf("期望MaxChunkSize为1000，实际为: %d", result.Config.MaxChunkSize)
		}
	})

	t.Run("V2配置缺少策略配置", func(t *testing.T) {
		config := &ChunkerConfig{
			MaxChunkSize: 500,
			LogLevel:     "DEBUG",
			// 缺少 ChunkingStrategy
		}

		result, err := MigrateConfig(config)
		if err != nil {
			t.Fatalf("迁移缺少策略的V2配置失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config.ChunkingStrategy == nil {
			t.Error("期望迁移后添加了策略配置")
		}

		if result.Config.ChunkingStrategy.Name != "element-level" {
			t.Errorf("期望默认策略为element-level，实际为: %s", result.Config.ChunkingStrategy.Name)
		}

		if len(result.Warnings) == 0 {
			t.Error("期望有迁移警告")
		}
	})

	t.Run("旧版本配置迁移", func(t *testing.T) {
		legacyConfig := &LegacyChunkerConfig{
			MaxChunkSize: 800,
			EnabledTypes: map[string]bool{
				"heading":   true,
				"paragraph": true,
				"code":      false,
			},
			LogLevel:          "INFO",
			FilterEmptyChunks: true,
			Version:           ConfigVersionV1,
		}

		result, err := MigrateConfig(legacyConfig)
		if err != nil {
			t.Fatalf("迁移旧版本配置失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.OriginalVersion != ConfigVersionV1 {
			t.Errorf("期望原始版本为V1，实际为: %s", result.OriginalVersion)
		}

		if result.TargetVersion != ConfigVersionV2 {
			t.Errorf("期望目标版本为V2，实际为: %s", result.TargetVersion)
		}

		// 检查基本配置是否正确迁移
		if result.Config.MaxChunkSize != 800 {
			t.Errorf("期望MaxChunkSize为800，实际为: %d", result.Config.MaxChunkSize)
		}

		if result.Config.LogLevel != "INFO" {
			t.Errorf("期望LogLevel为INFO，实际为: %s", result.Config.LogLevel)
		}

		if !result.Config.FilterEmptyChunks {
			t.Error("期望FilterEmptyChunks为true")
		}

		// 检查策略配置
		if result.Config.ChunkingStrategy == nil {
			t.Fatal("期望迁移后包含策略配置")
		}

		if result.Config.ChunkingStrategy.Name != "element-level" {
			t.Errorf("期望策略名称为element-level，实际为: %s", result.Config.ChunkingStrategy.Name)
		}

		// 检查类型过滤是否正确转换
		expectedTypes := []string{"heading", "paragraph"}
		if len(result.Config.ChunkingStrategy.IncludeTypes) != len(expectedTypes) {
			t.Errorf("期望包含类型数量为%d，实际为: %d",
				len(expectedTypes), len(result.Config.ChunkingStrategy.IncludeTypes))
		}

		// 检查大小限制是否正确转换
		if result.Config.ChunkingStrategy.MaxChunkSize != 800 {
			t.Errorf("期望策略MaxChunkSize为800，实际为: %d", result.Config.ChunkingStrategy.MaxChunkSize)
		}

		if len(result.Notes) == 0 {
			t.Error("期望有迁移说明")
		}
	})

	t.Run("JSON配置迁移", func(t *testing.T) {
		// V1 JSON配置
		v1JSON := map[string]any{
			"max_chunk_size":      1200,
			"filter_empty_chunks": true,
			"log_level":           "DEBUG",
			"enabled_types": map[string]any{
				"heading":   true,
				"paragraph": true,
				"table":     true,
			},
		}

		result, err := MigrateConfig(v1JSON)
		if err != nil {
			t.Fatalf("迁移V1 JSON配置失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.OriginalVersion != ConfigVersionV1 {
			t.Errorf("期望原始版本为V1，实际为: %s", result.OriginalVersion)
		}

		if result.Config.MaxChunkSize != 1200 {
			t.Errorf("期望MaxChunkSize为1200，实际为: %d", result.Config.MaxChunkSize)
		}

		if result.Config.ChunkingStrategy == nil {
			t.Fatal("期望迁移后包含策略配置")
		}
	})

	t.Run("JSON字节数组迁移", func(t *testing.T) {
		jsonConfig := `{
			"max_chunk_size": 600,
			"log_level": "WARN",
			"enabled_types": {
				"heading": true,
				"code": true
			}
		}`

		result, err := MigrateConfig([]byte(jsonConfig))
		if err != nil {
			t.Fatalf("迁移JSON字节数组失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config.MaxChunkSize != 600 {
			t.Errorf("期望MaxChunkSize为600，实际为: %d", result.Config.MaxChunkSize)
		}

		if result.Config.LogLevel != "WARN" {
			t.Errorf("期望LogLevel为WARN，实际为: %s", result.Config.LogLevel)
		}
	})

	t.Run("JSON字符串迁移", func(t *testing.T) {
		jsonConfig := `{"max_chunk_size": 400, "log_level": "ERROR"}`

		result, err := MigrateConfig(jsonConfig)
		if err != nil {
			t.Fatalf("迁移JSON字符串失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config.MaxChunkSize != 400 {
			t.Errorf("期望MaxChunkSize为400，实际为: %d", result.Config.MaxChunkSize)
		}
	})

	t.Run("不支持的配置类型", func(t *testing.T) {
		_, err := MigrateConfig(123) // 不支持的类型
		if err == nil {
			t.Error("期望不支持的配置类型返回错误")
		}

		if chunkerErr, ok := err.(*ChunkerError); ok {
			if chunkerErr.Type != ErrorTypeConfigInvalid {
				t.Errorf("期望错误类型为ConfigInvalid，实际为: %v", chunkerErr.Type)
			}
		} else {
			t.Error("期望返回ChunkerError类型的错误")
		}
	})
}

func TestIsLegacyConfig(t *testing.T) {
	t.Run("识别旧版本配置", func(t *testing.T) {
		legacyConfig := &LegacyChunkerConfig{
			MaxChunkSize: 1000,
		}

		if !IsLegacyConfig(legacyConfig) {
			t.Error("期望识别为旧版本配置")
		}
	})

	t.Run("识别新版本配置", func(t *testing.T) {
		newConfig := &ChunkerConfig{
			MaxChunkSize:     1000,
			ChunkingStrategy: ElementLevelConfig(),
		}

		if IsLegacyConfig(newConfig) {
			t.Error("期望识别为新版本配置")
		}
	})

	t.Run("识别JSON配置", func(t *testing.T) {
		// 旧版本JSON（没有策略配置）
		v1JSON := map[string]any{
			"max_chunk_size": 1000,
			"log_level":      "INFO",
		}

		if !IsLegacyConfig(v1JSON) {
			t.Error("期望识别为旧版本JSON配置")
		}

		// 新版本JSON（有策略配置）
		v2JSON := map[string]any{
			"max_chunk_size": 1000,
			"chunking_strategy": map[string]any{
				"name": "element-level",
			},
		}

		if IsLegacyConfig(v2JSON) {
			t.Error("期望识别为新版本JSON配置")
		}
	})
}

func TestGetConfigVersion(t *testing.T) {
	t.Run("获取旧版本配置版本", func(t *testing.T) {
		legacyConfig := &LegacyChunkerConfig{
			Version: ConfigVersionV1,
		}

		version := GetConfigVersion(legacyConfig)
		if version != ConfigVersionV1 {
			t.Errorf("期望版本为V1，实际为: %s", version)
		}
	})

	t.Run("获取新版本配置版本", func(t *testing.T) {
		newConfig := &ChunkerConfig{
			ChunkingStrategy: ElementLevelConfig(),
		}

		version := GetConfigVersion(newConfig)
		if version != ConfigVersionV2 {
			t.Errorf("期望版本为V2，实际为: %s", version)
		}
	})

	t.Run("获取JSON配置版本", func(t *testing.T) {
		// 明确指定版本的JSON
		jsonWithVersion := map[string]any{
			"version":        "v1",
			"max_chunk_size": 1000,
		}

		version := GetConfigVersion(jsonWithVersion)
		if version != ConfigVersionV1 {
			t.Errorf("期望版本为V1，实际为: %s", version)
		}

		// 通过策略配置推断版本的JSON
		jsonWithStrategy := map[string]any{
			"chunking_strategy": map[string]any{
				"name": "hierarchical",
			},
		}

		version = GetConfigVersion(jsonWithStrategy)
		if version != ConfigVersionV2 {
			t.Errorf("期望版本为V2，实际为: %s", version)
		}

		// 没有策略配置的JSON（推断为V1）
		jsonWithoutStrategy := map[string]any{
			"max_chunk_size": 1000,
		}

		version = GetConfigVersion(jsonWithoutStrategy)
		if version != ConfigVersionV1 {
			t.Errorf("期望版本为V1，实际为: %s", version)
		}
	})
}

func TestMigrateConfigWithLogger(t *testing.T) {
	t.Run("带日志的配置迁移", func(t *testing.T) {
		logger := log.NewLogger(&log.Options{
			Level:      "debug",
			Format:     "console",
			Directory:  "./logs",
			TimeLayout: "2006-01-02 15:04:05.000",
		})

		legacyConfig := &LegacyChunkerConfig{
			MaxChunkSize: 1000,
			LogLevel:     "INFO",
		}

		result, err := MigrateConfigWithLogger(legacyConfig, logger)
		if err != nil {
			t.Fatalf("带日志的配置迁移失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config == nil {
			t.Error("期望迁移后的配置不为空")
		}
	})

	t.Run("使用默认日志器", func(t *testing.T) {
		legacyConfig := &LegacyChunkerConfig{
			MaxChunkSize: 500,
		}

		result, err := MigrateConfigWithLogger(legacyConfig, nil)
		if err != nil {
			t.Fatalf("使用默认日志器的配置迁移失败: %v", err)
		}

		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}
	})
}

func TestCreateMigrationGuide(t *testing.T) {
	t.Run("创建迁移指南", func(t *testing.T) {
		guide := CreateMigrationGuide()

		if guide == "" {
			t.Error("期望迁移指南不为空")
		}

		// 检查指南是否包含关键信息
		expectedSections := []string{
			"# 配置迁移指南",
			"## 概述",
			"## 自动迁移",
			"## 使用方法",
			"## 配置变化",
			"## 兼容性保证",
			"## 迁移示例",
			"## 注意事项",
			"## 故障排除",
		}

		for _, section := range expectedSections {
			if !containsText(guide, section) {
				t.Errorf("期望迁移指南包含章节: %s", section)
			}
		}

		// 检查是否包含代码示例
		if !containsText(guide, "```go") {
			t.Error("期望迁移指南包含Go代码示例")
		}

		if !containsText(guide, "```json") {
			t.Error("期望迁移指南包含JSON示例")
		}
	})
}

func TestConfigMigrationIntegration(t *testing.T) {
	t.Run("完整的迁移流程测试", func(t *testing.T) {
		// 1. 创建旧版本配置
		legacyConfig := &LegacyChunkerConfig{
			MaxChunkSize: 1500,
			EnabledTypes: map[string]bool{
				"heading":   true,
				"paragraph": true,
				"list":      true,
			},
			LogLevel:            "DEBUG",
			FilterEmptyChunks:   true,
			PreserveWhitespace:  false,
			EnableObjectPooling: true,
		}

		// 2. 执行迁移
		result, err := MigrateConfig(legacyConfig)
		if err != nil {
			t.Fatalf("配置迁移失败: %v", err)
		}

		// 3. 验证迁移结果
		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config == nil {
			t.Fatal("期望迁移后的配置不为空")
		}

		// 4. 使用迁移后的配置创建分块器
		chunker := NewMarkdownChunkerWithConfig(result.Config)
		if chunker == nil {
			t.Fatal("期望能够使用迁移后的配置创建分块器")
		}

		// 5. 测试分块功能
		testMarkdown := `# 标题1

这是第一段内容。

## 标题2

这是第二段内容。

- 列表项1
- 列表项2`

		chunks, err := chunker.ChunkDocument([]byte(testMarkdown))
		if err != nil {
			t.Fatalf("使用迁移后的配置分块失败: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("期望分块结果不为空")
		}

		// 6. 验证分块行为符合预期
		// 由于配置了类型过滤，应该只包含heading、paragraph和list类型的块
		for _, chunk := range chunks {
			if chunk.Type != "heading" && chunk.Type != "paragraph" && chunk.Type != "list" {
				t.Errorf("意外的块类型: %s", chunk.Type)
			}
		}
	})

	t.Run("JSON配置文件迁移测试", func(t *testing.T) {
		// 模拟从配置文件读取的JSON
		configJSON := `{
			"max_chunk_size": 2000,
			"enabled_types": {
				"heading": true,
				"paragraph": true,
				"code": true,
				"table": false
			},
			"log_level": "INFO",
			"filter_empty_chunks": true,
			"preserve_whitespace": false
		}`

		// 解析JSON
		var jsonConfig map[string]any
		err := json.Unmarshal([]byte(configJSON), &jsonConfig)
		if err != nil {
			t.Fatalf("JSON解析失败: %v", err)
		}

		// 执行迁移
		result, err := MigrateConfig(jsonConfig)
		if err != nil {
			t.Fatalf("JSON配置迁移失败: %v", err)
		}

		// 验证迁移结果
		if !result.Migrated {
			t.Error("期望迁移标志为true")
		}

		if result.Config.MaxChunkSize != 2000 {
			t.Errorf("期望MaxChunkSize为2000，实际为: %d", result.Config.MaxChunkSize)
		}

		if result.Config.ChunkingStrategy == nil {
			t.Fatal("期望迁移后包含策略配置")
		}

		// 验证类型过滤转换
		expectedIncludeTypes := []string{"heading", "paragraph", "code"}
		if len(result.Config.ChunkingStrategy.IncludeTypes) != len(expectedIncludeTypes) {
			t.Errorf("期望包含类型数量为%d，实际为: %d",
				len(expectedIncludeTypes), len(result.Config.ChunkingStrategy.IncludeTypes))
		}
	})
}

// containsText 检查字符串是否包含子字符串
func containsText(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func Test_migrateLegacyConfig(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		legacyConfig *LegacyChunkerConfig
		want         *ConfigMigrationResult
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := migrateLegacyConfig(tt.legacyConfig)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("migrateLegacyConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("migrateLegacyConfig() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("migrateLegacyConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
