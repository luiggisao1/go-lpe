# Golpe (auto-golpe)

<p align="center">
  <strong>An Autonomous Software Engineering (SWE) Agent Harness in Go</strong><br>
  High-throughput, sandboxed multi-agent orchestrator built with Go 1.22+, Redis Asynq, Google Gemini, and Docker.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go" alt="Go Version" />
  <img src="https://img.shields.io/badge/Slack-Event%20API-4A154B?style=flat&logo=slack" alt="Slack" />
  <img src="https://img.shields.io/badge/Queue-Asynq%20%2B%20Redis-DC382D?style=flat&logo=redis" alt="Asynq" />
  <img src="https://img.shields.io/badge/LLM-Google%20Gemini-4285F4?style=flat&logo=google" alt="Gemini" />
  <img src="https://img.shields.io/badge/Sandbox-Docker%20Engine-2496ED?style=flat&logo=docker" alt="Docker" />
  <img src="https://img.shields.io/badge/Status-Active%20WIP-orange?style=flat" alt="Status" />
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License" />
</p>

> 💡 **TL;DR:** **Golpe** is a Slack-native autonomous Software Engineering (SWE) assistant and agent harness. It features non-blocking webhook ingestion, full multi-turn conversational memory with Google Gemini, asynchronous Redis task offloading via Asynq, and ephemeral Docker sandbox management—built in pure Go as an ultra-fast alternative to heavyweight Python agent stacks.

> 🚧 **Project Status:** Active Work in Progress (`v0.1.0-alpha`).  
> The core event ingestion pipeline, asynchronous task offloading, multi-turn Slack thread memory, Gemini integration, and the Docker Sandbox Engine are operational. The Tool Execution Harness and subagent delegation are currently under active development.

---

## ⚡ Overview

**Golpe** is a production-grade, autonomous Software Engineering (SWE) agent harness designed to plan, code, execute, test, and ship features and bug fixes directly from team communications (Slack) and issue trackers (Notion / GitHub).

Built to replace traditional, heavyweight Python/Celery agent pipelines, Golpe leverages Go's native concurrency model, sub-millisecond memory footprint, and static binary distribution to deliver an ultra-responsive, highly scalable agent runtime.

---

## 💡 Motivation: The Case for a Go-Native SWE Agent

Most developer agents today (like early AutoGPT architectures, SWE-bench runners, or custom Celery orchestrators) are written in Python. While Python is great for exploratory scripting, it introduces major architectural friction when building a real-time, production SWE platform:

1. **Slack Webhook Deadlines (`< 3,000ms`):** Slack requires webhook endpoints to acknowledge receipt within 3 seconds. Heavy Python frameworks under load frequently experience latency spikes, triggering Slack's retry mechanism and creating duplicate task runs or bot conversation loops. Golpe uses a native Go `net/http` router with cryptographic HMAC-SHA256 validation to acknowledge events in under a few milliseconds and offload processing to Redis.
2. **The "Host Pollution" Security Trap:** Running AI-generated shell commands (`npm install`, `go build`, arbitrary bash scripts) directly on a developer machine or host server is dangerous. Golpe provisions isolated, resource-capped Docker sandboxes per session, guaranteeing that untrusted third-party code and dependencies never touch the host filesystem.
3. **Credential & Secret Isolation:** Many agent setups inject the LLM API key directly into the container where third-party code runs. A malicious repository script could easily execute `printenv` and steal the keys. Golpe runs the agent orchestrator **outside** the sandbox; commands are injected via the Docker Engine API (`docker exec`), keeping API keys and Slack tokens 100% isolated.
4. **Minimal Resource Footprint:** Python/Node-based agent containers easily consume 500MB–2GB of memory. Golpe compiles into a single static binary running inside a **15.9 MB** Alpine container with negligible idle CPU and RAM usage.

---

## 💬 Slack-Native Interaction Model

Golpe meets software engineering teams right where they already collaborate:

```text
1. Trigger: A developer mentions the bot in a Slack channel:
   "@bot can you explain how goroutines and channels differ from OS threads?"

2. Non-Blocking Event Ingestion:
   The server verifies the cryptographic HMAC-SHA256 signature, acknowledges Slack in milliseconds
   to eliminate webhook timeouts, and enqueues the job into Redis via Asynq.

3. Multi-Turn Thread Memory:
   The background worker fetches the entire conversation history using `conversations.replies`,
   aligning user and model roles so the agent retains full conversational continuity across follow-up questions.

4. Contextual Delivery:
   The agent generates technical guidance using Google Gemini and responds directly
   inside the Slack thread without echo loops.
```

