package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/amem/mcp-server/pkg/config"
	"github.com/amem/mcp-server/pkg/models"
	"github.com/amem/mcp-server/pkg/services"
	"go.uber.org/zap"
)

// WorkspaceInitTool implements smart workspace initialization
type WorkspaceInitTool struct {
	workspaceService *services.WorkspaceService
	logger           *zap.Logger
}

// NewWorkspaceInitTool creates a new workspace init tool
func NewWorkspaceInitTool(workspaceService *services.WorkspaceService, logger *zap.Logger) *WorkspaceInitTool {
	return &WorkspaceInitTool{
		workspaceService: workspaceService,
		logger:           logger,
	}
}

func (t *WorkspaceInitTool) Name() string {
	return "workspace_init"
}

func (t *WorkspaceInitTool) Description() string {
	return "Smart workspace initialization - creates new workspace or retrieves existing one. If no identifier provided, uses current working directory."
}

func (t *WorkspaceInitTool) UsageTriggers() []string {
	return []string{
		"At the start of each new project or when switching to a different codebase",
		"When beginning work in a new directory or repository",
		"Before storing the first memory in a new context",
		"When organizing memories by project or domain-specific themes",
		"At the beginning of each coding session to establish workspace context",
	}
}

func (t *WorkspaceInitTool) BestPractices() []string {
	return []string{
		"Use filesystem paths for project-specific workspaces (e.g., '/Users/dev/my-project')",
		"Use logical names for cross-project themes (e.g., 'react-patterns', 'debugging-techniques')",
		"Provide descriptive names to make workspaces easily identifiable",
		"Initialize workspace before storing any memories to ensure proper organization",
		"Use current working directory as default when working within a specific project",
	}
}

func (t *WorkspaceInitTool) Synergies() map[string][]string {
	return map[string][]string{
		"precedes": {"store_coding_memory", "retrieve_relevant_memories", "workspace_retrieve"},
		"succeeds": {},
	}
}

func (t *WorkspaceInitTool) WorkflowSnippets() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"goal": "Start a new coding session in a project",
			"steps": []string{
				"1. workspace_init with project directory path as identifier",
				"2. retrieve_relevant_memories to load existing context",
				"3. Begin coding with workspace context established",
			},
		},
		{
			"goal": "Create a theme-based workspace for learning",
			"steps": []string{
				"1. workspace_init with logical name (e.g., 'machine-learning-patterns')",
				"2. Provide descriptive name for easy identification",
				"3. store_coding_memory with relevant examples and insights",
			},
		},
	}
}

func (t *WorkspaceInitTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"identifier": map[string]interface{}{
				"type":        "string",
				"description": "Path or name for the workspace. If not provided, uses current working directory",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable name for the workspace (optional)",
			},
		},
		"required": []string{},
	}
}

func (t *WorkspaceInitTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var req models.WorkspaceRequest

	if identifier, ok := args["identifier"].(string); ok {
		req.Identifier = identifier
	}

	if name, ok := args["name"].(string); ok {
		req.Name = name
	}

	// Initialize workspace
	workspace, created, err := t.workspaceService.InitializeWorkspace(ctx, &req)
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error initializing workspace: %v", err),
			}},
		}, nil
	}

	// Create response
	response := models.WorkspaceResponse{
		Workspace: *workspace,
		Created:   created,
	}

	action := "Retrieved"
	if created {
		action = "Created"
	}

	// Serialize response to JSON
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error serializing response: %v", err),
			}},
		}, nil
	}

	return &models.MCPToolResult{
		IsError: false,
		Content: []models.MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("%s workspace '%s' (%s)\n\nWorkspace Details:\n```json\n%s\n```",
					action, workspace.Name, workspace.ID, string(responseJSON)),
			},
		},
	}, nil
}

// WorkspaceCreateTool implements explicit workspace creation
type WorkspaceCreateTool struct {
	workspaceService *services.WorkspaceService
	logger           *zap.Logger
}

