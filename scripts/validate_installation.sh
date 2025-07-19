#!/bin/bash
# A-MEM Installation Validation Script
# Validates that A-MEM is properly installed and configured

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

# Print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
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

# Validate prerequisites
validate_prerequisites() {
    print_info "Validating prerequisites..."
    
    local all_good=true
    
    # Check Docker
    if command_exists docker; then
        print_status "Docker is installed"
    else
        print_error "Docker is not installed"
        all_good=false
    fi
    
    # Check Docker Compose
    if command_exists docker-compose; then
        print_status "Docker Compose is installed"
    else
        print_error "Docker Compose is not installed"
        all_good=false
    fi
    
    # Check Go
    if command_exists go; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_status "Go is installed (version $go_version)"
    else
        print_error "Go is not installed"
        all_good=false
    fi
    
    if [ "$all_good" = false ]; then
        print_error "Prerequisites validation failed"
        return 1
    fi
    
    print_status "All prerequisites are satisfied"
}

# Validate A-MEM server
validate_server() {
    print_info "Validating A-MEM server..."
    
    cd "$PROJECT_ROOT"
    
    # Check if server exists
    if [ -f "amem-server" ]; then
        print_status "A-MEM server binary exists"
    else
        print_error "A-MEM server binary not found"
        print_info "Run 'make build' to build the server"
        return 1
    fi
    
    # Check if executable
    if [ -x "amem-server" ]; then
        print_status "A-MEM server is executable"
    else
        print_error "A-MEM server is not executable"
        print_info "Run 'chmod +x amem-server' to fix"
        return 1
    fi
    
    # Check configuration files
    if [ -f "config/production.yaml" ]; then
        print_status "Production configuration exists"
    else
        print_error "Production configuration not found"
        return 1
    fi
    
    if [ -f "config/development.yaml" ]; then
        print_status "Development configuration exists"
    else
        print_warning "Development configuration not found"
    fi
    
    print_status "A-MEM server validation passed"
}

# Validate Docker services
validate_services() {
    print_info "Validating Docker services..."
    
    cd "$PROJECT_ROOT"
    
    # Check if docker-compose.yml exists
    if [ ! -f "docker-compose.yml" ]; then
        print_error "docker-compose.yml not found"
        return 1
    fi
    
    # Validate docker-compose configuration
    if docker-compose config >/dev/null 2>&1; then
        print_status "Docker Compose configuration is valid"
    else
        print_error "Docker Compose configuration is invalid"
        return 1
    fi
    
    # Check if services are running
    local running_services=$(docker-compose ps --services --filter "status=running" 2>/dev/null || echo "")
    
    if echo "$running_services" | grep -q "chromadb"; then
        print_status "ChromaDB service is running"
    else
        print_warning "ChromaDB service is not running"
    fi
    
    if echo "$running_services" | grep -q "redis"; then
        print_status "Redis service is running"
    else
        print_warning "Redis service is not running"
    fi
    
    if echo "$running_services" | grep -q "sentence-transformers"; then
        print_status "Sentence Transformers service is running"
    else
        print_warning "Sentence Transformers service is not running"
    fi
    
    print_status "Docker services validation completed"
}

# Test service connectivity
test_connectivity() {
    print_info "Testing service connectivity..."
    
    # Test ChromaDB
    if curl -s http://localhost:8004/api/v1/heartbeat >/dev/null 2>&1; then
        print_status "ChromaDB is accessible"
    else
        print_warning "ChromaDB is not accessible (may not be running)"
    fi
    
    # Test Sentence Transformers
    if curl -s http://localhost:8005/health >/dev/null 2>&1; then
        print_status "Sentence Transformers is accessible"
    else
        print_warning "Sentence Transformers is not accessible (may still be starting)"
    fi
    
    # Test Redis (basic connection)
    if docker-compose exec -T redis redis-cli ping >/dev/null 2>&1; then
        print_status "Redis is accessible"
    else
        print_warning "Redis is not accessible"
    fi
    
    print_status "Connectivity tests completed"
}

# Test MCP protocol
test_mcp_protocol() {
    print_info "Testing MCP protocol..."
    
    cd "$PROJECT_ROOT"
    
    # Test basic MCP initialization
    local test_request='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}'
    
    local response=$(echo "$test_request" | timeout 10 ./amem-server -config config/production.yaml 2>/dev/null || echo "")
    
    if echo "$response" | grep -q '"result"'; then
        print_status "MCP protocol test passed"
    else
        print_warning "MCP protocol test failed (this may be normal if services aren't running)"
    fi
    
    print_status "MCP protocol test completed"
}

# Validate Claude configuration
validate_claude_config() {
    print_info "Validating Claude configuration..."
    
    local claude_found=false
    
    # Check Claude Code configuration
    if [ -f "$CLAUDE_CODE_CONFIG" ]; then
        print_status "Claude Code configuration file exists"
        
        # Check if A-MEM is configured
        if grep -q "amem-augmented" "$CLAUDE_CODE_CONFIG" 2>/dev/null; then
            print_status "A-MEM is configured in Claude Code"
            claude_found=true
        else
            print_warning "A-MEM is not configured in Claude Code"
        fi
        
        # Validate JSON syntax
        if cat "$CLAUDE_CODE_CONFIG" | jq . >/dev/null 2>&1; then
            print_status "Claude Code configuration JSON is valid"
        else
            print_error "Claude Code configuration JSON is invalid"
        fi
    else
        print_info "Claude Code configuration not found"
    fi
    
    # Check Claude Desktop configuration
    if [ -f "$CLAUDE_DESKTOP_CONFIG" ]; then
        print_status "Claude Desktop configuration file exists"
        
        # Check if A-MEM is configured
        if grep -q "amem-augmented" "$CLAUDE_DESKTOP_CONFIG" 2>/dev/null; then
            print_status "A-MEM is configured in Claude Desktop"
            claude_found=true
        else
            print_warning "A-MEM is not configured in Claude Desktop"
        fi
        
        # Validate JSON syntax
        if cat "$CLAUDE_DESKTOP_CONFIG" | jq . >/dev/null 2>&1; then
            print_status "Claude Desktop configuration JSON is valid"
        else
            print_error "Claude Desktop configuration JSON is invalid"
        fi
    else
        print_info "Claude Desktop configuration not found"
    fi
    
    if [ "$claude_found" = true ]; then
        print_status "A-MEM is configured for at least one Claude installation"
    else
        print_warning "A-MEM is not configured for any Claude installation"
        print_info "Run the installer or manually configure Claude"
    fi
    
    print_status "Claude configuration validation completed"
}

