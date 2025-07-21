# MCP Server Implementation Template for Golang

This comprehensive guide synthesizes patterns and best practices from the ZetMem MCP server implementation, providing a complete template for building production-ready MCP servers in Go.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Core Server Implementation](#core-server-implementation)
3. [Tool System Design](#tool-system-design)
4. [Service Layer Architecture](#service-layer-architecture)
5. [Storage and Persistence](#storage-and-persistence)
6. [Error Handling Strategy](#error-handling-strategy)
7. [Configuration Management](#configuration-management)
8. [Deployment and Operations](#deployment-and-operations)

## Architecture Overview

### Key Design Principles

1. **Service-Oriented Architecture**: Clear separation between transport, business logic, and data layers
2. **Dependency Injection**: Constructor-based DI for testability and flexibility
3. **Interface-Driven Design**: Interfaces for all major components enabling easy substitution
4. **Stateless Operation**: No session state between MCP calls for scalability
5. **Observability First**: Structured logging, metrics, and tracing throughout

### Component Hierarchy

```
cmd/
├── server/
│   └── main.go          # Entry point with initialization sequence
pkg/
├── mcp/
│   ├── server.go        # Core MCP server implementation
│   ├── models.go        # Protocol data structures
│   └── errors.go        # Error definitions
├── tools/
│   ├── interfaces.go    # Tool interface definitions
│   ├── registry.go      # Tool registration system
│   └── middleware.go    # Tool middleware support
├── services/
│   ├── interfaces.go    # Service interfaces
│   └── [domain]/        # Domain-specific services
├── config/
│   └── config.go        # Configuration management
└── pkg/
    └── [utilities]/     # Shared utilities
```

## Core Server Implementation

### 1. Main Entry Point

```go
// cmd/server/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/joho/godotenv"
    "go.uber.org/zap"
    
    "myproject/pkg/config"
    "myproject/pkg/mcp"
    "myproject/pkg/services"
    "myproject/pkg/tools"
)

func main() {
    // Command-line flags
    var (
        configPath = flag.String("config", "config.yaml", "Path to configuration file")
        envFile    = flag.String("env", ".env", "Path to environment file")
        logLevel   = flag.String("log-level", "info", "Log level")
    )
    flag.Parse()

    // Load environment
    if err := godotenv.Load(*envFile); err != nil {
        // Non-fatal: environment variables might be set elsewhere
    }

    // Initialize logger
    logger, err := initLogger(*logLevel)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
        os.Exit(1)
    }
    defer logger.Sync()

    // Load configuration
    cfg, err := config.Load(*configPath)
    if err != nil {
        logger.Fatal("Failed to load configuration", zap.Error(err))
    }

    // Create context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Initialize services in dependency order
    services, err := initializeServices(ctx, cfg, logger)
    if err != nil {
        logger.Fatal("Failed to initialize services", zap.Error(err))
    }

    // Create MCP server
    server := mcp.NewServer(logger.Named("mcp"))

    // Register tools
    if err := registerTools(server, services, logger); err != nil {
        logger.Fatal("Failed to register tools", zap.Error(err))
    }

    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // Start server
    errChan := make(chan error, 1)
    go func() {
        errChan <- server.Start(ctx)
    }()

    // Wait for shutdown
    select {
    case sig := <-sigChan:
        logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
        cancel()
    case err := <-errChan:
        if err != nil {
            logger.Error("Server error", zap.Error(err))
        }
    }

    // Graceful shutdown
    logger.Info("Shutting down server")
}

func initLogger(level string) (*zap.Logger, error) {
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(parseLogLevel(level))
    config.OutputPaths = []string{"stdout"}
    config.ErrorOutputPaths = []string{"stderr"}
    return config.Build()
}

func initializeServices(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*services.Container, error) {
    container := &services.Container{}
    
    // Initialize services in dependency order
    // Example: database -> cache -> business services
    
    return container, nil
}

func registerTools(server *mcp.Server, services *services.Container, logger *zap.Logger) error {
    // Register each tool with the server
    // Example:
    // tool := tools.NewExampleTool(services.Example, logger.Named("example-tool"))
    // server.RegisterTool(tool)
    
    return nil
}
```

### 2. MCP Server Core

```go
// pkg/mcp/server.go
package mcp

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"
    
    "go.uber.org/zap"
)

type Server struct {
    logger       *zap.Logger
    tools        map[string]Tool
    initialized  bool
    mu           sync.RWMutex
    
    // I/O
    reader       *bufio.Reader
    writer       io.Writer
}

func NewServer(logger *zap.Logger) *Server {
    return &Server{
        logger: logger,
        tools:  make(map[string]Tool),
        reader: bufio.NewReader(os.Stdin),
        writer: os.Stdout,
    }
}

func (s *Server) RegisterTool(tool Tool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.tools[tool.Name()] = tool
    s.logger.Info("Registered tool", zap.String("name", tool.Name()))
}

func (s *Server) Start(ctx context.Context) error {
    s.logger.Info("Starting MCP server")
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := s.handleRequest(ctx); err != nil {
                if err == io.EOF {
                    s.logger.Info("Client disconnected")
                    return nil
                }
                s.logger.Error("Request handling error", zap.Error(err))
            }
        }
    }
}

func (s *Server) handleRequest(ctx context.Context) error {
    // Read line from stdin
    line, err := s.reader.ReadString('\n')
    if err != nil {
        return err
    }

    // Try to parse as request
    var request Request
    if err := json.Unmarshal([]byte(line), &request); err != nil {
        return s.sendError(nil, ParseError, "Invalid JSON", nil)
    }

    // Route based on method
    switch request.Method {
    case "initialize":
        return s.handleInitialize(ctx, request)
    case "tools/list":
        return s.handleToolsList(ctx, request)
    case "tools/call":
        return s.handleToolCall(ctx, request)
    default:
        if request.ID != nil {
            return s.sendError(request.ID, MethodNotFound, "Method not found", nil)
        }
        // Ignore unknown notifications
        return nil
    }
}

func (s *Server) handleInitialize(ctx context.Context, req Request) error {
    s.mu.Lock()
    s.initialized = true
    s.mu.Unlock()

    response := InitializeResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result: InitializeResult{
            ProtocolVersion: "2024-11-05",
            ServerInfo: ServerInfo{
                Name:    "mcp-golang-template",
                Version: "1.0.0",
            },
            Capabilities: ServerCapabilities{
                Tools: &ToolsCapability{},
            },
        },
    }

    return s.sendResponse(response)
}

func (s *Server) handleToolsList(ctx context.Context, req Request) error {
    s.mu.RLock()
    defer s.mu.RUnlock()

    tools := make([]ToolInfo, 0, len(s.tools))
    for _, tool := range s.tools {
        info := ToolInfo{
            Name:        tool.Name(),
            Description: tool.Description(),
            InputSchema: tool.InputSchema(),
        }
        
        // Add enhanced metadata if available
        if enhanced, ok := tool.(EnhancedTool); ok {
            info.UsageTriggers = enhanced.UsageTriggers()
            info.BestPractices = enhanced.BestPractices()
            info.Synergies = enhanced.Synergies()
            info.WorkflowSnippets = enhanced.WorkflowSnippets()
        }
        
        tools = append(tools, info)
    }

    response := ToolsListResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result:  ToolsListResult{Tools: tools},
    }

    return s.sendResponse(response)
}

func (s *Server) handleToolCall(ctx context.Context, req Request) error {
    var params ToolCallParams
    if err := json.Unmarshal(req.Params, &params); err != nil {
        return s.sendError(req.ID, InvalidParams, "Invalid parameters", nil)
    }

    s.mu.RLock()
    tool, exists := s.tools[params.Name]
    s.mu.RUnlock()

    if !exists {
        return s.sendError(req.ID, ToolNotFound, "Tool not found", nil)
    }

    // Execute tool
    result, err := tool.Execute(ctx, params.Arguments)
    if err != nil {
        s.logger.Error("Tool execution failed",
            zap.String("tool", params.Name),
            zap.Error(err))
        return s.sendError(req.ID, InternalError, "Tool execution failed", err.Error())
    }

    response := ToolCallResponse{
        JSONRPC: "2.0",
        ID:      req.ID,
        Result:  result,
    }

    return s.sendResponse(response)
}

func (s *Server) sendResponse(response interface{}) error {
    data, err := json.Marshal(response)
    if err != nil {
        return fmt.Errorf("marshal response: %w", err)
    }

    _, err = fmt.Fprintf(s.writer, "%s\n", data)
    return err
}

func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
    response := ErrorResponse{
        JSONRPC: "2.0",
        ID:      id,
        Error: Error{
            Code:    code,
            Message: message,
            Data:    data,
        },
    }

    return s.sendResponse(response)
}
```

## Tool System Design

### 1. Tool Interfaces

```go
// pkg/tools/interfaces.go
package tools

import (
    "context"
)

// Tool is the base interface for all MCP tools
type Tool interface {
    // Name returns the tool's unique identifier
    Name() string
    
    // Description returns a human-readable description
    Description() string
    
    // InputSchema returns the JSON Schema for tool arguments
    InputSchema() map[string]interface{}
    
    // Execute runs the tool with provided arguments
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// EnhancedTool provides additional metadata for better AI integration
type EnhancedTool interface {
    Tool
    
    // UsageTriggers returns scenarios when this tool should be used
    UsageTriggers() []string
    
    // BestPractices returns guidelines for effective tool usage
    BestPractices() []string
    
    // Synergies returns relationships with other tools
    Synergies() map[string][]string
    
    // WorkflowSnippets returns example usage patterns
    WorkflowSnippets() []map[string]interface{}
}

// ToolResult represents the output of a tool execution
type ToolResult struct {
    IsError bool        `json:"isError,omitempty"`
    Content []Content   `json:"content"`
}

// Content represents a piece of tool output
type Content struct {
    Type string `json:"type"`
    Text string `json:"text"`
}
```

### 2. Tool Implementation Example

```go
// pkg/tools/example/example_tool.go
package example

import (
    "context"
    "fmt"
    
    "go.uber.org/zap"
    
    "myproject/pkg/services"
    "myproject/pkg/tools"
)

type ExampleTool struct {
    service *services.ExampleService
    logger  *zap.Logger
}

func NewExampleTool(service *services.ExampleService, logger *zap.Logger) *ExampleTool {
    return &ExampleTool{
        service: service,
        logger:  logger,
    }
}

func (t *ExampleTool) Name() string {
    return "example_tool"
}

func (t *ExampleTool) Description() string {
    return "An example tool demonstrating the implementation pattern"
}

func (t *ExampleTool) InputSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type":        "string",
                "description": "The input to process",
            },
            "options": map[string]interface{}{
                "type":        "object",
                "description": "Optional processing options",
                "properties": map[string]interface{}{
                    "format": map[string]interface{}{
                        "type":        "string",
                        "description": "Output format",
                        "enum":        []string{"json", "text", "markdown"},
                        "default":     "text",
                    },
                },
            },
        },
        "required": []string{"input"},
    }
}

func (t *ExampleTool) Execute(ctx context.Context, args map[string]interface{}) (*tools.ToolResult, error) {
    // Parse arguments
    input, ok := args["input"].(string)
    if !ok {
        return &tools.ToolResult{
            IsError: true,
            Content: []tools.Content{{
                Type: "text",
                Text: "Error: 'input' parameter is required and must be a string",
            }},
        }, nil
    }

    // Parse options
    format := "text"
    if options, ok := args["options"].(map[string]interface{}); ok {
        if f, ok := options["format"].(string); ok {
            format = f
        }
    }

    // Execute business logic
    result, err := t.service.ProcessInput(ctx, input, format)
    if err != nil {
        t.logger.Error("Processing failed", 
            zap.String("input", input),
            zap.Error(err))
        
        return &tools.ToolResult{
            IsError: true,
            Content: []tools.Content{{
                Type: "text",
                Text: fmt.Sprintf("Processing failed: %v", err),
            }},
        }, nil
    }

    // Return success
    return &tools.ToolResult{
        Content: []tools.Content{{
            Type: "text",
            Text: result,
        }},
    }, nil
}

// Enhanced tool methods
func (t *ExampleTool) UsageTriggers() []string {
    return []string{
        "When the user needs to process text input",
        "When formatting conversion is required",
        "As part of a larger text processing workflow",
    }
}

func (t *ExampleTool) BestPractices() []string {
    return []string{
        "Provide clear, concise input text",
        "Specify format option for non-text output",
        "Use in combination with other text processing tools",
    }
}

func (t *ExampleTool) Synergies() map[string][]string {
    return map[string][]string{
        "precedes": {"output_formatter", "data_validator"},
        "succeeds": {"input_collector", "text_cleaner"},
    }
}

func (t *ExampleTool) WorkflowSnippets() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "goal": "Process and format user input",
            "steps": []string{
                "1. Collect input using input_collector",
                "2. Process with example_tool",
                "3. Format output with output_formatter",
            },
        },
    }
}
```

### 3. Tool Middleware

```go
// pkg/tools/middleware.go
package tools

import (
    "context"
    "time"
    
    "go.uber.org/zap"
)

// Middleware wraps tool execution with cross-cutting concerns
type Middleware func(Tool) Tool

// LoggingMiddleware adds structured logging to tool execution
func LoggingMiddleware(logger *zap.Logger) Middleware {
    return func(next Tool) Tool {
        return &loggingTool{
            Tool:   next,
            logger: logger,
        }
    }
}

type loggingTool struct {
    Tool
    logger *zap.Logger
}

func (t *loggingTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
    start := time.Now()
    
    t.logger.Info("Tool execution started",
        zap.String("tool", t.Name()),
        zap.Any("args", args))
    
    result, err := t.Tool.Execute(ctx, args)
    
    t.logger.Info("Tool execution completed",
        zap.String("tool", t.Name()),
        zap.Duration("duration", time.Since(start)),
        zap.Bool("error", result != nil && result.IsError),
        zap.Error(err))
    
    return result, err
}

// MetricsMiddleware adds performance metrics
func MetricsMiddleware(metrics MetricsCollector) Middleware {
    return func(next Tool) Tool {
        return &metricsTool{
            Tool:    next,
            metrics: metrics,
        }
    }
}

// TimeoutMiddleware adds execution timeout
func TimeoutMiddleware(timeout time.Duration) Middleware {
    return func(next Tool) Tool {
        return &timeoutTool{
            Tool:    next,
            timeout: timeout,
        }
    }
}
```

## Service Layer Architecture

### 1. Service Interfaces

```go
// pkg/services/interfaces.go
package services

import (
    "context"
)

// Container holds all service instances
type Container struct {
    Example ExampleService
    Storage StorageService
    Cache   CacheService
    // Add more services as needed
}

// ExampleService demonstrates a business logic service
type ExampleService interface {
    ProcessInput(ctx context.Context, input, format string) (string, error)
}

// StorageService provides persistence operations
type StorageService interface {
    Store(ctx context.Context, key string, value interface{}) error
    Retrieve(ctx context.Context, key string) (interface{}, error)
    Delete(ctx context.Context, key string) error
}

// CacheService provides caching operations
type CacheService interface {
    Get(ctx context.Context, key string) (interface{}, bool)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

### 2. Service Implementation Pattern

```go
// pkg/services/example/service.go
package example

import (
    "context"
    "fmt"
    
    "go.uber.org/zap"
    
    "myproject/pkg/config"
)

type Service struct {
    config *config.ExampleConfig
    logger *zap.Logger
    cache  CacheService
}

func NewService(config *config.ExampleConfig, cache CacheService, logger *zap.Logger) *Service {
    return &Service{
        config: config,
        logger: logger,
        cache:  cache,
    }
}

func (s *Service) ProcessInput(ctx context.Context, input, format string) (string, error) {
    // Check cache
    cacheKey := fmt.Sprintf("process:%s:%s", input, format)
    if cached, found := s.cache.Get(ctx, cacheKey); found {
        return cached.(string), nil
    }

    // Perform processing
    result, err := s.process(ctx, input, format)
    if err != nil {
        return "", fmt.Errorf("processing failed: %w", err)
    }

    // Cache result
    if err := s.cache.Set(ctx, cacheKey, result, s.config.CacheTTL); err != nil {
        s.logger.Warn("Failed to cache result", zap.Error(err))
    }

    return result, nil
}
```

## Storage and Persistence

### 1. Vector Storage Interface

```go
// pkg/storage/vector/interface.go
package vector

import (
    "context"
)

type VectorStore interface {
    // Store saves vectors with metadata
    Store(ctx context.Context, vectors []Vector) error
    
    // Search finds similar vectors
    Search(ctx context.Context, query Vector, limit int, filters map[string]interface{}) ([]SearchResult, error)
    
    // Update modifies metadata
    Update(ctx context.Context, id string, metadata map[string]interface{}) error
    
    // Delete removes vectors
    Delete(ctx context.Context, ids []string) error
}

type Vector struct {
    ID       string
    Values   []float32
    Metadata map[string]interface{}
}

type SearchResult struct {
    Vector   Vector
    Distance float32
    Score    float32
}
```

### 2. ChromaDB Implementation

```go
// pkg/storage/chromadb/client.go
package chromadb

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    baseURL      string
    httpClient   *http.Client
    collection   string
    collectionID string
}

func NewClient(baseURL, collection string) *Client {
    return &Client{
        baseURL:    baseURL,
        collection: collection,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
    }
}

func (c *Client) Store(ctx context.Context, vectors []vector.Vector) error {
    // Prepare batch request
    ids := make([]string, len(vectors))
    embeddings := make([][]float32, len(vectors))
    metadatas := make([]map[string]interface{}, len(vectors))
    
    for i, v := range vectors {
        ids[i] = v.ID
        embeddings[i] = v.Values
        metadatas[i] = v.Metadata
    }
    
    payload := map[string]interface{}{
        "ids":        ids,
        "embeddings": embeddings,
        "metadatas":  metadatas,
    }
    
    // Make request
    url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collectionID)
    return c.makeRequest(ctx, "POST", url, payload, nil)
}
```

## Error Handling Strategy

### 1. Error Types

```go
// pkg/errors/types.go
package errors

import "fmt"

// MCPError represents protocol-level errors
type MCPError struct {
    Code    int
    Message string
    Data    interface{}
}

func (e *MCPError) Error() string {
    return fmt.Sprintf("MCP Error %d: %s", e.Code, e.Message)
}

// Standard MCP error codes
const (
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
)

// AppError represents application-level errors
type AppError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType string

const (
    ValidationError ErrorType = "validation"
    NotFoundError   ErrorType = "not_found"
    ConflictError   ErrorType = "conflict"
    RateLimitError  ErrorType = "rate_limit"
    ExternalError   ErrorType = "external"
)

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}
```

### 2. Error Handler

```go
// pkg/errors/handler.go
package errors

import (
    "context"
    "errors"
    "runtime/debug"
    
    "go.uber.org/zap"
)

type Handler struct {
    logger *zap.Logger
}

func NewHandler(logger *zap.Logger) *Handler {
    return &Handler{logger: logger}
}

func (h *Handler) HandlePanic(ctx context.Context) (err error) {
    if r := recover(); r != nil {
        h.logger.Error("Panic recovered",
            zap.Any("panic", r),
            zap.String("stack", string(debug.Stack())))
        
        switch x := r.(type) {
        case error:
            err = fmt.Errorf("panic: %w", x)
        default:
            err = fmt.Errorf("panic: %v", x)
        }
    }
    return
}

func (h *Handler) MapToMCPError(err error) (int, string, interface{}) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        switch appErr.Type {
        case ValidationError:
            return InvalidParams, appErr.Message, nil
        case NotFoundError:
            return MethodNotFound, appErr.Message, nil
        default:
            return InternalError, "Internal server error", nil
        }
    }
    
    var mcpErr *MCPError
    if errors.As(err, &mcpErr) {
        return mcpErr.Code, mcpErr.Message, mcpErr.Data
    }
    
    return InternalError, "Internal server error", nil
}
```

## Configuration Management

### 1. Configuration Structure

```go
// pkg/config/config.go
package config

import (
    "fmt"
    "os"
    "time"
    
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server     ServerConfig     `yaml:"server"`
    Storage    StorageConfig    `yaml:"storage"`
    Services   ServicesConfig   `yaml:"services"`
    Monitoring MonitoringConfig `yaml:"monitoring"`
}

type ServerConfig struct {
    LogLevel        string        `yaml:"log_level" env:"LOG_LEVEL" default:"info"`
    MaxRequestSize  int           `yaml:"max_request_size" env:"MAX_REQUEST_SIZE" default:"10485760"`
    RequestTimeout  time.Duration `yaml:"request_timeout" env:"REQUEST_TIMEOUT" default:"30s"`
    ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" default:"10s"`
}

type StorageConfig struct {
    Type     string                 `yaml:"type" env:"STORAGE_TYPE" default:"chromadb"`
    Settings map[string]interface{} `yaml:"settings"`
}

type ServicesConfig struct {
    CacheTTL time.Duration `yaml:"cache_ttl" env:"CACHE_TTL" default:"5m"`
    // Add service-specific configs
}

type MonitoringConfig struct {
    MetricsEnabled bool   `yaml:"metrics_enabled" env:"METRICS_ENABLED" default:"true"`
    MetricsPort    int    `yaml:"metrics_port" env:"METRICS_PORT" default:"9090"`
    TracingEnabled bool   `yaml:"tracing_enabled" env:"TRACING_ENABLED" default:"false"`
}

// Load reads configuration from file and environment
func Load(path string) (*Config, error) {
    // Read file
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read config file: %w", err)
    }
    
    // Parse YAML
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    
    // Apply environment overrides
    if err := applyEnvOverrides(&config); err != nil {
        return nil, fmt.Errorf("apply env overrides: %w", err)
    }
    
    // Validate
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }
    
    return &config, nil
}

