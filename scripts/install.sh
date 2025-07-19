#!/bin/bash
# A-MEM MCP Server Automated Installation Script
# This script sets up A-MEM for integration with Claude Code or Claude Desktop

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
INSTALL_LOG="$PROJECT_ROOT/install.log"

# Logging function
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$INSTALL_LOG"
}

# Print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
    log "SUCCESS: $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
    log "WARNING: $1"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
    log "ERROR: $1"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
    log "INFO: $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Get Claude configuration paths
get_claude_paths() {
    local os="$1"
    
    case "$os" in
        "macos")
            CLAUDE_CODE_CONFIG="$HOME/Library/Application Support/Code/User/settings.json"
            CLAUDE_DESKTOP_CONFIG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
            ;;
        "linux")
            CLAUDE_CODE_CONFIG="$HOME/.config/Code/User/settings.json"
            CLAUDE_DESKTOP_CONFIG="$HOME/.config/claude/claude_desktop_config.json"
            ;;
        "windows")
            CLAUDE_CODE_CONFIG="$APPDATA/Code/User/settings.json"
            CLAUDE_DESKTOP_CONFIG="$APPDATA/Claude/claude_desktop_config.json"
            ;;
        *)
            print_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    local missing_deps=()
    
    # Check Docker
    if ! command_exists docker; then
        missing_deps+=("docker")
    fi
    
    # Check Docker Compose
    if ! command_exists docker-compose; then
        missing_deps+=("docker-compose")
    fi
    
    # Check Git
    if ! command_exists git; then
        missing_deps+=("git")
    fi
    
    # Check Go
    if ! command_exists go; then
        missing_deps+=("go")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        echo ""
        echo "Please install the missing dependencies and run this script again."
        echo ""
        echo "Installation guides:"
        echo "- Docker: https://docs.docker.com/get-docker/"
        echo "- Docker Compose: https://docs.docker.com/compose/install/"
        echo "- Git: https://git-scm.com/downloads"
        echo "- Go: https://golang.org/dl/"
        exit 1
    fi
    
    print_status "All prerequisites are installed"
}

# Detect Claude installations
detect_claude() {
    local claude_installations=()

    # Check for Claude Code (VS Code extension)
    if [ -f "$CLAUDE_CODE_CONFIG" ]; then
        claude_installations+=("code")
    fi

    # Check for Claude Desktop
    if [ -f "$CLAUDE_DESKTOP_CONFIG" ] || [ -d "$(dirname "$CLAUDE_DESKTOP_CONFIG")" ]; then
        claude_installations+=("desktop")
    fi

    echo "${claude_installations[@]}"
}

