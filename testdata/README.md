# Test Data

This directory contains sample OVH API responses and test fixtures for the docker-machine-driver-ovh test suite.

## Files

- **project_response.json** - Sample OVH Cloud project response
- **regions_response.json** - List of available regions
- **flavors_response.json** - List of available VM flavors
- **images_response.json** - List of available OS images
- **instance_response.json** - Sample instance creation response
- **mks_cluster_response.json** - Sample MKS cluster response
- **cloud-init-userdata.yaml** - Sample cloud-init configuration for testing userdata

## Usage

These files are used by the test suite to verify JSON parsing, response handling, and to provide realistic test data during development.

To use in tests:

```go
data, err := os.ReadFile("testdata/project_response.json")
if err != nil {
    t.Fatal(err)
}

var project Project
if err := json.Unmarshal(data, &project); err != nil {
    t.Fatal(err)
}
```
