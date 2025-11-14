# CRUDER - User Management API

A CRUD API for user management built with Go, Gin framework, and PostgreSQL. Except for terraform, all the other features are implemented mentioned in the [TASK.md](./TASK.md) file.

**Quick look:**
- **36MB Docker image** - 98% size reduction from naive build (1.87GB → 36MB)
- **Live deployment** - Production: http://136.110.146.135 | Staging: http://34.49.250.233
- **CI/CD pipeline** - Automated quality gates (lint, security, tests) + deployment
- **Structured logging** - JSON logs with request tracing and automatic log levels

The original task requirements can be found in [TASK.md](./TASK.md)


## Reviewer's Guide: Engineering Decisions

- int vs. uuid: The task mentioned :uuid for endpoints, but the initial code used int IDs. To fix this at the foundation, the database migration was corrected to use UUIDs as primary keys for security and scalability.

- Service Layer Tests: The task's example showed an integration test, but the request was for service-layer tests. internal/service/users_test.go implements these as true unit tests, mocking the repository to test business logic (like validation and normalization) in isolation.

- Docker ("Minimal as possible"): This was a key focus. The final image is 36.4MB (a 98% reduction from a 1.8GB naive build). This was achieved with a distroless base, which also runs as a nonroot user for security. See the full analysis in [DOCKER_SIZE_OPTIMIZATION.md](./docs/DOCKER_SIZE_OPTIMIZATION.md).

- Configuration: All hardcoded credentials were removed. The app now follows environment-based configuration (no config files). All settings (database host, port, credentials, server port) come from environment variables, making it cloud ready for any platform.

- JSON Logging: A structured log/slog JSON logging middleware was implemented (internal/middleware/logger.go). It injects a request_id for traceability and automatically logs request/response metadata. See [JSON_LOGGING_IMPLEMENTATION.md](./docs/JSON_LOGGING_IMPLEMENTATION.md) for details.

- Informative Repository: The DELETE and UPDATE functions in the repository report facts (e.g., return ErrUserNotFound if no row was affected).

- Controller-as-Policy: The DeleteUser controller catches this ErrUserNotFound and makes a policy decision to return 204 No Content, preserving HTTP idempotency while still having the information to log the event.

- CI/CD Pipeline: The CI pipeline in .github/workflows/ci.yml runs lint, security scans, tests, and builds in parallel for fast, reliable feedback.

---
## Getting Started

The application requires PostgreSQL. Run it locally with Docker Compose:

```bash
# Start database and application
docker-compose up --build

# In another terminal, run database migrations
make migrate-up
```

The API will be available at `http://localhost:8080/api/v1`

**For local development** (database in Docker, app runs locally):
```bash
make db          # Start PostgreSQL container
make migrate-up  # Run migrations
make run         # Run Go application
```

---

## API Endpoints

All endpoints use base URL: `http://localhost:8080/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Liveness probe (Kubernetes health check) |
| GET | `/ready` | Readiness probe (Kubernetes readiness check) |
| GET | `/users` | Get all users |
| GET | `/users/username/:username` | Get user by username |
| GET | `/users/id/:id` | Get user by UUID |
| POST | `/users` | Create new user |
| PATCH | `/users/id/:id` | Update user by UUID |
| DELETE | `/users/id/:id` | Delete user by UUID |

**Note**: Health probe endpoints (`/health` and `/ready`) are at the root level and do not require authentication.

Example:
```bash
# Create a new user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username": "johndoe", "email": "john@example.com", "full_name": "John Doe"}'
```



## Key Features

- UUID-based primary keys for security and scalability
- Clean architecture with layered design
- Comprehensive input validation
- SQL injection prevention
- Proper HTTP status codes
- Structured error responses
- **X-API-Key authentication** - Optional header-based authentication (401/403 responses)
- **97.5% test coverage** - Comprehensive unit tests
- **JSON structured logging** - Request tracing with unique IDs, latency tracking, automatic log levels
- **Dockerized application** - Production-ready container (36MB)
- **Docker Compose setup** - One-command development environment
- **Multi-stage builds** - Optimized for size and security
- **CI/CD pipeline** - Automated linting, security scanning, testing (golangci-lint + gosec)
- **Kubernetes deployment** - Kubernetes deployment to GKE Autopilot cluster.
- **Neon PostgreSQL** - Neon PostgreSQL as the managed database service.
---

## Docker Implementation

### Architecture

Multi-stage Docker build using Alpine for compilation and Distroless for runtime:

```dockerfile
# Stage 1: Build static binary with Go 1.25 on Alpine
FROM golang:1.25-alpine AS builder
# ... build process ...

