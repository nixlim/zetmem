package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/zetmem/mcp-server/pkg/config"
	"go.uber.org/zap"
)

// EmbeddingService handles text embedding generation
type EmbeddingService struct {
	config     config.EmbeddingConfig
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
}

// EmbeddingRequest represents a request to the embedding service
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse represents a response from the embedding service
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// SentenceTransformersRequest for local sentence-transformers service
type SentenceTransformersRequest struct {
	Sentences []string `json:"sentences"`
	Model     string   `json:"model,omitempty"`
}

// SentenceTransformersResponse from local sentence-transformers service
type SentenceTransformersResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(cfg config.EmbeddingConfig, logger *zap.Logger) *EmbeddingService {
	baseURL := cfg.URL // Use configured URL
	if cfg.Service == "openai" {
		baseURL = "https://api.openai.com/v1"
	}
	// Environment variable can override config
	if url := getEnvString("EMBEDDING_SERVICE_URL", ""); url != "" {
		baseURL = url
	}

	return &EmbeddingService{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GenerateEmbedding generates an embedding for the given text
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	switch s.config.Service {
	case "openai":
		return s.generateOpenAIEmbedding(ctx, text)
	case "sentence-transformers":
		return s.generateSentenceTransformersEmbedding(ctx, text)
	default:
		return s.generateFallbackEmbedding(text), nil
	}
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
func (s *EmbeddingService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	switch s.config.Service {
	case "sentence-transformers":
		return s.generateBatchSentenceTransformersEmbeddings(ctx, texts)
	default:
		// Fallback to individual calls
		embeddings := make([][]float32, len(texts))
		for i, text := range texts {
			embedding, err := s.GenerateEmbedding(ctx, text)
			if err != nil {
				return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
			}
			embeddings[i] = embedding
		}
		return embeddings, nil
	}
}

// generateOpenAIEmbedding generates embedding using OpenAI API
func (s *EmbeddingService) generateOpenAIEmbedding(ctx context.Context, text string) ([]float32, error) {
	request := EmbeddingRequest{
		Input: text,
		Model: "text-embedding-ada-002", // Default OpenAI embedding model
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		s.baseURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getOpenAIKey())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response EmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embeddings in response")
	}

	s.logger.Debug("OpenAI embedding generated",
		zap.Int("prompt_tokens", response.Usage.PromptTokens),
		zap.Int("embedding_dim", len(response.Data[0].Embedding)))

	return response.Data[0].Embedding, nil
}

// generateSentenceTransformersEmbedding generates embedding using sentence-transformers
func (s *EmbeddingService) generateSentenceTransformersEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.generateBatchSentenceTransformersEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// generateBatchSentenceTransformersEmbeddings generates batch embeddings using sentence-transformers
func (s *EmbeddingService) generateBatchSentenceTransformersEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	request := SentenceTransformersRequest{
		Sentences: texts,
		Model:     s.config.Model,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		s.baseURL+"/embeddings", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Sentence-Transformers API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response SentenceTransformersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	s.logger.Debug("Sentence-Transformers embeddings generated",
		zap.Int("count", len(response.Embeddings)),
		zap.Int("embedding_dim", len(response.Embeddings[0])))

	return response.Embeddings, nil
}

// generateFallbackEmbedding generates a simple hash-based embedding as fallback
func (s *EmbeddingService) generateFallbackEmbedding(text string) []float32 {
	s.logger.Warn("Using fallback embedding generation")

	// Simple hash-based embedding (same as before)
	embedding := make([]float32, 384) // Common embedding dimension
	hash := simpleHash(text)

	for i := range embedding {
		embedding[i] = float32((hash >> (i % 32)) & 1)
	}

	return embedding
}

// getOpenAIKey retrieves OpenAI API key from environment
func getOpenAIKey() string {
	// This should be injected via configuration in production
	return "your-openai-key" // Placeholder
}

// simpleHash creates a simple hash of the input string
func simpleHash(s string) uint64 {
	var hash uint64 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}

// getEnvString retrieves environment variable with default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
