# ZetMem API Patterns and Tool Interfaces Documentation

## Overview

This document provides comprehensive documentation for creating extensible tool interfaces in a Go MCP (Model Context Protocol) server, based on the ZetMem implementation. It covers protocol implementation details, JSON-RPC compliance, tool hierarchy patterns, API versioning strategies, and request/response lifecycle management.

## Table of Contents

1. [MCP Protocol Implementation Details](#mcp-protocol-implementation-details)
2. [JSON-RPC 2.0 Compliance and Message Structure](#json-rpc-20-compliance-and-message-structure)
3. [Tool Interface Hierarchy and Extension Patterns](#tool-interface-hierarchy-and-extension-patterns)
4. [API Versioning and Backward Compatibility](#api-versioning-and-backward-compatibility)
5. [Request/Response Lifecycle Patterns](#requestresponse-lifecycle-patterns)
6. [Tool Discovery and Metadata Exposure](#tool-discovery-and-metadata-exposure)
7. [Extensible Tool Interface Patterns](#extensible-tool-interface-patterns)

## 1. MCP Protocol Implementation Details

### Core Server Architecture

The ZetMem MCP server implements a tool registry pattern with the following key components:

```go
// Server represents the MCP server
type Server struct {
    logger      *zap.Logger
    tools       map[string]Tool
    initialized bool
    reader      *bufio.Reader
    writer      io.Writer
}
```

### Key Protocol Features

1. **Stream-based I/O**: Uses buffered readers/writers for efficient JSON-RPC message processing
2. **Tool Registry**: Dynamic tool registration system for extensibility
3. **Protocol Version**: Fixed at "2024-11-05" for MCP compliance
4. **Initialization State**: Tracks server initialization status for proper request validation

### Server Lifecycle

```go
// Server initialization and startup pattern
func main() {
    // 1. Initialize services and dependencies
    logger := initLogger()
    services := initializeServices(config)
    
    // 2. Create MCP server
    mcpServer := mcp.NewServer(logger)
    
    // 3. Register tools dynamically
    mcpServer.RegisterTool(NewTool1(services))
    mcpServer.RegisterTool(NewTool2(services))
    
    // 4. Start server with context for graceful shutdown
    ctx := context.WithCancel(context.Background())
    mcpServer.Start(ctx)
}
```

## 2. JSON-RPC 2.0 Compliance and Message Structure

### Message Type Hierarchy

ZetMem implements full JSON-RPC 2.0 compliance with distinct message types:

```go
// Request (with ID) - Requires response
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// Notification (no ID) - No response expected
type MCPNotification struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// Success Response
type MCPSuccessResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result"`
}

// Error Response
type MCPErrorResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Error   MCPError    `json:"error"`
}
```

### Message Parsing Strategy

```go
func (s *Server) handleRequest(ctx context.Context) error {
    line, err := s.reader.ReadString('\n')
    if err != nil {
        return err
    }

    // Try parsing as request first (has ID)
    var request MCPRequest
    if err := json.Unmarshal([]byte(line), &request); err == nil && request.ID != nil {
        return s.handleJSONRPCRequest(ctx, request)
    }

    // Try parsing as notification (no ID)
    var notification MCPNotification
    if err := json.Unmarshal([]byte(line), &notification); err == nil {
        return s.handleJSONRPCNotification(notification)
    }

    // Invalid JSON - send error only for requests
    s.sendError(nil, ParseError, "Invalid JSON", nil)
    return nil
}
```

### Standard Error Codes

```go
const (
    ParseError     = -32700  // Invalid JSON
    InvalidRequest = -32600  // Invalid request structure
    MethodNotFound = -32601  // Unknown method
    InvalidParams  = -32602  // Invalid method parameters
    InternalError  = -32603  // Internal server error
)
```

## 3. Tool Interface Hierarchy and Extension Patterns

### Base Tool Interface

```go
// Minimal tool interface - all tools must implement
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*MCPToolResult, error)
}
```

### Enhanced Tool Interface

```go
// Extended interface for tools with strategic guidance
type EnhancedTool interface {
    Tool
    UsageTriggers() []string              // When to use this tool
    BestPractices() []string              // How to use effectively
    Synergies() map[string][]string       // Tool relationships
    WorkflowSnippets() []map[string]interface{} // Example workflows
}
```

### Dynamic Interface Detection

```go
func (s *Server) handleListTools(request MCPRequest) error {
    tools := make([]MCPTool, 0, len(s.tools))
    
    for _, tool := range s.tools {
        mcpTool := MCPTool{
            Name:        tool.Name(),
            Description: tool.Description(),
            InputSchema: tool.InputSchema(),
        }
        
        // Check for enhanced capabilities
        if enhancedTool, ok := tool.(EnhancedTool); ok {
            mcpTool.UsageTriggers = enhancedTool.UsageTriggers()
            mcpTool.BestPractices = enhancedTool.BestPractices()
            mcpTool.Synergies = enhancedTool.Synergies()
            mcpTool.WorkflowSnippets = enhancedTool.WorkflowSnippets()
        }
        
        tools = append(tools, mcpTool)
    }
    
    return s.sendResponse(request.ID, map[string]interface{}{
        "tools": tools,
    })
}
```

## 4. API Versioning and Backward Compatibility

### Protocol Version Management

```go
// Fixed protocol version for MCP compliance
const MCPProtocolVersion = "2024-11-05"

// Server version follows semantic versioning
const ServerVersion = "1.0.0"

func (s *Server) handleInitialize(request MCPRequest) error {
    result := map[string]interface{}{
        "protocolVersion": MCPProtocolVersion,
        "serverInfo": map[string]interface{}{
            "name":    "ZetMem MCP Server",
            "version": ServerVersion,
        },
        "capabilities": map[string]interface{}{
            "tools": map[string]interface{}{},
        },
    }
    return s.sendResponse(request.ID, result)
}
```

### Backward Compatibility Patterns

#### 1. Field Deprecation Pattern

```go
type StoreMemoryRequest struct {
    Content     string `json:"content"`
    WorkspaceID string `json:"workspace_id"`      // New field
    ProjectPath string `json:"project_path"`      // Deprecated but supported
    CodeType    string `json:"code_type"`
}

func (t *StoreCodingMemoryTool) Execute(ctx context.Context, args map[string]interface{}) (*MCPToolResult, error) {
    // Support both old and new field names
    if workspaceID, ok := args["workspace_id"].(string); ok {
        req.WorkspaceID = workspaceID
    } else if projectPath, ok := args["project_path"].(string); ok {
        // Fallback to deprecated field
        req.WorkspaceID = projectPath
    }
}
```

#### 2. Optional Field Introduction

```go
func (t *Tool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "required_field": map[string]interface{}{
                "type": "string",
                "description": "Always required",
            },
            "new_optional_field": map[string]interface{}{
                "type": "string",
                "description": "Added in v1.1.0 (optional)",
            },
        },
        "required": []string{"required_field"}, // New fields not in required
    }
}
```

#### 3. Tool Aliasing for Compatibility

```go
// Register tool with multiple names for compatibility
oldTool := NewStoreCodingMemoryTool(system, logger)
mcpServer.RegisterTool(oldTool)
mcpServer.RegisterTool(&ToolAlias{Tool: oldTool, name: "store_memory"}) // Old name
```

## 5. Request/Response Lifecycle Patterns

### Complete Request Flow

```go
// 1. Message Reception
line, err := s.reader.ReadString('\n')

// 2. Message Parsing
var request MCPRequest
json.Unmarshal([]byte(line), &request)

// 3. Method Routing
switch request.Method {
case MethodInitialize:
    return s.handleInitialize(request)
case MethodListTools:
    return s.handleListTools(request)
case MethodCallTool:
    return s.handleCallTool(ctx, request)
default:
    s.sendError(request.ID, MethodNotFound, "Unknown method", nil)
}

// 4. Tool Execution
func (s *Server) handleCallTool(ctx context.Context, request MCPRequest) error {
    // Validate server state
    if !s.initialized {
        s.sendError(request.ID, InvalidRequest, "Server not initialized", nil)
        return nil
    }
    
    // Parse parameters
    params := request.Params.(map[string]interface{})
    toolName := params["name"].(string)
    arguments := params["arguments"].(map[string]interface{})
    
    // Find and execute tool
    tool, exists := s.tools[toolName]
    if !exists {
        s.sendError(request.ID, MethodNotFound, "Tool not found", nil)
        return nil
    }
    
    result, err := tool.Execute(ctx, arguments)
    if err != nil {
        s.sendError(request.ID, InternalError, err.Error(), nil)
        return nil
    }
    
    return s.sendResponse(request.ID, result)
}
```

### Error Handling Strategy

```go
// Consistent error response pattern
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
    response := MCPErrorResponse{
        JSONRPC: "2.0",
        ID:      id,
        Error: MCPError{
            Code:    code,
            Message: message,
            Data:    data,
        },
    }
    
    responseData, _ := json.Marshal(response)
    fmt.Fprintf(s.writer, "%s\n", responseData)
    return nil
}

