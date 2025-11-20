# Docker Image Optimization

**Goal**: Create a minimal, secure Docker image for production deployment.

---

## Results

| Approach | Base Image | Size | Reduction |
|----------|-----------|------|-----------|
| Naive (single-stage) | `golang:1.25` | 1.87 GB | Baseline |
| **Optimized (multi-stage)** | `gcr.io/distroless/static-debian12` | **36.1 MB** | **98% smaller** |

**Savings**: 1.83 GB reduction

---

## Implementation

### Multi-Stage Build

**Stage 1 - Builder** (golang:1.25-alpine):
- Compiles Go binary with optimization flags
- Strips debug symbols (`-ldflags="-w -s"`)
- Creates static binary (`CGO_ENABLED=0`)

**Stage 2 - Runtime** (distroless):
- Copies only the compiled binary
- Includes timezone data (accurate logging)
- Runs as non-root user (security)

### Key Optimizations

1. **Multi-stage build**: Separates build tools from runtime
2. **Static binary**: No runtime dependencies (`CGO_ENABLED=0`)
3. **Stripped symbols**: Remove debug info (`-ldflags="-w -s"`)
4. **Distroless base**: Minimal attack surface (no shell, no package manager)
5. **Non-root user**: Security best practice

---

## Size Breakdown

```
Component              Size        % of Total
────────────────────────────────────────────
Go Binary (stripped)   20.5 MB     56.3%
Distroless Base        ~14 MB      38.5%
Timezone Data          1.55 MB     4.3%
Migration Files        12.3 KB     0.03%
────────────────────────────────────────────
TOTAL                  36.1 MB     100%
```

---

## Security Benefits

- No shell or debugging tools (reduced attack surface)
- Non-root user by default
- Minimal dependencies (only what's needed)
- Distroless provides only essential runtime files

---

## Verification

```bash
# Check image size
docker images ghcr.io/fahadaziz44/cruder:latest

# Inspect layers
docker history ghcr.io/fahadaziz44/cruder:latest --human

# Test the image
docker run -e POSTGRES_USER=user -e POSTGRES_PASSWORD=pass \
  -e POSTGRES_HOST=localhost -e POSTGRES_PORT=5432 \
  -e POSTGRES_DB=db ghcr.io/fahadaziz44/cruder:latest
```

---

## Key Learnings

1. Multi-stage builds are essential - Never ship build tools to production
2. Base image choice matters - Distroless provides optimal size/security balance
3. Static binaries enable minimal images - Go's `CGO_ENABLED=0` is powerful
4. Security and size often align - Fewer components = smaller image + smaller attack surface

---

## Base Image Selection

Evaluated multiple options for runtime base image:

| Base Image | Size | Has CA Certs? | Pros | Cons |
|-----------|------|---------------|------|------|
| `golang:1.25` | ~1.8 GB | Yes | Everything included | Way too large for runtime |
| `alpine:latest` | ~5 MB | Yes | Small, popular | Has shell/package manager (attack surface) |
| `distroless/static` | ~2 MB | Yes | Minimal, no shell, includes essentials | Harder to debug |
| `scratch` | 0 MB | No | Ultimate minimal | No CA certs → PostgreSQL TLS fails |

**Decision: Distroless**

Chose `gcr.io/distroless/static-debian12` because:
- Includes CA certificates (for PostgreSQL TLS connections)
- Includes timezone data (accurate logging timestamps)
- No shell or package managers (minimal attack surface)
- Optimal size/security balance

**Why not scratch?** The application can connect to PostgreSQL with `sslmode=require`, which requires CA certificates to verify the server's TLS certificate. Scratch has no files, causing `certificate signed by unknown authority` errors.

---