# Print Claude detection results
print_claude_detection() {
    local claude_installations=("$@")

    print_info "Detecting Claude installations..."

    if [[ " ${claude_installations[@]} " =~ " code " ]]; then
        print_status "Found Claude Code configuration"
    fi

    if [[ " ${claude_installations[@]} " =~ " desktop " ]]; then
        print_status "Found Claude Desktop installation"
    fi

    if [ ${#claude_installations[@]} -eq 0 ]; then
        print_warning "No Claude installations detected"
        echo ""
        echo "Please install Claude Code or Claude Desktop first:"
        echo "- Claude Code: Install the Claude extension in VS Code"
        echo "- Claude Desktop: Download from https://claude.ai/download"
        echo ""
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Setup environment
setup_environment() {
    print_info "Setting up environment..."
    
    cd "$PROJECT_ROOT"
    
    # Copy environment template if it doesn't exist
    if [ ! -f ".env" ]; then
        cp .env.example .env
        print_status "Created .env file from template"
    else
        print_info ".env file already exists"
    fi
    
    # Prompt for API keys
    echo ""
    echo "🔑 API Key Configuration"
    echo "========================"
    echo ""
    echo "A-MEM works best with an OpenAI API key for enhanced embeddings and analysis."
    echo "You can also use it without an API key (with reduced functionality)."
    echo ""
    
    read -p "Do you have an OpenAI API key? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        read -p "Enter your OpenAI API key: " -s openai_key
        echo ""
        
        if [ -n "$openai_key" ]; then
            # Update .env file
            if grep -q "OPENAI_API_KEY=" .env; then
                sed -i.bak "s/OPENAI_API_KEY=.*/OPENAI_API_KEY=$openai_key/" .env
            else
                echo "OPENAI_API_KEY=$openai_key" >> .env
            fi
            print_status "OpenAI API key configured"
        fi
    else
        print_warning "Continuing without OpenAI API key (reduced functionality)"
    fi
}

# Cleanup existing containers and processes
cleanup_existing_containers() {
    print_info "Checking for existing A-MEM containers and processes..."

    # Kill any existing amem-server processes (e.g., started by Claude Desktop)
    print_status "Stopping any running A-MEM server processes..."

    # Kill processes using A-MEM ports
    for port in 8080 9092; do
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
    done

    # Also check for any amem-server processes by name
    pkill -f "amem-server" 2>/dev/null || true

    # Give processes a moment to terminate
    sleep 1

    local containers_found=false
    local compose_containers=""
    local orphaned_containers=""

    # Check for compose-managed containers
    if command -v docker-compose >/dev/null 2>&1; then
        compose_containers=$(docker-compose ps -q 2>/dev/null || true)
    fi

    # Check for orphaned containers with our naming pattern
    orphaned_containers=$(docker ps -aq --filter "name=amemcontext_augment" 2>/dev/null || true)

    # Determine if any containers were found
    if [ -n "$compose_containers" ] || [ -n "$orphaned_containers" ]; then
        containers_found=true
    fi

    if [ "$containers_found" = true ]; then
        echo ""
        print_warning "Found existing A-MEM containers:"

        # Show compose containers
        if [ -n "$compose_containers" ]; then
            echo "Compose-managed containers:"
            docker-compose ps 2>/dev/null || true
        fi

        # Show orphaned containers
        if [ -n "$orphaned_containers" ]; then
            echo "Orphaned containers:"
            docker ps -a --filter "name=amemcontext_augment" --format "table {{.Names}}\t{{.Status}}" 2>/dev/null || true
        fi

        echo ""
        echo "These containers will be stopped and removed."
        echo "Data volumes will be preserved."
        echo ""

        read -p "Clean up existing containers before installation? (Y/n): " -n 1 -r
        echo

        if [[ $REPLY =~ ^[Nn]$ ]]; then
            print_warning "Skipping container cleanup - this may cause port conflicts"
            return 0
        fi

        # Perform cleanup
        print_info "Cleaning up existing containers..."

        # Use docker-compose for project containers
        if [ -n "$compose_containers" ]; then
            print_status "Stopping compose-managed containers..."
            if docker-compose down --remove-orphans 2>/dev/null; then
                print_status "Compose containers cleaned up successfully"
            else
                print_warning "Some compose containers may not have been cleaned up"
            fi
        fi

        # Handle orphaned containers individually
        if [ -n "$orphaned_containers" ]; then
            print_status "Cleaning up orphaned containers..."
            if docker stop $orphaned_containers 2>/dev/null; then
                print_status "Orphaned containers stopped"
            fi
            if docker rm $orphaned_containers 2>/dev/null; then
                print_status "Orphaned containers removed"
            fi
        fi

        # Verify cleanup
        remaining=$(docker ps -aq --filter "name=amemcontext_augment" 2>/dev/null || true)
        if [ -z "$remaining" ]; then
            print_status "Container cleanup completed successfully"
        else
            print_warning "Some containers may still be running - check manually if needed"
        fi

        echo ""
    else
        print_status "No existing A-MEM containers found"
    fi

    print_status "A-MEM process and container cleanup completed"
}

# Start services
start_services() {
    print_info "Starting A-MEM services..."

    cd "$PROJECT_ROOT"

    # Clean up existing containers and processes first
    cleanup_existing_containers

    # Check for port conflicts
    local ports=(8004 8005 6382 9091 9092)
    local conflicts=()
    
    for port in "${ports[@]}"; do
        if lsof -i ":$port" >/dev/null 2>&1; then
            conflicts+=("$port")
        fi
    done
    
    if [ ${#conflicts[@]} -ne 0 ]; then
        print_warning "Port conflicts detected: ${conflicts[*]}"
        echo "These ports are required by A-MEM services."
        echo "Please stop services using these ports or modify docker-compose.yml"
        echo ""
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Start services
    print_info "Starting Docker services..."
    docker-compose up -d
    
    # Wait for services to be ready
    print_info "Waiting for services to start..."
    sleep 10
    
    # Check service health
    local services_ok=true
    
    # Check ChromaDB
    if curl -s http://localhost:8004/api/v1/heartbeat >/dev/null 2>&1; then
        print_status "ChromaDB is running"
    else
        print_error "ChromaDB failed to start"
        services_ok=false
    fi
    
    # Check Redis
    if docker-compose ps redis | grep -q "Up"; then
        print_status "Redis is running"
    else
        print_error "Redis failed to start"
        services_ok=false
    fi
    
    # Check Sentence Transformers (may take longer to start)
    print_info "Waiting for Sentence Transformers to download models..."
    local retries=0
    while [ $retries -lt 30 ]; do
        if curl -s http://localhost:8003/health >/dev/null 2>&1; then
            print_status "Sentence Transformers is running"
            break
        fi
        sleep 5
        retries=$((retries + 1))
    done
    
    if [ $retries -eq 30 ]; then
        print_warning "Sentence Transformers is still starting (this is normal for first run)"
    fi
    
    if [ "$services_ok" = false ]; then
        print_error "Some services failed to start. Check logs with: docker-compose logs"
        exit 1
    fi
}

# Build A-MEM server
build_server() {
    print_info "Building A-MEM server..."
    
    cd "$PROJECT_ROOT"
    
    # Download dependencies
    go mod tidy
    
    # Build server
    go build -o amem-server cmd/server/main.go
    
    if [ -f "amem-server" ]; then
        print_status "A-MEM server built successfully"
        
        # Make executable
        chmod +x amem-server
        
        # Get absolute path
        AMEM_SERVER_PATH="$(pwd)/amem-server"
        AMEM_CONFIG_PATH="$(pwd)/config/production.yaml"
    else
        print_error "Failed to build A-MEM server"
        exit 1
    fi
}

# Configure Claude integration
configure_claude() {
    local claude_type="$1"
    
    print_info "Configuring Claude $claude_type integration..."
    
    case "$claude_type" in
        "code")
            configure_claude_code
            ;;
        "desktop")
            configure_claude_desktop
            ;;
        *)
            print_error "Unknown Claude type: $claude_type"
            return 1
            ;;
    esac
}

# Configure Claude Code
configure_claude_code() {
    local config_file="$CLAUDE_CODE_CONFIG"
    local config_dir="$(dirname "$config_file")"
    
    # Create config directory if it doesn't exist
    mkdir -p "$config_dir"
    
    # Create or update settings.json
    if [ ! -f "$config_file" ]; then
        echo '{}' > "$config_file"
    fi
    
    # Create temporary config
    local temp_config=$(mktemp)
    
    # Read the actual API key from .env file
    local api_key=""
    if [ -f ".env" ] && grep -q "OPENAI_API_KEY=" .env; then
        api_key=$(grep "OPENAI_API_KEY=" .env | cut -d'=' -f2- | sed 's/^"//' | sed 's/"$//')
    fi

    # Read current config and add A-MEM MCP server (preserve existing servers)
    cat "$config_file" | jq --arg server_path "$AMEM_SERVER_PATH" --arg config_path "$AMEM_CONFIG_PATH" --arg api_key "$api_key" '
        .["claude.mcpServers"] = (.["claude.mcpServers"] // {}) + {
            "amem-augmented": {
                "command": $server_path,
                "args": ["-config", $config_path],
                "env": {
                    "OPENAI_API_KEY": $api_key
                }
            }
        }
    ' > "$temp_config"
    
    # Replace original config
    mv "$temp_config" "$config_file"
    
    print_status "Claude Code configuration updated"
    print_info "Config file: $config_file"
}

# Configure Claude Desktop
configure_claude_desktop() {
    local config_file="$CLAUDE_DESKTOP_CONFIG"
    local config_dir="$(dirname "$config_file")"
    
    # Create config directory if it doesn't exist
    mkdir -p "$config_dir"
    
    # Create or update claude_desktop_config.json
    if [ ! -f "$config_file" ]; then
        echo '{}' > "$config_file"
    fi
    
    # Create temporary config
    local temp_config=$(mktemp)
    
    # Read the actual API key from .env file
    local api_key=""
    if [ -f ".env" ] && grep -q "OPENAI_API_KEY=" .env; then
        api_key=$(grep "OPENAI_API_KEY=" .env | cut -d'=' -f2- | sed 's/^"//' | sed 's/"$//')
    fi

    # Read current config and add A-MEM MCP server (preserve existing servers)
    cat "$config_file" | jq --arg server_path "$AMEM_SERVER_PATH" --arg config_path "$AMEM_CONFIG_PATH" --arg api_key "$api_key" '
        .mcpServers = (.mcpServers // {}) + {
            "amem-augmented": {
                "command": $server_path,
                "args": ["-config", $config_path],
                "env": {
                    "OPENAI_API_KEY": $api_key
                }
            }
        }
    ' > "$temp_config"
    
    # Replace original config
    mv "$temp_config" "$config_file"
    
    print_status "Claude Desktop configuration updated"
    print_info "Config file: $config_file"
}

# Test installation
test_installation() {
    print_info "Testing installation..."
    
    cd "$PROJECT_ROOT"
    
    # Test server build
    if [ ! -f "amem-server" ]; then
        print_error "A-MEM server not found"
        return 1
    fi
    
    # Test basic functionality
    print_info "Running basic tests..."
    
    # Test MCP protocol
    echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}' | timeout 10 ./amem-server -config config/production.yaml > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        print_status "MCP protocol test passed"
    else
        print_warning "MCP protocol test failed (this may be normal)"
    fi
    
    # Test services
    if curl -s http://localhost:8002/api/v1/heartbeat >/dev/null 2>&1; then
        print_status "ChromaDB connection test passed"
    else
        print_error "ChromaDB connection test failed"
        return 1
    fi
    
    print_status "Installation test completed"
}

# Main installation function
main() {
    echo ""
    echo "🚀 A-MEM MCP Server Installation"
    echo "================================="
    echo ""
    echo "This script will install and configure A-MEM for use with Claude."
    echo ""
    
    # Initialize log
    echo "Installation started at $(date)" > "$INSTALL_LOG"
    
    # Detect OS
    local os=$(detect_os)
    print_info "Detected OS: $os"
    
    # Get Claude paths
    get_claude_paths "$os"
    
    # Check prerequisites
    check_prerequisites
    
    # Detect Claude installations
    local claude_installations=($(detect_claude))
    print_claude_detection "${claude_installations[@]}"
    
    # Setup environment
    setup_environment
    
    # Start services
    start_services
    
    # Build server
    build_server
    
    # Configure Claude integrations
    if [ ${#claude_installations[@]} -gt 0 ]; then
        echo ""
        echo "🔧 Claude Integration Configuration"
        echo "==================================="
        echo ""
        
        for installation in "${claude_installations[@]}"; do
            configure_claude "$installation"
        done
    fi
    
    # Test installation
    test_installation
    
    # Success message
    echo ""
    echo "🎉 Installation Complete!"
    echo "========================"
    echo ""
    print_status "A-MEM MCP Server has been successfully installed and configured"
    echo ""
    echo "Next steps:"
    echo "1. Restart Claude Code (VS Code) or Claude Desktop"
    echo "2. In Claude, you should now see these new tools:"
    echo "   - store_coding_memory"
    echo "   - retrieve_relevant_memories"
    echo "   - evolve_memory_network"
    echo ""
    echo "3. Test the integration by asking Claude:"
    echo "   'Can you store this code snippet in memory?'"
    echo ""
    echo "Configuration files:"
    echo "- A-MEM Server: $AMEM_SERVER_PATH"
    echo "- Configuration: $AMEM_CONFIG_PATH"
    echo "- Environment: $PROJECT_ROOT/.env"
    
    if [ ${#claude_installations[@]} -gt 0 ]; then
        echo "- Claude Config: ${claude_installations[*]}"
    fi
    
    echo ""
    echo "Logs: $INSTALL_LOG"
    echo ""
    echo "For troubleshooting, see: INSTALLATION_GUIDE.md"
    echo ""
    
    # Final service status
    echo "Service Status:"
    docker-compose ps
}

# Check if jq is installed (required for JSON manipulation)
if ! command_exists jq; then
    print_error "jq is required for configuration file manipulation"
    echo "Please install jq and run this script again:"
    echo ""
    case "$(detect_os)" in
        "macos")
            echo "  brew install jq"
            ;;
        "linux")
            echo "  sudo apt-get install jq  # Ubuntu/Debian"
            echo "  sudo yum install jq      # CentOS/RHEL"
            ;;
        "windows")
            echo "  Download from: https://stedolan.github.io/jq/download/"
            ;;
    esac
    exit 1
fi

# Run main installation
main "$@"