// Tool execution with error handling
func (t *Tool) Execute(ctx context.Context, args map[string]interface{}) (*MCPToolResult, error) {
    // Return MCP-formatted error result
    if err := validateArgs(args); err != nil {
        return &MCPToolResult{
            IsError: true,
            Content: []MCPContent{{
                Type: "text",
                Text: fmt.Sprintf("Validation error: %v", err),
            }},
        }, nil
    }
    
    // Successful execution
    return &MCPToolResult{
        Content: []MCPContent{{
            Type: "text",
            Text: "Success message",
        }},
    }, nil
}
```

## 6. Tool Discovery and Metadata Exposure

### Rich Tool Discovery

```go
type MCPTool struct {
    // Basic metadata
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

### Tool Metadata Example

```go
func (t *StoreCodingMemoryTool) UsageTriggers() []string {
    return []string{
        "After solving a non-trivial problem",
        "When discovering new patterns or techniques",
        "After implementing significant features",
        "When gaining insights about the codebase",
        "Before context switches to preserve knowledge",
    }
}

func (t *StoreCodingMemoryTool) Synergies() map[string][]string {
    return map[string][]string{
        "precedes": {"retrieve_relevant_memories", "evolve_memory_network"},
        "succeeds": {"workspace_init", "retrieve_relevant_memories"},
    }
}

func (t *StoreCodingMemoryTool) WorkflowSnippets() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "goal": "Store a problem-solution pattern",
            "steps": []string{
                "1. store_coding_memory with problem and solution",
                "2. Include context explaining why it works",
                "3. Specify code_type and workspace_id",
            },
        },
    }
}
```

### Strategy Guide Integration

```go
func (s *Server) handleListTools(request MCPRequest) error {
    // ... tool collection logic ...
    
    result := map[string]interface{}{
        "tools": tools,
        "strategyGuideSummary": "This server follows the Zetmem strategic principles...",
    }
    
    return s.sendResponse(request.ID, result)
}
```

## 7. Extensible Tool Interface Patterns

### Pattern 1: Composition-Based Tool Design

```go
// Base tool implementation with common functionality
type BaseTool struct {
    name        string
    description string
    logger      *zap.Logger
}

func (b *BaseTool) Name() string        { return b.name }
func (b *BaseTool) Description() string { return b.description }

// Specific tool embeds base and adds functionality
type MyCustomTool struct {
    BaseTool
    service MyService
}

func (t *MyCustomTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "custom_param": map[string]interface{}{
                "type": "string",
            },
        },
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, args map[string]interface{}) (*MCPToolResult, error) {
    // Custom implementation
}
```

### Pattern 2: Service Injection

```go
// Tools receive services through constructor injection
type AnalysisTool struct {
    llm        LLMService
    embeddings EmbeddingService
    storage    StorageService
    logger     *zap.Logger
}

func NewAnalysisTool(llm LLMService, embeddings EmbeddingService, 
                     storage StorageService, logger *zap.Logger) *AnalysisTool {
    return &AnalysisTool{
        llm:        llm,
        embeddings: embeddings,
        storage:    storage,
        logger:     logger,
    }
}
```

### Pattern 3: Middleware Pattern

```go
// Tool middleware for cross-cutting concerns
type ToolMiddleware func(Tool) Tool

// Logging middleware
func WithLogging(logger *zap.Logger) ToolMiddleware {
    return func(tool Tool) Tool {
        return &LoggingTool{
            Tool:   tool,
            logger: logger,
        }
    }
}

type LoggingTool struct {
    Tool
    logger *zap.Logger
}

func (l *LoggingTool) Execute(ctx context.Context, args map[string]interface{}) (*MCPToolResult, error) {
    l.logger.Info("Tool execution started", 
        zap.String("tool", l.Name()),
        zap.Any("args", args))
    
    result, err := l.Tool.Execute(ctx, args)
    
    if err != nil {
        l.logger.Error("Tool execution failed", 
            zap.String("tool", l.Name()),
            zap.Error(err))
    } else {
        l.logger.Info("Tool execution completed", 
            zap.String("tool", l.Name()))
    }
    
    return result, err
}

// Usage
tool := NewMyTool(services)
tool = WithLogging(logger)(tool)
tool = WithMetrics(metrics)(tool)
server.RegisterTool(tool)
```

### Pattern 4: Schema Validation Pattern

```go
// Reusable schema components
var CommonSchemas = map[string]interface{}{
    "workspace_id": map[string]interface{}{
        "type":        "string",
        "description": "Workspace identifier",
    },
    "code_type": map[string]interface{}{
        "type":        "string",
        "description": "Programming language",
        "enum":        []string{"go", "python", "javascript", "typescript"},
    },
}

// Schema builder with validation
type SchemaBuilder struct {
    properties map[string]interface{}
    required   []string
}

func NewSchemaBuilder() *SchemaBuilder {
    return &SchemaBuilder{
        properties: make(map[string]interface{}),
        required:   []string{},
    }
}

func (b *SchemaBuilder) AddProperty(name string, schema interface{}, required bool) *SchemaBuilder {
    b.properties[name] = schema
    if required {
        b.required = append(b.required, name)
    }
    return b
}

func (b *SchemaBuilder) Build() map[string]interface{} {
    return map[string]interface{}{
        "type":       "object",
        "properties": b.properties,
        "required":   b.required,
    }
}

// Usage in tool
func (t *MyTool) InputSchema() map[string]interface{} {
    return NewSchemaBuilder().
        AddProperty("workspace_id", CommonSchemas["workspace_id"], true).
        AddProperty("code_type", CommonSchemas["code_type"], false).
        AddProperty("content", map[string]interface{}{
            "type": "string",
            "description": "Content to process",
        }, true).
        Build()
}
```

### Pattern 5: Tool Versioning

```go
// Versioned tool interface
type VersionedTool interface {
    Tool
    Version() string
    DeprecatedIn() string
    ReplacedBy() string
}

// Registry with version support
type VersionedToolRegistry struct {
    tools map[string][]VersionedTool
}

func (r *VersionedToolRegistry) RegisterTool(tool VersionedTool) {
    name := tool.Name()
    r.tools[name] = append(r.tools[name], tool)
    // Sort by version
}

func (r *VersionedToolRegistry) GetTool(name string, version string) (Tool, error) {
    versions := r.tools[name]
    if version == "" {
        // Return latest non-deprecated version
        for i := len(versions) - 1; i >= 0; i-- {
            if versions[i].DeprecatedIn() == "" {
                return versions[i], nil
            }
        }
    }
    // Find specific version
}
```

### Pattern 6: Configuration-Driven Tools

```go
// Tool configuration
type ToolConfig struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  []ParameterConfig      `json:"parameters"`
    Handler     string                 `json:"handler"`
}

type ParameterConfig struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Required    bool   `json:"required"`
    Description string `json:"description"`
}

// Dynamic tool creation from config
func CreateToolFromConfig(config ToolConfig, handlers map[string]Handler) Tool {
    return &ConfigurableTool{
        config:  config,
        handler: handlers[config.Handler],
    }
}

type ConfigurableTool struct {
    config  ToolConfig
    handler Handler
}

func (t *ConfigurableTool) InputSchema() map[string]interface{} {
    builder := NewSchemaBuilder()
    for _, param := range t.config.Parameters {
        schema := map[string]interface{}{
            "type":        param.Type,
            "description": param.Description,
        }
        builder.AddProperty(param.Name, schema, param.Required)
    }
    return builder.Build()
}
```

### Pattern 7: Result Builder Pattern

```go
// Fluent result builder
type ResultBuilder struct {
    content []MCPContent
    isError bool
}

func NewResultBuilder() *ResultBuilder {
    return &ResultBuilder{
        content: []MCPContent{},
    }
}

func (b *ResultBuilder) AddText(text string) *ResultBuilder {
    b.content = append(b.content, MCPContent{
        Type: "text",
        Text: text,
    })
    return b
}

func (b *ResultBuilder) AddError(err error) *ResultBuilder {
    b.isError = true
    b.content = append(b.content, MCPContent{
        Type: "text",
        Text: fmt.Sprintf("Error: %v", err),
    })
    return b
}

func (b *ResultBuilder) Build() *MCPToolResult {
    return &MCPToolResult{
        Content: b.content,
        IsError: b.isError,
    }
}

// Usage in tool
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (*MCPToolResult, error) {
    builder := NewResultBuilder()
    
    // Process...
    if err != nil {
        return builder.AddError(err).Build(), nil
    }
    
    return builder.
        AddText("Operation completed successfully").
        AddText(fmt.Sprintf("Processed %d items", count)).
        Build(), nil
}
```

## Summary

The ZetMem implementation provides a comprehensive template for building extensible Go-based MCP servers with:

1. **Full MCP Protocol Compliance**: Proper JSON-RPC 2.0 message handling with request/notification distinction
2. **Extensible Tool System**: Base and enhanced tool interfaces with dynamic capability detection
3. **Robust Error Handling**: Consistent error responses following JSON-RPC standards
4. **Rich Tool Discovery**: Comprehensive metadata exposure including usage guidance and workflow examples
5. **Backward Compatibility**: Field deprecation and optional field patterns
6. **Service-Oriented Architecture**: Dependency injection and composition patterns
7. **Middleware Support**: Cross-cutting concerns like logging and metrics

This architecture enables developers to create sophisticated MCP tools that are maintainable, testable, and extensible while maintaining protocol compliance and providing excellent developer experience.