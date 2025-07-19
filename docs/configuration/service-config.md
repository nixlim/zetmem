# Service Configuration Guide

## Environment Variables

### Required Variables

```bash
# OpenAI API Key (Required for LiteLLM and OpenAI embeddings)
OPENAI_API_KEY=sk-your-api-key-here

# ChromaDB Configuration
CHROMADB_URL=http://localhost:8000
CHROMADB_COLLECTION=zetmem_memories
```

### Optional Variables

```bash
# Embedding Service Configuration
EMBEDDING_SERVICE=sentence-transformers  # Options: openai, sentence-transformers, fallback
EMBEDDING_SERVICE_URL=http://localhost:8080  # For sentence-transformers
EMBEDDING_MODEL=all-MiniLM-L6-v2

# LiteLLM Configuration
LLM_DEFAULT_MODEL=gpt-3.5-turbo
LLM_MAX_RETRIES=3
LLM_TIMEOUT=30s
LLM_FALLBACK_MODELS=gpt-3.5-turbo-16k,claude-instant-1

# Workspace Configuration
DEFAULT_WORKSPACE=default
WORKSPACE_AUTO_CREATE=true

# Prompt Manager Configuration
PROMPTS_DIR=./prompts
PROMPTS_CACHE_ENABLED=true
PROMPTS_HOT_RELOAD=true
```

## Configuration Files

### Main Configuration (config/config.yaml)

```yaml
# Service Configuration
services:
  # ChromaDB Vector Database
  chromadb:
    url: "${CHROMADB_URL:-http://localhost:8000}"
    collection: "${CHROMADB_COLLECTION:-zetmem_memories}"
    timeout: 30s
    max_retries: 3
  
  # Embedding Service
  embedding:
    service: "${EMBEDDING_SERVICE:-sentence-transformers}"
    url: "${EMBEDDING_SERVICE_URL:-http://localhost:8080}"
    model: "${EMBEDDING_MODEL:-all-MiniLM-L6-v2}"
    dimension: 384  # 384 for MiniLM, 1536 for OpenAI
    batch_size: 32
  
  # LiteLLM Service
  litellm:
    default_model: "${LLM_DEFAULT_MODEL:-gpt-3.5-turbo}"
    max_retries: ${LLM_MAX_RETRIES:-3}
    timeout: ${LLM_TIMEOUT:-30s}
    fallback_models:
      - gpt-3.5-turbo-16k
      - claude-instant-1
    temperature: 0.1
    max_tokens: 1000
  
  # Prompt Manager
  prompts:
    directory: "${PROMPTS_DIR:-./prompts}"
    cache_enabled: ${PROMPTS_CACHE_ENABLED:-true}
    hot_reload: ${PROMPTS_HOT_RELOAD:-true}
    default_model_config:
      temperature: 0.1
      max_tokens: 1000
  
  # Workspace Service
  workspace:
    auto_create: ${WORKSPACE_AUTO_CREATE:-true}
    default_id: "${DEFAULT_WORKSPACE:-default}"
    validation:
      max_length: 255
      allowed_chars: "a-zA-Z0-9._/-"

# Logging Configuration
logging:
  level: "${LOG_LEVEL:-info}"
  format: "${LOG_FORMAT:-json}"
  output: "${LOG_OUTPUT:-stdout}"

# Performance Settings
performance:
  connection_pool_size: 100
  request_timeout: 30s
  idle_timeout: 90s
  max_concurrent_requests: 50
```

### Prompt Templates (prompts/)

#### prompts/extract_memory.yaml

```yaml
name: extract_memory
version: 1.0.0
description: Extract structured memory from code content
model_config:
  temperature: 0.1
  max_tokens: 1000
  top_p: 0.95
variables:
  focus_areas:
    - purpose
    - algorithms
    - dependencies
    - patterns
template: |
  Analyze the following code and extract key information for memory storage:
  
  Code Type: {{.CodeType}}
  Project: {{.ProjectPath}}
  
  Content:
  ```{{.CodeType}}
  {{.Content}}
  ```
  
  Extract and structure the following information:
  1. Primary purpose and functionality
  2. Key algorithms and logic patterns
  3. External dependencies and imports
  4. Notable design patterns used
  5. Potential areas of improvement
  
  Format your response as JSON with the following structure:
  {
    "summary": "brief description",
    "purpose": "main functionality",
    "algorithms": ["list", "of", "algorithms"],
    "dependencies": ["list", "of", "dependencies"],
    "patterns": ["list", "of", "patterns"],
    "tags": ["relevant", "tags"],
    "keywords": ["search", "keywords"]
  }
```

#### prompts/search_context.yaml

```yaml
name: search_context
version: 1.0.0
description: Generate context for memory search
model_config:
  temperature: 0.3
  max_tokens: 500
template: |
  Based on the search query and retrieved memories, provide a comprehensive answer:
  
  Query: {{.Query}}
  
  Retrieved Memories:
  {{range .Memories}}
  ---
  Content: {{.Content}}
  Context: {{.Context}}
  Tags: {{.Tags}}
  ---
  {{end}}
  
  Synthesize the information and provide:
  1. Direct answer to the query
  2. Related code examples if applicable
  3. Best practices or recommendations
  4. Additional context that might be helpful
```

## Docker Compose Configuration

