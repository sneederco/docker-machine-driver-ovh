# Troubleshooting Guide

Common issues, error messages, and solutions for the OVHcloud Driver.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Authentication Issues](#authentication-issues)
- [Instance Creation Issues](#instance-creation-issues)
- [SSH Connection Issues](#ssh-connection-issues)
- [Network Issues](#network-issues)
- [Rancher Integration Issues](#rancher-integration-issues)
- [Managed Kubernetes (MKS) Issues](#managed-kubernetes-mks-issues)
- [Performance Issues](#performance-issues)
- [Getting Help](#getting-help)

## Installation Issues

### "docker-machine: command not found"

**Cause:** Docker Machine is not installed or not in PATH.

**Solution:**

```bash
# Linux/macOS
base=https://github.com/docker/machine/releases/download/v0.16.2
curl -L $base/docker-machine-$(uname -s)-$(uname -m) >/usr/local/bin/docker-machine
chmod +x /usr/local/bin/docker-machine

# Verify
docker-machine version
```

### "Could not find driver ovh"

**Cause:** The OVH driver binary is not in your PATH or not executable.

**Solution:**

```bash
# Check if driver exists
which docker-machine-driver-ovh

# If not found, install it
go install github.com/sneederco/docker-machine-driver-ovh@latest
ln -s $(go env GOPATH)/bin/docker-machine-driver-ovh /usr/local/bin/

# Make it executable
chmod +x /usr/local/bin/docker-machine-driver-ovh

# Verify
docker-machine create --driver ovh --help
```

### "Permission denied" when running driver

**Cause:** Driver binary doesn't have execute permissions.

**Solution:**

```bash
chmod +x /usr/local/bin/docker-machine-driver-ovh
```

### Build fails with "go: module not found"

**Cause:** Go modules are not downloaded.

**Solution:**

```bash
go mod download
go mod vendor
make build
```

## Authentication Issues

### "Could not create a connection to OVH API"

**Full error:**
```
Could not create a connection to OVH API. You may want to visit: 
https://github.com/sneederco/docker-machine-driver-ovh#example-usage.
```

**Causes & Solutions:**

#### 1. Missing credentials

Check your configuration:

```bash
# Check environment variables
echo $OVH_APPLICATION_KEY
echo $OVH_APPLICATION_SECRET
echo $OVH_CONSUMER_KEY

# Or check config file
cat ~/.ovh.conf
```

**Fix:** Create `~/.ovh.conf`:

```ini
[default]
endpoint=ovh-eu

[ovh-eu]
application_key=YOUR_KEY
application_secret=YOUR_SECRET
consumer_key=YOUR_CONSUMER
```

#### 2. Wrong endpoint

**Error symptoms:** API calls fail, authentication errors

**Fix:** Ensure endpoint matches your OVH region:

```bash
docker-machine create -d ovh \
  --ovh-endpoint ovh-eu \
  my-node
```

**Available endpoints:**
- `ovh-eu` — Most common (Europe)
- `ovh-ca` — Canada
- `ovh-us` — United States

#### 3. Invalid API keys

**Test your keys:**

```bash
curl -X GET \
  "https://eu.api.ovh.com/1.0/me" \
  -H "X-Ovh-Application: $OVH_APPLICATION_KEY" \
  -H "X-Ovh-Timestamp: $(date +%s)" \
  -H "X-Ovh-Consumer: $OVH_CONSUMER_KEY"
```

**Fix:** Regenerate keys at https://api.ovh.com/createToken/

#### 4. Insufficient permissions

**Error:** API calls return 403 Forbidden

**Fix:** Ensure API keys have required permissions:
- `GET /*`
- `POST /*`
- `PUT /*`
- `DELETE /*`

### "Invalid signature" errors

**Cause:** System clock is out of sync.

**Solution:**

```bash
# Linux
sudo ntpdate pool.ntp.org

# macOS
sudo sntp -sS time.apple.com
```

### "Consumer key is not valid"

**Cause:** Consumer key has expired or been revoked.

**Solution:**

1. Go to https://api.ovh.com/createToken/
2. Generate new keys
3. Update your `~/.ovh.conf`

## Instance Creation Issues

### "Project not found" or "You must specify a project"

**Cause:** You have multiple OVH projects and didn't specify which to use.

**Solution:**

```bash
# List your projects
# (You'll need to use OVH API or console)

# Specify project
docker-machine create -d ovh \
  --ovh-project "my-project-name" \
  my-node
```

**Find project ID:**
1. Go to https://www.ovh.com/manager/cloud/
2. Select your project
3. Project ID is in URL: `#/iaas/pci/project/{PROJECT_ID}`

### "Flavor not found"

**Error:**
```
Error creating machine: Error in driver during machine creation: 
Flavor 'invalid-flavor' not found
```

**Cause:** Specified flavor doesn't exist or isn't available in selected region.

**Solution:**

```bash
# Use a valid flavor
docker-machine create -d ovh \
  --ovh-flavor vps-ssd-1 \
  my-node
```

**Common valid flavors:**
- `vps-ssd-1`, `vps-ssd-2`, `vps-ssd-3`
- `b2-7`, `b2-15`, `b2-30`, `b2-60`, `b2-120`

**Check available flavors in OVH console:**
https://www.ovh.com/manager/cloud/ → Your Project → Instances

### "Image not found"

**Error:**
```
Error creating machine: Error in driver during machine creation: 
Image 'invalid-image' not found
```

**Cause:** Image name doesn't match exactly or isn't available in region.

**Solution:**

Use exact image name from OVH:

```bash
docker-machine create -d ovh \
  --ovh-image "Ubuntu 22.04" \
  my-node
```

**Common valid images:**
- `Ubuntu 22.04`
- `Ubuntu 20.04`
- `Debian 12`
- `Debian 11`

**Note:** Image names may vary slightly by region. Check OVH console.

### "Quota exceeded"

**Error:**
```
Error creating machine: Error in driver during machine creation: 
Quota exceeded for resource 'instances'
```

**Cause:** Your OVH project has reached its instance quota.

**Solutions:**

1. **Delete unused instances** in OVH console
2. **Request quota increase:**
   - Go to: https://www.ovh.com/manager/cloud/
   - Your Project → Quota and Regions
   - Click **Increase quota**

3. **Use a different region** (quotas are per-region):
   ```bash
   docker-machine create -d ovh --ovh-region SBG5 my-node
   ```

### "Region not available"

**Cause:** Specified region doesn't exist or your project isn't activated for it.

**Solution:**

1. **Use a valid region:**
   ```bash
   docker-machine create -d ovh --ovh-region GRA1 my-node
   ```

2. **Activate region in your project:**
   - Go to: https://www.ovh.com/manager/cloud/
   - Your Project → Quota and Regions
   - Enable desired region

**Valid regions:** GRA1, GRA3, GRA5, SBG1, SBG3, SBG5, BHS1, BHS3, BHS5, DE1, UK1, WAW1, SGP1, SYD1

## SSH Connection Issues

### "Waiting for SSH to be available..." (hangs)

**Cause:** SSH isn't accessible on the instance.

**Troubleshoot:**

1. **Check instance status in OVH console:**
   - Should show "Active"
   - Should have a public IP

2. **Test SSH manually:**
   ```bash
   ssh ubuntu@<instance-ip>
   ```

3. **Check security groups:**
   - Ensure port 22 is open
   - Check OVH firewall settings

4. **Check cloud-init logs:**
   ```bash
   # If you can SSH in
   sudo journalctl -u cloud-init
   ```

**Solutions:**

- Wait longer (cloud-init can take 2-3 minutes)
- Check OVH console for instance errors
- Recreate the instance if stuck

### "Permission denied (publickey)"

**Cause:** SSH key mismatch or incorrect SSH user.

**Solutions:**

#### 1. Wrong SSH user

Different images use different default users:

```bash
# Ubuntu
docker-machine create -d ovh --ovh-ssh-user ubuntu my-node

# Debian
docker-machine create -d ovh --ovh-ssh-user debian --ovh-image "Debian 12" my-node

# CoreOS
docker-machine create -d ovh --ovh-ssh-user core --ovh-image "CoreOS stable" my-node
```

#### 2. SSH key not found

If using `--ovh-ssh-key`:

```bash
# Key must exist in OVH project AND be in ~/.ssh/
ls -la ~/.ssh/

# Make sure private key has correct permissions
chmod 600 ~/.ssh/id_rsa
```

### "Host key verification failed"

**Cause:** SSH host key changed (common after recreating instances).

**Solution:**

```bash
# Remove old host key
ssh-keygen -R <instance-ip>

# Or edit ~/.ssh/known_hosts and remove the line
```

### Cannot SSH after machine is created

**Cause:** Security group or firewall blocking SSH.

**Solution:**

1. **Check OVH console security groups**
2. **Test with verbose SSH:**
   ```bash
   ssh -vvv ubuntu@<instance-ip>
   ```
3. **Try from different network** (to rule out local firewall)

## Network Issues

### Private network interface not working

**Problem:** Instance created with `--ovh-private-network` but interface isn't active.

**Cause:** Private interface requires manual configuration.

**Solution:**

```bash
docker-machine ssh my-node

# Configure interface
sudo tee /etc/network/interfaces.d/99-vrack.cfg << 'EOF'
auto ens4
iface ens4 inet dhcp
EOF

# Enable interface
sudo ifup ens4

# Verify
ip addr show ens4
```

**For persistence across reboots**, add to cloud-init or startup script.

### "Network not found" error

**Cause:** Specified VLAN doesn't exist in your vRack.

**Solution:**

1. **Check vRack setup:**
   - Go to: https://www.ovh.com/manager/cloud/index.html#/vrack
   - Verify project is attached
   - Check private networks exist

2. **Create VLAN:**
   - Your Project → Private Networks
   - Create network with specific VLAN ID

3. **Use correct VLAN ID:**
   ```bash
   docker-machine create -d ovh --ovh-private-network 42 my-node
   ```

### Docker daemon not accessible

**Error:**
```
Cannot connect to the Docker daemon at tcp://<ip>:2376. 
Is the docker daemon running?
```

**Solutions:**

```bash
# Regenerate certificates
docker-machine regenerate-certs my-node

# Restart Docker on the machine
docker-machine ssh my-node sudo systemctl restart docker

# Check Docker status
docker-machine ssh my-node sudo systemctl status docker
```

## Rancher Integration Issues

### Driver not appearing in Rancher

**Problem:** OVH driver not visible in Rancher node drivers list.

**Solutions:**

1. **Check download URL** is correct in driver configuration
2. **Verify Rancher can reach GitHub:**
   ```bash
   kubectl exec -n cattle-system $(kubectl get pods -n cattle-system -l app=rancher -o name | head -1) -- curl -I https://github.com
   ```
3. **Check Rancher logs:**
   ```bash
   kubectl logs -n cattle-system -l app=rancher --tail=100
   ```
4. **Re-add driver** with correct URL

### Node provisioning fails in Rancher

**Problem:** Nodes stuck in "Provisioning" state.

**Solutions:**

1. **Check node logs** in Rancher UI
2. **Verify credentials** in node template
3. **Check OVH quotas** (common cause)
4. **Try creating manually** with docker-machine to isolate issue:
   ```bash
   docker-machine create -d ovh \
     --ovh-application-key "..." \
     --ovh-application-secret "..." \
     --ovh-consumer-key "..." \
     --ovh-region GRA1 \
     test-node
   ```

### Nodes active but cluster won't provision

**Problem:** Nodes show "Active" but cluster stays in "Provisioning".

**Solutions:**

1. **Check Rancher agent logs** on nodes:
   ```bash
   docker-machine ssh node docker logs rancher-agent
   ```

2. **Verify network connectivity:**
   - Nodes can reach Rancher server
   - Security groups allow required ports
   - DNS resolution works

3. **Check Rancher server logs:**
   ```bash
   kubectl logs -n cattle-system -l app=rancher
   ```

## Managed Kubernetes (MKS) Issues

### "MKS cluster name is required"

**Cause:** Using `--ovh-hosted-mks` without specifying cluster name.

**Solution:**

```bash
docker-machine create -d ovh \
  --ovh-hosted-mks \
  --ovh-mks-cluster-name my-cluster \
  --ovh-mks-nodepool-flavor b2-7 \
  --ovh-mks-nodepool-size 3 \
  k8s-cluster
```

### Cannot connect to MKS cluster

**Problem:** Cluster created but can't access with kubectl.

**Solution:**

1. **Get kubeconfig from OVH console:**
   - Go to: https://www.ovh.com/manager/cloud/
   - Your Project → Managed Kubernetes
   - Click cluster → Download kubeconfig

2. **Use kubeconfig:**
   ```bash
   export KUBECONFIG=~/Downloads/kubeconfig.yml
   kubectl get nodes
   ```

### MKS cluster stuck in creating state

**Cause:** OVH cluster provisioning can take 5-10 minutes.

**Solution:**

1. **Wait** — first-time setup is slow
2. **Check OVH console** for errors
3. **Verify region has MKS available**

## Performance Issues

### Slow instance creation

**Cause:** Normal for OVH provisioning.

**Expected times:**
- Instance creation: 30-60 seconds
- SSH availability: 1-2 minutes
- Cloud-init completion: 2-3 minutes
- Total: 3-5 minutes

**If longer:**
- Check OVH status page: https://status.ovhcloud.com/
- Try different region
- Check your internet connection

### Docker commands slow on created machine

**Cause:** Instance undersized for workload.

**Solution:**

Use larger flavor:

```bash
docker-machine create -d ovh \
  --ovh-flavor b2-15 \
  larger-node
```

## Debugging Tips

### Enable debug mode

```bash
docker-machine -D create -d ovh my-node
```

### Check driver logs

```bash
docker-machine -D create -d ovh my-node 2>&1 | tee debug.log
```

### Test OVH API directly

```bash
# Test authentication
curl "https://eu.api.ovh.com/1.0/me" \
  -H "X-Ovh-Application: $OVH_APPLICATION_KEY"

# List projects
curl "https://eu.api.ovh.com/1.0/cloud/project" \
  -H "X-Ovh-Application: $OVH_APPLICATION_KEY"
```

### Inspect machine state

```bash
# Check machine status
docker-machine status my-node

# Get machine details
docker-machine inspect my-node

# View machine config
cat ~/.docker/machine/machines/my-node/config.json
```

### Check instance in OVH console

1. Go to: https://www.ovh.com/manager/cloud/
2. Your Project → Instances
3. Find your instance by name
4. Check logs and console

## Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| `connection timeout` | Network/firewall issue | Check security groups, try different network |
| `invalid grant` | Expired consumer key | Regenerate API keys |
| `insufficient quota` | Project quota exceeded | Delete instances or request increase |
| `not found` | Resource doesn't exist | Verify names/IDs in OVH console |
| `already exists` | Name conflict | Use different machine name |
| `unauthorized` | Invalid credentials | Check API keys |

## Getting Help

### Before asking for help

Gather this information:

1. **Driver version:**
   ```bash
   docker-machine-driver-ovh --version
   ```

2. **Docker Machine version:**
   ```bash
   docker-machine --version
   ```

3. **System info:**
   ```bash
   uname -a
   ```

4. **Debug logs:**
   ```bash
   docker-machine -D create -d ovh my-node 2>&1 | tee debug.log
   ```

5. **Command used** (redact sensitive info)

### Where to get help

- **GitHub Issues:** https://github.com/sneederco/docker-machine-driver-ovh/issues
- **GitHub Discussions:** https://github.com/sneederco/docker-machine-driver-ovh/discussions
- **OVH Support:** https://help.ovhcloud.com/

### Filing a bug report

Include:
- Description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Debug logs (with sensitive data redacted)
- System information
- Screenshots (if applicable)

## Additional Resources

- [Installation Guide](installation.md)
- [Configuration Reference](configuration.md)
- [Rancher Integration](rancher-integration.md)
- [OVH API Documentation](https://api.ovh.com/)
- [Docker Machine Documentation](https://docs.docker.com/machine/)

---

[← Rancher Integration](rancher-integration.md) | [Back to README](../README.md) | [Architecture →](architecture.md)
