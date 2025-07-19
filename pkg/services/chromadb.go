package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zetmem/mcp-server/pkg/config"
	"github.com/zetmem/mcp-server/pkg/models"
	"go.uber.org/zap"
)

// ChromaDBService handles vector database operations
type ChromaDBService struct {
	config       config.ChromaDBConfig
	logger       *zap.Logger
	httpClient   *http.Client
	baseURL      string
	collectionID string // Cache the collection UUID
}

// ChromaAddRequest represents a request to add documents to ChromaDB
type ChromaAddRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float32              `json:"embeddings"`
	Metadatas  []map[string]interface{} `json:"metadatas"`
	Documents  []string                 `json:"documents"`
}

// ChromaQueryRequest represents a query request to ChromaDB
type ChromaQueryRequest struct {
	QueryEmbeddings [][]float32            `json:"query_embeddings"`
	NResults        int                    `json:"n_results"`
	Include         []string               `json:"include"`
	Where           map[string]interface{} `json:"where,omitempty"`
}

// ChromaQueryResponse represents a query response from ChromaDB
type ChromaQueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float32                `json:"distances"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Documents [][]string                 `json:"documents"`
}

// NewChromaDBService creates a new ChromaDB service
func NewChromaDBService(cfg config.ChromaDBConfig, logger *zap.Logger) *ChromaDBService {
	return &ChromaDBService{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: cfg.URL,
	}
}

// getCollectionID gets the UUID for a collection by name
func (c *ChromaDBService) getCollectionID(ctx context.Context) (string, error) {
	if c.collectionID != "" {
		return c.collectionID, nil // Return cached ID
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, c.config.Collection), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ChromaDB get collection error: %d - %s", resp.StatusCode, string(body))
	}

	var collection struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return "", fmt.Errorf("failed to decode collection response: %w", err)
	}

	c.collectionID = collection.ID // Cache the ID
	return c.collectionID, nil
}

// Initialize initializes the ChromaDB collection
func (c *ChromaDBService) Initialize(ctx context.Context) error {
	// Create collection if it doesn't exist
	collectionData := map[string]interface{}{
		"name": c.config.Collection,
		"metadata": map[string]interface{}{
			"description": "ZetMem memory storage",
		},
	}

	requestBody, err := json.Marshal(collectionData)
	if err != nil {
		return fmt.Errorf("failed to marshal collection data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/v1/collections", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	// 409 means collection already exists, which is fine
	// 500 with UniqueConstraintError also means collection exists
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict {
		// Success or already exists
	} else if resp.StatusCode == http.StatusInternalServerError {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if strings.Contains(bodyStr, "already exists") {
			// Collection already exists, this is fine
		} else {
			return fmt.Errorf("ChromaDB API error: %d - %s", resp.StatusCode, bodyStr)
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ChromaDB API error: %d - %s", resp.StatusCode, string(body))
	}

	c.logger.Info("ChromaDB collection initialized",
		zap.String("collection", c.config.Collection))

	return nil
}

// StoreMemory stores a memory in ChromaDB
func (c *ChromaDBService) StoreMemory(ctx context.Context, memory *models.Memory) error {
	if len(memory.Embedding) == 0 {
		return fmt.Errorf("memory embedding is required")
	}

	// Get collection UUID
	collectionID, err := c.getCollectionID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get collection ID: %w", err)
	}

	// Prepare metadata
	metadata := map[string]interface{}{
		"context":      memory.Context,
		"keywords":     strings.Join(memory.Keywords, ","),
		"tags":         strings.Join(memory.Tags, ","),
		"project_path": memory.ProjectPath, // Keep for backward compatibility
		"workspace_id": memory.WorkspaceID,
		"code_type":    memory.CodeType,
		"created_at":   memory.CreatedAt.Unix(),
		"updated_at":   memory.UpdatedAt.Unix(),
	}

	// Add custom metadata
	for k, v := range memory.Metadata {
		metadata[k] = v
	}

	request := ChromaAddRequest{
		IDs:        []string{memory.ID},
		Embeddings: [][]float32{memory.Embedding},
		Metadatas:  []map[string]interface{}{metadata},
		Documents:  []string{memory.Content},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, collectionID),
		bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ChromaDB store error: %d - %s", resp.StatusCode, string(body))
	}

	c.logger.Debug("Memory stored in ChromaDB",
		zap.String("memory_id", memory.ID))

	return nil
}

// SearchSimilar searches for similar memories
func (c *ChromaDBService) SearchSimilar(ctx context.Context, queryEmbedding []float32, limit int, filters map[string]interface{}) ([]*models.Memory, []float32, error) {
	// Get collection UUID
	collectionID, err := c.getCollectionID(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get collection ID: %w", err)
	}

	request := ChromaQueryRequest{
		QueryEmbeddings: [][]float32{queryEmbedding},
		NResults:        limit,
		Include:         []string{"metadatas", "documents", "distances"},
	}

	if len(filters) > 0 {
		request.Where = filters
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, collectionID),
		bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query ChromaDB: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("ChromaDB query error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response ChromaQueryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.IDs) == 0 || len(response.IDs[0]) == 0 {
		return []*models.Memory{}, []float32{}, nil
	}

	memories := make([]*models.Memory, 0, len(response.IDs[0]))
	distances := response.Distances[0]

	for i, id := range response.IDs[0] {
		memory := &models.Memory{
			ID:      id,
			Content: response.Documents[0][i],
		}

		// Reconstruct memory from metadata
		if len(response.Metadatas) > 0 && len(response.Metadatas[0]) > i {
			metadata := response.Metadatas[0][i]

			if context, ok := metadata["context"].(string); ok {
				memory.Context = context
			}
			if keywords, ok := metadata["keywords"].(string); ok && keywords != "" {
				memory.Keywords = strings.Split(keywords, ",")
			}
			if tags, ok := metadata["tags"].(string); ok && tags != "" {
				memory.Tags = strings.Split(tags, ",")
			}
			if projectPath, ok := metadata["project_path"].(string); ok {
				memory.ProjectPath = projectPath
			}
			if workspaceID, ok := metadata["workspace_id"].(string); ok {
				memory.WorkspaceID = workspaceID
			}
			if codeType, ok := metadata["code_type"].(string); ok {
				memory.CodeType = codeType
			}
			if createdAt, ok := metadata["created_at"].(float64); ok {
				memory.CreatedAt = time.Unix(int64(createdAt), 0)
			}
			if updatedAt, ok := metadata["updated_at"].(float64); ok {
				memory.UpdatedAt = time.Unix(int64(updatedAt), 0)
			}

			memory.Metadata = metadata
		}

		memories = append(memories, memory)
	}

	c.logger.Debug("ChromaDB search completed",
		zap.Int("results", len(memories)))

	return memories, distances, nil
}
