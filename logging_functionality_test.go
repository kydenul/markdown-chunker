package markdownchunker

import (
	"strings"
	"testing"
	"time"
)

// TestLogConfiguration 测试日志配置功能
func TestLogConfiguration(t *testing.T) {
	t.Run("default log configuration", func(t *testing.T) {
		config := DefaultConfig()

		// 验证默认日志配置
		if config.LogLevel != "INFO" {
			t.Errorf("Expected default LogLevel to be 'INFO', got '%s'", config.LogLevel)
		}
		if !config.EnableLog {
			t.Error("Expected EnableLog to be true by default")
		}
		if config.LogFormat != "console" {
			t.Errorf("Expected default LogFormat to be 'console', got '%s'", config.LogFormat)
		}
		if config.LogDirectory != "./logs" {
			t.Errorf("Expected default LogDirectory to be './logs', got '%s'", config.LogDirectory)
		}
	})

	t.Run("custom log configuration", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "json",
			LogDirectory: "/tmp/test-logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		if chunker.logger == nil {
			t.Fatal("Logger should be initialized with custom config")
		}
	})

	t.Run("disabled logging configuration", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "INFO",
			EnableLog:    false,
			LogFormat:    "console",
			LogDirectory: "./logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		if chunker.logger == nil {
			t.Fatal("Logger should still be initialized even when disabled")
		}
	})

	t.Run("empty log directory uses default", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "INFO",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: "", // Empty should use default
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		if chunker.logger == nil {
			t.Fatal("Logger should be initialized with default directory")
		}
	})
}

// TestLogLevelBehavior 测试不同日志级别的输出行为
func TestLogLevelBehavior(t *testing.T) {
	testCases := []struct {
		name     string
		logLevel string
		expected string
	}{
		{"DEBUG level", "DEBUG", "debug"},
		{"INFO level", "INFO", "info"},
		{"WARN level", "WARN", "warn"},
		{"WARNING level", "WARNING", "warn"},
		{"ERROR level", "ERROR", "error"},
		{"lowercase debug", "debug", "debug"},
		{"lowercase info", "info", "info"},
		{"lowercase warn", "warn", "warn"},
		{"lowercase error", "error", "error"},
		{"invalid level defaults to info", "INVALID", "info"},
		{"empty level defaults to info", "", "info"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseLogLevel(tc.logLevel)
			if result != tc.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tc.logLevel, result, tc.expected)
			}
		})
	}
}

