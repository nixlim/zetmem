# ZetMem MCP Server - Project Structure Analysis

## Overview

The ZetMem (Adaptive Memory) MCP Server is a sophisticated Go-based Model Context Protocol (MCP) server designed to provide advanced memory management capabilities for AI assistants. The project implements a scalable, event-driven architecture with comprehensive monitoring, embedding services, and memory evolution features.

## Technology Stack

- **Primary Language**: Go 1.23
- **Core Dependencies**:
  - `go.uber.org/zap` - Structured logging
  - `github.com/prometheus/client_golang` - Metrics and monitoring
  - `github.com/google/uuid` - UUID generation
  - `github.com/joho/godotenv` - Environment variable management
  - `gopkg.in/yaml.v3` - YAML configuration parsing

- **Infrastructure Components**:
  - **ChromaDB** - Vector database for memory embeddings
  - **Redis** - Caching and session management
  - **RabbitMQ** - Message queue for evolution workers (optional)
  - **Prometheus** - Metrics collection and monitoring
  - **Sentence Transformers** - Embedding service (Docker container)

## Directory Structure

```
zetmem/
├── cmd/                          # Application entry points
│   └── server/
│       └── main.go              # Main server executable
├── pkg/                         # Core packages
│   ├── config/                  # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── mcp/                     # MCP protocol implementation
│   │   └── server.go
│   ├── memory/                  # Memory system core
│   │   ├── system.go           # Memory system coordinator
│   │   ├── evolution.go        # Memory evolution manager
│   │   ├── tools.go            # MCP tool implementations
│   │   └── workspace_tools.go  # Workspace management tools
│   ├── models/                  # Data models
│   │   ├── mcp.go              # MCP protocol models
│   │   ├── memory.go           # Memory system models
│   │   └── memory_test.go
│   ├── monitoring/              # Observability
│   │   └── metrics.go          # Prometheus metrics
│   ├── scheduler/               # Task scheduling
│   │   └── scheduler.go        # Evolution scheduler
│   └── services/                # External service integrations
│       ├── chromadb.go         # Vector database service
│       ├── embedding.go        # Embedding service client
│       ├── litellm.go          # LLM service integration
│       ├── prompts.go          # Prompt management
│       └── workspace.go        # Workspace service
├── config/                      # Configuration files
│   ├── development.yaml        # Development environment config
│   ├── docker.yaml            # Docker environment config
│   ├── production.yaml        # Production environment config
│   └── claude_*.json          # Claude integration configs
├── docker/                      # Docker-related files
│   └── sentence-transformers/  # Embedding service container
│       ├── Dockerfile
│       └── app.py
├── prompts/                     # Prompt templates
│   ├── enhanced_note_construction.yaml
│   ├── memory_evolution.yaml
│   └── note_construction.yaml
├── scripts/                     # Utility scripts
│   ├── install.sh              # Installation script
│   ├── setup.sh                # Setup script
│   ├── test_mcp.py             # MCP testing script
│   ├── test_phase2.py          # Phase 2 testing script
│   └── validate_installation.sh # Validation script
├── monitoring/                  # Monitoring configuration
│   └── prometheus.yml          # Prometheus config
├── memory/                      # Runtime memory storage
│   ├── agents/                 # Agent memory storage
│   ├── sessions/               # Session data
│   └── claude-flow-data.json   # Claude Flow integration data
├── coordination/                # Coordination artifacts
│   ├── memory_bank/            # Shared memory bank
│   ├── orchestration/          # Task orchestration
│   └── subtasks/               # Subtask definitions
├── docs/                        # Documentation
│   ├── api/                    # API documentation
│   ├── architecture/           # Architecture docs
│   ├── components/             # Component documentation
│   ├── deployment/             # Deployment guides
│   ├── guides/                 # User guides
│   └── infrastructure/         # Infrastructure docs
├── Dockerfile                   # Main server Docker image
├── docker-compose.yml          # Docker Compose configuration
├── Makefile                    # Build automation
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── CLAUDE.md                   # Claude Code configuration
└── documentation_workflow.json  # Documentation workflow
```

## Core Components

### 1. MCP Server (`pkg/mcp/server.go`)
- Implements the Model Context Protocol
- Manages tool registration and execution
- Handles JSON-RPC communication
- Coordinates with memory system

