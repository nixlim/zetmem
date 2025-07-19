# Service Integration Guide

## Overview

This guide explains how to integrate and extend the A-MEM service layer for custom implementations and advanced use cases.

## Service Initialization Order

Services must be initialized in the correct order due to dependencies:

```go
// 1. Initialize ChromaDB first (storage backend)
chromaDB := services.NewChromaDBService(chromaConfig, logger)
if err := chromaDB.Initialize(ctx); err != nil {
    log.Fatal("Failed to initialize ChromaDB:", err)
}

// 2. Initialize Embedding Service
embeddingService := services.NewEmbeddingService(embeddingConfig, logger)

// 3. Initialize LiteLLM Service
llmService := services.NewLiteLLMService(llmConfig, logger)

// 4. Initialize Prompt Manager
promptManager := services.NewPromptManager(promptConfig, logger)

// 5. Initialize Workspace Service (depends on ChromaDB)
workspaceService := services.NewWorkspaceService(chromaDB, logger)
```

## Configuration Patterns

### Environment-Based Configuration

```go
// config/config.go
func LoadConfig() (*Config, error) {
    cfg := &Config{
        ChromaDB: ChromaDBConfig{
            URL:        getEnv("CHROMADB_URL", "http://localhost:8000"),
            Collection: getEnv("CHROMADB_COLLECTION", "amem_memories"),
        },
        Embedding: EmbeddingConfig{
            Service: getEnv("EMBEDDING_SERVICE", "sentence-transformers"),
            URL:     getEnv("EMBEDDING_URL", "http://localhost:8080"),
            Model:   getEnv("EMBEDDING_MODEL", "all-MiniLM-L6-v2"),
        },
        LiteLLM: LiteLLMConfig{
            DefaultModel:   getEnv("LLM_MODEL", "gpt-3.5-turbo"),
            MaxRetries:     getEnvInt("LLM_MAX_RETRIES", 3),
            Timeout:        getEnvDuration("LLM_TIMEOUT", 30*time.Second),
            FallbackModels: getEnvList("LLM_FALLBACK_MODELS", ","),
        },
    }
    return cfg, nil
}
```

### YAML Configuration

```yaml
# config.yaml
services:
  chromadb:
    url: "http://localhost:8000"
    collection: "amem_memories"
  
  embedding:
    service: "sentence-transformers"
    url: "http://localhost:8080"
    model: "all-MiniLM-L6-v2"
  
  litellm:
    default_model: "gpt-3.5-turbo"
    max_retries: 3
    timeout: 30s
    fallback_models:
      - "gpt-3.5-turbo-16k"
      - "claude-instant-1"
```

## Common Integration Patterns

### 1. Memory Processing Pipeline

```go
type MemoryPipeline struct {
    workspace    *services.WorkspaceService
    llm          *services.LiteLLMService
    embedding    *services.EmbeddingService
    chromaDB     *services.ChromaDBService
    prompts      *services.PromptManager
}

func (p *MemoryPipeline) ProcessCode(ctx context.Context, code CodeInput) error {
    // 1. Ensure workspace exists
    ws, _, err := p.workspace.InitializeWorkspace(ctx, &models.WorkspaceRequest{
        Identifier: code.ProjectPath,
    })
    if err != nil {
        return fmt.Errorf("workspace init failed: %w", err)
    }
    
    // 2. Extract memory using LLM
    promptData := services.PromptData{
        Content:     code.Content,
        ProjectPath: ws.ID,
        CodeType:    code.Language,
    }
    
    prompt, err := p.prompts.RenderPrompt("extract_memory", promptData)
    if err != nil {
        return fmt.Errorf("prompt render failed: %w", err)
    }
    
    analysis, err := p.llm.CallWithRetry(ctx, prompt, true)
    if err != nil {
        return fmt.Errorf("LLM analysis failed: %w", err)
    }
    
    // 3. Generate embedding
    embedding, err := p.embedding.GenerateEmbedding(ctx, code.Content)
    if err != nil {
        return fmt.Errorf("embedding generation failed: %w", err)
    }
    
    // 4. Store memory
    memory := &models.Memory{
        ID:          generateID(),
        Content:     code.Content,
        Context:     analysis,
        WorkspaceID: ws.ID,
        Embedding:   embedding,
        Tags:        extractTags(analysis),
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    return p.chromaDB.StoreMemory(ctx, memory)
}
```

