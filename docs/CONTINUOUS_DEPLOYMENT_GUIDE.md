# Continuous Deployment Guide

## Quick Answer: What is Continuous Deployment?

**Continuous Deployment** means code changes **automatically** flow from your repository to your production environment after passing tests.

**For your project**:
- ✅ **Push to `master`** → Automatically deploys to **staging**
- ⏸️ **Manual approval** → Then deploys to **production**

This is actually **Continuous Delivery** (with manual production approval), which is safer than full Continuous Deployment.

---

## Table of Contents
1. [Understanding CD vs CI vs GitOps](#understanding-cd-vs-ci-vs-gitops)
2. [Your CI/CD Pipeline](#your-cicd-pipeline)
3. [Setup Instructions](#setup-instructions)
4. [How It Works](#how-it-works)
5. [Safety Features](#safety-features)
6. [Troubleshooting](#troubleshooting)
7. [Best Practices](#best-practices)

---

## Understanding CD vs CI vs GitOps

### Continuous Integration (CI) - Build & Test

**What happens**:
```
Developer pushes code → GitHub Actions:
                         ├─ Checkout code
                         ├─ Run tests
                         ├─ Run linter
                         ├─ Run security scan
                         ├─ Build Docker image
                         └─ Push to registry

❌ STOPS HERE - No deployment
```

**Benefit**: Catches bugs early, ensures code quality

---

### Continuous Delivery (CD) - Your Pipeline ✅

**What happens**:
```
Developer pushes to master → GitHub Actions:
                            ├─ Run CI pipeline (tests, build)
                            ├─ Deploy to STAGING (automatic)
                            └─ ⏸️  WAIT FOR APPROVAL

Team lead clicks "Approve" → GitHub Actions:
                              └─ Deploy to PRODUCTION
```

**Benefits**:
- ✅ Fast feedback (staging deployed in ~5 minutes)
- ✅ Safe (manual production approval)
- ✅ Confidence (test in staging first)

---

### Continuous Deployment - Full Automation

**What happens**:
```
Developer pushes to master → GitHub Actions:
                            ├─ Run CI pipeline
                            ├─ Deploy to STAGING
                            ├─ Run integration tests
                            ├─ Check health metrics
                            └─ Deploy to PRODUCTION (automatic!)
```

**Requirements**:
- Excellent test coverage (>80%)
- Comprehensive integration tests
- Monitoring and alerting
- Automatic rollback capabilities
- Mature DevOps culture

**Use when**: You have a mature pipeline and high confidence in automated testing

---

### GitOps - Infrastructure as Code

**What happens**:
```
Developer updates manifests in Git → ArgoCD/Flux:
                                      ├─ Detects changes
                                      ├─ Syncs with cluster
                                      └─ Applies manifests

Cluster always matches Git state
```

**Benefits**:
- Git is source of truth
- Declarative deployments
- Easy rollbacks (git revert)
- Audit trail (git history)

**Phase 2 recommendation** after mastering Continuous Delivery

---

## Your CI/CD Pipeline

I've created `.github/workflows/deploy.yml` with 4 jobs:

### Job 1: Test & Quality Checks ✅
**Runs on**: Every push and pull request

```yaml
Steps:
├─ Checkout code
├─ Set up Go
├─ Cache dependencies
├─ Run tests with coverage
├─ Verify coverage > 50%
├─ Run golangci-lint
└─ Run gosec (security scan)
```

**What it does**: Ensures code quality before building

---

### Job 2: Build Docker Image 🐳
**Runs on**: Push to `master` branch (after tests pass)

```yaml
Steps:
├─ Login to GitHub Container Registry
├─ Generate image tags (commit SHA + latest)
├─ Build multi-platform image (linux/amd64)
└─ Push to ghcr.io
```

**Output**: `ghcr.io/fahadaziz44/cruder:abc1234`

---

### Job 3: Deploy to Staging 🚀
**Runs on**: After successful build (automatic)

```yaml
Steps:
├─ Authenticate to GKE
├─ Get cluster credentials
├─ Update deployment image
├─ Wait for rollout
├─ Verify pods are ready
├─ Run smoke tests (health check)
└─ Report success
```

**Time**: ~3-5 minutes
**Trigger**: Automatic (no approval needed)

---

### Job 4: Deploy to Production 🎯
**Runs on**: After staging deployment (manual approval required)

```yaml
Steps:
├─ ⏸️  WAIT FOR MANUAL APPROVAL
├─ Authenticate to GKE
├─ Get cluster credentials
├─ Update deployment image
├─ Wait for rollout (up to 10 mins)
├─ Verify pods are ready
├─ Run smoke tests
└─ Report success
```

**Time**: ~5-10 minutes (after approval)
**Trigger**: Manual click in GitHub UI

---

## Setup Instructions

### Step 1: Create GCP Service Account

Your GitHub Actions need permission to deploy to GKE.

```bash
# 1. Create service account
gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions CI/CD"

# 2. Grant necessary permissions
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/container.developer"

# 3. Create and download key
gcloud iam service-accounts keys create gke-sa-key.json \
  --iam-account=github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com

# 4. Copy the JSON key content (you'll need it for GitHub)
cat gke-sa-key.json
```

**Security**: Never commit this key to Git! Delete after adding to GitHub.

---

### Step 2: Add GitHub Secrets

**Go to**: `https://github.com/YOUR_USERNAME/YOUR_REPO/settings/secrets/actions`

**Add these secrets**:

| Secret Name | Value | Description |
|-------------|-------|-------------|
| `GKE_SA_KEY` | Contents of `gke-sa-key.json` | GCP service account key for deployment |

**How to add**:
1. Click "New repository secret"
2. Name: `GKE_SA_KEY`
3. Value: Paste entire JSON content from `gke-sa-key.json`
4. Click "Add secret"

---

### Step 3: Configure GitHub Environments

**Staging environment** (no approval needed):

1. Go to `Settings → Environments`
2. Click "New environment"
3. Name: `staging`
4. Leave protection rules empty (auto-deploy)
5. Click "Save protection rules"

**Production environment** (requires approval):

1. Go to `Settings → Environments`
2. Click "New environment"
3. Name: `production`
4. Check "Required reviewers"
5. Add your GitHub username (or team)
6. Optionally: Set "Wait timer" (e.g., 5 minutes minimum wait)
7. Click "Save protection rules"

---

### Step 4: Test the Pipeline

**Trigger a deployment**:

```bash
# Make a simple change
echo "# CI/CD Test" >> README.md

# Commit and push
git add README.md
git commit -m "test: trigger CI/CD pipeline"
git push origin master
```

**Watch the pipeline**:
1. Go to: `https://github.com/YOUR_USERNAME/YOUR_REPO/actions`
2. Click on the latest workflow run
3. Watch each job execute

**Expected flow**:
```
✅ Test & Quality Checks (2-3 mins)
    ↓
✅ Build Docker Image (3-5 mins)
    ↓
✅ Deploy to Staging (3-5 mins)
    ↓
⏸️  Deploy to Production (waiting for approval)
```

**Approve production deployment**:
1. In the workflow run, click "Review deployments"
2. Check "production"
3. Click "Approve and deploy"
4. Watch production deployment complete

---

## How It Works

### Workflow Visualization

```
┌─────────────────────────────────────────────────────────┐
│                   Developer                              │
└────────────────────┬────────────────────────────────────┘
                     │ git push origin master
                     ▼
┌─────────────────────────────────────────────────────────┐
│                GitHub Repository                         │
│  Triggers: .github/workflows/deploy.yml                  │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│            Job 1: Test & Quality Checks                  │
│  ├─ Run unit tests                                       │
│  ├─ Check code coverage                                  │
│  ├─ Run linter (golangci-lint)                          │
│  └─ Run security scan (gosec)                           │
└────────────────────┬────────────────────────────────────┘
                     │ ✅ Tests passed
                     ▼
┌─────────────────────────────────────────────────────────┐
│            Job 2: Build Docker Image                     │
│  ├─ Login to ghcr.io                                     │
│  ├─ Build linux/amd64 image                             │
│  ├─ Tag with commit SHA (abc1234)                       │
│  └─ Push to GitHub Container Registry                   │
└────────────────────┬────────────────────────────────────┘
                     │ 🐳 Image ready
                     ▼
┌─────────────────────────────────────────────────────────┐
│          Job 3: Deploy to Staging (Auto)                 │
│  ├─ Authenticate to GKE                                  │
│  ├─ kubectl set image (new SHA)                         │
│  ├─ kubectl rollout status                              │
│  ├─ Run smoke tests                                      │
│  └─ ✅ Staging updated                                   │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│         Job 4: Deploy to Production (Manual)             │
│  ⏸️  Waiting for approval...                             │
│                                                           │
│  [Team Lead clicks "Approve"]                            │
│                                                           │
│  ├─ Authenticate to GKE                                  │
│  ├─ kubectl set image (same SHA as staging)             │
│  ├─ kubectl rollout status                              │
│  ├─ Run smoke tests                                      │
│  └─ 🚀 Production updated                                │
└─────────────────────────────────────────────────────────┘
```

---

### What Gets Deployed

**Same image deployed to both environments**:

```
Staging:  ghcr.io/fahadaziz44/cruder:abc1234
               ↓
          (Test in staging)
               ↓
          (Manual approval)
               ↓
Production: ghcr.io/fahadaziz44/cruder:abc1234 (same image!)
```

**Why this matters**: You're deploying the **exact same artifact** that was tested in staging. No surprises!

---

## Safety Features

### 1. Automatic Rollback on Health Check Failure

```yaml
- name: Verify deployment
  run: |
    kubectl wait --for=condition=ready pod \
      -l app=cruder,environment=production \
      -n production \
      --timeout=600s
```

**If this fails**: Kubernetes automatically rolls back to previous version

---

### 2. Manual Approval for Production

```yaml
environment:
  name: production
  # Requires approval from configured reviewers
```

**Why**: Prevents accidental production deployments

---

### 3. Smoke Tests After Deployment

```yaml
- name: Run smoke tests
  run: |
    kubectl run test-pod --image=curlimages/curl:latest --rm -i --restart=Never -- \
      curl -f http://cruder-service.production.svc.cluster.local/health || exit 1
```

**If health check fails**: Pipeline fails, you can rollback

---

### 4. Test Coverage Enforcement

```yaml
- name: Check test coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 50" | bc -l) )); then
      echo "❌ Test coverage is below 50%"
      exit 1
    fi
```

**Prevents**: Deploying code with insufficient tests

---

### 5. Same Image for Staging and Production

**Ensures**: What you test in staging is what goes to production

---

## Rollback Procedure

### If Deployment Fails

**Automatic**: Kubernetes rolling update will rollback automatically if:
- New pods fail to start
- New pods fail health checks
- Rollout timeout exceeds 10 minutes

**Manual rollback**:

```bash
# View rollout history
kubectl rollout history deployment/cruder-app -n production

# Rollback to previous version
kubectl rollout undo deployment/cruder-app -n production

# Or rollback to specific revision
kubectl rollout undo deployment/cruder-app -n production --to-revision=3
```

---

### If Deployment Succeeds But Has Bugs

**Option 1: Quick rollback** (fastest)
```bash
kubectl rollout undo deployment/cruder-app -n production
```

**Option 2: Fix and redeploy** (recommended)
```bash
# Fix the bug
git commit -m "fix: critical bug in user endpoint"
git push origin master

# Pipeline will automatically:
# 1. Deploy to staging
# 2. Wait for your approval
# 3. Deploy to production (after you approve)
```

---

## Troubleshooting

### Pipeline Fails at "Test & Quality Checks"

**Problem**: Tests failing or coverage too low

**Solution**:
```bash
# Run tests locally
make test

# Check coverage
make coverage

# Fix failing tests
# Increase coverage to > 50%
```

---

### Pipeline Fails at "Build Docker Image"

**Problem**: Docker build errors

**Solution**:
```bash
# Test Docker build locally
docker build --platform linux/amd64 -t test .

# Check Dockerfile syntax
# Ensure all dependencies are available
```

---

### Pipeline Fails at "Deploy to Staging"

**Problem**: GKE authentication or deployment errors

**Check**:
1. Is `GKE_SA_KEY` secret configured correctly?
2. Does service account have `roles/container.developer` permission?
3. Are cluster name and region correct in workflow file?

**Debug**:
```bash
# Test GKE access locally
gcloud container clusters get-credentials autopilot-cluster-1 \
  --region=europe-central2

# Check deployment status
kubectl get deployment cruder-app -n staging
```

---

### Production Deployment Stuck "Waiting for Approval"

**This is normal!** Production deployments require manual approval.

**To approve**:
1. Go to GitHub Actions workflow run
2. Click "Review deployments"
3. Select "production"
4. Click "Approve and deploy"

---

### Smoke Tests Fail

**Problem**: Health or ready endpoints not responding

**Check**:
```bash
# Check pod status
kubectl get pods -n staging

# Check pod logs
kubectl logs deployment/cruder-app -n staging

# Check service
kubectl get service cruder-service -n staging

# Test internally
kubectl run test-pod --image=curlimages/curl --rm -i --restart=Never -- \
  curl -v http://cruder-service.staging.svc.cluster.local/health
```

---

## Best Practices

### 1. Always Test in Staging First ✅

**Never** skip staging deployment:
```yaml
deploy-production:
  needs: [build, deploy-staging]  # ← Production depends on staging
```

---

### 2. Use Feature Branches for Development

**Workflow**:
```bash
# Create feature branch
git checkout -b feature/new-endpoint

# Make changes
git commit -m "feat: add new endpoint"
git push origin feature/new-endpoint

# Create pull request
# → GitHub Actions runs tests (but NO deployment)

# After PR approval, merge to master
# → Triggers deployment to staging → production
```

---

### 3. Write Good Commit Messages

**Why**: Helps track what's deployed

**Format** (Conventional Commits):
```
feat: add user deletion endpoint
fix: resolve database connection pooling issue
chore: update dependencies
docs: add deployment guide
test: add integration tests for users API
```

---

### 4. Monitor Deployments

**During deployment**:
```bash
# Watch pods update
watch kubectl get pods -n production

# Stream logs
kubectl logs -f deployment/cruder-app -n production
```

**After deployment**:
- Check Google Cloud Monitoring
- Verify metrics (request rate, error rate, latency)
- Monitor logs for errors

---

### 5. Maintain Test Coverage

**Keep coverage > 50%** (enforced by pipeline)

```bash
# Check current coverage
make coverage

# Write tests for new features
# Update tests when changing code
```

---

### 6. Tag Releases

**For important deployments**:
```bash
# Tag after successful production deployment
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

---

## Timeline: How Fast Is It?

### From Code to Staging

```
Developer pushes code
    ↓ (0 sec)
GitHub receives push
    ↓ (5-10 sec - GitHub processing)
Test & Quality Checks start
    ↓ (2-3 mins - tests, linting, security scan)
Build Docker Image
    ↓ (3-5 mins - build, push to registry)
Deploy to Staging
    ↓ (3-5 mins - GKE update, health checks)
✅ STAGING UPDATED

Total: 8-13 minutes (fully automated)
```

---

### From Staging to Production

```
Staging deployment complete
    ↓ (0 sec)
Waiting for approval
    ↓ (Variable - depends on team lead)
Approval granted
    ↓ (5-10 mins - GKE update, health checks)
✅ PRODUCTION UPDATED

Total: 5-10 minutes + approval wait time
```

**Comparison to manual deployment**: 15+ minutes → 8-13 minutes (and less error-prone!)

---

## Cost of CI/CD

### GitHub Actions Free Tier

- **Public repos**: Unlimited
- **Private repos**: 2,000 minutes/month (free)

### Typical Usage

**Per deployment**:
- Test & Quality: 2-3 minutes
- Build: 3-5 minutes
- Deploy Staging: 3-5 minutes
- Deploy Production: 5-10 minutes

**Total**: ~15-25 minutes per deployment

**Monthly estimate** (10 deployments/month):
- 10 × 25 = 250 minutes
- Well within free tier (2,000 minutes)

**Cost**: $0 (free tier) ✅

---

## Next Steps

### Phase 1: Basic CI/CD (Start Here) ✅
- ✅ Automated testing
- ✅ Automated staging deployment
- ✅ Manual production approval

### Phase 2: Advanced CI/CD
- [ ] Integration tests (API tests)
- [ ] Database migrations in pipeline
- [ ] Slack/Discord notifications
- [ ] Deployment metrics

### Phase 3: GitOps
- [ ] ArgoCD or Flux setup
- [ ] Manifests in Git
- [ ] Auto-sync cluster state
- [ ] Git-based rollback

---

## Summary

### What You Get

**Before (Manual)**:
- ❌ 15+ minutes per deployment
- ❌ Error-prone (easy to forget steps)
- ❌ No testing before production
- ❌ Stressful production deployments

**After (CI/CD)**:
- ✅ 8-13 minutes to staging (automatic)
- ✅ All changes tested before deployment
- ✅ Same artifact deployed to staging and production
- ✅ Manual approval for production (safe)
- ✅ Automatic health checks
- ✅ Easy rollback

### Your Deployment Flow

```
Push to master → Tests → Build → Deploy Staging → Approve → Deploy Production
      ↓           ↓       ↓          ↓              ↓            ↓
   0 sec      2-3 min  3-5 min    3-5 min       Manual      5-10 min
                                                  Click
```

**Total time**: 13-23 minutes (mostly automated!) 🚀

---

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GKE Deployment Best Practices](https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-workloads-overview)
- [Kubernetes Rolling Updates](https://kubernetes.io/docs/tutorials/kubernetes-basics/update/update-intro/)
- [Conventional Commits](https://www.conventionalcommits.org/)
