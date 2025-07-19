# Services API Reference

## LiteLLMService API

### Constructor

```go
func NewLiteLLMService(cfg config.LiteLLMConfig, logger *zap.Logger) *LiteLLMService
```

Creates a new LiteLLM service instance.

**Parameters:**
- `cfg`: LiteLLM configuration
- `logger`: Zap logger instance

### Methods

#### CallWithRetry

```go
func (s *LiteLLMService) CallWithRetry(
    ctx context.Context, 
    prompt string, 
    retryOnJSON bool
) (string, error)
```

Calls LLM with automatic retry logic and fallback models.

**Parameters:**
- `ctx`: Context for cancellation
- `prompt`: The prompt to send to the LLM
- `retryOnJSON`: Whether to retry if response is not valid JSON

**Returns:**
- `string`: LLM response
- `error`: Error if all retries fail

**Example:**
```go
response, err := llmService.CallWithRetry(
    context.Background(),
    "Extract key concepts from this code",
    true
)
```

## ChromaDBService API

### Constructor

```go
func NewChromaDBService(cfg config.ChromaDBConfig, logger *zap.Logger) *ChromaDBService
```

### Methods

#### Initialize

```go
func (c *ChromaDBService) Initialize(ctx context.Context) error
```

Initializes the ChromaDB collection. Creates if not exists.

**Returns:**
- `error`: Error if initialization fails

#### StoreMemory

```go
func (c *ChromaDBService) StoreMemory(
    ctx context.Context, 
    memory *models.Memory
) error
```

Stores a memory with its embedding and metadata.

**Parameters:**
- `ctx`: Context for cancellation
- `memory`: Memory object with embedding

**Returns:**
- `error`: Error if storage fails

**Example:**
```go
memory := &models.Memory{
    ID:          "mem_123",
    Content:     "function implementation",
    Embedding:   embedding,
    WorkspaceID: "project1",
    Tags:        []string{"golang", "http"},
}
err := chromaDB.StoreMemory(ctx, memory)
```

#### SearchSimilar

```go
func (c *ChromaDBService) SearchSimilar(
    ctx context.Context,
    queryEmbedding []float32,
    limit int,
    filters map[string]interface{}
) ([]*models.Memory, []float32, error)
```

Searches for similar memories using vector similarity.

**Parameters:**
- `ctx`: Context for cancellation
- `queryEmbedding`: Query vector
- `limit`: Maximum results to return
- `filters`: Metadata filters (e.g., workspace_id)

**Returns:**
- `[]*models.Memory`: Matching memories
- `[]float32`: Distance scores
- `error`: Error if search fails

**Example:**
```go
filters := map[string]interface{}{
    "workspace_id": "project1",
    "code_type": "golang",
}
memories, distances, err := chromaDB.SearchSimilar(
    ctx, 
    queryEmbedding, 
    10, 
    filters
)
```

## EmbeddingService API

### Constructor

```go
func NewEmbeddingService(cfg config.EmbeddingConfig, logger *zap.Logger) *EmbeddingService
```

### Methods

#### GenerateEmbedding

```go
func (s *EmbeddingService) GenerateEmbedding(
    ctx context.Context, 
    text string
) ([]float32, error)
```

Generates an embedding vector for a single text.

**Parameters:**
- `ctx`: Context for cancellation
- `text`: Text to embed

**Returns:**
- `[]float32`: Embedding vector
- `error`: Error if generation fails

#### GenerateBatchEmbeddings

```go
func (s *EmbeddingService) GenerateBatchEmbeddings(
    ctx context.Context,
    texts []string
) ([][]float32, error)
```

Generates embeddings for multiple texts efficiently.

**Parameters:**
- `ctx`: Context for cancellation
- `texts`: Array of texts to embed

**Returns:**
- `[][]float32`: Array of embedding vectors
- `error`: Error if generation fails

**Example:**
```go
texts := []string{
    "function implementation",
    "error handling code",
    "database query",
}
embeddings, err := embeddingService.GenerateBatchEmbeddings(ctx, texts)
```

## PromptManager API

### Constructor

```go
func NewPromptManager(cfg config.PromptsConfig, logger *zap.Logger) *PromptManager
```

### Methods

#### LoadPrompt

```go
func (pm *PromptManager) LoadPrompt(name string) (*PromptTemplate, error)
```

Loads a prompt template by name.

**Parameters:**
- `name`: Template name (without .yaml extension)

**Returns:**
- `*PromptTemplate`: Loaded template
- `error`: Error if loading fails

#### RenderPrompt

```go
func (pm *PromptManager) RenderPrompt(
    name string, 
    data PromptData
) (string, error)
```

Renders a prompt template with provided data.

**Parameters:**
- `name`: Template name
- `data`: Data for template rendering

**Returns:**
- `string`: Rendered prompt
- `error`: Error if rendering fails

**Example:**
```go
data := PromptData{
    Content:     codeContent,
    ProjectPath: "/path/to/project",
    CodeType:    "golang",
    Query:       "find error handling",
}
prompt, err := promptManager.RenderPrompt("code_analysis", data)
```

#### GetModelConfig

```go
func (pm *PromptManager) GetModelConfig(name string) (*ModelConfig, error)
```

Gets model configuration for a prompt.

**Returns:**
- `*ModelConfig`: Model-specific settings
- `error`: Error if prompt not found

#### ListPrompts

```go
func (pm *PromptManager) ListPrompts() ([]string, error)
```

Lists all available prompt names.

**Returns:**
- `[]string`: Array of prompt names
- `error`: Error if listing fails

## WorkspaceService API

### Constructor

```go
func NewWorkspaceService(
    chromaDB *ChromaDBService, 
    logger *zap.Logger
) *WorkspaceService
```

### Methods

#### InitializeWorkspace

