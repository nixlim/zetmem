# Key Architectural Decisions in ZetMem MCP Server

This document captures the important architectural decisions made in the ZetMem implementation that should be considered when creating a Go MCP server template.

## 1. Transport Layer: stdio vs HTTP/WebSocket

### Decision: stdio (stdin/stdout)
**Rationale:**
- MCP protocol is designed for stdio communication
- Simplifies deployment (no network configuration)
- Natural fit for CLI integration
- Secure by default (no network exposure)

**Trade-offs:**
- Single client connection at a time
- No multiplexing capability
- Requires process management for scaling

**Alternative Considered:**
- HTTP/WebSocket for web-based clients
- Could be added as additional transport layer

## 2. State Management: Stateless Design

### Decision: Completely Stateless Server
**Rationale:**
- Each tool call is independent
- Simplifies scaling and reliability
- No session management complexity
- State persisted in external storage (ChromaDB)

**Implementation:**
- Workspace context passed in each request
- No in-memory session state
- Tools are pure functions with injected dependencies

**Benefits:**
- Horizontal scaling possible
- Crash recovery is simple
- Testing is straightforward

## 3. Tool Interface Design: Dual Interface Pattern

### Decision: Base Tool + Enhanced Tool Interfaces
**Rationale:**
- Base interface for minimal MCP compliance
- Enhanced interface for AI-optimized metadata
- Backward compatibility maintained
- Progressive enhancement approach

**Implementation:**
```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

type EnhancedTool interface {
    Tool
    UsageTriggers() []string
    BestPractices() []string
    Synergies() map[string][]string
    WorkflowSnippets() []map[string]interface{}
}
```

**Benefits:**
- Tools can be simple or sophisticated
- AI assistants get rich metadata when available
- Clean separation of concerns

## 4. Error Handling: Never Return Go Errors

### Decision: Convert All Errors to MCP Protocol Responses
**Rationale:**
- MCP protocol expects specific error format
- Go errors would break protocol compliance
- User-friendly error messages required

**Implementation:**
- Tools return `(*ToolResult, error)` but error is always nil
- Actual errors converted to `ToolResult{IsError: true}`
- Internal errors logged but not exposed

**Benefits:**
- Protocol compliance guaranteed
- Security through error sanitization
- Consistent error handling

## 5. Configuration: Three-Layer System

### Decision: Defaults → YAML → Environment Variables
**Rationale:**
- Flexibility for different deployment scenarios
- Environment variables for secrets
- YAML for complex configuration
- Defaults for zero-config startup

**Priority Order:**
1. Environment variables (highest)
2. YAML configuration file
3. Default values in code (lowest)

**Benefits:**
- Works in containers and traditional deployments
- Secrets never in configuration files
- Easy local development

## 6. Service Architecture: Dependency Injection

### Decision: Constructor-Based DI
**Rationale:**
- Explicit dependencies
- Testability through mocking
- No global state
- Clear initialization order

**Implementation:**
```go
func NewMemoryService(storage StorageService, embedder EmbeddingService, llm LLMService, logger *zap.Logger) *MemoryService {
    return &MemoryService{
        storage:  storage,
        embedder: embedder,
        llm:      llm,
        logger:   logger,
    }
}
```

**Benefits:**
- Easy to test with mocks
- Dependencies are explicit
- Circular dependencies impossible

## 7. Storage: Vector Database First

### Decision: ChromaDB as Primary Storage
**Rationale:**
- Semantic search is primary use case
- Vector embeddings central to functionality
- Metadata storage included
- REST API allows language flexibility

**Trade-offs:**
- Additional infrastructure requirement
- Network latency for operations
- Limited transaction support

**Abstraction:**
- Interface-based design allows swapping
- Could support Qdrant, Weaviate, Pinecone

## 8. Logging: Structured Logging Throughout

### Decision: Zap Logger with Named Instances
**Rationale:**
- High performance
- Structured fields for analysis
- Named loggers for component isolation
- JSON output for log aggregation

**Pattern:**
```go
logger.Info("Operation completed",
    zap.String("component", "memory"),
    zap.String("operation", "store"),
    zap.Duration("duration", time.Since(start)),
    zap.Error(err))
```

**Benefits:**
- Easy debugging with component filtering
- Machine-readable logs
- Performance metrics included

## 9. Workspace Model: Unified Identity System

### Decision: Support Both Paths and Logical Names
**Rationale:**
- Filesystem paths natural for developers
- Logical names better for SaaS/multi-tenant
- Smooth migration path
- User flexibility

**Implementation:**
- Normalize all identifiers
- Check both path and name in queries
- Backward compatibility maintained

## 10. Memory Evolution: Scheduled Background Processing

### Decision: Batch Processing with LLM Analysis
**Rationale:**
- LLM calls expensive
- Batch processing more efficient
- Asynchronous improvement
- System gets smarter over time

**Implementation:**
- Configurable batch sizes
- Rate limiting for LLM calls
- Gradual network evolution
- Non-blocking operation

## 11. Tool Discovery: Rich Metadata

### Decision: Extended Tool Information
**Rationale:**
- AI assistants need context
- Usage patterns improve UX
- Workflow examples guide users
- Synergies enable tool chaining

**Metadata Includes:**
- Usage triggers
- Best practices
- Tool relationships
- Workflow examples

## 12. Prompt Management: File-Based with Hot Reload

### Decision: YAML Files with Template Support
**Rationale:**
- Easy prompt iteration
- Version control friendly
- No rebuild required
- Template variables for flexibility

**Features:**
- Hot reload in development
- Caching in production
- Go template syntax
- Model-specific configuration

## Architectural Principles Summary

1. **Simplicity First**: Avoid over-engineering
2. **Protocol Compliance**: Strict adherence to MCP spec
3. **Testability**: Every component independently testable
4. **Observability**: Metrics and logging built-in
5. **Flexibility**: Interfaces enable substitution
6. **Security**: Input validation and error sanitization
7. **Performance**: Caching and batch processing where appropriate
8. **Developer Experience**: Clear errors and good defaults

These architectural decisions create a solid foundation for building MCP servers that are:
- Easy to understand and maintain
- Scalable and reliable
- Secure by default
- Pleasant to work with

When implementing your own MCP server, consider these decisions but adapt based on your specific requirements and constraints.