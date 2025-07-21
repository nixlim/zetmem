# Prompt System Template for MCP Servers

This document provides a complete template for implementing a prompt management system similar to ZetMem's architecture.

## Overview

The prompt system serves as the bridge between your MCP tools and LLM services, enabling:
- Dynamic prompt loading and caching
- Template-based prompt generation
- Model-specific configurations
- Hot reloading for development
- Version management

## Core Components

### 1. Prompt Manager Service

```go
// pkg/services/prompts/manager.go
package prompts

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "text/template"
    "time"
    
    "go.uber.org/zap"
    "gopkg.in/yaml.v3"
)

type Manager struct {
    config    Config
    logger    *zap.Logger
    cache     map[string]*PromptTemplate
    templates map[string]*template.Template
    mu        sync.RWMutex
    lastLoad  time.Time
}

type Config struct {
    PromptDir      string        `yaml:"prompt_dir" env:"PROMPT_DIR" default:"/app/prompts"`
    CacheEnabled   bool          `yaml:"cache_enabled" env:"CACHE_ENABLED" default:"true"`
    HotReload      bool          `yaml:"hot_reload" env:"HOT_RELOAD" default:"true"`
    ReloadInterval time.Duration `yaml:"reload_interval" env:"RELOAD_INTERVAL" default:"5s"`
}

type PromptTemplate struct {
    Name         string                 `yaml:"name"`
    Description  string                 `yaml:"description"`
    Version      string                 `yaml:"version"`
    Template     string                 `yaml:"template"`
    Model        ModelConfig            `yaml:"model"`
    Variables    []VariableDefinition   `yaml:"variables"`
    Metadata     map[string]interface{} `yaml:"metadata"`
}

type ModelConfig struct {
    Provider    string  `yaml:"provider"`
    Model       string  `yaml:"model"`
    Temperature float64 `yaml:"temperature"`
    MaxTokens   int     `yaml:"max_tokens"`
}

type VariableDefinition struct {
    Name        string `yaml:"name"`
    Type        string `yaml:"type"`
    Required    bool   `yaml:"required"`
    Default     string `yaml:"default"`
    Description string `yaml:"description"`
}

func NewManager(config Config, logger *zap.Logger) *Manager {
    return &Manager{
        config:    config,
        logger:    logger,
        cache:     make(map[string]*PromptTemplate),
        templates: make(map[string]*template.Template),
    }
}

func (m *Manager) Initialize(ctx context.Context) error {
    // Load all prompts on startup
    if err := m.loadPrompts(); err != nil {
        return fmt.Errorf("load prompts: %w", err)
    }
    
    // Start hot reload if enabled
    if m.config.HotReload {
        go m.watchForChanges(ctx)
    }
    
    return nil
}

func (m *Manager) GetPrompt(name string) (*PromptTemplate, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Check if reload needed
    if m.shouldReload() {
        m.mu.RUnlock()
        m.mu.Lock()
        if err := m.loadPrompts(); err != nil {
            m.logger.Error("Failed to reload prompts", zap.Error(err))
        }
        m.mu.Unlock()
        m.mu.RLock()
    }
    
    prompt, exists := m.cache[name]
    if !exists {
        return nil, fmt.Errorf("prompt not found: %s", name)
    }
    
    return prompt, nil
}

func (m *Manager) ExecutePrompt(name string, data interface{}) (string, error) {
    m.mu.RLock()
    tmpl, exists := m.templates[name]
    m.mu.RUnlock()
    
    if !exists {
        return "", fmt.Errorf("template not found: %s", name)
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("execute template: %w", err)
    }
    
    return buf.String(), nil
}

func (m *Manager) loadPrompts() error {
    newCache := make(map[string]*PromptTemplate)
    newTemplates := make(map[string]*template.Template)
    
    // Walk prompt directory
    err := filepath.Walk(m.config.PromptDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        // Skip non-YAML files
        if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
            return nil
        }
        
        // Load prompt
        prompt, err := m.loadPromptFile(path)
        if err != nil {
            m.logger.Error("Failed to load prompt", 
                zap.String("file", path),
                zap.Error(err))
            return nil // Continue loading other prompts
        }
        
        // Compile template
        tmpl, err := template.New(prompt.Name).Parse(prompt.Template)
        if err != nil {
            m.logger.Error("Failed to compile template",
                zap.String("prompt", prompt.Name),
                zap.Error(err))
            return nil
        }
        
        newCache[prompt.Name] = prompt
        newTemplates[prompt.Name] = tmpl
        
        return nil
    })
    
    if err != nil {
        return fmt.Errorf("walk prompt directory: %w", err)
    }
    
    // Update cache atomically
    m.mu.Lock()
    m.cache = newCache
    m.templates = newTemplates
    m.lastLoad = time.Now()
    m.mu.Unlock()
    
    m.logger.Info("Loaded prompts", 
        zap.Int("count", len(newCache)))
    
    return nil
}

func (m *Manager) loadPromptFile(path string) (*PromptTemplate, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read file: %w", err)
    }
    
    var prompt PromptTemplate
    if err := yaml.Unmarshal(data, &prompt); err != nil {
        return nil, fmt.Errorf("unmarshal yaml: %w", err)
    }
    
    // Validate prompt
    if err := m.validatePrompt(&prompt); err != nil {
        return nil, fmt.Errorf("validate prompt: %w", err)
    }
    
    return &prompt, nil
}

func (m *Manager) validatePrompt(prompt *PromptTemplate) error {
    if prompt.Name == "" {
        return fmt.Errorf("prompt name is required")
    }
    
    if prompt.Template == "" {
        return fmt.Errorf("prompt template is required")
    }
    
    if prompt.Version == "" {
        prompt.Version = "1.0.0"
    }
    
    return nil
}

func (m *Manager) shouldReload() bool {
    if !m.config.HotReload || !m.config.CacheEnabled {
        return true
    }
    
    return time.Since(m.lastLoad) > m.config.ReloadInterval
}

func (m *Manager) watchForChanges(ctx context.Context) {
    ticker := time.NewTicker(m.config.ReloadInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := m.checkAndReload(); err != nil {
                m.logger.Error("Failed to check and reload prompts", zap.Error(err))
            }
        }
    }
}

func (m *Manager) checkAndReload() error {
    // Check if any files have been modified
    needsReload := false
    
    err := filepath.Walk(m.config.PromptDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() || !strings.HasSuffix(path, ".yaml") {
            return nil
        }
        
        if info.ModTime().After(m.lastLoad) {
            needsReload = true
            return filepath.SkipDir // Stop walking
        }
        
        return nil
    })
    
    if err != nil && err != filepath.SkipDir {
        return fmt.Errorf("walk directory: %w", err)
    }
    
    if needsReload {
        m.logger.Info("Reloading prompts due to file changes")
        return m.loadPrompts()
    }
    
    return nil
}
```