func (c *Config) Validate() error {
    // Add validation logic
    return nil
}
```

### 2. Environment Override

```go
// pkg/config/env.go
package config

import (
    "os"
    "reflect"
    "strconv"
    "time"
)

func applyEnvOverrides(config interface{}) error {
    return applyEnvToStruct(reflect.ValueOf(config).Elem(), "")
}

func applyEnvToStruct(v reflect.Value, prefix string) error {
    t := v.Type()
    
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fieldValue := v.Field(i)
        
        // Get env tag
        envName := field.Tag.Get("env")
        if envName == "" {
            continue
        }
        
        // Add prefix if present
        if prefix != "" {
            envName = prefix + "_" + envName
        }
        
        // Get environment value
        envValue := os.Getenv(envName)
        if envValue == "" {
            continue
        }
        
        // Set field value based on type
        if err := setFieldValue(fieldValue, envValue); err != nil {
            return fmt.Errorf("set field %s: %w", field.Name, err)
        }
    }
    
    return nil
}

func setFieldValue(field reflect.Value, value string) error {
    switch field.Kind() {
    case reflect.String:
        field.SetString(value)
    case reflect.Int, reflect.Int64:
        n, err := strconv.ParseInt(value, 10, 64)
        if err != nil {
            return err
        }
        field.SetInt(n)
    case reflect.Bool:
        b, err := strconv.ParseBool(value)
        if err != nil {
            return err
        }
        field.SetBool(b)
    case reflect.Float64:
        f, err := strconv.ParseFloat(value, 64)
        if err != nil {
            return err
        }
        field.SetFloat(f)
    default:
        // Handle time.Duration
        if field.Type() == reflect.TypeOf(time.Duration(0)) {
            d, err := time.ParseDuration(value)
            if err != nil {
                return err
            }
            field.Set(reflect.ValueOf(d))
        }
    }
    
    return nil
}
```

## Deployment and Operations

### 1. Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mcp-server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/mcp-server .

# Copy default config
COPY config/default.yaml ./config.yaml

# Create non-root user
RUN addgroup -g 1000 mcp && \
    adduser -D -u 1000 -G mcp mcp

USER mcp

# Expose metrics port
EXPOSE 9090

ENTRYPOINT ["./mcp-server"]
```

