#!/bin/bash
# deploy-autoscaler.sh - Deploy cluster autoscaler to a Rancher-managed cluster
# Usage: ./deploy-autoscaler.sh <cluster-name> [min-nodes] [max-nodes]
#
# Environment variables:
#   RANCHER_URL   - Rancher server URL (auto-detected if not set)
#   RANCHER_TOKEN - Rancher API token (required)
#
# Example:
#   export RANCHER_TOKEN='token-xxxxx:yyyyyyyyyyyyyyy'
#   ./deploy-autoscaler.sh my-cluster 1 10

set -e

CLUSTER_NAME="${1:?Usage: $0 <cluster-name> [min] [max]}"
MIN_NODES="${2:-1}"
MAX_NODES="${3:-10}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Validate token
if [ -z "$RANCHER_TOKEN" ]; then
    echo "Error: RANCHER_TOKEN not set"
    echo "Create a Rancher API key and export it:"
    echo "  export RANCHER_TOKEN='token-xxxxx:yyyyyyyyyyyyyyy'"
    exit 1
fi

# Get Rancher URL from cluster settings if not provided
RANCHER_URL="${RANCHER_URL:-$(kubectl get settings.management.cattle.io server-url -o jsonpath='{.value}' 2>/dev/null)}"
if [ -z "$RANCHER_URL" ]; then
    echo "Error: Could not determine RANCHER_URL"
    exit 1
fi

echo "=== Deploying Cluster Autoscaler ==="
echo "  Cluster:    $CLUSTER_NAME"
echo "  Rancher:    $RANCHER_URL"
echo "  Min nodes:  $MIN_NODES"
echo "  Max nodes:  $MAX_NODES"

# Get cluster ID
CLUSTER_ID=$(kubectl get cluster.provisioning.cattle.io "$CLUSTER_NAME" -n fleet-default -o jsonpath='{.status.clusterName}' 2>/dev/null)
if [ -z "$CLUSTER_ID" ]; then
    echo "Error: Cluster '$CLUSTER_NAME' not found in fleet-default namespace"
    exit 1
fi
echo "  Cluster ID: $CLUSTER_ID"

# Generate kubeconfig for downstream cluster
echo ""
echo "=== Fetching downstream cluster kubeconfig ==="
TMPKUBE=$(mktemp)
curl -sk "$RANCHER_URL/v3/clusters/$CLUSTER_ID?action=generateKubeconfig" \
    -H "Authorization: Bearer $RANCHER_TOKEN" \
    -X POST | jq -r '.config' > "$TMPKUBE"

if [ ! -s "$TMPKUBE" ]; then
    echo "Error: Failed to generate kubeconfig"
    rm -f "$TMPKUBE"
    exit 1
fi

# Add autoscaler annotations to machine pool
echo ""
echo "=== Setting autoscaler annotations ==="
kubectl patch cluster.provisioning.cattle.io "$CLUSTER_NAME" -n fleet-default --type=json -p="[
  {\"op\":\"add\",\"path\":\"/spec/rkeConfig/machinePools/0/machineDeploymentAnnotations\",\"value\":{}},
  {\"op\":\"add\",\"path\":\"/spec/rkeConfig/machinePools/0/machineDeploymentAnnotations/cluster.provisioning.cattle.io~1autoscaler-min-size\",\"value\":\"$MIN_NODES\"},
  {\"op\":\"add\",\"path\":\"/spec/rkeConfig/machinePools/0/machineDeploymentAnnotations/cluster.provisioning.cattle.io~1autoscaler-max-size\",\"value\":\"$MAX_NODES\"}
]" 2>/dev/null || echo "  Annotations may already exist, continuing..."

# Deploy autoscaler using the manifest
echo ""
echo "=== Deploying autoscaler to downstream cluster ==="

# Use envsubst to inject variables into manifest
export RANCHER_URL RANCHER_TOKEN CLUSTER_NAME
envsubst < "$SCRIPT_DIR/../manifests/cluster-autoscaler.yaml" | \
    KUBECONFIG="$TMPKUBE" kubectl apply --server-side -f -

rm -f "$TMPKUBE"

echo ""
echo "✅ Cluster autoscaler deployed successfully!"
echo ""
echo "Monitor with:"
echo "  kubectl -n cluster-autoscaler logs -f deployment/cluster-autoscaler"
