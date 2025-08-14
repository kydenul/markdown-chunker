package markdownchunker

import (
	"errors"
	"testing"
	"time"

	"github.com/yuin/goldmark/ast"
)

// MockFailingStrategy 模拟失败的策略
type MockFailingStrategy struct {
	name        string
	shouldFail  bool
	failureType string
}

func (s *MockFailingStrategy) GetName() string {
	return s.name
}

func (s *MockFailingStrategy) GetDescription() string {
	return "Mock strategy for testing error handling"
}

func (s *MockFailingStrategy) ChunkDocument(doc ast.Node, source []byte, chunker *MarkdownChunker) ([]Chunk, error) {
	if s.shouldFail {
		switch s.failureType {
		case "panic":
			panic("mock strategy panic")
		case "timeout":
			time.Sleep(100 * time.Millisecond)
			return nil, errors.New("strategy timeout")
		case "invalid_output":
			// 返回无效输出
			return []Chunk{
				{ID: -1, Type: "invalid", Content: "content1", Text: "text1", Metadata: nil},
				{ID: 1, Type: "invalid", Content: "content2", Text: "text2", Metadata: nil},
				{ID: 1, Type: "paragraph", Content: "duplicate id", Text: "duplicate id", Metadata: nil}, // 重复ID
			}, nil
		default:
			return nil, errors.New("mock strategy failure")
		}
	}

	// 正常情况返回简单的块
	return []Chunk{
		{
			ID:       0,
			Type:     "paragraph",
			Content:  string(source),
			Text:     string(source),
			Metadata: make(map[string]string),
		},
	}, nil
}

func (s *MockFailingStrategy) ValidateConfig(config *StrategyConfig) error {
	return nil
}

func (s *MockFailingStrategy) Clone() ChunkingStrategy {
	return &MockFailingStrategy{
		name:        s.name,
		shouldFail:  s.shouldFail,
		failureType: s.failureType,
	}
}

func TestStrategyErrorHandling_StrictMode(t *testing.T) {
	config := DefaultConfig()
	config.ErrorHandling = ErrorModeStrict

	chunker := NewMarkdownChunkerWithConfig(config)

	// 注册失败的策略
	failingStrategy := &MockFailingStrategy{
		name:       "failing-strategy",
		shouldFail: true,
	}

	err := chunker.RegisterStrategy(failingStrategy)
	if err != nil {
		t.Fatalf("Failed to register failing strategy: %v", err)
	}

	// 设置失败的策略
	err = chunker.SetStrategy("failing-strategy", nil)
	if err != nil {
		t.Fatalf("Failed to set failing strategy: %v", err)
	}

	// 尝试分块，应该失败
	content := []byte("# Test\nContent")
	chunks, err := chunker.ChunkDocument(content)

	if err == nil {
		t.Error("Expected error in strict mode, but got none")
	}

	if len(chunks) != 0 {
		t.Errorf("Expected no chunks in strict mode, got %d", len(chunks))
	}

	// 检查错误类型
	if chunkerErr, ok := err.(*ChunkerError); ok {
		if chunkerErr.Type != ErrorTypeStrategyExecutionFailed {
			t.Errorf("Expected ErrorTypeStrategyExecutionFailed, got %v", chunkerErr.Type)
		}
	} else {
		t.Error("Expected ChunkerError type")
	}
}

func TestStrategyErrorHandling_PermissiveMode(t *testing.T) {
	config := DefaultConfig()
	config.ErrorHandling = ErrorModePermissive

	chunker := NewMarkdownChunkerWithConfig(config)

	// 注册失败的策略
	failingStrategy := &MockFailingStrategy{
		name:       "failing-strategy",
		shouldFail: true,
	}

	err := chunker.RegisterStrategy(failingStrategy)
	if err != nil {
		t.Fatalf("Failed to register failing strategy: %v", err)
	}

	// 设置失败的策略
	err = chunker.SetStrategy("failing-strategy", nil)
	if err != nil {
		t.Fatalf("Failed to set failing strategy: %v", err)
	}

	// 尝试分块，应该恢复到默认策略
	content := []byte("# Test\nContent")
	chunks, err := chunker.ChunkDocument(content)
	if err != nil {
		t.Errorf("Expected no error in permissive mode, got: %v", err)
	}

	if len(chunks) == 0 {
		t.Error("Expected chunks after recovery, got none")
	}

	// 检查是否恢复到默认策略
	strategyName, _ := chunker.GetCurrentStrategy()
	if strategyName != "element-level" {
		t.Errorf("Expected recovery to element-level strategy, got %s", strategyName)
	}
}

