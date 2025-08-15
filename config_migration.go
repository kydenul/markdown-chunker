package markdownchunker

import (
	"encoding/json"
	"reflect"

	"github.com/kydenul/log"
)

// ConfigVersion 表示配置版本
type ConfigVersion string

const (
	// ConfigVersionV1 版本1配置（策略系统之前）
	ConfigVersionV1 ConfigVersion = "v1"
	// ConfigVersionV2 版本2配置（策略系统）
	ConfigVersionV2 ConfigVersion = "v2"
)

// LegacyChunkerConfig 旧版本的分块器配置（策略系统之前）
type LegacyChunkerConfig struct {
	// 基本配置
	MaxChunkSize        int                 `json:"max_chunk_size"`
	EnabledTypes        map[string]bool     `json:"enabled_types"`
	CustomExtractors    []MetadataExtractor `json:"custom_extractors"`
	ErrorHandling       ErrorHandlingMode   `json:"error_handling"`
	PerformanceMode     PerformanceMode     `json:"performance_mode"`
	FilterEmptyChunks   bool                `json:"filter_empty_chunks"`
	PreserveWhitespace  bool                `json:"preserve_whitespace"`
	MemoryLimit         int64               `json:"memory_limit"`
	EnableObjectPooling bool                `json:"enable_object_pooling"`

	// 日志配置
	LogLevel     string `json:"log_level"`
	EnableLog    bool   `json:"enable_log"`
	LogFormat    string `json:"log_format"`
	LogDirectory string `json:"log_directory"`

	// 版本标识
	Version ConfigVersion `json:"version,omitempty"`
}

// ConfigMigrationResult 配置迁移结果
type ConfigMigrationResult struct {
	// 迁移后的配置
	Config *ChunkerConfig `json:"config"`
	// 是否进行了迁移
	Migrated bool `json:"migrated"`
	// 原始版本
	OriginalVersion ConfigVersion `json:"original_version"`
	// 目标版本
	TargetVersion ConfigVersion `json:"target_version"`
	// 迁移警告
	Warnings []string `json:"warnings"`
	// 迁移说明
	Notes []string `json:"notes"`
}

// MigrateConfig 迁移配置到最新版本
// 这个函数处理从旧版本配置到新版本配置的迁移
func MigrateConfig(config any) (*ConfigMigrationResult, error) {
	if config == nil {
		// 如果配置为空，创建默认配置
		return &ConfigMigrationResult{
			Config:          DefaultConfig(),
			Migrated:        true,
			OriginalVersion: "",
			TargetVersion:   ConfigVersionV2,
			Warnings:        []string{"配置为空，使用默认配置"},
			Notes:           []string{"创建了默认的元素级策略配置"},
		}, nil
	}

	// 检查配置类型
	switch cfg := config.(type) {
	case *ChunkerConfig:
		// 已经是新版本配置
		return migrateV2Config(cfg)
	case *LegacyChunkerConfig:
		// 旧版本配置，需要迁移
		return migrateLegacyConfig(cfg)
	case map[string]any:
		// JSON 格式的配置，需要解析后迁移
		return migrateJSONConfig(cfg)
	case []byte:
		// JSON 字节数组，需要解析后迁移
		return migrateJSONBytes(cfg)
	case string:
		// JSON 字符串，需要解析后迁移
		return migrateJSONString(cfg)
	default:
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "不支持的配置类型", nil).
			WithContext("function", "MigrateConfig").
			WithContext("config_type", reflect.TypeOf(config).String())
	}
}

