# Rancher Integration

Step-by-step guide for integrating the OVHcloud Driver with Rancher for automated node provisioning and cluster management.

## Overview

This driver integrates seamlessly with Rancher, allowing you to:
- ✅ Provision OVHcloud instances directly from Rancher UI
- ✅ Scale Kubernetes clusters with OVHcloud nodes
- ✅ Manage node pools across multiple regions
- ✅ Automate infrastructure provisioning

## Prerequisites

- **Rancher** 2.5+ installed and running
- **OVHcloud account** with API credentials
- **Admin access** to Rancher
- The **OVH driver binary** for your Rancher server's architecture

## Installation Steps

### Step 1: Add the Node Driver

1. **Navigate to Node Drivers:**
   - Log into Rancher
   - Click **☰** menu → **Cluster Management**
   - Select **Drivers** from the left sidebar
   - Click the **Node Drivers** tab

2. **Add Custom Driver:**
   - Click **Add Node Driver** button

3. **Configure the driver:**
   
   **Download URL:**
   ```
   https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-linux-amd64
   ```
   
   **For ARM-based Rancher servers:**
   ```
   https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-linux-arm64
   ```
   
   **Custom UI URL:** *(Leave empty for now)*
   
   **Whitelist Domains:** *(Leave empty)*

4. **Click Create**

5. **Activate the driver:**
   - Find "OVH" in the driver list
   - Click the **⋮** menu
   - Select **Activate**

### Step 2: Wait for Driver Download

Rancher will download and install the driver. This may take 1-2 minutes.

**Check status:**
- Status should change to **Active**
- If it shows an error, check the download URL and try again

### Step 3: Create a Node Template

1. **Navigate to Node Templates:**
   - Click **☰** menu → **Cluster Management**
   - Select **RKE1 Configuration** → **Node Templates**
   - Click **Add Template**

2. **Select OVH driver:**
   - Find and click **OVH** in the driver list

3. **Configure Account Access:**

   **OVH API Credentials:**
   - **Application Key:** Your OVH application key
   - **Application Secret:** Your OVH application secret
   - **Consumer Key:** Your OVH consumer key
   - **Endpoint:** Select your OVH region (usually `ovh-eu`)
   
   **Get credentials at:** https://api.ovh.com/createToken/

4. **Configure Instance Options:**

   **Project:**
   - **OVH Project:** Your project name or ID (leave empty if you have only one)

   **Instance Configuration:**
   - **Region:** Select datacenter (e.g., `GRA1`, `SBG5`, `BHS5`)
   - **Flavor:** Select instance type (e.g., `b2-7`, `b2-15`)
   - **Image:** Select OS image (e.g., `Ubuntu 22.04`)
   - **SSH User:** Enter SSH username (e.g., `ubuntu`)

   **Network:**
   - **Private Network:** (Optional) VLAN ID for vRack networking
   - **SSH Key:** (Optional) Name of existing SSH key in OVH project

   **Billing:**
   - **Billing Period:** `hourly` (recommended) or `monthly`

5. **Configure Engine Options:**

   **Docker Settings:**
   - **Engine Install URL:** *(Usually auto-detected)*
   - **Docker Version:** *(Leave default)*

6. **Name the template:**
   - Enter a name: e.g., `ovh-gra1-b2-7-ubuntu`

7. **Click Create**

### Step 4: Create a Cluster

#### Option A: Create New RKE1 Cluster

1. **Start cluster creation:**
   - Click **☰** → **Cluster Management**
   - Click **Create**
   - Select **Custom** under **Infrastructure Provider**

2. **Configure cluster:**
   - **Cluster Name:** Enter a name (e.g., `ovh-production`)
   - **Kubernetes Version:** Select version
   - **Network Provider:** Select (e.g., `Canal`, `Calico`)
   - **Cloud Provider:** Leave as `None` or select if needed

3. **Add node pools:**
   - Click **Add Node Template**
   - Select your OVH template
   - Configure roles:
     - ☑️ **etcd** (for control plane nodes)
     - ☑️ **Control Plane** (for control plane nodes)
     - ☑️ **Worker** (for worker nodes)
   - Set **Count:** Number of nodes (e.g., `3`)
   - Click **Create**

