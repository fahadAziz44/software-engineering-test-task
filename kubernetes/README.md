# Kubernetes Deployment

Production-ready Kubernetes setup for CRUDER API with multi-environment support.

## What's Implemented

### Multi-Environment Setup
- **2 Namespaces**: `staging` and `production` for complete environment isolation
- **Staging**: 2 replicas, 2Gi storage, debug mode
- **Production**: 3 replicas, 10Gi storage, release mode, zero-downtime deployments

### Core Features
- **Health Probes**: Liveness (`/health`) and Readiness (`/ready`) endpoints with database checks
- **Graceful Shutdown**: preStop hooks (5s staging, 10s production) for zero-downtime rolling updates
- **Persistent Storage**: StatefulSets with PersistentVolumes for PostgreSQL data durability
- **Configuration Management**: Secrets (credentials) and ConfigMaps (app config) separated per environment
- **Resource Limits**: CPU and memory limits to prevent resource exhaustion
- **Security**: Non-root containers, restrictive security contexts, dropped capabilities
- **Ingress Routing**: NGINX Ingress for domain-based HTTP routing (`staging.local`, `api.local`)

### Architecture
```
Ingress (NGINX) → Service → Deployment (Pods) → PostgreSQL (StatefulSet + PVC)
                    ↓                              ↓
              Load Balancer               Persistent Volume
```

## Quick Deploy

```bash
cd kubernetes/manifests

# Deploy in order
kubectl apply -f namespace.yaml
kubectl apply -f persistent-volume.yaml
kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml
kubectl apply -f postgres-statefulset.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# Verify
kubectl get all -n staging
kubectl get all -n production
```

## 📋 Manifest Files

| File | What It Does |
|------|--------------|
| `namespace.yaml` | Creates staging & production namespaces |
| `persistent-volume.yaml` | Storage for PostgreSQL databases |
| `postgres-statefulset.yaml` | PostgreSQL with stable identity & persistent data |
| `secret.yaml` | Database credentials (base64 encoded) |
| `configmap.yaml` | Application configuration (env vars) |
| `deployment.yaml` | Application pods with health probes & resource limits |
| `service.yaml` | ClusterIP services for internal communication |
| `ingress.yaml` | HTTP routing to staging.local and api.local |

## Production Readiness

### Already Implemented
- Health probes (liveness & readiness)
- Graceful shutdown (preStop hooks)
- Resource limits (CPU, memory)
- Persistent storage for databases(I need to look more into this about options for production)
- Multi-environment isolation
- Zero-downtime rolling updates (production)
- Security contexts (non-root, dropped capabilities)

### TODO Before Production

Look for `TODO:` tags in manifest files for production enhancements:
1. TLS/SSL Certificates (`ingress.yaml`)
   - Add Google Cloud Load Balancer for SSL certificates
   - Enable HTTPS with proper TLS configuration
2. Specify POSTGRES_SSL_MODE for production(`configmap.yaml`)
   - Use "require" in real production
3. Resource tuning(`deployment.yaml`)
   - Monitor actual CPU/memory usage
   - Adjust limits based on real workload


## 🔍 Key Configuration Differences

| Setting | Staging | Production |
|---------|---------|------------|
| **Replicas** | 2 | 3 |
| **Storage** | 2Gi | 10Gi |
| **Memory** | 128Mi-256Mi | 256Mi-512Mi |
| **CPU** | 100m-500m | 250m-1000m |
| **Mode** | debug | release |
| **Downtime** | Allowed (50%) | Zero (0 maxUnavailable) |

## Tested On

- Docker Desktop Kubernetes v1.34.1


---

**Note**: For production deployment, review and complete all `TODO:` items in manifest files and this README.
