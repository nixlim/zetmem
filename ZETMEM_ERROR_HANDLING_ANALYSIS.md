# ZetMem Error Handling and Response Patterns Analysis

## Executive Summary

This document provides a comprehensive analysis of error handling and response patterns in the ZetMem MCP server implementation. The analysis identifies key patterns, error classification strategies, propagation mechanisms, and provides recommendations for implementing robust error handling in a Go MCP server template.

## 1. Error Types and Classification

### 1.1 MCP Protocol Error Codes

The system defines standard JSON-RPC 2.0 error codes in `pkg/models/mcp.go`:

```go
const (
    ParseError     = -32700  // Invalid JSON
    InvalidRequest = -32600  // Invalid request structure
    MethodNotFound = -32601  // Method does not exist
    InvalidParams  = -32602  // Invalid method parameters
    InternalError  = -32603  // Internal server error
)
```

### 1.2 Application-Level Error Categories

1. **Validation Errors**
   - Empty or invalid workspace IDs
   - Missing required parameters
   - Invalid character patterns in identifiers
   - Type conversion failures

2. **Service Integration Errors**
   - ChromaDB connection failures
   - LiteLLM API errors
   - Embedding service failures
   - HTTP request timeouts

3. **Business Logic Errors**
   - Workspace already exists
   - No memories found
   - Evolution process failures
   - Link generation failures

4. **System Errors**
   - Context cancellation
   - JSON marshaling/unmarshaling errors
   - File system errors
   - Configuration errors

## 2. Error Propagation Through Service Layers

### 2.1 Layered Architecture Pattern

```
MCP Server Layer (Top)
    ↓
Tool Execution Layer
    ↓
Memory System Layer
    ↓
Service Layer (ChromaDB, LiteLLM, Embedding)
    ↓
External APIs (Bottom)
```

### 2.2 Error Wrapping Strategy

The codebase uses Go's error wrapping with `fmt.Errorf` and `%w` verb:

```go
// Example from workspace.go
if err := w.ValidateWorkspaceID(workspaceID); err != nil {
    return nil, fmt.Errorf("invalid workspace ID: %w", err)
}

// Example from chromadb.go
if err != nil {
    return "", fmt.Errorf("failed to create request: %w", err)
}
```

### 2.3 Error Context Preservation

Errors are enriched with context at each layer:
- Original error message is preserved
- Additional context is added using descriptive prefixes
- Structured logging captures error details

## 3. MCP Error Response Formatting

### 3.1 Success Response Structure

```go
type MCPSuccessResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result"`
}
```

### 3.2 Error Response Structure

```go
type MCPErrorResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Error   MCPError    `json:"error"`
}

type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### 3.3 Tool-Level Error Handling

Tools return errors in MCPToolResult with IsError flag:

```go
return &models.MCPToolResult{
    IsError: true,
    Content: []models.MCPContent{{
        Type: "text",
        Text: fmt.Sprintf("Error: %v", err),
    }},
}, nil
```

## 4. Logging Strategies and Error Context

### 4.1 Structured Logging with Zap

The system uses structured logging throughout:

```go
logger.Error("Failed to store memory", 
    zap.String("memory_id", memoryID),
    zap.Error(err))

logger.Warn("LiteLLM call failed, retrying",
    zap.Int("attempt", i+1),
    zap.Error(err))
```

### 4.2 Log Levels for Errors

- **Fatal**: Configuration loading failures, initialization errors
- **Error**: Tool execution failures, service errors
- **Warn**: Retryable failures, fallback scenarios
- **Info**: Successful operations, state changes
- **Debug**: Detailed operation flow, API responses

### 4.3 Contextual Information

Logs include relevant context:
- Operation identifiers (memory_id, workspace_id)
- Retry attempts and fallback models
- Request/response sizes and token counts
- Timing information

## 5. User-Friendly Error Messages

### 5.1 Error Message Patterns

1. **Parameter Validation**
   ```go
   "Error: 'content' parameter is required and must be a string"
   "Error: 'query' parameter is required and must be a string"
   ```

2. **Service Failures**
   ```go
   "Failed to store memory: %v"
   "Failed to retrieve memories: %v"
   "Memory network evolution failed: %v"
   ```

3. **Not Found Scenarios**
   ```go
   "No relevant memories found for your query. Try adjusting your search terms or lowering the relevance threshold."
   ```

### 5.2 Error Message Guidelines

- Start with "Error:" for clear identification
- Include parameter names in validation errors
- Provide actionable suggestions when possible
- Avoid exposing internal implementation details

## 6. Recovery and Fallback Mechanisms

### 6.1 Retry Strategies

LiteLLM service implements exponential backoff:

```go
for i := 0; i < s.config.MaxRetries; i++ {
    response, err := s.call(ctx, prompt, s.config.DefaultModel)
    if err != nil {
        // Exponential backoff
        if i < s.config.MaxRetries-1 {
            time.Sleep(time.Second * time.Duration(1<<i))
        }
        continue
    }
}
```

### 6.2 Fallback Models

```go
// Try fallback models
for _, model := range s.config.FallbackModels {
    response, err := s.call(ctx, prompt, model)
    if err != nil {
        continue
    }
    return response, nil
}
```

### 6.3 Graceful Degradation

1. **Link Generation**: Continues without links on failure
2. **Embedding Service**: Falls back to hash-based embeddings
3. **Workspace Operations**: Creates default workspace if needed
4. **Evolution**: Logs warning but doesn't fail memory creation

## 7. Error Handling Patterns for Go MCP Server Template

### 7.1 Core Error Types

