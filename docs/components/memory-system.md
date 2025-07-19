# Memory System Architecture

## Overview

The AMEM memory system is a sophisticated vector-based memory management solution that enables intelligent storage, retrieval, and evolution of coding memories. It leverages LLM analysis, vector embeddings, and graph-based relationships to create a dynamic knowledge network.

## Core Components

### 1. Memory System (`pkg/memory/system.go`)

The central component that orchestrates all memory operations.

```go
type System struct {
    logger           *zap.Logger
    llmService       *services.LiteLLMService
    chromaDB         *services.ChromaDBService
    embeddingService *services.EmbeddingService
    workspaceService *services.WorkspaceService
}
```

**Key Responsibilities:**
- Memory creation with AI-powered analysis
- Vector-based memory retrieval
- Link generation between related memories
- Workspace-aware memory organization

### 2. Evolution Manager (`pkg/memory/evolution.go`)

Handles the intelligent evolution of memory networks over time.

```go
type EvolutionManager struct {
    system *System
    logger *zap.Logger
}
```

**Key Features:**
- Batch processing of memories for optimization
- LLM-driven pattern recognition
- Automatic link creation and strengthening
- Context and tag updates based on learned patterns

### 3. MCP Tools (`pkg/memory/tools.go`)

Provides Model Context Protocol interfaces for memory operations:

- **StoreCodingMemoryTool**: Store new memories with AI analysis
- **RetrieveRelevantMemoriesTool**: Search memories using vector similarity
- **EvolveMemoryNetworkTool**: Trigger network evolution

### 4. Workspace Tools (`pkg/memory/workspace_tools.go`)

Advanced workspace management capabilities:

- **WorkspaceInitTool**: Smart initialization (create or retrieve)
- **WorkspaceCreateTool**: Explicit workspace creation
- **WorkspaceRetrieveTool**: Fetch workspace metadata

## Data Model

### Memory Structure

```go
type Memory struct {
    ID          string                 // Unique identifier
    Content     string                 // Raw content
    Context     string                 // AI-generated context
    Keywords    []string               // Extracted keywords
    Tags        []string               // Categorical tags
    ProjectPath string                 // Deprecated
    WorkspaceID string                 // Workspace identifier
    CodeType    string                 // Programming language
    Embedding   []float32              // Vector embedding
    Links       []MemoryLink           // Relationships
    CreatedAt   time.Time              
    UpdatedAt   time.Time
    Metadata    map[string]interface{} // Extensible metadata
}
```

### Memory Links

```go
type MemoryLink struct {
    TargetID string  // Target memory ID
    LinkType string  // solution|pattern|technology|debugging|progression
    Strength float32 // Connection strength (0.0-1.0)
    Reason   string  // Human-readable explanation
}
```

## Memory Creation Workflow

1. **Content Analysis**
   - LLM analyzes the content to extract structure
   - Generates keywords, tags, and contextual summary
   - Uses sophisticated prompts for coding-specific analysis

2. **Embedding Generation**
   - Converts content to high-dimensional vector
   - Enables semantic similarity searches
   - Powers relationship discovery

3. **Memory Storage**
   - Stores in ChromaDB with vector index
   - Maintains workspace isolation
   - Preserves metadata for future evolution

4. **Link Generation**
   - Searches for similar existing memories
   - Creates typed links based on similarity
   - Builds knowledge graph connections

## Memory Retrieval Process

1. **Query Processing**
   - Converts search query to embedding
   - Applies workspace and type filters
   - Supports backward compatibility

2. **Vector Search**
   - Performs similarity search in ChromaDB
   - Uses L2 distance metric
   - Configurable relevance thresholds

3. **Result Ranking**
   - Converts distances to similarity scores
   - Filters by minimum relevance
   - Generates match reasons

## Evolution Algorithm

The evolution system continuously improves the memory network:

1. **Memory Selection**
   - Retrieves memories based on scope (recent/all/project)
   - Processes in configurable batches
   - Prioritizes high-value memories

2. **LLM Analysis**
   - Analyzes memory patterns and relationships
   - Identifies improvement opportunities
   - Suggests new connections and updates

3. **Network Updates**
   - Creates new inter-memory links
   - Updates contexts with improved descriptions
   - Refines tags for better categorization
   - Strengthens existing connections

## Workspace Management

Workspaces provide logical isolation and organization:

- **Smart Initialization**: Create new or retrieve existing
- **Path Normalization**: Handles both paths and logical names
- **Memory Isolation**: Each workspace has separate memory space
- **Backward Compatibility**: Supports legacy project_path fields

## Key Features

### 1. AI-Powered Analysis
- Automatic keyword extraction
- Contextual summarization
- Intelligent tagging

### 2. Vector Similarity
- High-dimensional embeddings
- Semantic search capabilities
- Relevance scoring

### 3. Knowledge Graph
- Typed relationships between memories
- Strength-based connections
- Reason tracking for explainability

### 4. Continuous Learning
- Network evolution over time
- Pattern recognition
- Automatic optimization

### 5. Flexible Organization
- Workspace-based isolation
- Project and code type filtering
- Extensible metadata

## Integration Points

The memory system integrates with:

1. **LiteLLM Service**: For AI analysis and evolution
2. **ChromaDB Service**: For vector storage and search
3. **Embedding Service**: For vector generation
4. **Workspace Service**: For organizational structure
5. **MCP Protocol**: For tool exposure

## Performance Considerations

- **Batch Processing**: Evolution processes memories in batches
- **Embedding Cache**: Reuses embeddings when possible
- **Async Operations**: Non-blocking memory operations
- **Configurable Limits**: Max results and relevance thresholds

## Future Enhancements

1. **Advanced Evolution**: More sophisticated learning algorithms
2. **Memory Compression**: Automatic summarization of old memories
3. **Cross-Workspace Links**: Relationships across boundaries
4. **Temporal Analysis**: Time-based memory patterns
5. **Memory Decay**: Automatic cleanup of unused memories