package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zetmem/mcp-server/pkg/models"
	"go.uber.org/zap"
)

// WorkspaceService handles workspace management operations
type WorkspaceService struct {
	logger   *zap.Logger
	chromaDB *ChromaDBService
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(chromaDB *ChromaDBService, logger *zap.Logger) *WorkspaceService {
	return &WorkspaceService{
		logger:   logger,
		chromaDB: chromaDB,
	}
}

// ValidateWorkspaceID validates a workspace identifier
func (w *WorkspaceService) ValidateWorkspaceID(id string) error {
	if id == "" {
		return fmt.Errorf("workspace ID cannot be empty")
	}

	// Check for invalid characters
	if strings.Contains(id, "\n") || strings.Contains(id, "\r") {
		return fmt.Errorf("workspace ID cannot contain newline characters")
	}

	// Allow both filesystem paths and logical names
	// Paths can be absolute or relative
	// Names should be alphanumeric with common separators
	validName := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !validName.MatchString(id) {
		return fmt.Errorf("workspace ID contains invalid characters")
	}

	return nil
}

// IsFilesystemPath determines if the identifier is a filesystem path
func (w *WorkspaceService) IsFilesystemPath(id string) bool {
	// Check for absolute paths
	if filepath.IsAbs(id) {
		return true
	}

	// Check for relative paths with directory separators
	if strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return true
	}

	// Check for common path patterns
	if strings.HasPrefix(id, "./") || strings.HasPrefix(id, "../") {
		return true
	}

	return false
}

// NormalizeWorkspaceID normalizes a workspace identifier
func (w *WorkspaceService) NormalizeWorkspaceID(id string) string {
	if id == "" {
		return "default"
	}

	// If it's a filesystem path, clean it
	if w.IsFilesystemPath(id) {
		return filepath.Clean(id)
	}

	// For logical names, just trim whitespace and convert to lowercase
	return strings.ToLower(strings.TrimSpace(id))
}

// GetDefaultWorkspaceID returns the default workspace ID
func (w *WorkspaceService) GetDefaultWorkspaceID() string {
	// Try to get current working directory
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}

	// Fallback to "default"
	return "default"
}

// WorkspaceExists checks if a workspace exists by looking for memories with that workspace_id
func (w *WorkspaceService) WorkspaceExists(ctx context.Context, workspaceID string) (bool, error) {
	// Query ChromaDB for any memories with this workspace_id
	filters := map[string]interface{}{
		"workspace_id": workspaceID,
	}

	// Use a dummy embedding for the query (we only care about metadata)
	dummyEmbedding := make([]float32, 384) // Match the embedding dimension
	memories, _, err := w.chromaDB.SearchSimilar(ctx, dummyEmbedding, 1, filters)
	if err != nil {
		return false, fmt.Errorf("failed to check workspace existence: %w", err)
	}

	return len(memories) > 0, nil
}

// GetWorkspaceInfo retrieves information about a workspace
func (w *WorkspaceService) GetWorkspaceInfo(ctx context.Context, workspaceID string) (*models.Workspace, error) {
	// Query ChromaDB for memories in this workspace to get count and dates
	filters := map[string]interface{}{
		"workspace_id": workspaceID,
	}

	// Use a dummy embedding for the query (we only care about metadata)
	dummyEmbedding := make([]float32, 384)
	memories, _, err := w.chromaDB.SearchSimilar(ctx, dummyEmbedding, 1000, filters) // Get many to count
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace info: %w", err)
	}

	workspace := &models.Workspace{
		ID:          workspaceID,
		Name:        w.generateWorkspaceName(workspaceID),
		Description: w.generateWorkspaceDescription(workspaceID),
		MemoryCount: len(memories),
		CreatedAt:   time.Now(), // Will be updated if we find memories
		UpdatedAt:   time.Now(),
	}

	// If we have memories, use the oldest for CreatedAt and newest for UpdatedAt
	if len(memories) > 0 {
		oldest := memories[0].CreatedAt
		newest := memories[0].UpdatedAt

		for _, memory := range memories {
			if memory.CreatedAt.Before(oldest) {
				oldest = memory.CreatedAt
			}
			if memory.UpdatedAt.After(newest) {
				newest = memory.UpdatedAt
			}
		}

		workspace.CreatedAt = oldest
		workspace.UpdatedAt = newest
	}

	return workspace, nil
}

// generateWorkspaceName generates a human-readable name for a workspace
func (w *WorkspaceService) generateWorkspaceName(workspaceID string) string {
	if workspaceID == "default" {
		return "Default Workspace"
	}

	if w.IsFilesystemPath(workspaceID) {
		return fmt.Sprintf("Project: %s", filepath.Base(workspaceID))
	}

	// For logical names, capitalize and format nicely
	name := strings.ReplaceAll(workspaceID, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return strings.Title(name)
}

// generateWorkspaceDescription generates a description for a workspace
func (w *WorkspaceService) generateWorkspaceDescription(workspaceID string) string {
	if workspaceID == "default" {
		return "Default workspace for memories without specific project assignment"
	}

	if w.IsFilesystemPath(workspaceID) {
		return fmt.Sprintf("Workspace for project at %s", workspaceID)
	}

	return fmt.Sprintf("Logical workspace: %s", workspaceID)
}

// CreateWorkspace creates a new workspace (explicit creation)
func (w *WorkspaceService) CreateWorkspace(ctx context.Context, req *models.WorkspaceRequest) (*models.Workspace, error) {
	workspaceID := w.NormalizeWorkspaceID(req.Identifier)

	if err := w.ValidateWorkspaceID(workspaceID); err != nil {
		return nil, fmt.Errorf("invalid workspace ID: %w", err)
	}

	// Check if workspace already exists
	exists, err := w.WorkspaceExists(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("workspace '%s' already exists", workspaceID)
	}

	// Create workspace info
	workspace := &models.Workspace{
		ID:          workspaceID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		MemoryCount: 0,
	}

	// Use generated name/description if not provided
	if workspace.Name == "" {
		workspace.Name = w.generateWorkspaceName(workspaceID)
	}
	if workspace.Description == "" {
		workspace.Description = w.generateWorkspaceDescription(workspaceID)
	}

	w.logger.Info("Created new workspace",
		zap.String("workspace_id", workspaceID),
		zap.String("name", workspace.Name))

	return workspace, nil
}

// InitializeWorkspace smart initialization - creates if not exists, retrieves if exists
func (w *WorkspaceService) InitializeWorkspace(ctx context.Context, req *models.WorkspaceRequest) (*models.Workspace, bool, error) {
	workspaceID := req.Identifier
	if workspaceID == "" {
		workspaceID = w.GetDefaultWorkspaceID()
	}

	workspaceID = w.NormalizeWorkspaceID(workspaceID)

	if err := w.ValidateWorkspaceID(workspaceID); err != nil {
		return nil, false, fmt.Errorf("invalid workspace ID: %w", err)
	}

	// Check if workspace exists
	exists, err := w.WorkspaceExists(ctx, workspaceID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check workspace existence: %w", err)
	}

	if exists {
		// Retrieve existing workspace
		workspace, err := w.GetWorkspaceInfo(ctx, workspaceID)
		if err != nil {
			return nil, false, fmt.Errorf("failed to retrieve workspace info: %w", err)
		}
		return workspace, false, nil
	}

	// Create new workspace
	createReq := &models.WorkspaceRequest{
		Identifier:  workspaceID,
		Name:        req.Name,
		Description: req.Description,
	}

	workspace, err := w.CreateWorkspace(ctx, createReq)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create workspace: %w", err)
	}

	return workspace, true, nil
}