// migrateV2Config 处理已经是V2版本的配置
func migrateV2Config(config *ChunkerConfig) (*ConfigMigrationResult, error) {
	if config == nil {
		return &ConfigMigrationResult{
			Config:          DefaultConfig(),
			Migrated:        true,
			OriginalVersion: ConfigVersionV2,
			TargetVersion:   ConfigVersionV2,
			Warnings:        []string{"配置为空，使用默认配置"},
			Notes:           []string{"创建了默认的元素级策略配置"},
		}, nil
	}

	// 创建配置的深拷贝，避免修改原配置
	newConfig := &ChunkerConfig{
		MaxChunkSize:        config.MaxChunkSize,
		EnabledTypes:        make(map[string]bool),
		CustomExtractors:    make([]MetadataExtractor, len(config.CustomExtractors)),
		ErrorHandling:       config.ErrorHandling,
		PerformanceMode:     config.PerformanceMode,
		FilterEmptyChunks:   config.FilterEmptyChunks,
		PreserveWhitespace:  config.PreserveWhitespace,
		MemoryLimit:         config.MemoryLimit,
		EnableObjectPooling: config.EnableObjectPooling,
		LogLevel:            config.LogLevel,
		EnableLog:           config.EnableLog,
		LogFormat:           config.LogFormat,
		LogDirectory:        config.LogDirectory,
	}

	// 深拷贝 EnabledTypes
	if config.EnabledTypes != nil {
		for k, v := range config.EnabledTypes {
			newConfig.EnabledTypes[k] = v
		}
	}

	// 深拷贝 CustomExtractors
	copy(newConfig.CustomExtractors, config.CustomExtractors)

	// 深拷贝策略配置
	if config.ChunkingStrategy != nil {
		newConfig.ChunkingStrategy = &StrategyConfig{
			Name:         config.ChunkingStrategy.Name,
			Parameters:   make(map[string]interface{}),
			MaxDepth:     config.ChunkingStrategy.MaxDepth,
			MinDepth:     config.ChunkingStrategy.MinDepth,
			MergeEmpty:   config.ChunkingStrategy.MergeEmpty,
			MinChunkSize: config.ChunkingStrategy.MinChunkSize,
			MaxChunkSize: config.ChunkingStrategy.MaxChunkSize,
			IncludeTypes: make([]string, len(config.ChunkingStrategy.IncludeTypes)),
			ExcludeTypes: make([]string, len(config.ChunkingStrategy.ExcludeTypes)),
		}

		// 深拷贝 Parameters
		for k, v := range config.ChunkingStrategy.Parameters {
			newConfig.ChunkingStrategy.Parameters[k] = v
		}

		// 深拷贝切片
		copy(newConfig.ChunkingStrategy.IncludeTypes, config.ChunkingStrategy.IncludeTypes)
		copy(newConfig.ChunkingStrategy.ExcludeTypes, config.ChunkingStrategy.ExcludeTypes)
	}

	// 检查是否需要修复或补充配置
	var warnings []string
	var notes []string
	migrated := false

	// 确保有策略配置
	if newConfig.ChunkingStrategy == nil {
		newConfig.ChunkingStrategy = ElementLevelConfig()
		migrated = true
		warnings = append(warnings, "缺少策略配置，添加了默认的元素级策略")
		notes = append(notes, "为了保持向后兼容性，使用元素级策略作为默认策略")
	}

	// 验证并修复策略配置
	EnsureDefaultStrategyConfig(newConfig)

	// 验证配置
	if err := ValidateConfig(newConfig); err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "配置验证失败", err).
			WithContext("function", "migrateV2Config")
	}

	return &ConfigMigrationResult{
		Config:          newConfig,
		Migrated:        migrated,
		OriginalVersion: ConfigVersionV2,
		TargetVersion:   ConfigVersionV2,
		Warnings:        warnings,
		Notes:           notes,
	}, nil
}

