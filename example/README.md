# Markdown Chunker Examples

This directory contains comprehensive examples demonstrating all features of the Markdown Chunker library.

## Examples Overview

### 1. Basic Configuration (`config_example/`)

Demonstrates basic configuration options and usage patterns.

**Features shown:**

- Default configuration
- Custom content type filtering
- Custom metadata extractors
- Chunk size limits
- **Logging functionality** (NEW)

**Run:**

```bash
cd config_example
go run config_example.go
```

### 2. Advanced Configuration (`advanced_configuration/`)

Shows advanced configuration options and complex use cases.

**Features shown:**

- Content type filtering
- Custom metadata extractors
- Error handling modes
- Performance modes
- Chunk size limits
- **Logging configuration** (NEW)
- Complete advanced configuration

**Run:**

```bash
cd advanced_configuration
go run advanced_config_example.go
```

### 3. Comprehensive Features (`comprehensive_features/`)

Demonstrates all library features with complex test documents.

**Features shown:**

- Basic usage
- Advanced features
- Error handling and recovery
- Performance monitoring
- **Logging features** (NEW)
- Metadata extraction
- Content analysis

**Run:**

```bash
cd comprehensive_features
go run comprehensive_example.go
```

### 4. Error Handling (`error_handling/`)

Focuses on error handling capabilities and recovery strategies.

**Run:**

```bash
cd error_handling
go run error_example.go
```

### 5. Performance Optimization (`performance_optimization/`)

Shows performance optimization techniques and monitoring.

**Run:**

```bash
cd performance_optimization
go run performance_example.go
```

### 6. Table Processing (`table_processing/`)

Demonstrates advanced table processing and analysis features.

**Run:**

```bash
cd table_processing
go run table_example.go
```

### 7. Logging Features (`logging_features/`)

**Comprehensive logging functionality demonstration.**

**Features shown:**

- Basic logging configuration
- Different log levels (DEBUG, INFO, WARN, ERROR)
- Different log formats (console, JSON)
- Error logging with context
- Performance logging
- Custom log directories
- Logging with complex configurations

**Run:**

```bash
cd logging_features
go run logging_example.go
```

### 8. Custom Strategy (`custom_strategy/`) - NEW

**Demonstrates custom chunking strategy development.**

**Features shown:**

- Custom strategy implementation
- Strategy builder usage
- Rule-based chunking logic
- Strategy registration and usage
- Performance comparison with built-in strategies

**Run:**

```bash
cd custom_strategy
go run custom_strategy_example.go
```

### 9. Configuration Migration (`config_migration/`) - NEW

**Shows how to migrate existing configurations to the new strategy system.**

**Features shown:**

- Configuration migration helpers
- Backward compatibility
- Strategy configuration
- Migration validation

**Run:**

```bash
cd config_migration
go run config_migration_example.go
```

### 10. Strategy Examples (`strategy_examples/`) - NEW

**Comprehensive demonstration of all chunking strategies.**

**Features shown:**

- All built-in strategies (element-level, hierarchical, document-level)
- Dynamic strategy switching
- Strategy performance comparison
- Size constraints and filtering
- Error handling with strategies
- Best practices for strategy selection

**Run:**

```bash
cd strategy_examples
go run strategy_examples.go
```

## Chunking Strategies - NEW

The library now supports multiple chunking strategies for different use cases:

### Built-in Strategies

#### Element-Level Strategy (Default)
Processes each Markdown element individually, maintaining the original behavior.

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.ElementLevelConfig()
```

#### Hierarchical Strategy
Groups content by heading levels, creating chunks that contain a heading and all its subordinate content.

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.HierarchicalConfig(3) // Max depth of 3 levels
```

#### Document-Level Strategy
Treats the entire document as a single chunk.

```go
config := mc.DefaultConfig()
config.ChunkingStrategy = mc.DocumentLevelConfig()
```

### Custom Strategies

Create custom strategies using the builder pattern:

```go
builder := mc.NewCustomStrategyBuilder("my-strategy", "Custom chunking logic")
builder.AddRule(
    mc.HeadingLevelCondition{MinLevel: 1, MaxLevel: 2},
    mc.CreateSeparateChunkAction{},
    10, // High priority
)
customStrategy := builder.Build()
```

### Strategy Selection Guide

| Use Case | Recommended Strategy | Reason |
|----------|---------------------|---------|
| Search indexing | Element-Level | Fine-grained matching |
| Documentation | Hierarchical | Preserves section context |
| RAG systems | Hierarchical | Maintains logical structure |
| Document classification | Document-Level | Complete context |
| Code analysis | Custom | Specialized requirements |

## Logging Features

The library includes comprehensive logging capabilities:

### Log Levels

- **DEBUG**: Detailed debugging information (5x performance impact)
- **INFO**: General processing information (2x performance impact)
- **WARN**: Warning messages (minimal impact)
- **ERROR**: Error messages only (minimal impact)

### Log Formats

- **console**: Human-readable format for development
- **json**: Structured format for log aggregation

### Configuration

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "INFO"        // DEBUG, INFO, WARN, ERROR
config.LogFormat = "console"    // console, json
config.LogDirectory = "./logs"  // Log file directory
```

### Log Output Examples

#### Console Format

```LOG
2024-01-15 10:30:45 INFO  [chunker.go:123] Starting document processing
2024-01-15 10:30:45 DEBUG [chunker.go:145] Processing heading node: "Introduction"
2024-01-15 10:30:45 INFO  [chunker.go:234] Document processing completed: 15 chunks, 2.3ms
```

#### JSON Format

```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "message": "Starting document processing",
  "function": "ChunkDocument",
  "file": "chunker.go",
  "line": 123,
  "context": {
    "document_size": 1024,
    "config": {"max_chunk_size": 1000}
  }
}
```

## Running All Examples

To run all examples and see the logging functionality:

```bash
# Basic configuration with logging
cd config_example && go run config_example.go

# Advanced configuration with logging
cd ../advanced_configuration && go run advanced_config_example.go

# Comprehensive features including logging
cd ../comprehensive_features && go run comprehensive_example.go

# Dedicated logging features demo
cd ../logging_features && go run logging_example.go
```

## Log Files

After running the examples, you'll find log files in various directories:

- `./example-logs/` - Basic configuration logs
- `./debug-logs/` - Debug level logs
- `./demo-logs/` - Demonstration logs from logging examples
- `./custom-logs/` - Custom directory examples

## Performance Impact

| Log Level | Performance Impact | Use Case |
|-----------|-------------------|----------|
| Disabled | Baseline | Production (no logging) |
| ERROR | +0-5% | Production (error only) |
| WARN | +5-10% | Production (monitoring) |
| INFO | +50-100% | Development/Staging |
| DEBUG | +300-500% | Development/Debugging |

## Best Practices

### Production

```go
config.EnableLog = true
config.LogLevel = "ERROR"  // or "WARN"
config.LogFormat = "json"  // for log aggregation
```

### Development

```go
config.EnableLog = true
config.LogLevel = "DEBUG"
config.LogFormat = "console"  // human-readable
```

### Performance Testing

```go
config.EnableLog = false  // disable for accurate benchmarks
```

## Documentation

For detailed logging documentation, see [LOGGING.md](../LOGGING.md) in the project root.

For general library documentation, see [README.md](../README.md) in the project root.
