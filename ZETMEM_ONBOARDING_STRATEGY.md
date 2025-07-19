# Zetmem Onboarding Strategy for Coding Agents
## Comprehensive Guide to Effective Memory Management

**Version:** 1.0  
**Last Updated:** July 18, 2025  
**Target Audience:** AI Coding Agents and Their Operators

---

## Table of Contents

1. [Quick-Start Guide](#quick-start-guide)
2. [Core Protocols](#core-protocols)
3. [Decision Matrices & Flow-Charts](#decision-matrices--flow-charts)
4. [Templates & Code Snippets](#templates--code-snippets)
5. [Project-Type Overlays](#project-type-overlays)
6. [FAQ / Troubleshooting](#faq--troubleshooting)
7. [Glossary & Versioning](#glossary--versioning)

---

## I. Quick-Start Guide

### Essential Setup (2-Minute Start)

```
1. Initialize workspace: workspace_init_zetmem(identifier="/path/to/project", name="Project Name")
2. Store first memory: store_coding_memory_zetmem(content="...", workspace_id="/path/to/project")
3. Retrieve memories: retrieve_relevant_memories_zetmem(query="...", workspace_id="/path/to/project")
```

### Core Principles

**CONSISTENCY OVER SPORADIC USE**
- Zetmem effectiveness scales exponentially with consistent usage
- Store memories after each significant coding session
- Retrieve memories before starting new tasks
- Evolve network weekly or after 10+ new memories

**WORKSPACE-FIRST APPROACH**
- Always specify workspace_id for scoped organization
- Use filesystem paths for project-specific workspaces
- Use logical names for cross-project themes
- Maintain clean separation between different contexts

### Default Thresholds

```
Storage Trigger: token_count ≤ 350 AND likely_needed_beyond_2_exchanges = true
Retrieval Default: min_relevance = 0.3 (cast wide net initially)
Precision Mode: min_relevance = 0.82 (high-quality matches)
Evolution Frequency: Weekly OR after 10+ new memories
```

---

## II. Core Protocols

### 1. Workspace Initialization Protocol

#### When to Initialize
- **New Project**: At the start of each new codebase or project
- **Context Switch**: When switching to a different codebase
- **Theme-Based Work**: For cross-project patterns or learning

#### How to Initialize
```python
# Project-specific workspace
workspace_init_zetmem(
    identifier="/Users/dev/my-project",
    name="My Project Development"
)

# Theme-based workspace
workspace_init_zetmem(
    identifier="react-patterns",
    name="React Design Patterns"
)
```

#### Naming Conventions
- **Filesystem Paths**: Use absolute paths for project-specific workspaces
- **Logical Names**: Use kebab-case for cross-project themes
- **Descriptive Names**: Provide clear, human-readable workspace names

### 2. Memory Storage Best Practices

#### What to Store
- **Code Snippets**: Reusable functions, patterns, configurations
- **Architectural Decisions**: Design choices and their rationale
- **Problem-Solution Pairs**: Debugging insights and fixes
- **Design Patterns**: Implementation examples and use cases
- **Learning Progressions**: Skill development and knowledge building

#### When to Store
- After solving a non-trivial problem
- When discovering new patterns or approaches
- After implementing significant features
- When gaining insights about the codebase
- Before context switches to preserve knowledge

#### How to Store
```python
store_coding_memory_zetmem(
    content="function debounce(func, delay) { ... }",
    workspace_id="/Users/dev/my-project",
    code_type="javascript",
    context="Performance optimization for search input handling"
)
```

### 3. Memory Retrieval Strategy

#### Threshold Management
```
Start Low → Filter Up Strategy:
1. Begin with min_relevance: 0.3 (wide net)
2. If results > 15: increase to 0.4
3. If results > 10: increase to 0.5
4. If results > 5: increase to 0.6
5. For precision: use 0.82+
```

#### Search Optimization
- Use specific keywords and technical terms
- Include code_types when relevant
- Always specify workspace_id for scoping
- Combine multiple queries for complex topics

#### Retrieval Examples
```python
# Wide exploration
retrieve_relevant_memories_zetmem(
    query="authentication middleware",
    workspace_id="/Users/dev/api-project",
    min_relevance=0.3,
    max_results=10
)

# Precision search
retrieve_relevant_memories_zetmem(
    query="JWT token validation error handling",
    workspace_id="/Users/dev/api-project",
    code_types=["javascript", "typescript"],
    min_relevance=0.82,
    max_results=5
)
```

### 4. Ongoing Memory Management

#### Regular Habits
- **Session Start**: Retrieve relevant memories for current task
- **Problem Solving**: Store insights as they emerge
- **Session End**: Store significant discoveries
- **Weekly Review**: Evolve memory network for optimization

#### Network Evolution
```python
# Trigger evolution after significant additions
evolve_memory_network_zetmem(
    scope="recent",
    max_memories=100,
    trigger_type="manual"
)
```

#### Quality Monitoring
- Review retrieval effectiveness
- Refine stored memories based on usage
- Update contexts for better searchability
- Remove or merge duplicate memories

### 5. Integration Workflow

#### Session-Based Workflow
```
1. SESSION START
   ├── Retrieve relevant memories for current task
   ├── Review previous solutions and patterns
   └── Set context for new work

2. DURING DEVELOPMENT
   ├── Store insights as they emerge
   ├── Document problem-solution pairs
   └── Capture architectural decisions

3. SESSION END
   ├── Store significant discoveries
   ├── Update memory contexts if needed
   └── Trigger evolution if 10+ new memories

4. WEEKLY MAINTENANCE
   ├── Evolve memory network
   ├── Review memory quality
   └── Clean up duplicates
```

---

## III. Decision Matrices & Flow-Charts

### Workspace Initialization Decision Tree

```
New Work Context?
├── YES: Is it a specific project/codebase?
│   ├── YES: Use filesystem path
│   │   └── workspace_init_zetmem(identifier="/path/to/project")
│   └── NO: Is it a cross-project theme?
│       ├── YES: Use logical name
│       │   └── workspace_init_zetmem(identifier="theme-name")
│       └── NO: Use default workspace
└── NO: Continue with existing workspace
```

### Memory Storage Decision Matrix

```
Should I store this memory?

Information Type     | Token Count | Future Need | Store?
--------------------|-------------|-------------|--------
Code Snippet        | ≤ 350       | Likely      | YES
Architecture Decision| Any         | Likely      | YES
Quick Fix           | ≤ 350       | Maybe       | YES
Temporary Workaround| Any         | Unlikely    | NO
Standard Library Use| ≤ 100       | Unlikely    | NO
Complex Algorithm   | > 350       | Likely      | SPLIT*

*SPLIT: Break into smaller, focused memories
```

### Retrieval Threshold Decision Flow

```
Query Results Analysis:
├── Results > 15: Increase threshold to 0.4
├── Results 10-15: Increase threshold to 0.5
├── Results 5-10: Current threshold optimal
├── Results 2-5: Consider lowering threshold
└── Results 0-1: Lower threshold or broaden query
```

---

## IV. Templates & Code Snippets

### Workspace Initialization Templates

#### Template 1: New Project Setup
```python
# Initialize workspace for new project
workspace_init_zetmem(
    identifier="/Users/dev/new-project",
    name="New Project Development"
)

# Store initial architecture memory
store_coding_memory_zetmem(
    content="Project uses React + TypeScript + Node.js stack with PostgreSQL database",
    workspace_id="/Users/dev/new-project",
    code_type="architecture",
    context="Initial project setup and technology stack decisions"
)
```

#### Template 2: Theme-Based Workspace
```python
# Initialize cross-project theme workspace
workspace_init_zetmem(
    identifier="performance-optimization",
    name="Performance Optimization Patterns"
)

# Store performance pattern
store_coding_memory_zetmem(
    content="const memoizedFunction = useMemo(() => expensiveCalculation(data), [data]);",
    workspace_id="performance-optimization",
    code_type="javascript",
    context="React useMemo pattern for expensive calculations"
)
```

### Memory Storage Templates

#### Template 3: Problem-Solution Pattern
```python
store_coding_memory_zetmem(
    content="""
    // Problem: CORS errors in development
    // Solution: Configure proxy in package.json
    {
      "name": "my-app",
      "proxy": "http://localhost:3001"
    }
    """,
    workspace_id="/Users/dev/react-app",
    code_type="json",
    context="CORS configuration for React development server"
)
```

#### Template 4: Architecture Decision
```python
store_coding_memory_zetmem(
    content="""
    Decision: Use Redux Toolkit instead of plain Redux
    Rationale: 
    - Reduces boilerplate code by 70%
    - Built-in Immer for immutable updates
    - Better TypeScript support
    - Recommended by Redux team
    """,
    workspace_id="/Users/dev/state-management",
    code_type="architecture",
    context="State management library selection for large React application"
)
```

### Retrieval Query Templates

#### Template 5: Broad Exploration
```python
memories = retrieve_relevant_memories_zetmem(
    query="error handling patterns",
    workspace_id="/Users/dev/current-project",
    min_relevance=0.3,
    max_results=15
)
```

#### Template 6: Specific Technical Search
```python
memories = retrieve_relevant_memories_zetmem(
    query="async/await error handling try-catch",
    workspace_id="/Users/dev/api-project",
    code_types=["javascript", "typescript"],
    min_relevance=0.82,
    max_results=5
)
```

---

## V. Project-Type Overlays

### Archetype A: Greenfield Single Repository

**Characteristics:**
- New project from scratch
- Single codebase
- Clear technology stack
- Focused team/individual

**Workspace Strategy:**
```python
# Single workspace for entire project
workspace_init_zetmem(
    identifier="/Users/dev/greenfield-project",
    name="Greenfield Project Development"
)
```

**Memory Organization:**
- Store architectural decisions early
- Document technology choices and rationale
- Capture design patterns as they emerge
- Focus on building reusable components

**Specific Protocols:**
- **Storage Frequency**: After each feature implementation
- **Retrieval Pattern**: Before starting new features
- **Evolution Trigger**: Weekly or after major milestones
- **Threshold Strategy**: Start with 0.3, increase for precision

### Archetype B: Large Monorepo with Microservices

**Characteristics:**
- Multiple services/modules
- Complex interdependencies
- Large team collaboration
- Diverse technology stack

**Workspace Strategy:**
```python
# Service-specific workspaces
workspace_init_zetmem(
    identifier="/Users/dev/monorepo/auth-service",
    name="Authentication Service"
)

workspace_init_zetmem(
    identifier="/Users/dev/monorepo/payment-service", 
    name="Payment Service"
)

# Cross-cutting concerns workspace
workspace_init_zetmem(
    identifier="monorepo-patterns",
    name="Monorepo Shared Patterns"
)
```

**Memory Organization:**
- Separate workspaces per service
- Shared workspace for cross-cutting patterns
- Document service interfaces and contracts
- Capture integration patterns

**Specific Protocols:**
- **Storage Frequency**: After service changes and integrations
- **Retrieval Pattern**: Service-specific + shared patterns
- **Evolution Trigger**: Bi-weekly for active services
- **Threshold Strategy**: Higher precision (0.5+) for service-specific queries

### Archetype C: Script-Style One-offs / Notebooks

**Characteristics:**
- Exploratory coding
- Data analysis scripts
- Proof of concepts
- Learning exercises

**Workspace Strategy:**
```python
# Theme-based workspaces
workspace_init_zetmem(
    identifier="data-analysis-patterns",
    name="Data Analysis Techniques"
)

workspace_init_zetmem(
    identifier="ml-experiments",
    name="Machine Learning Experiments"
)
```

**Memory Organization:**
- Group by technique or domain
- Store successful patterns and snippets
- Document data processing workflows
- Capture learning insights

**Specific Protocols:**
- **Storage Frequency**: After successful experiments
- **Retrieval Pattern**: Technique-focused queries
- **Evolution Trigger**: Monthly or after major learning
- **Threshold Strategy**: Lower thresholds (0.3-0.4) for exploration

---

## VI. FAQ / Troubleshooting

### Common Issues

**Q: Too many irrelevant results in retrieval**
A: Increase min_relevance threshold incrementally (0.3 → 0.4 → 0.5 → 0.6)

**Q: No results found for specific queries**
A: Lower min_relevance to 0.3 or broaden query terms

**Q: Duplicate memories being stored**
A: Use retrieve before store to check for existing similar memories

**Q: Workspace organization becoming messy**
A: Review workspace strategy, consider splitting by domain or merging related ones

**Q: Memory evolution taking too long**
A: Reduce max_memories parameter or increase evolution frequency

### Best Practices Troubleshooting

**Issue: Inconsistent memory quality**
- Solution: Establish clear storage criteria and review regularly

**Issue: Poor retrieval relevance**
- Solution: Improve memory contexts and use more specific queries

**Issue: Workspace proliferation**
- Solution: Follow naming conventions and merge related workspaces

---

## VII. Glossary & Versioning

### Key Terms

**Workspace**: Logical grouping of memories by project or theme
**Memory**: Stored code snippet, insight, or knowledge with metadata
**Evolution**: AI-driven optimization of memory network connections
**Threshold**: Minimum relevance score for memory retrieval
**Context**: Descriptive information about a memory's purpose and usage

### Version History

- **v1.0** (July 18, 2025): Initial comprehensive strategy document
- **Future**: Quarterly reviews aligned with zetmem feature releases

### Feedback Loop

This document is designed as a living asset. Updates should be made based on:
- Agent performance data
- User feedback and usage patterns
- New zetmem features and capabilities
- Emerging best practices from the community

---

**Document Status**: Production Ready  
**Next Review**: October 18, 2025  
**Maintainer**: Zetmem Development Team