# Stage 2: Minimal runtime on distroless
FROM gcr.io/distroless/static-debian12
# ... only binary + migrations ...
```


**Some points about my docker work:**
- Image Size 36MB (98% reduction from 1.87 GB naive build)
- Build Time ~25s (With layer caching)
- Base Image with No shell, no package manager & runs as non-root (uid 65532) to minimize attack surface and security vulnerabilities.
- Optimised while keeping in mind storage, bandwidth and deployment time.
- Static Binaries Enabled to support minimal image.

#### Image Versioning

Images are tagged with commit SHA for immutable, traceable deployments. :

```bash
# Each commit builds a unique image
ghcr.io/fahadaziz44/cruder:2458f78  # Commit SHA tag
```
This will help us to rollback, traceability and reproducibility to any previous version if needed.

**How it works:**
- CI/CD automatically tags images with short commit SHA (`git rev-parse --short HEAD`)
- Kubernetes deployments reference specific SHA tags
- Rollback: `kubectl set image deployment/cruder-app cruder-app=ghcr.io/.../cruder:PREVIOUS_SHA -n production`

### Docker Compose

Full development environment (app + database):

```yaml
services:
  db:
    image: postgres:latest
    healthcheck: pg_isready

  app:
    build: .
    depends_on:
      db:
        condition: service_healthy
```

### Technical Decisions

1. **Distroless over Alpine** for runtime: Smaller attack surface, built-in CA certs
2. **Static binary compilation**: No runtime dependencies, portable
3. **Build cache optimization**: Dependencies cached separately from source code
4. **.dockerignore**: Reduces build context from ~250MB to ~2MB

- [DOCKER_SIZE_OPTIMIZATION.md](./docs/DOCKER_SIZE_OPTIMIZATION.md) - Docker size analysis and optimization

---

## CI/CD Pipeline

Two-workflow architecture following production best practices: **CI for quality gates, CD for deployment**.

#### CI Workflow (`.github/workflows/ci.yml`)
**Triggers**: All branches and PRs
**Jobs** (parallel, ~45s): Lint → Security Scan → Tests → Build Verification

#### CD Workflow (`.github/workflows/deploy.yml`)
**Triggers**: Push to `master` only
**Jobs** (sequential, ~10-15 mins):
1. Build Docker Image → Push to ghcr.io
2. Deploy to Staging (automatic, 2 replicas)
3. Deploy to Production (manual approval, 3 replicas)

---

## Logging

The application uses **structured JSON logging** with automatic request tracing:

**Features:**
- All logs in JSON format (startup, requests, errors)
- Automatic log levels based on HTTP status codes
- Unique Request-ID header added to all responses `X-Request-ID` header
**Log Level**
- INFO: 2xx status codes
- WARN: 4xx status codes
- ERROR: 5xx status codes and request failures

**Example log output:**
```json
{"time":"2025-10-31T16:58:03Z","level":"INFO","msg":"Request completed","request_id":"5ca149c4-a6cc-4fb4-a151-075828504e48","method":"GET","path":"/api/v1/users","status_code":200,"latency":23959166,"client_ip":"::1","user_agent":"curl/8.7.1"}
```

**How to use:**
- **Controllers**: Access request logger via Gin context
- **Services**: Pass logger for important business events
- **Startup/errors**: Use structured logger for consistency

**For detailed logging documentation**, see [JSON_LOGGING_IMPLEMENTATION.md](./docs/JSON_LOGGING_IMPLEMENTATION.md)

---

## Testing
The project has comprehensive unit tests for the service layer.

```bash
# Run all unit tests
make test

