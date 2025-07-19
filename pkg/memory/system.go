package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zetmem/mcp-server/pkg/models"
	"github.com/zetmem/mcp-server/pkg/services"
	"go.uber.org/zap"
)

// System represents the core memory management system
type System struct {
	logger           *zap.Logger
	llmService       *services.LiteLLMService
	chromaDB         *services.ChromaDBService
	embeddingService *services.EmbeddingService
	workspaceService *services.WorkspaceService
}

// NewSystem creates a new memory system
func NewSystem(logger *zap.Logger, llmService *services.LiteLLMService, chromaDB *services.ChromaDBService, embeddingService *services.EmbeddingService, workspaceService *services.WorkspaceService) *System {
	return &System{
		logger:           logger,
		llmService:       llmService,
		chromaDB:         chromaDB,
		embeddingService: embeddingService,
		workspaceService: workspaceService,
	}
}

// CreateMemory creates a new memory from the given content
func (s *System) CreateMemory(ctx context.Context, req models.StoreMemoryRequest) (*models.StoreMemoryResponse, error) {
	// Determine workspace ID (with backward compatibility)
	workspaceID := req.WorkspaceID
	if workspaceID == "" && req.ProjectPath != "" {
		// Backward compatibility: use project_path as workspace_id
		workspaceID = req.ProjectPath
	}
	if workspaceID == "" {
		// Use default workspace
		workspaceID = s.workspaceService.GetDefaultWorkspaceID()
	}

	// Normalize and validate workspace ID
	workspaceID = s.workspaceService.NormalizeWorkspaceID(workspaceID)
	if err := s.workspaceService.ValidateWorkspaceID(workspaceID); err != nil {
		return nil, fmt.Errorf("invalid workspace ID: %w", err)
	}

	s.logger.Info("Creating memory",
		zap.String("workspace_id", workspaceID),
		zap.String("project_path", req.ProjectPath),
		zap.String("code_type", req.CodeType))

	// Generate unique ID
	memoryID := uuid.New().String()

	// Step 1: Construct note using LLM
	noteResult, err := s.constructNote(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to construct note: %w", err)
	}

	// Step 2: Generate embedding
	embedding, err := s.embeddingService.GenerateEmbedding(ctx, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Step 3: Create memory object
	memory := &models.Memory{
		ID:          memoryID,
		Content:     req.Content,
		Context:     noteResult.Context,
		Keywords:    noteResult.Keywords,
		Tags:        noteResult.Tags,
		ProjectPath: req.ProjectPath, // Keep for backward compatibility
		WorkspaceID: workspaceID,
		CodeType:    req.CodeType,
		Embedding:   embedding,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	// Step 4: Generate links to existing memories
	links, err := s.generateLinks(ctx, memory)
	if err != nil {
		s.logger.Warn("Failed to generate links", zap.Error(err))
		// Continue without links rather than failing
	}
	memory.Links = links

	// Step 5: Store in ChromaDB
	if err := s.chromaDB.StoreMemory(ctx, memory); err != nil {
		return nil, fmt.Errorf("failed to store memory: %w", err)
	}

	s.logger.Info("Memory created successfully",
		zap.String("memory_id", memoryID),
		zap.Int("links_created", len(links)))

	return &models.StoreMemoryResponse{
		MemoryID:     memoryID,
		Keywords:     noteResult.Keywords,
		Tags:         noteResult.Tags,
		LinksCreated: len(links),
		EventEmitted: true,
	}, nil
}

// RetrieveMemories retrieves relevant memories based on query
func (s *System) RetrieveMemories(ctx context.Context, req models.RetrieveMemoryRequest) (*models.RetrieveMemoryResponse, error) {
	// Determine workspace ID (with backward compatibility)
	workspaceID := req.WorkspaceID
	if workspaceID == "" && req.ProjectFilter != "" {
		// Backward compatibility: use project_filter as workspace_id
		workspaceID = req.ProjectFilter
	}
	if workspaceID == "" {
		// Use default workspace
		workspaceID = s.workspaceService.GetDefaultWorkspaceID()
	}

	// Normalize workspace ID
	workspaceID = s.workspaceService.NormalizeWorkspaceID(workspaceID)

	s.logger.Info("Retrieving memories",
		zap.String("query", req.Query),
		zap.String("workspace_id", workspaceID),
		zap.Int("max_results", req.MaxResults))

	// Set defaults
	if req.MaxResults <= 0 {
		req.MaxResults = 5
	}
	if req.MinRelevance <= 0 {
		req.MinRelevance = 0.3
	}

	// Step 1: Generate query embedding
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Step 2: Build filters with proper ChromaDB query structure
	var conditions []map[string]interface{}

	// Create workspace filter with OR logic for backward compatibility
	workspaceFilter := []map[string]interface{}{
		{"workspace_id": workspaceID},
	}

	// Add project_path to OR clause if specified (backward compatibility)
	if req.ProjectFilter != "" && req.ProjectFilter != workspaceID {
		workspaceFilter = append(workspaceFilter, map[string]interface{}{
			"project_path": req.ProjectFilter,
		})
	}

	// Add workspace condition
	if len(workspaceFilter) > 1 {
		conditions = append(conditions, map[string]interface{}{
			"$or": workspaceFilter,
		})
	} else {
		conditions = append(conditions, map[string]interface{}{
			"workspace_id": workspaceID,
		})
	}

	// Add code type filter
	if len(req.CodeTypes) > 0 {
		conditions = append(conditions, map[string]interface{}{
			"code_type": map[string]interface{}{
				"$in": req.CodeTypes,
			},
		})
	}

	// Build final filter structure
	var filters map[string]interface{}
	if len(conditions) == 1 {
		// Single condition - use it directly
		filters = conditions[0]
	} else if len(conditions) > 1 {
		// Multiple conditions - wrap in $and
		filters = map[string]interface{}{
			"$and": conditions,
		}
	} else {
		// No conditions
		filters = make(map[string]interface{})
	}

	// Step 3: Search similar memories
	memories, distances, err := s.chromaDB.SearchSimilar(ctx, queryEmbedding, req.MaxResults*2, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	// Step 4: Rank and filter results
	retrievedMemories := make([]models.RetrievedMemory, 0, req.MaxResults)
	for i, memory := range memories {
		if i >= len(distances) {
			break
		}

		// Convert distance to similarity score (inverse relationship for L2 distance)
		relevanceScore := 1.0 / (1.0 + distances[i])

		if relevanceScore < req.MinRelevance {
			continue
		}

		retrievedMemory := models.RetrievedMemory{
			Memory:         *memory,
			RelevanceScore: relevanceScore,
			MatchReason:    s.generateMatchReason(req.Query, memory),
		}

		retrievedMemories = append(retrievedMemories, retrievedMemory)

		if len(retrievedMemories) >= req.MaxResults {
			break
		}
	}

	s.logger.Info("Memory retrieval completed",
		zap.Int("total_found", len(retrievedMemories)))

	return &models.RetrieveMemoryResponse{
		Memories:   retrievedMemories,
		TotalFound: len(retrievedMemories),
	}, nil
}

// constructNote uses LLM to analyze content and extract structured information
func (s *System) constructNote(ctx context.Context, req models.StoreMemoryRequest) (*models.NoteConstructionResult, error) {
	prompt := fmt.Sprintf(`Generate a structured analysis of the following coding content by:
1. Identifying the most salient keywords (focus on technical terms, functions, concepts)
2. Extracting core programming themes and contextual elements
3. Creating relevant categorical tags for coding classification

For coding context, consider:
- Programming language and frameworks used
- Problem domain (web dev, algorithms, data structures, etc.)
- Solution patterns and techniques
- Error types and debugging context
- Libraries and dependencies mentioned

Format the response as a JSON object:
{
  "keywords": [// 3-7 specific technical keywords, ordered by importance],
  "context": // one sentence summarizing the coding problem/solution/concept,
  "tags": [// 3-6 broad categories: language, domain, pattern type, difficulty]
}

Content for analysis: %s
Project Path: %s
Code Type: %s`, req.Content, req.ProjectPath, req.CodeType)

	response, err := s.llmService.CallWithRetry(ctx, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var result models.NoteConstructionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Validate and set defaults
	if len(result.Keywords) == 0 {
		result.Keywords = []string{"code", "programming"}
	}
	if result.Context == "" {
		result.Context = "Code snippet or programming concept"
	}
	if len(result.Tags) == 0 {
		result.Tags = []string{"general", "code"}
	}

	return &result, nil
}

// generateLinks creates links between the new memory and existing similar memories
func (s *System) generateLinks(ctx context.Context, memory *models.Memory) ([]models.MemoryLink, error) {
	// Search for similar memories
	similarMemories, distances, err := s.chromaDB.SearchSimilar(ctx, memory.Embedding, 10, nil)
	if err != nil {
		return nil, err
	}

	links := make([]models.MemoryLink, 0)
	for i, similarMemory := range similarMemories {
		if similarMemory.ID == memory.ID {
			continue // Skip self
		}

		if i >= len(distances) {
			break
		}

		// Convert distance to similarity
		similarity := 1.0 - distances[i]

		if similarity > 0.7 { // Threshold for creating links
			linkType := s.determineLinkType(memory, similarMemory)
			reason := s.generateLinkReason(memory, similarMemory, similarity)

			link := models.MemoryLink{
				TargetID: similarMemory.ID,
				LinkType: linkType,
				Strength: similarity,
				Reason:   reason,
			}

			links = append(links, link)
		}
	}

	return links, nil
}

// determineLinkType determines the type of link between two memories
func (s *System) determineLinkType(memory1, memory2 *models.Memory) string {
	// Simple heuristics for link type determination
	if memory1.CodeType == memory2.CodeType {
		return "technology"
	}

	// Check for common keywords
	for _, keyword1 := range memory1.Keywords {
		for _, keyword2 := range memory2.Keywords {
			if keyword1 == keyword2 {
				return "pattern"
			}
		}
	}

	return "solution"
}

// generateLinkReason generates a human-readable reason for the link
func (s *System) generateLinkReason(memory1, memory2 *models.Memory, similarity float32) string {
	return fmt.Sprintf("Similar content with %.1f%% relevance", similarity*100)
}

// generateMatchReason generates a reason why a memory matched the query
func (s *System) generateMatchReason(query string, memory *models.Memory) string {
	// Simple keyword matching for now
	for _, keyword := range memory.Keywords {
		if contains(query, keyword) {
			return fmt.Sprintf("Keyword match: %s", keyword)
		}
	}
	return "Content similarity match"
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
