# Go MCP Server Template

## Quick Start Template

This template provides a production-ready starting point for building MCP servers in Go, based on the ZetMem patterns.

### Project Structure

```
my-mcp-server/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── pkg/
│   ├── mcp/
│   │   ├── server.go         # Core MCP server
│   │   ├── interfaces.go     # Tool interfaces
│   │   └── middleware.go     # Middleware support
│   ├── models/
│   │   └── mcp.go           # MCP protocol models
│   ├── tools/
│   │   ├── base.go          # Base tool implementation
│   │   └── example.go       # Example tool
│   └── services/
│       └── example.go       # Business logic services
├── config/
│   └── config.yaml          # Configuration
├── go.mod
└── README.md
```

### Core Files

#### `pkg/models/mcp.go` - Protocol Models

```go
package models

// JSON-RPC 2.0 Request
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// JSON-RPC 2.0 Notification
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

type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Tool Result
type MCPToolResult struct {
    Content []MCPContent `json:"content"`
    IsError bool         `json:"isError,omitempty"`
}

type MCPContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// Tool Definition
type MCPTool struct {
    Name             string                   `json:"name"`
    Description      string                   `json:"description"`
    InputSchema      map[string]interface{}   `json:"inputSchema"`
    UsageTriggers    []string                 `json:"usageTriggers,omitempty"`
    BestPractices    []string                 `json:"bestPractices,omitempty"`
    Synergies        map[string][]string      `json:"synergies,omitempty"`
    WorkflowSnippets []map[string]interface{} `json:"workflowSnippets,omitempty"`
}

// Standard error codes
const (
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
)

// Method names
const (
    MethodInitialize = "initialize"
    MethodListTools  = "tools/list"
    MethodCallTool   = "tools/call"
)
```

#### `pkg/mcp/interfaces.go` - Tool Interfaces

```go
package mcp

import (
    "context"
    "github.com/yourusername/my-mcp-server/pkg/models"
)

// Base tool interface
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error)
}

// Enhanced tool with additional metadata
type EnhancedTool interface {
    Tool
    UsageTriggers() []string
    BestPractices() []string
    Synergies() map[string][]string
    WorkflowSnippets() []map[string]interface{}
}

// Versioned tool support
type VersionedTool interface {
    Tool
    Version() string
    DeprecatedIn() string
    ReplacedBy() string
}

// Tool middleware
type ToolMiddleware func(Tool) Tool

// Tool factory
type ToolFactory interface {
    CreateTool(config map[string]interface{}) (Tool, error)
}
```

#### `pkg/mcp/server.go` - Core Server Implementation

