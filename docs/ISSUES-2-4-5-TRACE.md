# Issues #2 / #4 / #5 — Traceability & execution notes

Epic: https://github.com/sneederco/docker-machine-driver-ovh/issues/2
- WS2: https://github.com/sneederco/docker-machine-driver-ovh/issues/4
- WS3: https://github.com/sneederco/docker-machine-driver-ovh/issues/5

## Scope split
- Issue #2 (epic): dual provisioning documentation, acceptance evidence, and runbooks
- Issue #4 (WS2): hosted path MVP (OVH MKS create/delete/scale)
- Issue #5 (WS3): node path hardening for OVH instances used by RKE2/K3s

## Existing doc anchor
- `README.md` remains legacy upstream usage + flags reference
- UI/adoption flow is documented in sibling repo: `sneederco/ui-driver-ovh/docs/README.md`

## Command consistency proof (current host)
Attempted in repo root:

```bash
go build -o /tmp/docker-machine-driver-ovh-proof .
```

Result: failed due to missing Go module context.

Attempted legacy mode:

```bash
GO111MODULE=off go build -o /tmp/docker-machine-driver-ovh-proof .
```

Result: failed due to missing `github.com/docker/machine/libmachine/*` and `github.com/ovh/go-ovh/ovh` dependencies in current GOPATH context.

## Blocker
- Owner: Platform/Build environment owner (repo maintainers)
- Needed: deterministic GOPATH+vendor build recipe (or go.mod migration) runnable outside CI
- Impact: cannot produce local WS2/WS3 runtime proof artifact on this host yet

## Issue #4 docs delivered
- Added `docs/ISSUE-4-HOSTED-MKS-SOP.md` covering:
  - Hosted MKS create flow
  - Nodepool hourly billing create + scale up/down
  - Delete order (nodepool -> cluster)
  - Hourly billing teardown checklist
  - Rollback/cleanup escalation path
  - Evidence bundle template for issue comments

## Next doc actions
1. Add CI-based proof references for amd64/arm64 deterministic artifacts once workflow outputs are confirmed.
2. Add Rancher registration dry-run transcript references aligned to WS2/WS3 acceptance criteria.
3. Attach real command transcript IDs from a successful hosted test run to close issue #4 acceptance evidence.
