# Configuration Reference

Complete reference for all configuration options, flags, and settings for the OVHcloud Driver.

## Table of Contents

- [Authentication Options](#authentication-options)
- [Instance Configuration](#instance-configuration)
- [Network Configuration](#network-configuration)
- [Managed Kubernetes (MKS)](#managed-kubernetes-mks)
- [Advanced Options](#advanced-options)
- [Configuration File](#configuration-file)
- [Environment Variables](#environment-variables)

## Authentication Options

### Application Key

**Flag:** `--ovh-application-key`  
**Environment:** `OVH_APPLICATION_KEY`  
**Config file:** `application_key` in `ovh.conf`  
**Required:** Yes (unless in config file)

Your OVH API application key. Generated at https://api.ovh.com/createToken/

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-application-key "Ab1cD2eF3gH4iJ5k" \
  my-node
```

### Application Secret

**Flag:** `--ovh-application-secret`  
**Environment:** `OVH_APPLICATION_SECRET`  
**Config file:** `application_secret` in `ovh.conf`  
**Required:** Yes (unless in config file)

Your OVH API application secret.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-application-secret "Xy9zA8bC7dE6fG5hI4jK3lM2nO1" \
  my-node
```

### Consumer Key

**Flag:** `--ovh-consumer-key`  
**Environment:** `OVH_CONSUMER_KEY`  
**Config file:** `consumer_key` in `ovh.conf`  
**Required:** Yes (unless in config file)

Your OVH API consumer key.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-consumer-key "Qw9eR8tY7uI6oP5aS4dF3gH2jK1" \
  my-node
```

### Endpoint

**Flag:** `--ovh-endpoint`  
**Environment:** `OVH_ENDPOINT`  
**Config file:** `endpoint` in `ovh.conf`  
**Default:** `ovh-eu`

OVH API endpoint for your region.

**Available endpoints:**
- `ovh-eu` — Europe (default)
- `ovh-ca` — Canada
- `ovh-us` — United States
- `soyoustart-eu` — So you Start Europe
- `soyoustart-ca` — So you Start Canada
- `kimsufi-eu` — Kimsufi Europe
- `kimsufi-ca` — Kimsufi Canada

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-endpoint ovh-ca \
  canada-node
```

## Instance Configuration

### Project

**Flag:** `--ovh-project`  
**Default:** Auto-detected (if you have only one project)  
**Required:** Only if you have multiple OVH Cloud projects

OVH Cloud project name, description, or ID.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-project "my-production-project" \
  prod-node
```

**Finding your project ID:**
1. Go to: https://www.ovh.com/manager/cloud/
2. Select your project
3. Project ID is in the URL: `#/iaas/pci/project/{PROJECT_ID}`

### Region

**Flag:** `--ovh-region`  
**Default:** `GRA1`

OVHcloud datacenter region.

**Available regions:**

| Region | Location | Description |
|--------|----------|-------------|
| `GRA1` | Gravelines, France | Default region |
| `GRA3` | Gravelines, France | Newer datacenter |
| `GRA5` | Gravelines, France | Latest datacenter |
| `GRA7` | Gravelines, France | Latest datacenter |
| `SBG1` | Strasbourg, France | Eastern France |
| `SBG3` | Strasbourg, France | Newer datacenter |
| `SBG5` | Strasbourg, France | Latest datacenter |
| `BHS1` | Beauharnois, Canada | North America |
| `BHS3` | Beauharnois, Canada | Newer datacenter |
| `BHS5` | Beauharnois, Canada | Latest datacenter |
| `DE1` | Frankfurt, Germany | Germany |
| `UK1` | London, United Kingdom | UK |
| `WAW1` | Warsaw, Poland | Eastern Europe |
| `SGP1` | Singapore | Asia-Pacific |
| `SYD1` | Sydney, Australia | Asia-Pacific |

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-region SBG5 \
  strasbourg-node
```

### Flavor (Instance Type)

**Flag:** `--ovh-flavor`  
**Default:** `vps-ssd-1`

Instance type/size. Can be name or ID.

**Common flavors:**

| Flavor | vCPUs | RAM | Disk | Price/hour* |
|--------|-------|-----|------|-------------|
| `vps-ssd-1` | 1 | 2 GB | 10 GB | ~€0.004 |
| `vps-ssd-2` | 1 | 4 GB | 20 GB | ~€0.008 |
| `vps-ssd-3` | 2 | 8 GB | 40 GB | ~€0.016 |
| `b2-7` | 2 | 7 GB | 50 GB | ~€0.031 |
| `b2-15` | 4 | 15 GB | 100 GB | ~€0.064 |
| `b2-30` | 8 | 30 GB | 200 GB | ~€0.127 |
| `b2-60` | 16 | 60 GB | 400 GB | ~€0.254 |
| `b2-120` | 32 | 120 GB | 400 GB | ~€0.507 |

\* Prices approximate, vary by region. Check https://www.ovhcloud.com/en/public-cloud/prices/ for current pricing.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-flavor b2-15 \
  large-node
```

**List available flavors:**
```bash
# Using OVH API CLI (if installed)
ovh-eu cloud project {PROJECT_ID} flavor list --region GRA1
```

### Image (Operating System)

**Flag:** `--ovh-image`  
**Default:** `Ubuntu 22.04`

OS image name or ID.

**Common images:**

| Image | SSH User | Notes |
|-------|----------|-------|
| `Ubuntu 22.04` | `ubuntu` | LTS, recommended |
| `Ubuntu 20.04` | `ubuntu` | Older LTS |
| `Debian 12` | `debian` | Latest stable |
| `Debian 11` | `debian` | Previous stable |
| `Fedora 38` | `fedora` | Latest Fedora |
| `CentOS Stream 9` | `centos` | RHEL-compatible |
| `Rocky Linux 9` | `rocky` | RHEL-compatible |
| `CoreOS stable` | `core` | Container-optimized |

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-image "Debian 12" \
  --ovh-ssh-user debian \
  debian-node
```

**Note:** Image names may vary by region. Use the exact name as shown in OVH console.

### SSH User

**Flag:** `--ovh-ssh-user`  
**Default:** `ubuntu`

SSH username for the instance. Must match the OS image.

**Common SSH users:**
- Ubuntu: `ubuntu`
- Debian: `debian`
- Fedora: `fedora`
- CentOS/Rocky: `centos` or `rocky`
- CoreOS: `core`
- Alpine: `alpine`

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-image "Debian 11" \
  --ovh-ssh-user debian \
  debian-node
```

### SSH Key

**Flag:** `--ovh-ssh-key`  
**Default:** Auto-generated

Name of an existing SSH key in your OVH project. If not specified, a new key will be generated.

**Using an existing key:**
```bash
docker-machine create -d ovh \
  --ovh-ssh-key "my-team-key" \
  team-node
```

**Benefits of using existing keys:**
- Reuse keys across multiple machines
- Avoid cluttering OVH with auto-generated keys
- Use keys already in your SSH agent

**Requirements:**
- Key must exist in your OVH project
- Private key must be accessible (in `~/.ssh/` or SSH agent)

**Creating keys in OVH:**
1. Go to: https://www.ovh.com/manager/cloud/
2. Select your project → **SSH Keys**
3. Click **Add a key**
4. Paste your public key

### Billing Period

**Flag:** `--ovh-billing-period`  
**Default:** `hourly`

Billing period for the instance.

**Options:**
- `hourly` — Pay per hour, billed monthly
- `monthly` — Fixed monthly price (discounted ~50%)

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-billing-period monthly \
  long-term-node
```

**Cost comparison:**
```
vps-ssd-1 hourly: €2.88/month
vps-ssd-1 monthly: €1.67/month
```

**Note:** Monthly instances cannot be deleted within the first month without charges for the full month.

## Network Configuration

### Private Network (vRack)

**Flag:** `--ovh-private-network`  
**Default:** Public network

Connect instance to an OVHcloud vRack private network.

**Value:** VLAN ID (number) or network name

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-private-network 42 \
  --ovh-region GRA1 \
  private-node
```

**Setting up vRack:**

1. **Create vRack** (if needed):
   - Go to: https://www.ovh.com/manager/cloud/index.html#/vrack/new
   - Free of charge

2. **Attach project to vRack:**
   - Go to: https://www.ovh.com/manager/cloud/index.html#/vrack
   - Add your Cloud project

3. **Create VLANs:**
   - Go to your Cloud project → **Private networks**
   - Create networks with VLAN IDs (0-4000)

4. **Use in driver:**
   ```bash
   docker-machine create -d ovh \
     --ovh-private-network 100 \
     vrack-node
   ```

**Post-creation network setup:**

The private interface needs manual configuration:

```bash
docker-machine ssh vrack-node

# Create interface config
sudo tee /etc/network/interfaces.d/99-vrack.cfg << 'EOF'
auto ens4
iface ens4 inet dhcp
EOF

# Enable interface
sudo ifup ens4

# Verify
ip addr show ens4
```

**For persistent setup across reboots:**
Add the configuration to cloud-init or a startup script.

## Managed Kubernetes (MKS)

Experimental support for OVHcloud Managed Kubernetes Service.

### Enable MKS Mode

**Flag:** `--ovh-hosted-mks`  
**Default:** `false`

Enable Managed Kubernetes cluster creation instead of a single VM.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-nodepool-flavor b2-7 \
  --ovh-mks-nodepool-size 3 \
  k8s-cluster
```

### MKS Cluster Name

**Flag:** `--ovh-mks-cluster-name`  
**Required when:** `--ovh-hosted-mks` is enabled

Name for the Kubernetes cluster.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name production-cluster \
  prod-k8s
```

### MKS Kubernetes Version

**Flag:** `--ovh-mks-version`  
**Default:** Latest available version

Kubernetes version for the cluster.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-version 1.28 \
  k8s-cluster
```

### MKS Nodepool Name

**Flag:** `--ovh-mks-nodepool-name`  
**Default:** `default`

Name for the nodepool within the cluster.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-nodepool-name worker-pool \
  k8s-cluster
```

### MKS Nodepool Flavor

**Flag:** `--ovh-mks-nodepool-flavor`  
**Default:** `vps-ssd-1`  
**Required when:** `--ovh-hosted-mks` is enabled

Instance type for nodepool nodes.

**Recommended flavors for K8s:**
- `b2-7` — Small clusters (2 vCPU, 7 GB RAM)
- `b2-15` — Medium clusters (4 vCPU, 15 GB RAM)
- `b2-30` — Large clusters (8 vCPU, 30 GB RAM)

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-nodepool-flavor b2-15 \
  k8s-cluster
```

### MKS Nodepool Size

**Flag:** `--ovh-mks-nodepool-size`  
**Default:** `1`  
**Required when:** `--ovh-hosted-mks` is enabled

Number of nodes in the nodepool.

**Example:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-nodepool-size 5 \
  k8s-cluster
```

**Note:** MKS mode is experimental. Current limitations:
- Limited scaling operations
- kubeconfig must be retrieved manually from OVH console
- See [Architecture](architecture.md) for details

## Advanced Options

### Docker Engine Options

Pass Docker engine flags during machine creation:

```bash
docker-machine create -d ovh \
  --engine-opt log-driver=syslog \
  --engine-opt log-opt="syslog-address=udp://syslog-server:514" \
  logging-node
```

### Storage Options

```bash
docker-machine create -d ovh \
  --engine-storage-driver overlay2 \
  storage-node
```

### Registry Mirrors

```bash
docker-machine create -d ovh \
  --engine-registry-mirror https://registry.example.com \
  registry-node
```

### Labels

```bash
docker-machine create -d ovh \
  --engine-label environment=production \
  --engine-label team=platform \
  labeled-node
```

## Configuration File

### Format

Create `~/.ovh.conf`:

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=YOUR_APPLICATION_KEY
application_secret=YOUR_APPLICATION_SECRET
consumer_key=YOUR_CONSUMER_KEY
```

### Multiple Endpoints

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=EU_APP_KEY
application_secret=EU_APP_SECRET
consumer_key=EU_CONSUMER_KEY

[ovh-ca]
application_key=CA_APP_KEY
application_secret=CA_APP_SECRET
consumer_key=CA_CONSUMER_KEY
```

Use with `--ovh-endpoint`:

```bash
docker-machine create -d ovh \
  --ovh-endpoint ovh-ca \
  canada-node
```

### File Locations

Checked in order (first found is used):
1. `./ovh.conf` (current directory)
2. `~/.ovh.conf` (user home)
3. `/etc/ovh.conf` (system-wide)

## Environment Variables

All flags can be set via environment variables:

```bash
export OVH_APPLICATION_KEY="your_key"
export OVH_APPLICATION_SECRET="your_secret"
export OVH_CONSUMER_KEY="your_consumer_key"
export OVH_ENDPOINT="ovh-eu"

docker-machine create -d ovh my-node
```

**Priority:** Flags > Environment > Config file

## Complete Example

```bash
docker-machine create -d ovh \
  --ovh-application-key "YOUR_KEY" \
  --ovh-application-secret "YOUR_SECRET" \
  --ovh-consumer-key "YOUR_CONSUMER" \
  --ovh-endpoint ovh-eu \
  --ovh-project "production-project" \
  --ovh-region SBG5 \
  --ovh-flavor b2-15 \
  --ovh-image "Ubuntu 22.04" \
  --ovh-ssh-user ubuntu \
  --ovh-ssh-key "team-key" \
  --ovh-private-network 100 \
  --ovh-billing-period monthly \
  --engine-label environment=production \
  production-node-1
```

## Next Steps

- [Rancher Integration](rancher-integration.md) — Use this driver with Rancher
- [Troubleshooting](troubleshooting.md) — Common issues and solutions
- [Architecture](architecture.md) — How the driver works

---

[← Installation](installation.md) | [Back to README](../README.md) | [Rancher Integration →](rancher-integration.md)
