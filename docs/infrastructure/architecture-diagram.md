# A-MEM MCP Server Architecture Diagrams

## System Architecture Overview

```mermaid
graph TB
    subgraph "External Clients"
        CC[Claude Code]
        CD[Claude Desktop]
        API[REST API Clients]
    end
    
    subgraph "Load Balancer"
        LB[Nginx/Traefik<br/>Port 80/443]
    end
    
    subgraph "Application Layer"
        AS[A-MEM Server<br/>Go Application<br/>Port 8080]
        MS[Metrics Server<br/>Port 9092]
    end
    
    subgraph "Data Layer"
        CB[ChromaDB<br/>Vector Store<br/>Port 8004]
        RD[Redis<br/>Cache/Queue<br/>Port 6382]
        RMQ[RabbitMQ<br/>Message Queue<br/>Port 5672]
    end
    
    subgraph "Processing Layer"
        ST[Sentence Transformers<br/>Embedding Service<br/>Port 8005]
        EW[Evolution Workers<br/>Background Jobs]
    end
    
    subgraph "Monitoring"
        PM[Prometheus<br/>Metrics Collection<br/>Port 9091]
        GF[Grafana<br/>Dashboards<br/>Port 3000]
    end
    
    CC --> LB
    CD --> LB
    API --> LB
    
    LB --> AS
    AS --> CB
    AS --> RD
    AS --> RMQ
    AS --> ST
    
    AS --> MS
    MS --> PM
    PM --> GF
    
    RMQ --> EW
    EW --> CB
    EW --> ST
```

## Container Network Architecture

```mermaid
graph LR
    subgraph "Docker Host"
        subgraph "amem-network (bridge)"
            AS[amem-server]
            CB[chromadb]
            RD[redis]
            RMQ[rabbitmq]
            ST[sentence-transformers]
            PM[prometheus]
        end
        
        subgraph "Volumes"
            V1[chromadb_data]
            V2[redis_data]
            V3[rabbitmq_data]
            V4[prometheus_data]
            V5[sentence_transformers_cache]
        end
        
        subgraph "Port Mapping"
            P1[8080:8080]
            P2[9092:9090]
            P3[8004:8000]
            P4[6382:6379]
            P5[5672:5672]
            P6[15672:15672]
            P7[8005:8000]
            P8[9091:9090]
        end
    end
    
    AS -.-> V1
    CB -.-> V1
    RD -.-> V2
    RMQ -.-> V3
    PM -.-> V4
    ST -.-> V5
    
    AS --> P1
    AS --> P2
    CB --> P3
    RD --> P4
    RMQ --> P5
    RMQ --> P6
    ST --> P7
    PM --> P8
```

## Request Flow Diagram

```mermaid
sequenceDiagram
    participant C as Client
    participant LB as Load Balancer
    participant AS as A-MEM Server
    participant ST as Sentence Transformer
    participant CB as ChromaDB
    participant RD as Redis
    participant RMQ as RabbitMQ
    
    C->>+LB: POST /api/v1/remember
    LB->>+AS: Forward Request
    
    AS->>+RD: Check Cache
    RD-->>-AS: Cache Miss
    
    AS->>+ST: Generate Embedding
    ST-->>-AS: Return Vector
    
    AS->>+CB: Store Memory + Vector
    CB-->>-AS: Confirm Storage
    
    AS->>+RD: Update Cache
    RD-->>-AS: Cache Updated
    
    AS->>+RMQ: Queue Evolution Task
    RMQ-->>-AS: Task Queued
    
    AS-->>-LB: Return Success
    LB-->>-C: 200 OK
```

## Memory Evolution Process

```mermaid
graph TD
    subgraph "Evolution Scheduler"
        S[Cron Scheduler<br/>Daily at 2 AM]
    end
    
    subgraph "Evolution Process"
        S --> Q[Query Memories]
        Q --> B[Batch Processing]
        B --> E[Enhance with LLM]
        E --> V[Validate Quality]
        V --> U[Update Memory]
        U --> I[Index New Vectors]
    end
    
    subgraph "Data Stores"
        CB2[ChromaDB]
        RD2[Redis Cache]
        ST2[Embeddings]
    end
    
    Q --> CB2
    E --> ST2
    U --> CB2
    U --> RD2
    I --> CB2
```

## High Availability Architecture

```mermaid
graph TB
    subgraph "Load Balancer Cluster"
        LB1[HAProxy Primary]
        LB2[HAProxy Secondary]
    end
    
    subgraph "Application Cluster"
        AS1[A-MEM Server 1]
        AS2[A-MEM Server 2]
        AS3[A-MEM Server 3]
    end
    
    subgraph "Data Layer HA"
        subgraph "ChromaDB Cluster"
            CB1[ChromaDB Primary]
            CB2[ChromaDB Replica]
        end
        
        subgraph "Redis Cluster"
            RD1[Redis Master]
            RD2[Redis Slave 1]
            RD3[Redis Slave 2]
        end
        
        subgraph "RabbitMQ Cluster"
            RMQ1[RabbitMQ Node 1]
            RMQ2[RabbitMQ Node 2]
            RMQ3[RabbitMQ Node 3]
        end
    end
    
    LB1 -.-> LB2
    LB1 --> AS1
    LB1 --> AS2
    LB1 --> AS3
    
    AS1 --> CB1
    AS2 --> CB1
    AS3 --> CB1
    CB1 -.-> CB2
    
    AS1 --> RD1
    AS2 --> RD1
    AS3 --> RD1
    RD1 -.-> RD2
    RD1 -.-> RD3
    
    AS1 --> RMQ1
    AS2 --> RMQ2
    AS3 --> RMQ3
    RMQ1 -.-> RMQ2
    RMQ2 -.-> RMQ3
```