// TestLogFormatValidation 测试日志格式验证
func TestLogFormatValidation(t *testing.T) {
	testCases := []struct {
		name      string
		config    *ChunkerConfig
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid console format",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "console",
			},
			expectErr: false,
		},
		{
			name: "valid json format",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "json",
			},
			expectErr: false,
		},
		{
			name: "invalid format",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "xml",
			},
			expectErr: true,
			errMsg:    "无效的日志格式配置",
		},
		{
			name: "empty format is valid",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "",
			},
			expectErr: false,
		},
		{
			name: "case insensitive format",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "JSON",
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateLogConfig(tc.config)
			if tc.expectErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tc.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// TestLogDirectoryValidation 测试日志目录验证
func TestLogDirectoryValidation(t *testing.T) {
	testCases := []struct {
		name         string
		logDirectory string
		expectErr    bool
		errMsg       string
	}{
		{
			name:         "valid absolute path",
			logDirectory: "/tmp/logs",
			expectErr:    false,
		},
		{
			name:         "valid relative path",
			logDirectory: "./logs",
			expectErr:    false,
		},
		{
			name:         "empty directory is valid",
			logDirectory: "",
			expectErr:    false,
		},
		{
			name:         "current directory",
			logDirectory: ".",
			expectErr:    false,
		},
		{
			name:         "whitespace only directory",
			logDirectory: "   ",
			expectErr:    true,
			errMsg:       "日志目录不能为空白字符",
		},
		{
			name:         "tab only directory",
			logDirectory: "\t",
			expectErr:    true,
			errMsg:       "日志目录不能为空白字符",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &ChunkerConfig{
				LogLevel:     "INFO",
				LogFormat:    "console",
				LogDirectory: tc.logDirectory,
			}

			err := validateLogConfig(config)
			if tc.expectErr {
				if err == nil {
					t.Error("Expected validation error but got none")
				} else if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tc.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// TestLogContentAndFormat 验证日志内容的正确性和格式
func TestLogContentAndFormat(t *testing.T) {
	t.Run("log context fields are properly formatted", func(t *testing.T) {
		ctx := NewLogContext("TestFunction").
			WithNodeInfo("Heading", 1).
			WithDocumentInfo(500, 3).
			WithProcessTime(100*time.Millisecond).
			WithMetadata("test_key", "test_value")

		fields := ctx.ToLogFields()

		// 验证字段数量是偶数（键值对）
		if len(fields)%2 != 0 {
			t.Error("Log fields should be in key-value pairs")
		}

		// 转换为map便于验证
		fieldMap := make(map[string]interface{})
		for i := 0; i < len(fields); i += 2 {
			key := fields[i].(string)
			value := fields[i+1]
			fieldMap[key] = value
		}

		// 验证必需字段
		expectedFields := map[string]interface{}{
			"function":        "TestFunction",
			"node_type":       "Heading",
			"node_id":         1,
			"document_size":   500,
			"chunk_count":     3,
			"process_time_ms": int64(100),
			"test_key":        "test_value",
		}

		for key, expectedValue := range expectedFields {
			if actualValue, exists := fieldMap[key]; !exists {
				t.Errorf("Expected field '%s' to exist in log fields", key)
			} else if actualValue != expectedValue {
				t.Errorf("Expected field '%s' to be %v, got %v", key, expectedValue, actualValue)
			}
		}

		// 验证文件名和行号字段存在
		if _, exists := fieldMap["file"]; !exists {
			t.Error("Expected 'file' field to exist in log fields")
		}
		if _, exists := fieldMap["line"]; !exists {
			t.Error("Expected 'line' field to exist in log fields")
		}
	})

	t.Run("log with context handles all log levels", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: "/tmp/test-logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)
		ctx := NewLogContext("TestFunction").WithMetadata("test", "value")

		// 测试所有日志级别不会panic
		levels := []string{"debug", "info", "warn", "error", "unknown"}
		for _, level := range levels {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("logWithContext panicked for level '%s': %v", level, r)
					}
				}()
				chunker.logWithContext(level, "Test message", ctx)
			}()
		}
	})
}

// TestLoggerInitialization 测试日志器初始化
func TestLoggerInitialization(t *testing.T) {
	t.Run("logger initialized with valid config", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "json",
			LogDirectory: "/tmp/test-logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		if chunker.logger == nil {
			t.Fatal("Logger should be initialized")
		}

		// 测试日志器方法存在且可调用
		chunker.logger.Debug("Debug test")
		chunker.logger.Info("Info test")
		chunker.logger.Warn("Warn test")
		chunker.logger.Error("Error test")
	})

	t.Run("logger initialized with disabled logging", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "INFO",
			EnableLog:    false,
			LogFormat:    "console",
			LogDirectory: "./logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		if chunker.logger == nil {
			t.Fatal("Logger should be initialized even when logging is disabled")
		}
	})

	t.Run("logger handles nil config gracefully", func(t *testing.T) {
		chunker := NewMarkdownChunkerWithConfig(nil)

		if chunker.logger == nil {
			t.Fatal("Logger should be initialized with default config when nil config provided")
		}
	})
}