---

### 🚦 Current State: What Works Today

This project is built iteratively following clean architecture principles. The following modules are live and tested:

* **⚡ Non-Blocking Webhook Ingestion:** Fast standard library `net/http` router with cryptographic HMAC-SHA256 signature verification, immediately acknowledging Slack events and eliminating webhook timeouts.
* **📬 Distributed Async Offloading:** Powered by `hibiken/asynq` + Redis, decoupling webhook receipt from long-running agent reasoning.
* **🧵 Multi-Turn Thread Memory:** Full conversational continuity inside Slack threads using `conversations.replies`, formatting message history with `genai.RoleUser` and `genai.RoleModel` while preventing infinite bot echo loops.
* **🤖 Google Gemini:** High-speed LLM reasoning integrated via the official `google.golang.org/genai` Go SDK.
* **📦 Ultra-Lean Multi-Stage Container:** Production Docker build resulting in a static, self-contained **15.9 MB** Alpine image.

---

## 🧠 Target Architecture: The Agent Harness Blueprint

Modern autonomous coding requires more than just calling an LLM endpoint. It demands a **robust agent harness** that coordinates observation, reasoning, safe action execution, and self-correction.

The diagram below outlines the target multi-agent architecture being implemented:

```mermaid
flowchart TD
    subgraph Ingestion ["1. Fast Ingestion Layer"]
        Slack[Slack Webhooks / Events] -->|HMAC-SHA256 Auth| Server[HTTP Ingestion Server\nnet/http ServeMux]
        Server -->|Immediate Enqueue| Redis[(Redis Broker)]
    end

    subgraph Orchestration ["2. Distributed Agent Harness"]
        Redis -->|Async Worker Task| Worker[Asynq Worker Pool]
        Worker --> ContextMgr[Context & Thread State Manager]
        ContextMgr --> Supervisor[Supervisor / Orchestrator Agent\nGoogle Gemini]
    end

    subgraph MultiAgent ["3. Hierarchical Subagent Delegation"]
        Supervisor -->|Delegates Task| Planner[Planner Subagent]
        Supervisor -->|Spawns Coding Run| Coder[Coder Subagent]
        Supervisor -->|Verification Run| Reviewer[Reviewer / Tester Subagent]
    end

    subgraph Tooling ["4. Tool Execution & Sandboxing"]
        Coder -->|Function Calling| Tools{Agent Tools}
        Tools -->|Read / Write Specs| Notion[Notion API]
        Tools -->|docker exec| DockerSandbox[Isolated Docker Sandbox\nWorkspace Volume]
        Tools -->|Commit & Open PR| GitAPI[GitHub API]
    end

    DockerSandbox -->|Stdout, Stderr, Exit Code| Coder
    Reviewer -->|Automated Evaluation Pass| SlackResp[Slack Thread Reply]
```

---

## 🏛️ Key System Capabilities

### 1. High-Performance, Resilient Ingestion
* **Non-Blocking Ingestion Pipeline:** Built on native Go `net/http` using modern pattern-matching routing (`Go 1.22+`), immediately acknowledging Slack events to eliminate timeouts and offloading workloads to Asynq.
* **Cryptographic Verification:** Strict HMAC-SHA256 signature verification preventing replay and tampering attacks.
* **Asynchronous Offloading:** Webhook handlers immediately acknowledge Slack events within milliseconds, enqueuing tasks into Redis via `hibiken/asynq` for reliable background processing.

### 2. Ephemeral Sandbox Execution (Docker Go SDK)
* **Zero Host Pollution:** All repository cloning, dependency resolution (`npm`, `go get`, `cargo`), test suites, and terminal commands execute inside isolated Docker containers.
* **Session-Persistent Workspaces:** Containers live throughout the lifecycle of a Slack thread session, preserving git state and compilation caches across multiple agent iterations.
* **Security & Resource Limits:** Strict CPU quotas (`NanoCPUs`), memory limits, and isolated networking prevent container escapes or runaway processes.

