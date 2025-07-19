# Changelog

All notable changes to the ZetMem MCP Server project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Placeholder for future features and improvements

### Changed
- Placeholder for future changes

### Fixed
- Placeholder for future bug fixes

## [1.0.0] - 2025-07-19

### Added
- Complete ZetMem MCP Server implementation with comprehensive rebranding
- Systematic memory documentation across multiple memory systems
- Comprehensive codebase review and validation process
- Production-ready configuration and infrastructure setup
- Full MCP (Model Context Protocol) compliance with updated tool names
- Docker containerization with zetmem-* service naming
- Prometheus monitoring with 14 zetmem_* prefixed metrics
- ChromaDB integration with zetmem_memories collections
- Claude Desktop and Claude Code integration support
- Automated installation and validation scripts
- Comprehensive documentation and API reference

### Changed
- **BREAKING**: Complete rebranding from A-MEM to ZetMem across entire codebase
- **BREAKING**: Go module path updated from `github.com/amem/mcp-server` to `github.com/zetmem/mcp-server`
- **BREAKING**: Binary name changed from `amem-server` to `zetmem-server`
- **BREAKING**: All environment variables updated from `AMEM_*` to `ZETMEM_*` pattern:
  - `AMEM_ENV` → `ZETMEM_ENV`
  - `AMEM_PORT` → `ZETMEM_PORT`
  - `AMEM_LOG_LEVEL` → `ZETMEM_LOG_LEVEL`
  - `AMEM_CONFIG_PATH` → `ZETMEM_CONFIG_PATH`
  - `AMEM_PROMPTS_PATH` → `ZETMEM_PROMPTS_PATH`
  - `AMEM_EVOLUTION_*` → `ZETMEM_EVOLUTION_*`
  - `AMEM_METRICS_*` → `ZETMEM_METRICS_*`
  - `AMEM_STRATEGY_GUIDE_*` → `ZETMEM_STRATEGY_GUIDE_*`
- **BREAKING**: Docker service names updated:
  - `amem-server` → `zetmem-server`
  - `amem-network` → `zetmem-network`
- **BREAKING**: Database collections renamed:
  - `amem_memories` → `zetmem_memories`
  - `amem_memories_dev` → `zetmem_memories_dev`
  - `amem_memories_prod` → `zetmem_memories_prod`
- **BREAKING**: All 14 Prometheus metrics renamed with `zetmem_*` prefix:
  - `amem_memory_operations_total` → `zetmem_memory_operations_total`
  - `amem_memory_operation_duration_seconds` → `zetmem_memory_operation_duration_seconds`
  - `amem_llm_requests_total` → `zetmem_llm_requests_total`
  - `amem_llm_request_duration_seconds` → `zetmem_llm_request_duration_seconds`
  - `amem_llm_tokens_total` → `zetmem_llm_tokens_total`
  - `amem_vector_searches_total` → `zetmem_vector_searches_total`
  - `amem_vector_search_duration_seconds` → `zetmem_vector_search_duration_seconds`
  - `amem_evolution_runs_total` → `zetmem_evolution_runs_total`
  - `amem_evolution_duration_seconds` → `zetmem_evolution_duration_seconds`
  - `amem_active_connections` → `zetmem_active_connections`
  - `amem_errors_total` → `zetmem_errors_total`
  - `amem_cache_hits_total` → `zetmem_cache_hits_total`
  - `amem_cache_misses_total` → `zetmem_cache_misses_total`
- **BREAKING**: MCP tool name updated from `amem-augmented` to `zetmem-augmented`
- **BREAKING**: Docker user and group changed from `amem:amem` to `zetmem:zetmem`
- **BREAKING**: RabbitMQ credentials updated from `amem/amem_password` to `zetmem/zetmem_password`
- Project documentation updated to reflect ZetMem branding throughout
- README.md title updated from "A-MEM MCP Server" to "ZetMem MCP Server"
- API documentation updated with new server descriptions and tool names
- Installation and configuration guides updated with new binary and service names
- All code comments and log messages updated to use ZetMem terminology
- Build system updated to produce zetmem-server binary
- Docker Compose configuration updated with new service and network names
- Prometheus monitoring configuration updated with new job names
- Claude Desktop and Claude Code configuration examples updated

### Fixed
- Resolved all import statement dependencies following module path change
- Corrected configuration loading to use new ZETMEM_* environment variables
- Fixed Docker container builds to use correct binary names
- Resolved MCP tool registration with updated tool names
- Corrected database collection references throughout codebase
- Fixed monitoring metrics registration with new naming convention

### Security
- No security vulnerabilities identified during comprehensive codebase review
- Maintained all existing security practices and input validation
- Preserved authentication and authorization mechanisms
- Container security maintained with updated user accounts

## Migration Guide

### Upgrading from A-MEM to ZetMem

#### Environment Variables
Update all environment variables in your configuration:
```bash
# Old (A-MEM)
AMEM_ENV=production
AMEM_PORT=8080
AMEM_LOG_LEVEL=info

# New (ZetMem)
ZETMEM_ENV=production
ZETMEM_PORT=8080
ZETMEM_LOG_LEVEL=info
```

#### Docker Services
Update your docker-compose.yml:
```yaml
# Old service name
services:
  amem-server:
    # ...

# New service name
services:
  zetmem-server:
    # ...
```

#### Binary Name
Update any scripts or commands:
```bash
# Old binary
./amem-server -config config/production.yaml

# New binary
./zetmem-server -config config/production.yaml
```

#### MCP Configuration
Update Claude Desktop/Code configuration:
```json
{
  "mcpServers": {
    "zetmem-augmented": {
      "command": "/path/to/zetmem-server",
      "args": ["-config", "/path/to/config/production.yaml"]
    }
  }
}
```

#### Database Collections
ChromaDB collections are automatically renamed. Update any external references:
- `amem_memories` → `zetmem_memories`
- `amem_memories_dev` → `zetmem_memories_dev`

#### Prometheus Metrics
Update monitoring dashboards and alerts to use new metric names with `zetmem_*` prefix.

---

**Note**: This version represents the completion of the comprehensive A-MEM to ZetMem rebranding project. All core functionality has been preserved while achieving complete naming consistency throughout the codebase, documentation, and infrastructure.