```yaml
version: '3.8'

services:
  # ChromaDB Vector Database
  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
    volumes:
      - chromadb_data:/chroma/chroma
    environment:
      - CHROMA_SERVER_AUTH_CREDENTIALS_FILE=/chroma/auth.json
      - CHROMA_SERVER_AUTH_CREDENTIALS_PROVIDER=chromadb.auth.simple.SimpleAuthCredentialsProvider
      - CHROMA_SERVER_AUTH_PROVIDER=chromadb.auth.simple.SimpleAuthenticationProvider
      - ANONYMIZED_TELEMETRY=false
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/api/v1/heartbeat"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Sentence Transformers Service (Optional)
  sentence-transformers:
    build:
      context: ./docker/sentence-transformers
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - MODEL_NAME=all-MiniLM-L6-v2
      - MAX_LENGTH=256
      - DEVICE=cpu
    volumes:
      - model_cache:/root/.cache
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # ZetMem Server
  zetmem:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:3000"  # REST API
      - "9090:9090"  # MCP Server
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - CHROMADB_URL=http://chromadb:8000
      - CHROMADB_COLLECTION=zetmem_memories
      - EMBEDDING_SERVICE=sentence-transformers
      - EMBEDDING_SERVICE_URL=http://sentence-transformers:8080
      - LOG_LEVEL=info
    depends_on:
      chromadb:
        condition: service_healthy
      sentence-transformers:
        condition: service_healthy
    volumes:
      - ./prompts:/app/prompts
      - ./config:/app/config

volumes:
  chromadb_data:
  model_cache:
```

## Service-Specific Configurations

### ChromaDB Configuration

```bash
# Start ChromaDB with authentication
docker run -d \
  --name chromadb \
  -p 8000:8000 \
  -v chromadb_data:/chroma/chroma \
  -e ANONYMIZED_TELEMETRY=false \
  chromadb/chroma:latest

# Health check
curl http://localhost:8000/api/v1/heartbeat
```

### Sentence Transformers Configuration

```dockerfile
# docker/sentence-transformers/Dockerfile
FROM python:3.9-slim

RUN pip install sentence-transformers fastapi uvicorn

COPY server.py /app/server.py
WORKDIR /app

CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8080"]
```

```python
# docker/sentence-transformers/server.py
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import os

app = FastAPI()
model_name = os.getenv("MODEL_NAME", "all-MiniLM-L6-v2")
model = SentenceTransformer(model_name)

class EmbeddingRequest(BaseModel):
    sentences: list[str]
    model: str = None

@app.post("/embeddings")
async def create_embeddings(request: EmbeddingRequest):
    embeddings = model.encode(request.sentences)
    return {"embeddings": embeddings.tolist()}

@app.get("/health")
async def health():
    return {"status": "healthy", "model": model_name}
```

## Development Configuration

### Local Development (.env)

```bash
# Development Environment
NODE_ENV=development
LOG_LEVEL=debug

# Service URLs (Local)
CHROMADB_URL=http://localhost:8000
EMBEDDING_SERVICE_URL=http://localhost:8080

# Development Features
PROMPTS_HOT_RELOAD=true
PROMPTS_CACHE_ENABLED=false
DEBUG_MODE=true

# Test Configuration
TEST_WORKSPACE=test_workspace
TEST_TIMEOUT=60s
```

### Test Configuration (config/test.yaml)

```yaml
# Test Configuration
test:
  # Use in-memory storage for tests
  storage:
    type: memory
    
  # Mock external services
  mocks:
    litellm: true
    embeddings: true
    
  # Test data
  fixtures:
    memories_dir: ./test/fixtures/memories
    prompts_dir: ./test/fixtures/prompts
    
  # Timeouts
  timeouts:
    default: 5s
    integration: 30s
```

## Production Configuration

### Production Environment Variables

```bash
# Production Settings
NODE_ENV=production
LOG_LEVEL=warn
LOG_FORMAT=json

# High Availability
CHROMADB_REPLICAS=3
CHROMADB_CLUSTER_MODE=true

# Performance Tuning
CONNECTION_POOL_SIZE=200
MAX_CONCURRENT_REQUESTS=100
REQUEST_TIMEOUT=60s

# Security
API_KEY_ROTATION_DAYS=30
ENABLE_AUTH=true
CORS_ORIGINS=https://yourdomain.com

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9091
TRACE_ENABLED=true
TRACE_ENDPOINT=http://jaeger:14268/api/traces
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: zetmem-config
  namespace: zetmem
data:
  config.yaml: |
    services:
      chromadb:
        url: "http://chromadb-service:8000"
        collection: "zetmem_memories_prod"
      embedding:
        service: "sentence-transformers"
        url: "http://embedding-service:8080"
      litellm:
        default_model: "gpt-3.5-turbo"
        max_retries: 5
        timeout: 60s
```

## Monitoring Configuration

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'zetmem'
    static_configs:
      - targets: ['zetmem:9091']
    metrics_path: '/metrics'
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "ZetMem Service Metrics",
    "panels": [
      {
        "title": "LLM Call Duration",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, litellm_call_duration_seconds)"
          }
        ]
      },
      {
        "title": "Embedding Generation Rate",
        "targets": [
          {
            "expr": "rate(embedding_generated_total[5m])"
          }
        ]
      },
      {
        "title": "ChromaDB Query Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.99, chromadb_query_duration_seconds)"
          }
        ]
      }
    ]
  }
}
```

## Security Configuration

### API Key Management

```yaml
# secrets.yaml (encrypted)
secrets:
  openai_api_key: !vault |
    $ANSIBLE_VAULT;1.1;AES256
    66383439383437363...
  
  chromadb_auth:
    username: admin
    password: !vault |
      $ANSIBLE_VAULT;1.1;AES256
      35663836643364336...
```

### TLS Configuration

```yaml
tls:
  enabled: true
  cert_file: /etc/ssl/certs/zetmem.crt
  key_file: /etc/ssl/private/zetmem.key
  ca_file: /etc/ssl/certs/ca-bundle.crt
  min_version: "1.2"
  cipher_suites:
    - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
    - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
```