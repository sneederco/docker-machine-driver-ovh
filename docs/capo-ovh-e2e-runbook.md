# CAPO on OVH (OpenStack) End-to-End Runbook

## What was required to make workers join + Rancher import succeed

1. `cluster-info` in `kube-public` must point to the actual join endpoint (internal API IP when joining privately).
2. kubeadm discovery RBAC must allow anonymous/bootstrap reads for `cluster-info`.
3. Bootstrap token RBAC must allow reading `kubeadm-config` in `kube-system`.
4. `kubelet-config` and `kube-proxy` ConfigMaps must exist.
5. Nodes needed explicit InternalIPs patched (OVH/CAPO test case) so CNI + kubelet proxy paths work.
6. `kube-proxy` needed direct API endpoint kubeconfig to avoid ClusterIP bootstrap deadlock.
7. Rancher import agent required a restart after networking stabilized.

## Key manifests/components used
- OpenStack Cloud Controller Manager (OCCM) in `kube-system`
- Flannel CNI daemonset
- kube-proxy daemonset
- Rancher `cattle-cluster-agent` import manifest

## Known failure signatures and fixes
- `failed to request cluster-info ConfigMap` => discovery RBAC and endpoint mismatch.
- `kubeadm-config forbidden` => bootstrap Role/RoleBinding missing in `kube-system`.
- `kubelet-config not found` => create ConfigMap.
- `dial tcp 10.96.0.1:443: i/o timeout` from rancher-agent/kube-proxy => kube-proxy not functional yet; switch kube-proxy to direct API server, then restart.

## Final validated state
- Rancher imported cluster became `active` with nodeCount > 0.
- Control plane and workers all `Ready`.