// migrateLegacyConfig 迁移旧版本配置到新版本
func migrateLegacyConfig(legacyConfig *LegacyChunkerConfig) (*ConfigMigrationResult, error) {
	if legacyConfig == nil {
		return &ConfigMigrationResult{
			Config:          DefaultConfig(),
			Migrated:        true,
			OriginalVersion: ConfigVersionV1,
			TargetVersion:   ConfigVersionV2,
			Warnings:        []string{"旧配置为空，使用默认配置"},
			Notes:           []string{"创建了默认的元素级策略配置"},
		}, nil
	}

	// 创建新版本配置
	newConfig := &ChunkerConfig{
		// 复制基本配置
		MaxChunkSize:        legacyConfig.MaxChunkSize,
		EnabledTypes:        legacyConfig.EnabledTypes,
		CustomExtractors:    legacyConfig.CustomExtractors,
		ErrorHandling:       legacyConfig.ErrorHandling,
		PerformanceMode:     legacyConfig.PerformanceMode,
		FilterEmptyChunks:   legacyConfig.FilterEmptyChunks,
		PreserveWhitespace:  legacyConfig.PreserveWhitespace,
		MemoryLimit:         legacyConfig.MemoryLimit,
		EnableObjectPooling: legacyConfig.EnableObjectPooling,

		// 复制日志配置
		LogLevel:     legacyConfig.LogLevel,
		EnableLog:    legacyConfig.EnableLog,
		LogFormat:    legacyConfig.LogFormat,
		LogDirectory: legacyConfig.LogDirectory,

		// 添加默认策略配置
		ChunkingStrategy: ElementLevelConfig(),
	}

	var warnings []string
	var notes []string

	// 检查是否有需要特殊处理的配置
	if len(legacyConfig.EnabledTypes) > 0 {
		// 如果旧配置中有类型过滤，将其转换为策略配置
		var includeTypes []string
		for typeName, enabled := range legacyConfig.EnabledTypes {
			if enabled {
				includeTypes = append(includeTypes, typeName)
			}
		}

		if len(includeTypes) > 0 && len(includeTypes) < 7 { // 不是所有类型都启用
			// 创建带类型过滤的元素级策略
			newConfig.ChunkingStrategy = ElementLevelConfigWithTypes(includeTypes, nil)
			notes = append(notes, "将旧版本的类型过滤配置转换为策略配置")
		}
	}

	// 检查大小限制
	if legacyConfig.MaxChunkSize > 0 {
		// 如果有大小限制，更新策略配置
		if newConfig.ChunkingStrategy.MaxChunkSize == 0 {
			newConfig.ChunkingStrategy.MaxChunkSize = legacyConfig.MaxChunkSize
			newConfig.ChunkingStrategy.Parameters["max_chunk_size"] = legacyConfig.MaxChunkSize
			notes = append(notes, "将旧版本的大小限制配置转换为策略配置")
		}
	}

	// 确保配置完整性
	EnsureDefaultStrategyConfig(newConfig)

	// 验证迁移后的配置
	if err := ValidateConfig(newConfig); err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "迁移后的配置验证失败", err).
			WithContext("function", "migrateLegacyConfig")
	}

	notes = append(notes, "成功从V1配置迁移到V2配置，保持了原有的分块行为")

	return &ConfigMigrationResult{
		Config:          newConfig,
		Migrated:        true,
		OriginalVersion: ConfigVersionV1,
		TargetVersion:   ConfigVersionV2,
		Warnings:        warnings,
		Notes:           notes,
	}, nil
}

// migrateJSONConfig 迁移JSON格式的配置
func migrateJSONConfig(jsonConfig map[string]any) (*ConfigMigrationResult, error) {
	// 检查是否包含策略配置字段
	if _, hasStrategy := jsonConfig["chunking_strategy"]; hasStrategy {
		// 看起来是V2配置，尝试解析为新版本配置
		return parseV2JSONConfig(jsonConfig)
	} else {
		// 看起来是V1配置，尝试解析为旧版本配置
		return parseV1JSONConfig(jsonConfig)
	}
}

// migrateJSONBytes 迁移JSON字节数组配置
func migrateJSONBytes(jsonBytes []byte) (*ConfigMigrationResult, error) {
	var jsonConfig map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonConfig); err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "JSON配置解析失败", err).
			WithContext("function", "migrateJSONBytes")
	}

	return migrateJSONConfig(jsonConfig)
}