### 2. Prompt Data Structures

```go
// pkg/services/prompts/data.go
package prompts

// PromptData represents the data passed to prompt templates
type PromptData struct {
    // Common fields
    Content     string                 `json:"content"`
    Context     string                 `json:"context"`
    Query       string                 `json:"query"`
    
    // Memory-specific fields
    ProjectPath string                 `json:"project_path"`
    CodeType    string                 `json:"code_type"`
    Memories    []MemoryData          `json:"memories"`
    
    // Extensible fields
    Custom      map[string]interface{} `json:"custom"`
}

type MemoryData struct {
    ID       string   `json:"id"`
    Content  string   `json:"content"`
    Context  string   `json:"context"`
    Keywords []string `json:"keywords"`
    Tags     []string `json:"tags"`
}

// Builder pattern for prompt data
type PromptDataBuilder struct {
    data PromptData
}

func NewPromptDataBuilder() *PromptDataBuilder {
    return &PromptDataBuilder{
        data: PromptData{
            Custom: make(map[string]interface{}),
        },
    }
}

func (b *PromptDataBuilder) WithContent(content string) *PromptDataBuilder {
    b.data.Content = content
    return b
}

func (b *PromptDataBuilder) WithContext(context string) *PromptDataBuilder {
    b.data.Context = context
    return b
}

func (b *PromptDataBuilder) WithQuery(query string) *PromptDataBuilder {
    b.data.Query = query
    return b
}

func (b *PromptDataBuilder) WithCustom(key string, value interface{}) *PromptDataBuilder {
    b.data.Custom[key] = value
    return b
}

func (b *PromptDataBuilder) Build() PromptData {
    return b.data
}
```