func TestStrategyErrorHandling_SilentMode(t *testing.T) {
	config := DefaultConfig()
	config.ErrorHandling = ErrorModeSilent

	chunker := NewMarkdownChunkerWithConfig(config)

	// 注册失败的策略
	failingStrategy := &MockFailingStrategy{
		name:       "failing-strategy",
		shouldFail: true,
	}

	err := chunker.RegisterStrategy(failingStrategy)
	if err != nil {
		t.Fatalf("Failed to register failing strategy: %v", err)
	}

	// 设置失败的策略
	err = chunker.SetStrategy("failing-strategy", nil)
	if err != nil {
		t.Fatalf("Failed to set failing strategy: %v", err)
	}

	// 尝试分块，应该静默恢复
	content := []byte("# Test\nContent")
	chunks, err := chunker.ChunkDocument(content)
	if err != nil {
		t.Errorf("Expected no error in silent mode, got: %v", err)
	}

	// 在静默模式下，即使恢复失败也应该返回空结果而不是错误
	if chunks == nil {
		t.Error("Expected non-nil chunks slice in silent mode")
	}
}

func TestStrategyErrorHandling_InvalidOutput(t *testing.T) {
	config := DefaultConfig()
	config.ErrorHandling = ErrorModePermissive

	chunker := NewMarkdownChunkerWithConfig(config)

	// 注册返回无效输出的策略
	invalidStrategy := &MockFailingStrategy{
		name:        "invalid-output-strategy",
		shouldFail:  true,
		failureType: "invalid_output",
	}

	err := chunker.RegisterStrategy(invalidStrategy)
	if err != nil {
		t.Fatalf("Failed to register invalid output strategy: %v", err)
	}

	// 设置策略
	err = chunker.SetStrategy("invalid-output-strategy", nil)
	if err != nil {
		t.Fatalf("Failed to set invalid output strategy: %v", err)
	}

	// 尝试分块
	content := []byte("# Test\nContent")
	chunks, err := chunker.ChunkDocument(content)
	if err != nil {
		t.Errorf("Expected no error after output sanitization, got: %v", err)
	}

	// 检查输出是否被修复
	if len(chunks) == 0 {
		t.Error("Expected sanitized chunks, got none")
	}

	// 检查ID是否被修复
	idMap := make(map[int]bool)
	for _, chunk := range chunks {
		if idMap[chunk.ID] {
			t.Errorf("Found duplicate ID after sanitization: %d", chunk.ID)
		}
		idMap[chunk.ID] = true

		if chunk.ID < 0 {
			t.Errorf("Found negative ID after sanitization: %d", chunk.ID)
		}

		// 检查是否添加了修复标记
		if chunk.Metadata["sanitized"] != "true" {
			t.Error("Expected sanitized metadata to be set")
		}
	}
}

func TestStrategyErrorHandling_RecoveryFailure(t *testing.T) {
	config := DefaultConfig()
	config.ErrorHandling = ErrorModePermissive

	chunker := NewMarkdownChunkerWithConfig(config)

	// 清空策略注册器，使恢复失败
	chunker.strategyRegistry = NewStrategyRegistry()

	// 设置一个空策略
	chunker.strategy = nil

	// 尝试分块
	content := []byte("# Test\nContent")
	chunks, err := chunker.ChunkDocument(content)

	if err == nil {
		t.Error("Expected error when recovery fails, got none")
	}

	if len(chunks) != 0 {
		t.Errorf("Expected no chunks when recovery fails, got %d", len(chunks))
	}
}

