# OVHcloud Driver for Rancher

[![OVHcloud Driver](https://raw.githubusercontent.com/sneederco/docker-machine-driver-ovh/master/img/logo.png)](https://github.com/sneederco/docker-machine-driver-ovh)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/sneederco/docker-machine-driver-ovh)](https://goreportcard.com/report/github.com/sneederco/docker-machine-driver-ovh)

**Maintained by [Sneederco](https://github.com/sneederco)**

A Docker Machine driver for OVHcloud Public Cloud. Provision compute instances and Managed Kubernetes clusters on OVHcloud with Docker Machine or Rancher.

```bash
# Create your first OVHcloud instance in seconds
docker-machine create -d ovh my-first-node
```

## ✨ Features

- 🚀 **Fast provisioning** — Create OVHcloud instances in under 60 seconds
- ☸️ **Rancher integration** — Native support for Rancher node drivers
- 🌍 **Multi-region** — Deploy across OVHcloud's global datacenters
- 🔐 **Secure by default** — SSH key management and private networking
- 💰 **Cost flexible** — Hourly or monthly billing options
- 🛡️ **vRack support** — Integrate with OVHcloud private networks
- ⎈ **Managed Kubernetes** — Experimental support for OVH MKS clusters

## 📦 Quick Start

Get up and running in 5 steps:

### 1. Install the driver

```bash
# Install via Go
go install github.com/sneederco/docker-machine-driver-ovh@latest

# Make it available to Docker Machine
ln -s $(go env GOPATH)/bin/docker-machine-driver-ovh /usr/local/bin/
```

Or download a pre-built binary from the [releases page](https://github.com/sneederco/docker-machine-driver-ovh/releases).

### 2. Get OVH API credentials

Create your API keys at: https://api.ovh.com/createToken/

**Required permissions:**
- `GET /*`
- `POST /*`
- `PUT /*`
- `DELETE /*`

You'll receive:
- Application Key
- Application Secret
- Consumer Key

### 3. Configure credentials

Create `~/.ovh.conf`:

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=YOUR_APPLICATION_KEY
application_secret=YOUR_APPLICATION_SECRET
consumer_key=YOUR_CONSUMER_KEY
```

Or use environment variables:

```bash
export OVH_APPLICATION_KEY="YOUR_APPLICATION_KEY"
export OVH_APPLICATION_SECRET="YOUR_APPLICATION_SECRET"
export OVH_CONSUMER_KEY="YOUR_CONSUMER_KEY"
```

### 4. Create your first instance

```bash
docker-machine create -d ovh my-node
```

### 5. Start using it!

```bash
# Get environment variables
eval $(docker-machine env my-node)

# Run a container
docker run hello-world
```

## 📋 Configuration Options

| Flag | Environment Variable | Description | Default | Required |
|------|---------------------|-------------|---------|----------|
| `--ovh-application-key` | `OVH_APPLICATION_KEY` | OVH API application key | - | Yes* |
| `--ovh-application-secret` | `OVH_APPLICATION_SECRET` | OVH API application secret | - | Yes* |
| `--ovh-consumer-key` | `OVH_CONSUMER_KEY` | OVH API consumer key | - | Yes* |
| `--ovh-endpoint` | `OVH_ENDPOINT` | OVH API endpoint | `ovh-eu` | No |
| `--ovh-project` | - | OVH Cloud project name or ID | Auto-detect | Conditional** |
| `--ovh-region` | - | OVH region | `GRA1` | No |
| `--ovh-flavor` | - | Instance flavor/type | `vps-ssd-1` | No |
| `--ovh-image` | - | OS image name or ID | `Ubuntu 22.04` | No |
| `--ovh-ssh-user` | - | SSH username | `ubuntu` | No |
| `--ovh-ssh-key` | - | Existing SSH key name | Auto-generated | No |
| `--ovh-private-network` | - | vRack network (VLAN ID or name) | `public` | No |
| `--ovh-billing-period` | - | Billing period: `hourly` or `monthly` | `hourly` | No |
| `--ovh-hosted-mks` | - | Enable Managed Kubernetes mode | `false` | No |
| `--ovh-mks-cluster-name` | - | MKS cluster name | - | If MKS enabled |
| `--ovh-mks-version` | - | Kubernetes version | Latest | No |
| `--ovh-mks-nodepool-name` | - | MKS nodepool name | `default` | No |
| `--ovh-mks-nodepool-flavor` | - | MKS nodepool flavor | `vps-ssd-1` | If MKS enabled |
| `--ovh-mks-nodepool-size` | - | MKS nodepool node count | `1` | If MKS enabled |

\* Can be stored in `ovh.conf` instead  
\*\* Required only if you have multiple OVH Cloud projects

**Available regions:** GRA1, GRA3, SBG1, SBG5, BHS5, DE1, UK1, WAW1  
**Popular flavors:** vps-ssd-1, vps-ssd-2, vps-ssd-3, b2-7, b2-15, b2-30  
**Common images:** Ubuntu 22.04, Ubuntu 20.04, Debian 12, Debian 11

[Full configuration reference →](docs/configuration.md)

## 🎯 Common Use Cases

### Deploy in a specific region with custom resources

```bash
docker-machine create -d ovh \
  --ovh-region SBG5 \
  --ovh-flavor b2-15 \
  --ovh-image "Debian 12" \
  --ovh-ssh-user debian \
  production-node
```

### Use private networking with vRack

```bash
docker-machine create -d ovh \
  --ovh-private-network 42 \
  --ovh-region GRA1 \
  private-node
```

### Create a Managed Kubernetes cluster (Experimental)

```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-nodepool-flavor b2-7 \
  --ovh-mks-nodepool-size 3 \
  k8s-cluster
```

## 🔗 Rancher Integration

This driver works seamlessly with Rancher for automated node provisioning.

**Step-by-step guide:** [Rancher Integration →](docs/rancher-integration.md)

### Quick setup:

1. Navigate to **Cluster Management** → **Drivers** → **Node Drivers**
2. Add a new Node Driver:
   - **Download URL:** `https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-linux-amd64`
   - **Custom UI URL:** *(optional)*
3. Activate the driver
4. Create a new cluster using the OVH node template

## 🔧 Troubleshooting

### Common issues:

**"Could not create a connection to OVH API"**
- Verify your credentials in `~/.ovh.conf`
- Check API endpoint matches your region
- Ensure API keys have sufficient permissions

**"Project not found"**
- Specify project explicitly with `--ovh-project`
- Verify project exists in OVH console

**SSH connection issues**
- Check SSH user matches your OS (`ubuntu`, `debian`, `core`, `admin`)
- Verify SSH keys are properly configured
- Ensure security groups allow SSH (port 22)

[Full troubleshooting guide →](docs/troubleshooting.md)

## 📚 Documentation

- [Installation Guide](docs/installation.md) — Detailed installation steps for all platforms
- [Configuration Reference](docs/configuration.md) — Complete flag and option documentation
- [Rancher Integration](docs/rancher-integration.md) — Step-by-step Rancher setup
- [Architecture Overview](docs/architecture.md) — How the driver works internally
- [Troubleshooting Guide](docs/troubleshooting.md) — Common issues and solutions

## 🤝 Contributing

We welcome contributions! Whether it's:
- 🐛 Bug reports
- 💡 Feature requests
- 📝 Documentation improvements
- 🔧 Code contributions

Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📖 Related Resources

- **OVH Cloud Console:** https://www.ovh.com/manager/cloud
- **OVH Cloud Offers:** https://www.ovhcloud.com/en/public-cloud/
- **OVH API Documentation:** https://api.ovh.com/
- **Docker Machine Docs:** https://docs.docker.com/machine/
- **Rancher Docs:** https://rancher.com/docs/

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Maintained by [Sneederco](https://github.com/sneederco)** | **Issues:** [GitHub Issues](https://github.com/sneederco/docker-machine-driver-ovh/issues) | **Repo:** [sneederco/docker-machine-driver-ovh](https://github.com/sneederco/docker-machine-driver-ovh)