# View coverage summary
make coverage

# Generate HTML coverage report
make coverage-html
```

### Run All Validation Checks

```bash
# Run linting + security scan + tests
make validate
```

---

## Authentication

The API supports optional **X-API-Key authentication** for securing endpoints:

**How it works:**
- If `API_KEY` environment variable is set → Authentication is **enabled**
- If `API_KEY` is not set → Authentication is **disabled** (development mode)

**Responses:**
- Valid key → Request proceeds normally
- Missing header → `401 Unauthorized`
- Wrong key → `403 Forbidden`

**Usage:**
```bash
# Enable authentication (add to .env)
API_KEY=your-secret-key-here

# Make authenticated request
curl -H "X-API-Key: {your-secret-key-here}" http://localhost:8080/api/v1/users

```

**Development**: Leave `API_KEY` commented out in `.env` to disable authentication during development.

---

## Configuration
We are using envconfig for configuration. I choose envconfig because of a standard way to handle configuration and it is ease of use.
The benefits of using envconfig are:
- Automatic validation
- Automatic type conversion
- Can enforce required and optional fields
- Default values can be set
- Clear error messages (Catch errors at compile time)
- Self-documenting code via struct tags, One struct shows everything needed for configuration, easy to understand and maintain
- Fail-Fast - App won't start if config is wrong
- Centralized, not scattered across multiple files

Environment-only configuration :
- All configuration via environment variables
- Single source of truth (`.env` file for local development)
- Sensible defaults for non-sensitive config

**Required variables**: `POSTGRES_USER`, `POSTGRES_PASSWORD`
**Optional variables** (with defaults): `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_DB`, `POSTGRES_SSL_MODE`, `PORT`

The application will fail with a clear error message if required environment variables are missing.

---

## Kubernetes Implementation

A complete Local Development Kubernetes deployment setup is available to showcase Kubernetes concepts including Persistent Volumes (PV), Persistent Volume Claims (PVC), StatefulSets, Deployments, Services, Ingress etc.

**Note**: The Kubernetes implementation with Persistent Volume storage for PostgreSQL is available in `./kubernetes/Readme.md` in the branch `feature/k8s_with_persistant-storage`. That branch demonstrates Kubernetes concepts including StatefulSets and persistent storage.

**Production Setup**: In the `master` branch, production deployments use a **managed database service** instead of running PostgreSQL in-cluster for better reliability and reduced operational complexity.

For detailed Kubernetes implementation documentation, see [Kubernetes implementation documentation](./kubernetes/README.md).

---

### Updated: Kubernetes Deployment on GKE Autopilot Cluster with Neon PostgreSQL as the managed database service.


The application is deployed to **Google Kubernetes Engine (GKE) Autopilot** with **Neon PostgreSQL** as the managed database service.

#### Live Deployment

The application is **currently deployed and running** on GKE Autopilot. You can test it directly:

- Production: `http://136.110.146.135`
  - Health check: `http://136.110.146.135/health`
  - API base: `http://136.110.146.135/api/v1`
  
- Staging: `http://34.49.250.233`
  - Health check: `http://34.49.250.233/health`
  - API base: `http://34.49.250.233/api/v1`

**Example:**
```bash
# Test production health endpoint
curl http://136.110.146.135/health

# Test production API
curl http://136.110.146.135/api/v1/users
```

#### Architecture Overview

 

