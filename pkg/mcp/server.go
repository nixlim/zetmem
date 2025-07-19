package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/zetmem/mcp-server/pkg/models"
	"go.uber.org/zap"
)

// Server represents the MCP server
type Server struct {
	logger      *zap.Logger
	tools       map[string]Tool
	initialized bool
	reader      *bufio.Reader
	writer      io.Writer
}

// Tool represents an MCP tool handler
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (*models.MCPToolResult, error)
}

// EnhancedTool represents an MCP tool with strategic guidance capabilities
type EnhancedTool interface {
	Tool
	UsageTriggers() []string
	BestPractices() []string
	Synergies() map[string][]string
	WorkflowSnippets() []map[string]interface{}
}

// NewServer creates a new MCP server
func NewServer(logger *zap.Logger) *Server {
	return &Server{
		logger: logger,
		tools:  make(map[string]Tool),
		reader: bufio.NewReader(os.Stdin),
		writer: os.Stdout,
	}
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool Tool) {
	s.tools[tool.Name()] = tool
	s.logger.Info("Registered tool", zap.String("name", tool.Name()))
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("MCP server shutting down")
			return ctx.Err()
		default:
			if err := s.handleRequest(ctx); err != nil {
				if err == io.EOF {
					s.logger.Info("Client disconnected")
					return nil
				}
				s.logger.Error("Error handling request", zap.Error(err))
				continue
			}
		}
	}
}

// handleRequest handles a single JSON-RPC request or notification
func (s *Server) handleRequest(ctx context.Context) error {
	line, err := s.reader.ReadString('\n')
	if err != nil {
		return err
	}

	// First try to parse as a request (with ID)
	var request models.MCPRequest
	if err := json.Unmarshal([]byte(line), &request); err == nil && request.ID != nil {
		s.logger.Debug("Received request",
			zap.String("method", request.Method),
			zap.Any("id", request.ID))
		return s.handleJSONRPCRequest(ctx, request)
	}

	// If that fails, try to parse as a notification (no ID)
	var notification models.MCPNotification
	if err := json.Unmarshal([]byte(line), &notification); err == nil {
		s.logger.Debug("Received notification",
			zap.String("method", notification.Method))
		return s.handleJSONRPCNotification(notification)
	}

	// If both fail, send error response (only for requests, not notifications)
	s.sendError(nil, models.ParseError, "Invalid JSON", nil)
	return nil
}

// handleJSONRPCRequest handles requests that require responses
func (s *Server) handleJSONRPCRequest(ctx context.Context, request models.MCPRequest) error {
	switch request.Method {
	case models.MethodInitialize:
		return s.handleInitialize(request)
	case models.MethodListTools:
		return s.handleListTools(request)
	case models.MethodCallTool:
		return s.handleCallTool(ctx, request)
	default:
		s.sendError(request.ID, models.MethodNotFound,
			fmt.Sprintf("Method not found: %s", request.Method), nil)
	}

	return nil
}

// handleJSONRPCNotification handles notifications that don't require responses
func (s *Server) handleJSONRPCNotification(notification models.MCPNotification) error {
	switch notification.Method {
	case "notifications/initialized":
		// Client has finished initialization - no response needed
		s.logger.Debug("Client initialization complete")
		return nil
	default:
		// Unknown notification - log but don't respond
		s.logger.Debug("Unknown notification received",
			zap.String("method", notification.Method))
		return nil
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(request models.MCPRequest) error {
	s.initialized = true

	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "ZetMem MCP Server",
			"version": "1.0.0",
		},
	}

	return s.sendResponse(request.ID, result)
}

// handleListTools handles the list tools request
func (s *Server) handleListTools(request models.MCPRequest) error {
	if !s.initialized {
		s.sendError(request.ID, models.InvalidRequest, "Server not initialized", nil)
		return nil
	}

	tools := make([]models.MCPTool, 0, len(s.tools))
	for _, tool := range s.tools {
		mcpTool := models.MCPTool{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		}

		// Check if tool implements EnhancedTool interface
		if enhancedTool, ok := tool.(EnhancedTool); ok {
			mcpTool.UsageTriggers = enhancedTool.UsageTriggers()
			mcpTool.BestPractices = enhancedTool.BestPractices()
			mcpTool.Synergies = enhancedTool.Synergies()
			mcpTool.WorkflowSnippets = enhancedTool.WorkflowSnippets()
		}

		tools = append(tools, mcpTool)
	}

	result := map[string]interface{}{
		"tools": tools,
		"strategyGuideSummary": "This server follows the Zetmem strategic principles for AI collaboration. Key concepts include workspace-first organization, consistent memory management habits, and iterative development patterns. For comprehensive onboarding guidance, see the Zetmem Onboarding Strategy document at: ZETMEM_ONBOARDING_STRATEGY.md",
	}

	return s.sendResponse(request.ID, result)
}

// handleCallTool handles tool call requests
func (s *Server) handleCallTool(ctx context.Context, request models.MCPRequest) error {
	if !s.initialized {
		s.sendError(request.ID, models.InvalidRequest, "Server not initialized", nil)
		return nil
	}

	params, ok := request.Params.(map[string]interface{})
	if !ok {
		s.sendError(request.ID, models.InvalidParams, "Invalid params", nil)
		return nil
	}

	toolName, ok := params["name"].(string)
	if !ok {
		s.sendError(request.ID, models.InvalidParams, "Tool name required", nil)
		return nil
	}

	tool, exists := s.tools[toolName]
	if !exists {
		s.sendError(request.ID, models.MethodNotFound,
			fmt.Sprintf("Tool not found: %s", toolName), nil)
		return nil
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	s.logger.Info("Executing tool",
		zap.String("tool", toolName),
		zap.Any("arguments", arguments))

	result, err := tool.Execute(ctx, arguments)
	if err != nil {
		s.logger.Error("Tool execution failed",
			zap.String("tool", toolName),
			zap.Error(err))
		s.sendError(request.ID, models.InternalError, err.Error(), nil)
		return nil
	}

	return s.sendResponse(request.ID, result)
}

// sendResponse sends a JSON-RPC response
func (s *Server) sendResponse(id interface{}, result interface{}) error {
	response := models.MCPSuccessResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	return err
}

// sendError sends a JSON-RPC error response
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) error {
	response := models.MCPErrorResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: models.MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "%s\n", responseData)
	return err
}
