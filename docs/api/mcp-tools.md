# MCP Tools API Reference

## Overview

The A-MEM MCP server provides seven tools for memory management and workspace organization. All tools follow the MCP (Model Context Protocol) specification and communicate via JSON-RPC 2.0.

## Available Tools

### 1. perform_onboarding

**⭐ RECOMMENDED FIRST COMMAND** - Comprehensive agent onboarding that initializes workspace and provides complete tool use strategy.

**Purpose**: Streamlined onboarding for AI agents to establish proper workspace context and receive complete strategic guidance for effective zetmem usage.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "project_path": {
            "type": "string",
            "description": "Path to the project directory (optional, uses current directory if not provided)"
        },
        "project_name": {
            "type": "string",
            "description": "Descriptive name for the project workspace (optional)"
        },
        "include_strategy_guide": {
            "type": "boolean",
            "description": "Whether to include the complete strategy guide in response (default: true)",
            "default": true
        }
    },
    "required": []
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "perform_onboarding",
        "arguments": {
            "project_path": "/Users/dev/my-project",
            "project_name": "My Development Project",
            "include_strategy_guide": true
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "content": [{
            "type": "text",
            "text": "🎯 **Zetmem Agent Onboarding Complete**\n\n## Workspace Initialization\nCreated workspace 'My Development Project' (/Users/dev/my-project)\n- **Workspace ID**: /Users/dev/my-project\n- **Memory Count**: 0\n- **Status**: Ready for use\n\n## Quick Start Commands\n1. **Store Memory**: store_coding_memory(content=\"...\", workspace_id=\"/Users/dev/my-project\", code_type=\"...\", context=\"...\")\n2. **Retrieve Memories**: retrieve_relevant_memories(query=\"...\", workspace_id=\"/Users/dev/my-project\", min_relevance=0.3)\n3. **Evolve Network**: evolve_memory_network(scope=\"recent\", max_memories=100)\n\n## Complete Strategy Guide\n[Full strategy guide content included]\n\n**You are now ready to use zetmem effectively!**"
        }]
    }
}
```

### 2. store_coding_memory

Store code snippets and programming context with AI-generated analysis, keywords, and embeddings.

**Purpose**: Capture and preserve coding knowledge for future retrieval and pattern recognition.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "content": {
            "type": "string",
            "description": "The code content or coding context to store"
        },
        "workspace_id": {
            "type": "string",
            "description": "Workspace identifier (path or name) for organizing memories"
        },
        "project_path": {
            "type": "string",
            "description": "Optional project path for context (deprecated: use workspace_id)"
        },
        "code_type": {
            "type": "string",
            "description": "Programming language or code type (e.g., 'javascript', 'python', 'go')"
        },
        "context": {
            "type": "string",
            "description": "Additional context about the code"
        }
    },
    "required": ["content"]
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "store_coding_memory",
        "arguments": {
            "content": "async function fetchUserData(userId) {\n  const response = await fetch(`/api/users/${userId}`);\n  if (!response.ok) {\n    throw new Error(`HTTP error! status: ${response.status}`);\n  }\n  return await response.json();\n}",
            "workspace_id": "web-app-project",
            "code_type": "javascript",
            "context": "Error handling pattern for API calls with async/await"
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "content": [{
            "type": "text",
            "text": "Memory stored successfully!\n\nMemory ID: mem_abc123\nKeywords: [\"async\", \"await\", \"fetch\", \"error-handling\", \"api\"]\nTags: [\"javascript\", \"async-patterns\", \"error-handling\"]\nLinks Created: 3\n\nThe memory has been analyzed and stored with AI-generated keywords and tags. It's now available for future retrieval and will be linked to related memories."
        }]
    }
}
```

### 3. retrieve_relevant_memories

Search and retrieve stored memories using vector similarity search.

