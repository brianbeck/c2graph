# C2Graph

Graph-based fraud detection platform for Solana blockchain. Ingests on-chain transaction data, builds a wallet/transaction graph in Neo4j, and scores wallets for fraud risk using heuristic signals.

## Architecture

```
┌───────────┐     ┌─────────────┐     ┌──────────┐
│  Frontend  │────▶│ API Gateway  │────▶│  Neo4j   │
│  (React)   │     │   (Go/Chi)   │     │ (Graph)  │
└───────────┘     └──────┬──────┘     └────▲─────┘
                         │                  │
                         │ RabbitMQ         │
                         ▼                  │
                  ┌─────────────┐           │
                  │   Worker    │───────────┘
                  │  (Go/AMQP)  │───▶ Solana RPC
                  └─────────────┘
```

- **Frontend** — React 19 + TypeScript + Vite, Supabase auth, interactive force-directed graph visualization with filtering and export
- **API Gateway** — Go (Chi router), REST API with Supabase JWT auth, publishes scan jobs to RabbitMQ
- **Worker** — Go, consumes scan jobs from RabbitMQ, fetches transactions from Solana RPC, writes the graph to Neo4j, runs fraud scoring
- **Neo4j** — Graph database storing wallets, transactions, programs, and their relationships
- **RabbitMQ** — Job queue for async scan processing

## What It Detects

### Fraud Risk Scoring

Each wallet receives a composite risk score (0.0–1.0) based on these heuristics:

| Signal | What It Detects | Max Contribution |
|---|---|---|
| **Cycle Detection** | Circular fund flows (wash trading) — funds leaving a wallet and returning within 30 days | 30 pts |
| **Fan-Out** | Funds distributed to 10+ wallets within 1-hour windows (structuring/splitting) | 25 pts |
| **Fan-In** | 10+ wallets sending to a single wallet within 1-hour windows (consolidation) | 25 pts |
| **Velocity Anomaly** | Transaction rate exceeding 5× the baseline of 5 tx/day | 20 pts |
| **Proximity to Blacklisted** | Shortest graph distance to known bad wallets (1-hop: 50, 2-hop: 20, 3-hop: 5) | 50 pts |
| **High-Risk Program Usage** | Interactions with known high-risk programs (mixers, etc.) | 15 pts/program |
| **Bot Behavior** | Automated wallet activity (see below) | 25 pts |

The raw score (0–100) is normalized to 0.0–1.0 and stored on the wallet node along with which factors contributed.

### Bot Detection

A separate bot likelihood score (0.0–1.0) identifies automated wallets using four signals:

| Signal | Weight | What It Measures |
|---|---|---|
| **Timing Regularity** | 30% | Coefficient of variation of inter-transaction intervals. Bots have very consistent timing (CV < 0.1) |
| **Velocity** | 30% | Transactions per day. >100/day is strongly bot-like |
| **Program Concentration** | 20% | Unique programs vs transaction count. Bots repeatedly call the same 1–2 programs |
| **Hour Spread** | 20% | Distinct hours of the day with activity. Bots transact 24/7 with no sleep pattern |

Wallets with bot likelihood ≥ 0.5 are flagged as bots. Each individual signal score is stored on the wallet node for transparency.

## Graph Schema

**Nodes:**
- `Wallet` — address, risk_score, bot_likelihood, tx_count, first_seen, last_seen, tags
- `Transaction` — signature, block_time, slot, fee, sol_amount
- `Program` — program_id, name, category, risk_level
- `Token` — mint_address, symbol, name
- `ScanJob` — job_id, status, root_address, requested_depth

