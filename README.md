## ZetMem MCP Server
#### version: alpha

---

This is an experiment in AI Assisted Software Engineering.
A human wrote [NONE] of the code directly (not yet, anyway).
USE AT OWN RISK :)

---

An AI-powered memory system for Claude Code/Claude Desktop that enables persistent, contextual awareness across coding sessions.

## Features

### Core Memory System
- **Memory Creation**: Store code snippets with AI-generated keywords, tags, and embeddings
- **Memory Retrieval**: Vector similarity search with ranking and filtering
- **Memory Evolution**: AI-driven analysis to update and optimize memory networks
- **MCP Integration**: JSON-RPC 2.0 server compatible with Claude Code

### Advanced Capabilities (Phase 2)
- **Real Embeddings**: Sentence-transformers and OpenAI embedding services
- **Intelligent Evolution**: Automated memory network optimization
- **Prompt Engineering**: Template-based LLM prompt management
- **Monitoring & Metrics**: Comprehensive Prometheus observability
- **Task Scheduling**: Cron-based automated maintenance
- **Multi-LLM Support**: LiteLLM proxy for fallback and model flexibility
- **Vector Storage**: ChromaDB for scalable similarity search

## Quick Start

### 🚀 Spin up Docker services

```bash
git clone git@github.com:nixlim/zetmem.git
cd zetmem
./scripts/install.sh
```

###  Edit Claude Desktop/Claude Code Configuration

```json
{
    "mcpServers": {
        "zetmem": {
            "command": "/absolute/path/to/zetmem/zetmem-server",
            "args": ["-config", "/absolute/path/to/zetmem/config/production.yaml"],
            "env": {
                "OPENAI_API_KEY": "${OPENAI_API_KEY- <your OPENAI_API_KEY>}"
            }
        }
    }
}

```

### List of Tools

ZetMem provides the following MCP tools for memory management and AI-assisted development:

#### Core Memory Tools
- **`store_coding_memory`** - Store code snippets, solutions, and development insights with AI-generated analysis
- **`retrieve_relevant_memories`** - Search and retrieve stored memories using vector similarity search
- **`evolve_memory_network`** - Trigger memory network evolution to identify patterns and optimize connections

#### Workspace Management
- **`workspace_init`** - Initialize or retrieve workspace for organizing memories by project
- **`workspace_create`** - Create new workspace with specified configuration
- **`workspace_retrieve`** - Get detailed workspace information and memory statistics

#### Agent Onboarding
- **`perform_onboarding`** - Comprehensive agent onboarding that initializes workspace and provides complete tool use strategy and best practices

Each tool includes comprehensive error handling, input validation, and detailed response formatting to ensure reliable operation in AI-assisted development workflows.


### 📚 Documentation

See /docs

### Verification

After installation, verify ZetMem is working:

```bash
# Check services
docker-compose ps

# Validate installation
./scripts/validate_installation.sh

# Test in Claude
# Ask Claude: "What tools do you have available?"
# You should see: store_coding_memory, retrieve_relevant_memories, evolve_memory_network
```

## Configuration

Configuration is managed through YAML files and environment variables:

- `config/development.yaml` - Development settings
- `config/production.yaml` - Production settings
- `.env` - Environment variables (API keys, overrides)

Key configuration sections:

- **server**: Port, logging, request limits
- **chromadb**: Vector database connection
- **litellm**: LLM proxy settings and fallbacks
- **evolution**: Memory evolution scheduling
- **monitoring**: Metrics and tracing

## Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌──────────────┐
│   Claude Code   │───▶│  MCP Server  │───▶│  Memory      │
│                 │    │              │    │  System      │
└─────────────────┘    └──────────────┘    └──────┬───────┘
                                                  │
                                                  ▼
                                        ┌──────────────┐
                                        │   LiteLLM    │
                                        │   Analysis   │
                                        └──────┬───────┘
                                               │
                                               ▼
                                        ┌──────────────┐
                                        │  ChromaDB    │
                                        │ Vector Store │
                                        └──────────────┘
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Development build
go build -o zetmem-server cmd/server/main.go

# Production build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zetmem-server cmd/server/main.go
```

## Monitoring

The server exposes Prometheus metrics on port 9090:

- Memory operation counts
- LLM request latency
- Vector search duration
- Error rates

Access metrics at: `http://localhost:9090/metrics`

## Troubleshooting

### Common Issues

1. **ChromaDB connection failed**:
    - Ensure ChromaDB is running: `docker-compose ps chromadb`
    - Check URL in config: `chromadb.url`

2. **LLM API errors**:
    - Verify API key in `.env` file
    - Check rate limits and quotas
    - Review fallback models in config

3. **Memory storage errors**:
    - Check ChromaDB logs: `docker-compose logs chromadb`
    - Verify collection initialization

### Logs

View server logs:
```bash
# Docker deployment
docker-compose logs zetmem-server

# Direct execution
./zetmem-server -log-level debug
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

MIT License

# The What and The Why

I have recently been interested in the problem of persistent context of AI Coding Agents and their context windows.
Gemini Pro has 2M tokens of context. Claude - 200K. Claude is hands down better coder at specific tasks. But big picture view - Gemini Pro is your best bet.

Augment Code has the best context management system I worked with so far. But, I wanted to give the agent things like semantic search, evolving memory, persistent context across sessions.

I went to Arxiv, found the A-MEM paper and built a ZetMem MCP Server with AI.

In 10 hours I delivered a project that would take me, working solo without the AI, about 2-3 weeks of 5 days a week, 8 hours a day.

It is by no means perfect. It is good enough though. I will continue to improve it.

It works. It has tests, startup scripts, local docker. Claude Desktop integration works, as should Claude Code.

I did not write a line of code. I paired with AI, I navigated and it drove. 

---

ACKNOWLEDGEMENTS:
This MCP Server was built on the basis of the following paper:

```
@article{xu2025mem,
title={A-mem: Agentic memory for llm agents},
author={Xu, Wujiang and Liang, Zujie and Mei, Kai and Gao, Hang and Tan, Juntao and Zhang, Yongfeng},
journal={arXiv preprint arXiv:2502.12110},
year={2025}
}
```
Link to pdf of paper: https://arxiv.org/pdf/2502.12110v1
Link to paper's github: https://github.com/WujiangXu/A-mem

The authors of the paper also have their own implementation of the system (don't think it's an MCP Server):
https://github.com/WujiangXu/A-mem-sys
___