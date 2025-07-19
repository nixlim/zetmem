package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/zetmem/mcp-server/pkg/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// PromptManager manages prompt templates //TODO: Consider wiring this in
type PromptManager struct {
	config    config.PromptsConfig
	logger    *zap.Logger
	cache     map[string]*PromptTemplate
	templates map[string]*template.Template
	mu        sync.RWMutex
	lastLoad  time.Time
}

// PromptTemplate represents a prompt template configuration
type PromptTemplate struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	ModelConfig ModelConfig            `yaml:"model_config"`
	Template    string                 `yaml:"template"`
	Variables   map[string]interface{} `yaml:"variables,omitempty"`
	Metadata    map[string]interface{} `yaml:"metadata,omitempty"`
}

// ModelConfig represents model-specific configuration for prompts
type ModelConfig struct {
	Temperature float32 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
	TopP        float32 `yaml:"top_p,omitempty"`
	TopK        int     `yaml:"top_k,omitempty"`
}

// PromptData represents data to be injected into a prompt template
type PromptData struct {
	Content     string
	ProjectPath string
	CodeType    string
	Context     string
	Query       string
	Memories    interface{}
	Custom      map[string]interface{}
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(cfg config.PromptsConfig, logger *zap.Logger) *PromptManager {
	return &PromptManager{
		config:    cfg,
		logger:    logger,
		cache:     make(map[string]*PromptTemplate),
		templates: make(map[string]*template.Template),
	}
}

// LoadPrompt loads a prompt template by name
func (pm *PromptManager) LoadPrompt(name string) (*PromptTemplate, error) {
	pm.mu.RLock()

	// Check cache first
	if pm.config.CacheEnabled {
		if cached, ok := pm.cache[name]; ok {
			// Check if hot reload is enabled and file has changed
			if !pm.config.HotReload || !pm.shouldReload(name) {
				pm.mu.RUnlock()
				return cached, nil
			}
		}
	}

	pm.mu.RUnlock()

	// Load from file
	prompt, err := pm.loadFromFile(name)
	if err != nil {
		return nil, err
	}

	// Cache the prompt
	if pm.config.CacheEnabled {
		pm.mu.Lock()
		pm.cache[name] = prompt
		pm.lastLoad = time.Now()
		pm.mu.Unlock()
	}

	return prompt, nil
}

// RenderPrompt renders a prompt template with the given data
func (pm *PromptManager) RenderPrompt(name string, data PromptData) (string, error) {
	prompt, err := pm.LoadPrompt(name)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %s: %w", name, err)
	}

	// Get or create compiled template
	tmpl, err := pm.getCompiledTemplate(name, prompt.Template)
	if err != nil {
		return "", fmt.Errorf("failed to compile template %s: %w", name, err)
	}

	// Prepare template data
	templateData := pm.prepareTemplateData(data, prompt.Variables)

	// Render template
	var result strings.Builder
	if err := tmpl.Execute(&result, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	pm.logger.Debug("Prompt rendered successfully",
		zap.String("name", name),
		zap.Int("length", len(result.String())))

	return result.String(), nil
}

// GetModelConfig returns the model configuration for a prompt
func (pm *PromptManager) GetModelConfig(name string) (*ModelConfig, error) {
	prompt, err := pm.LoadPrompt(name)
	if err != nil {
		return nil, err
	}

	return &prompt.ModelConfig, nil
}

// ListPrompts returns a list of available prompt names
func (pm *PromptManager) ListPrompts() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(pm.config.Directory, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to list prompt files: %w", err)
	}

	var names []string
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".yaml")
		names = append(names, name)
	}

	return names, nil
}

// ClearCache clears the prompt cache
func (pm *PromptManager) ClearCache() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.cache = make(map[string]*PromptTemplate)
	pm.templates = make(map[string]*template.Template)
	pm.logger.Info("Prompt cache cleared")
}

// loadFromFile loads a prompt template from a YAML file
func (pm *PromptManager) loadFromFile(name string) (*PromptTemplate, error) {
	path := filepath.Join(pm.config.Directory, name+".yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file %s: %w", path, err)
	}

	var prompt PromptTemplate
	if err := yaml.Unmarshal(data, &prompt); err != nil {
		return nil, fmt.Errorf("failed to parse prompt file %s: %w", path, err)
	}

	// Validate prompt
	if err := pm.validatePrompt(&prompt); err != nil {
		return nil, fmt.Errorf("invalid prompt %s: %w", name, err)
	}

	pm.logger.Debug("Prompt loaded from file",
		zap.String("name", name),
		zap.String("version", prompt.Version))

	return &prompt, nil
}

// getCompiledTemplate gets or creates a compiled template
func (pm *PromptManager) getCompiledTemplate(name, templateStr string) (*template.Template, error) {
	pm.mu.RLock()
	if tmpl, ok := pm.templates[name]; ok {
		pm.mu.RUnlock()
		return tmpl, nil
	}
	pm.mu.RUnlock()

	// Compile template
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return nil, err
	}

	// Cache compiled template
	pm.mu.Lock()
	pm.templates[name] = tmpl
	pm.mu.Unlock()

	return tmpl, nil
}

// prepareTemplateData prepares data for template execution
func (pm *PromptManager) prepareTemplateData(data PromptData, variables map[string]interface{}) map[string]interface{} {
	templateData := map[string]interface{}{
		"Content":     data.Content,
		"ProjectPath": data.ProjectPath,
		"CodeType":    data.CodeType,
		"Context":     data.Context,
		"Query":       data.Query,
		"Memories":    data.Memories,
	}

	// Add custom data
	for k, v := range data.Custom {
		templateData[k] = v
	}

	// Add template variables
	for k, v := range variables {
		templateData[k] = v
	}

	return templateData
}

// validatePrompt validates a prompt template
func (pm *PromptManager) validatePrompt(prompt *PromptTemplate) error {
	if prompt.Name == "" {
		return fmt.Errorf("prompt name is required")
	}

	if prompt.Template == "" {
		return fmt.Errorf("prompt template is required")
	}

	// Validate model config
	if prompt.ModelConfig.Temperature < 0 || prompt.ModelConfig.Temperature > 2 {
		return fmt.Errorf("invalid temperature: %f", prompt.ModelConfig.Temperature)
	}

	if prompt.ModelConfig.MaxTokens <= 0 {
		return fmt.Errorf("invalid max_tokens: %d", prompt.ModelConfig.MaxTokens)
	}

	return nil
}

// shouldReload checks if a prompt should be reloaded
func (pm *PromptManager) shouldReload(name string) bool {
	if !pm.config.HotReload {
		return false
	}

	path := filepath.Join(pm.config.Directory, name+".yaml")
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.ModTime().After(pm.lastLoad)
}

// GetPromptNames returns all available prompt names
func (pm *PromptManager) GetPromptNames() []string {
	names, err := pm.ListPrompts()
	if err != nil {
		pm.logger.Error("Failed to list prompts", zap.Error(err))
		return []string{}
	}
	return names
}
