# MCP Server Implementation Guide

This guide provides step-by-step instructions for implementing an MCP server in Go based on the ZetMem patterns.

## Quick Start

### 1. Project Setup

```bash
# Create project structure
mkdir mcp-server-go
cd mcp-server-go

# Initialize Go module
go mod init github.com/yourorg/mcp-server-go

# Create directory structure
mkdir -p cmd/server pkg/{mcp,tools,services,config,errors,storage} config
```

### 2. Core Dependencies

```bash
# Add essential dependencies
go get go.uber.org/zap
go get github.com/joho/godotenv
go get gopkg.in/yaml.v3
go get github.com/prometheus/client_golang
```

### 3. Basic Implementation Steps

1. **Start with the MCP Protocol**
   - Implement JSON-RPC 2.0 message parsing
   - Create request/response types
   - Build the stdio transport layer

2. **Add Tool Support**
   - Define Tool interface
   - Create tool registry
   - Implement example tools

3. **Layer in Services**
   - Design service interfaces
   - Implement core services
   - Wire dependencies

4. **Add Configuration**
   - Create config structures
   - Implement file + env loading
   - Add validation

5. **Implement Error Handling**
   - Define error types
   - Add error mapping
   - Implement recovery

## Implementation Patterns

### Pattern 1: Tool Registration

```go
// In main.go
func registerTools(server *mcp.Server, services *services.Container, logger *zap.Logger) error {
    // Create tools with dependencies
    memoryTool := memory.NewStoreTool(services.Memory, logger.Named("store-memory"))
    searchTool := memory.NewSearchTool(services.Memory, logger.Named("search-memory"))
    
    // Apply middleware
    memoryTool = tools.ApplyMiddleware(memoryTool,
        tools.LoggingMiddleware(logger),
        tools.MetricsMiddleware(metrics),
        tools.TimeoutMiddleware(30*time.Second),
    )
    
    // Register with server
    server.RegisterTool(memoryTool)
    server.RegisterTool(searchTool)
    
    return nil
}
```

### Pattern 2: Service Initialization

```go
func initializeServices(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*services.Container, error) {
    container := &services.Container{}
    
    // Initialize in dependency order
    container.Storage = storage.NewChromaDB(cfg.Storage, logger.Named("storage"))
    container.Cache = cache.NewMemoryCache(cfg.Cache, logger.Named("cache"))
    container.Memory = memory.NewService(container.Storage, container.Cache, logger.Named("memory"))
    
    // Start background services
    if err := container.Storage.Connect(ctx); err != nil {
        return nil, fmt.Errorf("connect storage: %w", err)
    }
    
    return container, nil
}
```

### Pattern 3: Graceful Shutdown

```go
func main() {
    // ... initialization ...
    
    // Shutdown handling
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    errChan := make(chan error, 1)
    go func() {
        errChan <- server.Start(ctx)
    }()
    
    select {
    case sig := <-sigChan:
        logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
        
        // Graceful shutdown with timeout
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer shutdownCancel()
        
        if err := shutdown(shutdownCtx, services); err != nil {
            logger.Error("Shutdown error", zap.Error(err))
        }
        
    case err := <-errChan:
        if err != nil {
            logger.Fatal("Server error", zap.Error(err))
        }
    }
}

func shutdown(ctx context.Context, services *services.Container) error {
    // Close services in reverse order
    if err := services.Memory.Close(ctx); err != nil {
        return fmt.Errorf("close memory service: %w", err)
    }
    
    if err := services.Storage.Close(ctx); err != nil {
        return fmt.Errorf("close storage: %w", err)
    }
    
    return nil
}
```

## Testing Strategy

### 1. Unit Tests

```go
// pkg/tools/example/example_tool_test.go
func TestExampleTool_Execute(t *testing.T) {
    tests := []struct {
        name    string
        args    map[string]interface{}
        want    *tools.ToolResult
        wantErr bool
    }{
        {
            name: "valid input",
            args: map[string]interface{}{
                "input": "test",
            },
            want: &tools.ToolResult{
                Content: []tools.Content{{
                    Type: "text",
                    Text: "Processed: test",
                }},
            },
        },
        {
            name: "missing required field",
            args: map[string]interface{}{},
            want: &tools.ToolResult{
                IsError: true,
                Content: []tools.Content{{
                    Type: "text",
                    Text: "Error: 'input' parameter is required and must be a string",
                }},
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create mocks
            mockService := &MockExampleService{}
            logger := zap.NewNop()
            
            // Create tool
            tool := NewExampleTool(mockService, logger)
            
            // Execute
            got, err := tool.Execute(context.Background(), tt.args)
            
            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Execute() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 2. Integration Tests

```go
// cmd/server/integration_test.go
func TestServerIntegration(t *testing.T) {
    // Start server in test mode
    server, cleanup := startTestServer(t)
    defer cleanup()
    
    // Create client
    client := NewTestClient(server)
    
    // Test initialize
    initResp, err := client.Initialize()
    require.NoError(t, err)
    assert.Equal(t, "2024-11-05", initResp.ProtocolVersion)
    
    // Test tools/list
    toolsResp, err := client.ListTools()
    require.NoError(t, err)
    assert.NotEmpty(t, toolsResp.Tools)
    
    // Test tool execution
    result, err := client.CallTool("example_tool", map[string]interface{}{
        "input": "test",
    })
    require.NoError(t, err)
    assert.False(t, result.IsError)
}
```

## Performance Optimization

### 1. Connection Pooling

```go
// For HTTP-based storage backends
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    DisableCompression:  true,
    DisableKeepAlives:   false,
}

