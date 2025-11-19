# Production-Ready Go Microservice Template

**A production-grade microservice architecture demonstrating zero-downtime Kubernetes deployments, automated CI/CD pipelines, and cloud-native best practices.**

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-GKE%20Autopilot-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)
[![Docker Image Size](https://img.shields.io/badge/Docker-36MB-2496ED?style=flat&logo=docker)](https://github.com/fahadAziz44/software-engineering-test-task/pkgs/container/cruder)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🎯 **What This Demonstrates**

This project showcases **production-ready backend engineering** with real-world deployment on Google Kubernetes Engine:

- ✅ **Zero-Downtime Deployments** - Rolling updates with health probes and graceful shutdown
- ✅ **Multi-Environment Architecture** - Isolated staging and production namespaces
- ✅ **Automated CI/CD Pipeline** - Quality gates, security scanning, progressive deployment
- ✅ **Cloud-Native Design** - Running live on GKE Autopilot with managed PostgreSQL
- ✅ **Observability** - Structured JSON logging with request tracing
- ✅ **Docker Optimization** - 98% size reduction (1.87GB → 36MB)

**Live Deployment:**
- 🌐 Production: `http://136.110.146.135` ([Health Check](http://136.110.146.135/health))
- 🔧 Staging: `http://34.49.250.233` ([Health Check](http://34.49.250.233/health))

---

## 🏗️ **Architecture Overview**

```
┌─────────────────────────────────────────────────────────┐
│                  GitHub Actions CI/CD                   │
│                                                         │
│  Lint → Security Scan → Tests → Build → Deploy         │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────┐
│              Google Kubernetes Engine (GKE)             │
│                                                         │
│  ┌──────────────────┐      ┌──────────────────┐        │
│  │ Staging Namespace│      │Production Namespace│       │
│  │                  │      │                  │        │
│  │  Load Balancer   │      │  Load Balancer   │        │
│  │  34.49.250.233   │      │  136.110.146.135 │        │
│  │       ↓          │      │       ↓          │        │
│  │  Service (LB)    │      │  Service (LB)    │        │
│  │       ↓          │      │       ↓          │        │
│  │  Deployment      │      │  Deployment      │        │
│  │  • 2 Replicas    │      │  • 3 Replicas    │        │
│  │  • Health Probes │      │  • Health Probes │        │
│  │  • Auto Rollback │      │  • Zero Downtime │        │
│  └────────┬─────────┘      └────────┬─────────┘        │
│           │                         │                  │
└───────────┼─────────────────────────┼──────────────────┘
            │                         │
            ↓                         ↓
    ┌───────────────┐         ┌───────────────┐
    │  Neon Staging │         │  Neon Production│
    │  PostgreSQL   │         │  PostgreSQL   │
    │  (SSL/TLS)    │         │  (SSL/TLS)    │
    └───────────────┘         └───────────────┘
```

---

## 🚀 **Key Features**

### **Zero-Downtime Deployments**
- **Rolling updates** with `maxUnavailable: 0` in production
- **Health probes** prevent traffic to unhealthy pods (liveness + readiness)
- **Graceful shutdown** with configurable preStop hooks
- **Automated smoke tests** validate deployments before traffic routing
- **Instant rollback** on deployment failure

### **Production-Grade CI/CD**
- **Parallel quality gates**: Linting, security scanning (gosec), unit tests
- **Progressive deployment**: Staging (automatic) → Production (manual approval)
- **Immutable deployments**: SHA-tagged Docker images for traceability
- **Environment isolation**: Separate namespaces, configs, and databases
- **Automated rollback**: Failed deployments revert automatically

### **Cloud-Native Architecture**
- **GKE Autopilot**: Fully managed Kubernetes with auto-scaling
- **Managed Database**: Neon PostgreSQL with SSL/TLS encryption
- **Secret Management**: Kubernetes secrets (not hardcoded credentials)
- **Resource Optimization**: CPU/memory limits prevent resource exhaustion
- **Security Hardening**: Non-root containers, minimal attack surface

### **Observability & Monitoring**
- **Structured JSON logging** with automatic log levels (INFO/WARN/ERROR)
- **Request tracing** via unique `X-Request-ID` headers
- **Health endpoints**: `/health` (liveness) and `/ready` (readiness)
- **Prometheus-ready**: Annotations for metrics scraping
- **Latency tracking**: Automatic request duration logging

### **Docker Optimization**
- **36MB final image** (98% reduction from naive 1.87GB build)
- **Multi-stage distroless build** for minimal attack surface
- **Static binary compilation** (no runtime dependencies)
- **Security**: Runs as non-root user (uid 65532)
- **Build caching**: Optimized layer structure for fast rebuilds

---

## 📦 **Tech Stack**

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.23+ | High-performance backend |
| **Web Framework** | Gin | HTTP routing and middleware |
| **Database** | PostgreSQL (Neon) | Managed, serverless SQL database |
| **Container** | Docker (Distroless) | Minimal, secure runtime |
| **Orchestration** | Kubernetes (GKE Autopilot) | Zero-downtime deployments |
| **CI/CD** | GitHub Actions | Automated testing and deployment |
| **Logging** | log/slog | Structured JSON logging |
| **Registry** | GitHub Container Registry (GHCR) | Docker image storage |

---

## 🏁 **Quick Start**

### **Local Development**

```bash
# Clone the repository
git clone https://github.com/fahadAziz44/software-engineering-test-task.git
cd software-engineering-test-task

# Start database and application with Docker Compose
docker-compose up --build

# Run database migrations (in another terminal)
make migrate-up

# The API will be available at http://localhost:8080
```

### **Test the Live Deployment**

```bash
# Production health check
curl http://136.110.146.135/health

# List all users
curl http://136.110.146.135/api/v1/users

# Create a user
curl -X POST http://136.110.146.135/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "full_name": "John Doe"
  }'

# Get user by username
curl http://136.110.146.135/api/v1/users/username/johndoe
```

### **Development Workflow**

```bash
# Run all validation checks (lint, security, tests)
make validate

# Run tests with coverage
make coverage

# Build Docker image locally
make docker-build

# Run application locally (DB in Docker)
make db          # Start PostgreSQL container
make migrate-up  # Run migrations
make run         # Start Go application
```

---

## 📚 **API Endpoints**

**Base URL:** `http://localhost:8080/api/v1` (local) or `http://136.110.146.135/api/v1` (production)

| Method | Endpoint | Description |
|--------|----------|-------------|
| **GET** | `/health` | Liveness probe (Kubernetes) |
| **GET** | `/ready` | Readiness probe (database connectivity) |
| **GET** | `/users` | List all users |
| **GET** | `/users/username/:username` | Get user by username |
| **GET** | `/users/id/:id` | Get user by UUID |
| **POST** | `/users` | Create new user |
| **PATCH** | `/users/id/:id` | Update user by UUID |
| **DELETE** | `/users/id/:id` | Delete user by UUID |

**Example Request:**
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "full_name": "Alice Johnson"
  }'
```

**Example Response:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "alice",
  "email": "alice@example.com",
  "full_name": "Alice Johnson",
  "created_at": "2024-11-19T10:30:00Z",
  "updated_at": "2024-11-19T10:30:00Z"
}
```

---

## 🔒 **Security Features**

- **UUID-based primary keys** (prevents enumeration attacks)
- **SQL injection prevention** (parameterized queries)
- **Input validation** (username, email, full_name constraints)
- **Non-root containers** (uid 65532, dropped capabilities)
- **Minimal Docker images** (distroless, no shell, no package manager)
- **TLS/SSL database connections** (Neon PostgreSQL requires encryption)
- **Optional X-API-Key authentication** (header-based access control)
- **Secret management** (Kubernetes secrets, not hardcoded)

---

## 🛠️ **CI/CD Pipeline**

### **CI Workflow** (`.github/workflows/ci.yml`)
**Triggers:** All branches and pull requests
**Duration:** ~45 seconds (parallel execution)

```
Lint (golangci-lint) ──┐
                       ├──> Quality Gate
Security Scan (gosec) ─┤
                       │
Unit Tests ────────────┤
                       │
Build Verification ────┘
```

### **CD Workflow** (`.github/workflows/deploy.yml`)
**Triggers:** Push to `master` only
**Duration:** ~10-15 minutes

```
Build & Push Docker Image (SHA-tagged)
            ↓
Deploy to Staging (automatic)
  • Update image with SHA tag
  • Rolling update (2 replicas)
  • Smoke tests (/health, /ready)
            ↓
Deploy to Production (manual approval required)
  • Update image with SHA tag
  • Zero-downtime rolling update (3 replicas)
  • Smoke tests (/health, /ready)
  • Auto-rollback on failure
```

**Key Features:**
- **Immutable deployments**: Every commit creates a unique SHA-tagged image
- **Progressive rollout**: Staging validates changes before production
- **Automated validation**: Health checks prevent bad deployments
- **Traceability**: Know exactly which commit is running in each environment

---

## 📊 **Kubernetes Deployment**

### **Multi-Environment Strategy**

| Setting | Staging | Production |
|---------|---------|------------|
| **Replicas** | 2 | 3 |
| **Downtime Tolerance** | 50% (1 pod) | 0% (zero-downtime) |
| **Memory** | 128Mi-256Mi | 256Mi-512Mi |
| **CPU** | 100m-500m | 250m-1000m |
| **Database** | Neon dev branch | Neon production branch |
| **Deployment** | Automatic | Manual approval |

### **Zero-Downtime Configuration**

**Production deployment strategy:**
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1        # Create 1 extra pod during update
    maxUnavailable: 0  # Never drop below 3 running pods
```

**Health probes prevent bad deployments:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 15
  periodSeconds: 20
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  failureThreshold: 3
```

**Graceful shutdown prevents connection drops:**
```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 10"]  # Drain connections
```

---

## 📖 **Documentation**

Comprehensive documentation is available in the `docs/` directory:

- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System design and zero-downtime deployment strategy
- **[RUNBOOK.md](./RUNBOOK.md)** - Operational guide for deployments and troubleshooting
- **[DOCKER_SIZE_OPTIMIZATION.md](./docs/DOCKER_SIZE_OPTIMIZATION.md)** - Docker optimization journey (1.87GB → 36MB)
- **[JSON_LOGGING_IMPLEMENTATION.md](./docs/JSON_LOGGING_IMPLEMENTATION.md)** - Structured logging design
- **[Kubernetes Deployment Guide](./kubernetes/README.md)** - Kubernetes manifest documentation
- **[Project Origin](./docs/assignment/README.md)** - How this project evolved from an assignment

---

## 🧪 **Testing**

```bash
# Run all unit tests
make test

# Run tests with coverage report
make coverage

# Generate HTML coverage report
make coverage-html
open coverage.html
```

**Test Coverage:** Service layer has comprehensive unit tests following the Given-When-Then pattern.

**Example test:**
```go
// Given: A valid user exists in the repository
func TestGetByID_Success(t *testing.T) {
    // When: Fetching user by ID
    user, err := service.GetByID(ctx, validID)

    // Then: User is returned without error
    assert.NoError(t, err)
    assert.Equal(t, "johndoe", user.Username)
}
```

---

## ⚙️ **Configuration**

The application uses **environment-based configuration** with validation and fail-fast behavior.

**Required Environment Variables:**
```bash
POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password
```

**Optional Environment Variables (with defaults):**
```bash
POSTGRES_HOST=localhost      # Database host
POSTGRES_PORT=5432          # Database port
POSTGRES_DB=cruder          # Database name
POSTGRES_SSL_MODE=disable   # SSL mode (use 'require' in production)
PORT=8080                   # Application port
API_KEY=                    # Optional API key for authentication
```

**Development Setup:**
1. Copy `.env.example` to `.env`
2. Update `POSTGRES_USER` and `POSTGRES_PASSWORD`
3. Start the application with `docker-compose up`

**Production Setup:**
- Configuration is managed via Kubernetes ConfigMaps and Secrets
- Sensitive credentials (database password, API keys) are stored in Kubernetes Secrets
- Non-sensitive config (database host, port) is stored in ConfigMaps

---

## 🔐 **Authentication** (Optional)

The API supports optional **X-API-Key authentication**:

**Enable authentication:**
```bash
# Add to .env file
API_KEY=your-secret-key-here
```

**Make authenticated requests:**
```bash
curl -H "X-API-Key: your-secret-key-here" \
  http://localhost:8080/api/v1/users
```

**Responses:**
- ✅ Valid key → Request proceeds
- ❌ Missing header → `401 Unauthorized`
- ❌ Wrong key → `403 Forbidden`

**Development mode:** Leave `API_KEY` unset to disable authentication during local development.

---

## 🎓 **What This Project Teaches**

This template demonstrates real-world backend engineering practices:

### **Backend Development**
- Clean architecture with layered design (controller → service → repository)
- Proper error handling and HTTP status codes
- Input validation and security best practices
- Database migrations and schema management
- Unit testing with mocks and dependency injection

### **DevOps & Cloud**
- Zero-downtime deployment strategies
- Multi-environment Kubernetes architecture
- CI/CD pipeline design and automation
- Docker optimization and security hardening
- Infrastructure as Code with Kubernetes manifests

### **Production Operations**
- Health probes and graceful shutdown
- Structured logging for observability
- Secret management and configuration
- Rollback strategies and incident response
- Resource management and auto-scaling

---

## 🚧 **Future Enhancements**

Potential improvements to make this even more production-ready:

- [ ] **HTTPS/TLS** - SSL certificates for secure communication
- [ ] **Rate Limiting** - Protect API from abuse (currently implemented at LB level via Cloud Armor)
- [ ] **Monitoring** - Prometheus/Grafana dashboards with alerts
- [ ] **Terraform** - Infrastructure as Code for GKE and Neon
- [ ] **Integration Tests** - End-to-end API validation in CI/CD
- [ ] **Database Backups** - Automated backup and restore procedures
- [ ] **API Documentation** - Swagger/OpenAPI specification
- [ ] **JWT Authentication** - Per-user authentication (currently using API key)
- [ ] **Pagination** - Handle large datasets efficiently
- [ ] **Feature Flags** - Gradual rollouts and safe feature deployment

---

## 📄 **License**

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 **Acknowledgments**

This project evolved from a technical assessment into a comprehensive exploration of production-grade backend architecture. It represents the type of system I'd build for real-world use, with all the operational considerations that come with running services in production.

**Key Learnings:**
- How to achieve zero-downtime deployments with Kubernetes
- The importance of health probes and graceful shutdown
- Docker optimization techniques (98% size reduction)
- Progressive deployment strategies (staging → production)
- Structured logging for production observability
- Security hardening at every layer

---

## 📬 **Contact**

**Built by:** Fahad Aziz
**GitHub:** [@fahadAziz44](https://github.com/fahadAziz44)

---

**⭐ If you find this useful, please consider giving it a star!**
