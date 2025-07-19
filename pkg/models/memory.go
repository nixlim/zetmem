package models

import (
	"time"
)

// Memory represents a stored coding memory with embeddings and links
type Memory struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Context     string                 `json:"context"`
	Keywords    []string               `json:"keywords"`
	Tags        []string               `json:"tags"`
	ProjectPath string                 `json:"project_path"` // Deprecated: use WorkspaceID
	WorkspaceID string                 `json:"workspace_id"`
	CodeType    string                 `json:"code_type"`
	Embedding   []float32              `json:"embedding"`
	Links       []MemoryLink           `json:"links"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// MemoryLink represents a connection between memories
type MemoryLink struct {
	TargetID string  `json:"target_id"`
	LinkType string  `json:"link_type"` // solution|pattern|technology|debugging|progression
	Strength float32 `json:"strength"`  // 0.0-1.0
	Reason   string  `json:"reason"`
}

// StoreMemoryRequest represents the request to store a new memory
type StoreMemoryRequest struct {
	Content     string `json:"content" validate:"required"`
	ProjectPath string `json:"project_path"` // Deprecated: use WorkspaceID
	WorkspaceID string `json:"workspace_id"`
	CodeType    string `json:"code_type"`
	Context     string `json:"context"`
}

// StoreMemoryResponse represents the response after storing a memory
type StoreMemoryResponse struct {
	MemoryID     string   `json:"memory_id"`
	Keywords     []string `json:"keywords"`
	Tags         []string `json:"tags"`
	LinksCreated int      `json:"links_created"`
	EventEmitted bool     `json:"event_emitted"`
}

// RetrieveMemoryRequest represents the request to retrieve memories
type RetrieveMemoryRequest struct {
	Query         string   `json:"query" validate:"required"`
	MaxResults    int      `json:"max_results"`
	ProjectFilter string   `json:"project_filter"` // Deprecated: use WorkspaceID
	WorkspaceID   string   `json:"workspace_id"`
	CodeTypes     []string `json:"code_types"`
	MinRelevance  float32  `json:"min_relevance"`
}

// RetrieveMemoryResponse represents the response with retrieved memories
type RetrieveMemoryResponse struct {
	Memories   []RetrievedMemory `json:"memories"`
	TotalFound int               `json:"total_found"`
}

// RetrievedMemory extends Memory with relevance information
type RetrievedMemory struct {
	Memory
	RelevanceScore float32 `json:"relevance_score"`
	MatchReason    string  `json:"match_reason"`
}

// EvolveNetworkRequest represents the request to evolve memory network
type EvolveNetworkRequest struct {
	TriggerType string `json:"trigger_type"` // manual|scheduled|event
	Scope       string `json:"scope"`        // recent|all|project
	MaxMemories int    `json:"max_memories"`
	ProjectPath string `json:"project_path"`
}

// EvolveNetworkResponse represents the response after network evolution
type EvolveNetworkResponse struct {
	MemoriesAnalyzed  int `json:"memories_analyzed"`
	MemoriesEvolved   int `json:"memories_evolved"`
	LinksCreated      int `json:"links_created"`
	LinksStrengthened int `json:"links_strengthened"`
	ContextsUpdated   int `json:"contexts_updated"`
	DurationMs        int `json:"duration_ms"`
}

// NoteConstructionResult represents the result of LLM-based note construction
type NoteConstructionResult struct {
	Keywords []string `json:"keywords"`
	Context  string   `json:"context"`
	Tags     []string `json:"tags"`
}

// EvolutionAnalysisResult represents the result of memory evolution analysis
type EvolutionAnalysisResult struct {
	ShouldEvolve         bool                   `json:"should_evolve"`
	Actions              []string               `json:"actions"`
	SuggestedConnections []MemoryLink           `json:"suggested_connections"`
	ContextUpdates       map[string]string      `json:"context_updates"`
	TagUpdates           map[string][]string    `json:"tag_updates"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// Workspace represents a logical grouping of memories
type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MemoryCount int       `json:"memory_count"`
}

// WorkspaceRequest represents a request to create or retrieve a workspace
type WorkspaceRequest struct {
	Identifier  string `json:"identifier"`  // Path or name for the workspace
	Name        string `json:"name"`        // Human-readable name (optional)
	Description string `json:"description"` // Workspace description (optional)
}

// WorkspaceResponse represents the response after workspace operations
type WorkspaceResponse struct {
	Workspace Workspace `json:"workspace"`
	Created   bool      `json:"created"` // True if workspace was created, false if retrieved
}
