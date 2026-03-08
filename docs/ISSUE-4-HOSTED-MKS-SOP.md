# Issue #4 — Hosted MKS Create/Delete/Scale + Hourly Billing Teardown SOP

Issue: https://github.com/sneederco/docker-machine-driver-ovh/issues/4

This SOP is for the **hosted MKS path** (OVH Managed Kubernetes Service) used by Rancher operators.

---

## 1) Prereqs

```bash
export OVH_ENDPOINT="ovh-eu"
export OVH_APPLICATION_KEY="<app-key>"
export OVH_APPLICATION_SECRET="<app-secret>"
export OVH_CONSUMER_KEY="<consumer-key>"

export OVH_PROJECT_ID="<project-id>"
export MKS_REGION="GRA9"
export MKS_NAME="rkf-mks-dev"
export K8S_VERSION="1.30"
```

Optional local helper alias (if `ovh` CLI is installed):

```bash
alias ovhapi='ovh --endpoint "$OVH_ENDPOINT"'
```

---

## 2) Create hosted MKS cluster

### API path (curl)

```bash
curl -sS -X POST "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube" \
  -H "X-Ovh-Application: ${OVH_APPLICATION_KEY}" \
  -H "X-Ovh-Consumer: ${OVH_CONSUMER_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${MKS_NAME}\",\"region\":\"${MKS_REGION}\",\"version\":\"${K8S_VERSION}\"}"
```

### CLI path (if available)

```bash
ovhapi api POST /cloud/project/${OVH_PROJECT_ID}/kube \
  name="${MKS_NAME}" region="${MKS_REGION}" version="${K8S_VERSION}"
```

Capture the resulting cluster id:

```bash
export MKS_ID="<kube-id-from-create-response>"
```

---

## 3) Create hourly-billed node pool

> Use `billingPeriod=hourly` explicitly to avoid accidental monthly billing.

```bash
export NODEPOOL_NAME="pool-a"
export FLAVOR="b2-7"
export DESIRED_NODES="3"

curl -sS -X POST "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool" \
  -H "X-Ovh-Application: ${OVH_APPLICATION_KEY}" \
  -H "X-Ovh-Consumer: ${OVH_CONSUMER_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${NODEPOOL_NAME}\",\"flavorName\":\"${FLAVOR}\",\"desiredNodes\":${DESIRED_NODES},\"billingPeriod\":\"hourly\"}"
```

Verify:

```bash
curl -sS "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool" | jq .
```

---

## 4) Scale node pool up/down

Get pool id:

```bash
export NODEPOOL_ID="$(curl -sS https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool | jq -r '.[0].id')"
```

Scale to 5:

```bash
curl -sS -X PUT "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool/${NODEPOOL_ID}" \
  -H "X-Ovh-Application: ${OVH_APPLICATION_KEY}" \
  -H "X-Ovh-Consumer: ${OVH_CONSUMER_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"desiredNodes":5}'
```

Scale to 0 (teardown prep):

```bash
curl -sS -X PUT "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool/${NODEPOOL_ID}" \
  -H "X-Ovh-Application: ${OVH_APPLICATION_KEY}" \
  -H "X-Ovh-Consumer: ${OVH_CONSUMER_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"desiredNodes":0}'
```

---

## 5) Delete flow (nodepool first, then cluster)

Delete pool:

```bash
curl -sS -X DELETE "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool/${NODEPOOL_ID}" \
  -H "X-Ovh-Application: ${OVH_APPLICATION_KEY}" \
  -H "X-Ovh-Consumer: ${OVH_CONSUMER_KEY}"
```

Delete cluster:

```bash
curl -sS -X DELETE "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}" \
  -H "X-Ovh-Application: ${OVH_APPLICATION_KEY}" \
  -H "X-Ovh-Consumer: ${OVH_CONSUMER_KEY}"
```

Confirm zero hosted MKS resources remain:

```bash
curl -sS "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube" | jq .
```

---

## 6) Hourly billing teardown checklist

Run this checklist every time issue #4 hosted test environments are retired.

- [ ] Scale all temporary node pools to `desiredNodes=0`.
- [ ] Delete all temporary node pools created for the test run.
- [ ] Delete the temporary MKS cluster(s) after nodepool deletion completes.
- [ ] Verify no attached public IP / LB / volume artifacts remain from workload tests.
- [ ] Verify OVH billing dashboard no longer shows active hourly compute for this run.
- [ ] Remove temporary kubeconfigs and API tokens from CI/local shell history.
- [ ] Post teardown evidence (command transcript + timestamp) to issue thread.

---

## 7) Rollback / cleanup if delete fails

1. Re-list current resources:
   ```bash
   curl -sS "https://api.ovh.com/1.0/cloud/project/${OVH_PROJECT_ID}/kube/${MKS_ID}/nodepool" | jq .
   ```
2. Retry nodepool delete until status is terminal (`DELETED`/absent).
3. If cluster delete fails with dependency errors, check for:
   - pending node pools
   - managed load balancers from in-cluster services
   - volumes or snapshots retained by CSI
4. Force cleanup of leftover cloud resources, then rerun cluster delete.
5. Escalate to OVH support with project id + cluster id + operation id if API returns a persistent 5xx/409.

Owner for escalation: Platform/Ops on-call (OVH account owner).

---

## 8) Evidence bundle template (paste into issue #4)

```text
Issue: #4
Project: <OVH_PROJECT_ID>
Cluster: <MKS_ID>
Nodepool: <NODEPOOL_ID>
Create time: <UTC>
Scale events: <UTC list>
Delete time: <UTC>
Billing verification: <screenshot/link>
Cleanup checklist: complete/incomplete (+why)
```
