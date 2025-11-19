# Operational Runbook

This runbook provides step-by-step operational procedures for deploying, monitoring, troubleshooting, and maintaining the microservice in production.

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

# Get GKE credentials
gcloud container clusters get-credentials autopilot-cluster-1 \
  --region=europe-central2 \
  --project=YOUR_PROJECT_ID

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
https://github.com/fahadAziz44/software-engineering-test-task/actions
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
curl http://34.49.250.233/health
curl http://34.49.250.233/ready

# Internal (from within cluster)
kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -f http://cruder-service.staging.svc.cluster.local/health
```

**Production:**
```bash
# External
curl http://136.110.146.135/health
curl http://136.110.146.135/ready

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
kubectl logs -n production -l app=cruder | grep "5ca149c4-a6cc-4fb4"
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
curl http://136.110.146.135/health
curl http://136.110.146.135/ready

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

### Database Backup (Neon)

**Automatic Backups:**
- Neon provides automatic point-in-time recovery (PITR)
- Retention: 7 days (configurable in Neon console)

**Manual Backup:**
```bash
# Export database dump
kubectl run pg-dump --image=postgres:latest --rm -i --restart=Never -- \
  pg_dump "postgres://USER:PASS@HOST:5432/DB?sslmode=require" > backup.sql

# Upload to cloud storage
gsutil cp backup.sql gs://your-backup-bucket/cruder-$(date +%Y%m%d).sql
```

### Database Restore

**Point-in-Time Recovery (Neon Console):**
1. Go to Neon console
2. Select production database
3. Click "Restore" → Choose timestamp
4. Create new branch from backup point

**Restore from SQL dump:**
```bash
# Restore from backup file
kubectl run pg-restore --image=postgres:latest --rm -i --restart=Never -- \
  psql "postgres://USER:PASS@HOST:5432/DB?sslmode=require" < backup.sql
```

---

## Incident Response

### High Severity (P1) - Service Down

**Immediate Actions:**
1. **Assess impact**: Check if staging or production affected
   ```bash
   curl http://136.110.146.135/health
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

## Common Issues & Solutions

### Issue: Database Connection Failures

**Symptoms:**
```json
{"level":"ERROR","msg":"database connection failed","error":"dial tcp: i/o timeout"}
```

**Solutions:**
1. Verify database credentials:
   ```bash
   kubectl get secret postgres-secret -n production -o yaml
   ```

2. Check Neon database status (Neon console)

3. Verify SSL mode is correct:
   ```bash
   kubectl get configmap cruder-config -n production -o yaml | grep SSL
   ```

4. Test connectivity from pod:
   ```bash
   kubectl run test-conn --image=postgres:latest --rm -i --restart=Never -- \
     psql "postgres://USER:PASS@HOST:5432/DB?sslmode=require" -c "SELECT 1"
   ```

---

### Issue: 401/403 Authentication Errors

**Symptoms:**
- External requests getting 401 or 403 responses

**Cause:**
- API_KEY authentication enabled but client not providing key

**Solutions:**
1. **Disable authentication** (development):
   ```bash
   # Remove API_KEY from ConfigMap
   kubectl edit configmap cruder-config -n staging
   # Delete API_KEY line, save
   kubectl rollout restart deployment/cruder-app -n staging
   ```

2. **Provide API key** (production):
   ```bash
   # Get API key from secret
   kubectl get secret api-secret -n production -o jsonpath='{.data.API_KEY}' | base64 -d

   # Use in requests
   curl -H "X-API-Key: YOUR_KEY_HERE" http://136.110.146.135/api/v1/users
   ```

---

### Issue: Deployment Timing Out

**Symptoms:**
- `kubectl rollout status` times out after 10 minutes
- Pods stuck in `ContainerCreating` or `Pending` state

**Investigation:**
```bash
# Check pod events
kubectl describe pod <pod-name> -n production

# Check node resources
kubectl describe nodes | grep -A 5 "Allocated resources"
```

**Common Causes:**
1. **Insufficient cluster resources**: GKE Autopilot provisioning new nodes (wait 5-10 min)
2. **Image pull timeout**: Large image or slow registry (check image size)
3. **Volume mount issues**: PVC not available (not applicable for this app)

**Solutions:**
- Wait for GKE Autopilot to provision nodes
- Optimize Docker image size (already optimized to 36MB)
- Check GKE quota limits in Google Cloud Console

---

### Issue: Readiness Probe Failures

**Symptoms:**
```bash
# Pod running but not ready
NAME                          READY   STATUS    RESTARTS
cruder-app-7d4f8b9c5d-abc12   0/1     Running   0
```

**Investigation:**
```bash
# Check readiness probe logs
kubectl describe pod <pod-name> -n production | grep -A 10 "Readiness"

# Test /ready endpoint from inside pod
kubectl run test-ready --image=curlimages/curl:latest --rm -i --restart=Never -- \
  curl -v http://<POD_IP>:8080/ready
```

**Common Causes:**
1. **Database unreachable**: `/ready` checks database connectivity
2. **Port not open**: Application not listening on port 8080
3. **Slow startup**: Application needs more time (increase initialDelaySeconds)

---

### Issue: Out of Memory (OOMKilled)

**Symptoms:**
```bash
# Pod restarting frequently
kubectl get pods -n production -l app=cruder
# Shows high RESTARTS count

# Check events
kubectl describe pod <pod-name> -n production
# Shows: "Reason: OOMKilled"
```

**Solutions:**
1. **Increase memory limits**:
   ```bash
   kubectl edit deployment cruder-app -n production
   # Increase limits.memory: "1Gi"
   ```

2. **Investigate memory leak**:
   ```bash
   # Check memory usage before OOM
   kubectl top pods -n production -l app=cruder --containers
   ```

3. **Profile application** (locally):
   ```bash
   go test -memprofile=mem.prof
   go tool pprof mem.prof
   ```

---

## Emergency Contacts

**On-Call Rotation:** [Add your on-call schedule link]

**Escalation:**
1. **L1**: DevOps Engineer (first responder)
2. **L2**: Senior Backend Engineer (architecture decisions)
3. **L3**: Engineering Manager (business impact decisions)

**Communication Channels:**
- **Slack**: `#incidents` channel
- **PagerDuty**: [Add PagerDuty integration]
- **Email**: `engineering@company.com`

---

## Maintenance Windows

**Recommended Maintenance Windows:**
- **Staging**: Anytime (low-impact)
- **Production**: Sundays 02:00-06:00 UTC (lowest traffic)

**Pre-Maintenance Checklist:**
- [ ] Announce maintenance in advance (24h notice)
- [ ] Verify rollback procedure is ready
- [ ] Backup database before changes
- [ ] Have at least 2 engineers available
- [ ] Test changes in staging first

**Post-Maintenance Checklist:**
- [ ] Verify all services healthy
- [ ] Check error rates in logs
- [ ] Monitor for 1 hour post-change
- [ ] Update incident log
- [ ] Send completion notification

---

## Useful Commands Cheat Sheet

```bash
# Quick Health Check
kubectl get all -n production
curl http://136.110.146.135/health

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
**Version:** 1.0
**Maintained By:** DevOps Team
