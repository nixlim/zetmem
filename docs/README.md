# ZetMem MCP Server Documentation

## 📚 Overview

ZetMem (Adaptive Memory) is an intelligent memory management system that provides AI-powered code memory storage, retrieval, and evolution capabilities through the Model Context Protocol (MCP).

## 🏗️ Architecture

- **[Project Structure](architecture/project-structure.md)** - Complete codebase organization and technology stack
- **[System Overview](architecture/system-overview.mermaid)** - Visual architecture diagram
- **[Memory Flow](architecture/memory-flow.mermaid)** - Sequence diagram of memory operations
- **[Deployment Topology](architecture/deployment-topology.mermaid)** - Container architecture

## 🧩 Core Components

- **[Memory System](components/memory-system.md)** - Adaptive memory storage with evolution algorithms
- **[MCP Server](components/mcp-server.md)** - JSON-RPC 2.0 protocol implementation
- **[Services Layer](components/services.md)** - LiteLLM, ChromaDB, embeddings, and more

## 📖 API Reference

- **[Memory API](api/memory-api.md)** - Complete memory system API documentation
- **[MCP Tools](api/mcp-tools.md)** - All available MCP tools with examples
- **[Services API](api/services-api.md)** - Service layer interfaces and methods

## 🚀 Deployment & Infrastructure

- **[Quick Start Guide](deployment/quick-start.md)** - Get started in 5 minutes
- **[Docker Setup](infrastructure/docker-setup.md)** - Container configuration details
- **[Deployment Guide](infrastructure/deployment.md)** - Production deployment patterns
- **[Configuration Reference](infrastructure/configuration.md)** - All configuration options

## 📘 Integration Guides

- **[Service Integration](guides/service-integration.md)** - How to integrate services
- **[Configuration Guide](configuration/service-config.md)** - Service configuration details

## 🔧 Key Features

### Intelligent Memory Management
- **Vector-based storage** using ChromaDB for semantic search
- **AI-powered analysis** extracting keywords, tags, and summaries
- **Automatic evolution** creating knowledge graphs from memories
- **Workspace isolation** for project-specific memory management

### MCP Protocol Implementation
- **JSON-RPC 2.0** compliant server
- **Six core tools** for memory and workspace operations
- **Extensible architecture** for adding new tools
- **Robust error handling** with standard error codes

### Scalable Architecture
- **Microservices design** with Docker orchestration
- **Horizontal scaling** support for API servers
- **Caching layer** with Redis for performance
- **Message queue** integration with RabbitMQ

### Enterprise Features
- **Prometheus monitoring** with custom metrics
- **Health check endpoints** for all services
- **Configurable logging** with structured output
- **Security hardening** with non-root containers

## 📊 Architecture Diagrams

### System Overview
View the [system architecture diagram](architecture/system-overview.mermaid) to understand component relationships.

### Memory Creation Flow
See the [memory flow sequence](architecture/memory-flow.mermaid) for detailed operation sequences.

### Deployment Architecture
Check the [deployment topology](architecture/deployment-topology.mermaid) for container organization.

## 🛠️ Technical Stack

- **Language**: Go 1.23
- **Vector Database**: ChromaDB
- **LLM Integration**: OpenAI GPT-4 / Anthropic Claude
- **Embeddings**: OpenAI Ada / Sentence Transformers
- **Caching**: Redis
- **Message Queue**: RabbitMQ
- **Monitoring**: Prometheus
- **Container**: Docker & Docker Compose

## 📚 Documentation Index

### By Component
1. [Memory System](components/memory-system.md)
2. [MCP Server](components/mcp-server.md)
3. [Services](components/services.md)

### By Use Case
1. [Quick Start](deployment/quick-start.md)
2. [API Integration](api/mcp-tools.md)
3. [Service Setup](guides/service-integration.md)

### By Role
- **Developers**: Start with [Quick Start](deployment/quick-start.md) and [API Reference](api/mcp-tools.md)
- **DevOps**: Review [Deployment Guide](infrastructure/deployment.md) and [Configuration](infrastructure/configuration.md)
- **Architects**: Explore [Architecture](architecture/project-structure.md) and [Services](components/services.md)

## 🔍 Search This Documentation

Use your browser's search (Ctrl+F / Cmd+F) or grep through the markdown files to find specific topics.

## 📝 Version

Documentation Version: 2.0.0  
Generated: 2025-07-18  
ZetMem Version: Latest

---

*This documentation was generated using the Claude Flow documentation workflow.*