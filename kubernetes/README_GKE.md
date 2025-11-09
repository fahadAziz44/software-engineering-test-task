# Kubernetes Deployment

Production deployment on **Google Kubernetes Engine (GKE) Autopilot** with **Neon PostgreSQL** as the managed database service.

## 🚀 Live Deployment

- **Production**: `http://136.110.146.135`
  - Health: `http://136.110.146.135/health`
  - API: `http://136.110.146.135/api/v1`
  
- **Staging**: `http://34.49.250.233`
  - Health: `http://34.49.250.233/health`
  - API: `http://34.49.250.233/api/v1`

## Architecture

```
GKE Autopilot Cluster
│
├── Staging Namespace
│   ├── GCE Ingress (34.49.250.233)
│   ├── Service (Load Balancer)
│   └── Deployment (2 replicas)
│
└── Production Namespace
    ├── GCE Ingress (136.110.146.135)
    ├── Service (Load Balancer)
    └── Deployment (3 replicas)
    
    ↓ (External)
    
Neon PostgreSQL (Managed)
├── Development Branch (Staging)
└── Production Branch (Production)
```

## Key Features

- **GKE Autopilot**: Fully managed Kubernetes (no node management)
- **Neon PostgreSQL**: Serverless managed database with SSL/TLS required
- **GCE Ingress**: Built-in Google Cloud Load Balancer (separate IP per namespace)
- **Multi-Environment**: Isolated staging and production namespaces
- **Health Probes**: Liveness (`/health`) and Readiness (`/ready`) endpoints
- **Graceful Shutdown**: preStop hooks for zero-downtime deployments
- **Resource Limits**: CPU and memory constraints per environment
- **Security**: Non-root containers, restrictive security contexts

## Quick Deploy

```bash
cd kubernetes/manifests

# Deploy in order
kubectl apply -f namespace.yaml
kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# Verify
kubectl get all -n staging
kubectl get all -n production
```

## Manifest Files

| File | Purpose |
|------|---------|
| `namespace.yaml` | Creates staging & production namespaces |
| `secret.yaml` | Database credentials & API keys (base64 encoded) |
| `configmap.yaml` | Application configuration (Neon connection strings, env vars) |
| `deployment.yaml` | Application pods with health probes & resource limits |
| `service.yaml` | ClusterIP services for internal communication |
| `ingress.yaml` | GCE Ingress for external HTTP routing |

## Configuration Differences

| Setting | Staging | Production |
|---------|---------|------------|
| **Replicas** | 2 | 3 |
| **Memory** | 128Mi-256Mi | 256Mi-512Mi |
| **CPU** | 100m-500m | 250m-1000m |
| **Mode** | debug | release |
| **Max Unavailable** | 50% | 0% (zero-downtime) |
| **Database** | Neon Dev Branch | Neon Prod Branch |

##  Decisions

- **Managed Database**: Neon PostgreSQL instead of in-cluster StatefulSet
  - Automatic backups, high availability, SSL/TLS enforced
  - Reduced operational complexity
  
- **GCE Ingress**: Built-in load balancer instead of NGINX Ingress
  - No additional controller needed as it gives Separate IP addresses per namespace.
  
- **GHCR**: Container images from GitHub Container Registry
  - Private registry with image pull secrets

## Tested On

- GKE Autopilot (Production)
- Docker Desktop Kubernetes v1.34.1 (Local testing)

---