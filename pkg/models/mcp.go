package models

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPNotification represents a JSON-RPC 2.0 notification (no ID field)
type MCPNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPSuccessResponse represents a successful JSON-RPC 2.0 response
type MCPSuccessResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result"`
}

// MCPErrorResponse represents an error JSON-RPC 2.0 response
type MCPErrorResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   MCPError    `json:"error"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPTool represents an MCP tool definition
type MCPTool struct {
	Name             string                            `json:"name"`
	Description      string                            `json:"description"`
	InputSchema      map[string]interface{}            `json:"inputSchema"`
	UsageTriggers    []string                          `json:"usageTriggers,omitempty"`
	BestPractices    []string                          `json:"bestPractices,omitempty"`
	Synergies        map[string][]string               `json:"synergies,omitempty"`
	WorkflowSnippets []map[string]interface{}          `json:"workflowSnippets,omitempty"`
}

// MCPToolCall represents a tool call request
type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPToolResult represents a tool call result
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// MCPContent represents content in MCP responses
type MCPContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Standard MCP error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP method names
const (
	MethodInitialize   = "initialize"
	MethodListTools    = "tools/list"
	MethodCallTool     = "tools/call"
	MethodNotification = "notifications/initialized"
)

// Tool names
const (
	ToolStoreCodingMemory        = "store_coding_memory"
	ToolRetrieveRelevantMemories = "retrieve_relevant_memories"
	ToolEvolveMemoryNetwork      = "evolve_memory_network"
)
