package models

import (
	"testing"
	"time"
)

func TestMemoryCreation(t *testing.T) {
	memory := Memory{
		ID:          "test-id",
		Content:     "function test() { return 'hello'; }",
		Context:     "Simple JavaScript function",
		Keywords:    []string{"function", "javascript", "test"},
		Tags:        []string{"javascript", "function", "simple"},
		ProjectPath: "/test/project",
		CodeType:    "javascript",
		Embedding:   []float32{0.1, 0.2, 0.3},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	if memory.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", memory.ID)
	}

	if len(memory.Keywords) != 3 {
		t.Errorf("Expected 3 keywords, got %d", len(memory.Keywords))
	}

	if memory.CodeType != "javascript" {
		t.Errorf("Expected CodeType to be 'javascript', got %s", memory.CodeType)
	}
}

func TestMemoryLink(t *testing.T) {
	link := MemoryLink{
		TargetID: "target-id",
		LinkType: "solution",
		Strength: 0.85,
		Reason:   "Similar problem domain",
	}

	if link.TargetID != "target-id" {
		t.Errorf("Expected TargetID to be 'target-id', got %s", link.TargetID)
	}

	if link.Strength != 0.85 {
		t.Errorf("Expected Strength to be 0.85, got %f", link.Strength)
	}
}

func TestStoreMemoryRequest(t *testing.T) {
	req := StoreMemoryRequest{
		Content:     "test content",
		ProjectPath: "/test",
		CodeType:    "go",
		Context:     "test context",
	}

	if req.Content != "test content" {
		t.Errorf("Expected Content to be 'test content', got %s", req.Content)
	}

	if req.CodeType != "go" {
		t.Errorf("Expected CodeType to be 'go', got %s", req.CodeType)
	}
}

func TestRetrieveMemoryRequest(t *testing.T) {
	req := RetrieveMemoryRequest{
		Query:         "test query",
		MaxResults:    5,
		ProjectFilter: "/test",
		CodeTypes:     []string{"go", "javascript"},
		MinRelevance:  0.7,
	}

	if req.Query != "test query" {
		t.Errorf("Expected Query to be 'test query', got %s", req.Query)
	}

	if req.MaxResults != 5 {
		t.Errorf("Expected MaxResults to be 5, got %d", req.MaxResults)
	}

	if len(req.CodeTypes) != 2 {
		t.Errorf("Expected 2 CodeTypes, got %d", len(req.CodeTypes))
	}
}
