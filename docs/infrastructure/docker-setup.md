# Docker Setup Guide

## Overview

This guide provides detailed information about the Docker configuration for the A-MEM MCP Server, including the multi-stage build process, service orchestration, and optimization techniques.

## Docker Architecture

### Multi-Stage Build Process

The A-MEM server uses a multi-stage Docker build to optimize the final image size:

```dockerfile
# Stage 1: Builder
FROM golang:1.23-alpine AS builder
- Compiles Go application
- Downloads dependencies
- Produces single binary

# Stage 2: Runtime
FROM alpine:latest
- Minimal runtime environment
- Only includes necessary files
- Runs as non-root user
```

**Benefits:**
- Reduced image size (from ~1.2GB to ~50MB)
- Improved security (no build tools in production)
- Faster deployment and scaling

### Service Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Docker Network: amem-network             │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │  amem-server    │  │    chromadb     │  │    redis    │ │
│  │  Go API Server  │  │  Vector Store   │  │  Cache/Queue│ │
│  │  Port: 8080     │  │  Port: 8004     │  │  Port: 6382 │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
│                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   rabbitmq      │  │sentence-transf  │  │ prometheus  │ │
│  │  Message Queue  │  │  Embeddings     │  │  Monitoring │ │
│  │  Port: 5672     │  │  Port: 8005     │  │  Port: 9091 │ │
│  └─────────────────┘  └─────────────────┘  └─────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Dockerfile Breakdown

### Build Stage

```dockerfile
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy and download dependencies (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o amem-server ./cmd/server
```

**Key optimizations:**
- Dependencies cached separately for faster rebuilds
- CGO disabled for static binary
- Alpine base for minimal size

### Runtime Stage

```dockerfile
FROM alpine:latest

# Security: Install only runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Security: Create non-root user
RUN addgroup -g 1001 -S amem && \
    adduser -u 1001 -S amem -G amem

WORKDIR /app

# Copy artifacts with proper ownership
COPY --from=builder /app/amem-server .
COPY --chown=amem:amem config/ ./config/
COPY --chown=amem:amem prompts/ ./prompts/

# Security: Run as non-root
USER amem

EXPOSE 8080

# Health check configuration
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD pgrep amem-server || exit 1

CMD ["./amem-server", "-config", "./config/docker.yaml"]
```

## Docker Compose Configuration

### Service Definitions

#### 1. A-MEM Server

```yaml
amem-server:
  build: .
  ports:
    - "8080:8080"    # API port
    - "9092:9090"    # Metrics port
  environment:
    - AMEM_ENV=development
    - OPENAI_API_KEY=${OPENAI_API_KEY}
    - CHROMADB_HOST=http://chromadb:8000
  volumes:
    - ./config:/app/config      # Configuration
    - ./prompts:/app/prompts    # Prompt templates
    - ./data:/app/data          # Persistent data
  depends_on:
    - chromadb
    - redis
  restart: unless-stopped
```

#### 2. ChromaDB (Vector Database)

```yaml
chromadb:
  image: chromadb/chroma:latest
  ports:
    - "8004:8000"
  volumes:
    - chromadb_data:/chroma/chroma
  environment:
    - CHROMA_SERVER_HOST=0.0.0.0
    - CHROMA_SERVER_PORT=8000
```

#### 3. Redis (Cache & Queue)

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6382:6379"
  volumes:
    - redis_data:/data
  command: redis-server --appendonly yes
```

#### 4. Sentence Transformers (Embeddings)

```yaml
sentence-transformers:
  build:
    context: ./docker/sentence-transformers
  ports:
    - "8005:8000"
  environment:
    - MODEL_NAME=all-MiniLM-L6-v2
    - MAX_BATCH_SIZE=32
  volumes:
    - sentence_transformers_cache:/root/.cache
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
    interval: 30s
```

#### 5. Monitoring Stack

```yaml
prometheus:
  image: prom/prometheus:latest
  ports:
    - "9091:9090"
  volumes:
    - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_data:/prometheus
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.retention.time=200h'
```

### Volume Management

```yaml
volumes:
  chromadb_data:      # Vector embeddings
  redis_data:         # Cache and queue data
  rabbitmq_data:      # Message queue persistence
  prometheus_data:    # Metrics history
  sentence_transformers_cache:  # Model cache
