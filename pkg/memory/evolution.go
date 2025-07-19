package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zetmem/mcp-server/pkg/models"
	"go.uber.org/zap"
)

// EvolutionManager handles memory network evolution
type EvolutionManager struct {
	system *System
	logger *zap.Logger
}

// NewEvolutionManager creates a new evolution manager
func NewEvolutionManager(system *System, logger *zap.Logger) *EvolutionManager {
	return &EvolutionManager{
		system: system,
		logger: logger,
	}
}

// EvolveNetwork evolves the memory network based on the given request
func (e *EvolutionManager) EvolveNetwork(ctx context.Context, req models.EvolveNetworkRequest) (*models.EvolveNetworkResponse, error) {
	e.logger.Info("Starting memory network evolution",
		zap.String("trigger_type", req.TriggerType),
		zap.String("scope", req.Scope),
		zap.Int("max_memories", req.MaxMemories))

	startTime := time.Now()

	// Step 1: Get memories to analyze based on scope
	memories, err := e.getMemoriesToAnalyze(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get memories to analyze: %w", err)
	}

	if len(memories) == 0 {
		e.logger.Info("No memories found for evolution")
		return &models.EvolveNetworkResponse{
			MemoriesAnalyzed:  0,
			MemoriesEvolved:   0,
			LinksCreated:      0,
			LinksStrengthened: 0,
			ContextsUpdated:   0,
			DurationMs:        int(time.Since(startTime).Milliseconds()),
		}, nil
	}

	e.logger.Info("Found memories for evolution", zap.Int("count", len(memories)))

	// Step 2: Analyze and evolve memories
	evolved := 0
	linksCreated := 0
	linksStrengthened := 0
	contextsUpdated := 0

	// Process memories in batches to avoid overwhelming the LLM
	batchSize := 10
	for i := 0; i < len(memories); i += batchSize {
		end := i + batchSize
		if end > len(memories) {
			end = len(memories)
		}

		batch := memories[i:end]
		batchEvolved, batchLinksCreated, batchLinksStrengthened, batchContextsUpdated, err := e.evolveBatch(ctx, batch)
		if err != nil {
			e.logger.Warn("Error evolving batch", zap.Error(err), zap.Int("batch_start", i))
			continue
		}

		evolved += batchEvolved
		linksCreated += batchLinksCreated
		linksStrengthened += batchLinksStrengthened
		contextsUpdated += batchContextsUpdated
	}

	duration := time.Since(startTime).Milliseconds()
	e.logger.Info("Memory network evolution completed",
		zap.Int("memories_analyzed", len(memories)),
		zap.Int("memories_evolved", evolved),
		zap.Int("links_created", linksCreated),
		zap.Int("links_strengthened", linksStrengthened),
		zap.Int("contexts_updated", contextsUpdated),
		zap.Int64("duration_ms", duration))

	return &models.EvolveNetworkResponse{
		MemoriesAnalyzed:  len(memories),
		MemoriesEvolved:   evolved,
		LinksCreated:      linksCreated,
		LinksStrengthened: linksStrengthened,
		ContextsUpdated:   contextsUpdated,
		DurationMs:        int(duration),
	}, nil
}

// getMemoriesToAnalyze retrieves memories based on scope
func (e *EvolutionManager) getMemoriesToAnalyze(ctx context.Context, req models.EvolveNetworkRequest) ([]*models.Memory, error) {
	// Build filters based on scope
	filters := make(map[string]interface{})

	if req.Scope == "project" && req.ProjectPath != "" {
		filters["project_path"] = req.ProjectPath
	}

	// For now, we'll use a simple approach to get memories
	// In a real implementation, this would be more sophisticated with time-based filtering
	// and better selection criteria

	// Use a dummy query to get recent memories
	dummyQuery := "recent memories"
	queryEmbedding, err := e.system.embeddingService.GenerateEmbedding(ctx, dummyQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Limit the number of memories to analyze
	limit := req.MaxMemories
	if limit <= 0 {
		limit = 100 // Default
	}

	memories, _, err := e.system.chromaDB.SearchSimilar(ctx, queryEmbedding, limit, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to search memories: %w", err)
	}

	return memories, nil
}

// evolveBatch evolves a batch of memories
func (e *EvolutionManager) evolveBatch(ctx context.Context, memories []*models.Memory) (int, int, int, int, error) {
	if len(memories) == 0 {
		return 0, 0, 0, 0, nil
	}

	// Step 1: Prepare context for LLM analysis
	analysisContext := e.prepareAnalysisContext(memories)

	// Step 2: Call LLM for analysis
	analysisResult, err := e.analyzeMemoryNetwork(ctx, analysisContext)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to analyze memory network: %w", err)
	}

	// Step 3: Apply evolution actions
	evolved := 0
	linksCreated := 0
	linksStrengthened := 0
	contextsUpdated := 0

	if !analysisResult.ShouldEvolve {
		e.logger.Info("LLM analysis suggests no evolution needed for this batch")
		return 0, 0, 0, 0, nil
	}

	// Apply context updates
	for memoryID, newContext := range analysisResult.ContextUpdates {
		if err := e.updateMemoryContext(ctx, memoryID, newContext); err != nil {
			e.logger.Warn("Failed to update memory context",
				zap.String("memory_id", memoryID),
				zap.Error(err))
			continue
		}
		contextsUpdated++
		evolved++
	}

	// Apply tag updates
	for memoryID, newTags := range analysisResult.TagUpdates {
		if err := e.updateMemoryTags(ctx, memoryID, newTags); err != nil {
			e.logger.Warn("Failed to update memory tags",
				zap.String("memory_id", memoryID),
				zap.Error(err))
			continue
		}
		evolved++
	}

	// Create new connections
	for _, link := range analysisResult.SuggestedConnections {
		if err := e.createMemoryLink(ctx, link); err != nil {
			e.logger.Warn("Failed to create memory link",
				zap.String("target_id", link.TargetID),
				zap.Error(err))
			continue
		}
		linksCreated++
	}

	return evolved, linksCreated, linksStrengthened, contextsUpdated, nil
}