### 3. Example Prompt Files

#### Memory Analysis Prompt
```yaml
# prompts/memory_analysis.yaml
name: memory_analysis
description: Analyzes code content to extract structured information
version: 1.0.0

model:
  provider: openai
  model: gpt-4
  temperature: 0.3
  max_tokens: 1000

variables:
  - name: content
    type: string
    required: true
    description: The code content to analyze
  - name: code_type
    type: string
    required: false
    default: unknown
    description: Programming language or code type

template: |
  You are an expert code analyst. Analyze the following code and extract structured information.

  Code Type: {{.CodeType}}
  
  Code Content:
  ```{{.CodeType}}
  {{.Content}}
  ```

  Please analyze this code and return a JSON object with the following structure:
  {
    "summary": "A concise summary of what this code does",
    "keywords": ["keyword1", "keyword2", ...],
    "tags": ["tag1", "tag2", ...],
    "context": "Detailed explanation of the code's purpose and functionality",
    "complexity": "low|medium|high",
    "dependencies": ["dep1", "dep2", ...]
  }

  Focus on:
  1. The main purpose and functionality
  2. Key algorithms or patterns used
  3. External dependencies
  4. Potential use cases
  5. Notable design decisions

metadata:
  category: analysis
  tags: [code, analysis, memory]
```

#### Evolution Analysis Prompt
```yaml
# prompts/evolution_analysis.yaml
name: evolution_analysis
description: Analyzes memory network for patterns and improvements
version: 1.0.0

model:
  provider: anthropic
  model: claude-3-sonnet
  temperature: 0.5
  max_tokens: 2000

variables:
  - name: memories
    type: array
    required: true
    description: Array of memory objects to analyze

template: |
  You are an AI specializing in knowledge graph analysis and optimization.
  
  Analyze the following memory network and identify:
  1. Common patterns and themes
  2. Missing connections between related memories
  3. Opportunities for knowledge synthesis
  4. Redundant or outdated information
  
  Memories to analyze:
  {{range $index, $memory := .Memories}}
  Memory {{$index}}:
  - ID: {{$memory.ID}}
  - Context: {{$memory.Context}}
  - Keywords: {{join $memory.Keywords ", "}}
  - Tags: {{join $memory.Tags ", "}}
  {{end}}
  
  Return a JSON object with:
  {
    "patterns": [
      {
        "pattern": "description of pattern",
        "memory_ids": ["id1", "id2"],
        "confidence": 0.0-1.0
      }
    ],
    "new_connections": [
      {
        "source_id": "id1",
        "target_id": "id2",
        "relationship": "type",
        "reason": "explanation"
      }
    ],
    "improvements": [
      {
        "memory_id": "id",
        "action": "update|merge|delete",
        "suggestion": "specific improvement"
      }
    ],
    "synthesis_opportunities": [
      {
        "memory_ids": ["id1", "id2"],
        "potential_insight": "description"
      }
    ]
  }

metadata:
  category: evolution
  tags: [analysis, network, improvement]
```

### 4. Integration with Services