4. **Create additional pools** (optional):
   - Add separate pools for workers
   - Use different instance sizes if needed

5. **Click Create** to provision the cluster

#### Option B: Scale Existing Cluster

1. **Open your cluster:**
   - Click **☰** → **Cluster Management**
   - Click your cluster name

2. **Edit cluster:**
   - Click **⋮** (three dots) → **Edit Config**

3. **Add node pool:**
   - Scroll to **Node Pools**
   - Click **Add Node Pool**
   - Select OVH template
   - Configure roles and count
   - Click **Save**

### Step 5: Monitor Provisioning

1. **Watch node creation:**
   - Nodes will appear in **Nodes** tab
   - Status will show: `Provisioning` → `Waiting` → `Active`

2. **Check logs if needed:**
   - Click node name → **⋮** → **View Logs**

**Typical provisioning time:** 3-5 minutes per node

## Using Multiple Node Templates

Create templates for different use cases:

### Control Plane Template
```
Name: ovh-control-gra1
Flavor: b2-7
Region: GRA1
Image: Ubuntu 22.04
Roles: etcd, Control Plane
```

### Worker Template
```
Name: ovh-worker-sbg5
Flavor: b2-15
Region: SBG5
Image: Ubuntu 22.04
Roles: Worker
```

### GPU Worker Template
```
Name: ovh-gpu-bhs5
Flavor: t1-45
Region: BHS5
Image: Ubuntu 22.04
Roles: Worker
```

## Multi-Region Setup

Deploy clusters across OVHcloud regions:

**Create templates per region:**
1. **Europe (France):** GRA1, GRA5, SBG5
2. **Europe (Germany):** DE1
3. **Europe (UK):** UK1
4. **North America:** BHS5
5. **Asia:** SGP1

**Add node pools from each:**
```
ovh-gra1-control (etcd + Control Plane)
ovh-sbg5-control (etcd + Control Plane)
ovh-gra1-worker (Worker)
ovh-sbg5-worker (Worker)
```

**Benefits:**
- High availability across datacenters
- Geographic distribution
- Fault tolerance

## Using vRack Private Networks

### Prerequisites

1. Create vRack at: https://www.ovh.com/manager/cloud/index.html#/vrack/new
2. Attach your OVH Cloud project to the vRack
3. Create private networks with VLAN IDs

### Configure in Rancher

1. **In node template:**
   - Set **Private Network** to your VLAN ID (e.g., `42`)

2. **After nodes provision:**
   - Nodes will have public and private interfaces
   - Private interface requires manual configuration

3. **Configure private networking:**

   **Using cloud-init:**
   Add to **Cloud Init** in template:
   ```yaml
   #cloud-config
   write_files:
     - path: /etc/network/interfaces.d/99-vrack.cfg
       content: |
         auto ens4
         iface ens4 inet dhcp
   runcmd:
     - ifup ens4
   ```

   **Or manually via SSH:**
   ```bash
   docker-machine ssh node-name sudo tee /etc/network/interfaces.d/99-vrack.cfg << 'EOF'
   auto ens4
   iface ens4 inet dhcp
   EOF
   docker-machine ssh node-name sudo ifup ens4
   ```

## Managed Kubernetes (MKS) Mode

⚠️ **Experimental Feature**

The driver supports provisioning OVHcloud Managed Kubernetes clusters, but this is currently experimental.

**Limitations:**
- No UI integration in Rancher yet
- Must use Docker Machine CLI directly
- Limited to basic cluster operations

**Usage:**
```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name rancher-mks \
  --ovh-mks-nodepool-flavor b2-15 \
  --ovh-mks-nodepool-size 3 \
  mks-cluster
```

**Import to Rancher:**
1. Get kubeconfig from OVH console
2. Import cluster using **Import Existing** in Rancher

## Automation with Terraform

Use Rancher's Terraform provider with OVH driver:

