# Configuration Guide

## Overview

The ZetMem MCP Server uses a hierarchical configuration system that supports multiple environments, hot reloading, and both YAML and environment variable configuration methods.

## Configuration Hierarchy

```
Priority (highest to lowest):
1. Environment Variables
2. Command-line flags
3. Configuration files (YAML)
4. Default values
```

## Configuration Files

### File Structure

```
config/
├── development.yaml      # Development environment
├── docker.yaml          # Docker environment
├── production.yaml      # Production environment
├── claude_code_settings.json    # Claude Code integration
├── claude_desktop_config.json   # Claude Desktop settings
└── claude_development.json      # Claude dev configuration
```

### Core Configuration Schema

```yaml
# config/docker.yaml
server:
  port: 8080
  log_level: info
  max_request_size: 10MB

chromadb:
  url: "http://chromadb:8000"
  collection: "zetmem_memories"
  batch_size: 100

litellm:
  default_model: "gpt-4-turbo-preview"
  fallback_models:
    - "gpt-3.5-turbo"
    - "gpt-4"
  max_retries: 3
  timeout: 30s
  rate_limit: 60

embedding:
  service: "sentence-transformers"
  model: "all-MiniLM-L6-v2"
  batch_size: 32
  url: "http://sentence-transformers:8000"

evolution:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 50
  worker_count: 3

prompts:
  directory: "/app/prompts"
  cache_enabled: true
  hot_reload: false

monitoring:
  metrics_port: 9092
  enable_tracing: true
  sample_rate: 0.1
```

## Environment Variables

### Complete Environment Variable Reference

#### Core Application Settings

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `AMEM_ENV` | Environment (development/production) | development | No |
| `AMEM_PORT` | API server port | 8080 | No |
| `AMEM_LOG_LEVEL` | Log level (debug/info/warn/error) | info | No |
| `AMEM_CONFIG_PATH` | Path to config directory | ./config | No |
| `AMEM_PROMPTS_PATH` | Path to prompts directory | ./prompts | No |

#### API Keys

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OPENAI_API_KEY` | OpenAI API key | - | Yes |
| `ANTHROPIC_API_KEY` | Anthropic API key | - | No |
| `COHERE_API_KEY` | Cohere API key | - | No |
| `HUGGINGFACE_API_KEY` | HuggingFace API key | - | No |

#### ChromaDB Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `CHROMADB_HOST` | ChromaDB host URL | http://localhost:8000 | No |
| `CHROMADB_COLLECTION` | Collection name | amem_memories_dev | No |
| `CHROMADB_BATCH_SIZE` | Batch size for operations | 100 | No |

#### LiteLLM Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LITELLM_DEFAULT_MODEL` | Default LLM model | gpt-4.1 | No |
| `LITELLM_MAX_RETRIES` | Maximum retry attempts | 3 | No |
| `LITELLM_TIMEOUT_SECONDS` | Request timeout | 30 | No |
| `LITELLM_RATE_LIMIT` | Requests per minute | 60 | No |
| `LITELLM_TEMPERATURE` | Model temperature | 0.7 | No |
| `LITELLM_MAX_TOKENS` | Max tokens per request | 2000 | No |

#### Embedding Service

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `EMBEDDING_SERVICE` | Service type | sentence-transformers | No |
| `EMBEDDING_MODEL` | Model name | all-MiniLM-L6-v2 | No |
| `EMBEDDING_BATCH_SIZE` | Batch size | 32 | No |

#### Memory Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MEMORY_RETENTION_DAYS` | Memory retention period | 30 | No |
| `MEMORY_MAX_ENTRIES_PER_AGENT` | Max entries per agent | 10000 | No |
| `MEMORY_CHUNK_SIZE` | Text chunk size | 512 | No |
| `MEMORY_OVERLAP_SIZE` | Chunk overlap size | 50 | No |

#### Security Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SECRET_KEY` | JWT secret key | - | Yes |
| `JWT_ALGORITHM` | JWT algorithm | HS256 | No |
| `JWT_EXPIRATION_DELTA` | Token expiration (seconds) | 3600 | No |
| `API_KEY_HEADER` | API key header name | X-API-Key | No |

#### Feature Flags

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENABLE_MEMORY_COMPRESSION` | Enable memory compression | true | No |
| `ENABLE_MEMORY_ENCRYPTION` | Enable memory encryption | false | No |
| `ENABLE_DISTRIBUTED_MEMORY` | Enable distributed memory | false | No |
| `ENABLE_MEMORY_SEARCH` | Enable memory search | true | No |
| `ENABLE_MEMORY_EXPORT` | Enable memory export | true | No |

## Configuration by Environment

### Development Configuration

```yaml
# config/development.yaml
server:
  port: 8080
  log_level: debug
  max_request_size: 50MB

chromadb:
  url: "http://localhost:8000"
  collection: "amem_memories_dev"

litellm:
  default_model: "gpt-3.5-turbo"
  rate_limit: 120

monitoring:
  metrics_port: 9090
  enable_tracing: true
  sample_rate: 1.0  # Full tracing in dev

prompts:
  hot_reload: true  # Enable hot reload for development
```

### Production Configuration

```yaml
# config/production.yaml
server:
  port: 8080
  log_level: info
  max_request_size: 10MB

chromadb:
  url: "${CHROMADB_URL}"
  collection: "amem_memories_prod"
  batch_size: 500

litellm:
  default_model: "gpt-4-turbo-preview"
  rate_limit: 60
  timeout: 60s