### 3. Dual-Tier Memory System
* **Short-Term / Thread Memory:** Real-time extraction of Slack thread conversation histories using `conversations.replies`, maintaining full context across multi-turn user interactions.
* **Role-Based Context Alignment:** Distinguishes user prompts from model completions (`genai.RoleUser` vs `genai.RoleModel`) to keep the conversational state intact without hallucination loops.
* **Long-Term / Knowledge Memory (Upcoming):** Vector search using `pgvector` and PostgreSQL to store architectural guidelines, past bug solutions, and workspace context.

### 4. Hierarchical Multi-Agent Delegation (Upcoming)
Complex software tasks are broken down using specialized subagents rather than a monolithic prompt:
* **Supervisor Agent:** Triage incoming requests, manage conversation flow, and delegate tasks.
* **Planner Subagent:** Analyzes codebases, maps dependencies, and designs step-by-step implementation plans.
* **Coder Subagent:** Writes idiomatic code, applies diffs, and iterates inside the sandbox.
* **Reviewer / QA Subagent:** Executes linting, static analysis, unit tests, and verifies edge cases before creating Pull Requests.

---

## 📂 Repository Structure

```text
├── cmd/
│   └── server/             # Application entrypoint (HTTP Server + Asynq Worker)
├── internal/
│   ├── config/             # Environment variable parsing and validation
│   ├── docker/             # Docker SDK integration for isolated execution sandboxes
│   ├── gemini/             # Google Gemini GenAI SDK client and function calling
│   ├── http/               # HTTP server, HMAC validation, and Slack webhook routing
│   ├── models/             # Persistence models and database abstractions
│   ├── notion/             # Notion API client for issue tracking and notes
│   ├── queue/              # Asynq task definitions and background event handlers
│   └── slack/              # Slack API client for messaging and thread history retrieval
├── Dockerfile              # Multi-stage static build (< 20 MB production image)
├── docker-compose.yml      # Local development environment (App + Redis + PostgreSQL)
└── go.mod                  # Go module definition
```

---

## 🗺️ Engineering Roadmap & Milestones

| Phase | Milestone | Status | Description |
| :---: | :--- | :---: | :--- |
| **01** | **Base Server & Routing** | ✅ Complete | Native `net/http` router, structured `slog` logging, recovery middleware. |
| **02** | **Containerization** | ✅ Complete | Multi-stage Docker build delivering minimal, secure Alpine/scratch images. |
| **03** | **Webhook Security** | ✅ Complete | Slack HMAC-SHA256 signature verification and payload validation. |
| **04** | **Distributed Queue** | ✅ Complete | Asynq + Redis producer/consumer pipeline running concurrently. |
| **05** | **Conversational Memory**| ✅ Complete | Slack thread context retrieval & multi-turn dialog with Gemini SDK. |
| **06** | **Docker Sandbox Engine**| ✅ Complete | Ephemeral container provisioning, command execution (`docker exec`), and stream demultiplexing. |
| **07** | **Tool Execution Harness**| 🔄 In Progress | Native function calling for file reading, editing, and terminal commands. |
| **08** | **Subagent Delegation** | ⏳ Planned | Hierarchical Planner, Coder, and Reviewer subagent orchestration. |
| **09** | **Issue Tracker Sync** | ⏳ Planned | Bidirectional sync with Notion databases and GitHub Pull Requests. |
| **10** | **Vector Memory** | ⏳ Planned | Long-term memory and RAG using PostgreSQL + `pgvector`. |

---

## 🚀 Getting Started

### Prerequisites
* **Go:** `1.22+` (tested on `Go 1.27.3`)
* **Docker:** Docker Desktop or Docker Engine
* **Redis:** For Asynq queue storage (or via Docker Compose)
* **ngrok:** For receiving Slack webhooks locally

### 1. Clone & Configure
```bash
git clone https://github.com/luiggisao/auto-golpe.git
cd auto-golpe

cp .env.example .env  # Or create .env with required keys
```

Ensure your `.env` contains:
```ini
PORT=8080
ENV=development
REDIS_ADDR=localhost:6379
SLACK_SIGNING_SECRET=your_slack_signing_secret
SLACK_BOT_TOKEN=xoxb-your-bot-token
GOOGLE_API_KEY=your_gemini_api_key
```

### 2. Start Dependencies
```bash
docker compose up -d redis
```

### 3. Run Application
```bash
go run cmd/server/main.go
```

The server will initialize both the HTTP API server on port `:8080` and the Asynq background worker concurrently.

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).