// migrateJSONString 迁移JSON字符串配置
func migrateJSONString(jsonString string) (*ConfigMigrationResult, error) {
	return migrateJSONBytes([]byte(jsonString))
}

// parseV2JSONConfig 解析V2版本的JSON配置
func parseV2JSONConfig(jsonConfig map[string]any) (*ConfigMigrationResult, error) {
	// 将JSON配置转换为结构体
	jsonBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "JSON配置序列化失败", err).
			WithContext("function", "parseV2JSONConfig")
	}

	var config ChunkerConfig
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "V2配置解析失败", err).
			WithContext("function", "parseV2JSONConfig")
	}

	return migrateV2Config(&config)
}

// parseV1JSONConfig 解析V1版本的JSON配置
func parseV1JSONConfig(jsonConfig map[string]any) (*ConfigMigrationResult, error) {
	// 将JSON配置转换为结构体
	jsonBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "JSON配置序列化失败", err).
			WithContext("function", "parseV1JSONConfig")
	}

	var legacyConfig LegacyChunkerConfig
	if err := json.Unmarshal(jsonBytes, &legacyConfig); err != nil {
		return nil, NewChunkerError(ErrorTypeConfigInvalid, "V1配置解析失败", err).
			WithContext("function", "parseV1JSONConfig")
	}

	return migrateLegacyConfig(&legacyConfig)
}

// IsLegacyConfig 检查配置是否为旧版本配置
func IsLegacyConfig(config any) bool {
	switch cfg := config.(type) {
	case *LegacyChunkerConfig:
		return true
	case *ChunkerConfig:
		return false
	case map[string]any:
		// 检查是否包含策略配置字段
		_, hasStrategy := cfg["chunking_strategy"]
		return !hasStrategy
	default:
		return false
	}
}

// GetConfigVersion 获取配置版本
func GetConfigVersion(config any) ConfigVersion {
	switch cfg := config.(type) {
	case *LegacyChunkerConfig:
		if cfg.Version != "" {
			return cfg.Version
		}
		return ConfigVersionV1
	case *ChunkerConfig:
		return ConfigVersionV2
	case map[string]any:
		if version, ok := cfg["version"].(string); ok {
			return ConfigVersion(version)
		}
		// 检查是否包含策略配置字段
		if _, hasStrategy := cfg["chunking_strategy"]; hasStrategy {
			return ConfigVersionV2
		}
		return ConfigVersionV1
	default:
		return ""
	}
}

// MigrateConfigWithLogger 带日志记录的配置迁移
func MigrateConfigWithLogger(config any, logger log.Logger) (*ConfigMigrationResult, error) {
	if logger == nil {
		// 如果没有提供日志器，使用默认的
		logger = log.NewLogger(&log.Options{
			Level:      "info",
			Format:     "console",
			Directory:  "./logs",
			TimeLayout: "2006-01-02 15:04:05.000",
		})
	}

	logger.Infow("开始配置迁移",
		"function", "MigrateConfigWithLogger",
		"config_type", reflect.TypeOf(config).String())

	result, err := MigrateConfig(config)
	if err != nil {
		logger.Errorw("配置迁移失败",
			"function", "MigrateConfigWithLogger",
			"error", err.Error())
		return nil, err
	}

	// 记录迁移结果
	logger.Infow("配置迁移完成",
		"function", "MigrateConfigWithLogger",
		"migrated", result.Migrated,
		"original_version", result.OriginalVersion,
		"target_version", result.TargetVersion,
		"warnings_count", len(result.Warnings),
		"notes_count", len(result.Notes))

	// 记录警告
	for i, warning := range result.Warnings {
		logger.Warnw("配置迁移警告",
			"function", "MigrateConfigWithLogger",
			"warning_index", i+1,
			"warning", warning)
	}

	// 记录说明
	for i, note := range result.Notes {
		logger.Infow("配置迁移说明",
			"function", "MigrateConfigWithLogger",
			"note_index", i+1,
			"note", note)
	}

	return result, nil
}

