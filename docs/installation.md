# Installation Guide

Complete installation instructions for the OVHcloud Driver for Rancher on all supported platforms.

## Prerequisites

Before installing, ensure you have:

- **Docker** (version 19.03+)
- **Docker Machine** (version 0.16.0+)
- An **OVHcloud account** with a Public Cloud project
- **Go 1.21+** (only if building from source)

## Method 1: Install via Go (Recommended)

The easiest way to install the latest version:

```bash
# Install the driver
go install github.com/sneederco/docker-machine-driver-ovh@latest

# Make it available to Docker Machine
ln -s $(go env GOPATH)/bin/docker-machine-driver-ovh /usr/local/bin/docker-machine-driver-ovh
```

**Verify installation:**

```bash
docker-machine create --driver ovh --help
```

## Method 2: Download Pre-Built Binary

Download the latest release for your platform:

### Linux (amd64)

```bash
# Download latest release
curl -L https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-linux-amd64 \
  -o /usr/local/bin/docker-machine-driver-ovh

# Make executable
chmod +x /usr/local/bin/docker-machine-driver-ovh
```

### Linux (arm64)

```bash
curl -L https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-linux-arm64 \
  -o /usr/local/bin/docker-machine-driver-ovh

chmod +x /usr/local/bin/docker-machine-driver-ovh
```

### macOS (Intel)

```bash
curl -L https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-darwin-amd64 \
  -o /usr/local/bin/docker-machine-driver-ovh

chmod +x /usr/local/bin/docker-machine-driver-ovh
```

### macOS (Apple Silicon)

```bash
curl -L https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-darwin-arm64 \
  -o /usr/local/bin/docker-machine-driver-ovh

chmod +x /usr/local/bin/docker-machine-driver-ovh
```

### Windows