evolution:
  enabled: true
  schedule: "0 2 * * *"
  worker_count: 10

monitoring:
  metrics_port: 9090
  enable_tracing: true
  sample_rate: 0.01  # 1% sampling in production

prompts:
  cache_enabled: true
  hot_reload: false
```

## Advanced Configuration Options

### Rate Limiting

```yaml
rate_limiting:
  enabled: true
  requests_per_minute: 60
  requests_per_hour: 1000
  burst_size: 10
  
  # Per-endpoint limits
  endpoints:
    /api/v1/remember:
      requests_per_minute: 100
    /api/v1/recall:
      requests_per_minute: 200
```

### Storage Configuration

```yaml
storage:
  type: local  # Options: local, s3, gcs, azure
  
  # Local storage
  local:
    path: ./data/storage
    
  # S3 storage
  s3:
    bucket: amem-storage
    region: us-east-1
    prefix: memories/
    
  # Backup configuration
  backup:
    enabled: true
    interval: 24h
    retention: 7d
    destination: ./data/backups
```

### Monitoring and Observability

```yaml
monitoring:
  metrics:
    enabled: true
    port: 9090
    path: /metrics
    
  tracing:
    enabled: true
    exporter: jaeger  # Options: jaeger, zipkin, otlp
    endpoint: http://jaeger:14268/api/traces
    sample_rate: 0.1
    
  logging:
    level: info
    format: json  # Options: json, text
    output: stdout
    file:
      enabled: true
      path: /var/log/amem/server.log
      max_size: 100MB
      max_age: 7d
      max_backups: 5
```

### Clustering Configuration

```yaml
cluster:
  enabled: false
  node_id: "${HOSTNAME}"
  
  # Cluster discovery
  discovery:
    type: consul  # Options: consul, etcd, kubernetes
    consul:
      address: consul:8500
      service_name: amem-server
      
  # Distributed locking
  locking:
    type: redis
    redis:
      address: redis:6379
      prefix: amem:locks:
```

## Claude Integration Configuration

### Claude Code Settings

```json
{
  "mcpServers": {
    "amem": {
      "command": "go",
      "args": ["run", "./cmd/mcp"],
      "env": {
        "AMEM_CONFIG": "./config/claude_development.json"
      }
    }
  }
}
```

### Claude Desktop Configuration

```json
{
  "amem": {
    "enabled": true,
    "host": "http://localhost:8080",
    "apiKey": "${AMEM_API_KEY}",
    "features": {
      "memory": true,
      "evolution": true,
      "search": true
    }
  }
}
```

## Configuration Validation

The server validates configuration on startup:

1. **Required fields**: Ensures all required fields are present
2. **Type validation**: Checks correct data types
3. **Range validation**: Validates numeric ranges
4. **Connection tests**: Tests database and service connections

### Validation Command

```bash
# Validate configuration without starting server
./amem-server -validate-config -config ./config/production.yaml
```

## Dynamic Configuration

### Hot Reload

Certain configurations can be reloaded without restart:

- Prompt templates
- Rate limits
- Feature flags
- Log levels

```bash
# Send reload signal
curl -X POST http://localhost:8080/admin/reload-config
```

### Runtime Configuration API

```bash
# Get current configuration
curl http://localhost:8080/admin/config

# Update configuration
curl -X PATCH http://localhost:8080/admin/config \
  -H "Content-Type: application/json" \
  -d '{"server": {"log_level": "debug"}}'
```

## Best Practices

### 1. Environment-Specific Files

Keep environment-specific configurations separate:
- `development.yaml` - Local development
- `docker.yaml` - Docker containers
- `production.yaml` - Production deployment
- `testing.yaml` - Test environment

### 2. Secrets Management

Never commit secrets to version control:
```bash
# Use environment variables
export OPENAI_API_KEY=$(vault kv get -field=api_key secret/openai)

# Or use secret files
docker secret create openai_key ./secrets/openai_key.txt
```

### 3. Configuration Templates

Use templates for team members:
```bash
# Create from template
cp config/development.yaml.template config/development.yaml
```

### 4. Validation

Always validate configuration changes:
```bash
# Validate before deployment
make validate-config

# Test configuration
docker compose run --rm amem-server -validate-config
```

## Troubleshooting

### Common Issues

1. **Missing Required Configuration**
   ```
   Error: required configuration 'openai_api_key' not found
   Solution: Set OPENAI_API_KEY environment variable
   ```

2. **Invalid Configuration Values**
   ```
   Error: invalid rate_limit value: -1
   Solution: Rate limit must be positive integer
   ```

3. **Connection Failures**
   ```
   Error: failed to connect to ChromaDB
   Solution: Check CHROMADB_HOST and network connectivity
   ```

### Debug Configuration

Enable debug mode to see configuration loading:
```bash
AMEM_LOG_LEVEL=debug ./amem-server -config ./config/development.yaml
```

## Migration Guide

### From v1.x to v2.x

```bash
# Migrate configuration
./scripts/migrate-config.sh v1-config.yaml v2-config.yaml

# Key changes:
# - 'llm' section renamed to 'litellm'
# - 'vector_db' renamed to 'chromadb'
# - New 'evolution' section added
```

## Next Steps

- Review [Deployment Guide](./deployment.md) for deployment configuration
- See [Docker Setup](./docker-setup.md) for container configuration
- Check [Quick Start](../deployment/quick-start.md) for minimal setup