### 2. Semantic Search Handler

```go
type SearchHandler struct {
    embedding *services.EmbeddingService
    chromaDB  *services.ChromaDBService
    workspace *services.WorkspaceService
}

func (h *SearchHandler) Search(ctx context.Context, query SearchQuery) (*SearchResults, error) {
    // 1. Generate query embedding
    queryEmbedding, err := h.embedding.GenerateEmbedding(ctx, query.Text)
    if err != nil {
        return nil, fmt.Errorf("query embedding failed: %w", err)
    }
    
    // 2. Build filters
    filters := make(map[string]interface{})
    if query.WorkspaceID != "" {
        filters["workspace_id"] = query.WorkspaceID
    }
    if query.CodeType != "" {
        filters["code_type"] = query.CodeType
    }
    for k, v := range query.CustomFilters {
        filters[k] = v
    }
    
    // 3. Search memories
    memories, distances, err := h.chromaDB.SearchSimilar(
        ctx,
        queryEmbedding,
        query.Limit,
        filters,
    )
    if err != nil {
        return nil, fmt.Errorf("search failed: %w", err)
    }
    
    // 4. Build results
    results := &SearchResults{
        Query:   query.Text,
        Results: make([]SearchResult, len(memories)),
    }
    
    for i, mem := range memories {
        results.Results[i] = SearchResult{
            Memory:     mem,
            Similarity: 1.0 - distances[i], // Convert distance to similarity
            Score:      calculateRelevanceScore(mem, query, distances[i]),
        }
    }
    
    return results, nil
}
```

### 3. Batch Processing

```go
type BatchProcessor struct {
    pipeline  *MemoryPipeline
    embedding *services.EmbeddingService
}

func (b *BatchProcessor) ProcessBatch(ctx context.Context, files []FileInput) error {
    // 1. Extract content from all files
    contents := make([]string, len(files))
    for i, file := range files {
        contents[i] = file.Content
    }
    
    // 2. Generate batch embeddings
    embeddings, err := b.embedding.GenerateBatchEmbeddings(ctx, contents)
    if err != nil {
        return fmt.Errorf("batch embedding failed: %w", err)
    }
    
    // 3. Process each file with its embedding
    var wg sync.WaitGroup
    errors := make(chan error, len(files))
    
    for i, file := range files {
        wg.Add(1)
        go func(idx int, f FileInput) {
            defer wg.Done()
            
            // Process with pre-computed embedding
            if err := b.processFileWithEmbedding(ctx, f, embeddings[idx]); err != nil {
                errors <- fmt.Errorf("file %s: %w", f.Path, err)
            }
        }(i, file)
    }
    
    wg.Wait()
    close(errors)
    
    // Collect errors
    var errs []error
    for err := range errors {
        errs = append(errs, err)
    }
    
    if len(errs) > 0 {
        return fmt.Errorf("batch processing failed: %v", errs)
    }
    
    return nil
}
```

## Extending Services

### Custom LLM Provider

```go
type CustomLLMProvider struct {
    baseURL string
    apiKey  string
    client  *http.Client
}

func (p *CustomLLMProvider) Call(ctx context.Context, prompt string) (string, error) {
    // Implement custom LLM API call
    request := map[string]interface{}{
        "prompt":      prompt,
        "temperature": 0.1,
        "max_tokens":  1000,
    }
    
    // Make API call
    resp, err := p.makeAPICall(ctx, request)
    if err != nil {
        return "", err
    }
    
    return resp.Text, nil
}

// Integrate with LiteLLMService
func ExtendLiteLLMService(service *services.LiteLLMService, provider *CustomLLMProvider) {
    // Add custom provider to fallback chain
    service.AddProvider("custom", provider)
}
```

