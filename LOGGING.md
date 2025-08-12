# Logging Guide

This document provides a comprehensive guide to the logging features in the Markdown Chunker library.

## Overview

The Markdown Chunker library includes a comprehensive logging system that helps with debugging, monitoring, and performance analysis. The logging system is built on top of the `github.com/kydenul/log` library and provides structured, configurable logging throughout the document processing pipeline.

## Features

- **Multiple Log Levels**: DEBUG, INFO, WARN, ERROR
- **Multiple Formats**: Console (human-readable) and JSON (structured)
- **Configurable Output**: Customizable log directory and file naming
- **Rich Context**: Function names, line numbers, processing metrics
- **Performance Integration**: Automatic performance metrics logging
- **Error Context**: Detailed error information with full context
- **Thread-Safe**: Safe for concurrent use

## Configuration

### Basic Configuration

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "INFO"
config.LogFormat = "console"
config.LogDirectory = "./logs"
```

### Configuration Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `EnableLog` | `bool` | Enable/disable logging | `false` |
| `LogLevel` | `string` | Log level (DEBUG, INFO, WARN, ERROR) | `"ERROR"` |
| `LogFormat` | `string` | Log format (console, json) | `"console"` |
| `LogDirectory` | `string` | Directory for log files | `"./logs"` |

## Log Levels

### DEBUG

- **Purpose**: Detailed debugging information
- **Content**: Node processing details, metadata extraction, internal state
- **Use Case**: Development and troubleshooting
- **Performance Impact**: High (5x processing time)

```go
config.LogLevel = "DEBUG"
```

**Example Output:**

```LOG
2024-01-15 10:30:45 DEBUG [chunker.go:234] Processing heading node
2024-01-15 10:30:45 DEBUG [extractors.go:45] Extracting links from paragraph
2024-01-15 10:30:45 DEBUG [chunker.go:156] Node processing completed: heading level 2
```

### INFO

- **Purpose**: General processing information
- **Content**: Processing progress, chunk creation, performance metrics
- **Use Case**: Production monitoring and general logging
- **Performance Impact**: Medium (2x processing time)

```go
config.LogLevel = "INFO"
```

**Example Output:**

```LOG
2024-01-15 10:30:45 INFO  [chunker.go:123] Starting document processing
2024-01-15 10:30:45 INFO  [chunker.go:234] Document processing completed: 15 chunks, 2.3ms
2024-01-15 10:30:45 INFO  [chunker.go:245] Performance: 6521 chunks/sec, 2.1MB/sec
```

### WARN

- **Purpose**: Warning messages for potential issues
- **Content**: Memory warnings, size limit warnings, format issues
- **Use Case**: Monitoring potential problems
- **Performance Impact**: Low (minimal overhead)

```go
config.LogLevel = "WARN"
```

**Example Output:**

```LOG
2024-01-15 10:30:45 WARN  [chunker.go:189] Chunk size approaching limit: 950/1000 characters
2024-01-15 10:30:45 WARN  [memory.go:67] Memory usage high: 45MB/50MB limit
```

### ERROR

- **Purpose**: Error messages for processing failures
- **Content**: Error details, context information, stack traces
- **Use Case**: Error monitoring and debugging
- **Performance Impact**: Minimal (only on errors)

```go
config.LogLevel = "ERROR"
```

**Example Output:**

```LOG
2024-01-15 10:30:45 ERROR [chunker.go:198] Chunk size exceeded: 1200/1000 characters
2024-01-15 10:30:45 ERROR [parser.go:89] Failed to parse table: malformed header
```

## Log Formats

### Console Format

Human-readable format suitable for development and debugging.

```go
config.LogFormat = "console"
```

**Example:**

```LOG
2024-01-15 10:30:45 INFO  [chunker.go:123] Starting document processing
2024-01-15 10:30:45 DEBUG [chunker.go:145] Processing heading node: "Introduction"
2024-01-15 10:30:45 INFO  [chunker.go:234] Document processing completed: 15 chunks, 2.3ms
```

### JSON Format

Structured format suitable for log aggregation and analysis.

```go
config.LogFormat = "json"
```

**Example:**

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
    "config": {"max_chunk_size": 1000, "log_level": "INFO"}
  }
}
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "DEBUG",
  "message": "Processing heading node",
  "function": "processHeading",
  "file": "chunker.go",
  "line": 145,
  "context": {
    "node_type": "heading",
    "node_id": 1,
    "heading_level": 2,
    "content": "Introduction"
  }
}
```

## Usage Examples

### Basic Logging

```go
package main

import (
    "fmt"
    mc "github.com/kydenul/markdown-chunker"
)

func main() {
    config := mc.DefaultConfig()
    config.EnableLog = true
    config.LogLevel = "INFO"
    config.LogFormat = "console"
    config.LogDirectory = "./logs"

    chunker := mc.NewMarkdownChunkerWithConfig(config)
    chunks, err := chunker.ChunkDocument([]byte(markdown))
    if err != nil {
        panic(err)
    }

    fmt.Printf("Processed %d chunks\n", len(chunks))
    fmt.Printf("Logs written to: %s\n", config.LogDirectory)
}
```

### Debug Logging with JSON Format

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "DEBUG"
config.LogFormat = "json"
config.LogDirectory = "./debug-logs"

