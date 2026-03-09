# Architecture Overview

Technical overview of how the OVHcloud Driver for Rancher works internally.

## Table of Contents

- [Overview](#overview)
- [Component Architecture](#component-architecture)
- [Driver Lifecycle](#driver-lifecycle)
- [API Integration](#api-integration)
- [SSH Key Management](#ssh-key-management)
- [Network Architecture](#network-architecture)
- [State Management](#state-management)
- [Rancher Integration](#rancher-integration)
- [MKS Mode Architecture](#mks-mode-architecture)

## Overview

The OVHcloud Driver is a Docker Machine driver that provisions and manages OVHcloud Public Cloud instances. It implements the Docker Machine driver interface and communicates with the OVH API to create, configure, and destroy compute resources.

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Machine / Rancher                  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │           Docker Machine Driver Interface            │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                │
│                            ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              OVHcloud Driver (this)                  │  │
│  │                                                       │  │
│  │  • Create/Delete instances                           │  │
│  │  • SSH key management                                │  │
│  │  • State management                                  │  │
│  │  • Network configuration                             │  │
│  └──────────────────────────────────────────────────────┘  │
│                            │                                │
│                            ▼                                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  OVH API Client                      │  │
│  │                                                       │  │
│  │  • Authentication                                     │  │
│  │  • API request signing                               │  │
│  │  • Error handling                                    │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                      OVH Public API                          │
│                                                              │
│  • Instance lifecycle (create/delete/status)                │
│  • Project management                                       │
│  • SSH key management                                       │
│  • Network management (vRack)                               │
│  • Managed Kubernetes (MKS)                                 │
└─────────────────────────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                  OVHcloud Infrastructure                     │
│                                                              │
│  • Compute instances (OpenStack-based)                      │
│  • Block storage                                            │
│  • Private networks (vRack)                                 │
│  • Managed Kubernetes                                       │
└─────────────────────────────────────────────────────────────┘
```

## Component Architecture

### Core Components

#### 1. Driver (`driver.go`)

The main driver struct implements the Docker Machine driver interface:

```go
type Driver struct {
    *drivers.BaseDriver
    
    // Configuration
    ProjectName        string
    FlavorName         string
    RegionName         string
    PrivateNetworkName string
    
    // OVH credentials
    ApplicationKey    string
    ApplicationSecret string
    ConsumerKey       string
    Endpoint          string
    
    // Internal state
    ProjectID   string
    FlavorID    string
    ImageID     string
    InstanceID  string
    KeyPairName string
    
    // API client
    client *API
}
```

**Key methods:**
- `Create()` — Provision new instance
- `Remove()` — Delete instance
- `Start()` — Start stopped instance
- `Stop()` — Stop running instance
- `Kill()` — Force stop instance
- `GetState()` — Query instance state
- `GetIP()` — Get instance IP address

#### 2. API Client (`api.go`)

Handles all communication with OVH API:

```go
type API struct {
    endpoint          string
    applicationKey    string
    applicationSecret string
    consumerKey       string
    client            *http.Client
}
```

**Key methods:**
- `Get()` — Execute GET request
- `Post()` — Execute POST request
- `Put()` — Execute PUT request
- `Delete()` — Execute DELETE request
- `signRequest()` — Sign API requests

**Authentication flow:**
1. Calculate timestamp
2. Generate signature: SHA1(AS+CK+METHOD+URL+BODY+TS)
3. Add headers:
   - `X-Ovh-Application`: Application Key
   - `X-Ovh-Timestamp`: Current timestamp
   - `X-Ovh-Consumer`: Consumer Key
   - `X-Ovh-Signature`: Computed signature

#### 3. Configuration

Configuration priority (highest to lowest):
1. **Command-line flags** — `--ovh-*`
2. **Environment variables** — `OVH_*`
3. **Configuration file** — `~/.ovh.conf`

**Configuration file format:**
```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=YOUR_KEY
application_secret=YOUR_SECRET
consumer_key=YOUR_CONSUMER
```

## Driver Lifecycle

### Instance Creation Flow

```
1. SetConfigFromFlags()
   ├─ Parse command-line flags
   ├─ Load credentials (flags > env > config file)
   └─ Validate required parameters

2. PreCreateCheck()
   ├─ Validate credentials
   ├─ Test API connectivity
   ├─ Resolve project ID (if name provided)
   └─ Verify region availability

3. Create()
   ├─ Resolve image ID (name → ID)
   ├─ Resolve flavor ID (name → ID)
   ├─ Create/upload SSH key
   ├─ Create instance via API
   ├─ Wait for instance to be ACTIVE
   ├─ Get public IP address
   └─ Wait for SSH to be available

4. Start()
   └─ Wait for Docker daemon to be ready

5. (Optional) ConfigurePrivateNetwork()
   └─ Attach to vRack VLAN
```

### Instance Deletion Flow

```
1. Remove()
   ├─ Get instance state
   ├─ Delete instance via API
   ├─ Wait for deletion to complete
   ├─ Delete SSH key (if auto-generated)
   └─ Clean up local state
```

### State Query Flow

```
1. GetState()
   ├─ Query instance status via API
   ├─ Map OVH status to Docker Machine state:
   │  • ACTIVE → Running
   │  • SHUTOFF → Stopped
   │  • BUILD → Starting
   │  • ERROR → Error
   │  • DELETED → NotFound
   └─ Return state
```

## API Integration

### OVH API Endpoints Used

#### Instance Management

```
GET    /cloud/project/{project}/instance/{instance}
POST   /cloud/project/{project}/instance
DELETE /cloud/project/{project}/instance/{instance}
POST   /cloud/project/{project}/instance/{instance}/start
POST   /cloud/project/{project}/instance/{instance}/stop
```

#### SSH Keys

```
GET    /cloud/project/{project}/sshkey
POST   /cloud/project/{project}/sshkey
DELETE /cloud/project/{project}/sshkey/{keyId}
```

#### Projects & Metadata

```
GET    /cloud/project
GET    /cloud/project/{project}
GET    /cloud/project/{project}/region
GET    /cloud/project/{project}/flavor
GET    /cloud/project/{project}/image
```

#### Networks (vRack)

```
GET    /cloud/project/{project}/network/private
GET    /cloud/project/{project}/network/private/{networkId}
```

#### Managed Kubernetes

```
POST   /cloud/project/{project}/kube
GET    /cloud/project/{project}/kube/{kubeId}
DELETE /cloud/project/{project}/kube/{kubeId}
POST   /cloud/project/{project}/kube/{kubeId}/nodepool
GET    /cloud/project/{project}/kube/{kubeId}/nodepool/{poolId}
```

### Error Handling

API errors are wrapped with context:

```go
if err != nil {
    return fmt.Errorf("failed to create instance: %w", err)
}
```

**Retry logic:**
- Exponential backoff for transient errors
- Timeout after 200 status checks (configurable)
- Max wait time: ~20 minutes

## SSH Key Management

### Auto-Generated Keys

When no `--ovh-ssh-key` is specified:

```
1. Generate key pair:
   ├─ Create RSA 2048-bit key pair
   ├─ Save private key: ~/.docker/machine/machines/{name}/id_rsa
   └─ Save public key: ~/.docker/machine/machines/{name}/id_rsa.pub

2. Upload to OVH:
   ├─ Generate random key name: docker-machine-{name}-{timestamp}
   ├─ Upload public key via API
   └─ Store key ID in driver state

3. On deletion:
   └─ Delete key from OVH project
```

### Existing Keys

When using `--ovh-ssh-key`:

```
1. Lookup key in OVH project
2. Verify key exists locally:
   ├─ Check ~/.ssh/{key_name}
   └─ Check SSH agent
3. Use key ID for instance creation
4. Skip deletion (key is managed externally)
```

## Network Architecture

### Public Network (Default)

```
Instance
  ├─ eth0 (public)
  │   ├─ Public IPv4
  │   ├─ Public IPv6 (optional)
  │   └─ Internet gateway
  └─ Docker bridge (docker0)
      └─ 172.17.0.0/16
```

### Private Network (vRack)

```
Instance
  ├─ eth0 (public)
  │   ├─ Public IPv4
  │   └─ Internet gateway
  ├─ eth1 (private - requires manual config)
  │   ├─ vRack VLAN
  │   ├─ DHCP from OVH
  │   └─ Private subnet (e.g., 10.0.0.0/24)
  └─ Docker bridge (docker0)
      └─ 172.17.0.0/16
```

**Private interface configuration:**

Not automatically configured by the driver. Must be done post-creation:

```bash
# Method 1: Manual
docker-machine ssh node sudo tee /etc/network/interfaces.d/99-vrack.cfg << 'EOF'
auto ens4
iface ens4 inet dhcp
EOF
docker-machine ssh node sudo ifup ens4

# Method 2: Cloud-init (at creation time)
# Add to cloud-init config
```

### Managed Kubernetes Network

```
MKS Cluster
  ├─ Managed control plane
  │   └─ (OVH-managed, not directly accessible)
  ├─ Nodepool
  │   ├─ Worker node 1
  │   ├─ Worker node 2
  │   └─ Worker node N
  └─ Cluster network
      ├─ Service CIDR (default: 10.96.0.0/12)
      └─ Pod CIDR (default: 10.244.0.0/16)
```

## State Management

### Local State Storage

Docker Machine stores state in `~/.docker/machine/machines/{name}/`:

```
{name}/
  ├─ config.json          # Machine configuration
  ├─ id_rsa               # SSH private key
  ├─ id_rsa.pub           # SSH public key
  ├─ ca.pem               # Certificate authority
  ├─ cert.pem             # Client certificate
  ├─ key.pem              # Client key
  └─ server.pem           # Server certificate
```

### OVH State

Driver stores OVH-specific IDs in config.json:

```json
{
  "Driver": {
    "ProjectID": "abc123...",
    "InstanceID": "def456...",
    "FlavorID": "ghi789...",
    "ImageID": "jkl012...",
    "KeyPairID": "mno345...",
    "NetworkIDs": ["pqr678..."]
  }
}
```

### State Synchronization

The driver reconciles local and remote state:

```
GetState() flow:
  1. Check local cache (config.json)
  2. Query OVH API for current status
  3. Map OVH status to Docker Machine state
  4. Update local cache if changed
  5. Return state
```

## Rancher Integration

### Node Driver Interface

Rancher calls the driver through Docker Machine:

```
Rancher Server
  ├─ Node Driver Manager
  │   └─ Downloads driver binary
  ├─ Node Template
  │   └─ Stores OVH credentials & config
  └─ Node Pool Controller
      └─ Creates/deletes machines via driver

Each node creation:
  1. Rancher calls: docker-machine create -d ovh ...
  2. Driver provisions instance
  3. Rancher waits for Docker to be ready
  4. Rancher installs Rancher agent
  5. Node joins cluster
```

### Authentication in Rancher

Credentials stored per node template:
- Encrypted at rest in Rancher database
- Passed to driver as environment variables
- Not visible in Rancher UI after saving

### Scaling Behavior

```
Scale up:
  For each new node:
    1. Clone node template config
    2. Generate unique hostname
    3. Call driver to create instance
    4. Wait for Docker ready
    5. Install Rancher agent

Scale down:
  For each removed node:
    1. Drain node (move workloads)
    2. Remove from cluster
    3. Call driver to delete instance
    4. Clean up resources
```

## MKS Mode Architecture

### Experimental MVP Implementation

MKS mode switches from single-instance to managed-cluster mode:

```
Standard mode:          MKS mode:
┌─────────────┐        ┌──────────────────────┐
│  Instance   │        │   MKS Cluster        │
│  (VM)       │        │  ┌────────────────┐  │
│             │   vs   │  │ Control Plane  │  │
│  • Docker   │        │  │ (managed)      │  │
│  • SSH      │        │  └────────────────┘  │
└─────────────┘        │  ┌────────────────┐  │
                       │  │   Nodepool      │  │
                       │  │  • Worker 1     │  │
                       │  │  • Worker 2     │  │
                       │  │  • Worker N     │  │
                       │  └────────────────┘  │
                       └──────────────────────┘
```

### MKS Creation Flow

```
1. Create() with --ovh-hosted-mks
   ├─ Create MKS cluster via API
   │  ├─ Set cluster name
   │  ├─ Set Kubernetes version
   │  └─ Set region
   ├─ Wait for cluster to be READY
   ├─ Create nodepool
   │  ├─ Set nodepool name
   │  ├─ Set flavor
   │  └─ Set desired node count
   ├─ Wait for nodepool to be READY
   └─ Store cluster ID and nodepool ID

2. GetState()
   ├─ Query cluster status
   ├─ Map to Docker Machine state
   └─ Return state

3. Remove()
   ├─ Delete nodepool
   ├─ Wait for nodepool deletion
   ├─ Delete cluster
   └─ Wait for cluster deletion
```

### Limitations

Current MKS mode limitations:
- No direct SSH access (managed control plane)
- Limited scaling operations
- Kubeconfig must be retrieved from OVH console
- Cannot use standard Docker Machine commands (ssh, etc.)

## Performance Considerations

### Instance Creation Time

Typical timings:
- API call to create instance: 2-5 seconds
- Instance provisioning: 30-60 seconds
- SSH availability: 1-2 minutes
- Cloud-init completion: 1-2 minutes
- **Total: 3-5 minutes**

### Bottlenecks

1. **OVH API rate limits:**
   - Default: 20 requests/second
   - Bulk operations may hit limits

2. **SSH connection:**
   - Timeout: 10 minutes (configurable)
   - Retries every 2 seconds

3. **Cloud-init:**
   - Package updates can be slow
   - Network-dependent

### Optimization

- Use monthly billing for long-lived instances (no API overhead)
- Reuse SSH keys (reduces API calls)
- Use existing images (faster than custom images)
- Deploy in same region as other resources

## Code Structure

```
docker-machine-driver-ovh/
  ├─ main.go              # Entry point
  ├─ driver.go            # Driver implementation
  ├─ api.go               # OVH API client
  ├─ api_test.go          # API tests
  ├─ go.mod               # Go module definition
  ├─ go.sum               # Go module checksums
  ├─ Makefile             # Build automation
  ├─ .golangci.yml        # Linter configuration
  ├─ README.md            # Documentation
  ├─ CONTRIBUTING.md      # Contribution guide
  └─ docs/                # Additional documentation
      ├─ installation.md
      ├─ configuration.md
      ├─ rancher-integration.md
      ├─ troubleshooting.md
      └─ architecture.md (this file)
```

## Extension Points

### Adding New Features

To add a new feature:

1. **Add configuration:**
   - Add flag in `GetCreateFlags()`
   - Add field to `Driver` struct
   - Parse in `SetConfigFromFlags()`

2. **Implement logic:**
   - Add API methods in `api.go` if needed
   - Implement in appropriate driver method

3. **Test:**
   - Add unit tests
   - Test manually with Docker Machine
   - Test in Rancher (if applicable)

4. **Document:**
   - Update README.md
   - Update configuration.md
   - Add examples

### API Client Extensions

To add new OVH API support:

```go
// api.go
func (a *API) NewOperation(projectID, param string) error {
    path := fmt.Sprintf("/cloud/project/%s/endpoint", projectID)
    body := map[string]string{"key": param}
    return a.Post(path, body, nil)
}

// driver.go
func (d *Driver) UseNewFeature() error {
    client, err := d.getClient()
    if err != nil {
        return err
    }
    return client.NewOperation(d.ProjectID, d.SomeParam)
}
```

## Security Considerations

### Credential Storage

- **Local:** Credentials in `~/.ovh.conf` (file permissions: 600)
- **Rancher:** Encrypted in database
- **Never logged:** Credentials filtered from debug output

### Network Security

- **SSH:** Key-based authentication only
- **Docker TLS:** Automatic certificate generation
- **API:** HTTPS only, request signing

### Instance Security

Default security:
- SSH port 22 open
- All other ports closed by default
- No password authentication

## References

- **OVH API Documentation:** https://api.ovh.com/
- **Docker Machine Driver API:** https://github.com/docker/machine/tree/master/libmachine/drivers
- **OVH Go SDK:** https://github.com/ovh/go-ovh
- **OpenStack API:** https://docs.openstack.org/api-ref/ (OVH uses OpenStack internally)

---

[← Troubleshooting](troubleshooting.md) | [Back to README](../README.md)
