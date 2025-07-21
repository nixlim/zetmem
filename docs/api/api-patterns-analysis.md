# ZetMem API Patterns and Tool Interfaces Documentation

## Executive Summary

This document provides a comprehensive analysis of the ZetMem MCP server's API patterns and tool interfaces, focusing on MCP protocol implementation, JSON-RPC 2.0 compliance, tool interface hierarchy, API versioning strategies, request/response lifecycle patterns, and tool discovery mechanisms. The analysis provides patterns for creating extensible tool interfaces in a Go MCP server template.

## Table of Contents

1. [MCP Protocol Implementation Details](#mcp-protocol-implementation-details)
2. [JSON-RPC 2.0 Compliance and Message Structure](#json-rpc-20-compliance-and-message-structure)
3. [Tool Interface Hierarchy and Extension Patterns](#tool-interface-hierarchy-and-extension-patterns)
4. [API Versioning and Backward Compatibility](#api-versioning-and-backward-compatibility)
5. [Request/Response Lifecycle Patterns](#requestresponse-lifecycle-patterns)
6. [Tool Discovery and Metadata Exposure](#tool-discovery-and-metadata-exposure)
7. [Patterns for Creating Extensible Tool Interfaces](#patterns-for-creating-extensible-tool-interfaces)

## MCP Protocol Implementation Details

### Core Server Architecture

The ZetMem MCP server implements the Model Context Protocol through a clean, modular architecture:

```go
// pkg/mcp/server.go
type Server struct {
    logger      *zap.Logger
    tools       map[string]Tool
    initialized bool
    reader      *bufio.Reader
    writer      io.Writer
}
```

**Key Design Decisions:**

1. **Tool Registry Pattern**: Tools are stored in a map for O(1) lookup performance
2. **Initialization State**: Server tracks initialization to enforce MCP protocol requirements
3. **Stream-Based I/O**: Uses buffered readers for efficient JSON-RPC message processing
4. **Structured Logging**: Integrated zap logger for comprehensive observability

### Protocol Version Management

The server declares MCP protocol version "2024-11-05" during initialization:

```go
result := map[string]interface{}{
    "protocolVersion": "2024-11-05",
    "capabilities": map[string]interface{}{
        "tools": map[string]interface{}{},
    },
    "serverInfo": map[string]interface{}{
        "name":    "ZetMem MCP Server",
        "version": "1.0.0",
    },
}
```

**Version Strategy:**
- Protocol version is hardcoded to ensure compatibility
- Server version follows semantic versioning
- Capabilities object allows future feature negotiation

## JSON-RPC 2.0 Compliance and Message Structure

### Request Handling Architecture

The server implements a robust message parsing strategy that handles both requests and notifications:

```go
// Handle requests with ID
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// Handle notifications without ID
type MCPNotification struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}
```

**Parsing Strategy:**
1. First attempt to parse as request (with ID)
2. If that fails, try parsing as notification (no ID)
3. Send error response only for malformed requests (not notifications)

### Response Message Patterns

The server uses separate response types for success and error cases:

```go
// Success responses
type MCPSuccessResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result"`
}

// Error responses
type MCPErrorResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Error   MCPError    `json:"error"`
}
```

**Error Code Standards:**
```go
const (
    ParseError     = -32700  // Invalid JSON
    InvalidRequest = -32600  // Invalid request structure
    MethodNotFound = -32601  // Unknown method
    InvalidParams  = -32602  // Invalid method parameters
    InternalError  = -32603  // Internal server error
)
```

## Tool Interface Hierarchy and Extension Patterns

### Base Tool Interface

The system defines a minimal base interface for all tools:

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error)
}
```

### Enhanced Tool Interface

For tools requiring strategic guidance capabilities:

```go
type EnhancedTool interface {
    Tool
    UsageTriggers() []string
    BestPractices() []string
    Synergies() map[string][]string
    WorkflowSnippets() []map[string]interface{}
}
```

**Extension Benefits:**
- **UsageTriggers**: Provides AI agents with context on when to use the tool
- **BestPractices**: Offers guidance for optimal tool usage
- **Synergies**: Maps relationships with other tools
- **WorkflowSnippets**: Provides example usage patterns

### Tool Registration Pattern

Dynamic tool detection of enhanced capabilities:

```go
func (s *Server) handleListTools(request models.MCPRequest) error {
    tools := make([]models.MCPTool, 0, len(s.tools))
    for _, tool := range s.tools {
        mcpTool := models.MCPTool{
            Name:        tool.Name(),
            Description: tool.Description(),
            InputSchema: tool.InputSchema(),
        }
        
        // Dynamic interface detection
        if enhancedTool, ok := tool.(EnhancedTool); ok {
            mcpTool.UsageTriggers = enhancedTool.UsageTriggers()
            mcpTool.BestPractices = enhancedTool.BestPractices()
            mcpTool.Synergies = enhancedTool.Synergies()
            mcpTool.WorkflowSnippets = enhancedTool.WorkflowSnippets()
        }
        
        tools = append(tools, mcpTool)
    }
    // ...
}
```

## API Versioning and Backward Compatibility

### Current Versioning Strategy

1. **Protocol Version**: Fixed at "2024-11-05" for MCP compliance
2. **Server Version**: Semantic versioning at "1.0.0"
3. **Tool Versions**: Not currently versioned individually

### Backward Compatibility Patterns

#### 1. Field Deprecation Pattern

```go
// Example from memory tools
"project_path": {
    "type": "string",
    "description": "Optional project path for context (deprecated: use workspace_id)"
}
```

**Pattern Benefits:**
- Old fields remain functional
- Clear migration path indicated
- No breaking changes for existing clients

#### 2. Optional Field Introduction

All new fields are introduced as optional to maintain compatibility:

```go
"include_strategy_guide": {
    "type": "boolean",
    "description": "Whether to include the complete strategy guide in response (default: true)",
    "default": true
}
```

#### 3. Tool Aliasing (Recommended Pattern)

For future tool evolution:

```go
// Register both old and new versions
server.RegisterTool(NewStoreCodingMemoryTool())      // v1
server.RegisterTool(NewStoreCodingMemoryV2Tool())    // v2
server.RegisterTool(NewStoreMemoryTool())            // Latest alias
```

## Request/Response Lifecycle Patterns

### Complete Request Flow

1. **Message Reception**
   ```go
   line, err := s.reader.ReadString('\n')
   ```

2. **Message Classification**
   - Parse as request (has ID) → Route to request handler
   - Parse as notification (no ID) → Route to notification handler
   - Parse failure → Send error response

3. **Method Routing**
   ```go
   switch request.Method {
   case models.MethodInitialize:
       return s.handleInitialize(request)
   case models.MethodListTools:
       return s.handleListTools(request)
   case models.MethodCallTool:
       return s.handleCallTool(ctx, request)
   default:
       s.sendError(request.ID, models.MethodNotFound, ...)
   }
   ```

4. **Tool Execution**
   ```go
   // Validate tool exists
   tool, exists := s.tools[toolName]
   if !exists {
       s.sendError(request.ID, models.MethodNotFound, ...)
       return nil
   }
   
   // Execute with context
   result, err := tool.Execute(ctx, arguments)
   ```

5. **Response Formatting**
   ```go
   type MCPToolResult struct {
       Content []MCPContent `json:"content"`
       IsError bool         `json:"isError,omitempty"`
   }
   ```

### Error Handling Patterns

Consistent error responses throughout the lifecycle:

```go
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
    response := models.MCPErrorResponse{
        JSONRPC: "2.0",
        ID:      id,
        Error: models.MCPError{
            Code:    code,
            Message: message,
            Data:    data,
        },
    }
    // ...
}
```

## Tool Discovery and Metadata Exposure

### Discovery Mechanism

The `tools/list` method provides comprehensive tool discovery:

```go
func (s *Server) handleListTools(request models.MCPRequest) error {
    if !s.initialized {
        s.sendError(request.ID, models.InvalidRequest, "Server not initialized", nil)
        return nil
    }
    
    // Build tool list with metadata
    tools := make([]models.MCPTool, 0, len(s.tools))
    // ... tool collection ...
    
    result := map[string]interface{}{
        "tools": tools,
        "strategyGuideSummary": "This server follows the Zetmem strategic principles...",
    }
    
    return s.sendResponse(request.ID, result)
}
```

### Tool Metadata Structure

Each tool exposes rich metadata:

```go
type MCPTool struct {
    // Core MCP fields
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
    
    // Enhanced metadata (optional)
    UsageTriggers    []string                   `json:"usageTriggers,omitempty"`
    BestPractices    []string                   `json:"bestPractices,omitempty"`
    Synergies        map[string][]string        `json:"synergies,omitempty"`
    WorkflowSnippets []map[string]interface{}   `json:"workflowSnippets,omitempty"`
}
```

### Input Schema Definition Pattern

Tools define JSON Schema for input validation:

```go
func (t *StoreCodingMemoryTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "content": map[string]interface{}{
                "type":        "string",
                "description": "The code content or coding context to store",
            },
            "workspace_id": map[string]interface{}{
                "type":        "string",
                "description": "Workspace identifier (path or name) for organizing memories",
            },
            // ... more properties ...
        },
        "required": []string{"content"},
    }
}
```

## Patterns for Creating Extensible Tool Interfaces

### 1. Composition-Based Tool Design

```go
// Base tool implementation
type BaseTool struct {
    name        string
    description string
    logger      *zap.Logger
}

// Composed tool with additional capabilities
type MyCustomTool struct {
    BaseTool
    service     MyService
    config      MyConfig
}

// Implement required methods
func (t *MyCustomTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            // Define schema
        },
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    // Implementation
}
```

### 2. Service Injection Pattern

```go
// Tool constructor with dependency injection
func NewMyTool(service MyService, config MyConfig, logger *zap.Logger) *MyTool {
    return &MyTool{
        service: service,
        config:  config,
        logger:  logger,
    }
}

// Main server setup
func main() {
    // Initialize services
    service := NewMyService()
    
    // Create tools with injected services
    tool := NewMyTool(service, config, logger)
    
    // Register tools
    server.RegisterTool(tool)
}
```

### 3. Middleware Pattern for Cross-Cutting Concerns

```go
// Tool middleware interface
type ToolMiddleware interface {
    Before(ctx context.Context, args map[string]interface{}) error
    After(ctx context.Context, result *models.MCPToolResult, err error) error
}

// Example: Logging middleware
type LoggingMiddleware struct {
    logger *zap.Logger
}

func (m *LoggingMiddleware) Before(ctx context.Context, args map[string]interface{}) error {
    m.logger.Info("Tool execution started", zap.Any("args", args))
    return nil
}

// Wrapped tool execution
type MiddlewareEnabledTool struct {
    Tool
    middlewares []ToolMiddleware
}

func (t *MiddlewareEnabledTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    // Run before middlewares
    for _, m := range t.middlewares {
        if err := m.Before(ctx, args); err != nil {
            return nil, err
        }
    }
    
    // Execute actual tool
    result, err := t.Tool.Execute(ctx, args)
    
    // Run after middlewares
    for _, m := range t.middlewares {
        if err := m.After(ctx, result, err); err != nil {
            return nil, err
        }
    }
    
    return result, err
}
```

### 4. Schema Validation Pattern

```go
// Reusable schema components
var CommonSchemas = map[string]interface{}{
    "workspace_id": map[string]interface{}{
        "type":        "string",
        "description": "Workspace identifier",
    },
    "pagination": map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "page": map[string]interface{}{
                "type":    "integer",
                "minimum": 1,
                "default": 1,
            },
            "limit": map[string]interface{}{
                "type":    "integer",
                "minimum": 1,
                "maximum": 100,
                "default": 10,
            },
        },
    },
}

// Schema builder helper
func BuildSchema(properties map[string]interface{}, required []string) map[string]interface{} {
    return map[string]interface{}{
        "type":       "object",
        "properties": properties,
        "required":   required,
    }
}
```

### 5. Tool Versioning Pattern

```go
// Version-aware tool interface
type VersionedTool interface {
    Tool
    Version() string
    DeprecationNotice() string
}

// Implementation
type MyToolV2 struct {
    // ... fields ...
}

func (t *MyToolV2) Name() string {
    return "my_tool" // Same name for compatibility
}

func (t *MyToolV2) Version() string {
    return "2.0.0"
}

func (t *MyToolV2) DeprecationNotice() string {
    return "" // Not deprecated
}

// Server can expose version info
func (s *Server) handleListTools(request models.MCPRequest) error {
    // ... existing code ...
    
    if versionedTool, ok := tool.(VersionedTool); ok {
        mcpTool.Version = versionedTool.Version()
        mcpTool.Deprecated = versionedTool.DeprecationNotice()
    }
}
```

### 6. Configuration-Driven Tool Creation

```go
// Tool factory pattern
type ToolFactory interface {
    CreateTool(config ToolConfig) (Tool, error)
}

// Configuration structure
type ToolConfig struct {
    Name        string                 `yaml:"name"`
    Type        string                 `yaml:"type"`
    Enabled     bool                   `yaml:"enabled"`
    Settings    map[string]interface{} `yaml:"settings"`
}

// Factory implementation
type DefaultToolFactory struct {
    services map[string]interface{}
    logger   *zap.Logger
}

func (f *DefaultToolFactory) CreateTool(config ToolConfig) (Tool, error) {
    if !config.Enabled {
        return nil, nil
    }
    
    switch config.Type {
    case "memory_store":
        return NewMemoryStoreTool(f.services["memory"], f.logger), nil
    case "workspace":
        return NewWorkspaceTool(f.services["workspace"], f.logger), nil
    default:
        return nil, fmt.Errorf("unknown tool type: %s", config.Type)
    }
}
```

### 7. Result Builder Pattern

```go
// Fluent result builder
type ResultBuilder struct {
    content []models.MCPContent
    isError bool
}

func NewResultBuilder() *ResultBuilder {
    return &ResultBuilder{
        content: make([]models.MCPContent, 0),
    }
}

func (b *ResultBuilder) AddText(text string) *ResultBuilder {
    b.content = append(b.content, models.MCPContent{
        Type: "text",
        Text: text,
    })
    return b
}

func (b *ResultBuilder) AddError(err error) *ResultBuilder {
    b.isError = true
    return b.AddText(fmt.Sprintf("Error: %v", err))
}

func (b *ResultBuilder) Build() *models.MCPToolResult {
    return &models.MCPToolResult{
        Content: b.content,
        IsError: b.isError,
    }
}

// Usage in tool
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    builder := NewResultBuilder()
    
    // Process...
    builder.AddText("Operation completed successfully")
    
    if err != nil {
        return builder.AddError(err).Build(), nil
    }
    
    return builder.Build(), nil
}
```

## Best Practices Summary

1. **Protocol Compliance**
   - Always validate server initialization state
   - Handle both requests and notifications properly
   - Use standard JSON-RPC error codes

2. **Tool Design**
   - Keep base interface minimal
   - Use composition for extended capabilities
   - Implement proper input validation

3. **Versioning**
   - Deprecate fields rather than removing them
   - Add new fields as optional with defaults
   - Consider tool aliasing for major changes

4. **Error Handling**
   - Return errors in result for tool-level issues
   - Use JSON-RPC errors for protocol-level issues
   - Provide descriptive error messages

5. **Extensibility**
   - Use dependency injection
   - Implement middleware for cross-cutting concerns
   - Support configuration-driven tool creation

6. **Documentation**
   - Include rich metadata in tool definitions
   - Provide usage examples in workflow snippets
   - Document field deprecations clearly

## Conclusion

The ZetMem MCP server demonstrates a well-architected approach to implementing the Model Context Protocol in Go. Its patterns for tool interfaces, protocol compliance, and extensibility provide a solid foundation for building MCP-compliant servers. The combination of clean interfaces, dependency injection, and metadata-rich tool definitions creates a system that is both powerful and maintainable.

Key takeaways for implementing MCP servers:
- Strict adherence to JSON-RPC 2.0 specification
- Clear separation between protocol handling and business logic
- Rich tool metadata for improved AI agent interactions
- Extensible architecture supporting tool evolution
- Comprehensive error handling at all layers

These patterns can serve as a template for creating new MCP servers or extending existing ones with additional capabilities.