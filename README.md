# url-monitor

![CI](https://github.com/Shakir0905/url-monitor/actions/workflows/ci.yml/badge.svg)
![Docker](https://github.com/Shakir0905/url-monitor/actions/workflows/docker.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)
![Kubernetes](https://img.shields.io/badge/Kubernetes-K3s-326CE5?logo=kubernetes&logoColor=white)

Distributed URL monitoring system built on Go microservices. The system periodically checks user-registered URLs, tracks uptime and response time, and exposes analytics through a REST gateway and a React dashboard.

## Architecture

```
┌──────────────┐       REST      ┌─────────────┐       gRPC
│  React UI    │ ──────────────▶ │  Gateway    │ ──────────────┐
│  (Vite, 5173)│                 │  (8000)     │               │
└──────────────┘                 └──────┬──────┘               │
                                        │                      │
                                        │ gRPC                 ▼
                          ┌─────────────┼─────────────┐   ┌──────────────┐
                          ▼             ▼             ▼   │  Analytics   │
                   ┌──────────┐  ┌──────────┐  ┌──────────┐  │  (50053)     │
                   │   Auth   │  │   URL    │  │ Analytics│  └──────┬───────┘
                   │ (50051)  │  │ (50052)  │  │ (50053)  │         │ consumes
                   └────┬─────┘  └────┬─────┘  └──────────┘         │ url.checked
                        │             │                              │
                        └──────┬──────┘                              │
                               ▼                                     │
                        ┌──────────────┐                             │
                        │  PostgreSQL  │ ◀───────────────────────────┘
                        │   (5432)     │
                        └──────────────┘
                               ▲
                               │ reads/writes
                        ┌──────┴───────┐         publishes
                        │ Monitor      │ ──────────────────▶ ┌──────────────┐
                        │ Worker       │                     │    Kafka     │
                        │ (cron 30s)   │                     │   (9092)     │
                        └──────────────┘                     └──────────────┘
```

## Services

| Service          | Port  | Purpose                                                    |
|------------------|-------|------------------------------------------------------------|
| `gateway`        | 8000  | REST API, JWT middleware, CORS, gRPC clients to backends   |
| `auth`           | 50051 | Registration, login, JWT issuing and validation            |
| `url`            | 50052 | CRUD for user-owned URLs                                   |
| `monitor-worker` | -     | Cron worker, pings URLs, publishes events to Kafka         |
| `analytics`      | 50053 | Aggregates check events, exposes dashboard stats           |
| `frontend`       | 5173  | React + Vite + Tailwind 4 dashboard                        |

## Tech Stack

**Backend:** Go 1.25, gRPC, Protocol Buffers, pgx/v5, golang-migrate, segmentio/kafka-go, golang-jwt/jwt/v5, log/slog
**Frontend:** React 18, Vite, Tailwind 4, react-router, axios
**Data:** PostgreSQL 16, Redis 7, Apache Kafka 3.8 (KRaft mode, no Zookeeper)
**Observability:** Prometheus, Grafana, cAdvisor, node-exporter, nvidia-gpu-exporter
**Infrastructure:** Docker Compose, Kubernetes (K3s/k3d)
**CI/CD:** GitHub Actions, GitHub Container Registry

## Quick Start

### Docker Compose (local development)

```bash
cp .env.example .env
docker compose up -d --build
```

Services:
- Frontend: http://localhost:5173
- Gateway API: http://localhost:8000
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090
- Kafka UI: http://localhost:8080

Run database migrations:
```bash
make migrate-up
```

### API examples

```bash
# Register
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  http://localhost:8000/api/auth/register

# Login
curl -X POST -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  http://localhost:8000/api/auth/login

# Add URL (use token from login response)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://google.com","check_interval_seconds":60}' \
  http://localhost:8000/api/urls
```

## Kubernetes Deployment

Full Kubernetes manifests in `k8s/`. Tested on K3s/k3d locally.

```bash
# Create cluster
k3d cluster create url-monitor --port "8090:80@loadbalancer"

# Namespace
kubectl create namespace url-monitor
kubectl config set-context --current --namespace=url-monitor

# Apply manifests
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/kafka.yaml
kubectl apply -f k8s/migrations.yaml
kubectl apply -f k8s/auth-service.yaml
kubectl apply -f k8s/url-service.yaml
kubectl apply -f k8s/monitor-worker.yaml
kubectl apply -f k8s/analytics-service.yaml
kubectl apply -f k8s/gateway.yaml

# Access gateway
kubectl port-forward service/gateway 8001:8000
```

**Manifest highlights:**
- `Deployment` for stateless services (auth, url, gateway, analytics, monitor-worker)
- `StatefulSet` for stateful components (postgres, kafka with persistent volumes)
- `Service` for internal service discovery via Kubernetes DNS
- `ConfigMap` for non-sensitive configuration
- `Secret` for credentials, shared across services via `secretKeyRef`
- `Job` for one-shot database migrations

## CI/CD

Continuous integration via **GitHub Actions** (`.github/workflows/`).

### CI workflow (`ci.yml`)

Runs on every push and PR to `main`:
- `gofmt` formatting check
- `go vet` static analysis
- Unit tests with race detector (`go test -race ./...`)
- Build all 5 Go services
- Frontend build

### Docker workflow (`docker.yml`)

Runs on push to `main` and version tags (`v*`):
- Parallel matrix build of 5 service images
- Multi-stage Dockerfiles, final image size ~20MB per service
- Published to GitHub Container Registry (`ghcr.io`)
- Auto-tagged: branch name, commit SHA, `latest` for main, semver for tags
- Build cache via GitHub Actions cache

### Pre-commit hooks

Local quality gate in `scripts/hooks/pre-commit`:
- `gofmt` formatting check
- `go vet`
- `golangci-lint` on new changes only
- `go test -short`

Install:
```bash
make install-hooks
```

### Container images

All 5 service images are public on GitHub Container Registry:

```bash
docker pull ghcr.io/shakir0905/url-monitor-auth:latest
docker pull ghcr.io/shakir0905/url-monitor-url:latest
docker pull ghcr.io/shakir0905/url-monitor-gateway:latest
docker pull ghcr.io/shakir0905/url-monitor-analytics:latest
docker pull ghcr.io/shakir0905/url-monitor-monitor-worker:latest
```

## Testing

```bash
go test ./...                            # All tests
go test -race ./...                      # With race detector
go test -cover ./...                     # With coverage
go test -v ./internal/auth/service       # Verbose, specific package
```

Unit tests use in-memory mock repositories (e.g., `internal/auth/service/auth_service_test.go`) to isolate service-layer logic from infrastructure, keeping tests fast and deterministic.

## Project Structure

```
url-monitor/
├── cmd/                    # Service entry points and Dockerfiles
│   ├── auth/
│   ├── url/
│   ├── monitor-worker/
│   ├── analytics/
│   └── gateway/
├── internal/               # Private application code
│   ├── auth/               # 5-layer pattern per service:
│   │   ├── domain/         #   business types and errors
│   │   ├── repository/     #   data access
│   │   ├── service/        #   business logic (tested)
│   │   ├── server/         #   gRPC handlers
│   │   └── config/         #   env-driven config
│   ├── url/
│   ├── monitor/
│   ├── analytics/
│   ├── gateway/
│   └── pkg/db/             # Shared database utilities
├── proto/                  # Protocol Buffers definitions and generated code
├── migrations/             # SQL migration files
├── frontend/               # React + Vite dashboard
├── k8s/                    # Kubernetes manifests
├── infra/                  # Prometheus, Grafana configuration
├── scripts/hooks/          # Git hooks (shared across team)
├── .github/workflows/      # GitHub Actions CI/CD
├── docker-compose.yml
└── Makefile
```

## License

MIT