// NewWorkspaceCreateTool creates a new workspace create tool
func NewWorkspaceCreateTool(workspaceService *services.WorkspaceService, logger *zap.Logger) *WorkspaceCreateTool {
	return &WorkspaceCreateTool{
		workspaceService: workspaceService,
		logger:           logger,
	}
}

func (t *WorkspaceCreateTool) Name() string {
	return "workspace_create"
}

func (t *WorkspaceCreateTool) Description() string {
	return "Explicit workspace creation - fails if workspace already exists. Supports both filesystem paths and logical names."
}

func (t *WorkspaceCreateTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"identifier": map[string]interface{}{
				"type":        "string",
				"description": "Path or name for the workspace (required)",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Human-readable name for the workspace (optional)",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Description of the workspace (optional)",
			},
		},
		"required": []string{"identifier"},
	}
}

func (t *WorkspaceCreateTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var req models.WorkspaceRequest

	if identifier, ok := args["identifier"].(string); ok {
		req.Identifier = identifier
	} else {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: "Error: 'identifier' parameter is required",
			}},
		}, nil
	}

	if name, ok := args["name"].(string); ok {
		req.Name = name
	}

	if description, ok := args["description"].(string); ok {
		req.Description = description
	}

	// Create workspace
	workspace, err := t.workspaceService.CreateWorkspace(ctx, &req)
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error creating workspace: %v", err),
			}},
		}, nil
	}

	// Create response
	response := models.WorkspaceResponse{
		Workspace: *workspace,
		Created:   true,
	}

	// Serialize response to JSON
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error serializing response: %v", err),
			}},
		}, nil
	}

	return &models.MCPToolResult{
		IsError: false,
		Content: []models.MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Created workspace '%s' (%s)\n\nWorkspace Details:\n```json\n%s\n```",
					workspace.Name, workspace.ID, string(responseJSON)),
			},
		},
	}, nil
}

// WorkspaceRetrieveTool implements explicit workspace retrieval
type WorkspaceRetrieveTool struct {
	workspaceService *services.WorkspaceService
	logger           *zap.Logger
}

// NewWorkspaceRetrieveTool creates a new workspace retrieve tool
func NewWorkspaceRetrieveTool(workspaceService *services.WorkspaceService, logger *zap.Logger) *WorkspaceRetrieveTool {
	return &WorkspaceRetrieveTool{
		workspaceService: workspaceService,
		logger:           logger,
	}
}

func (t *WorkspaceRetrieveTool) Name() string {
	return "workspace_retrieve"
}

func (t *WorkspaceRetrieveTool) Description() string {
	return "Explicit workspace retrieval - fails if workspace doesn't exist. Returns comprehensive workspace metadata including memory count."
}

func (t *WorkspaceRetrieveTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"identifier": map[string]interface{}{
				"type":        "string",
				"description": "Path or name of the workspace to retrieve (required)",
			},
		},
		"required": []string{"identifier"},
	}
}

func (t *WorkspaceRetrieveTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var req models.WorkspaceRequest

	if identifier, ok := args["identifier"].(string); ok {
		req.Identifier = identifier
	} else {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: "Error: 'identifier' parameter is required",
			}},
		}, nil
	}

	// Normalize workspace ID
	workspaceID := t.workspaceService.NormalizeWorkspaceID(req.Identifier)

	// Check if workspace exists
	exists, err := t.workspaceService.WorkspaceExists(ctx, workspaceID)
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error checking workspace existence: %v", err),
			}},
		}, nil
	}

	if !exists {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Workspace '%s' does not exist", workspaceID),
			}},
		}, nil
	}

	// Get workspace info
	workspace, err := t.workspaceService.GetWorkspaceInfo(ctx, workspaceID)
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error retrieving workspace info: %v", err),
			}},
		}, nil
	}

	// Create response
	response := models.WorkspaceResponse{
		Workspace: *workspace,
		Created:   false,
	}

	// Serialize response to JSON
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error serializing response: %v", err),
			}},
		}, nil
	}

	return &models.MCPToolResult{
		IsError: false,
		Content: []models.MCPContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Retrieved workspace '%s' (%s) with %d memories\n\nWorkspace Details:\n```json\n%s\n```",
					workspace.Name, workspace.ID, workspace.MemoryCount, string(responseJSON)),
			},
		},
	}, nil
}