// prepareAnalysisContext prepares the context for LLM analysis
func (e *EvolutionManager) prepareAnalysisContext(memories []*models.Memory) string {
	// Format memories for LLM analysis
	context := "Memory Network Analysis Context:\n\n"

	for i, memory := range memories {
		context += fmt.Sprintf("Memory %d:\n", i+1)
		context += fmt.Sprintf("ID: %s\n", memory.ID)
		context += fmt.Sprintf("Content: %s\n", memory.Content)
		context += fmt.Sprintf("Context: %s\n", memory.Context)
		context += fmt.Sprintf("Keywords: %v\n", memory.Keywords)
		context += fmt.Sprintf("Tags: %v\n", memory.Tags)
		context += fmt.Sprintf("Project Path: %s\n", memory.ProjectPath)
		context += fmt.Sprintf("Code Type: %s\n", memory.CodeType)

		if len(memory.Links) > 0 {
			context += "Links:\n"
			for _, link := range memory.Links {
				context += fmt.Sprintf("- Target: %s, Type: %s, Strength: %.2f, Reason: %s\n",
					link.TargetID, link.LinkType, link.Strength, link.Reason)
			}
		}

		context += "\n---\n\n"
	}

	return context
}

// analyzeMemoryNetwork calls LLM to analyze the memory network
func (e *EvolutionManager) analyzeMemoryNetwork(ctx context.Context, analysisContext string) (*models.EvolutionAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the following memory network and suggest evolution actions:

%s

Your task is to identify patterns, redundancies, and opportunities for improvement in this memory network.
Consider:
1. Memories that should have improved context descriptions
2. Memories that should be linked together
3. Tags that should be updated for better categorization

Respond with a JSON object in the following format:
{
  "should_evolve": true/false,
  "actions": ["action1", "action2", ...],
  "suggested_connections": [
    {"target_id": "memory_id", "link_type": "pattern|solution|technology", "strength": 0.8, "reason": "reason for connection"}
  ],
  "context_updates": {
    "memory_id": "improved context description"
  },
  "tag_updates": {
    "memory_id": ["tag1", "tag2", "tag3"]
  }
}

Only suggest changes if they would significantly improve the memory network.`, analysisContext)

	response, err := e.system.llmService.CallWithRetry(ctx, prompt, true)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var result models.EvolutionAnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &result, nil
}

// updateMemoryContext updates a memory's context
func (e *EvolutionManager) updateMemoryContext(ctx context.Context, memoryID, newContext string) error {
	// In a real implementation, this would update the memory in ChromaDB
	// For now, we'll just log it
	e.logger.Info("Would update memory context",
		zap.String("memory_id", memoryID),
		zap.String("new_context", newContext))

	// Placeholder for actual implementation
	return nil
}

// updateMemoryTags updates a memory's tags
func (e *EvolutionManager) updateMemoryTags(ctx context.Context, memoryID string, newTags []string) error {
	// In a real implementation, this would update the memory in ChromaDB
	// For now, we'll just log it
	e.logger.Info("Would update memory tags",
		zap.String("memory_id", memoryID),
		zap.Strings("new_tags", newTags))

	// Placeholder for actual implementation
	return nil
}

// createMemoryLink creates a new link between memories
func (e *EvolutionManager) createMemoryLink(ctx context.Context, link models.MemoryLink) error {
	// In a real implementation, this would create a link in ChromaDB
	// For now, we'll just log it
	e.logger.Info("Would create memory link",
		zap.String("target_id", link.TargetID),
		zap.String("link_type", link.LinkType),
		zap.Float32("strength", link.Strength),
		zap.String("reason", link.Reason))

	// Placeholder for actual implementation
	return nil
}