### 2. Memory System (`pkg/memory/`)
- **system.go**: Core memory storage and retrieval
- **evolution.go**: Automatic memory network evolution
- **tools.go**: MCP tool implementations for memory operations
- **workspace_tools.go**: Workspace-specific memory management

### 3. Services Layer (`pkg/services/`)
- **ChromaDB Service**: Vector database operations for similarity search
- **Embedding Service**: Text-to-vector conversion using Sentence Transformers
- **LiteLLM Service**: LLM integration for memory evolution
- **Prompt Manager**: Dynamic prompt template management
- **Workspace Service**: Isolated memory workspace management

### 4. Monitoring & Observability (`pkg/monitoring/`)
- Prometheus metrics collection
- Custom metrics for memory operations
- Performance tracking
- Health endpoints

### 5. Scheduler (`pkg/scheduler/`)
- Cron-based task scheduling
- Memory evolution automation
- Background job management

## Entry Point Analysis

The main entry point (`cmd/server/main.go`) performs the following initialization sequence:

1. **Configuration Loading**
   - Command-line flag parsing
   - Environment variable loading (.env file)
   - YAML configuration parsing

2. **Logger Initialization**
   - Structured JSON logging with zap
   - Configurable log levels (debug, info, warn, error)
   - Performance-optimized sampling

3. **Service Initialization** (in order):
   - LiteLLM Service (LLM provider integration)
   - Embedding Service (text vectorization)
   - ChromaDB Service (vector storage)
   - Prompt Manager (template management)
   - Workspace Service (memory isolation)
   - Memory System (core memory operations)
   - Evolution Manager (memory network evolution)

4. **Monitoring Setup**
   - Prometheus metrics server on separate port
   - Background metrics collection

5. **Scheduler Initialization**
   - Evolution task scheduling
   - Periodic memory optimization

6. **MCP Tool Registration**
   - `store_coding_memory` - Store new memories
   - `retrieve_relevant_memories` - Query memories
   - `evolve_memory_network` - Trigger evolution
   - `workspace_init` - Initialize workspace
   - `workspace_create` - Create new workspace
   - `workspace_retrieve` - Get workspace data

7. **Server Startup**
   - Graceful shutdown handling
   - Signal interruption support
   - Context-based lifecycle management

## Configuration Architecture

The system uses a layered configuration approach:

1. **Environment Variables** (.env file)
   - API keys (OpenAI, etc.)
   - Service URLs
   - Runtime settings

2. **YAML Configuration** (config/*.yaml)
   - Service-specific settings
   - Feature toggles
   - Performance tuning

3. **Command-line Flags**
   - Config file selection
   - Log level override
   - Environment specification

## Docker Architecture

The project uses Docker Compose to orchestrate multiple services:

1. **zetmem-server** - Main MCP server
2. **chromadb** - Vector database
3. **redis** - Caching layer
4. **rabbitmq** - Message queue (optional)
5. **sentence-transformers** - Embedding service
6. **prometheus** - Metrics collection

## Build System

The Makefile provides comprehensive build automation:

- **Development**: `make dev` - Local development with hot reload
- **Docker**: `make docker-run` - Full stack deployment
- **Testing**: `make test` - Run test suite
- **Building**: `make build` - Create binary
- **Release**: `make release` - Multi-platform builds

## Integration Points

1. **Claude Flow Integration**
   - Configuration files for Claude Code and Desktop
   - Memory persistence for swarm coordination
   - Workflow templates

2. **MCP Protocol**
   - Standard MCP tool interface
   - JSON-RPC communication
   - Async operation support

3. **External Services**
   - OpenAI/LiteLLM for LLM operations
   - Sentence Transformers for embeddings
   - ChromaDB for vector search

## Key Design Patterns

1. **Service-Oriented Architecture**
   - Clear separation of concerns
   - Dependency injection
   - Interface-based design

2. **Event-Driven Processing**
   - Asynchronous memory evolution
   - Background task scheduling
   - Message queue integration

3. **Observability-First**
   - Comprehensive metrics
   - Structured logging
   - Health monitoring

4. **Configuration Management**
   - Environment-specific configs
   - Hot-reloading support
   - Secure credential handling

## Summary

The ZetMem MCP Server represents a well-architected, production-ready system for adaptive memory management in AI applications. Its modular design, comprehensive monitoring, and scalable architecture make it suitable for both development and production deployments. The integration with Claude Flow and adherence to MCP standards ensures compatibility with the broader AI assistant ecosystem.