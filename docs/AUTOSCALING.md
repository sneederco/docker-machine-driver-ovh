# OVH Cluster Autoscaler Integration

This guide explains how to enable automatic scaling for OVH clusters managed by Rancher.

## How It Works

1. **User creates cluster** with autoscaler annotations (via YAML mode)
2. **Autoscaler controller** (running on Rancher management cluster) detects the annotations
3. **Controller auto-deploys** cluster-autoscaler to the downstream cluster
4. **Cluster-autoscaler** monitors pods and scales machine pools up/down

## Prerequisites

- Rancher 2.7+ with OVH node driver installed
- Autoscaler controller deployed to management cluster (see [Controller Setup](#controller-setup))

## Enabling Autoscaling (YAML Method)

When creating or editing a cluster in Rancher:

1. Click **Edit as YAML** in the cluster creation wizard
2. Add `machineDeploymentAnnotations` to your machine pool:

```yaml
apiVersion: provisioning.cattle.io/v1
kind: Cluster
metadata:
  name: my-cluster
  namespace: fleet-default
spec:
  cloudCredentialSecretName: cattle-global-data:cc-xxxxx
  kubernetesVersion: v1.31.6+rke2r1
  rkeConfig:
    machinePools:
      - name: pool1
        quantity: 1
        etcdRole: true
        controlPlaneRole: true
        workerRole: true
        # ADD THESE ANNOTATIONS:
        machineDeploymentAnnotations:
          cluster.provisioning.cattle.io/autoscaler-min-size: "1"
          cluster.provisioning.cattle.io/autoscaler-max-size: "10"
        machineConfigRef:
          kind: OvhConfig
          name: my-cluster-pool1
```

3. Save and create the cluster
4. Once the cluster is `Ready`, the autoscaler controller will automatically deploy cluster-autoscaler

## Annotation Reference

| Annotation | Description | Example |
|------------|-------------|---------|
| `cluster.provisioning.cattle.io/autoscaler-min-size` | Minimum nodes in pool | `"1"` |
| `cluster.provisioning.cattle.io/autoscaler-max-size` | Maximum nodes in pool | `"10"` |

## Verifying Autoscaler Deployment

After cluster becomes Ready:

```bash
# Check if autoscaler was deployed (on management cluster)
kubectl get clusters.provisioning.cattle.io <cluster-name> -n fleet-default \
  -o jsonpath='{.metadata.labels.autoscaler\.sneederco\.io/deployed}'
# Should return: true

# Check autoscaler is running (on downstream cluster)
kubectl --kubeconfig=<downstream-kubeconfig> get pods -n cluster-autoscaler
# Should show: cluster-autoscaler-xxxxx  1/1  Running
```

## Monitoring Autoscaler

```bash
# View autoscaler logs (on downstream cluster)
kubectl --kubeconfig=<downstream-kubeconfig> logs -n cluster-autoscaler -l app=cluster-autoscaler -f

# Check scale events
kubectl --kubeconfig=<downstream-kubeconfig> get events -n cluster-autoscaler --sort-by='.lastTimestamp'
```

## Scale-Up Behavior

The autoscaler will add nodes when:
- Pods are `Pending` due to insufficient resources
- Node affinity/taints prevent scheduling on existing nodes

Scale-up respects:
- `autoscaler-max-size` annotation limit
- OVH API rate limits and quotas

## Scale-Down Behavior

The autoscaler will remove nodes when:
- Node utilization is below 50% for 5+ minutes
- All pods can be rescheduled to other nodes
- No system-critical pods would be evicted

Protected from scale-down:
- Nodes with pods that have `cluster-autoscaler.kubernetes.io/safe-to-evict: "false"`
- Nodes running DaemonSets (unless all pods are DaemonSets)

## Controller Setup

The autoscaler controller must be deployed to your Rancher management cluster:

```bash
# Set your Rancher token
export RANCHER_TOKEN='token-xxxxx:xxxxxxxxxx'

# Apply controller manifests
kubectl apply -f deploy/autoscaler-controller.yaml
```

The controller:
- Runs in `autoscaler-controller` namespace
- Watches clusters every 30 seconds
- Labels deployed clusters with `autoscaler.sneederco.io/deployed=true`

## Troubleshooting

### Autoscaler not deploying

Check controller logs:
```bash
kubectl logs -n autoscaler-controller -l app=autoscaler-controller
```

Common issues:
- Cluster not `Ready` yet (controller waits for ready state)
- Missing annotations on machinePool
- Already labeled as deployed

### Scale-up not happening

Check autoscaler logs for:
- "Pod is unschedulable" messages
- "Scale-up: group is at max size" (hit your max limit)
- Quota errors from OVH API

### Scale-down not happening

Reasons scale-down might be blocked:
- `--scale-down-enabled=false` in autoscaler config
- Pods with local storage
- Pods with PDBs that would be violated
- Recent scale-up (5 minute cooldown by default)

## Configuration

Default autoscaler settings deployed by the controller:

| Setting | Value | Description |
|---------|-------|-------------|
| `--scale-down-enabled` | `true` | Allow removing underutilized nodes |
| `--scale-down-delay-after-add` | `5m` | Wait after scale-up before scale-down |
| `--scale-down-unneeded-time` | `5m` | How long node must be unneeded |
| `--v` | `4` | Log verbosity |

To customize, edit the deployed ConfigMap in the downstream cluster:
```bash
kubectl --kubeconfig=<downstream> edit configmap cluster-autoscaler-cloud-config -n cluster-autoscaler
```
