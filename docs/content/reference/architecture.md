---
title: Architecture
description: Comprehensive architecture documentation for LocalAGI
---

# LocalAGI Architecture

This document provides a comprehensive overview of the LocalAGI architecture, covering system components, data flows, deployment patterns, and design decisions. It is intended for developers and architects working on or integrating with the project.

## Table of Contents

- [High-Level System Overview](#high-level-system-overview)
- [Component Architecture](#component-architecture)
- [Data Flow](#data-flow)
- [Backend Architecture](#backend-architecture)
- [Model Loading and Inference Pipeline](#model-loading-and-inference-pipeline)
- [API Layer Architecture](#api-layer-architecture)
- [Web UI Architecture](#web-ui-architecture)
- [Plugin and Action System](#plugin-and-action-system)
- [Connector Architecture](#connector-architecture)
- [Knowledge Base and RAG Architecture](#knowledge-base-and-rag-architecture)
- [Deployment Architecture](#deployment-architecture)
- [Scalability Considerations](#scalability-considerations)
- [Performance Bottlenecks and Optimization](#performance-bottlenecks-and-optimization)
- [Security Architecture](#security-architecture)

---

## High-Level System Overview

LocalAGI is a modular, event-driven AI agent platform written in Go. It enables users to create, configure, and manage autonomous AI agents that can interact through multiple channels (Slack, Discord, Telegram, web UI, etc.), execute actions (browse the web, search, manage GitHub issues, etc.), and maintain persistent knowledge bases.

```mermaid
graph TB
    subgraph External["External Services"]
        LLM["LLM Backend<br/>(LocalAI / OpenAI-compatible)"]
        Slack["Slack"]
        Discord["Discord"]
        Telegram["Telegram"]
        GitHub["GitHub"]
        Twitter["Twitter/X"]
        Email["Email"]
        MCP_Ext["MCP Servers"]
    end

    subgraph LocalAGI["LocalAGI Platform"]
        WebUI["Web UI<br/>(React SPA)"]
        API["API Layer<br/>(Fiber HTTP)"]
        SSE["SSE Manager<br/>(Real-time Updates)"]
        Pool["Agent Pool"]
        Agent1["Agent 1"]
        Agent2["Agent 2"]
        AgentN["Agent N"]
        Actions["Action System<br/>(45+ Built-in)"]
        Connectors["Connector System"]
        Skills["Skills System"]
        RAG["RAG / Knowledge Base"]
        Scheduler["Task Scheduler"]
        State["State Persistence"]
    end

    subgraph Storage["Storage Layer"]
        FS["File System<br/>(JSON State)"]
        VectorDB["Vector Store<br/>(Chromem / LocalAI)"]
        PostgreSQL["PostgreSQL<br/>(LocalRecall)"]
    end

    WebUI --> API
    API --> Pool
    Pool --> Agent1
    Pool --> Agent2
    Pool --> AgentN
    Agent1 --> Actions
    Agent1 --> RAG
    Agent1 --> Scheduler
    Agent1 --> SSE
    Connectors --> Pool
    Actions --> LLM
    Agent1 --> LLM
    Connectors --> Slack
    Connectors --> Discord
    Connectors --> Telegram
    Connectors --> GitHub
    Connectors --> Twitter
    Connectors --> Email
    Skills --> MCP_Ext
    RAG --> VectorDB
    RAG --> PostgreSQL
    State --> FS
```

**Key architectural principles:**

- **Modular design** — Components are loosely coupled through interfaces and dependency injection.
- **OpenAI-compatible API** — The LLM integration layer uses the OpenAI SDK, allowing any compatible backend (LocalAI, OpenAI, Ollama, etc.).
- **Event-driven communication** — Agents process jobs from a channel-based queue; real-time updates flow through SSE.
- **Plugin-oriented extensibility** — Actions, connectors, dynamic prompts, MCP tools, and skills can all be added without modifying core code.

---

## Component Architecture

```mermaid
graph LR
    subgraph Core["core/"]
        direction TB
        AgentPkg["agent/<br/>Agent engine, options,<br/>identity, prompts, MCP,<br/>knowledge base, state"]
        StatePkg["state/<br/>Agent pool, config,<br/>RAG provider, compaction"]
        SchedulerPkg["scheduler/<br/>Task scheduling,<br/>JSON store"]
        ConvPkg["conversations/<br/>Conversation tracker"]
        SSEPkg["sse/<br/>SSE broadcast manager"]
        TypesPkg["types/<br/>Action, Job, State,<br/>Filter, Observable"]
    end

    subgraph Pkg["pkg/"]
        direction TB
        LLMPkg["llm/<br/>OpenAI client wrapper"]
        VectorPkg["vectorstore/<br/>Chromem, LocalAI backends"]
        LocalRAGPkg["localrag/<br/>RAG HTTP client"]
        DeepFacePkg["deepface/<br/>Face detection client"]
        ConfigPkg["config/<br/>UI metadata"]
        ClientPkg["client/<br/>HTTP client SDK"]
    end

    subgraph Services["services/"]
        direction TB
        ActionsSvc["actions/<br/>45+ action implementations"]
        ConnectorsSvc["connectors/<br/>Slack, Discord, Telegram,<br/>GitHub, IRC, Matrix, etc."]
        SkillsSvc["skills/<br/>Plugin manager,<br/>skillserver integration"]
        FiltersSvc["filters/<br/>Message filtering"]
        PromptsSvc["prompts/<br/>Dynamic prompt injection"]
    end

    subgraph WebUI["webui/"]
        direction TB
        AppHandler["app.go<br/>HTTP handlers"]
        Routes["routes.go<br/>Route registration"]
        Collections["collections/<br/>RAG management"]
        ReactUI["react-ui/<br/>React SPA (Bun build)"]
    end

    Core --> Pkg
    Services --> Core
    WebUI --> Core
    WebUI --> Services
```

### Component Responsibilities

| Component | Directory | Responsibility |
|-----------|-----------|----------------|
| **Agent Engine** | `core/agent/` | Core agent logic: LLM interaction, job execution, prompt construction, action dispatch, knowledge base lookup, MCP tool integration |
| **Agent Pool** | `core/state/` | Manages lifecycle of multiple agents, configuration loading, RAG provider setup |
| **Scheduler** | `core/scheduler/` | Cron-style task execution for reminders and periodic agent runs |
| **SSE Manager** | `core/sse/` | Broadcasts real-time observable state updates to connected web clients |
| **Type System** | `core/types/` | Shared type definitions: actions, jobs, state, observables, filters |
| **LLM Client** | `pkg/llm/` | Wraps the OpenAI Go SDK with configurable base URL, timeout, and API key |
| **Vector Store** | `pkg/vectorstore/` | Abstraction over vector databases (Chromem in-memory, LocalAI embeddings) |
| **Actions** | `services/actions/` | 45+ built-in action implementations (search, browse, GitHub, image gen, memory, etc.) |
| **Connectors** | `services/connectors/` | Input/output channel integrations (Slack, Discord, Telegram, GitHub, etc.) |
| **Skills** | `services/skills/` | Git-based plugin system using skillserver and MCP protocol |
| **Web API** | `webui/` | HTTP API (Fiber), SSE endpoints, static file serving, collections management |
| **React UI** | `webui/react-ui/` | Single-page application for agent management, chat, and configuration |

---

## Data Flow

### Chat Request Flow

```mermaid
sequenceDiagram
    participant User
    participant WebUI as React UI
    participant API as Fiber API
    participant Pool as Agent Pool
    participant Agent
    participant LLM as LLM Backend
    participant Actions as Action System
    participant SSE as SSE Manager

    User->>WebUI: Send message
    WebUI->>API: POST /api/chat/:name
    API->>Pool: Lookup agent
    Pool->>Agent: Submit job to queue
    Agent->>Agent: Build prompt<br/>(system + KB + history)
    Agent->>LLM: ChatCompletion request
    LLM-->>Agent: Response (text or tool calls)

    alt Tool Call Response
        Agent->>Actions: Execute action(s)
        Actions-->>Agent: Action results
        Agent->>LLM: Follow-up with results
        LLM-->>Agent: Final response
    end

    Agent->>SSE: Publish observable state
    SSE-->>WebUI: Stream updates
    API-->>WebUI: Return ResponseBody
    WebUI-->>User: Display response
```

### Connector-Initiated Flow

```mermaid
sequenceDiagram
    participant Platform as External Platform<br/>(Slack/Discord/etc.)
    participant Connector
    participant Pool as Agent Pool
    participant Agent
    participant LLM as LLM Backend
    participant Actions as Action System

    Platform->>Connector: Incoming event<br/>(message, mention, webhook)
    Connector->>Pool: Route to agent
    Pool->>Agent: Submit job to queue
    Agent->>Agent: Build prompt with<br/>connector context
    Agent->>LLM: ChatCompletion request
    LLM-->>Agent: Response

    alt Tool Calls
        Agent->>Actions: Execute action(s)
        Actions-->>Agent: Results
        Agent->>LLM: Follow-up
        LLM-->>Agent: Final response
    end

    Agent-->>Connector: Return result
    Connector->>Platform: Send response<br/>(message, thread reply, etc.)
```

### Knowledge Base Query Flow

```mermaid
sequenceDiagram
    participant Agent
    participant KB as Knowledge Base
    participant VectorDB as Vector Store
    participant LLM as LLM Backend

    Agent->>Agent: Receive user message
    Agent->>KB: knowledgeBaseLookup(message)
    KB->>VectorDB: Semantic search
    VectorDB-->>KB: Similar entries
    KB-->>Agent: Relevant context

    Agent->>Agent: Inject KB results<br/>into prompt
    Agent->>LLM: ChatCompletion<br/>(with KB context)
    LLM-->>Agent: Response
```

---

## Backend Architecture

### Entry Point and Initialization

The application starts in `main.go` with the following initialization sequence:

```mermaid
graph TD
    A["main()"] --> B["Load environment variables<br/>(model names, API URLs, keys)"]
    B --> C["Create SkillsService<br/>(optional plugin system)"]
    C --> D["Create AgentPool<br/>(with service factories)"]
    D --> E["Initialize WebUI App<br/>(Fiber + React SPA)"]
    E --> F["Set RAG provider<br/>(HTTP or embedded)"]
    F --> G["Start AgentPool<br/>(load persisted agents)"]
    G --> H["Listen on :3000"]
```

### Agent Pool and Lifecycle

The `AgentPool` (`core/state/pool.go`) manages the full lifecycle of agents:

- **Creation** — Agents are created from `AgentConfig` (60+ configuration fields) via `CreateAgent()`.
- **Persistence** — Agent configurations are serialized to `pool.json` in the state directory.
- **Recreation** — `RecreateAgent()` stops and restarts an agent with updated config.
- **Deletion** — Agents are removed from the pool and their state is cleaned up.

Each agent runs independently with its own:
- **Job queue** — A Go channel (`chan *types.Job`) for serialized job processing.
- **Conversation tracker** — Manages conversation windows per connector/session.
- **OpenAI client** — Configured with the agent's model and API settings.
- **State** — Persistent internal state (goals, memories, current task).

### Concurrency Model

```mermaid
graph TB
    subgraph AgentGoroutines["Per Agent"]
        JQ["Job Queue<br/>(buffered channel)"]
        Worker["Job Worker<br/>(goroutine)"]
        SchedPoll["Scheduler Polling<br/>(goroutine)"]
    end

    subgraph SSEGoroutines["SSE System"]
        Workers["Worker Pool<br/>(goroutines)"]
        Broadcast["Broadcast Channel"]
    end

    subgraph ConnectorGoroutines["Connectors"]
        SlackListener["Slack SocketMode<br/>(goroutine)"]
        DiscordListener["Discord Gateway<br/>(goroutine)"]
        TelegramPoller["Telegram Poller<br/>(goroutine)"]
    end

    JQ --> Worker
    Worker --> Broadcast
    SchedPoll --> JQ
    SlackListener --> JQ
    DiscordListener --> JQ
    TelegramPoller --> JQ
    Broadcast --> Workers
```

**Synchronization primitives used:**
- `chan` — Job queues, SSE broadcast channels
- `sync.Map` — Thread-safe SSE client tracking
- `sync.RWMutex` — Protecting shared agent state
- `context.Context` — Cancellation propagation for jobs and tasks

### Design Patterns

| Pattern | Usage |
|---------|-------|
| **Builder (Options)** | Agent configuration via functional options (`core/agent/options.go`) |
| **Factory** | Connector and action creation from config (`services/connectors.go`, `services/actions.go`) |
| **Observer** | SSE manager broadcasts state to subscribed clients |
| **Adapter** | RAG provider adapters (HTTP vs. in-process), KB compaction adapter |
| **Pool** | Agent pool manages multiple independent agent instances |
| **Strategy** | Swappable RAG backends (HTTP, embedded Chromem, LocalAI) |

---

## Model Loading and Inference Pipeline

LocalAGI does not load models directly. Instead, it delegates model inference to an external OpenAI-compatible API server (typically [LocalAI](https://localai.io)).

```mermaid
graph LR
    subgraph LocalAGI
        Agent["Agent"]
        LLMClient["LLM Client<br/>(pkg/llm)"]
    end

    subgraph ExternalLLM["LLM Backend (e.g., LocalAI)"]
        ModelServer["Model Server"]
        Model1["Chat Model<br/>(e.g., gemma-3-4b)"]
        Model2["Multimodal Model"]
        Model3["Embedding Model<br/>(e.g., granite-embedding)"]
        Model4["TTS Model"]
        Model5["Transcription Model"]
    end

    Agent --> LLMClient
    LLMClient -->|"OpenAI SDK<br/>HTTP POST /v1/chat/completions"| ModelServer
    ModelServer --> Model1
    ModelServer --> Model2
    ModelServer --> Model3
    ModelServer --> Model4
    ModelServer --> Model5
```

### Configuration

Models are configured via environment variables:

| Variable | Purpose | Example |
|----------|---------|---------|
| `LOCALAGI_MODEL` | Primary chat/reasoning model | `gemma-3-4b-it` |
| `LOCALAGI_MULTIMODAL_MODEL` | Vision/multimodal model | `gemma-3-4b-it` |
| `LOCALAGI_TRANSCRIPTION_MODEL` | Speech-to-text model | `whisper-1` |
| `LOCALAGI_TTS_MODEL` | Text-to-speech model | — |
| `LOCALAGI_LLM_API_URL` | LLM API base URL | `http://localai:8080` |
| `LOCALAGI_LLM_API_KEY` | API authentication key | — |
| `EMBEDDING_MODEL` | Embedding model for vector search | `granite-embedding-107m-multilingual` |

### Inference Flow

1. The agent constructs a prompt combining system instructions, identity guidance, knowledge base context, conversation history, and available tool definitions.
2. The `pkg/llm` client sends the prompt to the LLM backend via the OpenAI `ChatCompletion` API.
3. If the LLM responds with tool calls, the agent executes the corresponding actions and feeds results back.
4. This loop continues until the LLM produces a final text response or the maximum evaluation loops are reached.

---

## API Layer Architecture

The API layer is built on the [Fiber](https://gofiber.io/) web framework and serves both the REST API and the embedded React SPA.

```mermaid
graph TB
    subgraph Middleware
        Auth["API Key Auth<br/>(v2keyauth)"]
        Static["Static File Server<br/>(embedded React dist)"]
        SPA["SPA Fallback<br/>(index.html)"]
    end

    subgraph Endpoints
        Chat["POST /api/chat/:name<br/>Chat with agent"]
        Create["POST /api/agent/create<br/>Create agent"]
        Delete["DELETE /api/agent/:name<br/>Delete agent"]
        Pause["PUT /api/agent/:name/pause<br/>Pause agent"]
        Start["PUT /api/agent/:name/start<br/>Resume agent"]
        Config["GET|PUT /api/agent/:name/config<br/>Agent configuration"]
        Meta["GET /api/agent/config/metadata<br/>Config field metadata"]
        SSEEndpoint["GET /sse/:name<br/>Real-time updates"]
        Responses["POST /v1/responses<br/>OpenAI-compatible"]
        CollUpload["POST /api/collections/upload<br/>Upload documents"]
        CollSearch["GET /api/collections/search<br/>Search documents"]
        CollList["GET /api/collections<br/>List collections"]
        SkillsList["GET /api/skills<br/>List skills"]
    end

    Auth --> Endpoints
    Static --> SPA
```

### Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/chat/:name` | Send a message to an agent and receive a response |
| `GET` | `/sse/:name` | Subscribe to real-time agent state updates via SSE |
| `POST` | `/api/agent/create` | Create a new agent with configuration |
| `DELETE` | `/api/agent/:name` | Delete an agent |
| `PUT` | `/api/agent/:name/pause` | Pause an agent |
| `PUT` | `/api/agent/:name/start` | Resume a paused agent |
| `GET` | `/api/agent/:name/config` | Retrieve agent configuration |
| `PUT` | `/api/agent/:name/config` | Update agent configuration |
| `POST` | `/v1/responses` | OpenAI-compatible responses endpoint |
| `POST` | `/api/collections/upload` | Upload documents to a knowledge base collection |
| `GET` | `/api/collections/search` | Semantic search across collections |
| `GET` | `/api/collections` | List all collections |

### SSE (Server-Sent Events)

The SSE system (`core/sse/`) provides real-time streaming of agent observable state to web clients:

- Each agent has a dedicated SSE broadcast manager.
- The manager maintains a worker pool for parallel client distribution.
- A history buffer (10 messages) enables late-joining clients to catch up.
- Standard SSE headers are set (`Content-Type: text/event-stream`, `Cache-Control: no-cache`).

---

## Web UI Architecture

```mermaid
graph TB
    subgraph Build["Build Pipeline"]
        BunBuild["Bun Bundler<br/>(Vite config)"]
        ReactSrc["React Source<br/>(webui/react-ui/src/)"]
        Dist["Built Assets<br/>(react-ui/dist/)"]
        GoEmbed["go:embed<br/>(compiled into binary)"]
    end

    subgraph Frontend["React SPA"]
        Pages["Pages<br/>(src/pages/)"]
        Components["Components<br/>(src/components/)"]
        Contexts["Contexts<br/>(src/contexts/)"]
        Hooks["Hooks<br/>(src/hooks/)"]
        Utils["Utilities<br/>(src/utils/)"]
    end

    subgraph Communication["Client-Server"]
        REST["REST API Calls"]
        SSEClient["SSE Event Source"]
    end

    ReactSrc --> BunBuild
    BunBuild --> Dist
    Dist --> GoEmbed

    Pages --> Components
    Pages --> Contexts
    Pages --> Hooks
    Components --> Utils
    Pages --> REST
    Pages --> SSEClient
```

### Key Characteristics

- **Framework:** React with Bun as the build tool (Vite configuration).
- **Embedding:** The built React assets are embedded into the Go binary using `//go:embed`, creating a single deployable binary.
- **SPA Routing:** A fallback route in Fiber serves `index.html` for all unmatched paths, enabling client-side routing.
- **Real-time Updates:** The UI connects to `/sse/:name` endpoints to receive live agent state updates (current task, reasoning, tool execution status).
- **Agent Management:** The UI provides forms for creating, configuring, pausing, and deleting agents, with field metadata driven by the `/api/agent/config/metadata` endpoint.

---

## Plugin and Action System

LocalAGI has a multi-layered extensibility model:

```mermaid
graph TB
    subgraph ActionTypes["Action Sources"]
        BuiltIn["Built-in Actions<br/>(services/actions/)<br/>45+ implementations"]
        UserDefined["User-Defined Actions<br/>(YAML/JSON config)"]
        MCPTools["MCP Tools<br/>(HTTP or stdio servers)"]
        SkillsActions["Skills<br/>(Git repos via skillserver)"]
    end

    subgraph Interface["Action Interface"]
        ActionI["Action Interface<br/>Run() + Definition()"]
    end

    subgraph Agent["Agent Runtime"]
        ToolDefs["Tool Definitions<br/>(sent to LLM)"]
        Dispatch["Action Dispatch"]
        Results["Action Results"]
    end

    BuiltIn --> ActionI
    UserDefined --> ActionI
    MCPTools --> ActionI
    SkillsActions --> ActionI
    ActionI --> ToolDefs
    ToolDefs --> Dispatch
    Dispatch --> Results
```

### Action Interface

All actions implement the `Action` interface (`core/types/actions.go`):

```go
type Action interface {
    Run(ctx context.Context, sharedState *AgentSharedState, params ActionParams) (ActionResult, error)
    Definition() ActionDefinition
}
```

- `Definition()` returns a JSON Schema describing the action's parameters — this is sent to the LLM as a tool definition.
- `Run()` executes the action with the provided parameters and returns a result (text and/or image).

### Built-in Actions (45+)

| Category | Actions |
|----------|---------|
| **Web** | Search, Browse, Scraper, Wikipedia |
| **GitHub** | Create/comment/list/close issues, create/review/merge PRs, manage repos |
| **Generation** | Image generation, PDF generation, Song generation |
| **Social** | Twitter posting, Email, Telegram messaging |
| **Memory** | Add/list/remove/search memory entries |
| **Scheduling** | One-time reminders, Recurring reminders |
| **System** | Shell commands, Webhooks, PiKVM control |
| **Multi-Agent** | Call other agents |

### MCP (Model Context Protocol)

Agents can integrate with external MCP servers (`core/agent/mcp.go`):

- **HTTP MCP servers** — Tools are listed via HTTP and wrapped as LocalAGI actions.
- **Stdio MCP servers** — Tools run as child processes communicating over stdin/stdout.
- MCP tools are dynamically discovered and registered as agent actions at startup.

### Skills System

The skills system (`services/skills/`) provides a Git-based plugin mechanism:

- Skills are stored in the `stateDir/skills` directory.
- The `skillserver` library manages skill discovery and indexing.
- Skills are exposed to agents via dynamic prompts (XML injection) and MCP sessions.

---

## Connector Architecture

Connectors bridge external communication platforms with the agent pool:

```mermaid
graph LR
    subgraph Platforms["External Platforms"]
        Slack["Slack<br/>(SocketMode)"]
        Discord["Discord<br/>(Gateway WebSocket)"]
        Telegram["Telegram<br/>(Long Polling)"]
        GitHub_C["GitHub<br/>(Webhooks)"]
        IRC_C["IRC<br/>(TCP)"]
        Matrix_C["Matrix<br/>(Sync API)"]
        Email_C["Email<br/>(IMAP/SMTP)"]
        Twitter_C["Twitter/X<br/>(API)"]
    end

    subgraph ConnectorLayer["Connector Layer"]
        direction TB
        Parse["Parse incoming event"]
        Route["Route to agent"]
        Format["Format response"]
    end

    subgraph AgentPool
        Agent["Agent Job Queue"]
    end

    Slack --> ConnectorLayer
    Discord --> ConnectorLayer
    Telegram --> ConnectorLayer
    GitHub_C --> ConnectorLayer
    IRC_C --> ConnectorLayer
    Matrix_C --> ConnectorLayer
    Email_C --> ConnectorLayer
    Twitter_C --> ConnectorLayer

    ConnectorLayer --> AgentPool
    AgentPool --> ConnectorLayer
    ConnectorLayer --> Platforms
```

Each connector handles:

1. **Authentication** — Platform-specific credentials and token management.
2. **Event Listening** — Receiving messages, mentions, or webhook payloads.
3. **Context Mapping** — Converting platform events into agent job requests.
4. **Response Formatting** — Translating agent responses back to platform-appropriate formats (threads, reactions, file uploads, etc.).
5. **State Tracking** — Managing conversation threads and active job tracking per channel/user.

Connectors are instantiated by factory functions in `services/connectors.go` based on the `ConnectorConfig` in each agent's configuration.

---

## Knowledge Base and RAG Architecture

```mermaid
graph TB
    subgraph Input["Document Input"]
        Upload["File Upload API"]
        Memory["Memory Actions<br/>(add to memory)"]
        Compaction["KB Compaction<br/>(conversation summaries)"]
    end

    subgraph RAGLayer["RAG Layer"]
        direction TB
        Backend["Collections Backend<br/>(interface)"]
        HTTP_BE["HTTP Backend<br/>(external LocalRAG)"]
        InProc["In-Process Backend<br/>(embedded LocalRecall)"]
    end

    subgraph VectorLayer["Vector Storage"]
        Chromem["Chromem-go<br/>(in-memory vectors)"]
        LocalAI_Emb["LocalAI Embeddings"]
        PostgreSQL_V["PostgreSQL<br/>(persistent storage)"]
    end

    subgraph Query["Query Path"]
        AgentQ["Agent receives message"]
        KBLookup["knowledgeBaseLookup()"]
        Search["Semantic search"]
        Inject["Inject context into prompt"]
    end

    Upload --> Backend
    Memory --> Backend
    Compaction --> Backend
    Backend --> HTTP_BE
    Backend --> InProc
    HTTP_BE --> PostgreSQL_V
    InProc --> Chromem
    Chromem --> LocalAI_Emb

    AgentQ --> KBLookup
    KBLookup --> Search
    Search --> Backend
    Backend --> Inject
```

The RAG system supports two backend strategies:

- **HTTP Backend** — Delegates to an external LocalRAG service backed by PostgreSQL.
- **In-Process Backend** — Uses the `localrecall` library with Chromem-go for in-memory vector storage and LocalAI for embeddings.

Knowledge base features include:

- **Auto-search** — Automatically queries the KB on every user message.
- **KB as Tools** — Exposes KB search as an LLM tool for on-demand retrieval.
- **Compaction** — Periodically summarizes conversation history into KB entries.
- **Collections** — Named document collections with upload, search, and management APIs.

---

## Deployment Architecture

### Docker Compose (Standard)

```mermaid
graph TB
    subgraph DockerCompose["Docker Compose Stack"]
        LocalAGI_D["LocalAGI<br/>Port: 8080→3000<br/>(Fiber + React)"]
        LocalAI_D["LocalAI<br/>Port: 8081→8080<br/>(Model Server)"]
        PostgreSQL_D["PostgreSQL<br/>Port: 5432<br/>(LocalRecall DB)"]
        SSHBox_D["SSHBox<br/>(Shell Access)"]
        DinD["Docker-in-Docker<br/>Port: 2375"]
    end

    LocalAGI_D -->|"LLM API"| LocalAI_D
    LocalAGI_D -->|"RAG Storage"| PostgreSQL_D
    LocalAGI_D -->|"Container Ops"| DinD
    SSHBox_D -->|"Docker CLI"| DinD

    User["User"] -->|"HTTP :8080"| LocalAGI_D
```

### Build Pipeline

The Docker build uses a multi-stage process:

```mermaid
graph LR
    Stage1["Stage 1: UI Builder<br/>(oven/bun:1)<br/>Build React with Bun"]
    Stage2["Stage 2: Go Builder<br/>(golang:1.26-alpine)<br/>Compile Go binary"]
    Stage3["Stage 3: Runtime<br/>(ubuntu:24.04)<br/>Minimal container"]

    Stage1 -->|"dist/"| Stage2
    Stage2 -->|"binary"| Stage3
```

### GPU-Accelerated Variants

| Compose File | Target Hardware |
|-------------|-----------------|
| `docker-compose.yaml` | CPU only |
| `docker-compose.nvidia.yaml` | NVIDIA GPUs (CUDA) |
| `docker-compose.intel.yaml` | Intel GPUs (oneAPI) |
| `docker-compose.amd.yaml` | AMD GPUs (ROCm) |

### Deployment Patterns

**Single-node deployment (recommended for getting started):**
- All services on one machine via Docker Compose.
- LocalAI handles model serving with optional GPU acceleration.
- Suitable for development and small-scale production.

**Separated model server:**
- LocalAGI connects to a remote LocalAI instance (or any OpenAI-compatible API).
- Set `LOCALAGI_LLM_API_URL` to point to the remote server.
- Allows dedicated GPU hardware for inference while running LocalAGI on a lighter machine.

**External LLM provider:**
- Use hosted APIs (OpenAI, Anthropic via proxy, etc.) by configuring the API URL and key.
- No local GPU required.

---

## Scalability Considerations

### Current Architecture Characteristics

- **Single-process** — LocalAGI runs as a single Go process managing all agents.
- **Vertical scaling** — Add more CPU/RAM to handle more concurrent agents and jobs.
- **Agent-level parallelism** — Each agent processes jobs from its own goroutine; agents run independently.
- **Configurable parallel jobs** — The `ParallelJobs` setting controls concurrent job execution per agent.

### Scaling Strategies

| Dimension | Approach |
|-----------|----------|
| **More agents** | Increase process memory; agents are lightweight goroutines |
| **Higher throughput** | Scale the LLM backend horizontally (multiple LocalAI instances behind a load balancer) |
| **Larger knowledge bases** | Use PostgreSQL-backed RAG (LocalRecall) instead of in-memory Chromem |
| **Multiple instances** | Run separate LocalAGI instances with shared LLM backend and database |
| **Connector scaling** | Each connector manages its own connection pool; no shared state between connectors |

### Bottleneck Analysis

The primary bottleneck is typically LLM inference latency. Agent processing (prompt construction, action execution, state management) is fast relative to model inference time.

---

## Performance Bottlenecks and Optimization

### Known Bottlenecks

1. **LLM Inference Latency** — The dominant factor in response time. Each agent request involves at least one LLM call; tool-using conversations may require multiple round trips.
2. **Vector Search at Scale** — In-memory Chromem performs well for moderate datasets but may become slow with very large knowledge bases.
3. **Connector Polling** — Some connectors (Telegram, email) use polling intervals that add latency to message processing.
4. **SSE Fan-out** — Broadcasting to many simultaneous web clients can consume goroutines.

### Optimization Strategies

| Bottleneck | Mitigation |
|------------|------------|
| LLM latency | Use GPU-accelerated LocalAI; choose smaller models for latency-sensitive tasks; configure appropriate timeouts (default 150s) |
| Vector search | Switch to PostgreSQL-backed LocalRecall for large datasets; tune `KBResults` count |
| Memory usage | Enable `KB Compaction` to summarize old conversations; configure `AutoCompactionThreshold` |
| Action execution | Actions execute in parallel when the LLM requests multiple tool calls; `MaxAttempts` limits retries |
| Conversation growth | `ConversationStorageMode` and conversation window settings prevent unbounded memory growth |
| Loop detection | `LoopDetection` and `MaxEvaluationLoops` prevent runaway agent behavior |

---

## Security Architecture

### Authentication and Authorization

```mermaid
graph LR
    Request["Incoming Request"] --> AuthMW["API Key Middleware<br/>(v2keyauth)"]
    AuthMW -->|"Valid Key"| Handler["Request Handler"]
    AuthMW -->|"Invalid/Missing"| Reject["401 Unauthorized"]
```

- The API is protected by API key authentication via Fiber's `v2keyauth` middleware.
- API keys are configured at startup.
- No built-in role-based access control — all authenticated requests have full access.

### Security Considerations

| Area | Current State | Recommendation |
|------|--------------|----------------|
| **API Authentication** | API key middleware | Use strong, unique keys; rotate regularly |
| **LLM API Communication** | HTTP with optional API key | Use HTTPS in production; restrict network access |
| **Action Execution** | Shell command action available | Disable shell action in production unless required; restrict with filters |
| **Connector Credentials** | Stored in agent configuration | Use environment variables or secrets management; avoid persisting tokens in plain-text config |
| **File Uploads** | Collections upload endpoint | Validate file types and sizes; sanitize filenames |
| **Docker Security** | Docker-in-Docker included | Restrict DinD access; use rootless Docker where possible |
| **State Persistence** | JSON files on disk | Protect state directory permissions; encrypt sensitive data at rest |
| **MCP Servers** | External process execution | Vet MCP servers; use stdio mode for isolation; restrict network access for HTTP MCP |

### Defense in Depth

- **Message Filters** — The filter system (`services/filters/`) can intercept and modify messages before they reach agents, enabling content moderation and input sanitization.
- **Action Allow-listing** — Only explicitly configured actions are available to each agent.
- **Evaluation Limits** — `MaxEvaluationLoops` prevents infinite tool-call loops.
- **Context Isolation** — Each agent has its own state, conversation history, and knowledge base; no cross-agent data leakage by default.
- **Network Segmentation** — In Docker deployments, services communicate over an internal network; only the LocalAGI port is exposed externally.
