# CRUDER - User Management API

A production-ready RESTful API for user management built with Go, Gin framework, and PostgreSQL. This project demonstrates clean architecture principles with proper separation of concerns.


The original task requirements can be found in [TASK.md](./TASK.md)


## Reviewer's Guide: Engineering Decisions
This project was approached not just as a "checklist" but as an exercise in building a truly production-ready system. Here are some of the key design decisions and where to find them:

- int vs. uuid: The task mentioned :uuid for endpoints, but the initial code used int IDs. To fix this at the foundation, the database migration was corrected to use UUIDs as primary keys for security and scalability.

- Service Layer Tests: The task's example showed an integration test, but the request was for service-layer tests. internal/service/users_test.go implements these as true unit tests, mocking the repository to test business logic (like validation and normalization) in isolation.

- Docker ("Minimal as possible"): This was a key focus. The final image is 36.4MB (a 98% reduction from a 1.8GB naive build). This was achieved with a distroless base, which also runs as a nonroot user for security. See the full analysis in [DOCKER_SIZE_OPTIMIZATION.md](./docs/DOCKER_SIZE_OPTIMIZATION.md).

- Hardcoded DSN: All hardcoded credentials were removed. The app now follows 12-Factor principles, loading non-sensitive config from config.yaml and reading sensitive secrets (DB_USER, DB_PASSWORD) from the environment.

- JSON Logging: A structured log/slog JSON logging middleware was implemented (internal/middleware/logger.go). It injects a request_id for traceability and automatically logs request/response metadata. See [JSON_LOGGING_IMPLEMENTATION.md](./docs/JSON_LOGGING_IMPLEMENTATION.md) for details.

- Informative Repository: The DELETE and UPDATE functions in the repository report facts (e.g., return ErrUserNotFound if no row was affected).

- Controller-as-Policy: The DeleteUser controller catches this ErrUserNotFound and makes a policy decision to return 204 No Content, preserving HTTP idempotency while still having the information to log the event.

- CI/CD Pipeline: The CI pipeline in .github/workflows/ci.yml runs lint, security scans, tests, and builds in parallel for fast, reliable feedback.

---
## Quick Start (Local Development - Recommended)

For fast iteration during development, run the database in Docker and the app locally:

### 1. Start Database
```bash
make db
```

### 2. Run Migrations
```bash
make migrate-up
```

### 3. Run Application
```bash
make run
```

The API will be available at `http://localhost:8080/api/v1`

---

## Docker Compose (Production Testing)

For production-parity testing or CI/CD, use Docker Compose to run the complete containerized environment:

### Start Everything
```bash
# Build and start both database and application
docker-compose up --build

# Or run in background
docker-compose up -d --build
```

### View Logs
```bash
# All services
docker-compose logs -f

# Just application
docker-compose logs -f app

# Just database
docker-compose logs -f db
```

### Stop Everything
```bash
# Stop and remove containers
docker-compose down

# Stop, remove containers, and clean volumes
docker-compose down -v
```

---

## Configuration

**Dynamic configuration separates sensitive data from code:**

**Points**:
- No more hardcoded passwords in source code
- Non-sensitive config in `config.yaml` (host, port, database name)
- Sensitive data in environment variables (username, password)

**How it works**:
```bash
# 1. Copy environment template (first time only)
cp .env.example .env

# 2. Edit .env with your credentials (all variables are required - no fallback)
DB_USER=postgres
DB_PASSWORD=your_secure_password

# 3. Run application - automatically loads config
make run
```

If `DB_USER` and `DB_PASSWORD` are not set in environment variables, the application will fail with a clear error message.

**For detailed configuration documentation**, see [CONFIGURATION_GUIDE.md](./docs/CONFIGURATION_GUIDE.md)

---

## API Endpoints

All endpoints use base URL: `http://localhost:8080/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | Get all users |
| GET | `/users/username/:username` | Get user by username |
| GET | `/users/id/:id` | Get user by UUID |
| POST | `/users` | Create new user |
| PATCH | `/users/id/:id` | Update user by UUID |
| DELETE | `/users/id/:id` | Delete user by UUID |

Example:
```bash
# Create a new user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username": "johndoe", "email": "john@example.com", "full_name": "John Doe"}'
```

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

See [TEST_API_README.md](./TEST_API_README.md) for complete testing documentation.

---

## Key Features

- ✅ UUID-based primary keys for security and scalability
- ✅ Clean architecture with layered design
- ✅ Comprehensive input validation
- ✅ SQL injection prevention
- ✅ Proper HTTP status codes
- ✅ Structured error responses
- ✅ **Dynamic configuration** - Separates sensitive data (env vars) from code (config.yaml)
- ✅ **97.5% test coverage** - Comprehensive unit tests
- ✅ **JSON structured logging** - Request tracing with unique IDs, latency tracking, automatic log levels
- ✅ **Dockerized application** - Production-ready container (36MB)
- ✅ **Docker Compose setup** - One-command development environment
- ✅ **Multi-stage builds** - Optimized for size and security
- ✅ **CI/CD pipeline** - Automated linting, security scanning, testing (golangci-lint + gosec)

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

### Docker Compose Setup

Full development environment (app + database) orchestrated with health checks:

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

### Usage

```bash
# Start everything (one command)
docker-compose up --build

# View logs
docker-compose logs -f app

# Stop
docker-compose down
```

### Technical Decisions

1. **Distroless over Alpine** for runtime: Smaller attack surface, built-in CA certs
2. **Static binary compilation**: No runtime dependencies, portable
3. **Build cache optimization**: Dependencies cached separately from source code
4. **.dockerignore**: Reduces build context from ~250MB to ~2MB

---

## Documentation

Comprehensive documentation available in `docs/`:
- [CI_PIPELINE_GUIDE.md](./docs/CI_PIPELINE_GUIDE.md) - CI/CD pipeline and GitHub Secrets setup
- [CONFIGURATION_GUIDE.md](./docs/CONFIGURATION_GUIDE.md) - Environment configuration and secret management
- [DOCKER_SIZE_OPTIMIZATION.md](./docs/DOCKER_SIZE_OPTIMIZATION.md) - Docker size analysis and optimization
- [JSON_LOGGING_IMPLEMENTATION.md](./docs/JSON_LOGGING_IMPLEMENTATION.md) - Structured logging implementation
- [REPOSITORY_PARTIAL_UPDATES_AND_DELETE.md](./docs/REPOSITORY_PARTIAL_UPDATES_AND_DELETE.md) - Repository layer design decisions
- [SECRETS_MANAGEMENT_SUMMARY.md](./docs/SECRETS_MANAGEMENT_SUMMARY.md) - Secrets management overview

---

## CI/CD Pipeline

Automated quality checks run on every push to master:

**Key Features:**
- CI catches bugs early - Before they reach production
- 4 automated checks - Lint, Security, Tests, Build
- Jobs run in parallel
- Must pass to merge
- Takes ~45 seconds to run in parallel

**Pipeline Jobs** (parallel execution):
- **Code Quality** - golangci-lint (50+ linters including go vet, go fmt, errcheck)
- **Security Scan** - gosec (SQL injection, hardcoded credentials, crypto issues)
- **Unit Tests** - go test with race detection
- **Build Verification** - Ensures code compiles successfully

**Security**: Uses GitHub Secrets for credentials (never hardcoded)

**Configuration**: `.github/workflows/ci.yml`

**Documentation**: See [CI_PIPELINE_GUIDE.md](./docs/CI_PIPELINE_GUIDE.md) for GitHub Secrets setup and detailed pipeline explanation