```go
package mcp

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"
    
    "github.com/yourusername/my-mcp-server/pkg/models"
    "go.uber.org/zap"
)

type Server struct {
    logger      *zap.Logger
    tools       map[string]Tool
    initialized bool
    reader      *bufio.Reader
    writer      io.Writer
    
    // Server metadata
    serverName    string
    serverVersion string
}

func NewServer(logger *zap.Logger, name, version string) *Server {
    return &Server{
        logger:        logger,
        tools:         make(map[string]Tool),
        reader:        bufio.NewReader(os.Stdin),
        writer:        os.Stdout,
        serverName:    name,
        serverVersion: version,
    }
}

func (s *Server) RegisterTool(tool Tool) {
    s.tools[tool.Name()] = tool
    s.logger.Info("Registered tool", zap.String("name", tool.Name()))
}

func (s *Server) Start(ctx context.Context) error {
    s.logger.Info("Starting MCP server", 
        zap.String("name", s.serverName),
        zap.String("version", s.serverVersion))
    
    for {
        select {
        case <-ctx.Done():
            s.logger.Info("Server shutting down")
            return ctx.Err()
        default:
            if err := s.handleRequest(ctx); err != nil {
                if err == io.EOF {
                    s.logger.Info("Client disconnected")
                    return nil
                }
                s.logger.Error("Error handling request", zap.Error(err))
                continue
            }
        }
    }
}

func (s *Server) handleRequest(ctx context.Context) error {
    line, err := s.reader.ReadString('\n')
    if err != nil {
        return err
    }
    
    // Try parsing as request
    var request models.MCPRequest
    if err := json.Unmarshal([]byte(line), &request); err == nil && request.ID != nil {
        s.logger.Debug("Received request",
            zap.String("method", request.Method),
            zap.Any("id", request.ID))
        return s.handleJSONRPCRequest(ctx, request)
    }
    
    // Try parsing as notification
    var notification models.MCPNotification
    if err := json.Unmarshal([]byte(line), &notification); err == nil {
        s.logger.Debug("Received notification",
            zap.String("method", notification.Method))
        return s.handleJSONRPCNotification(notification)
    }
    
    s.sendError(nil, models.ParseError, "Invalid JSON", nil)
    return nil
}

func (s *Server) handleJSONRPCRequest(ctx context.Context, request models.MCPRequest) error {
    switch request.Method {
    case models.MethodInitialize:
        return s.handleInitialize(request)
    case models.MethodListTools:
        return s.handleListTools(request)
    case models.MethodCallTool:
        return s.handleCallTool(ctx, request)
    default:
        s.sendError(request.ID, models.MethodNotFound,
            fmt.Sprintf("Method not found: %s", request.Method), nil)
    }
    return nil
}

func (s *Server) handleJSONRPCNotification(notification models.MCPNotification) error {
    switch notification.Method {
    case "notifications/initialized":
        s.logger.Debug("Client initialization complete")
        return nil
    default:
        s.logger.Debug("Unknown notification received",
            zap.String("method", notification.Method))
        return nil
    }
}

func (s *Server) handleInitialize(request models.MCPRequest) error {
    s.initialized = true
    
    result := map[string]interface{}{
        "protocolVersion": "2024-11-05",
        "capabilities": map[string]interface{}{
            "tools": map[string]interface{}{},
        },
        "serverInfo": map[string]interface{}{
            "name":    s.serverName,
            "version": s.serverVersion,
        },
    }
    
    return s.sendResponse(request.ID, result)
}

func (s *Server) handleListTools(request models.MCPRequest) error {
    if !s.initialized {
        s.sendError(request.ID, models.InvalidRequest, "Server not initialized", nil)
        return nil
    }
    
    tools := make([]models.MCPTool, 0, len(s.tools))
    for _, tool := range s.tools {
        mcpTool := models.MCPTool{
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
    
    result := map[string]interface{}{
        "tools": tools,
    }
    
    return s.sendResponse(request.ID, result)
}

func (s *Server) handleCallTool(ctx context.Context, request models.MCPRequest) error {
    if !s.initialized {
        s.sendError(request.ID, models.InvalidRequest, "Server not initialized", nil)
        return nil
    }
    
    params, ok := request.Params.(map[string]interface{})
    if !ok {
        s.sendError(request.ID, models.InvalidParams, "Invalid params", nil)
        return nil
    }
    
    toolName, ok := params["name"].(string)
    if !ok {
        s.sendError(request.ID, models.InvalidParams, "Tool name required", nil)
        return nil
    }
    
    tool, exists := s.tools[toolName]
    if !exists {
        s.sendError(request.ID, models.MethodNotFound,
            fmt.Sprintf("Tool not found: %s", toolName), nil)
        return nil
    }
    
    arguments, ok := params["arguments"].(map[string]interface{})
    if !ok {
        arguments = make(map[string]interface{})
    }
    
    s.logger.Info("Executing tool",
        zap.String("tool", toolName),
        zap.Any("arguments", arguments))
    
    result, err := tool.Execute(ctx, arguments)
    if err != nil {
        s.logger.Error("Tool execution failed",
            zap.String("tool", toolName),
            zap.Error(err))
        s.sendError(request.ID, models.InternalError, err.Error(), nil)
        return nil
    }
    
    return s.sendResponse(request.ID, result)
}

func (s *Server) sendResponse(id interface{}, result interface{}) error {
    response := models.MCPSuccessResponse{
        JSONRPC: "2.0",
        ID:      id,
        Result:  result,
    }
    
    data, err := json.Marshal(response)
    if err != nil {
        return err
    }
    
    _, err = fmt.Fprintf(s.writer, "%s\n", data)
    return err
}

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
    
    responseData, err := json.Marshal(response)
    if err != nil {
        return err
    }
    
    _, err = fmt.Fprintf(s.writer, "%s\n", responseData)
    return err
}
```