```

## Building and Running

### Development Build

```bash
# Build all services
docker compose build

# Build specific service
docker compose build amem-server

# Build with no cache
docker compose build --no-cache
```

### Running Services

```bash
# Start all services
docker compose up -d

# Start specific services
docker compose up -d amem-server chromadb redis

# View logs
docker compose logs -f amem-server

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v
```

### Development Workflow

1. **Hot Reload Configuration**
   ```yaml
   volumes:
     - ./config:/app/config:ro
     - ./prompts:/app/prompts:ro
   ```

2. **Override for Development**
   ```bash
   # Create docker-compose.override.yml
   services:
     amem-server:
       environment:
         - AMEM_LOG_LEVEL=debug
         - AMEM_ENV=development
   ```

3. **Local Development**
   ```bash
   # Run dependencies only
   docker compose up -d chromadb redis prometheus
   
   # Run server locally
   go run ./cmd/server -config ./config/development.yaml
   ```

## Performance Optimization

### 1. Build Optimization

```dockerfile
# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Then copy source
COPY . .
```

### 2. Layer Caching

Structure Dockerfile to maximize cache hits:
- Static dependencies first
- Frequently changing files last
- Separate build and runtime stages

### 3. Image Size Reduction

```dockerfile
# Use Alpine base
FROM alpine:latest

# Install only required packages
RUN apk --no-cache add ca-certificates

# Remove unnecessary files
RUN rm -rf /var/cache/apk/*
```

### 4. Resource Limits

```yaml
services:
  amem-server:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

## Security Best Practices

### 1. Non-Root User

```dockerfile
# Create and use non-root user
RUN addgroup -g 1001 -S amem && \
    adduser -u 1001 -S amem -G amem
USER amem
```

### 2. Read-Only Filesystem

```yaml
services:
  amem-server:
    read_only: true
    tmpfs:
      - /tmp
    volumes:
      - ./data:/app/data:rw
```

### 3. Network Isolation

```yaml
networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true
```

### 4. Secrets Management

```yaml
secrets:
  openai_api_key:
    file: ./secrets/openai_api_key.txt

services:
  amem-server:
    secrets:
      - openai_api_key
```

## Troubleshooting

### Common Issues

1. **Build Failures**
   ```bash
   # Clear build cache
   docker builder prune
   
   # Check build context size
   du -sh .
   ```

2. **Container Won't Start**
   ```bash
   # Check logs
   docker compose logs amem-server
   
   # Debug interactively
   docker compose run --entrypoint sh amem-server
   ```

3. **Network Issues**
   ```bash
   # List networks
   docker network ls
   
   # Inspect network
   docker network inspect amem-network
   ```

4. **Volume Permissions**
   ```bash
   # Fix ownership
   docker compose exec amem-server chown -R amem:amem /app/data
   ```

### Health Checks

Monitor service health:

```bash
# Check all services
docker compose ps

# Check specific service health
docker inspect amem-server | jq '.[0].State.Health'

# Manual health check
curl http://localhost:8080/health
```

## Advanced Configuration

### Custom Networks

```yaml
networks:
  amem-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### External Services

```yaml
services:
  amem-server:
    external_links:
      - external-redis:redis
    networks:
      - amem-network
      - external-network
```

### Multi-Environment Setup

```bash
# Production
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Staging
docker compose -f docker-compose.yml -f docker-compose.staging.yml up -d
```

## Maintenance

### Image Updates

```bash
# Update base images
docker compose pull

# Rebuild with latest dependencies
docker compose build --pull
```

### Cleanup

```bash
# Remove unused images
docker image prune -a

# Remove unused volumes
docker volume prune

# Complete cleanup
docker system prune -a --volumes
```

## Next Steps

- Review [Configuration Guide](./configuration.md) for detailed configuration options
- See [Deployment Guide](./deployment.md) for production deployment
- Check [Quick Start](../deployment/quick-start.md) for getting started quickly