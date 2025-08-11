package markdownchunker

import (
	"strings"
	"testing"
)

// testConfigWithLogging 创建一个用于测试的配置，使用临时目录避免污染工作目录
func testConfigWithLogging(level, format, directory string, enabled bool) *ChunkerConfig {
	if directory == "" {
		directory = "/tmp/markdown-chunker-test-logs"
	}
	return &ChunkerConfig{
		LogLevel:     level,
		EnableLog:    enabled,
		LogFormat:    format,
		LogDirectory: directory,
	}
}

func TestLoggingIntegration(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChunkerConfig
		expected string
	}{
		{
			name: "Default logging configuration",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				EnableLog: true,
				LogFormat: "console",
			},
			expected: "INFO",
		},
		{
			name: "Debug level logging",
			config: &ChunkerConfig{
				LogLevel:  "DEBUG",
				EnableLog: true,
				LogFormat: "console",
			},
			expected: "DEBUG",
		},
		{
			name: "JSON format logging",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				EnableLog: true,
				LogFormat: "json",
			},
			expected: "JSON",
		},
		{
			name: "Logging disabled",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				EnableLog: false,
				LogFormat: "console",
			},
			expected: "ERROR", // Should default to ERROR level when disabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunker := NewMarkdownChunkerWithConfig(tt.config)

			// Verify logger is initialized
			if chunker.logger == nil {
				t.Fatal("Logger should be initialized")
			}

			// Verify logger is properly configured
			// Note: kydenul/log doesn't expose level getter, so we test by functionality
			// Test that logger works by attempting to log
			// We can't directly verify the level, but we can verify the logger exists and functions
			if chunker.logger == nil {
				t.Error("Logger should not be nil")
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"DEBUG", "debug"},
		{"debug", "debug"},
		{"INFO", "info"},
		{"info", "info"},
		{"WARN", "warn"},
		{"warn", "warn"},
		{"WARNING", "warn"},
		{"warning", "warn"},
		{"ERROR", "error"},
		{"error", "error"},
		{"INVALID", "info"}, // Should default to info
		{"", "info"},        // Should default to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateLogConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *ChunkerConfig
		expectErr bool
	}{
		{
			name:      "nil config",
			config:    nil,
			expectErr: true,
		},
		{
			name: "valid config",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "console",
			},
			expectErr: false,
		},
		{
			name: "invalid log level",
			config: &ChunkerConfig{
				LogLevel:  "INVALID",
				LogFormat: "console",
			},
			expectErr: true,
		},
		{
			name: "invalid log format",
			config: &ChunkerConfig{
				LogLevel:  "INFO",
				LogFormat: "invalid",
			},
			expectErr: true,
		},
		{
			name: "empty log level and format",
			config: &ChunkerConfig{
				LogLevel:  "",
				LogFormat: "",
			},
			expectErr: false, // Empty values should be allowed
		},
		{
			name: "valid log directory",
			config: &ChunkerConfig{
				LogLevel:     "INFO",
				LogFormat:    "console",
				LogDirectory: "/var/log/app",
			},
			expectErr: false,
		},
		{
			name: "empty log directory",
			config: &ChunkerConfig{
				LogLevel:     "INFO",
				LogFormat:    "console",
				LogDirectory: "",
			},
			expectErr: false, // Empty directory should be allowed (will use default)
		},
		{
			name: "whitespace-only log directory",
			config: &ChunkerConfig{
				LogLevel:     "INFO",
				LogFormat:    "console",
				LogDirectory: "   ",
			},
			expectErr: true, // Whitespace-only should be invalid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	config := testConfigWithLogging("debug", "console", "", true)

	chunker := NewMarkdownChunkerWithConfig(config)

	// Test that logger works by calling log methods
	// Note: kydenul/log doesn't provide SetOutput, so we just test that methods exist
	chunker.logger.Debug("Debug message")
	chunker.logger.Info("Info message")
	chunker.logger.Warn("Warning message")
	chunker.logger.Error("Error message")

	// If we get here without panicking, the logger is working
	t.Log("Logger methods executed successfully")
}

func TestDefaultConfigLogging(t *testing.T) {
	config := DefaultConfig()

	// Verify default logging configuration
	if config.LogLevel != "INFO" {
		t.Errorf("Expected default log level INFO, got %s", config.LogLevel)
	}
	if !config.EnableLog {
		t.Error("Expected logging to be enabled by default")
	}
	if config.LogFormat != "console" {
		t.Errorf("Expected default log format console, got %s", config.LogFormat)
	}
	if config.LogDirectory != "./logs" {
		t.Errorf("Expected default log directory ./logs, got %s", config.LogDirectory)
	}
}

func TestValidateConfigWithLogging(t *testing.T) {
	// Test that ValidateConfig includes log validation
	config := &ChunkerConfig{
		LogLevel:  "INVALID",
		LogFormat: "console",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Expected validation error for invalid log level")
	}

	if !strings.Contains(err.Error(), "invalid log level") {
		t.Errorf("Expected error message about invalid log level, got: %v", err)
	}
}

func TestLogDirectoryConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		logDirectory string
		expectError  bool
	}{
		{
			name:         "custom directory",
			logDirectory: "/tmp/test-logs",
			expectError:  false,
		},
		{
			name:         "relative directory",
			logDirectory: "./test-logs",
			expectError:  false,
		},
		{
			name:         "empty directory (should use default)",
			logDirectory: "",
			expectError:  false,
		},
		{
			name:         "current directory",
			logDirectory: ".",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ChunkerConfig{
				LogLevel:     "INFO",
				EnableLog:    true,
				LogFormat:    "console",
				LogDirectory: tt.logDirectory,
			}

			chunker := NewMarkdownChunkerWithConfig(config)

			// Verify logger is initialized
			if chunker.logger == nil {
				t.Fatal("Logger should be initialized")
			}

			// Test that chunker can process content
			content := []byte("# Test\n\nThis is a test.")
			chunks, err := chunker.ChunkDocument(content)
			if err != nil {
				t.Fatalf("Error processing document: %v", err)
			}

			if len(chunks) == 0 {
				t.Error("Expected at least one chunk")
			}
		})
	}
}