#### `pkg/tools/base.go` - Base Tool Implementation

```go
package tools

import (
    "context"
    "fmt"
    
    "github.com/yourusername/my-mcp-server/pkg/models"
    "go.uber.org/zap"
)

// BaseTool provides common functionality
type BaseTool struct {
    name        string
    description string
    logger      *zap.Logger
}

func NewBaseTool(name, description string, logger *zap.Logger) BaseTool {
    return BaseTool{
        name:        name,
        description: description,
        logger:      logger,
    }
}

func (b *BaseTool) Name() string        { return b.name }
func (b *BaseTool) Description() string { return b.description }

// ResultBuilder for fluent result construction
type ResultBuilder struct {
    content []models.MCPContent
    isError bool
}

func NewResultBuilder() *ResultBuilder {
    return &ResultBuilder{
        content: []models.MCPContent{},
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
    b.content = append(b.content, models.MCPContent{
        Type: "text",
        Text: fmt.Sprintf("Error: %v", err),
    })
    return b
}

func (b *ResultBuilder) Build() *models.MCPToolResult {
    return &models.MCPToolResult{
        Content: b.content,
        IsError: b.isError,
    }
}

// Schema builder for input validation
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

func (b *SchemaBuilder) AddProperty(name string, propType string, description string, required bool) *SchemaBuilder {
    b.properties[name] = map[string]interface{}{
        "type":        propType,
        "description": description,
    }
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
```

#### `pkg/tools/example.go` - Example Tool Implementation

```go
package tools

import (
    "context"
    "fmt"
    
    "github.com/yourusername/my-mcp-server/pkg/models"
    "go.uber.org/zap"
)

// ExampleTool demonstrates tool implementation
type ExampleTool struct {
    BaseTool
    // Add service dependencies here
}

func NewExampleTool(logger *zap.Logger) *ExampleTool {
    return &ExampleTool{
        BaseTool: NewBaseTool(
            "example_tool",
            "An example tool that demonstrates the pattern",
            logger,
        ),
    }
}

func (t *ExampleTool) InputSchema() map[string]interface{} {
    return NewSchemaBuilder().
        AddProperty("input", "string", "Input text to process", true).
        AddProperty("mode", "string", "Processing mode", false).
        Build()
}

func (t *ExampleTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    builder := NewResultBuilder()
    
    // Extract and validate arguments
    input, ok := args["input"].(string)
    if !ok {
        return builder.AddError(fmt.Errorf("'input' parameter is required")).Build(), nil
    }
    
    mode, _ := args["mode"].(string)
    if mode == "" {
        mode = "default"
    }
    
    // Simulate processing
    t.logger.Info("Processing input",
        zap.String("input", input),
        zap.String("mode", mode))
    
    // Return success result
    return builder.
        AddText(fmt.Sprintf("Processed input: %s", input)).
        AddText(fmt.Sprintf("Mode: %s", mode)).
        Build(), nil
}

// Optional: Implement EnhancedTool interface
func (t *ExampleTool) UsageTriggers() []string {
    return []string{
        "When you need to process text",
        "For demonstration purposes",
    }
}

func (t *ExampleTool) BestPractices() []string {
    return []string{
        "Always provide the 'input' parameter",
        "Use 'mode' to control processing behavior",
    }
}

func (t *ExampleTool) Synergies() map[string][]string {
    return map[string][]string{
        "precedes": {"other_tool"},
        "succeeds": {"setup_tool"},
    }
}

func (t *ExampleTool) WorkflowSnippets() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "goal": "Process text with custom mode",
            "steps": []string{
                "1. Call example_tool with input text",
                "2. Specify mode='custom' for special processing",
                "3. Use the result in subsequent operations",
            },
        },
    }
}
```

