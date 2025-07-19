# A-MEM to ZetMem Rebranding - Comprehensive Discovery Inventory

## Discovery Summary
Completed systematic search for all "amem", "AMEM", "A-MEM", and variations across the entire zetmem project codebase. Found extensive references requiring coordinated updates.

## Reference Categories by Priority

### CRITICAL (Must Change First)
**Go Module & Import Paths:**
- `go.mod`: `module github.com/amem/mcp-server` → `github.com/zetmem/mcp-server`
- All Go files with imports: `github.com/amem/mcp-server/pkg/*` → `github.com/zetmem/mcp-server/pkg/*`
- Files: cmd/server/main.go, all pkg/ files, test files

**Binary Names:**
- `amem-server` → `zetmem-server` in Makefile, Dockerfile, scripts
- Build targets and executable references

### HIGH PRIORITY (Infrastructure)
**Environment Variables:**
- `AMEM_*` → `ZETMEM_*` pattern across all configs
- Variables: AMEM_ENV, AMEM_PORT, AMEM_LOG_LEVEL, AMEM_CONFIG_PATH, etc.
- Files: .env.example, config files, docker-compose.yml

**Docker & Service Names:**
- `amem-server` service → `zetmem-server`
- `amem-network` → `zetmem-network`
- `amem-augmented` MCP tool name → `zetmem-augmented`
- Container names, volume names, network references

**Database Collections:**
- `amem_memories` → `zetmem_memories`
- `amem_memories_dev` → `zetmem_memories_dev`
- ChromaDB collection names in configs

### MEDIUM PRIORITY (User-Facing)
**Documentation:**
- "A-MEM MCP Server" → "ZetMem MCP Server" in titles
- Project descriptions and branding
- README.md, docs/ directory files
- API documentation and examples

**Configuration Files:**
- Claude Desktop integration configs
- MCP tool descriptions and names
- Service discovery identifiers

### LOW PRIORITY (Internal)
**Monitoring & Metrics:**
- Prometheus metric names: `amem_*` → `zetmem_*`
- Log messages and service names
- Health check references

**Comments & Internal References:**
- Code comments mentioning "A-MEM"
- Error messages and user-facing text
- Test names and descriptions

## Files Requiring Updates (Partial List)
- go.mod (CRITICAL)
- All .go files with imports (CRITICAL)
- Makefile, Dockerfile (HIGH)
- docker-compose.yml (HIGH)
- All config/*.yaml files (HIGH)
- All config/*.json files (HIGH)
- README.md (MEDIUM)
- All docs/ files (MEDIUM)
- All scripts/ files (MEDIUM)
- .env.example (HIGH)
- .gitignore (LOW)
- monitoring/prometheus.yml (LOW)

## Estimated Scope
- **Files to modify**: 50+ files
- **References to update**: 200+ individual references
- **Critical path dependencies**: Go module → imports → builds → configs → docs