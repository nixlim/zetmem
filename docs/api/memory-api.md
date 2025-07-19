# Memory System API Reference

## MCP Tools

The memory system exposes its functionality through Model Context Protocol (MCP) tools. These tools can be invoked by LLM assistants to manage coding memories.

### 1. store_coding_memory

Store a coding memory with AI-generated analysis, keywords, tags, and embeddings.

**Parameters:**
- `content` (string, required): The code content or coding context to store
- `workspace_id` (string, optional): Workspace identifier for organizing memories
- `project_path` (string, optional, deprecated): Use workspace_id instead
- `code_type` (string, optional): Programming language (e.g., 'javascript', 'python')
- `context` (string, optional): Additional context about the code

**Example Request:**
```json
{
  "content": "def fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)",
  "workspace_id": "algorithms-practice",
  "code_type": "python",
  "context": "Recursive implementation of Fibonacci sequence"
}
```

**Example Response:**
```json
{
  "memory_id": "550e8400-e29b-41d4-a716-446655440000",
  "keywords": ["fibonacci", "recursion", "algorithm", "sequence"],
  "tags": ["python", "algorithms", "recursion", "mathematics"],
  "links_created": 3,
  "event_emitted": true
}
```

### 2. retrieve_relevant_memories

Retrieve relevant coding memories based on a query using vector similarity search.

**Parameters:**
- `query` (string, required): Search query (code snippet, problem description, or keywords)
- `workspace_id` (string, optional): Filter by workspace
- `max_results` (integer, optional): Maximum results to return (default: 5)
- `project_filter` (string, optional, deprecated): Use workspace_id
- `code_types` (array[string], optional): Filter by programming languages
- `min_relevance` (number, optional): Minimum relevance score 0.0-1.0 (default: 0.7)

**Example Request:**
```json
{
  "query": "implement dynamic programming fibonacci",
  "workspace_id": "algorithms-practice",
  "max_results": 10,
  "code_types": ["python", "javascript"],
  "min_relevance": 0.6
}
```

**Example Response:**
```json
{
  "memories": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "content": "def fibonacci_dp(n):\n    dp = [0, 1]\n    for i in range(2, n+1):\n        dp.append(dp[i-1] + dp[i-2])\n    return dp[n]",
      "context": "Dynamic programming solution for Fibonacci",
      "keywords": ["fibonacci", "dynamic-programming", "optimization"],
      "tags": ["python", "algorithms", "dp"],
      "relevance_score": 0.92,
      "match_reason": "Keyword match: dynamic programming"
    }
  ],
  "total_found": 1
}
```

### 3. evolve_memory_network

Trigger evolution of the memory network to identify patterns and optimize connections.

**Parameters:**
- `trigger_type` (string, optional): Type of trigger - 'manual', 'scheduled', 'event' (default: 'manual')
- `scope` (string, optional): Evolution scope - 'recent', 'all', 'project' (default: 'recent')
- `max_memories` (integer, optional): Maximum memories to analyze (default: 100)
- `project_path` (string, optional): Required when scope is 'project'

**Example Request:**
```json
{
  "trigger_type": "manual",
  "scope": "recent",
  "max_memories": 50
}
```

**Example Response:**
```json
{
  "memories_analyzed": 45,
  "memories_evolved": 12,
  "links_created": 8,
  "links_strengthened": 15,
  "contexts_updated": 7,
  "duration_ms": 2341
}
```

### 4. workspace_init

Smart workspace initialization - creates new workspace or retrieves existing one.

**Parameters:**
- `identifier` (string, optional): Path or name for workspace. Uses current directory if not provided
- `name` (string, optional): Human-readable workspace name

**Example Request:**
```json
{
  "identifier": "machine-learning-projects",
  "name": "ML Projects Workspace"
}
```

**Example Response:**
```json
{
  "workspace": {
    "id": "machine-learning-projects",
    "name": "ML Projects Workspace",
    "description": "",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "memory_count": 0
  },
  "created": true
}
```

### 5. workspace_create

Explicit workspace creation - fails if workspace already exists.

**Parameters:**
- `identifier` (string, required): Path or name for the workspace
- `name` (string, optional): Human-readable name
- `description` (string, optional): Workspace description

**Example Request:**
```json
{
  "identifier": "web-dev-2024",
  "name": "Web Development 2024",
  "description": "Full-stack web development projects and snippets"
}
```

### 6. workspace_retrieve

Explicit workspace retrieval - fails if workspace doesn't exist.

**Parameters:**
- `identifier` (string, required): Path or name of workspace to retrieve

**Example Request:**
```json
{
  "identifier": "web-dev-2024"
}
```

## Internal API Methods

### Memory System

#### CreateMemory(ctx, request) -> (response, error)

Creates a new memory with AI analysis and vector embedding.

**Process:**
1. Validates and normalizes workspace ID
2. Generates unique memory ID
3. Constructs note using LLM analysis
4. Generates vector embedding
5. Creates memory links to similar memories
6. Stores in ChromaDB

#### RetrieveMemories(ctx, request) -> (response, error)

Retrieves memories using vector similarity search.

**Process:**
1. Normalizes workspace filter
2. Generates query embedding
3. Builds ChromaDB filter structure
4. Performs vector similarity search
5. Ranks and filters results
6. Generates match reasons

### Evolution Manager

#### EvolveNetwork(ctx, request) -> (response, error)

Evolves the memory network through AI analysis.

**Process:**
1. Retrieves memories based on scope
2. Processes in batches (default: 10)
3. Analyzes with LLM for patterns
4. Applies evolution actions:
   - Updates contexts
   - Updates tags
   - Creates new links
   - Strengthens existing links

## Data Structures

### Memory Fields

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Unique identifier (UUID) |
| Content | string | Raw code or context |
| Context | string | AI-generated summary |
| Keywords | []string | Extracted keywords |
| Tags | []string | Categorical tags |
| WorkspaceID | string | Workspace identifier |
| CodeType | string | Programming language |
| Embedding | []float32 | Vector representation |
| Links | []MemoryLink | Relationships |
| CreatedAt | time.Time | Creation timestamp |
| UpdatedAt | time.Time | Last update timestamp |
| Metadata | map[string]interface{} | Extensible data |

### Link Types

| Type | Description |
|------|-------------|
| solution | Related solutions to similar problems |
| pattern | Similar design patterns |
| technology | Same technology stack |
| debugging | Related debugging scenarios |
| progression | Learning progression |

### Evolution Actions

| Action | Description |
|--------|-------------|
| Context Update | Improve memory descriptions |
| Tag Update | Refine categorization |
| Link Creation | Add new relationships |
| Link Strengthening | Increase connection weights |

## Error Handling

All API methods return structured errors:

```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}
```

Common error codes:
- `INVALID_WORKSPACE`: Invalid workspace identifier
- `MEMORY_NOT_FOUND`: Memory ID doesn't exist
- `EMBEDDING_FAILED`: Vector generation failed
- `LLM_ERROR`: AI analysis failed
- `STORAGE_ERROR`: Database operation failed

## Best Practices

1. **Workspace Organization**
   - Use meaningful workspace identifiers
   - Group related memories together
   - Consider project structure

2. **Memory Content**
   - Include complete code snippets
   - Add contextual information
   - Specify correct code types

3. **Search Queries**
   - Use descriptive search terms
   - Include problem context
   - Adjust relevance thresholds

4. **Evolution Triggers**
   - Run evolution periodically
   - Target specific scopes
   - Monitor evolution metrics

5. **Performance**
   - Limit max_results for searches
   - Use appropriate batch sizes
   - Cache frequently accessed memories