### 2. Docker Compose

```yaml
version: '3.8'

services:
  mcp-server:
    build: .
    environment:
      - LOG_LEVEL=info
      - STORAGE_TYPE=chromadb
      - CHROMADB_URL=http://chromadb:8000
      - METRICS_ENABLED=true
    volumes:
      - ./config:/app/config:ro
    depends_on:
      - chromadb
    stdin_open: true
    tty: true

  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
    volumes:
      - chromadb_data:/chroma/chroma
    environment:
      - IS_PERSISTENT=TRUE

volumes:
  chromadb_data:
```

### 3. Claude Desktop Configuration

```json
{
  "mcpServers": {
    "golang-template": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--name", "mcp-golang-template",
        "mcp-golang-template:latest"
      ],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### 4. Monitoring and Observability

```go
// pkg/monitoring/metrics.go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    ToolExecutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_tool_executions_total",
            Help: "Total number of tool executions",
        },
        []string{"tool", "status"},
    )
    
    ToolDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "mcp_tool_duration_seconds",
            Help:    "Tool execution duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"tool"},
    )
    
    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "mcp_active_connections",
            Help: "Number of active MCP connections",
        },
    )
)
```

## Best Practices Summary

1. **Architecture**
   - Use dependency injection for testability
   - Keep transport and business logic separate
   - Design for stateless operation

2. **Error Handling**
   - Never expose internal errors to clients
   - Use structured error types
   - Implement retry logic for transient failures

3. **Configuration**
   - Support both file and environment configuration
   - Use sensible defaults
   - Validate configuration on startup

4. **Observability**
   - Use structured logging throughout
   - Export metrics for monitoring
   - Include request IDs for tracing

5. **Testing**
   - Write unit tests for all tools
   - Use interface mocks for dependencies
   - Test error scenarios thoroughly

6. **Security**
   - Validate all inputs
   - Sanitize error messages
   - Use least privilege principles

This template provides a solid foundation for building production-ready MCP servers in Go. Adapt and extend based on your specific requirements while maintaining the core architectural principles.