client := &http.Client{
    Timeout:   30 * time.Second,
    Transport: transport,
}
```

### 2. Caching Strategy

```go
// Implement multi-level caching
type CacheLayer struct {
    l1 *MemoryCache  // Fast in-memory cache
    l2 *RedisCache   // Distributed cache
}

func (c *CacheLayer) Get(ctx context.Context, key string) (interface{}, bool) {
    // Check L1 first
    if val, found := c.l1.Get(ctx, key); found {
        return val, true
    }
    
    // Check L2
    if val, found := c.l2.Get(ctx, key); found {
        // Populate L1
        c.l1.Set(ctx, key, val, 5*time.Minute)
        return val, true
    }
    
    return nil, false
}
```

### 3. Batch Processing

```go
// For memory evolution or bulk operations
type BatchProcessor struct {
    batchSize int
    timeout   time.Duration
    process   func([]interface{}) error
}

func (b *BatchProcessor) Process(ctx context.Context, items []interface{}) error {
    for i := 0; i < len(items); i += b.batchSize {
        end := i + b.batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        
        if err := b.process(batch); err != nil {
            return fmt.Errorf("process batch %d: %w", i/b.batchSize, err)
        }
    }
    
    return nil
}
```

## Security Considerations

### 1. Input Validation

```go
func validateInput(input string) error {
    // Check for null bytes
    if strings.Contains(input, "\x00") {
        return errors.New("null bytes not allowed")
    }
    
    // Check length
    if len(input) > MaxInputLength {
        return errors.New("input too long")
    }
    
    // Additional validation...
    return nil
}
```

### 2. Rate Limiting

```go
// Implement per-tool rate limiting
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
}

func (r *RateLimiter) Allow(tool string) bool {
    r.mu.RLock()
    limiter, exists := r.limiters[tool]
    r.mu.RUnlock()
    
    if !exists {
        r.mu.Lock()
        limiter = rate.NewLimiter(rate.Limit(10), 100) // 10 req/s, burst 100
        r.limiters[tool] = limiter
        r.mu.Unlock()
    }
    
    return limiter.Allow()
}
```

## Monitoring and Debugging

### 1. Structured Logging

```go
// Use consistent log fields
logger.Info("Tool execution",
    zap.String("tool", toolName),
    zap.String("request_id", requestID),
    zap.Duration("duration", duration),
    zap.Bool("success", !result.IsError),
    zap.Any("args", sanitizeArgs(args)),
)
```

### 2. Health Checks

```go
// Implement health endpoint for monitoring
func (s *Server) HealthCheck(ctx context.Context) error {
    // Check critical dependencies
    checks := []struct {
        name  string
        check func(context.Context) error
    }{
        {"storage", s.services.Storage.Ping},
        {"cache", s.services.Cache.Ping},
    }
    
    for _, c := range checks {
        if err := c.check(ctx); err != nil {
            return fmt.Errorf("%s unhealthy: %w", c.name, err)
        }
    }
    
    return nil
}
```

## Common Pitfalls and Solutions

### 1. Stdin/Stdout Handling

**Problem**: Buffering issues with stdio communication

**Solution**:
```go
// Always use buffered reader
reader := bufio.NewReader(os.Stdin)

// Flush after writing
fmt.Fprintf(os.Stdout, "%s\n", response)
os.Stdout.Sync() // Force flush
```

### 2. Context Cancellation

**Problem**: Goroutine leaks from improper context handling

**Solution**:
```go
// Always respect context cancellation
select {
case <-ctx.Done():
    return ctx.Err()
case result := <-resultChan:
    return result, nil
}
```

### 3. Error Message Exposure

**Problem**: Leaking internal details in error messages

**Solution**:
```go
// Map internal errors to user-safe messages
func sanitizeError(err error) string {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Message // Already sanitized
    }
    
    // Don't expose internal errors
    logger.Error("Internal error", zap.Error(err))
    return "An error occurred. Please try again."
}
```

## Next Steps

1. **Extend Tool Set**: Add domain-specific tools
2. **Add Storage Backends**: Support multiple vector databases
3. **Implement Authentication**: Add API key or token support
4. **Add Telemetry**: OpenTelemetry integration
5. **Build CLI**: Create management CLI for operations

This implementation guide provides the foundation for building a robust, production-ready MCP server in Go following the patterns established by ZetMem.