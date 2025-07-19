package memory

import (
	"context"
	"fmt"

	"github.com/amem/mcp-server/pkg/models"
	"go.uber.org/zap"
)

// StoreCodingMemoryTool implements the store_coding_memory MCP tool
type StoreCodingMemoryTool struct {
	system *System
	logger *zap.Logger
}

// NewStoreCodingMemoryTool creates a new store coding memory tool
func NewStoreCodingMemoryTool(system *System, logger *zap.Logger) *StoreCodingMemoryTool {
	return &StoreCodingMemoryTool{
		system: system,
		logger: logger,
	}
}

func (t *StoreCodingMemoryTool) Name() string {
	return models.ToolStoreCodingMemory
}

func (t *StoreCodingMemoryTool) Description() string {
	return "Store a coding memory with AI-generated analysis, keywords, tags, and embeddings for future retrieval"
}

func (t *StoreCodingMemoryTool) UsageTriggers() []string {
	return []string{
		"After solving a non-trivial problem or implementing a solution",
		"When discovering new patterns, techniques, or architectural decisions",
		"After implementing significant features or code components",
		"When gaining insights about the codebase or debugging complex issues",
		"Before context switches to preserve important knowledge",
	}
}

func (t *StoreCodingMemoryTool) BestPractices() []string {
	return []string{
		"Always specify workspace_id to maintain proper memory organization",
		"Provide rich context explaining the problem, solution, and rationale",
		"Include relevant code_type for better categorization and retrieval",
		"Store memories with token_count ≤ 350 for optimal processing",
		"Focus on reusable patterns and insights rather than one-off fixes",
	}
}

func (t *StoreCodingMemoryTool) Synergies() map[string][]string {
	return map[string][]string{
		"precedes": {"retrieve_relevant_memories", "evolve_memory_network"},
		"succeeds": {"workspace_init", "retrieve_relevant_memories"},
	}
}

func (t *StoreCodingMemoryTool) WorkflowSnippets() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"goal": "Store a problem-solution pattern after debugging",
			"steps": []string{
				"1. store_coding_memory with problem description and solution code",
				"2. Include context explaining why this solution works",
				"3. Specify code_type and workspace_id for proper organization",
			},
		},
		{
			"goal": "Capture architectural decision for future reference",
			"steps": []string{
				"1. store_coding_memory with decision rationale and implementation",
				"2. Include context about alternatives considered and trade-offs",
				"3. Tag with relevant architectural patterns and technologies",
			},
		},
	}
}

func (t *StoreCodingMemoryTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The code content or coding context to store",
			},
			"workspace_id": map[string]interface{}{
				"type":        "string",
				"description": "Workspace identifier (path or name) for organizing memories",
			},
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Optional project path for context (deprecated: use workspace_id)",
			},
			"code_type": map[string]interface{}{
				"type":        "string",
				"description": "Programming language or code type (e.g., 'javascript', 'python', 'go')",
			},
			"context": map[string]interface{}{
				"type":        "string",
				"description": "Additional context about the code",
			},
		},
		"required": []string{"content"},
	}
}

func (t *StoreCodingMemoryTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var req models.StoreMemoryRequest

	if content, ok := args["content"].(string); ok {
		req.Content = content
	} else {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: "Error: 'content' parameter is required and must be a string",
			}},
		}, nil
	}

	if workspaceID, ok := args["workspace_id"].(string); ok {
		req.WorkspaceID = workspaceID
	}

	if projectPath, ok := args["project_path"].(string); ok {
		req.ProjectPath = projectPath
	}

	if codeType, ok := args["code_type"].(string); ok {
		req.CodeType = codeType
	}

	if context, ok := args["context"].(string); ok {
		req.Context = context
	}

	// Execute memory creation
	response, err := t.system.CreateMemory(ctx, req)
	if err != nil {
		t.logger.Error("Failed to store memory", zap.Error(err))
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to store memory: %v", err),
			}},
		}, nil
	}

	// Format response
	resultText := fmt.Sprintf(`Memory stored successfully!

Memory ID: %s
Keywords: %v
Tags: %v
Links Created: %d

The memory has been analyzed and stored with AI-generated keywords and tags. It's now available for future retrieval and will be linked to related memories.`,
		response.MemoryID,
		response.Keywords,
		response.Tags,
		response.LinksCreated)

	return &models.MCPToolResult{
		Content: []models.MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}, nil
}

