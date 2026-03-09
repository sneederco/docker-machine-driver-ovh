# Cluster Autoscaler Integration

The OVH Node Driver supports automatic cluster scaling through the Kubernetes Cluster Autoscaler with Rancher integration.

## How It Works

1. **Enable in UI**: Check "Enable Cluster Autoscaler" when creating a cluster
2. **Set Limits**: Configure minimum and maximum nodes per pool
3. **Auto-Deploy**: The driver deploys the cluster-autoscaler to your cluster
4. **Scale Up**: When pods are pending due to insufficient resources, new nodes are added
5. **Scale Down**: When nodes are underutilized for 5+ minutes, they're removed (respecting min)

## Requirements

- Rancher 2.8+ with RKE2 provisioning
- OVH Node Driver v1.0.2+
- Machine pool annotations for min/max

## Configuration

### Via UI (Recommended)

1. Create/edit cluster with OVHcloud driver
2. In "Autoscaling" section:
   - Enable "Enable Cluster Autoscaler"
   - Set "Minimum Nodes" (default: 1)
   - Set "Maximum Nodes" (default: 10)

### Via kubectl

```bash
# Add autoscaling annotations to machine pool
kubectl patch cluster.provisioning.cattle.io <cluster-name> -n fleet-default --type=json -p='[
  {
    "op": "add",
    "path": "/spec/rkeConfig/machinePools/0/machineDeploymentAnnotations",
    "value": {
      "cluster.provisioning.cattle.io/autoscaler-min-size": "1",
      "cluster.provisioning.cattle.io/autoscaler-max-size": "10"
    }
  }
]'
```

## Manual Deployment

If autoscaler isn't auto-deployed, use the manifest in `manifests/cluster-autoscaler.yaml`:

1. Set environment variables:
   ```bash
   export RANCHER_URL="https://rancher.example.com"
   export RANCHER_TOKEN="token-xxxxx:yyyyy"
   export CLUSTER_NAME="my-cluster"
   ```

2. Apply the manifest:
   ```bash
   envsubst < manifests/cluster-autoscaler.yaml | kubectl apply -f -
   ```

## Monitoring

```bash
# Check autoscaler logs
kubectl logs -n cluster-autoscaler deployment/cluster-autoscaler

# Check scaling events
kubectl get events -n cluster-autoscaler

# Check current node count
kubectl get nodes
```

## Cost Control Tips

1. **Use hourly billing** for autoscaled pools (monthly can't scale down)
2. **Set reasonable max** to avoid runaway costs
3. **Monitor scale events** in Rancher UI or kubectl
4. **Consider pod priority** to control which workloads trigger scaling
