# Deployment Verification Fix

**Date**: 2025-11-13  
**Status**: Resolved  
**Severity**: Medium  
**Impact**: CI/CD pipeline failures during rolling updates (no production downtime)

---

## Summary

Fixed CI/CD deployment verification step that was failing during Kubernetes rolling updates due to race conditions with terminating pods.

---

## Problem

- CI/CD pipeline failing during deployment verification step
- Error: `pod not found` or `pod terminated` errors
- Affected both staging and production deployments
- Occurred specifically during rolling updates

### Root Cause

The original verification approach used `kubectl wait` with pod label selectors:

```bash
# PROBLEMATIC APPROACH (removed)
kubectl wait --for=condition=ready pod -l app=cruder -n production --timeout=600s
```

Why it failed:

1. Label selector matches multiple deployment labels: During rolling updates, Kubernetes creates a new ReplicaSet while the old one is terminating
2. Race condition: `kubectl wait` tries to wait for pods that are being terminated
3. "Pod not found" errors: Terminating pods disappear before `kubectl wait` can check their status

### Example Error

```
Error from server (NotFound): pods "cruder-app-abc123" not found
```

This happened because the pod was from the old ReplicaSet and was already terminated.

---

## Solution


```bash
# Get deployment status (only counts pods from current ReplicaSet)
READY_REPLICAS=$(kubectl get deployment cruder-app -n production -o jsonpath='{.status.readyReplicas}')
DESIRED_REPLICAS=$(kubectl get deployment cruder-app -n production -o jsonpath='{.spec.replicas}')

# Verify all replicas are ready
if [ "$READY_REPLICAS" -eq "$DESIRED_REPLICAS" ]; then
  echo "✅ All replicas are ready"
else
  echo "❌ Not all replicas are ready"
  exit 1
fi
```

---

## Technical Details

### Kubernetes Rolling Update Behavior

During a rolling update:

1. New ReplicaSet created: With new image
2. Old ReplicaSet scaled down: Pods terminated gradually
3. Both ReplicaSets exist temporarily: During the transition
4. Pod labels identical: Both use `app=cruder` label

### Why Pod-Level Checks Fail

```bash
# This matches pods from BOTH ReplicaSets
kubectl wait --for=condition=ready pod -l app=cruder

# Old ReplicaSet pods:
cruder-app-old-abc123 (Terminating) ❌
cruder-app-old-def456 (Terminating) ❌

# New ReplicaSet pods:
cruder-app-new-xyz789 (Running) ✅
cruder-app-new-qwe012 (Running) ✅
```

When `kubectl wait` tries to check terminating pods, they may already be gone → "pod not found" error.

### Why Deployment-Level Checks Work

```bash
# Deployment status only counts current ReplicaSet
kubectl get deployment cruder-app -o jsonpath='{.status.readyReplicas}'
# Returns: 2 (only new ReplicaSet pods)

kubectl get deployment cruder-app -o jsonpath='{.spec.replicas}'
# Returns: 2 (desired count)
```

Deployment controller automatically filters out terminating pods from old ReplicaSets.

---

**Last Updated**: 2025-11-13  
**Author**: Fahad Aziz

