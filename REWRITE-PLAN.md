# Sneederco OVH Driver Rewrite Plan

**Goal:** Full modernization of docker-machine-driver-ovh with sneederco branding, modern Go, and proper Rancher 2.x integration.

## Repos
- **Backend:** `sneederco/docker-machine-driver-ovh` (Go driver)
- **Frontend:** `sneederco/ui-driver-ovh` (Rancher UI extension)

---

## Phase 1: Core Driver Rewrite (Go Backend)

### Agent 1: Project Structure & Dependencies
**Branch:** `rewrite/modernize-core`
**Tasks:**
- [ ] Initialize Go modules (`go mod init github.com/sneederco/docker-machine-driver-ovh`)
- [ ] Update to Go 1.22+
- [ ] Replace vendored deps with modern versions:
  - `github.com/ovh/go-ovh` v1.6+ (has `ovh-us`)
  - `github.com/rancher/machine` (latest compatible)
  - `github.com/sirupsen/logrus` (logging)
- [ ] Add `.golangci.yml` for linting
- [ ] Update Makefile for modern Go build
- [ ] Remove old vendor/, .gopath/, .travis.yml

### Agent 2: OVH API Client Rewrite
**Branch:** `rewrite/api-client`
**Tasks:**
- [ ] Rewrite `api.go` with modern patterns:
  - Context support for cancellation
  - Structured errors
  - Rate limiting awareness
  - Retry logic with backoff
- [ ] Add new API methods:
  - `ListRegions()` - dynamic region discovery
  - `ListFlavors(region)` - dynamic flavor listing
  - `ListImages(region)` - dynamic image listing
  - `ListSSHKeys()` - user's SSH keys
  - `GetQuotas(region)` - quota checking
- [ ] Add proper logging throughout
- [ ] Unit tests with mocked OVH API

### Agent 3: Driver Implementation Rewrite
**Branch:** `rewrite/driver-core`
**Tasks:**
- [ ] Rewrite `driver.go` with:
  - All Rancher machine driver interface methods
  - Proper state machine for VM lifecycle
  - Cloud-init support for custom scripts
  - Private network support
  - SSH key management (create if not exists)
- [ ] Update default values:
  - Default region: `US-EAST-VA-1` (not `GRA1`)
  - Default image: `Ubuntu 24.04` (not `Ubuntu 16.04`)
  - Default flavor: `b3-8`
- [ ] Add driver flags:
  - `--ovh-ssh-key-name` (use existing or create)
  - `--ovh-private-network` (VPC/vRack)
  - `--ovh-userdata` (cloud-init script)
  - `--ovh-tags` (metadata tags)
- [ ] Implement proper cleanup on failure

### Agent 4: Hosted MKS Support (Optional)
**Branch:** `rewrite/mks-integration`
**Tasks:**
- [ ] Add OVH Managed Kubernetes (MKS) API client
- [ ] Implement MKS node pool driver mode
- [ ] Add MKS-specific flags:
  - `--ovh-mks-cluster-id`
  - `--ovh-mks-nodepool-name`
  - `--ovh-mks-autoscale-min/max`

---

## Phase 2: UI Driver Rewrite (Rancher Extension)

### Agent 5: UI Modernization
**Branch:** `rewrite/ui-modern`
**Repo:** `sneederco/ui-driver-ovh`
**Tasks:**
- [ ] Update to Rancher UI driver SDK 2.x
- [ ] Add sneederco branding (logo, colors)
- [ ] Implement dynamic dropdowns (API-backed):
  - Region selector (fetches from OVH API)
  - Flavor selector (filtered by region)
  - Image selector (filtered by region)
  - SSH key selector
- [ ] Add field validation
- [ ] Add cost estimation display (hourly rates)
- [ ] Responsive design improvements
- [ ] Add tooltips/help text for all fields

---

## Phase 3: CI/CD & Release

### Agent 6: CI/CD Pipeline
**Branch:** `rewrite/ci-cd`
**Tasks:**
- [ ] GitHub Actions workflow:
  - Build for linux-amd64, linux-arm64, darwin-amd64, darwin-arm64
  - Run tests and linting
  - Create GitHub releases with binaries
  - Push to GHCR (container image optional)
- [ ] Dependabot configuration
- [ ] Security scanning (CodeQL, gosec)
- [ ] Automated changelog generation

---

## Phase 4: Documentation & Testing

### Agent 7: Documentation
**Branch:** `rewrite/docs`
**Tasks:**
- [ ] Rewrite README.md:
  - Sneederco branding
  - Quick start guide
  - All configuration options
  - Troubleshooting section
- [ ] Add CONTRIBUTING.md
- [ ] Add docs/:
  - `installation.md`
  - `configuration.md`
  - `rancher-integration.md`
  - `troubleshooting.md`
- [ ] Architecture diagram

### Agent 8: Integration Testing
**Branch:** `rewrite/testing`
**Tasks:**
- [ ] Add integration test suite:
  - Create VM test
  - Delete VM test
  - State transitions
  - Error handling
- [ ] Add mock OVH API for CI testing
- [ ] Add Rancher e2e test (optional)

---

## Execution Order

```
┌─────────────────────────────────────────────────────────┐
│  PARALLEL PHASE 1                                        │
├─────────────────────────────────────────────────────────┤
│  Agent 1: Project Structure    Agent 2: API Client      │
│  Agent 5: UI Modernization     Agent 6: CI/CD           │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│  SEQUENTIAL PHASE 2 (depends on Agent 1 & 2)            │
├─────────────────────────────────────────────────────────┤
│  Agent 3: Driver Implementation                          │
│  Agent 4: MKS Support (optional)                         │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│  PARALLEL PHASE 3                                        │
├─────────────────────────────────────────────────────────┤
│  Agent 7: Documentation        Agent 8: Testing         │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
                    [MERGE & RELEASE]
```

---

## Branding

- **Package name:** `github.com/sneederco/docker-machine-driver-ovh`
- **Binary name:** `docker-machine-driver-ovh` (standard)
- **Display name in Rancher:** "OVHcloud (Sneederco)"
- **Logo:** Sneederco logo or OVH logo with sneederco badge
- **Description:** "OVHcloud driver for Rancher, maintained by Sneederco"

---

## Success Criteria

1. ✅ Driver builds with Go 1.22+
2. ✅ `ovh-us` endpoint works out of the box
3. ✅ Dynamic region/flavor/image selection in UI
4. ✅ VM provisioning works in US-EAST-VA-1
5. ✅ Rancher cluster creation succeeds end-to-end
6. ✅ Autoscaling works (scale up/down)
7. ✅ All tests pass in CI
8. ✅ GitHub releases automated

---

## Timeline Estimate

- **Phase 1:** 2-3 hours (parallel agents)
- **Phase 2:** 1-2 hours (sequential)
- **Phase 3:** 1 hour (parallel)
- **Total:** ~4-6 hours with sub-agents

---

## Notes

- Keep backward compatibility with existing Rancher node driver registration
- Credential schema (`ovhcredentialconfig`) must match existing format
- Test with both OVH EU and OVH US endpoints