// Enable detailed metadata extraction for more debug info
config.CustomExtractors = []mc.MetadataExtractor{
    &mc.LinkExtractor{},
    &mc.ImageExtractor{},
    &mc.CodeComplexityExtractor{},
}

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))
```

### Error Logging with Context

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "ERROR"
config.LogFormat = "console"
config.LogDirectory = "./error-logs"
config.MaxChunkSize = 100  // Small limit to trigger errors
config.ErrorHandling = mc.ErrorModePermissive

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))

// Errors are logged automatically and can be retrieved
if chunker.HasErrors() {
    for _, err := range chunker.GetErrors() {
        fmt.Printf("Error: %s - %s\n", err.Type.String(), err.Message)
    }
}
```

### Performance Logging

```go
config := mc.DefaultConfig()
config.EnableLog = true
config.LogLevel = "INFO"
config.LogFormat = "json"
config.LogDirectory = "./perf-logs"
config.PerformanceMode = mc.PerformanceModeSpeedOptimized

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))

// Performance metrics are automatically logged
stats := chunker.GetPerformanceStats()
fmt.Printf("Processing time: %v\n", stats.ProcessingTime)
fmt.Printf("Memory used: %d KB\n", stats.MemoryUsed/1024)
```

## Log File Organization

### File Naming

Log files are automatically named with timestamps:

- `chunker-2024-01-15.log` (console format)
- `chunker-2024-01-15.json` (JSON format)

### Directory Structure

```LOG
logs/
├── chunker-2024-01-15.log
├── chunker-2024-01-15.json
└── archived/
    ├── chunker-2024-01-14.log
    └── chunker-2024-01-13.log
```

### Log Rotation

- Daily log rotation
- Automatic archiving of old logs
- Configurable retention period

## Performance Impact

| Log Level | Processing Time Impact | Memory Impact | Use Case |
|-----------|----------------------|---------------|----------|
| Disabled | Baseline | Baseline | Production (no logging) |
| ERROR | +0-5% | Minimal | Production (error only) |
| WARN | +5-10% | Low | Production (monitoring) |
| INFO | +50-100% | Medium | Development/Staging |
| DEBUG | +300-500% | High | Development/Debugging |

## Best Practices

### Production Use

```go
config.EnableLog = true
config.LogLevel = "ERROR"  // or "WARN"
config.LogFormat = "json"  // for log aggregation
config.LogDirectory = "/var/log/app"
```

### Development Use

```go
config.EnableLog = true
config.LogLevel = "DEBUG"
config.LogFormat = "console"  // human-readable
config.LogDirectory = "./dev-logs"
```

### Performance Testing

```go
config.EnableLog = false  // disable for accurate benchmarks
// or
config.LogLevel = "ERROR"  // minimal impact
```

### Log Analysis

```go
config.LogFormat = "json"  // structured for analysis
config.LogLevel = "INFO"   // good balance of detail/performance
```

## Integration with Other Features

### Error Handling

Logging is fully integrated with the error handling system:

```go
config.ErrorHandling = mc.ErrorModePermissive
config.LogLevel = "WARN"  // logs errors as warnings

chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))

// Errors are logged and can be retrieved
errors := chunker.GetErrors()
```

### Performance Monitoring

Performance metrics are automatically logged:

```go
config.PerformanceMode = mc.PerformanceModeSpeedOptimized
config.LogLevel = "INFO"

// Performance metrics appear in logs automatically
chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))
```

### Metadata Extraction

Metadata extraction details are logged at DEBUG level:

```go
config.CustomExtractors = []mc.MetadataExtractor{
    &mc.LinkExtractor{},
    &mc.CodeComplexityExtractor{},
}
config.LogLevel = "DEBUG"

// Extraction details appear in debug logs
chunker := mc.NewMarkdownChunkerWithConfig(config)
chunks, err := chunker.ChunkDocument([]byte(markdown))
```

## Troubleshooting

### No Log Files Created

1. Check if logging is enabled: `config.EnableLog = true`
2. Verify log directory permissions
3. Check log directory path exists or can be created

### Log Files Empty

1. Verify log level is appropriate for expected messages
2. Check if processing completed successfully
3. Ensure log directory is writable

### Performance Issues

1. Lower log level (ERROR < WARN < INFO < DEBUG)
2. Use JSON format for better performance
3. Consider disabling logging for performance-critical code

### Large Log Files

1. Implement log rotation
2. Use higher log levels (ERROR, WARN)
3. Monitor disk space usage

## Advanced Configuration

### Custom Log Directory Structure

```go
import (
    "path/filepath"
    "time"
)

// Create date-based subdirectories
today := time.Now().Format("2006-01-02")
config.LogDirectory = filepath.Join("./logs", today)
```

### Conditional Logging

```go
// Enable debug logging only in development
if os.Getenv("ENV") == "development" {
    config.LogLevel = "DEBUG"
} else {
    config.LogLevel = "ERROR"
}
```

### Multiple Log Configurations

```go
// Different configurations for different use cases
debugConfig := mc.DefaultConfig()
debugConfig.EnableLog = true
debugConfig.LogLevel = "DEBUG"
debugConfig.LogDirectory = "./debug"

prodConfig := mc.DefaultConfig()
prodConfig.EnableLog = true
prodConfig.LogLevel = "ERROR"
prodConfig.LogDirectory = "/var/log/app"
```

## Log Context Information

The logging system automatically includes rich context information:

### Function Context

- Function name where log was generated
- File name and line number
- Call stack information (for errors)

### Processing Context

- Document size and processing progress
- Current node type and ID
- Chunk count and processing statistics

### Configuration Context

- Active configuration parameters
- Enabled features and extractors
- Performance mode and settings

### Error Context

- Error type and severity
- Related configuration that may have caused the error
- Suggested remediation steps

This comprehensive logging system provides visibility into all aspects of the document processing pipeline, making it easier to debug issues, monitor performance, and understand the behavior of the chunking process.