func TestValidateStrategyOutput(t *testing.T) {
	chunker := NewMarkdownChunker()

	// 测试有效输出
	validChunks := []Chunk{
		{ID: 0, Type: "heading", Content: "# Test", Text: "Test", Metadata: make(map[string]string)},
		{ID: 1, Type: "paragraph", Content: "Content", Text: "Content", Metadata: make(map[string]string)},
	}

	context := make(map[string]interface{})
	err := chunker.validateStrategyOutput(validChunks, context)
	if err != nil {
		t.Errorf("Expected no error for valid output, got: %v", err)
	}

	// 测试重复ID
	invalidChunks := []Chunk{
		{ID: 0, Type: "heading", Content: "# Test", Text: "Test", Metadata: make(map[string]string)},
		{ID: 0, Type: "paragraph", Content: "Content", Text: "Content", Metadata: make(map[string]string)},
	}

	err = chunker.validateStrategyOutput(invalidChunks, context)
	if err == nil {
		t.Error("Expected error for duplicate IDs, got none")
	}
}

func TestSanitizeStrategyOutput(t *testing.T) {
	config := DefaultConfig()
	config.FilterEmptyChunks = true

	chunker := NewMarkdownChunkerWithConfig(config)

	// 创建有问题的输出
	problematicChunks := []Chunk{
		{ID: -1, Type: "invalid", Content: "Valid content", Text: "Valid content", Metadata: nil},
		{ID: 1, Type: "paragraph", Content: "", Text: "", Metadata: make(map[string]string)},         // 空块
		{ID: 1, Type: "heading", Content: "# Test", Text: "Test", Metadata: make(map[string]string)}, // 重复ID
	}

	sanitized := chunker.sanitizeStrategyOutput(problematicChunks)

	// 检查结果
	if len(sanitized) != 2 { // 空块应该被过滤
		t.Errorf("Expected 2 chunks after sanitization, got %d", len(sanitized))
	}

	// 检查ID是否被修复
	idMap := make(map[int]bool)
	for _, chunk := range sanitized {
		if idMap[chunk.ID] {
			t.Errorf("Found duplicate ID after sanitization: %d", chunk.ID)
		}
		idMap[chunk.ID] = true

		if chunk.ID < 0 {
			t.Errorf("Found negative ID after sanitization: %d", chunk.ID)
		}

		// 检查类型是否被修复
		validTypes := map[string]bool{
			"heading": true, "paragraph": true, "code": true,
			"table": true, "list": true, "blockquote": true,
			"thematic_break": true, "document": true,
		}
		if !validTypes[chunk.Type] {
			t.Errorf("Found invalid type after sanitization: %s", chunk.Type)
		}

		// 检查元数据是否被初始化
		if chunk.Metadata == nil {
			t.Error("Metadata should be initialized after sanitization")
		}

		// 检查修复标记
		if chunk.Metadata["sanitized"] != "true" {
			t.Error("Expected sanitized metadata to be set")
		}
	}
}

func TestCreateStrategyExecutionContext(t *testing.T) {
	config := DefaultConfig()
	config.MaxChunkSize = 1000
	config.FilterEmptyChunks = true
	config.PerformanceMode = PerformanceModeSpeedOptimized

	chunker := NewMarkdownChunkerWithConfig(config)

	// 设置策略
	err := chunker.SetStrategy("element-level", nil)
	if err != nil {
		t.Fatalf("Failed to set strategy: %v", err)
	}

	content := []byte("# Test\nContent")

	// 创建简单的AST节点用于测试
	doc := ast.NewDocument()

	context := chunker.createStrategyExecutionContext(doc, content)

	// 检查上下文内容
	expectedKeys := []string{
		"function", "document_size_bytes", "strategy", "error_handling_mode",
		"max_chunk_size", "filter_empty_chunks", "performance_mode",
	}

	for _, key := range expectedKeys {
		if _, exists := context[key]; !exists {
			t.Errorf("Expected context key %s not found", key)
		}
	}

	// 检查具体值
	if context["document_size_bytes"] != len(content) {
		t.Errorf("Expected document_size_bytes %d, got %v", len(content), context["document_size_bytes"])
	}

	if context["strategy"] != "element-level" {
		t.Errorf("Expected strategy element-level, got %v", context["strategy"])
	}

	if context["max_chunk_size"] != 1000 {
		t.Errorf("Expected max_chunk_size 1000, got %v", context["max_chunk_size"])
	}
}