#### `pkg/mcp/middleware.go` - Middleware Support

```go
package mcp

import (
    "context"
    "time"
    
    "github.com/yourusername/my-mcp-server/pkg/models"
    "go.uber.org/zap"
)

// Logging middleware
func WithLogging(logger *zap.Logger) ToolMiddleware {
    return func(tool Tool) Tool {
        return &loggingTool{
            Tool:   tool,
            logger: logger,
        }
    }
}

type loggingTool struct {
    Tool
    logger *zap.Logger
}

func (l *loggingTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    start := time.Now()
    
    l.logger.Info("Tool execution started",
        zap.String("tool", l.Name()),
        zap.Any("args", args))
    
    result, err := l.Tool.Execute(ctx, args)
    
    duration := time.Since(start)
    
    if err != nil {
        l.logger.Error("Tool execution failed",
            zap.String("tool", l.Name()),
            zap.Error(err),
            zap.Duration("duration", duration))
    } else {
        l.logger.Info("Tool execution completed",
            zap.String("tool", l.Name()),
            zap.Duration("duration", duration))
    }
    
    return result, err
}

// Metrics middleware
type MetricsCollector interface {
    RecordToolExecution(toolName string, duration time.Duration, success bool)
}

func WithMetrics(metrics MetricsCollector) ToolMiddleware {
    return func(tool Tool) Tool {
        return &metricsTool{
            Tool:    tool,
            metrics: metrics,
        }
    }
}

type metricsTool struct {
    Tool
    metrics MetricsCollector
}

func (m *metricsTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    start := time.Now()
    
    result, err := m.Tool.Execute(ctx, args)
    
    duration := time.Since(start)
    success := err == nil && (result == nil || !result.IsError)
    
    m.metrics.RecordToolExecution(m.Name(), duration, success)
    
    return result, err
}

// Timeout middleware
func WithTimeout(timeout time.Duration) ToolMiddleware {
    return func(tool Tool) Tool {
        return &timeoutTool{
            Tool:    tool,
            timeout: timeout,
        }
    }
}

type timeoutTool struct {
    Tool
    timeout time.Duration
}

func (t *timeoutTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    ctx, cancel := context.WithTimeout(ctx, t.timeout)
    defer cancel()
    
    resultChan := make(chan *models.MCPToolResult, 1)
    errChan := make(chan error, 1)
    
    go func() {
        result, err := t.Tool.Execute(ctx, args)
        if err != nil {
            errChan <- err
            return
        }
        resultChan <- result
    }()
    
    select {
    case result := <-resultChan:
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return &models.MCPToolResult{
            IsError: true,
            Content: []models.MCPContent{{
                Type: "text",
                Text: "Tool execution timed out",
            }},
        }, nil
    }
}
```

#### `cmd/server/main.go` - Server Entry Point

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/yourusername/my-mcp-server/pkg/mcp"
    "github.com/yourusername/my-mcp-server/pkg/tools"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // Parse command line flags
    var (
        logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
        timeout  = flag.Duration("timeout", 30*time.Second, "Tool execution timeout")
    )
    flag.Parse()
    
    // Initialize logger
    logger, err := initLogger(*logLevel)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
        os.Exit(1)
    }
    defer logger.Sync()
    
    // Create context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Initialize MCP server
    server := mcp.NewServer(logger.Named("mcp"), "My MCP Server", "1.0.0")
    
    // Create and register tools with middleware
    exampleTool := tools.NewExampleTool(logger.Named("example_tool"))
    
    // Apply middleware
    var tool mcp.Tool = exampleTool
    tool = mcp.WithLogging(logger.Named("middleware"))(tool)
    tool = mcp.WithTimeout(*timeout)(tool)
    
    server.RegisterTool(tool)
    
    // Add more tools here...
    
    logger.Info("All tools registered successfully")
    
    // Set up graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        logger.Info("Received shutdown signal")
        cancel()
    }()
    
    // Start server
    logger.Info("Starting MCP server")
    if err := server.Start(ctx); err != nil {
        logger.Error("Server error", zap.Error(err))
        os.Exit(1)
    }
}