# Validate environment
validate_environment() {
    print_info "Validating environment configuration..."
    
    cd "$PROJECT_ROOT"
    
    # Check .env file
    if [ -f ".env" ]; then
        print_status ".env file exists"
        
        # Check for OpenAI API key
        if grep -q "OPENAI_API_KEY=" .env && ! grep -q "OPENAI_API_KEY=$" .env && ! grep -q "OPENAI_API_KEY=your_" .env; then
            print_status "OpenAI API key is configured"
        else
            print_warning "OpenAI API key is not configured (reduced functionality)"
        fi
    else
        print_warning ".env file not found"
        print_info "Copy .env.example to .env and configure"
    fi
    
    print_status "Environment validation completed"
}

# Generate validation report
generate_report() {
    print_info "Generating validation report..."
    
    local report_file="$PROJECT_ROOT/validation_report.txt"
    
    cat > "$report_file" << EOF
A-MEM Installation Validation Report
Generated: $(date)

System Information:
- OS: $(detect_os)
- Docker: $(docker --version 2>/dev/null || echo "Not installed")
- Docker Compose: $(docker-compose --version 2>/dev/null || echo "Not installed")
- Go: $(go version 2>/dev/null || echo "Not installed")

A-MEM Server:
- Binary exists: $([ -f "$PROJECT_ROOT/amem-server" ] && echo "Yes" || echo "No")
- Executable: $([ -x "$PROJECT_ROOT/amem-server" ] && echo "Yes" || echo "No")
- Production config: $([ -f "$PROJECT_ROOT/config/production.yaml" ] && echo "Yes" || echo "No")

Docker Services:
- ChromaDB: $(curl -s http://localhost:8002/api/v1/heartbeat >/dev/null 2>&1 && echo "Running" || echo "Not accessible")
- Redis: $(docker-compose exec -T redis redis-cli ping >/dev/null 2>&1 && echo "Running" || echo "Not accessible")
- Sentence Transformers: $(curl -s http://localhost:8003/health >/dev/null 2>&1 && echo "Running" || echo "Not accessible")

Claude Configuration:
- Claude Code config: $([ -f "$CLAUDE_CODE_CONFIG" ] && echo "Exists" || echo "Not found")
- Claude Desktop config: $([ -f "$CLAUDE_DESKTOP_CONFIG" ] && echo "Exists" || echo "Not found")

Environment:
- .env file: $([ -f "$PROJECT_ROOT/.env" ] && echo "Exists" || echo "Not found")
- OpenAI API key: $(grep -q "OPENAI_API_KEY=" "$PROJECT_ROOT/.env" 2>/dev/null && ! grep -q "OPENAI_API_KEY=$" "$PROJECT_ROOT/.env" 2>/dev/null && echo "Configured" || echo "Not configured")

EOF
    
    print_status "Validation report saved to: $report_file"
}

# Main validation function
main() {
    echo ""
    echo "🔍 A-MEM Installation Validation"
    echo "================================="
    echo ""
    
    # Detect OS and set paths
    local os=$(detect_os)
    get_claude_paths "$os"
    
    # Run validation steps
    local validation_passed=true
    
    validate_prerequisites || validation_passed=false
    echo ""
    
    validate_server || validation_passed=false
    echo ""
    
    validate_services
    echo ""
    
    test_connectivity
    echo ""
    
    test_mcp_protocol
    echo ""
    
    validate_claude_config
    echo ""
    
    validate_environment
    echo ""
    
    # Generate report
    generate_report
    echo ""
    
    # Final status
    if [ "$validation_passed" = true ]; then
        echo "🎉 Validation Complete!"
        echo "======================"
        print_status "A-MEM installation appears to be working correctly"
        echo ""
        echo "Next steps:"
        echo "1. Restart Claude Code (VS Code) or Claude Desktop"
        echo "2. Test the integration by asking Claude about available tools"
        echo "3. Try storing and retrieving memories"
    else
        echo "⚠️  Validation Issues Found"
        echo "=========================="
        print_warning "Some validation checks failed"
        echo ""
        echo "Common fixes:"
        echo "1. Run 'make build' to build the server"
        echo "2. Run 'docker-compose up -d' to start services"
        echo "3. Run './scripts/install.sh' to reconfigure"
        echo "4. Check the validation report for details"
    fi
    
    echo ""
    echo "For help, see:"
    echo "- Installation Guide: INSTALLATION_GUIDE.md"
    echo "- Configuration Guide: MCP_CONFIGURATION_GUIDE.md"
    echo "- Validation Report: validation_report.txt"
}

# Check if jq is available for JSON validation
if ! command_exists jq; then
    print_warning "jq not found - JSON validation will be skipped"
    print_info "Install jq for complete validation"
fi

# Run validation
main "$@"
