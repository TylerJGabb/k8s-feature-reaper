#! /bin/bash

set -e

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl not found. Please install kubectl and ensure it's in your PATH."
  exit 1
fi

if ! command -v kind >/dev/null 2>&1; then
  echo "kind not found. Please install kind and ensure it's in your PATH."
  exit 1
fi

echo "Creating a kind cluster for chart integration tests..."
make kind-create
trap 'make kind-delete' EXIT
# Wait for the cluster to be ready
echo "Waiting for the cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=30s
echo "Cluster is ready."

# Create test namespaces: ns-old (>72h old) and ns-new (recent)
OLD_TS=$(date -u -v-73H +%Y%m%d%H%M%S)
NEW_TS=$(date -u -v-1H +%Y%m%d%H%M%S)

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: ns-old
  labels:
    isFeature: "true"
  annotations:
    updatedAt: "$OLD_TS"
---
apiVersion: v1
kind: Namespace
metadata:
  name: ns-new
  labels:
    isFeature: "true"
  annotations:
    updatedAt: "$NEW_TS"
EOF

echo "Created namespaces ns-old (old) and ns-new (recent) for testing."

make load
make upgrade-install

echo "Checking for feature-reaper CronJob in the feature-reaper namespace..."
if kubectl get cronjob feature-reaper -n feature-reaper >/dev/null 2>&1; then
  echo "feature-reaper CronJob exists in the feature-reaper namespace."
else
  echo "feature-reaper CronJob not found in the feature-reaper namespace."
  exit 1
fi

# Trigger the feature-reaper CronJob manually
JOB_NAME=feature-reaper-manual-$(date +%s)
echo "Triggering the feature-reaper CronJob manually..."
kubectl create job --from=cronjob/feature-reaper $JOB_NAME -n feature-reaper

# Wait for the job to complete
echo "Waiting for the feature-reaper job to complete..."
kubectl wait --for=condition=complete --timeout=5s job $JOB_NAME -n feature-reaper

# Check namespaces with Active status
echo "Checking that only 'ns-new' namespace is Active..."
ACTIVE_NAMESPACES=$(kubectl get ns --field-selector=status.phase=Active -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | grep '^ns-')

if [ "$ACTIVE_NAMESPACES" = "ns-new" ]; then
  echo "Test passed: Only 'ns-new' namespace is Active."
else
  echo "Test failed: Active namespaces are: $ACTIVE_NAMESPACES"
  exit 1
fi
