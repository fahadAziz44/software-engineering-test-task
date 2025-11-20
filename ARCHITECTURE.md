# System Architecture

This document provides a comprehensive overview of the microservice architecture, deployment strategy, and design decisions.

---

## Table of Contents

- [High-Level Architecture](#high-level-architecture)
- [Application Architecture](#application-architecture)
- [Zero-Downtime Deployment Strategy](#zero-downtime-deployment-strategy)
- [Multi-Environment Strategy](#multi-environment-strategy)
- [Database Architecture](#database-architecture)
- [Security Architecture](#security-architecture)
- [Observability & Monitoring](#observability--monitoring)
- [CI/CD Pipeline Architecture](#cicd-pipeline-architecture)
- [Design Decisions & Trade-offs](#design-decisions--trade-offs)

---

## High-Level Architecture

The system follows a cloud-native microservice architecture designed for deployment on Google Kubernetes Engine (GKE Autopilot) with a managed PostgreSQL database.

**Note:** Deployments are currently paused for cost optimization. See [kubernetes/README_GKE.md](./kubernetes/README_GKE.md) for deployment setup.

```
┌────────────────────────────────────────────────────────────────────┐
│                         GitHub Repository                          │
│                                                                    │
│  Code Push → GitHub Actions CI/CD → Docker Image (GHCR)           │
└────────────────────────┬───────────────────────────────────────────┘
                         │
                         ↓
┌────────────────────────────────────────────────────────────────────┐
│                  Google Kubernetes Engine (GKE)                    │
│                                                                    │
│  ┌──────────────────────────┐    ┌──────────────────────────┐    │
│  │   Staging Namespace      │    │  Production Namespace    │    │
│  │                          │    │                          │    │
│  │  ┌────────────────────┐  │    │  ┌────────────────────┐  │    │
│  │  │  GCE Load Balancer │  │    │  │  GCE Load Balancer │  │    │
│  │  │  <STAGING_IP>      │  │    │  │  <PRODUCTION_IP>   │  │    │
│  │  └──────────┬─────────┘  │    │  └──────────┬─────────┘  │    │
│  │             ↓            │    │             ↓            │    │
│  │  ┌────────────────────┐  │    │  ┌────────────────────┐  │    │
│  │  │ K8s Service (LB)   │  │    │  │ K8s Service (LB)   │  │    │
│  │  └──────────┬─────────┘  │    │  └──────────┬─────────┘  │    │
│  │             ↓            │    │             ↓            │    │
│  │  ┌────────────────────┐  │    │  ┌────────────────────┐  │    │
│  │  │   Deployment       │  │    │  │   Deployment       │  │    │
│  │  │                    │  │    │  │                    │  │    │
│  │  │  ┌──────┐ ┌──────┐ │  │    │  │ ┌──────┐ ┌──────┐ │  │    │
│  │  │  │ Pod1 │ │ Pod2 │ │  │    │  │ │ Pod1 │ │ Pod2 │ │  │    │
│  │  │  └──────┘ └──────┘ │  │    │  │ └──────┘ └──────┘ │  │    │
│  │  │                    │  │    │  │          ┌──────┐ │  │    │
│  │  │  2 Replicas        │  │    │  │          │ Pod3 │ │  │    │
│  │  │  Health Probes     │  │    │  │          └──────┘ │  │    │
│  │  │  Graceful Shutdown │  │    │  │                    │  │    │
│  │  └────────────────────┘  │    │  │  3 Replicas        │  │    │
│  │                          │    │  │  Health Probes     │  │    │
│  └────────────┬─────────────┘    │  │  Graceful Shutdown │  │    │
│               │                  │  │  Zero Downtime     │  │    │
│               │                  │  └────────────────────┘  │    │
└───────────────┼──────────────────┴─────────────┬─────────────────┘
                │                                │
                ↓                                ↓
        ┌───────────────┐                ┌───────────────┐
        │  Neon Staging │                │ Neon Production│
        │  PostgreSQL   │                │  PostgreSQL   │
        │  (Dev Branch) │                │ (Prod Branch) │
        │  SSL/TLS      │                │  SSL/TLS      │
        └───────────────┘                └───────────────┘
```

### Components

**Infrastructure Layer:**
- **GKE Autopilot**: Fully managed Kubernetes cluster with automatic node provisioning and scaling
- **Neon PostgreSQL**: Serverless, managed PostgreSQL database with branch-based development
- **GitHub Container Registry (GHCR)**: Private Docker image registry

**Application Layer:**
- **Go Microservice**: RESTful API built with Gin framework
- **Docker Container**: Distroless runtime image (36MB)
- **Kubernetes Deployments**: Rolling update strategy with health probes

**CI/CD Layer:**
- **GitHub Actions**: Automated testing, building, and deployment
- **Quality Gates**: Linting, security scanning, unit tests
- **Progressive Deployment**: Staging → Production with manual approval

---

## Application Architecture

The application follows **Clean Architecture** principles with clear separation of concerns.

### Layer Structure

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Layer (Gin)                     │
│                                                         │
│  Router → Middleware → Controllers                      │
└────────────────────────┬────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────┐
│                   Controller Layer                      │
│                                                         │
│  • HTTP request/response handling                       │
│  • Input validation                                     │
│  • Status code mapping                                  │
│  • Error response formatting                            │
└────────────────────────┬────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────┐
│                    Service Layer                        │
│                                                         │
│  • Business logic                                       │
│  • Data validation & normalization                      │
│  • Error handling                                       │
│  • Transaction coordination                             │
└────────────────────────┬────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────┐
│                  Repository Layer                       │
│                                                         │
│  • Database operations (CRUD)                           │
│  • SQL query execution                                  │
│  • Data mapping (DB ↔ Model)                            │
│  • Connection management                                │
└────────────────────────┬────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────┐
│                  PostgreSQL Database                    │
└─────────────────────────────────────────────────────────┘
```

### Directory Structure

```
.
├── cmd/                    # Application entry point
│   └── main.go            # Server initialization, routing setup
│
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   │   └── config.go     # Environment variable parsing (envconfig)
│   │
│   ├── controller/       # HTTP request handlers
│   │   ├── controllers.go
│   │   ├── users.go      # User CRUD endpoints
│   │   └── health.go     # Health/readiness probes
│   │
│   ├── service/          # Business logic layer
│   │   ├── services.go
│   │   ├── users.go      # User business logic
│   │   └── users_test.go # Unit tests with mocks
│   │
│   ├── repository/       # Data access layer
│   │   ├── repositories.go
│   │   ├── users.go      # SQL operations
│   │   └── connection.go # Database connection
│   │
│   ├── model/            # Domain models
│   │   └── users.go      # User entity definition
│   │
│   ├── middleware/       # HTTP middleware
│   │   ├── logger.go     # Structured JSON logging
│   │   └── auth.go       # Optional API key authentication
│   │
│   ├── handler/          # Router configuration
│   │   └── router.go     # Route registration
│   │
│   └── errors/           # Custom error types
│       └── errors.go     # Domain-specific errors
│
├── migrations/           # Database schema migrations
│   └── 001_create_users_table.sql
│
├── kubernetes/           # Kubernetes manifests
│   └── manifests/       # Deployment, service, ingress configs
│
└── .github/workflows/    # CI/CD pipelines
    ├── ci.yml           # Quality gates
    └── deploy.yml       # Deployment automation
```

### Request Flow Example

**Example: Create User Request**

```
1. HTTP Request
   POST /api/v1/users
   Body: {"username": "alice", "email": "alice@example.com", "full_name": "Alice"}

2. Middleware Layer
   → Logger Middleware: Generate request_id, log request
   → Auth Middleware (optional): Validate X-API-Key header

3. Controller Layer (internal/controller/users.go)
   → Parse & validate JSON body
   → Call service layer: userService.CreateUser(ctx, user)
   → Map result to HTTP response (201 Created or error)

4. Service Layer (internal/service/users.go)
   → Validate business rules (username format, email format)
   → Normalize data (lowercase username, trim whitespace)
   → Call repository: userRepo.CreateUser(ctx, user)
   → Handle repository errors (e.g., duplicate username)

5. Repository Layer (internal/repository/users.go)
   → Execute SQL INSERT with parameterized query
   → Return created user with generated UUID and timestamps
   → Handle database errors (connection, constraints)

6. HTTP Response
   201 Created
   Body: {"id": "uuid...", "username": "alice", ...}
   Header: X-Request-ID: "5ca149c4..."

7. Logger Middleware
   → Log response: status=201, latency=23ms, level=INFO
```

---

## Zero-Downtime Deployment Strategy

The system achieves **zero downtime** during deployments through a combination of Kubernetes features and application design.

### Key Mechanisms

#### 1. Rolling Update Strategy

**Production Configuration:**
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxSurge: 1        # Create 1 new pod before terminating old ones
    maxUnavailable: 0  # Never allow pods to drop below 3
```

**How it works:**
1. Deployment has 3 replicas running (v1)
2. Create 1 new pod with new version (v2) → Total: 4 pods (3×v1, 1×v2)
3. Wait for v2 pod to pass readiness probe
4. Terminate 1 old v1 pod → Total: 3 pods (2×v1, 1×v2)
5. Create another v2 pod → Total: 4 pods (2×v1, 2×v2)
6. Repeat until all pods are v2

**Result:** At least 3 healthy pods running at all times = zero downtime

#### 2. Health Probes

**Liveness Probe** (Pod is alive):
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 15   # Wait 15s before first check
  periodSeconds: 20         # Check every 20s
  timeoutSeconds: 5         # Request timeout
  failureThreshold: 3       # Restart after 3 failures
```

**Readiness Probe** (Pod is ready for traffic):
```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10   # Start checking after 10s
  periodSeconds: 10         # Check every 10s
  timeoutSeconds: 5
  failureThreshold: 3       # Remove from service after 3 failures
```

**Difference:**
- **Liveness**: Unhealthy → restart pod
- **Readiness**: Unhealthy → remove from load balancer (don't restart)

**Implementation:**
```go
// internal/controller/health.go

// /health - Liveness probe (basic check)
func (c *HealthController) Health(ctx *gin.Context) {
    ctx.JSON(200, gin.H{"status": "healthy"})
}

// /ready - Readiness probe (checks database connectivity)
func (c *HealthController) Ready(ctx *gin.Context) {
    // Ping database to ensure connectivity
    if err := c.db.Ping(); err != nil {
        ctx.JSON(503, gin.H{"status": "not ready", "error": "database unavailable"})
        return
    }
    ctx.JSON(200, gin.H{"status": "ready"})
}
```

**Why this matters:**
- New pods aren't added to load balancer until `/ready` returns 200
- Traffic never reaches unhealthy pods
- Deployments with broken database connections fail safely

#### 3. Graceful Shutdown

**PreStop Hook:**
```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 10"]
```

**Shutdown Sequence:**
1. Kubernetes sends SIGTERM to pod
2. PreStop hook executes (sleep 10s)
3. Pod removed from service endpoints (no new requests)
4. Application has 10s to finish in-flight requests
5. After 10s, if still running, SIGKILL sent

**Application Implementation:**
```go
// cmd/main.go - Graceful shutdown handler

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit  // Block until signal received

log.Info("Shutting down server...")

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Error("Server forced to shutdown", "error", err)
}

log.Info("Server exiting")
```

**Result:** In-flight requests complete before pod terminates

#### 4. Automated Smoke Tests

**Post-deployment validation:**
```bash
# .github/workflows/deploy.yml

# Test health endpoint
kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -f http://cruder-service.production.svc.cluster.local/health

# Test readiness endpoint
kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -f http://cruder-service.production.svc.cluster.local/ready
```

**If smoke tests fail:**
- Deployment marked as failed
- Automatic rollback triggered
- Previous version remains in production

### Deployment Timeline

**Example Production Deployment (3 replicas):**

```
Time    Event                                      Healthy Pods
────────────────────────────────────────────────────────────────
0:00    Deploy v2 triggered                        3 (all v1)
0:05    Create new pod (v2)                        3 v1 + 1 v2*
0:15    v2 pod ready (passes /ready probe)         3 v1 + 1 v2
0:16    Terminate 1 v1 pod (preStop: 10s)          2 v1 + 1 v2
0:26    v1 pod terminated, create new v2           2 v1 + 1 v2
0:31    2nd v2 pod ready                           2 v1 + 2 v2
0:32    Terminate 2nd v1 pod                       1 v1 + 2 v2
0:42    2nd v1 pod terminated, create 3rd v2       1 v1 + 2 v2
0:47    3rd v2 pod ready                           1 v1 + 3 v2
0:48    Terminate last v1 pod                      0 v1 + 3 v2
0:58    Deployment complete                        3 v2

* Pod exists but not receiving traffic (readiness probe not passed)
```

**Total deployment time:** ~60 seconds
**Minimum healthy pods:** 3 (throughout entire deployment)
**Downtime:** 0 seconds

---

## Multi-Environment Strategy

The system maintains two isolated environments with different characteristics.

### Environment Comparison

| Aspect | Staging | Production |
|--------|---------|------------|
| **Purpose** | Pre-production validation | Live user traffic |
| **Replicas** | 2 | 3 |
| **CPU** | 100m-500m | 250m-1000m |
| **Memory** | 128Mi-256Mi | 256Mi-512Mi |
| **Database** | Neon dev branch | Neon prod branch |
| **Deployment** | Automatic on merge | Manual approval required |
| **Downtime Tolerance** | 50% (1 pod can be down) | 0% (zero-downtime) |
| **Namespace** | `staging` | `production` |
| **PreStop Hook** | 5s | 10s |

### Configuration Management

**ConfigMaps** (non-sensitive):
```yaml
# kubernetes/manifests/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cruder-config
  namespace: staging
data:
  POSTGRES_HOST: "ep-staging-123.neon.tech"
  POSTGRES_PORT: "5432"
  POSTGRES_DB: "cruder"
  POSTGRES_SSL_MODE: "require"
  PORT: "8080"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cruder-config
  namespace: production
data:
  POSTGRES_HOST: "ep-production-456.neon.tech"
  POSTGRES_PORT: "5432"
  POSTGRES_DB: "cruder"
  POSTGRES_SSL_MODE: "require"
  PORT: "8080"
```

**Secrets** (sensitive):
```yaml
# kubernetes/manifests/secret.yaml (base64 encoded)
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: staging
type: Opaque
data:
  POSTGRES_USER: c3RhZ2luZ191c2Vy        # staging_user
  POSTGRES_PASSWORD: c3RhZ2luZ19wYXNz    # staging_pass
---
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: production
type: Opaque
data:
  POSTGRES_USER: cHJvZF91c2Vy            # prod_user
  POSTGRES_PASSWORD: cHJvZF9wYXNz        # prod_pass
```

### Deployment Flow

```
Developer Push to master
         ↓
    GitHub Actions CI
    (lint, test, scan)
         ↓
    Build Docker Image
    (SHA-tagged: ghcr.io/.../cruder:a1b2c3d)
         ↓
    Push to GHCR
         ↓
┌────────────────────┐
│ Deploy to Staging  │ (Automatic)
│ • kubectl set image│
│ • Rolling update   │
│ • Smoke tests      │
└────────┬───────────┘
         │
         ↓
    ✅ Staging Success
         │
         ↓
┌────────────────────┐
│Manual Approval     │ (GitHub Environment Protection)
│Required            │
└────────┬───────────┘
         │
         ↓
┌────────────────────┐
│Deploy to Production│ (Manual)
│ • kubectl set image│
│ • Zero-downtime    │
│ • Smoke tests      │
└────────┬───────────┘
         │
         ↓
    ✅ Production Success
```

---

## Database Architecture

### Neon PostgreSQL (Managed Database)

**Why Neon:**
- **Serverless**: Automatically scales to zero when inactive
- **Branching**: Separate database branches for staging/production
- **SSL/TLS**: Encrypted connections (required)
- **Backups**: Automatic point-in-time recovery
- **Maintenance**: No manual patching or updates

### Schema Design

**Users Table:**
```sql
-- migrations/001_create_users_table.sql

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

**Design Decisions:**
- **UUID Primary Keys**: Prevents enumeration attacks, globally unique
- **Unique Constraints**: Enforce data integrity at database level
- **Timestamps**: Track creation and modification times
- **Indexes**: Fast lookups by username and email
- **VARCHAR Limits**: Prevent abuse, ensure data quality

### Connection Management

**Connection String:**
```go
// internal/repository/connection.go

connectionString := fmt.Sprintf(
    "postgres://%s:%s@%s:%s/%s?sslmode=%s",
    config.PostgresUser,
    config.PostgresPassword,
    config.PostgresHost,
    config.PostgresPort,
    config.PostgresDB,
    config.PostgresSSLMode,
)

db, err := sql.Open("postgres", connectionString)

// Connection pooling (default settings)
db.SetMaxOpenConns(25)      // Maximum connections
db.SetMaxIdleConns(5)       // Idle connections in pool
db.SetConnMaxLifetime(5 * time.Minute)
```

**Health Checks:**
```go
// Readiness probe checks database connectivity
func (c *HealthController) Ready(ctx *gin.Context) {
    if err := c.db.Ping(); err != nil {
        ctx.JSON(503, gin.H{"status": "not ready"})
        return
    }
    ctx.JSON(200, gin.H{"status": "ready"})
}
```

### Migrations

**Migration Strategy:**
- **Version-controlled**: Migrations stored in `migrations/` directory
- **Sequentially numbered**: `001_`, `002_`, etc.
- **Idempotent**: `CREATE TABLE IF NOT EXISTS`
- **Forward-only**: No automatic rollbacks (manual intervention required)

**Execution:**
```bash
# Local development
make migrate-up

# Production
# Migrations run manually before deployment
kubectl exec -it <pod-name> -n production -- ./migrate up
```

---

## Security Architecture

### Container Security

**Distroless Base Image:**
- No shell (`/bin/sh` doesn't exist)
- No package manager (can't install tools)
- Minimal attack surface (only Go binary + CA certs)
- Non-root user (UID 65532)

**Security Context:**
```yaml
securityContext:
  allowPrivilegeEscalation: false  # Cannot gain additional privileges
  runAsNonRoot: true              # Enforce non-root user
  runAsUser: 65532                # Distroless nonroot user
  capabilities:
    drop:
      - ALL                       # Drop all Linux capabilities
  readOnlyRootFilesystem: false   # Allow temporary files (logs)
```

### Application Security

**SQL Injection Prevention:**
```go
// ❌ Vulnerable (concatenation)
query := "SELECT * FROM users WHERE username = '" + username + "'"

// ✅ Safe (parameterized queries)
query := "SELECT * FROM users WHERE username = $1"
db.QueryRow(query, username)
```

**Input Validation:**
```go
// internal/service/users.go

func validateUsername(username string) error {
    if len(username) < 3 || len(username) > 50 {
        return errors.New("username must be 3-50 characters")
    }
    if !regexp.MustCompile(`^[a-z0-9_]+$`).MatchString(username) {
        return errors.New("username can only contain lowercase letters, numbers, and underscores")
    }
    return nil
}
```

**UUID-based IDs:**
- Prevents enumeration attacks (`/users/1`, `/users/2`, etc.)
- Unpredictable resource identifiers
- Globally unique across distributed systems

### Network Security

**TLS/SSL:**
- Database connections require SSL (`POSTGRES_SSL_MODE=require`)
- Load balancer supports HTTPS (future enhancement)

**Secret Management:**
- Kubernetes Secrets for credentials (base64 encoded)
- Environment variable injection (not hardcoded)
- No secrets in Docker images or Git repository

**Optional API Key Authentication:**
```go
// internal/middleware/auth.go

func APIKeyAuth(apiKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        providedKey := c.GetHeader("X-API-Key")

        if providedKey == "" {
            c.JSON(401, gin.H{"error": "Missing API key"})
            c.Abort()
            return
        }

        if providedKey != apiKey {
            c.JSON(403, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

---

## Observability & Monitoring

### Structured Logging

**JSON Format:**
```json
{
  "time": "2024-11-19T10:30:00Z",
  "level": "INFO",
  "msg": "Request completed",
  "request_id": "5ca149c4-a6cc-4fb4-a151-075828504e48",
  "method": "GET",
  "path": "/api/v1/users",
  "status_code": 200,
  "latency": 23959166,
  "client_ip": "::1",
  "user_agent": "curl/8.7.1"
}
```

**Log Levels:**
- `INFO`: 2xx responses (successful operations)
- `WARN`: 4xx responses (client errors)
- `ERROR`: 5xx responses (server errors)

**Request Tracing:**
- Every request gets unique `request_id`
- ID returned in `X-Request-ID` response header
- Used to correlate logs for a single request

**Implementation:**
```go
// internal/middleware/logger.go

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        start := time.Now()
        c.Next()
        latency := time.Since(start)

        level := slog.LevelInfo
        if c.Writer.Status() >= 500 {
            level = slog.LevelError
        } else if c.Writer.Status() >= 400 {
            level = slog.LevelWarn
        }

        logger.Log(c, level, "Request completed",
            "request_id", requestID,
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status_code", c.Writer.Status(),
            "latency", latency.Nanoseconds(),
        )
    }
}
```

### Health Monitoring

**Kubernetes Health Checks:**
- **Liveness**: `/health` - Is the application running?
- **Readiness**: `/ready` - Can the application serve traffic?

**Metrics (Future Enhancement):**
- Prometheus annotations present in manifests
- Ready for `/metrics` endpoint implementation
- Can track: request rate, latency, error rate, active connections

---

## CI/CD Pipeline Architecture

### CI Workflow (Quality Gates)

**Parallel Execution for Speed:**
```yaml
jobs:
  lint:      # ~10s - golangci-lint
  security:  # ~15s - gosec security scan
  test:      # ~5s  - unit tests
  build:     # ~30s - verify Docker build
```

**Total CI time:** ~45 seconds (parallel execution)

### CD Workflow (Deployment)

**Sequential Stages:**
```
1. Build & Push (5 min)
   • Docker build (multi-stage)
   • Push to GHCR
   • Tag with commit SHA

2. Deploy Staging (3 min)
   • Update deployment image
   • Rolling update (2 replicas)
   • Smoke tests

3. Manual Approval Gate
   • Requires team approval

4. Deploy Production (7 min)
   • Update deployment image
   • Zero-downtime rolling update (3 replicas)
   • Smoke tests
   • Auto-rollback on failure
```

**Total CD time:** ~15 minutes (staging to production)

---

## Design Decisions & Trade-offs

### 1. Distroless vs Alpine Runtime

**Decision:** Use `gcr.io/distroless/static-debian12`

**Rationale:**
- Smaller attack surface (no shell, no package manager)
- Minimal size (similar to Alpine)
- Includes CA certificates (needed for HTTPS)
- Cannot execute arbitrary commands (security)

**Trade-off:** Harder to debug (no shell access to container)

### 2. Managed Database vs In-Cluster PostgreSQL

**Decision:** Use Neon PostgreSQL (managed service)

**Rationale:**
- No operational burden (backups, updates, scaling)
- Built-in SSL/TLS encryption
- Automatic failover and high availability
- Branching for environment isolation

**Trade-off:** External dependency, slight latency increase

### 3. Environment Variables vs Config Files

**Decision:** Environment-only configuration (no config files)

**Rationale:**
- 12-factor app compliance
- Kubernetes-native (ConfigMaps/Secrets)
- No config file management in Docker images
- Environment-specific configs separated

**Trade-off:** More verbose (many env vars vs single config file)

### 4. Rolling Updates vs Blue-Green Deployment

**Decision:** Rolling updates with zero downtime

**Rationale:**
- Resource-efficient (no double infrastructure)
- Gradual rollout (reduces blast radius)
- Built-in Kubernetes support (no custom tooling)
- Fast rollback (previous ReplicaSet still exists)

**Trade-off:** Cannot switch traffic instantly (gradual transition)

### 5. Monorepo vs Separate Repositories

**Decision:** Single repository for application + infrastructure

**Rationale:**
- Atomic changes (code + manifests in one commit)
- Simplified CI/CD (single pipeline)
- Easier versioning (SHA applies to everything)

**Trade-off:** Larger repository, potential for tight coupling

---

## Performance Characteristics

**Expected Performance:**
- **Request Latency**: <50ms (simple CRUD operations)
- **Throughput**: ~1000 req/sec (single pod, non-optimized)
- **Database Queries**: <10ms (indexed lookups)
- **Cold Start**: ~2s (container startup + readiness)
- **Rolling Update**: ~60s (3 replicas, production)

**Scalability:**
- **Horizontal**: Add more pods (Kubernetes HPA)
- **Vertical**: Increase CPU/memory limits
- **Database**: Neon auto-scales connection pooling

**Bottlenecks:**
- Database connections (max 25 per pod)
- Memory limits (256Mi-512Mi)
- Network I/O (GCE Load Balancer)

---

## Future Architectural Improvements

1. **Horizontal Pod Autoscaler (HPA)**: Auto-scale based on CPU/memory
2. **Redis Caching**: Reduce database load for read-heavy endpoints
3. **Message Queue**: Async processing for long-running operations
4. **Service Mesh (Istio)**: Advanced traffic management, observability
5. **Multi-Region Deployment**: Geographic redundancy
6. **CDN**: Static asset delivery (if UI added)
7. **Rate Limiting**: Per-IP throttling at application level
8. **Circuit Breaker**: Prevent cascade failures

---

**Last Updated:** November 2024
**Version:** 1.0
