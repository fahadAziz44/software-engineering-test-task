# Operational Runbook

This runbook details production operational procedures for deploying, monitoring, troubleshooting, and maintaining the microservice on Kubernetes. It covers real-world DevOps practices including zero-downtime deployments, health monitoring, and incident response.

**Note:** Replace `<PRODUCTION_IP>` and `<STAGING_IP>` with your actual GKE Load Balancer IPs when deployed.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Deployment Procedures](#deployment-procedures)
- [Monitoring & Health Checks](#monitoring--health-checks)
- [Troubleshooting Guide](#troubleshooting-guide)
- [Rollback Procedures](#rollback-procedures)
- [Scaling Operations](#scaling-operations)
- [Database Operations](#database-operations)
- [Incident Response](#incident-response)
- [Common Issues & Solutions](#common-issues--solutions)

---

## Prerequisites

**⚠️ Note:** The deployment workflows in `.github/workflows/deploy.yml` are **paused** (`if: false`) for cost optimization. The deployment code remains fully functional. To resume automated deployments to GKE, see the [Enabling Deployments section](../kubernetes/README_GKE.md#-enabling-deployments) in `kubernetes/README_GKE.md`.

### Required Access
- **GitHub**: Write access to repository
- **GKE Cluster**: `kubectl` access to staging and production namespaces
- **Google Cloud**: Service account credentials for GKE
- **GHCR**: Read access to GitHub Container Registry

### Required Tools
```bash
# Install kubectl
brew install kubectl  # macOS
# or
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Install gcloud CLI
brew install google-cloud-sdk  # macOS
# or
curl https://sdk.cloud.google.com | bash

# Install gke-gcloud-auth-plugin
gcloud components install gke-gcloud-auth-plugin

# Verify installations
kubectl version --client
gcloud --version
```

### Authenticate to GKE
```bash
# Authenticate to Google Cloud
gcloud auth login

# Get GKE credentials (replace with your cluster details)
gcloud container clusters get-credentials <CLUSTER_NAME> \
  --region=<YOUR_REGION> \
  --project=<YOUR_PROJECT_ID>

# Verify access
kubectl get namespaces
# Should see: staging, production
```

---

## Deployment Procedures

### Automated Deployment (Recommended)

**Trigger:** Push to `master` branch

**Process:**
1. Push code to `master` branch
2. CI workflow runs automatically (lint, test, security scan)
3. CD workflow builds Docker image and deploys to staging
4. **Manual approval required for production**
5. Approve production deployment in GitHub Actions UI
6. Production deployment executes with zero downtime

**GitHub Actions URL:**
```
https://github.com/fahadAziz44/zero-downtime-go-api/actions
```

**Approve Production Deployment:**
1. Go to GitHub Actions → CD workflow run
2. Click "Review deployments"
3. Select "production" environment
4. Click "Approve and deploy"

---

### Manual Deployment

**Use Case:** Emergency hotfix, debugging deployment issues

#### Step 1: Build Docker Image Locally
```bash
# Build image
docker build -t ghcr.io/fahadaziz44/cruder:manual-$(date +%s) .

# Test image locally
docker run -p 8080:8080 \
  -e POSTGRES_HOST=localhost \
  -e POSTGRES_USER=test \
  -e POSTGRES_PASSWORD=test \
  ghcr.io/fahadaziz44/cruder:manual-$(date +%s)

# Push to registry
docker push ghcr.io/fahadaziz44/cruder:manual-$(date +%s)
```

#### Step 2: Deploy to Staging
```bash
# Set image
kubectl set image deployment/cruder-app \
  cruder-app=ghcr.io/fahadaziz44/cruder:manual-1234567890 \
  -n staging

# Watch rollout
kubectl rollout status deployment/cruder-app -n staging

# Verify deployment
kubectl get pods -n staging -l app=cruder
```

#### Step 3: Validate Staging
```bash
# Get external IP
STAGING_IP=$(kubectl get svc cruder-service -n staging -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Test health endpoint
curl http://$STAGING_IP/health

# Test API
curl http://$STAGING_IP/api/v1/users
```

#### Step 4: Deploy to Production
```bash
# Deploy (zero-downtime rolling update)
kubectl set image deployment/cruder-app \
  cruder-app=ghcr.io/fahadaziz44/cruder:manual-1234567890 \
  -n production

# Watch rollout (should complete in ~60s)
kubectl rollout status deployment/cruder-app -n production --timeout=600s

# Verify all replicas are ready
kubectl get pods -n production -l app=cruder

# Run smoke tests
PROD_IP=$(kubectl get svc cruder-service -n production -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -f http://$PROD_IP/health || echo "FAILED"
curl -f http://$PROD_IP/ready || echo "FAILED"
```

---

## Monitoring & Health Checks

### Check Deployment Status
```bash
# Staging
kubectl get deployment cruder-app -n staging
kubectl get pods -n staging -l app=cruder

# Production
kubectl get deployment cruder-app -n production
kubectl get pods -n production -l app=cruder
```

**Expected Output:**
```
NAME         READY   UP-TO-DATE   AVAILABLE   AGE
cruder-app   3/3     3            3           5d
```

### Check Pod Health
```bash
# Get pod details
kubectl get pods -n production -l app=cruder -o wide

# Check pod logs
kubectl logs -n production -l app=cruder --tail=100 --follow

# Check specific pod
kubectl logs -n production <pod-name> --tail=100
```

### Test Health Endpoints

**Staging:**
```bash
# External (via Load Balancer)
curl http://<STAGING_IP>/health
curl http://<STAGING_IP>/ready

# Internal (from within cluster)
kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -f http://cruder-service.staging.svc.cluster.local/health
```

**Production:**
```bash
# External
curl http://<PRODUCTION_IP>/health
curl http://<PRODUCTION_IP>/ready

# Internal
kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -f http://cruder-service.production.svc.cluster.local/health
```

### Check Logs

**Recent logs (all pods):**
```bash
kubectl logs -n production -l app=cruder --tail=50 --timestamps
```

**Follow logs in real-time:**
```bash
kubectl logs -n production -l app=cruder --follow
```

**Filter logs by level:**
```bash
# Errors only
kubectl logs -n production -l app=cruder | grep '"level":"ERROR"'

# Warnings and errors
kubectl logs -n production -l app=cruder | grep -E '"level":"(WARN|ERROR)"'
```

**Search logs for request ID:**
```bash
# Replace <REQUEST_ID> with actual request ID from X-Request-ID header
kubectl logs -n production -l app=cruder | grep "<REQUEST_ID>"
```

### Check Resource Usage
```bash
# CPU and memory usage
kubectl top pods -n production -l app=cruder

# Node resource usage
kubectl top nodes
```

---

## Troubleshooting Guide

### Deployment Stuck or Failing

**Symptoms:**
- Rollout status shows "Waiting for deployment to finish"
- Pods in `CrashLoopBackOff` or `ImagePullBackOff` state

**Investigation:**
```bash
# Check rollout status
kubectl rollout status deployment/cruder-app -n production

# Check deployment events
kubectl describe deployment cruder-app -n production

# Check pod status
kubectl get pods -n production -l app=cruder

# Check pod events (look for errors)
kubectl describe pod <pod-name> -n production

# Check pod logs
kubectl logs <pod-name> -n production
```

**Common Causes & Solutions:**

**1. ImagePullBackOff**
```bash
# Cause: Cannot pull Docker image from GHCR
# Solution: Verify image exists and secrets are correct

# Check image exists
docker pull ghcr.io/fahadaziz44/cruder:TAG

# Verify image pull secret
kubectl get secret ghcr-secret -n production

# Recreate secret if needed
kubectl delete secret ghcr-secret -n production
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_USERNAME \
  --docker-password=YOUR_PAT \
  -n production
```

**2. CrashLoopBackOff**
```bash
# Cause: Application crashes on startup
# Solution: Check logs for startup errors

kubectl logs <pod-name> -n production --previous  # Previous crashed container

# Common issues:
# - Missing environment variables
# - Database connection failure
# - Configuration error

# Verify configuration
kubectl get configmap cruder-config -n production -o yaml
kubectl get secret postgres-secret -n production -o yaml
```

**3. Readiness Probe Failing**
```bash
# Cause: Pod not passing /ready health check
# Solution: Check database connectivity

# Exec into pod (if not distroless)
kubectl exec -it <pod-name> -n production -- /bin/sh

# Test database connectivity from pod
kubectl run test-db --image=postgres:latest --rm -i --restart=Never -- \
  psql "postgres://USER:PASS@HOST:5432/DB?sslmode=require" -c "SELECT 1"
```

---

### Pods Not Receiving Traffic

**Symptoms:**
- Pods running but health checks return errors
- External IP not responding

**Investigation:**
```bash
# Check service
kubectl get svc cruder-service -n production

# Verify endpoints (should list pod IPs)
kubectl get endpoints cruder-service -n production

# Check if pods are ready
kubectl get pods -n production -l app=cruder -o wide
```

**Solution:**
```bash
# If no endpoints, pods are not passing readiness probe
# Check /ready endpoint inside pod
kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -v http://<POD_IP>:8080/ready
```

---

### High Latency or Errors

**Symptoms:**
- Increased response times
- 5xx errors in logs

**Investigation:**
```bash
# Check pod resource usage
kubectl top pods -n production -l app=cruder

# Check if pods are being OOMKilled
kubectl describe pods -n production -l app=cruder | grep -i "OOMKilled"

# Check database connectivity
kubectl logs -n production -l app=cruder | grep "database"

# Check error rates in logs
kubectl logs -n production -l app=cruder --tail=1000 | grep '"level":"ERROR"' | wc -l
```

**Solutions:**

**If CPU/Memory limits reached:**
```bash
# Increase resource limits (edit deployment)
kubectl edit deployment cruder-app -n production

# Or scale horizontally (add more pods)
kubectl scale deployment cruder-app --replicas=5 -n production
```

**If database issues:**
```bash
# Check database connection pool exhaustion
kubectl logs -n production -l app=cruder | grep "connection"

# Verify Neon database status (check Neon console)
# May need to increase connection limits in Neon settings
```

---

## Rollback Procedures

### Automatic Rollback

**Scenario:** Deployment fails smoke tests
- CD pipeline automatically reverts to previous version
- No manual intervention required

### Manual Rollback

**Emergency rollback to previous version:**
```bash
# Rollback to previous ReplicaSet (instant)
kubectl rollout undo deployment/cruder-app -n production

# Watch rollback progress
kubectl rollout status deployment/cruder-app -n production

# Verify previous version is running
kubectl get pods -n production -l app=cruder -o jsonpath='{.items[0].spec.containers[0].image}'
```

**Rollback to specific version:**
```bash
# View rollout history
kubectl rollout history deployment/cruder-app -n production

# Rollback to specific revision
kubectl rollout undo deployment/cruder-app -n production --to-revision=3

# Verify
kubectl rollout status deployment/cruder-app -n production
```

**Rollback to specific image:**
```bash
# Deploy known-good image
kubectl set image deployment/cruder-app \
  cruder-app=ghcr.io/fahadaziz44/cruder:KNOWN_GOOD_SHA \
  -n production

# Wait for rollout
kubectl rollout status deployment/cruder-app -n production
```

**Post-Rollback:**
```bash
# Verify health
curl http://<PRODUCTION_IP>/health
curl http://<PRODUCTION_IP>/ready

# Check logs for errors
kubectl logs -n production -l app=cruder --tail=100

# Monitor for 5-10 minutes to ensure stability
watch kubectl get pods -n production -l app=cruder
```

---

## Scaling Operations

### Horizontal Scaling (Add/Remove Pods)

**Scale up for increased traffic:**
```bash
# Staging (2 → 4 replicas)
kubectl scale deployment cruder-app --replicas=4 -n staging

# Production (3 → 5 replicas)
kubectl scale deployment cruder-app --replicas=5 -n production

# Verify
kubectl get deployment cruder-app -n production
```

**Scale down after traffic decreases:**
```bash
# Return to baseline
kubectl scale deployment cruder-app --replicas=3 -n production

# Verify
kubectl get pods -n production -l app=cruder
```

### Vertical Scaling (Increase Resources)

**Edit deployment to increase CPU/memory:**
```bash
# Edit deployment manifest
kubectl edit deployment cruder-app -n production

# Update resources section:
resources:
  requests:
    memory: "512Mi"  # Increased from 256Mi
    cpu: "500m"      # Increased from 250m
  limits:
    memory: "1Gi"    # Increased from 512Mi
    cpu: "2000m"     # Increased from 1000m

# Save and exit
# Kubernetes will perform rolling update with new resources
```

**Verify resource changes:**
```bash
kubectl describe pod <pod-name> -n production | grep -A 10 "Limits"
```

---

## Database Operations

### Run Migrations

**Prerequisites:**
- Migration files in `migrations/` directory
- Database credentials in Kubernetes secrets

**Execute migrations:**
```bash
# Option 1: From local machine
# Create connection string from secrets
kubectl get secret postgres-secret -n production -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d

# Run migrations locally
DATABASE_URL="postgres://USER:PASS@HOST:5432/DB?sslmode=require" make migrate-up

# Option 2: From within pod
kubectl exec -it <pod-name> -n production -- /app/migrate up
```



---

## Incident Response

### High Severity (P1) - Service Down

**Immediate Actions:**
1. **Assess impact**: Check if staging or production affected
   ```bash
   curl http://<PRODUCTION_IP>/health
   kubectl get pods -n production -l app=cruder
   ```

2. **Check recent changes**: Was there a recent deployment?
   ```bash
   kubectl rollout history deployment/cruder-app -n production
   ```

3. **Rollback if recent deployment** (within last hour):
   ```bash
   kubectl rollout undo deployment/cruder-app -n production
   ```

4. **Check logs for errors**:
   ```bash
   kubectl logs -n production -l app=cruder --tail=100 | grep "ERROR"
   ```

5. **Notify stakeholders**: Post in incident channel

6. **Document timeline**: Record all actions and observations

### Medium Severity (P2) - Degraded Performance

**Investigation Steps:**
1. **Check resource usage**:
   ```bash
   kubectl top pods -n production -l app=cruder
   ```

2. **Check database performance**: Review Neon metrics

3. **Analyze logs for slow queries**:
   ```bash
   kubectl logs -n production -l app=cruder | grep "latency"
   ```

4. **Scale if needed**:
   ```bash
   kubectl scale deployment cruder-app --replicas=5 -n production
   ```

### Low Severity (P3) - Intermittent Errors

**Investigation Steps:**
1. **Collect logs for analysis**:
   ```bash
   kubectl logs -n production -l app=cruder --since=1h > incident-logs.txt
   ```

2. **Check for patterns**: Specific endpoints, users, or times

3. **Monitor for escalation**: Watch error rates

4. **Schedule fix**: Plan deployment during low-traffic window



---

## Useful Commands Cheat Sheet

```bash
# Quick Health Check
kubectl get all -n production
curl http://<PRODUCTION_IP>/health

# View Logs
kubectl logs -n production -l app=cruder --tail=100 --follow

# Describe Deployment
kubectl describe deployment cruder-app -n production

# Scale Replicas
kubectl scale deployment cruder-app --replicas=5 -n production

# Restart Deployment (rolling restart)
kubectl rollout restart deployment/cruder-app -n production

# Rollback Deployment
kubectl rollout undo deployment/cruder-app -n production

# Update Image
kubectl set image deployment/cruder-app cruder-app=ghcr.io/fahadaziz44/cruder:NEW_SHA -n production

# Get External IP
kubectl get svc cruder-service -n production -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# Exec into Pod (if shell available)
kubectl exec -it <pod-name> -n production -- /bin/sh

# Port Forward (local testing)
kubectl port-forward svc/cruder-service 8080:8080 -n production

# Watch Pods
watch kubectl get pods -n production -l app=cruder
```

---

**Last Updated:** November 2024  
