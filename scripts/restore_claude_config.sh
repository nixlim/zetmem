#!/bin/bash
# Script to restore Claude Desktop config if it was overwritten by the installer

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLAUDE_DESKTOP_CONFIG="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
BACKUP_CONFIG="${CLAUDE_DESKTOP_CONFIG}.backup"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

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

echo ""
echo "🔧 Claude Desktop Config Restoration"
echo "===================================="
echo ""

# Check if backup exists
if [ -f "$BACKUP_CONFIG" ]; then
    print_info "Found backup config at: $BACKUP_CONFIG"
    
    echo ""
    echo "Current config:"
    cat "$CLAUDE_DESKTOP_CONFIG" | jq . 2>/dev/null || echo "Invalid JSON or file not found"
    
    echo ""
    echo "Backup config:"
    cat "$BACKUP_CONFIG" | jq . 2>/dev/null || echo "Invalid JSON in backup"
    
    echo ""
    read -p "Do you want to restore from backup? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cp "$BACKUP_CONFIG" "$CLAUDE_DESKTOP_CONFIG"
        print_status "Config restored from backup"
        
        # Now add A-MEM to the restored config
        echo ""
        read -p "Do you want to add A-MEM to the restored config? (y/N): " -n 1 -r
        echo
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ZETMEM_SERVER_PATH="$PROJECT_ROOT/zetmem-server"
            ZETMEM_CONFIG_PATH="$PROJECT_ROOT/config/production.yaml"
            
            # Create temporary config
            temp_config=$(mktemp)
            
            # Add ZetMem to existing config
            cat "$CLAUDE_DESKTOP_CONFIG" | jq --arg server_path "$ZETMEM_SERVER_PATH" --arg config_path "$ZETMEM_CONFIG_PATH" '
                .mcpServers = (.mcpServers // {}) + {
                    "zetmem-augmented": {
                        "command": $server_path,
                        "args": ["-config", $config_path],
                        "env": {
                            "OPENAI_API_KEY": "${OPENAI_API_KEY:-}"
                        }
                    }
                }
            ' > "$temp_config"
            
            # Replace original config
            mv "$temp_config" "$CLAUDE_DESKTOP_CONFIG"
            
            print_status "A-MEM added to restored config"
        fi
    fi
else
    print_warning "No backup config found"
    
    echo ""
    echo "Manual restoration options:"
    echo "1. If you have your own backup, restore it manually"
    echo "2. Recreate your config from scratch"
    echo "3. Add A-MEM to current config (preserving existing servers)"
    echo ""
    
    read -p "Do you want to add A-MEM to current config? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Create backup first
        if [ -f "$CLAUDE_DESKTOP_CONFIG" ]; then
            cp "$CLAUDE_DESKTOP_CONFIG" "${CLAUDE_DESKTOP_CONFIG}.pre-amem-backup"
            print_info "Created backup at: ${CLAUDE_DESKTOP_CONFIG}.pre-amem-backup"
        fi
        
        ZETMEM_SERVER_PATH="$PROJECT_ROOT/zetmem-server"
        ZETMEM_CONFIG_PATH="$PROJECT_ROOT/config/production.yaml"
        
        # Create config directory if it doesn't exist
        mkdir -p "$(dirname "$CLAUDE_DESKTOP_CONFIG")"
        
        # Create or update config
        if [ ! -f "$CLAUDE_DESKTOP_CONFIG" ]; then
            echo '{}' > "$CLAUDE_DESKTOP_CONFIG"
        fi
        
        # Create temporary config
        temp_config=$(mktemp)
        
        # Add A-MEM to existing config (preserve existing servers)
        cat "$CLAUDE_DESKTOP_CONFIG" | jq --arg server_path "$ZETMEM_SERVER_PATH" --arg config_path "$ZETMEM_CONFIG_PATH" '
            .mcpServers = (.mcpServers // {}) + {
                "zetmem-augmented": {
                    "command": $server_path,
                    "args": ["-config", $config_path],
                    "env": {
                        "OPENAI_API_KEY": "${OPENAI_API_KEY:-}"
                    }
                }
            }
        ' > "$temp_config"
        
        # Replace original config
        mv "$temp_config" "$CLAUDE_DESKTOP_CONFIG"
        
        print_status "A-MEM added to Claude Desktop config"
    fi
fi

echo ""
echo "Final config:"
cat "$CLAUDE_DESKTOP_CONFIG" | jq . 2>/dev/null || echo "Invalid JSON"

echo ""
print_info "Remember to restart Claude Desktop for changes to take effect"
echo ""
echo "Config file location: $CLAUDE_DESKTOP_CONFIG"