```go
// pkg/services/memory/llm_integration.go
package memory

import (
    "context"
    "encoding/json"
    "fmt"
    
    "myproject/pkg/services/llm"
    "myproject/pkg/services/prompts"
)

type LLMIntegration struct {
    llm     llm.Service
    prompts *prompts.Manager
}

func NewLLMIntegration(llm llm.Service, prompts *prompts.Manager) *LLMIntegration {
    return &LLMIntegration{
        llm:     llm,
        prompts: prompts,
    }
}

func (l *LLMIntegration) AnalyzeCode(ctx context.Context, content, codeType string) (*CodeAnalysis, error) {
    // Get prompt template
    promptTemplate, err := l.prompts.GetPrompt("memory_analysis")
    if err != nil {
        return nil, fmt.Errorf("get prompt: %w", err)
    }
    
    // Build prompt data
    data := prompts.NewPromptDataBuilder().
        WithContent(content).
        WithCustom("CodeType", codeType).
        Build()
    
    // Execute template
    prompt, err := l.prompts.ExecutePrompt("memory_analysis", data)
    if err != nil {
        return nil, fmt.Errorf("execute prompt: %w", err)
    }
    
    // Call LLM
    response, err := l.llm.Complete(ctx, llm.Request{
        Prompt:      prompt,
        Model:       promptTemplate.Model.Model,
        Temperature: promptTemplate.Model.Temperature,
        MaxTokens:   promptTemplate.Model.MaxTokens,
    })
    if err != nil {
        return nil, fmt.Errorf("llm complete: %w", err)
    }
    
    // Parse response
    var analysis CodeAnalysis
    if err := json.Unmarshal([]byte(response), &analysis); err != nil {
        return nil, fmt.Errorf("parse response: %w", err)
    }
    
    return &analysis, nil
}

type CodeAnalysis struct {
    Summary      string   `json:"summary"`
    Keywords     []string `json:"keywords"`
    Tags         []string `json:"tags"`
    Context      string   `json:"context"`
    Complexity   string   `json:"complexity"`
    Dependencies []string `json:"dependencies"`
}
```

### 5. Testing Prompts

```go
// pkg/services/prompts/manager_test.go
package prompts

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/zap/zaptest"
)

func TestManager_LoadPrompts(t *testing.T) {
    // Create temp directory with test prompts
    tempDir := t.TempDir()
    
    testPrompt := `
name: test_prompt
description: Test prompt
version: 1.0.0
model:
  provider: test
  model: test-model
  temperature: 0.5
  max_tokens: 100
template: |
  Hello {{.Name}}!
`
    
    err := os.WriteFile(
        filepath.Join(tempDir, "test.yaml"),
        []byte(testPrompt),
        0644,
    )
    require.NoError(t, err)
    
    // Create manager
    manager := NewManager(Config{
        PromptDir:    tempDir,
        CacheEnabled: true,
        HotReload:    false,
    }, zaptest.NewLogger(t))
    
    // Initialize
    err = manager.Initialize(context.Background())
    require.NoError(t, err)
    
    // Get prompt
    prompt, err := manager.GetPrompt("test_prompt")
    require.NoError(t, err)
    assert.Equal(t, "test_prompt", prompt.Name)
    assert.Equal(t, "Test prompt", prompt.Description)
    
    // Execute prompt
    result, err := manager.ExecutePrompt("test_prompt", map[string]string{
        "Name": "World",
    })
    require.NoError(t, err)
    assert.Equal(t, "Hello World!\n", result)
}

func TestManager_HotReload(t *testing.T) {
    tempDir := t.TempDir()
    
    // Create initial prompt
    initialPrompt := `
name: reload_test
template: Version 1
`
    
    promptPath := filepath.Join(tempDir, "reload.yaml")
    err := os.WriteFile(promptPath, []byte(initialPrompt), 0644)
    require.NoError(t, err)
    
    // Create manager with hot reload
    manager := NewManager(Config{
        PromptDir:      tempDir,
        CacheEnabled:   true,
        HotReload:      true,
        ReloadInterval: 100 * time.Millisecond,
    }, zaptest.NewLogger(t))
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    err = manager.Initialize(ctx)
    require.NoError(t, err)
    
    // Verify initial version
    result, err := manager.ExecutePrompt("reload_test", nil)
    require.NoError(t, err)
    assert.Equal(t, "Version 1", result)
    
    // Update prompt file
    updatedPrompt := `
name: reload_test
template: Version 2
`
    err = os.WriteFile(promptPath, []byte(updatedPrompt), 0644)
    require.NoError(t, err)
    
    // Wait for reload
    time.Sleep(200 * time.Millisecond)
    
    // Verify updated version
    result, err = manager.ExecutePrompt("reload_test", nil)
    require.NoError(t, err)
    assert.Equal(t, "Version 2", result)
}
```