```go
func (w *WorkspaceService) InitializeWorkspace(
    ctx context.Context,
    req *models.WorkspaceRequest
) (*models.Workspace, bool, error)
```

Smart workspace initialization - creates if not exists, retrieves if exists.

**Parameters:**
- `ctx`: Context for cancellation
- `req`: Workspace request with identifier

**Returns:**
- `*models.Workspace`: Workspace information
- `bool`: True if newly created
- `error`: Error if operation fails

**Example:**
```go
req := &models.WorkspaceRequest{
    Identifier:  "/path/to/project",
    Name:        "My Project",
    Description: "Project workspace",
}
workspace, created, err := workspaceService.InitializeWorkspace(ctx, req)
```

#### ValidateWorkspaceID

```go
func (w *WorkspaceService) ValidateWorkspaceID(id string) error
```

Validates a workspace identifier.

**Parameters:**
- `id`: Workspace identifier to validate

**Returns:**
- `error`: Validation error or nil

#### GetWorkspaceInfo

```go
func (w *WorkspaceService) GetWorkspaceInfo(
    ctx context.Context,
    workspaceID string
) (*models.Workspace, error)
```

Retrieves detailed workspace information.

**Parameters:**
- `ctx`: Context for cancellation
- `workspaceID`: Workspace identifier

**Returns:**
- `*models.Workspace`: Workspace details with memory count
- `error`: Error if retrieval fails

## Data Models

### Memory

```go
type Memory struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Context     string                 `json:"context"`
    Keywords    []string               `json:"keywords"`
    Tags        []string               `json:"tags"`
    WorkspaceID string                 `json:"workspace_id"`
    ProjectPath string                 `json:"project_path"` // Deprecated
    CodeType    string                 `json:"code_type"`
    Embedding   []float32              `json:"embedding"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

### Workspace

```go
type Workspace struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    MemoryCount int       `json:"memory_count"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### PromptTemplate

```go
type PromptTemplate struct {
    Name        string                 `yaml:"name"`
    Version     string                 `yaml:"version"`
    ModelConfig ModelConfig            `yaml:"model_config"`
    Template    string                 `yaml:"template"`
    Variables   map[string]interface{} `yaml:"variables"`
    Metadata    map[string]interface{} `yaml:"metadata"`
}
```

### ModelConfig

```go
type ModelConfig struct {
    Temperature float32 `yaml:"temperature"`
    MaxTokens   int     `yaml:"max_tokens"`
    TopP        float32 `yaml:"top_p,omitempty"`
    TopK        int     `yaml:"top_k,omitempty"`
}
```

## Error Handling

### Common Errors

```go
// LiteLLMService
var (
    ErrAllRetriesFailed = errors.New("all retries and fallbacks failed")
    ErrInvalidJSON      = errors.New("invalid JSON response")
    ErrNoAPIKey         = errors.New("OPENAI_API_KEY not set")
)

// ChromaDBService
var (
    ErrCollectionNotFound = errors.New("collection not found")
    ErrEmbeddingRequired  = errors.New("memory embedding is required")
    ErrStorageFailed      = errors.New("failed to store memory")
)

// EmbeddingService
var (
    ErrNoEmbeddingService = errors.New("no embedding service configured")
    ErrEmbeddingFailed    = errors.New("failed to generate embedding")
)

// PromptManager
var (
    ErrPromptNotFound     = errors.New("prompt template not found")
    ErrInvalidTemplate    = errors.New("invalid prompt template")
    ErrRenderFailed       = errors.New("failed to render template")
)

// WorkspaceService
var (
    ErrInvalidWorkspaceID = errors.New("invalid workspace identifier")
    ErrWorkspaceExists    = errors.New("workspace already exists")
)
```

## Usage Examples

### Complete Memory Creation Flow

```go
// 1. Initialize services
chromaDB := services.NewChromaDBService(chromaConfig, logger)
embedding := services.NewEmbeddingService(embeddingConfig, logger)
llm := services.NewLiteLLMService(llmConfig, logger)
prompts := services.NewPromptManager(promptConfig, logger)
workspace := services.NewWorkspaceService(chromaDB, logger)

// 2. Initialize workspace
wsReq := &models.WorkspaceRequest{
    Identifier: "/path/to/project",
}
ws, _, err := workspace.InitializeWorkspace(ctx, wsReq)

// 3. Load and render prompt
promptData := services.PromptData{
    Content:     codeContent,
    ProjectPath: ws.ID,
    CodeType:    "golang",
}
prompt, err := prompts.RenderPrompt("extract_memory", promptData)

// 4. Call LLM for analysis
analysis, err := llm.CallWithRetry(ctx, prompt, true)

// 5. Generate embedding
embed, err := embedding.GenerateEmbedding(ctx, codeContent)

// 6. Create and store memory
memory := &models.Memory{
    ID:          uuid.New().String(),
    Content:     codeContent,
    Context:     analysis,
    WorkspaceID: ws.ID,
    Embedding:   embed,
    Tags:        []string{"golang", "http"},
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
}
err = chromaDB.StoreMemory(ctx, memory)
```

### Memory Search Flow

```go
// 1. Generate query embedding
queryEmbed, err := embedding.GenerateEmbedding(ctx, "error handling")

// 2. Search with filters
filters := map[string]interface{}{
    "workspace_id": ws.ID,
    "code_type":    "golang",
}
memories, distances, err := chromaDB.SearchSimilar(
    ctx, 
    queryEmbed, 
    10, 
    filters
)

// 3. Process results
for i, mem := range memories {
    fmt.Printf("Match %d (distance: %.4f):\n", i+1, distances[i])
    fmt.Printf("  Content: %s\n", mem.Content[:100])
    fmt.Printf("  Tags: %v\n", mem.Tags)
}
```