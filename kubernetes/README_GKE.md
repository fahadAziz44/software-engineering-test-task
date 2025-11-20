# Kubernetes Deployment

Production deployment on **Google Kubernetes Engine (GKE) Autopilot** with **Neon PostgreSQL** as the managed database service.

## 🚀 Enabling Deployments

**⚠️ Deployment Paused:** The GitHub Actions deployment workflow (`.github/workflows/deploy.yml`) is **paused** to minimize cloud costs. The deployment code remains fully functional.

### **Quick Enable Steps**

1. **Create GKE Cluster** (if needed):
   ```bash
   gcloud container clusters create-auto autopilot-cluster-1 \
     --region=europe-central2 \
     --project=YOUR_PROJECT_ID
   ```

2. **Set up Service Account**:
   ```bash
   gcloud iam service-accounts create github-actions-gke \
     --display-name="GitHub Actions GKE Deployer"
   
   gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
     --member="serviceAccount:github-actions-gke@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/container.developer"
   
   gcloud iam service-accounts keys create key.json \
     --iam-account=github-actions-gke@YOUR_PROJECT_ID.iam.gserviceaccount.com
   ```

3. **Add GitHub Secret**:
   - Repository → Settings → Secrets → Actions → New secret
   - Name: `GKE_SA_KEY`
   - Value: Contents of `key.json`

4. **Enable Workflow**:
   - Edit `.github/workflows/deploy.yml`
   - Remove `if: false` from `deploy-staging` and `deploy-production` jobs
   - Update `GKE_CLUSTER` and `GKE_REGION` environment variables
   - Replace `<STAGING_IP>` and `<PRODUCTION_IP>` placeholders

5. **Deploy Manifests** (see Quick Deploy section below)

**After deployment, get Load Balancer IPs:**
```bash
# Staging
kubectl get svc cruder-service -n staging -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# Production  
kubectl get svc cruder-service -n production -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

**Deployment URLs** (replace with your actual IPs):
- **Production**: `http://<PRODUCTION_IP>` ([Health Check](http://<PRODUCTION_IP>/health))
- **Staging**: `http://<STAGING_IP>` ([Health Check](http://<STAGING_IP>/health))

## Architecture

```
GKE Autopilot Cluster
│
├── Staging Namespace
│   ├── GCE Ingress (<STAGING_IP>)
│   ├── Service (Load Balancer)
│   └── Deployment (2 replicas)
│
└── Production Namespace
    ├── GCE Ingress (<PRODUCTION_IP>)
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