// CreateMigrationGuide 创建迁移指南
func CreateMigrationGuide() string {
	return `# 配置迁移指南

## 概述

本指南帮助您将旧版本的 Markdown 分块器配置迁移到新的策略系统。

## 自动迁移

系统提供自动迁移功能，可以处理以下情况：

1. **空配置**: 自动创建默认的元素级策略配置
2. **V1配置**: 将旧版本配置转换为新版本，保持原有行为
3. **V2配置**: 验证并补充缺失的策略配置

## 使用方法

### 方法1: 使用 MigrateConfig 函数

` + "```go" + `
// 迁移任何类型的配置
result, err := markdownchunker.MigrateConfig(oldConfig)
if err != nil {
    log.Fatal(err)
}

// 使用迁移后的配置
chunker := markdownchunker.NewMarkdownChunkerWithConfig(result.Config)
` + "```" + `

### 方法2: 使用带日志的迁移

` + "```go" + `
logger := log.NewLogger(&log.Options{Level: "info"})
result, err := markdownchunker.MigrateConfigWithLogger(oldConfig, logger)
if err != nil {
    log.Fatal(err)
}
` + "```" + `

## 配置变化

### V1 到 V2 的主要变化

1. **新增策略配置**: 添加了 ` + "`ChunkingStrategy`" + ` 字段
2. **类型过滤**: 旧的 ` + "`EnabledTypes`" + ` 转换为策略的类型过滤
3. **大小限制**: 旧的 ` + "`MaxChunkSize`" + ` 转换为策略的大小限制

### 兼容性保证

- 所有旧的API调用继续工作
- 默认行为保持不变
- 现有配置自动迁移

## 迁移示例

### 示例1: 基本配置迁移

旧配置:
` + "```json" + `
{
  "max_chunk_size": 1000,
  "enabled_types": {
    "heading": true,
    "paragraph": true
  },
  "log_level": "INFO"
}
` + "```" + `

迁移后:
` + "```json" + `
{
  "max_chunk_size": 1000,
  "enabled_types": {
    "heading": true,
    "paragraph": true
  },
  "log_level": "INFO",
  "chunking_strategy": {
    "name": "element-level",
    "include_types": ["heading", "paragraph"],
    "max_chunk_size": 1000,
    "parameters": {
      "include_types": ["heading", "paragraph"],
      "max_chunk_size": 1000
    }
  }
}
` + "```" + `

### 示例2: 代码迁移

旧代码:
` + "```go" + `
config := &markdownchunker.ChunkerConfig{
    MaxChunkSize: 1000,
    EnabledTypes: map[string]bool{
        "heading": true,
        "paragraph": true,
    },
}
chunker := markdownchunker.NewMarkdownChunkerWithConfig(config)
` + "```" + `

新代码:
` + "```go" + `
// 方法1: 手动创建新配置
config := markdownchunker.DefaultConfig()
config.ChunkingStrategy = markdownchunker.ElementLevelConfigWithTypes(
    []string{"heading", "paragraph"}, nil)
config.ChunkingStrategy.MaxChunkSize = 1000

// 方法2: 使用迁移功能
result, _ := markdownchunker.MigrateConfig(oldConfig)
config := result.Config

chunker := markdownchunker.NewMarkdownChunkerWithConfig(config)
` + "```" + `

## 注意事项

1. **备份配置**: 迁移前请备份原始配置
2. **测试验证**: 迁移后请测试确保行为符合预期
3. **日志检查**: 查看迁移日志了解具体变化
4. **性能测试**: 新策略系统可能有不同的性能特征

## 故障排除

### 常见问题

1. **配置验证失败**: 检查配置格式和字段值
2. **策略不存在**: 确保使用支持的策略名称
3. **类型转换错误**: 检查JSON配置的数据类型

### 获取帮助

如果遇到迁移问题，请：
1. 检查迁移结果中的警告和说明
2. 查看详细的错误日志
3. 参考API文档和示例代码
`
}