// RetrieveRelevantMemoriesTool implements the retrieve_relevant_memories MCP tool
type RetrieveRelevantMemoriesTool struct {
	system *System
	logger *zap.Logger
}

// NewRetrieveRelevantMemoriesTool creates a new retrieve relevant memories tool
func NewRetrieveRelevantMemoriesTool(system *System, logger *zap.Logger) *RetrieveRelevantMemoriesTool {
	return &RetrieveRelevantMemoriesTool{
		system: system,
		logger: logger,
	}
}

func (t *RetrieveRelevantMemoriesTool) Name() string {
	return models.ToolRetrieveRelevantMemories
}

func (t *RetrieveRelevantMemoriesTool) Description() string {
	return "Retrieve relevant coding memories based on a query using vector similarity search"
}

func (t *RetrieveRelevantMemoriesTool) UsageTriggers() []string {
	return []string{
		"At the start of each coding session to retrieve relevant context",
		"Before implementing features to find similar patterns or solutions",
		"When encountering problems to find previous solutions or approaches",
		"During debugging to locate related error handling patterns",
		"When exploring new technologies to find existing knowledge",
	}
}

func (t *RetrieveRelevantMemoriesTool) BestPractices() []string {
	return []string{
		"Start with min_relevance: 0.3 to cast a wide net for potentially relevant memories",
		"If too many results (>15), incrementally increase threshold (0.4, 0.5, 0.6)",
		"Use specific keywords and technical terms in queries for better results",
		"Always specify workspace_id to scope searches to relevant project context",
		"Combine multiple queries for complex topics requiring different perspectives",
	}
}

func (t *RetrieveRelevantMemoriesTool) Synergies() map[string][]string {
	return map[string][]string{
		"precedes": {"store_coding_memory", "evolve_memory_network"},
		"succeeds": {"workspace_init", "workspace_retrieve"},
	}
}

func (t *RetrieveRelevantMemoriesTool) WorkflowSnippets() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"goal": "Find relevant patterns before implementing a feature",
			"steps": []string{
				"1. retrieve_relevant_memories with feature description and workspace_id",
				"2. Review returned patterns and solutions",
				"3. Adapt existing patterns or create new approach",
				"4. store_coding_memory with new implementation",
			},
		},
		{
			"goal": "Debug an issue using previous solutions",
			"steps": []string{
				"1. retrieve_relevant_memories with error description and code_types",
				"2. Start with min_relevance: 0.3 for broad search",
				"3. If too many results, increase threshold for precision",
				"4. Apply solution and store_coding_memory with fix",
			},
		},
	}
}

func (t *RetrieveRelevantMemoriesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query (code snippet, problem description, or keywords)",
			},
			"workspace_id": map[string]interface{}{
				"type":        "string",
				"description": "Workspace identifier to filter results (optional)",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 5)",
				"default":     5,
			},
			"project_filter": map[string]interface{}{
				"type":        "string",
				"description": "Optional project path to filter results (deprecated: use workspace_id)",
			},
			"code_types": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional array of code types to filter by",
			},
			"min_relevance": map[string]interface{}{
				"type":        "number",
				"description": "Minimum relevance score (0.0-1.0, default: 0.7)",
				"default":     0.7,
			},
		},
		"required": []string{"query"},
	}
}

func (t *RetrieveRelevantMemoriesTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var req models.RetrieveMemoryRequest

	if query, ok := args["query"].(string); ok {
		req.Query = query
	} else {
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: "Error: 'query' parameter is required and must be a string",
			}},
		}, nil
	}

	if workspaceID, ok := args["workspace_id"].(string); ok {
		req.WorkspaceID = workspaceID
	}

	if maxResults, ok := args["max_results"].(float64); ok {
		req.MaxResults = int(maxResults)
	} else {
		req.MaxResults = 5
	}

	if projectFilter, ok := args["project_filter"].(string); ok {
		req.ProjectFilter = projectFilter
	}

	if codeTypesInterface, ok := args["code_types"].([]interface{}); ok {
		codeTypes := make([]string, len(codeTypesInterface))
		for i, ct := range codeTypesInterface {
			if ctStr, ok := ct.(string); ok {
				codeTypes[i] = ctStr
			}
		}
		req.CodeTypes = codeTypes
	}

	if minRelevance, ok := args["min_relevance"].(float64); ok {
		req.MinRelevance = float32(minRelevance)
	} else {
		req.MinRelevance = 0.7
	}

	// Execute memory retrieval
	response, err := t.system.RetrieveMemories(ctx, req)
	if err != nil {
		t.logger.Error("Failed to retrieve memories", zap.Error(err))
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Failed to retrieve memories: %v", err),
			}},
		}, nil
	}

	// Format response
	if len(response.Memories) == 0 {
		return &models.MCPToolResult{
			Content: []models.MCPContent{{
				Type: "text",
				Text: "No relevant memories found for your query. Try adjusting your search terms or lowering the relevance threshold.",
			}},
		}, nil
	}

	resultText := fmt.Sprintf("Found %d relevant memories:\n\n", response.TotalFound)

	for i, memory := range response.Memories {
		resultText += fmt.Sprintf("**Memory %d** (Relevance: %.1f%%)\nID: %s\nContext: %s\nKeywords: %v\nTags: %v\nProject: %s\nCode Type: %s\nMatch Reason: %s\n\nContent:\n```\n%s\n```\n\n---\n\n",
			i+1, memory.RelevanceScore*100, memory.ID, memory.Context,
			memory.Keywords, memory.Tags, memory.ProjectPath,
			memory.CodeType, memory.MatchReason, memory.Content)
	}

	return &models.MCPToolResult{
		Content: []models.MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}, nil
}

// EvolveMemoryNetworkTool implements the evolve_memory_network MCP tool
type EvolveMemoryNetworkTool struct {
	evolutionMgr *EvolutionManager
	logger       *zap.Logger
}

// NewEvolveMemoryNetworkTool creates a new evolve memory network tool
func NewEvolveMemoryNetworkTool(evolutionMgr *EvolutionManager, logger *zap.Logger) *EvolveMemoryNetworkTool {
	return &EvolveMemoryNetworkTool{
		evolutionMgr: evolutionMgr,
		logger:       logger,
	}
}

func (t *EvolveMemoryNetworkTool) Name() string {
	return models.ToolEvolveMemoryNetwork
}

func (t *EvolveMemoryNetworkTool) Description() string {
	return "Trigger evolution of the memory network to identify patterns, optimize connections, and update memories"
}

func (t *EvolveMemoryNetworkTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"trigger_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of trigger: 'manual', 'scheduled', or 'event'",
				"default":     "manual",
			},
			"scope": map[string]interface{}{
				"type":        "string",
				"description": "Scope of evolution: 'recent', 'all', or 'project'",
				"default":     "recent",
			},
			"max_memories": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of memories to analyze (default: 100)",
				"default":     100,
			},
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Project path when scope is 'project'",
			},
		},
	}
}

func (t *EvolveMemoryNetworkTool) Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error) {
	// Parse arguments
	var req models.EvolveNetworkRequest

	req.TriggerType = "manual"
	if tt, ok := args["trigger_type"].(string); ok {
		req.TriggerType = tt
	}

	req.Scope = "recent"
	if s, ok := args["scope"].(string); ok {
		req.Scope = s
	}

	req.MaxMemories = 100
	if mm, ok := args["max_memories"].(float64); ok {
		req.MaxMemories = int(mm)
	}

	if projectPath, ok := args["project_path"].(string); ok {
		req.ProjectPath = projectPath
	}

	t.logger.Info("Evolution triggered",
		zap.String("trigger_type", req.TriggerType),
		zap.String("scope", req.Scope),
		zap.Int("max_memories", req.MaxMemories))

	// Execute evolution
	response, err := t.evolutionMgr.EvolveNetwork(ctx, req)
	if err != nil {
		t.logger.Error("Evolution failed", zap.Error(err))
		return &models.MCPToolResult{
			IsError: true,
			Content: []models.MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("Memory network evolution failed: %v", err),
			}},
		}, nil
	}

	// Format response
	resultText := fmt.Sprintf(`Memory network evolution completed!

Results:
- Memories Analyzed: %d
- Memories Evolved: %d
- Links Created: %d
- Links Strengthened: %d
- Contexts Updated: %d
- Duration: %d ms

The memory network has been analyzed and optimized. New connections have been identified and memory contexts have been improved based on AI analysis.`,
		response.MemoriesAnalyzed,
		response.MemoriesEvolved,
		response.LinksCreated,
		response.LinksStrengthened,
		response.ContextsUpdated,
		response.DurationMs)

	return &models.MCPToolResult{
		Content: []models.MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}, nil
}
