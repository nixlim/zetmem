# A-MEM MCP Server Deployment Guide

## Overview

The A-MEM MCP Server is deployed as a containerized microservices architecture using Docker and Docker Compose. This guide provides comprehensive instructions for deploying the system in various environments.

## Architecture Overview

The deployment consists of the following components:

```
┌─────────────────────────────────────────────────────────────┐
│                        Load Balancer                         │
│                         (Port 8080)                          │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                      A-MEM Server                            │
│                  (Go Application)                            │
│              ┌─────────────────────┐                        │
│              │   HTTP API (8080)   │                        │
│              │  Metrics API (9092) │                        │
│              └─────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
        │               │               │               │
        ▼               ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐
│  ChromaDB   │ │    Redis    │ │  RabbitMQ   │ │Sentence   │
│  (Vector)   │ │   (Cache)   │ │   (Queue)   │ │Transform  │
│  Port 8004  │ │  Port 6382  │ │  Port 5672  │ │Port 8005  │
└─────────────┘ └─────────────┘ └─────────────┘ └───────────┘
```

## Prerequisites

### System Requirements

- **CPU**: Minimum 4 cores, recommended 8+ cores
- **RAM**: Minimum 8GB, recommended 16GB+
- **Storage**: Minimum 20GB free space for containers and data
- **OS**: Linux (Ubuntu 20.04+ recommended), macOS, or Windows with WSL2

### Software Requirements

1. **Docker Engine** (v20.10+)
   ```bash
   # Ubuntu/Debian
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh
   
   # macOS
   brew install --cask docker
   ```

2. **Docker Compose** (v2.0+)
   ```bash
   # Usually included with Docker Desktop
   docker compose version
   ```

3. **Go** (v1.23+) - Only for local development
   ```bash
   # Download from https://golang.org/dl/
   go version
   ```

## Quick Start Deployment

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/amem-mcp-server.git
cd amem-mcp-server
```

### 2. Configure Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit the .env file with your API keys
nano .env
```

**Required environment variables:**
- `OPENAI_API_KEY`: Your OpenAI API key
- `SECRET_KEY`: Generate with `openssl rand -hex 32`

### 3. Start the Services

```bash
# Start all services
docker compose up -d

# Check service status
docker compose ps

# View logs
docker compose logs -f
```

### 4. Verify Deployment

```bash
# Check health endpoint
curl http://localhost:8080/health

# Check metrics endpoint
curl http://localhost:9092/metrics

# Access services:
# - API: http://localhost:8080
# - ChromaDB: http://localhost:8004
# - RabbitMQ Management: http://localhost:15672 (amem/amem_password)
# - Prometheus: http://localhost:9091
```

## Production Deployment

### 1. Environment-Specific Configuration

For production, use the production configuration:

```bash
# Use production config
cp config/production.yaml config/docker.yaml

# Update docker-compose for production
cp docker-compose.prod.yml docker-compose.override.yml
```

### 2. Security Hardening

1. **Update default passwords**:
   ```yaml
   # docker-compose.override.yml
   services:
     rabbitmq:
       environment:
         - RABBITMQ_DEFAULT_PASS=${RABBITMQ_PASSWORD}
   ```

2. **Enable TLS/SSL**:
   - Use a reverse proxy (Nginx/Traefik) with Let's Encrypt
   - Configure internal service communication over TLS

3. **Network isolation**:
   ```yaml
   networks:
     frontend:
       driver: bridge
     backend:
       driver: bridge
       internal: true
   ```

### 3. Resource Limits

Set appropriate resource limits in production:

```yaml
# docker-compose.override.yml
services:
  amem-server:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G
        reservations:
          cpus: '2'
          memory: 2G
```

### 4. Persistence and Backups

Ensure data persistence with named volumes:

```yaml
volumes:
  chromadb_data:
    driver: local
    driver_opts:
      type: none
      device: /data/chromadb
      o: bind
```

### 5. Monitoring Setup

Configure Prometheus for production monitoring:

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  external_labels:
    environment: 'production'
    
alerting:
  alertmanagers:
    - static_configs:
        - targets: ['alertmanager:9093']
```

## Deployment Options

### Option 1: Docker Compose (Recommended for Single Node)

```bash
# Production deployment
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Option 2: Kubernetes Deployment

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/deployments/
kubectl apply -f k8s/services/
```

### Option 3: Cloud Platform Deployment

#### AWS ECS
```bash
# Build and push to ECR
aws ecr get-login-password | docker login --username AWS --password-stdin $ECR_REGISTRY
docker build -t amem-server .
docker tag amem-server:latest $ECR_REGISTRY/amem-server:latest
docker push $ECR_REGISTRY/amem-server:latest

# Deploy with ECS CLI
ecs-cli compose up
```

#### Google Cloud Run
```bash
# Build and deploy
gcloud builds submit --tag gcr.io/$PROJECT_ID/amem-server
gcloud run deploy amem-server --image gcr.io/$PROJECT_ID/amem-server
```

## Scaling Considerations

### Horizontal Scaling

1. **API Server**: Stateless, can scale horizontally
   ```yaml
   services:
     amem-server:
       deploy:
         replicas: 3
   ```

2. **ChromaDB**: Use external managed service for production
3. **Redis**: Configure Redis Cluster or use managed service
4. **RabbitMQ**: Set up RabbitMQ cluster for HA

### Vertical Scaling

Adjust resource allocations based on workload:
- Memory-intensive: Increase RAM for embedding operations
- CPU-intensive: More cores for concurrent request handling

## Troubleshooting

### Common Issues

1. **Container fails to start**
   ```bash
   docker compose logs amem-server
   # Check for missing environment variables or config issues
   ```

2. **Connection errors**
   ```bash
   # Check network connectivity
   docker compose exec amem-server ping chromadb
   ```

3. **Performance issues**
   ```bash
   # Check resource usage
   docker stats
   ```

### Health Checks

All services include health checks:
- A-MEM Server: `http://localhost:8080/health`
- ChromaDB: `http://localhost:8004/api/v1/heartbeat`
- Sentence Transformers: `http://localhost:8005/health`

## Maintenance

### Updates

```bash
# Pull latest images
docker compose pull

# Restart services with zero downtime
docker compose up -d --no-deps --build amem-server
```

### Backups

```bash
# Backup volumes
docker run --rm -v amem_chromadb_data:/data -v $(pwd):/backup alpine tar czf /backup/chromadb-backup.tar.gz -C /data .
```

### Log Management

```bash
# Configure log rotation
# docker-compose.yml
services:
  amem-server:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Security Best Practices

1. **Use secrets management**: Never commit sensitive data
2. **Enable authentication**: Configure API keys for all services
3. **Network segmentation**: Isolate services appropriately
4. **Regular updates**: Keep all images and dependencies updated
5. **Monitoring**: Set up alerts for anomalous behavior

## Next Steps

- Review [Docker Setup Guide](./docker-setup.md) for detailed Docker configuration
- Check [Configuration Guide](./configuration.md) for service configuration options
- See [Quick Start](../deployment/quick-start.md) for development setup