**Relationships:**
- `(:Wallet)-[:INITIATED]->(:Transaction)` — wallet signed the transaction
- `(:Transaction)-[:SOL_TRANSFER]->(:Wallet)` — SOL transferred to wallet
- `(:Transaction)-[:TOKEN_TRANSFER]->(:Wallet)` — token transferred to wallet
- `(:Transaction)-[:CONTAINS_INSTRUCTION]->(:Instruction)-[:EXECUTED_BY]->(:Program)`

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | No | Health check |
| `POST` | `/api/scan` | JWT | Start a scan (`{"address": "...", "depth": 1-3}`) |
| `GET` | `/api/status/{job_id}` | JWT | Check scan job status |
| `GET` | `/api/graph/{address}?depth=1` | JWT | Fetch wallet subgraph |

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [Go 1.25+](https://go.dev/dl/) (for local development)
- [Node.js 20+](https://nodejs.org/) (for frontend development)
- A Solana RPC endpoint (the public mainnet endpoint works but is rate-limited)
- A [Supabase](https://supabase.com/) project (free tier, for authentication)

## Quick Start

### 1. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and fill in:
- `SOLANA_RPC_URL` — your Solana RPC endpoint (or set `SOLANA_RPC_URLS` for multiple endpoints with fallback)
- `SUPABASE_URL`, `SUPABASE_ANON_KEY`, `SUPABASE_JWT_SECRET` — from your Supabase project dashboard
- `VITE_SUPABASE_URL`, `VITE_SUPABASE_ANON_KEY` — same Supabase values for the frontend

### 2. Start everything

```bash
make up
```

This builds and starts all services:

| Service | URL |
|---|---|
| Frontend | http://localhost:5173 |
| API Gateway | http://localhost:8080 |
| Neo4j Browser | http://localhost:7474 |
| RabbitMQ Management | http://localhost:15672 |

### 3. Use the app

1. Open http://localhost:5173 and sign in with Supabase
2. Enter a Solana wallet address and click Scan
3. The worker fetches transactions from Solana, builds the graph, and scores the wallet
4. Explore the interactive graph visualization — filter by node type, risk score; export as JSON or CSV

### Stop

```bash
make down
```

## Local Development

For development, run infrastructure in Docker and services locally:

```bash
# Start Neo4j + RabbitMQ
make dev-infra

# In separate terminals:
make dev-api        # API gateway on :8080
make dev-worker     # Worker (connects to RabbitMQ + Neo4j)
make dev-frontend   # Vite dev server on :5173
```

## Build

```bash
make build          # Build all three services
make build-api      # Build API gateway only
make build-worker   # Build worker only
make build-frontend # Build frontend only
```

## Test

```bash
make test           # Run all Go tests
make test-api       # API gateway tests only
make test-worker    # Worker tests only
```

## Utilities

### Rescore existing wallets

Re-run the scoring engine on all wallets already in the database:

```bash
# Via Docker
docker compose exec worker c2graph-rescore

# Or locally
cd worker && go run ./cmd/rescore/
```

### Neo4j manual queries

```bash
# Apply schema constraints
make neo4j-constraints

# Seed known Solana programs
make neo4j-seed
```

## Configuration Reference

All configuration is via environment variables. See [`.env.example`](.env.example) for the full list.

| Variable | Default | Description |
|---|---|---|
| `NEO4J_URI` | `bolt://localhost:7687` | Neo4j Bolt URI |
| `NEO4J_USER` | `neo4j` | Neo4j username |
| `NEO4J_PASSWORD` | `c2graph-dev-password` | Neo4j password |
| `RABBITMQ_URI` | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URI |
| `SOLANA_RPC_URL` | `https://api.mainnet-beta.solana.com` | Solana RPC endpoint |
| `SOLANA_RPC_RATE_LIMIT` | `2` | Max requests/sec to Solana RPC |
| `SOLANA_RPC_URLS` | — | Comma-separated RPC URLs (overrides single URL) |
| `SOLANA_RPC_RATE_LIMITS` | — | Per-endpoint rate limits matching `SOLANA_RPC_URLS` |
| `WORKER_CONCURRENCY` | `4` | Parallel scan workers |
| `BATCH_SIZE` | `50` | Transactions fetched per RPC call |
| `MAX_WALLETS_PER_SCAN` | `10000` | Cap on wallets per scan job |
| `SCAN_FRESHNESS_HOURS` | `24` | Skip re-scanning wallets scanned within this window |
| `API_PORT` | `8080` | API gateway listen port |
| `API_CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin |

## Tech Stack

- **Go 1.25** — API gateway and worker (Chi, zerolog, neo4j-go-driver, amqp091-go)
- **TypeScript / React 19 / Vite 7** — Frontend (react-force-graph-2d, TanStack Query, Supabase JS)
- **Neo4j 5.26** with APOC — Graph database
- **RabbitMQ 3** — Message queue
- **Supabase** — Authentication (JWT)