## Security Architecture

```mermaid
graph TB
    subgraph "Internet"
        I[Internet Traffic]
    end
    
    subgraph "Security Layer"
        WAF[Web Application Firewall]
        SSL[SSL/TLS Termination]
    end
    
    subgraph "DMZ"
        LB3[Load Balancer]
        RP[Reverse Proxy]
    end
    
    subgraph "Application Zone"
        FW1[Firewall]
        AS4[A-MEM Servers]
    end
    
    subgraph "Data Zone"
        FW2[Firewall]
        DB[Databases]
        VAULT[HashiCorp Vault<br/>Secrets Management]
    end
    
    subgraph "Monitoring Zone"
        FW3[Firewall]
        MON[Monitoring Stack]
        SIEM[Security Monitoring]
    end
    
    I --> WAF
    WAF --> SSL
    SSL --> LB3
    LB3 --> RP
    RP --> FW1
    FW1 --> AS4
    AS4 --> FW2
    FW2 --> DB
    FW2 --> VAULT
    AS4 --> FW3
    FW3 --> MON
    FW3 --> SIEM
```

## Deployment Pipeline

```mermaid
graph LR
    subgraph "Development"
        DEV[Local Docker]
        TEST[Unit Tests]
    end
    
    subgraph "CI/CD"
        GIT[Git Push]
        CI[GitHub Actions]
        BUILD[Docker Build]
        SCAN[Security Scan]
        PUSH[Push to Registry]
    end
    
    subgraph "Staging"
        STAGE[Staging Cluster]
        E2E[E2E Tests]
        PERF[Performance Tests]
    end
    
    subgraph "Production"
        PROD[Production Cluster]
        MON2[Monitoring]
        ALERT[Alerting]
    end
    
    DEV --> TEST
    TEST --> GIT
    GIT --> CI
    CI --> BUILD
    BUILD --> SCAN
    SCAN --> PUSH
    PUSH --> STAGE
    STAGE --> E2E
    E2E --> PERF
    PERF --> PROD
    PROD --> MON2
    MON2 --> ALERT
```

## Data Flow Architecture

```mermaid
graph TD
    subgraph "Input Sources"
        API2[REST API]
        MCP[MCP Protocol]
        WS[WebSocket]
    end
    
    subgraph "Processing Pipeline"
        VAL[Input Validation]
        AUTH[Authentication]
        RL[Rate Limiting]
        PROC[Request Processing]
    end
    
    subgraph "Memory Operations"
        EMBED[Embedding Generation]
        STORE[Vector Storage]
        INDEX[Indexing]
        CACHE[Caching]
    end
    
    subgraph "Background Jobs"
        EVOL[Evolution Tasks]
        CLEAN[Cleanup Tasks]
        BACKUP[Backup Tasks]
    end
    
    API2 --> VAL
    MCP --> VAL
    WS --> VAL
    
    VAL --> AUTH
    AUTH --> RL
    RL --> PROC
    
    PROC --> EMBED
    EMBED --> STORE
    STORE --> INDEX
    STORE --> CACHE
    
    STORE --> EVOL
    STORE --> CLEAN
    STORE --> BACKUP
```

## Infrastructure Components

| Component | Purpose | Technology | Scaling Strategy |
|-----------|---------|------------|------------------|
| Load Balancer | Traffic distribution | Nginx/HAProxy | Active-passive HA |
| API Server | Request handling | Go + Gin | Horizontal scaling |
| Vector Database | Memory storage | ChromaDB | Replication |
| Cache Layer | Performance | Redis | Master-slave |
| Message Queue | Async processing | RabbitMQ | Clustering |
| Embedding Service | Vector generation | Sentence Transformers | Load balanced |
| Monitoring | Observability | Prometheus + Grafana | Single instance |

## Network Security Zones

| Zone | Components | Access Rules |
|------|------------|--------------|
| Public | Load Balancer | HTTPS only |
| DMZ | Reverse Proxy | From Public only |
| Application | A-MEM Servers | From DMZ only |
| Data | Databases | From Application only |
| Management | Monitoring | VPN access only |

## Resource Requirements

| Component | CPU | Memory | Storage | Network |
|-----------|-----|--------|---------|---------|
| A-MEM Server | 2-4 cores | 2-4 GB | 10 GB | 1 Gbps |
| ChromaDB | 4-8 cores | 8-16 GB | 100+ GB SSD | 1 Gbps |
| Redis | 2 cores | 4-8 GB | 10 GB | 10 Gbps |
| RabbitMQ | 2 cores | 2-4 GB | 10 GB | 1 Gbps |
| Sentence Transformers | 4 cores | 4-8 GB | 20 GB | 1 Gbps |
| Prometheus | 2 cores | 2-4 GB | 50+ GB | 1 Gbps |