```hcl
resource "rancher2_node_template" "ovh_template" {
  name = "ovh-production"
  
  driver_id = "ovh"
  
  ovh_config {
    application_key    = var.ovh_application_key
    application_secret = var.ovh_application_secret
    consumer_key       = var.ovh_consumer_key
    region             = "GRA5"
    flavor             = "b2-15"
    image              = "Ubuntu 22.04"
    ssh_user           = "ubuntu"
    billing_period     = "hourly"
  }
}

resource "rancher2_node_pool" "ovh_workers" {
  cluster_id        = rancher2_cluster.cluster.id
  name              = "ovh-workers"
  hostname_prefix   = "worker-"
  node_template_id  = rancher2_node_template.ovh_template.id
  quantity          = 3
  control_plane     = false
  etcd              = false
  worker            = true
}
```

## Troubleshooting Rancher Integration

### Driver Not Showing Up

**Problem:** OVH driver not visible in node drivers list.

**Solution:**
1. Verify download URL is correct
2. Check Rancher server can access GitHub
3. Check Rancher logs: `kubectl logs -n cattle-system -l app=rancher`

### Provisioning Fails

**Problem:** Nodes stuck in "Provisioning" state.

**Solution:**
1. Check node logs in Rancher UI
2. Verify OVH API credentials
3. Ensure project has quota for instances
4. Check selected flavor is available in region

### Authentication Errors

**Problem:** "Could not create a connection to OVH API"

**Solution:**
1. Verify API keys at https://api.ovh.com/console/
2. Check endpoint matches your OVH account region
3. Ensure keys have required permissions (GET, POST, PUT, DELETE on /*)

### SSH Connection Issues

**Problem:** Node provisioned but Rancher can't connect via SSH.

**Solution:**
1. Verify SSH user matches OS image
2. Check security groups allow SSH (port 22)
3. Ensure SSH key is properly configured
4. Test SSH manually: `ssh ubuntu@<node-ip>`

### Cluster Creation Hangs

**Problem:** Cluster stuck waiting for nodes to become active.

**Solution:**
1. Check node status in OVH console
2. Verify instances have internet access
3. Check cloud-init logs on instance:
   ```bash
   docker-machine ssh node sudo cloud-init status --wait
   docker-machine ssh node sudo journalctl -u cloud-init
   ```

### Network Issues in vRack

**Problem:** Nodes can't communicate over private network.

**Solution:**
1. Verify VLAN ID is correct
2. Check all nodes are in same vRack
3. Ensure private interface is configured (see [vRack section](#using-vrack-private-networks))
4. Test connectivity:
   ```bash
   docker-machine ssh node1 ping <node2-private-ip>
   ```

## Best Practices

### Security

- ✅ Use separate OVH projects for different environments
- ✅ Limit API key permissions to specific projects
- ✅ Use vRack for sensitive workloads
- ✅ Enable monthly billing for long-lived nodes
- ✅ Use specific SSH keys per cluster

### Scaling

- ✅ Create separate worker and control plane pools
- ✅ Use auto-scaling for worker pools
- ✅ Distribute nodes across regions for HA
- ✅ Use appropriate instance sizes for workload

### Cost Optimization

- ✅ Use hourly billing for dev/test environments
- ✅ Use monthly billing for production (50% discount)
- ✅ Right-size instances (avoid over-provisioning)
- ✅ Delete unused nodes promptly
- ✅ Use spot instances for non-critical workloads (if available)

## Next Steps

- [Configuration Reference](configuration.md) — All available options
- [Troubleshooting Guide](troubleshooting.md) — Common issues
- [Architecture Overview](architecture.md) — How it works

## Resources

- **Rancher Docs:** https://rancher.com/docs/rancher/v2.x/en/cluster-provisioning/rke-clusters/node-pools/
- **OVH Cloud Console:** https://www.ovh.com/manager/cloud/
- **Rancher Forums:** https://forums.rancher.com/

---

[← Configuration](configuration.md) | [Back to README](../README.md) | [Troubleshooting →](troubleshooting.md)