// PerformOnboardingTool implements comprehensive agent onboarding
type PerformOnboardingTool struct {
	workspaceService *services.WorkspaceService
	config           config.OnboardingConfig
	logger           *zap.Logger
	strategyGuide    string // Cached strategy guide content
}

// NewPerformOnboardingTool creates a new perform onboarding tool
func NewPerformOnboardingTool(workspaceService *services.WorkspaceService, onboardingConfig config.OnboardingConfig, logger *zap.Logger) *PerformOnboardingTool {
	tool := &PerformOnboardingTool{
		workspaceService: workspaceService,
		config:           onboardingConfig,
		logger:           logger,
	}

	// Load and cache strategy guide at initialization
	content, err := tool.loadStrategyGuide()
	if err != nil {
		logger.Warn("Could not load strategy guide at initialization", zap.Error(err))
		tool.strategyGuide = "Strategy guide not available. Please refer to ZETMEM_ONBOARDING_STRATEGY.md for complete guidance."
	} else {
		tool.strategyGuide = content
		logger.Info("Strategy guide loaded and cached successfully", zap.Int("size", len(content)))
	}

	return tool
}

func (t *PerformOnboardingTool) Name() string {
	return "perform_onboarding"
}

func (t *PerformOnboardingTool) Description() string {
	return "Comprehensive agent onboarding - initializes workspace and provides complete tool use strategy and best practices"
}

func (t *PerformOnboardingTool) UsageTriggers() []string {
	return []string{
		"At the very beginning of agent interaction with a new codebase or project",
		"When starting work in a new directory or switching to a different project context",
		"Before beginning any coding session to establish proper workspace and strategy context",
		"When an agent needs to understand the complete zetmem workflow and best practices",
		"As the first command in any new coding collaboration session",
	}
}

func (t *PerformOnboardingTool) BestPractices() []string {
	return []string{
		"Always run this as the first command when starting work in a new context",
		"Provide a descriptive project name to make the workspace easily identifiable",
		"Read and internalize the complete strategy guide returned by this command",
		"Use the workspace_id returned for all subsequent memory operations",
		"Follow the workflow patterns and thresholds specified in the strategy guide",
	}
}

func (t *PerformOnboardingTool) Synergies() map[string][]string {
	return map[string][]string{
		"precedes": {"store_coding_memory", "retrieve_relevant_memories", "evolve_memory_network"},
		"succeeds": {},
	}
}

func (t *PerformOnboardingTool) WorkflowSnippets() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"goal": "Complete agent onboarding for new project",
			"steps": []string{
				"1. perform_onboarding with project directory and descriptive name",
				"2. Review the complete strategy guide and internalize best practices",
				"3. Begin coding session with proper workspace context established",
				"4. Follow the workflow patterns specified in the strategy guide",
			},
		},
	}
}

func (t *PerformOnboardingTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the project directory (optional, uses current directory if not provided)",
			},
			"project_name": map[string]interface{}{
				"type":        "string",
				"description": "Descriptive name for the project workspace (optional)",
			},
			"include_strategy_guide": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to include the complete strategy guide in response (default: true)",
				"default":     true,
			},
		},
		"required": []string{},
	}
}