```go
package errors

import "fmt"

// MCPError represents a structured MCP protocol error
type MCPError struct {
    Code    int
    Message string
    Data    interface{}
}

func (e MCPError) Error() string {
    return fmt.Sprintf("MCP Error %d: %s", e.Code, e.Message)
}

// Common MCP errors
var (
    ErrParseError     = MCPError{Code: -32700, Message: "Parse error"}
    ErrInvalidRequest = MCPError{Code: -32600, Message: "Invalid request"}
    ErrMethodNotFound = MCPError{Code: -32601, Message: "Method not found"}
    ErrInvalidParams  = MCPError{Code: -32602, Message: "Invalid params"}
    ErrInternalError  = MCPError{Code: -32603, Message: "Internal error"}
)

// ServiceError wraps service-level errors with context
type ServiceError struct {
    Service string
    Op      string
    Err     error
}

func (e ServiceError) Error() string {
    return fmt.Sprintf("%s.%s: %v", e.Service, e.Op, e.Err)
}

func (e ServiceError) Unwrap() error {
    return e.Err
}
```

### 7.2 Error Handler Middleware

```go
package middleware

import (
    "context"
    "errors"
    "runtime/debug"
    
    "go.uber.org/zap"
)

type ErrorHandler struct {
    logger *zap.Logger
}

func (h *ErrorHandler) Handle(ctx context.Context, fn func() error) error {
    defer func() {
        if r := recover(); r != nil {
            h.logger.Error("Panic recovered",
                zap.Any("panic", r),
                zap.String("stack", string(debug.Stack())))
        }
    }()
    
    return fn()
}
```

### 7.3 Tool Error Wrapper

```go
package tools

type ToolError struct {
    Tool    string
    Message string
    Cause   error
    IsUser  bool // true if user error, false if system error
}

func (e ToolError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Tool, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Tool, e.Message)
}

func (e ToolError) Unwrap() error {
    return e.Cause
}

// Helper functions
func NewUserError(tool, message string) error {
    return ToolError{Tool: tool, Message: message, IsUser: true}
}

func NewSystemError(tool, message string, cause error) error {
    return ToolError{Tool: tool, Message: message, Cause: cause, IsUser: false}
}
```

### 7.4 Error Response Builder

```go
package response

import (
    "errors"
    "github.com/yourdomain/mcp-server/pkg/errors"
)

type ErrorResponseBuilder struct {
    logger *zap.Logger
}

func (b *ErrorResponseBuilder) Build(err error) MCPErrorResponse {
    var mcpErr errors.MCPError
    var toolErr errors.ToolError
    
    switch {
    case errors.As(err, &mcpErr):
        return MCPErrorResponse{
            JSONRPC: "2.0",
            Error: MCPError{
                Code:    mcpErr.Code,
                Message: mcpErr.Message,
                Data:    mcpErr.Data,
            },
        }
        
    case errors.As(err, &toolErr):
        if toolErr.IsUser {
            return MCPErrorResponse{
                JSONRPC: "2.0",
                Error: MCPError{
                    Code:    errors.ErrInvalidParams.Code,
                    Message: toolErr.Message,
                },
            }
        }
        
        b.logger.Error("Tool error", 
            zap.String("tool", toolErr.Tool),
            zap.Error(toolErr.Cause))
            
        return MCPErrorResponse{
            JSONRPC: "2.0",
            Error: MCPError{
                Code:    errors.ErrInternalError.Code,
                Message: "Internal error occurred",
                Data:    map[string]string{"tool": toolErr.Tool},
            },
        }
        
    default:
        b.logger.Error("Unhandled error", zap.Error(err))
        return MCPErrorResponse{
            JSONRPC: "2.0",
            Error: MCPError{
                Code:    errors.ErrInternalError.Code,
                Message: "Internal server error",
            },
        }
    }
}
```

### 7.5 Service Error Handling Pattern

```go
package services

import (
    "context"
    "fmt"
    "time"
)

type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn func() error) error {
    delay := cfg.InitialDelay
    
    for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if attempt == cfg.MaxAttempts {
            return fmt.Errorf("after %d attempts: %w", attempt, err)
        }
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            delay = time.Duration(float64(delay) * cfg.Multiplier)
            if delay > cfg.MaxDelay {
                delay = cfg.MaxDelay
            }
        }
    }
    
    return fmt.Errorf("retry failed")
}
```

## 8. Best Practices and Recommendations

### 8.1 Error Handling Guidelines

1. **Always wrap errors with context** using `fmt.Errorf` with `%w`
2. **Use structured error types** for different error categories
3. **Log errors at appropriate levels** based on severity and recoverability
4. **Include relevant context** in error messages and logs
5. **Implement retry logic** for transient failures
6. **Provide fallback mechanisms** where appropriate
7. **Sanitize error messages** shown to users

### 8.2 Testing Error Scenarios

1. **Unit tests** for each error condition
2. **Integration tests** for service failures
3. **Chaos testing** for network and service disruptions
4. **Load testing** to identify timeout scenarios

### 8.3 Monitoring and Alerting

1. **Track error rates** by type and service
2. **Monitor retry attempts** and fallback usage
3. **Alert on error spikes** or sustained failures
4. **Dashboard for error trends** and patterns

## 9. Conclusion

The ZetMem error handling implementation demonstrates several robust patterns:

- Clear error classification and consistent error codes
- Contextual error wrapping throughout the call stack
- Structured logging with relevant metadata
- User-friendly error messages with actionable guidance
- Retry mechanisms with exponential backoff
- Graceful degradation and fallback strategies

These patterns provide a solid foundation for implementing error handling in any Go-based MCP server, ensuring reliability, debuggability, and a good user experience even when things go wrong.