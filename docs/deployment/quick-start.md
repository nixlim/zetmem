# Quick Start Guide

Get the A-MEM MCP Server running in under 5 minutes with this quick start guide.

## Prerequisites

Ensure you have the following installed:
- Docker Desktop (includes Docker and Docker Compose)
- Git
- An OpenAI API key

## 1. Clone and Setup (1 minute)

```bash
# Clone the repository
git clone https://github.com/your-org/amem-mcp-server.git
cd amem-mcp-server

# Copy environment template
cp .env.example .env
```

## 2. Configure API Key (1 minute)

Edit the `.env` file and add your OpenAI API key:

```bash
# Using your favorite editor
nano .env

# Or use sed on macOS/Linux
sed -i '' 's/your_openai_api_key_here/YOUR_ACTUAL_API_KEY/' .env

# Generate a secret key
echo "SECRET_KEY=$(openssl rand -hex 32)" >> .env
```

## 3. Start Services (2 minutes)

```bash
# Start all services in background
docker compose up -d

# Watch the logs (optional)
docker compose logs -f
```

## 4. Verify Installation (1 minute)

```bash
# Check if services are running
docker compose ps

# Test the API endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","version":"1.0.0","services":{"chromadb":"connected","redis":"connected"}}
```

## 🎉 You're Ready!

The A-MEM MCP Server is now running with:
- **API Server**: http://localhost:8080
- **ChromaDB UI**: http://localhost:8004
- **Prometheus Metrics**: http://localhost:9091
- **RabbitMQ Management**: http://localhost:15672 (user: amem, pass: amem_password)

## Quick Test

### Create a Memory

```bash
curl -X POST http://localhost:8080/api/v1/remember \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "content": "The user prefers dark mode and uses VSCode",
    "agent_id": "assistant-001",
    "metadata": {
      "category": "preferences",
      "confidence": 0.95
    }
  }'
```

### Recall Memories

```bash
curl -X POST http://localhost:8080/api/v1/recall \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "query": "What are the user preferences?",
    "agent_id": "assistant-001",
    "limit": 5
  }'
```

## Quick Commands Reference

```bash
# Start services
docker compose up -d

# Stop services
docker compose down

# View logs
docker compose logs -f [service-name]

# Restart a service
docker compose restart amem-server

# Update and restart
docker compose pull && docker compose up -d

# Check service health
curl http://localhost:8080/health
```

## Common Quick Fixes

### Port Already in Use

```bash
# Change ports in docker-compose.yml or use:
AMEM_PORT=8081 docker compose up -d
```

### API Key Issues

```bash
# Verify environment variable
docker compose exec amem-server env | grep OPENAI

# Restart after changing .env
docker compose restart amem-server
```

### Service Won't Start

```bash
# Check logs for specific service
docker compose logs amem-server

# Run in foreground to see errors
docker compose up amem-server
```

## Quick Configuration Changes

### Enable Debug Logging

Add to `.env`:
```bash
AMEM_LOG_LEVEL=debug
```

Then restart:
```bash
docker compose restart amem-server
```

### Change Memory Model

Add to `.env`:
```bash
LITELLM_DEFAULT_MODEL=gpt-3.5-turbo
```

### Adjust Rate Limits

Add to `.env`:
```bash
LITELLM_RATE_LIMIT=120
```

## Quick Development Setup

For local development with hot reload:

```bash
# Run only dependencies
docker compose up -d chromadb redis prometheus

# Run server locally with hot reload
go run ./cmd/server -config ./config/development.yaml
```

## Quick Production Deployment

```bash
# Use production configuration
cp config/production.yaml config/docker.yaml

# Set production environment
echo "AMEM_ENV=production" >> .env

# Start with production settings
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## Next Steps in 5 Minutes

1. **Integrate with Claude Desktop**
   ```bash
   # Add to Claude Desktop config
   cat config/claude_desktop_config.json
   ```

2. **Enable Memory Evolution**
   ```bash
   # Already enabled by default!
   # Check evolution status:
   curl http://localhost:8080/api/v1/evolution/status
   ```

3. **Set Up Monitoring**
   ```bash
   # Prometheus already running!
   # Add Grafana:
   docker compose -f docker-compose.monitoring.yml up -d
   ```

## Quick Links

- **API Documentation**: http://localhost:8080/swagger
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:9092/metrics
- **ChromaDB Collections**: http://localhost:8004/api/v1/collections

## Troubleshooting in 30 Seconds

```bash
# Everything in one command:
docker compose down -v && docker compose up -d && docker compose logs -f
```

If that doesn't work:
1. Check `.env` file has API keys
2. Ensure ports 8080, 8004, 6382 are free
3. Check Docker has enough resources (4GB RAM minimum)

## Getting Help

- **Logs**: `docker compose logs [service-name]`
- **Shell Access**: `docker compose exec amem-server sh`
- **Documentation**: See `/docs` directory
- **Issues**: GitHub Issues page

---

**Congratulations!** You now have a fully functional A-MEM MCP Server running locally. The system is ready to store and retrieve memories for your AI agents.