**Purpose**: Find relevant code examples, patterns, and solutions from past experiences.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "query": {
            "type": "string",
            "description": "The search query (code snippet, problem description, or keywords)"
        },
        "workspace_id": {
            "type": "string",
            "description": "Workspace identifier to filter results (optional)"
        },
        "max_results": {
            "type": "integer",
            "description": "Maximum number of results to return (default: 5)",
            "default": 5
        },
        "project_filter": {
            "type": "string",
            "description": "Optional project path to filter results (deprecated: use workspace_id)"
        },
        "code_types": {
            "type": "array",
            "items": {"type": "string"},
            "description": "Optional array of code types to filter by"
        },
        "min_relevance": {
            "type": "number",
            "description": "Minimum relevance score (0.0-1.0, default: 0.7)",
            "default": 0.7
        }
    },
    "required": ["query"]
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
        "name": "retrieve_relevant_memories",
        "arguments": {
            "query": "error handling in async functions",
            "workspace_id": "web-app-project",
            "max_results": 3,
            "code_types": ["javascript", "typescript"],
            "min_relevance": 0.8
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "result": {
        "content": [{
            "type": "text",
            "text": "Found 3 relevant memories:\n\n**Memory 1** (Relevance: 92.3%)\nID: mem_abc123\nContext: Error handling pattern for API calls with async/await\nKeywords: [\"async\", \"await\", \"fetch\", \"error-handling\"]\nTags: [\"javascript\", \"async-patterns\"]\nProject: web-app-project\nCode Type: javascript\nMatch Reason: High similarity in async error handling patterns\n\nContent:\n```\nasync function fetchUserData(userId) {\n  const response = await fetch(`/api/users/${userId}`);\n  if (!response.ok) {\n    throw new Error(`HTTP error! status: ${response.status}`);\n  }\n  return await response.json();\n}\n```\n\n---\n\n"
        }]
    }
}
```

### 4. evolve_memory_network

Trigger AI-powered evolution of the memory network to identify patterns and optimize connections.

**Purpose**: Continuously improve memory organization and discover hidden relationships.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "trigger_type": {
            "type": "string",
            "description": "Type of trigger: 'manual', 'scheduled', or 'event'",
            "default": "manual"
        },
        "scope": {
            "type": "string",
            "description": "Scope of evolution: 'recent', 'all', or 'project'",
            "default": "recent"
        },
        "max_memories": {
            "type": "integer",
            "description": "Maximum number of memories to analyze (default: 100)",
            "default": 100
        },
        "project_path": {
            "type": "string",
            "description": "Project path when scope is 'project'"
        }
    }
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
        "name": "evolve_memory_network",
        "arguments": {
            "trigger_type": "manual",
            "scope": "recent",
            "max_memories": 50
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "result": {
        "content": [{
            "type": "text",
            "text": "Memory network evolution completed!\n\nResults:\n- Memories Analyzed: 50\n- Memories Evolved: 12\n- Links Created: 8\n- Links Strengthened: 15\n- Contexts Updated: 7\n- Duration: 1250 ms\n\nThe memory network has been analyzed and optimized. New connections have been identified and memory contexts have been improved based on AI analysis."
        }]
    }
}
```

### 5. workspace_init

Smart workspace initialization - creates new workspace or retrieves existing one.