1. Download: [docker-machine-driver-ovh-windows-amd64.exe](https://github.com/sneederco/docker-machine-driver-ovh/releases/latest/download/docker-machine-driver-ovh-windows-amd64.exe)
2. Rename to `docker-machine-driver-ovh.exe`
3. Move to a directory in your `PATH` (e.g., `C:\Program Files\Docker Machine\`)

## Method 3: Build from Source

For development or specific versions:

### Clone the repository

```bash
git clone https://github.com/sneederco/docker-machine-driver-ovh.git
cd docker-machine-driver-ovh
```

### Build the driver

```bash
# Download dependencies
go mod download
go mod vendor

# Build
make build

# Or build manually
go build -o docker-machine-driver-ovh .
```

### Install locally

```bash
# Using Make
make install

# Or manually
cp docker-machine-driver-ovh /usr/local/bin/
chmod +x /usr/local/bin/docker-machine-driver-ovh
```

## Installing Docker Machine

If you don't have Docker Machine installed:

### Linux/macOS

```bash
# Download
base=https://github.com/docker/machine/releases/download/v0.16.2
curl -L $base/docker-machine-$(uname -s)-$(uname -m) >/usr/local/bin/docker-machine

# Make executable
chmod +x /usr/local/bin/docker-machine
```

### Windows (PowerShell as Administrator)

```powershell
if ([Environment]::Is64BitOperatingSystem) {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest `
        -Uri "https://github.com/docker/machine/releases/download/v0.16.2/docker-machine-Windows-x86_64.exe" `
        -OutFile "C:\Program Files\Docker\docker-machine.exe"
}
```

**Verify installation:**

```bash
docker-machine version
```

## Setting Up OVH API Credentials

### Generate API Keys

1. Go to: https://api.ovh.com/createToken/
2. Log in with your OVH account
3. Set permissions:
   - `GET /*`
   - `POST /*`
   - `PUT /*`
   - `DELETE /*`
4. Set validity period (recommend 1+ years)
5. Click **Create keys**

You'll receive:
- **Application Key**
- **Application Secret**
- **Consumer Key**

### Configure Credentials

#### Option 1: Configuration File (Recommended)

Create `~/.ovh.conf`:

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=YOUR_APPLICATION_KEY
application_secret=YOUR_APPLICATION_SECRET
consumer_key=YOUR_CONSUMER_KEY
```

**File locations** (checked in order):
1. `./ovh.conf` (current directory)
2. `~/.ovh.conf` (user home)
3. `/etc/ovh.conf` (system-wide)

#### Option 2: Environment Variables

```bash
export OVH_APPLICATION_KEY="YOUR_APPLICATION_KEY"
export OVH_APPLICATION_SECRET="YOUR_APPLICATION_SECRET"
export OVH_CONSUMER_KEY="YOUR_CONSUMER_KEY"
export OVH_ENDPOINT="ovh-eu"  # Optional
```

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) to persist.

#### Option 3: Command-Line Flags

```bash
docker-machine create -d ovh \
  --ovh-application-key "YOUR_KEY" \
  --ovh-application-secret "YOUR_SECRET" \
  --ovh-consumer-key "YOUR_CONSUMER_KEY" \
  my-node
```

**Priority order:** Flags > Environment > Config file

## Verifying Installation

### Check driver availability

```bash
docker-machine create --driver ovh --help
```

You should see OVH-specific flags listed.

### Create a test machine

```bash
docker-machine create -d ovh test-node
```

If successful, you'll see output like:

```
Running pre-create checks...
Creating machine...
(test-node) Creating SSH key...
(test-node) Creating instance...
(test-node) Waiting for IP address...
(test-node) Waiting for SSH to be available...
Waiting for machine to be running, this may take a few minutes...
```

### Test the machine

```bash
# Get environment
eval $(docker-machine env test-node)

# Run a container
docker run hello-world

# Cleanup
docker-machine rm -f test-node
```

## Troubleshooting Installation

### "docker-machine: command not found"

Docker Machine is not installed or not in your PATH.

**Solution:** Follow [Installing Docker Machine](#installing-docker-machine) above.

### "Could not find driver ovh"

The driver binary is not in a location Docker Machine can find.

**Solution:** Ensure `docker-machine-driver-ovh` is in your PATH:

```bash
# Check PATH
echo $PATH

# Verify binary location
which docker-machine-driver-ovh

# Should output: /usr/local/bin/docker-machine-driver-ovh
```

### "Permission denied" when running driver

The driver binary is not executable.

**Solution:**

```bash
chmod +x /usr/local/bin/docker-machine-driver-ovh
```

### "Could not create a connection to OVH API"

Your OVH credentials are not configured or are incorrect.

**Solution:**
1. Verify credentials at https://api.ovh.com/console/
2. Check your `~/.ovh.conf` file format
3. Ensure the endpoint matches your OVH region
4. Test API access:

```bash
curl -X GET \
  "https://eu.api.ovh.com/1.0/me" \
  -H "X-Ovh-Application: $OVH_APPLICATION_KEY"
```

### Build fails with "package not found"

Dependencies are not downloaded.

**Solution:**

```bash
go mod download
go mod vendor
```

## Updating the Driver

### From Go install

```bash
go install github.com/sneederco/docker-machine-driver-ovh@latest
```

### From binary

Download and replace the binary following [Method 2](#method-2-download-pre-built-binary) above.

### Check version

```bash
docker-machine-driver-ovh --version
```

## Uninstalling

```bash
# Remove the driver
rm /usr/local/bin/docker-machine-driver-ovh

# Remove configuration (optional)
rm ~/.ovh.conf

# Remove Docker Machine (optional)
rm /usr/local/bin/docker-machine
```

## Next Steps

- **Configuration:** See [Configuration Reference](configuration.md)
- **Rancher Setup:** See [Rancher Integration](rancher-integration.md)
- **Usage Examples:** See the main [README](../README.md)

## Getting Help

- **Issues:** https://github.com/sneederco/docker-machine-driver-ovh/issues
- **Discussions:** https://github.com/sneederco/docker-machine-driver-ovh/discussions
- **OVH Support:** https://help.ovhcloud.com/

---

[← Back to README](../README.md) | [Configuration Reference →](configuration.md)
