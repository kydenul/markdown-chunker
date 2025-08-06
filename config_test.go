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