**Purpose**: Establish a workspace context for organizing memories by project or domain.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "identifier": {
            "type": "string",
            "description": "Path or name for the workspace. If not provided, uses current working directory"
        },
        "name": {
            "type": "string",
            "description": "Human-readable name for the workspace (optional)"
        }
    },
    "required": []
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
        "name": "workspace_init",
        "arguments": {
            "identifier": "/home/user/projects/web-app",
            "name": "Web Application Project"
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 4,
    "result": {
        "content": [{
            "type": "text",
            "text": "Retrieved workspace 'Web Application Project' (ws_xyz789)\n\nWorkspace Details:\n```json\n{\n  \"workspace\": {\n    \"id\": \"ws_xyz789\",\n    \"name\": \"Web Application Project\",\n    \"path\": \"/home/user/projects/web-app\",\n    \"memory_count\": 42,\n    \"created_at\": \"2024-01-15T10:30:00Z\",\n    \"updated_at\": \"2024-01-20T15:45:00Z\"\n  },\n  \"created\": false\n}\n```"
        }]
    }
}
```

### 6. workspace_create

Explicit workspace creation - fails if workspace already exists.

**Purpose**: Create a new workspace with explicit control over naming and configuration.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "identifier": {
            "type": "string",
            "description": "Path or name for the workspace (required)"
        },
        "name": {
            "type": "string",
            "description": "Human-readable name for the workspace (optional)"
        },
        "description": {
            "type": "string",
            "description": "Description of the workspace (optional)"
        }
    },
    "required": ["identifier"]
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
        "name": "workspace_create",
        "arguments": {
            "identifier": "ml-experiments",
            "name": "Machine Learning Experiments",
            "description": "Workspace for ML model development and experimentation"
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 5,
    "result": {
        "content": [{
            "type": "text",
            "text": "Created workspace 'Machine Learning Experiments' (ws_ml123)\n\nWorkspace Details:\n```json\n{\n  \"workspace\": {\n    \"id\": \"ws_ml123\",\n    \"name\": \"Machine Learning Experiments\",\n    \"description\": \"Workspace for ML model development and experimentation\",\n    \"memory_count\": 0,\n    \"created_at\": \"2024-01-20T16:00:00Z\",\n    \"updated_at\": \"2024-01-20T16:00:00Z\"\n  },\n  \"created\": true\n}\n```"
        }]
    }
}
```

### 7. workspace_retrieve

Explicit workspace retrieval - fails if workspace doesn't exist.

**Purpose**: Get detailed information about an existing workspace.

**Input Schema**:
```json
{
    "type": "object",
    "properties": {
        "identifier": {
            "type": "string",
            "description": "Path or name of the workspace to retrieve (required)"
        }
    },
    "required": ["identifier"]
}
```

**Example Request**:
```json
{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
        "name": "workspace_retrieve",
        "arguments": {
            "identifier": "web-app-project"
        }
    }
}
```

**Example Response**:
```json
{
    "jsonrpc": "2.0",
    "id": 6,
    "result": {
        "content": [{
            "type": "text",
            "text": "Retrieved workspace 'Web Application Project' (ws_xyz789) with 42 memories\n\nWorkspace Details:\n```json\n{\n  \"workspace\": {\n    \"id\": \"ws_xyz789\",\n    \"name\": \"Web Application Project\",\n    \"path\": \"/home/user/projects/web-app\",\n    \"memory_count\": 42,\n    \"created_at\": \"2024-01-15T10:30:00Z\",\n    \"updated_at\": \"2024-01-20T15:45:00Z\",\n    \"metadata\": {\n      \"language\": \"javascript\",\n      \"framework\": \"react\"\n    }\n  },\n  \"created\": false\n}\n```"
        }]
    }
}
```

## Common Patterns and Best Practices

### 1. Workspace-First Approach

Always initialize or create a workspace before storing memories:

```json
// Step 1: Initialize workspace
{
    "method": "tools/call",
    "params": {
        "name": "workspace_init",
        "arguments": {
            "identifier": "my-project"
        }
    }
}

// Step 2: Store memories in that workspace
{
    "method": "tools/call",
    "params": {
        "name": "store_coding_memory",
        "arguments": {
            "content": "...",
            "workspace_id": "my-project"
        }
    }
}
```

### 2. Contextual Memory Storage

Provide rich context when storing memories:

```json
{
    "name": "store_coding_memory",
    "arguments": {
        "content": "class RateLimiter { ... }",
        "workspace_id": "api-project",
        "code_type": "typescript",
        "context": "Token bucket rate limiter implementation with Redis backend for distributed systems"
    }
}
```

### 3. Filtered Retrieval

Use filters to get more relevant results:

```json
{
    "name": "retrieve_relevant_memories",
    "arguments": {
        "query": "rate limiting implementation",
        "workspace_id": "api-project",
        "code_types": ["typescript", "javascript"],
        "min_relevance": 0.85,
        "max_results": 5
    }
}
```

### 4. Regular Evolution

Periodically evolve the memory network to maintain quality:

```json
{
    "name": "evolve_memory_network",
    "arguments": {
        "trigger_type": "scheduled",
        "scope": "recent",
        "max_memories": 100
    }
}
```

## Error Handling

All tools return errors in a consistent format:

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "content": [{
            "type": "text",
            "text": "Error: [error message]"
        }],
        "isError": true
    }
}
```

Common error scenarios:
- Missing required parameters
- Invalid workspace identifiers
- Workspace already exists (for create)
- Workspace not found (for retrieve)
- Evolution failures
- Storage failures

## Integration Tips

1. **Session Management**: Use workspace IDs consistently across a coding session
2. **Memory Evolution**: Run evolution after significant memory additions
3. **Query Optimization**: Use specific queries and appropriate relevance thresholds
4. **Error Recovery**: Implement retry logic for transient failures
5. **Batch Operations**: Store related memories together for better linking

## Performance Considerations

- **Vector Search**: Retrieval performance depends on the number of memories
- **Evolution**: Can be resource-intensive for large memory sets
- **Workspace Operations**: Fast lookups using normalized identifiers
- **Memory Storage**: Includes AI analysis which may add latency

## Security and Privacy

- Memories are stored locally in the configured vector database
- No external API calls except for embeddings (if configured)
- Workspace isolation ensures project separation
- All data remains under user control