## Best Practices

### 1. Prompt Organization

```
prompts/
├── analysis/
│   ├── code_analysis.yaml
│   ├── complexity_analysis.yaml
│   └── dependency_analysis.yaml
├── evolution/
│   ├── network_evolution.yaml
│   └── pattern_detection.yaml
├── generation/
│   ├── documentation.yaml
│   └── test_generation.yaml
└── common/
    └── base_instructions.yaml
```

### 2. Version Management

```yaml
# Include version in prompt files
version: 1.2.0

# Track changes in metadata
metadata:
  changelog:
    - version: 1.2.0
      changes: ["Improved keyword extraction", "Added complexity analysis"]
    - version: 1.1.0
      changes: ["Initial release"]
```

### 3. Template Functions

```go
// Add custom template functions
func (m *Manager) loadPrompts() error {
    // ... existing code ...
    
    // Add custom functions
    funcMap := template.FuncMap{
        "join": strings.Join,
        "lower": strings.ToLower,
        "upper": strings.ToUpper,
        "contains": strings.Contains,
        "default": func(def, val interface{}) interface{} {
            if val == nil || val == "" {
                return def
            }
            return val
        },
    }
    
    tmpl, err := template.New(prompt.Name).Funcs(funcMap).Parse(prompt.Template)
    // ...
}
```

### 4. Prompt Testing

```go
// Create prompt test harness
type PromptTester struct {
    manager *Manager
    llm     *MockLLMService
}

func (pt *PromptTester) TestPrompt(t *testing.T, promptName string, input interface{}, expectedOutput interface{}) {
    // Execute prompt
    prompt, err := pt.manager.ExecutePrompt(promptName, input)
    require.NoError(t, err)
    
    // Set mock response
    pt.llm.SetResponse(expectedOutput)
    
    // Verify prompt structure
    assert.Contains(t, prompt, "required_keyword")
    assert.NotContains(t, prompt, "forbidden_phrase")
    
    // Test with actual LLM if available
    if os.Getenv("TEST_WITH_LLM") == "true" {
        response, err := pt.llm.Complete(context.Background(), prompt)
        require.NoError(t, err)
        
        // Validate response structure
        var result map[string]interface{}
        err = json.Unmarshal([]byte(response), &result)
        require.NoError(t, err)
    }
}
```

### 5. Performance Optimization

```go
// Implement prompt result caching
type CachedPromptManager struct {
    *Manager
    cache      *lru.Cache
    cacheTTL   time.Duration
}

func (m *CachedPromptManager) ExecutePromptCached(name string, data interface{}) (string, error) {
    // Generate cache key
    key := fmt.Sprintf("%s:%v", name, data)
    
    // Check cache
    if cached, ok := m.cache.Get(key); ok {
        return cached.(string), nil
    }
    
    // Execute prompt
    result, err := m.ExecutePrompt(name, data)
    if err != nil {
        return "", err
    }
    
    // Cache result
    m.cache.Add(key, result)
    
    return result, nil
}
```

## Deployment Considerations

### 1. Prompt Directory Structure in Docker

```dockerfile
# Copy prompts to container
COPY prompts /app/prompts

# Set permissions
RUN chown -R app:app /app/prompts && \
    chmod -R 644 /app/prompts/*.yaml
```

### 2. Environment Configuration

```bash
# Development
PROMPT_DIR=./prompts
HOT_RELOAD=true
CACHE_ENABLED=false

# Production
PROMPT_DIR=/app/prompts
HOT_RELOAD=false
CACHE_ENABLED=true
```

### 3. Monitoring

```go
// Add metrics for prompt usage
var (
    PromptExecutions = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "prompt_executions_total",
            Help: "Total number of prompt executions",
        },
        []string{"prompt", "status"},
    )
    
    PromptDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "prompt_duration_seconds",
            Help: "Prompt execution duration",
        },
        []string{"prompt"},
    )
)
```

This prompt system template provides a flexible, maintainable foundation for managing LLM interactions in your MCP server. The design supports easy iteration on prompts, model-specific configurations, and clean integration with your services.