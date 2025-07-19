#!/bin/bash
# A-MEM MCP Server Setup Script

set -e

echo "🚀 A-MEM MCP Server Setup"
echo "========================="
echo ""

# Check prerequisites
echo "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.23+ from https://golang.org/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go $GO_VERSION found"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker from https://docker.com/"
    exit 1
fi
echo "✅ Docker found"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi
echo "✅ Docker Compose found"

echo ""

# Setup environment
echo "Setting up environment..."

if [ ! -f .env ]; then
    cp .env.example .env
    echo "✅ Created .env file from template"
    echo "⚠️  Please edit .env file and add your API keys"
else
    echo "✅ .env file already exists"
fi

# Create data directory
mkdir -p data
echo "✅ Created data directory"

# Download dependencies
echo ""
echo "Downloading Go dependencies..."
go mod tidy
echo "✅ Dependencies downloaded"

# Build the server
echo ""
echo "Building server..."
make build
echo "✅ Server built successfully"

# Start supporting services
echo ""
echo "Starting supporting services (ChromaDB, Redis)..."
docker-compose up -d chromadb redis
echo "✅ Supporting services started"

# Wait for services to be ready
echo ""
echo "Waiting for services to be ready..."
sleep 5

# Check ChromaDB
if curl -s http://localhost:8000/api/v1/heartbeat > /dev/null; then
    echo "✅ ChromaDB is ready"
else
    echo "⚠️  ChromaDB may not be ready yet. Please wait a moment and try again."
fi

# Run tests
echo ""
echo "Running basic tests..."
if go test ./pkg/models ./pkg/config; then
    echo "✅ Basic tests passed"
else
    echo "⚠️  Some tests failed. Check the output above."
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your API keys (especially OPENAI_API_KEY)"
echo "2. Run 'make dev' to start the development server"
echo "3. Test with 'python3 scripts/test_mcp.py' (requires server to be running)"
echo ""
echo "Available commands:"
echo "  make dev          - Start development server"
echo "  make docker-run   - Run with Docker Compose"
echo "  make test         - Run all tests"
echo "  make help         - Show all available commands"
echo ""
echo "Documentation:"
echo "  README.md         - Full documentation"
echo "  config/           - Configuration examples"
echo "  prompts/          - LLM prompt templates"