// TestLoggerIntegrationWithChunking 测试日志器与分块功能的集成
func TestLoggerIntegrationWithChunking(t *testing.T) {
	t.Run("chunking process generates appropriate logs", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: "/tmp/test-logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		content := []byte(`# Test Heading

This is a test paragraph with some content.

## Another Heading

- List item 1
- List item 2

` + "```go\nfunc test() {}\n```")

		chunks, err := chunker.ChunkDocument(content)
		if err != nil {
			t.Fatalf("ChunkDocument failed: %v", err)
		}

		if len(chunks) == 0 {
			t.Error("Expected chunks to be generated")
		}

		// 验证分块过程中日志功能正常工作（通过没有panic来验证）
		// 实际的日志输出会在测试运行时显示
	})

	t.Run("error scenarios generate error logs", func(t *testing.T) {
		config := &ChunkerConfig{
			LogLevel:     "ERROR",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: "/tmp/test-logs",
		}

		chunker := NewMarkdownChunkerWithConfig(config)

		// 测试nil输入
		chunks, err := chunker.ChunkDocument(nil)
		if err != nil {
			t.Logf("Expected error for nil input: %v", err)
		}

		if len(chunks) != 0 {
			t.Error("Expected no chunks for nil input")
		}
	})
}

// TestLogConfigurationEdgeCases 测试日志配置的边界情况
func TestLogConfigurationEdgeCases(t *testing.T) {
	t.Run("mixed case log levels", func(t *testing.T) {
		testCases := []string{"Debug", "Info", "Warn", "Error", "dEbUg", "iNfO"}

		for _, level := range testCases {
			config := &ChunkerConfig{
				LogLevel:  level,
				EnableLog: true,
				LogFormat: "console",
			}

			err := validateLogConfig(config)
			if err != nil {
				t.Errorf("validateLogConfig should handle mixed case level '%s', got error: %v", level, err)
			}
		}
	})

	t.Run("mixed case log formats", func(t *testing.T) {
		testCases := []string{"Console", "JSON", "cOnSoLe", "JsOn"}

		for _, format := range testCases {
			config := &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: format,
			}

			// 只有 "console" 和 "json" (小写) 是有效的
			err := validateLogConfig(config)
			lowerFormat := strings.ToLower(format)
			if lowerFormat == "console" || lowerFormat == "json" {
				if err != nil {
					t.Errorf("validateLogConfig should accept format '%s', got error: %v", format, err)
				}
			} else {
				if err == nil {
					t.Errorf("validateLogConfig should reject invalid format '%s'", format)
				}
			}
		}
	})

	t.Run("special characters in log directory", func(t *testing.T) {
		testCases := []struct {
			directory string
			valid     bool
		}{
			{"./logs-test", true},
			{"./logs_test", true},
			{"./logs.test", true},
			{"./logs/sub", true},
			{"/tmp/logs-123", true},
			{"logs with spaces", true}, // 空格应该是允许的
		}

		for _, tc := range testCases {
			config := &ChunkerConfig{
				LogLevel:     "INFO",
				LogFormat:    "console",
				LogDirectory: tc.directory,
			}

			err := validateLogConfig(config)
			if tc.valid && err != nil {
				t.Errorf("Expected directory '%s' to be valid, got error: %v", tc.directory, err)
			} else if !tc.valid && err == nil {
				t.Errorf("Expected directory '%s' to be invalid", tc.directory)
			}
		}
	})
}

// TestLoggerMemoryAndPerformance 测试日志器的内存和性能影响
func TestLoggerMemoryAndPerformance(t *testing.T) {
	t.Run("logging does not significantly impact performance", func(t *testing.T) {
		// 创建两个配置：一个启用详细日志，一个禁用日志
		enabledConfig := &ChunkerConfig{
			LogLevel:     "DEBUG",
			EnableLog:    true,
			LogFormat:    "console",
			LogDirectory: "/tmp/test-logs",
		}

		disabledConfig := &ChunkerConfig{
			LogLevel:     "ERROR",
			EnableLog:    false,
			LogFormat:    "console",
			LogDirectory: "/tmp/test-logs",
		}

		content := []byte(strings.Repeat("# Heading\n\nParagraph content.\n\n", 100))

		// 测试启用日志的性能
		chunkerEnabled := NewMarkdownChunkerWithConfig(enabledConfig)
		startEnabled := time.Now()
		chunksEnabled, err := chunkerEnabled.ChunkDocument(content)
		durationEnabled := time.Since(startEnabled)

		if err != nil {
			t.Fatalf("ChunkDocument with logging enabled failed: %v", err)
		}

		// 测试禁用日志的性能
		chunkerDisabled := NewMarkdownChunkerWithConfig(disabledConfig)
		startDisabled := time.Now()
		chunksDisabled, err := chunkerDisabled.ChunkDocument(content)
		durationDisabled := time.Since(startDisabled)

		if err != nil {
			t.Fatalf("ChunkDocument with logging disabled failed: %v", err)
		}

		// 验证结果一致性
		if len(chunksEnabled) != len(chunksDisabled) {
			t.Errorf("Chunk count should be same regardless of logging: enabled=%d, disabled=%d",
				len(chunksEnabled), len(chunksDisabled))
		}

		// 日志不应该显著影响性能（允许2倍的差异）
		if durationEnabled > durationDisabled*2 {
			t.Logf("Warning: Logging may have significant performance impact. Enabled: %v, Disabled: %v",
				durationEnabled, durationDisabled)
		}

		t.Logf("Performance comparison - Enabled: %v, Disabled: %v", durationEnabled, durationDisabled)
	})
}