func initLogger(level string) (*zap.Logger, error) {
    config := zap.NewProductionConfig()
    
    // Set log level
    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    default:
        return nil, fmt.Errorf("invalid log level: %s", level)
    }
    
    // Customize output format
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    return config.Build()
}
```

### Usage Example

#### 1. Create a New Tool

```go
package tools

import (
    "context"
    "github.com/yourusername/my-mcp-server/pkg/models"
    "go.uber.org/zap"
)

type AnalysisTool struct {
    BaseTool
    analyzer AnalysisService
}

func NewAnalysisTool(analyzer AnalysisService, logger *zap.Logger) *AnalysisTool {
    return &AnalysisTool{
        BaseTool: NewBaseTool(
            "analyze_code",
            "Analyzes code for patterns and issues",
            logger,
        ),
        analyzer: analyzer,
    }
}

func (t *AnalysisTool) InputSchema() map[string]interface{} {
    return NewSchemaBuilder().
        AddProperty("code", "string", "Code to analyze", true).
        AddProperty("language", "string", "Programming language", true).
        AddProperty("checks", "array", "Specific checks to run", false).
        Build()
}

func (t *AnalysisTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    code, _ := args["code"].(string)
    language, _ := args["language"].(string)
    
    result, err := t.analyzer.Analyze(ctx, code, language)
    if err != nil {
        return NewResultBuilder().AddError(err).Build(), nil
    }
    
    return NewResultBuilder().
        AddText(result.Summary).
        AddText(fmt.Sprintf("Issues found: %d", len(result.Issues))).
        Build(), nil
}
```

#### 2. Register in Main

```go
// In main.go
analysisTool := tools.NewAnalysisTool(analysisService, logger)
tool = mcp.WithLogging(logger)(analysisTool)
tool = mcp.WithTimeout(60 * time.Second)(tool)
server.RegisterTool(tool)
```

#### 3. Run the Server

```bash
go run cmd/server/main.go --log-level=debug
```

### Testing

```go
package mcp_test

import (
    "context"
    "testing"
    
    "github.com/yourusername/my-mcp-server/pkg/mcp"
    "github.com/yourusername/my-mcp-server/pkg/models"
    "github.com/stretchr/testify/assert"
    "go.uber.org/zap/zaptest"
)

// Mock tool for testing
type mockTool struct {
    name string
    err  error
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Description() string { return "Mock tool" }
func (m *mockTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{"type": "object"}
}
func (m *mockTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
    if m.err != nil {
        return nil, m.err
    }
    return &models.MCPToolResult{
        Content: []models.MCPContent{{Type: "text", Text: "Success"}},
    }, nil
}

func TestToolExecution(t *testing.T) {
    logger := zaptest.NewLogger(t)
    server := mcp.NewServer(logger, "Test Server", "1.0.0")
    
    tool := &mockTool{name: "test_tool"}
    server.RegisterTool(tool)
    
    // Test successful execution
    result, err := tool.Execute(context.Background(), map[string]interface{}{})
    assert.NoError(t, err)
    assert.False(t, result.IsError)
    assert.Len(t, result.Content, 1)
    assert.Equal(t, "Success", result.Content[0].Text)
}
```

### Deployment

#### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mcp-server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-server .
CMD ["./mcp-server"]
```

#### Claude Desktop Configuration

```json
{
  "mcpServers": {
    "my-mcp-server": {
      "command": "/path/to/my-mcp-server",
      "args": ["--log-level=info"],
      "env": {
        "API_KEY": "your-api-key"
      }
    }
  }
}
```

This template provides a complete, production-ready starting point for building MCP servers in Go with all the extensibility patterns from ZetMem.