```

┌─────────────────────────────────────────────────┐

│           GKE Autopilot Cluster                 │

│                                                 │

│  ┌──────────────────┐    ┌──────────────────┐  │

│  │ Staging Namespace│    │ Production NS    │  │

│  │                  │    │                  │  │

│  │  GCE Ingress     │    │  GCE Ingress     │  │

│  │  34.49.250.233   │    │  136.110.146.135 │  │

│  │       ↓          │    │       ↓          │  │

│  │  Service (LB)    │    │  Service (LB)    │  │

│  │       ↓          │    │       ↓          │  │

│  │  Deployment      │    │  Deployment      │  │

│  │  (2 replicas)    │    │  (3 replicas)    │  │

│  └────────┬─────────┘    └────────┬─────────┘  │

│           │                       │            │

└───────────┼───────────────────────┼────────────┘

            │                       │

            ↓                       ↓

    ┌───────────────┐       ┌───────────────┐

    │ Neon Dev DB   │       │ Neon Prod DB  │

    │ (SSL required)│       │ (SSL required)│

    └───────────────┘       └───────────────┘

```

 **Edit**: I have scaled down the replicas to 1 for staging and 2 for production to save costs.
 **Edit2**: Rate limiting implemented through google cloud armor policy at the load balancer level.(10 requests per IP per 30 seconds)

#### Key Features

 

- **Managed Database**: Neon PostgreSQL (fully managed, serverless)

  - Staging → Neon development branch

  - Production → Neon production branch

  - SSL/TLS encryption(unencrypted connections are not allowed)

 

- **GKE Autopilot**: Fully managed Kubernetes cluster
  - Built-in security and compliance

- **GCE Ingress Controller**: Built-in Google Cloud Load Balancer

  - No need for nginx ingress controller because GCE ingress controller is built-in and we can use it to route traffic to our application. further more GKE Autopilot gives each namespace a separate load balancer and ip address.
  - Production: `136.110.146.135`
  - Staging: `34.49.250.233`

- **Multi-Environment Setup**:

  - Separate namespaces for staging and production

  - Environment-specific configurations via ConfigMaps

  - Different replica counts (2 for staging, 3 for production)

 

### Changes for GKE Autopilot

 We removed PostgreSQL StatefulSets (replaced with Neon managed database), Persistent Volumes, PVCs (no in-cluster database), NGINX Ingress Controller (GKE provides GCE ingress)

 

We added Neon PostgreSQL connection strings in ConfigMaps, SSL/TLS requirement for database connections, GCE Ingress configuration, GHCR (GitHub Container Registry) for container images, Image pull secrets for private registry access
Spot Pods for Staging.
 

For detailed deployment instructions and manifest files, see [Kubernetes Deployment Guide](./kubernetes/README.md).


### Future Enhancements
- **Rate limiting** - Protect API from abuse and DoS attacks
- **Integration tests** - End-to-end API validation in CI/CD pipeline
- **HTTPS/TLS** - SSL certificates for secure production communication

- **Terraform** - Infrastructure as Code for GKE Autopilot cluster and Neon PostgreSQL
- **Monitoring setup** - Prometheus/Grafana or Google Cloud Monitoring with alerts
- **Automated Database Migrations** - Run migrations as part of deployment pipeline
- **API documentation** - Swagger/OpenAPI for self-documenting API
- **JWT authentication** - Per-user authentication (currently using API key)
- **Pagination** - Handle large user lists efficiently
- **Google Cloud Armor** - WAF and DDoS protection at load balancer level
- **Feature flags** - Gradual rollouts and safe feature deployment

---

## Documentation

Comprehensive documentation available in `docs/`:
- [DOCKER_SIZE_OPTIMIZATION.md](./docs/DOCKER_SIZE_OPTIMIZATION.md) - Docker size analysis and optimization
- [JSON_LOGGING_IMPLEMENTATION.md](./docs/JSON_LOGGING_IMPLEMENTATION.md) - Structured logging implementation
- [Kubernetes Deployment Guide](./kubernetes/README.md) - Kubernetes deployment guide

---