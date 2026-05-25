# URL Monitor

Distributed URL monitoring system with real-time analytics, built with **Go microservices** and **React**.

Monitors HTTP/HTTPS endpoints on configurable intervals, tracks uptime and response times, surfaces status changes via Kafka events.

## Architecture

```
        ┌──────────────────────────────────────────┐
        │           React Frontend                 │
        │         (Vite + Tailwind 4)              │
        └────────────────┬─────────────────────────┘
                         │ REST + JWT
        ┌────────────────▼─────────────────────────┐
        │           API Gateway                    │
        │         (HTTP → gRPC)                    │
        └─────┬──────────┬──────────────┬──────────┘
              │ gRPC     │ gRPC         │ gRPC
        ┌─────▼────┐ ┌───▼────┐  ┌──────▼─────────┐
        │   Auth   │ │  URL   │  │   Analytics    │
        │  Service │ │Service │  │   Service      │
        └─────┬────┘ └───┬────┘  └──────┬─────────┘
              │          │              │
              └──────────┼──────────────┘
                         ▼
              ┌──────────────────────┐
              │     PostgreSQL       │
              └──────────────────────┘
                         ▲
              ┌──────────┴──────────┐
              │   Monitor Worker    │──┐
              │  (HTTP prober)      │  │
              └─────────────────────┘  ▼
                                 ┌──────────┐
                                 │  Kafka   │
                                 └──────────┘
                                       │
                            (consumed by Analytics)
```

## Services

| Service | Stack | Port | Purpose |
|---|---|---|---|
| **auth-service** | Go, gRPC, JWT, bcrypt | 50051 | User registration, login, token validation |
| **url-service** | Go, gRPC, pgx | 50052 | CRUD operations for monitored URLs |
| **monitor-worker** | Go, cron, kafka-go | — | Pings URLs every N seconds, publishes events |
| **analytics-service** | Go, gRPC, kafka-go | 50053 | Aggregates check results, computes uptime stats |
| **gateway** | Go, HTTP, gRPC clients | 8000 | REST API for frontend, JWT middleware |
| **frontend** | React 18, Vite, Tailwind 4 | 5173 | User UI with real-time dashboard |

## Infrastructure

- **PostgreSQL 16** — primary datastore (users, urls, checks)
- **Apache Kafka** — event bus (url.checked, url.status_changed)
- **Redis 7** — caching layer
- **Prometheus + Grafana** — metrics and dashboards
- **cAdvisor + node-exporter + nvidia-gpu-exporter** — system metrics

## Quick Start

Prerequisites: Docker, Docker Compose, Go 1.26+, Node 20+

```bash
git clone https://github.com/Shakir0905/url-monitor.git
cd url-monitor

cp .env.example .env
# Edit .env and set JWT_SECRET (openssl rand -hex 32)

docker compose up -d

cd frontend
npm install
npm run dev
```

Open http://localhost:5173

## Endpoints

### Public REST API (via Gateway)

```
POST /api/auth/register      Create new user
POST /api/auth/login         Login, returns JWT

# Below require Authorization: Bearer <token>
GET    /api/urls             List user's URLs
POST   /api/urls             Add URL to monitor
GET    /api/urls/:id         Get single URL
PUT    /api/urls/:id         Update URL settings
DELETE /api/urls/:id         Stop monitoring URL

GET /api/dashboard           User's monitoring overview
GET /api/urls/:id/stats      Uptime stats for a URL
```

### Internal gRPC

Each service exposes gRPC reflection:

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50052 list
grpcurl -plaintext localhost:50053 list
```

## Observability

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Kafka UI**: http://localhost:8080

## Project Structure

```
url-monitor/
├── cmd/                     # Service entry points (1 dir = 1 binary)
│   ├── auth/
│   ├── url/
│   ├── monitor-worker/
│   ├── analytics/
│   └── gateway/
├── internal/                # Service implementations
│   ├── auth/{domain,repository,service,server,config}/
│   ├── url/
│   ├── monitor/
│   ├── analytics/
│   ├── gateway/
│   └── pkg/db/              # Shared Postgres pool
├── proto/                   # gRPC contracts
├── migrations/              # SQL migrations
├── infra/                   # Prometheus / Grafana configs
├── frontend/                # React app
└── docker-compose.yml
```

## Tech Highlights

- **Clean Architecture**: each service split into domain/repository/service/server layers
- **gRPC + Protobuf**: typed contracts between services
- **Kafka pub/sub**: decoupled event-driven analytics
- **JWT auth** with bcrypt password hashing
- **Graceful shutdown** in all services
- **Structured logging** (slog JSON)
- **Multi-stage Docker builds** (~20MB final images, non-root user)
- **Connection pooling** (pgx)
- **CORS middleware** for browser clients
- **Tailwind CSS 4** with glassmorphism design

## License

MIT
