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

	"github.com/amem/mcp-server/pkg/config"
	"go.uber.org/zap"
)

// LiteLLMService handles LLM interactions via LiteLLM proxy
type LiteLLMService struct {
	config     config.LiteLLMConfig
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
}

// LiteLLMRequest represents a request to LiteLLM
type LiteLLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LiteLLMResponse represents a response from LiteLLM
type LiteLLMResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewLiteLLMService creates a new LiteLLM service
func NewLiteLLMService(cfg config.LiteLLMConfig, logger *zap.Logger) *LiteLLMService {
	return &LiteLLMService{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL: "https://api.openai.com/v1", // OpenAI API URL
	}
}

// CallWithRetry calls LiteLLM with retry logic
func (s *LiteLLMService) CallWithRetry(ctx context.Context, prompt string, retryOnJSON bool) (string, error) {
	var lastErr error

	for i := 0; i < s.config.MaxRetries; i++ {
		response, err := s.call(ctx, prompt, s.config.DefaultModel)
		if err != nil {
			lastErr = err
			s.logger.Warn("LiteLLM call failed, retrying",
				zap.Int("attempt", i+1),
				zap.Error(err))

			// Exponential backoff
			if i < s.config.MaxRetries-1 {
				time.Sleep(time.Second * time.Duration(1<<i))
			}
			continue
		}

		// Validate JSON if required
		if retryOnJSON {
			var test json.RawMessage
			if err := json.Unmarshal([]byte(response), &test); err != nil {
				lastErr = fmt.Errorf("invalid JSON response: %w", err)
				s.logger.Warn("Invalid JSON response, retrying",
					zap.Int("attempt", i+1),
					zap.String("response", response))
				continue
			}
		}

		return response, nil
	}

	// Try fallback models
	for _, model := range s.config.FallbackModels {
		s.logger.Info("Trying fallback model", zap.String("model", model))
		response, err := s.call(ctx, prompt, model)
		if err != nil {
			s.logger.Warn("Fallback model failed",
				zap.String("model", model),
				zap.Error(err))
			continue
		}

		if retryOnJSON {
			var test json.RawMessage
			if err := json.Unmarshal([]byte(response), &test); err != nil {
				s.logger.Warn("Fallback model returned invalid JSON",
					zap.String("model", model))
				continue
			}
		}

		return response, nil
	}

	return "", fmt.Errorf("all retries and fallbacks failed: %w", lastErr)
}

// call makes a single call to LiteLLM
func (s *LiteLLMService) call(ctx context.Context, prompt, model string) (string, error) {
	request := LiteLLMRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.1,
		MaxTokens:   1000,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		s.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add OpenAI API key from environment
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	} else {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LiteLLM API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var response LiteLLMResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	s.logger.Debug("LiteLLM call successful",
		zap.String("model", model),
		zap.Int("prompt_tokens", response.Usage.PromptTokens),
		zap.Int("completion_tokens", response.Usage.CompletionTokens))

	return response.Choices[0].Message.Content, nil
}

// GenerateEmbedding generates an embedding for the given text
func (s *LiteLLMService) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// For now, we'll use a simple hash-based embedding
	// In production, this should call an actual embedding service
	s.logger.Debug("Generating embedding", zap.String("text", text[:min(100, len(text))]))

	// Simple hash-based embedding (placeholder)
	embedding := make([]float32, 384) // Common embedding dimension
	hash := simpleHash(text)

	for i := range embedding {
		embedding[i] = float32((hash >> (i % 32)) & 1)
	}

	return embedding, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
