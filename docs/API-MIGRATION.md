# API Migration Guide

## Overview

The `api.go` file has been rewritten to use modern Go patterns. This guide explains the changes and how to migrate existing code.

## Key Changes

### 1. Context Support

All API methods now require a `context.Context` parameter:

**Old:**
```go
projects, err := client.GetProjects()
project, err := client.GetProject(projectID)
```

**New:**
```go
projects, err := client.GetProjects(ctx)
project, err := client.GetProject(ctx, projectID)
```

### 2. Structured Error Types

Errors are now wrapped in `APIError` with additional context:

```go
err := client.GetProject(ctx, "invalid-id")
if apiErr, ok := err.(*APIError); ok {
    if apiErr.IsNotFound() {
        // Handle 404
    }
    if apiErr.IsRetryable() {
        // Error is retryable (5xx, 429)
    }
}
```

### 3. Retry Logic

Automatic retry with exponential backoff for:
- Rate limits (429)
- Server errors (5xx)

Configure via `APIConfig`:
```go
config := &APIConfig{
    MaxRetries:    3,
    RetryDelay:    1 * time.Second,
    RateLimitWait: 100 * time.Millisecond,
    Logger:        logrus.New(),
}
api, err := NewAPIWithConfig(endpoint, appKey, appSecret, consumerKey, config)
```

### 4. Comprehensive Logging

All operations are logged with structured fields:
```go
logger := logrus.New()
logger.SetLevel(logrus.DebugLevel)
config := &APIConfig{Logger: logger}
```

### 5. New API Methods

#### ListRegions
Returns detailed region information (not just names):
```go
regions, err := api.ListRegions(ctx, projectID)
for _, region := range regions {
    fmt.Printf("%s: %s (%s)\n", region.Name, region.Status, region.Continent)
}
```

Legacy method still available:
```go
names, err := api.GetRegions(ctx, projectID) // Returns []string
```

#### ListFlavors
```go
flavors, err := api.ListFlavors(ctx, projectID, "GRA7")
```

#### ListImages
```go
images, err := api.ListImages(ctx, projectID, "GRA7")
```

#### ListSSHKeys
```go
keys, err := api.ListSSHKeys(ctx, projectID, "GRA7")
```

#### CreateSSHKey
```go
key, err := api.CreateSSHKey(ctx, projectID, "my-key", publicKeyContent)
```

#### GetQuotas
```go
quota, err := api.GetQuotas(ctx, projectID, "GRA7")
fmt.Printf("Available instances: %d\n", quota.Instance)
fmt.Printf("Available cores: %d\n", quota.Cores)
```

## Migration Steps for driver.go

### Step 1: Import context
```go
import (
    "context"
    // ...
)
```

### Step 2: Create a context for API calls

For most driver methods, use `context.Background()`:
```go
ctx := context.Background()
```

For operations that should be cancellable (e.g., Create, Remove), consider passing context through:
```go
func (d *Driver) Create() error {
    ctx := context.Background()
    // or use a context with timeout:
    // ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    // defer cancel()
    
    // Use ctx in all API calls
    project, err := d.client.GetProject(ctx, d.ProjectID)
    // ...
}
```

### Step 3: Update all API method calls

Replace all calls to add `ctx` as the first parameter:

```go
// PreCreate validation
project, err := client.GetProjectByName(ctx, d.ProjectName)
projects, err := client.GetProjects(ctx)
project, err := client.GetProject(ctx, projectID)

// Region validation
regions, err := client.GetRegions(ctx, d.ProjectID)

// Flavor lookup
flavor, err := client.GetFlavorByName(ctx, d.ProjectID, d.RegionName, d.FlavorName)
flavors, err := client.GetFlavors(ctx, d.ProjectID, d.RegionName)

// Image lookup
image, err := client.GetImageByName(ctx, d.ProjectID, d.RegionName, d.ImageName)

// SSH key operations
sshkey, err := client.GetSshkeyByName(ctx, d.ProjectID, d.RegionName, d.SSHKeyName)
sshkey, err := client.CreateSshkey(ctx, d.ProjectID, keyName, string(publicKey))
err := client.DeleteSshkey(ctx, d.ProjectID, d.SSHKeyID)

// Network operations
networks, err := client.GetNetworks(ctx, d.ProjectID, false)
network, err := client.GetPrivateNetworkByName(ctx, d.ProjectID, d.NetworkName)

// Instance operations
instance, err := client.CreateInstance(ctx, d.ProjectID, d.MachineName, ...)
instance, err := client.GetInstance(ctx, d.ProjectID, d.InstanceID)
err := client.StartInstance(ctx, d.ProjectID, d.InstanceID)
err := client.StopInstance(ctx, d.ProjectID, d.InstanceID)
err := client.RebootInstance(ctx, d.ProjectID, d.InstanceID, false)
err := client.DeleteInstance(ctx, d.ProjectID, d.InstanceID)

// MKS operations
clusters, err := client.ListMKSClusters(ctx, d.ProjectID)
cluster, err := client.CreateMKSCluster(ctx, d.ProjectID, req)
err := client.DeleteMKSCluster(ctx, d.ProjectID, clusterID)
nodePool, err := client.CreateMKSNodePool(ctx, d.ProjectID, clusterID, req)
err := client.ScaleMKSNodePool(ctx, d.ProjectID, clusterID, nodePoolID, desiredNodes)
```

### Step 4: Enhanced error handling (optional)

Take advantage of structured errors:

```go
instance, err := client.GetInstance(ctx, d.ProjectID, d.InstanceID)
if err != nil {
    if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
        return ErrInstanceNotFound
    }
    return fmt.Errorf("failed to get instance: %w", err)
}
```

## Type Renames

Some types have been renamed for consistency:

| Old Name | New Name |
|----------|----------|
| `Sshkey` | `SSHKey` |
| `Sshkeys` | `SSHKeys` |
| `SshkeyReq` | `SSHKeyReq` |

The old names are still available via type aliases for backward compatibility in legacy methods.

## Testing

Unit tests with mocked HTTP responses are provided in `api_test.go`. To run tests:

```bash
go test -v ./...
```

Note: Most tests are currently skipped as they require proper OVH client mocking. To enable full testing, integrate with a mocking library like `httptest` or `gomock`.

## Performance

The new implementation includes:
- **Retry logic**: Automatic retry on transient failures
- **Rate limiting**: Built-in backoff for rate limits
- **Logging**: Structured logging for debugging
- **Context support**: Proper cancellation and timeout handling

## Backward Compatibility

Legacy method signatures are preserved where possible:
- `GetRegions()` returns `[]string` (use `ListRegions()` for detailed info)
- `GetFlavors()`, `GetImages()`, `GetSshkeys()` still work
- `CreateSshkey()` still works (use `CreateSSHKey()` for new code)

## Example: Complete Migration

**Before:**
```go
func (d *Driver) validateConfig() error {
    project, err := d.client.GetProjectByName(d.ProjectName)
    if err != nil {
        return err
    }
    d.ProjectID = project.ID
    
    regions, err := d.client.GetRegions(d.ProjectID)
    if err != nil {
        return err
    }
    // ...
}
```

**After:**
```go
func (d *Driver) validateConfig() error {
    ctx := context.Background()
    
    project, err := d.client.GetProjectByName(ctx, d.ProjectName)
    if err != nil {
        return err
    }
    d.ProjectID = project.ID
    
    regions, err := d.client.GetRegions(ctx, d.ProjectID)
    if err != nil {
        return err
    }
    // ...
}
```

## Questions?

See the full API documentation in `api.go` or consult the unit tests in `api_test.go` for usage examples.