### Custom Embedding Model

```go
type LocalEmbeddingModel struct {
    model    *transformers.Model
    tokenizer *transformers.Tokenizer
}

func (m *LocalEmbeddingModel) Embed(text string) ([]float32, error) {
    // Tokenize input
    tokens := m.tokenizer.Encode(text)
    
    // Run model inference
    outputs := m.model.Forward(tokens)
    
    // Extract embeddings
    embeddings := outputs.LastHiddenState.Mean(1) // Mean pooling
    
    return embeddings.ToFloat32(), nil
}

// Create custom embedding service
type CustomEmbeddingService struct {
    model *LocalEmbeddingModel
}

func (s *CustomEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
    return s.model.Embed(text)
}
```

### Custom Storage Backend

```go
type PostgresVectorStore struct {
    db *sql.DB
}

func (p *PostgresVectorStore) Store(ctx context.Context, memory *models.Memory) error {
    query := `
        INSERT INTO memories (id, content, embedding, metadata, workspace_id)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    metadata, _ := json.Marshal(memory.Metadata)
    _, err := p.db.ExecContext(
        ctx,
        query,
        memory.ID,
        memory.Content,
        pq.Array(memory.Embedding),
        metadata,
        memory.WorkspaceID,
    )
    
    return err
}

func (p *PostgresVectorStore) SearchSimilar(
    ctx context.Context,
    embedding []float32,
    limit int,
) ([]*models.Memory, error) {
    query := `
        SELECT id, content, metadata, 
               1 - (embedding <=> $1) as similarity
        FROM memories
        WHERE workspace_id = $2
        ORDER BY embedding <=> $1
        LIMIT $3
    `
    
    // Execute query and scan results
    // ...
}
```

## Testing Strategies

### Unit Testing Services

```go
func TestLiteLLMService_CallWithRetry(t *testing.T) {
    // Mock HTTP client
    mockClient := &MockHTTPClient{
        Responses: []MockResponse{
            {StatusCode: 500, Error: errors.New("server error")},
            {StatusCode: 200, Body: `{"choices":[{"message":{"content":"test"}}]}`},
        },
    }
    
    service := &services.LiteLLMService{
        httpClient: mockClient,
        config: config.LiteLLMConfig{
            MaxRetries: 3,
        },
        logger: zap.NewNop(),
    }
    
    result, err := service.CallWithRetry(context.Background(), "test prompt", false)
    assert.NoError(t, err)
    assert.Equal(t, "test", result)
    assert.Equal(t, 2, mockClient.CallCount) // Failed once, succeeded on retry
}
```

### Integration Testing

```go
func TestMemoryPipeline_Integration(t *testing.T) {
    // Use test containers for ChromaDB
    chromaContainer, err := testcontainers.GenericContainer(
        context.Background(),
        testcontainers.GenericContainerRequest{
            ContainerRequest: testcontainers.ContainerRequest{
                Image: "chromadb/chroma:latest",
                ExposedPorts: []string{"8000/tcp"},
            },
        },
    )
    require.NoError(t, err)
    defer chromaContainer.Terminate(context.Background())
    
    // Get ChromaDB URL
    host, err := chromaContainer.Host(context.Background())
    port, err := chromaContainer.MappedPort(context.Background(), "8000")
    chromaURL := fmt.Sprintf("http://%s:%s", host, port.Port())
    
    // Initialize services with test config
    // ... test pipeline
}
```

## Performance Optimization

### Connection Pooling

```go
// HTTP client with connection pooling
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
    DisableCompression:  true,
}