func (t *PerformOnboardingTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var projectPath string
	var projectName string
	includeStrategyGuide := true

	if path, ok := args["project_path"].(string); ok {
		projectPath = path
	}

	if name, ok := args["project_name"].(string); ok {
		projectName = name
	}

	if include, ok := args["include_strategy_guide"].(bool); ok {
		includeStrategyGuide = include
	}

	// Validate project_path if provided
	if projectPath != "" {
		if err := t.validateProjectPath(projectPath); err != nil {
			return &models.MCPToolResult{
				IsError: true,
				Content: []models.MCPContent{{
					Type: "text",
					Text: fmt.Sprintf("Invalid project_path: %v", err),
				}},
			}, nil
		}
	}

	// Step 1: Initialize workspace
	workspaceReq := models.WorkspaceRequest{
		Identifier: projectPath,
		Name:       projectName,
	}

	workspace, created, err := t.workspaceService.InitializeWorkspace(ctx, &workspaceReq)
	if err != nil {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Error initializing workspace: %v", err),
			}},
		}, nil
	}

	// Step 2: Use cached strategy guide if requested
	var strategyGuide string
	if includeStrategyGuide {
		strategyGuide = t.strategyGuide
	}

	// Step 3: Create comprehensive response
	action := "Retrieved"
	if created {
		action = "Created"
	}

	response := fmt.Sprintf(`🎯 **Zetmem Agent Onboarding Complete**

## Workspace Initialization
%s workspace '%s' (%s)
- **Workspace ID**: %s
- **Memory Count**: %d
- **Status**: Ready for use

## Quick Start Commands
1. **Store Memory**: store_coding_memory(content="...", workspace_id="%s", code_type="...", context="...")
2. **Retrieve Memories**: retrieve_relevant_memories(query="...", workspace_id="%s", min_relevance=0.3)
3. **Evolve Network**: evolve_memory_network(scope="recent", max_memories=100)

## Key Principles
- **Workspace-First**: Always specify workspace_id for proper organization
- **Consistent Habits**: Store memories after solving problems, retrieve before starting tasks
- **Threshold Management**: Start with min_relevance=0.3, increase for precision
- **Regular Evolution**: Run evolution weekly or after 10+ new memories

`, action, workspace.Name, workspace.ID, workspace.ID, workspace.MemoryCount, workspace.ID, workspace.ID)

	if includeStrategyGuide && strategyGuide != "" {
		response += fmt.Sprintf(`
## Complete Strategy Guide

%s

---

**You are now ready to use zetmem effectively!** Follow the patterns above for optimal memory management and knowledge preservation.`, strategyGuide)
	} else {
		response += `
## Next Steps
- Review the complete strategy guide at: ZETMEM_ONBOARDING_STRATEGY.md
- Follow the workflow patterns for effective memory management
- Use the workspace_id provided for all memory operations

**You are now ready to use zetmem effectively!**`
	}

	return &models.MCPToolResult{
		IsError: false,
		Content: []models.MCPContent{
			{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

// loadStrategyGuide loads the strategy guide from the configured path with validation
func (t *PerformOnboardingTool) loadStrategyGuide() (string, error) {
	path := t.config.StrategyGuidePath
	if path == "" {
		return "", fmt.Errorf("strategy guide path not configured")
	}

	// Check if file exists and get size
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("strategy guide not found at path: %s", path)
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat strategy guide: %w", err)
	}

	// Validate file size
	if info.Size() > t.config.MaxFileSize {
		return "", fmt.Errorf("strategy guide exceeds size limit of %d bytes (actual: %d bytes)",
			t.config.MaxFileSize, info.Size())
	}

	// Read file content using modern Go practices
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read strategy guide: %w", err)
	}

	return string(content), nil
}

// validateProjectPath validates the project path parameter
func (t *PerformOnboardingTool) validateProjectPath(path string) error {
	// Check for null bytes (security)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null bytes")
	}

	// Check reasonable length limit
	if len(path) > 4096 {
		return fmt.Errorf("path too long (max 4096 characters)")
	}

	// Check for empty path
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be empty")
	}

	return nil
}
