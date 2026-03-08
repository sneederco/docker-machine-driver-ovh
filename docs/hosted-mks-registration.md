# Hosted OVH MKS: kubeconfig + Rancher registration notes (MVP)

This document captures the operational flow for WS2 hosted-provider MVP.

## 1) Create cluster + nodepool via driver

Use hosted mode flags during `docker-machine create`:

```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-project "$OVH_PROJECT" \
  --ovh-region GRA1 \
  --ovh-mks-cluster-name dm-mks-cluster \
  --ovh-mks-nodepool-name default \
  --ovh-mks-nodepool-flavor b2-7 \
  --ovh-mks-nodepool-size 3 \
  dm-ovh-mks
```

MVP behavior:
- creates one MKS cluster
- creates one nodepool
- stores `MKSClusterID` and `MKSNodePoolID` in machine state

## 2) Inspect/list hosted clusters

Programmatic helper path in the driver:
- `ListHostedMKSClusters()` returns clusters in current project
- `GetState()` in hosted mode resolves cluster status from `/kube/{clusterId}`

## 3) Scale nodepool

Programmatic helper path in the driver:
- `ScaleHostedMKSNodePool(desiredNodes)`
- sends `PUT /cloud/project/{projectId}/kube/{clusterId}/nodepool/{nodePoolId}`

## 4) Delete hosted cluster

`docker-machine rm <name>` in hosted mode:
- deletes MKS cluster by `MKSClusterID`
- OVH cascades nodepool removal with cluster deletion

## 5) Kubeconfig retrieval (operator step)

The driver MVP does not yet fetch kubeconfig directly.
Retrieve kubeconfig from OVH API/CLI after cluster is READY, then import into Rancher hosted provider flow.

Example with OVH CLI/API (pseudo):

```bash
# obtain kubeconfig for cluster id
ovhcloud kubeconfig get --project "$OVH_PROJECT" --cluster-id "$MKS_CLUSTER_ID" > kubeconfig.yaml
```

Then use Rancher import path:
1. Cluster Management → Import Existing
2. Upload/paste kubeconfig
3. Validate nodepool nodes and CNI status

## Known MVP gaps

- no built-in kubeconfig download in driver yet
- Start/Stop/Restart are VM-only operations (not supported in hosted mode)
- no nodepool autoscaling policy management in driver