httpClient := &http.Client{
    Transport: transport,
    Timeout:   30 * time.Second,
}
```

### Caching Layer

```go
type CachedEmbeddingService struct {
    service services.EmbeddingService
    cache   *lru.Cache
}

func (c *CachedEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
    // Check cache
    if cached, ok := c.cache.Get(text); ok {
        return cached.([]float32), nil
    }
    
    // Generate embedding
    embedding, err := c.service.GenerateEmbedding(ctx, text)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    c.cache.Add(text, embedding)
    
    return embedding, nil
}
```

## Monitoring and Observability

### Metrics Collection

```go
type MetricsMiddleware struct {
    service      services.LiteLLMService
    callDuration prometheus.Histogram
    errorCount   prometheus.Counter
}

func (m *MetricsMiddleware) CallWithRetry(ctx context.Context, prompt string, retryOnJSON bool) (string, error) {
    start := time.Now()
    
    result, err := m.service.CallWithRetry(ctx, prompt, retryOnJSON)
    
    m.callDuration.Observe(time.Since(start).Seconds())
    if err != nil {
        m.errorCount.Inc()
    }
    
    return result, err
}
```

### Distributed Tracing

```go
func (s *TracedChromaDBService) StoreMemory(ctx context.Context, memory *models.Memory) error {
    span, ctx := opentracing.StartSpanFromContext(ctx, "ChromaDB.StoreMemory")
    defer span.Finish()
    
    span.SetTag("memory.id", memory.ID)
    span.SetTag("workspace.id", memory.WorkspaceID)
    span.SetTag("embedding.size", len(memory.Embedding))
    
    err := s.chromaDB.StoreMemory(ctx, memory)
    if err != nil {
        span.SetTag("error", true)
        span.LogFields(log.Error(err))
    }
    
    return err
}
```

## Security Best Practices

### API Key Management

```go
// Use environment variables
apiKey := os.Getenv("OPENAI_API_KEY")
if apiKey == "" {
    return errors.New("OPENAI_API_KEY not set")
}

// Or use secret management service
secret, err := secretManager.GetSecret(ctx, "openai-api-key")
if err != nil {
    return fmt.Errorf("failed to get API key: %w", err)
}
```

### Input Validation

```go
func ValidateMemoryInput(memory *models.Memory) error {
    if memory.ID == "" {
        return errors.New("memory ID is required")
    }
    
    if len(memory.Content) > 1_000_000 { // 1MB limit
        return errors.New("content exceeds size limit")
    }
    
    if len(memory.Embedding) == 0 {
        return errors.New("embedding is required")
    }
    
    if memory.WorkspaceID == "" {
        return errors.New("workspace ID is required")
    }
    
    return nil
}
```

### Rate Limiting

```go
type RateLimitedLLMService struct {
    service     services.LiteLLMService
    rateLimiter *rate.Limiter
}

func (r *RateLimitedLLMService) CallWithRetry(ctx context.Context, prompt string, retryOnJSON bool) (string, error) {
    // Wait for rate limiter
    if err := r.rateLimiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limit exceeded: %w", err)
    }
    
    return r.service.CallWithRetry(ctx, prompt, retryOnJSON)
}
```

## Troubleshooting

### Common Issues

1. **ChromaDB Connection Errors**
   - Check if ChromaDB is running: `docker ps`
   - Verify URL configuration
   - Check network connectivity

2. **Embedding Service Timeout**
   - Increase timeout in configuration
   - Check if embedding service is healthy
   - Consider batch processing for large inputs

3. **LLM Rate Limits**
   - Implement exponential backoff
   - Use multiple API keys
   - Consider caching responses

4. **Memory Issues**
   - Limit batch sizes
   - Implement streaming for large datasets
   - Use connection pooling

### Debug Logging

```go
// Enable debug logging
logger := zap.NewDevelopment()
defer logger.Sync()

// Log service interactions
logger.Debug("Calling LLM",
    zap.String("prompt", prompt[:100]),
    zap.String("model", model),
)
```