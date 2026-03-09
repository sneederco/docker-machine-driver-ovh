//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/machine/libmachine/drivers"
)

// Integration tests that require real OVH API credentials
// Run with: go test -v -tags=integration

// getTestCredentials loads OVH credentials from environment variables
func getTestCredentials(t *testing.T) (appKey, appSecret, consumerKey, endpoint string) {
	t.Helper()
	
	appKey = os.Getenv("OVH_APPLICATION_KEY")
	appSecret = os.Getenv("OVH_APPLICATION_SECRET")
	consumerKey = os.Getenv("OVH_CONSUMER_KEY")
	endpoint = os.Getenv("OVH_ENDPOINT")
	
	if appKey == "" || appSecret == "" || consumerKey == "" {
		t.Skip("Skipping integration test: OVH credentials not set. Set OVH_APPLICATION_KEY, OVH_APPLICATION_SECRET, and OVH_CONSUMER_KEY environment variables.")
	}
	
	if endpoint == "" {
		endpoint = "ovh-us"
	}
	
	return
}

// TestIntegrationCreateDeleteVM tests the full lifecycle of a VM
func TestIntegrationCreateDeleteVM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	appKey, appSecret, consumerKey, endpoint := getTestCredentials(t)
	
	// Get test configuration from environment or use defaults
	projectName := os.Getenv("OVH_TEST_PROJECT")
	if projectName == "" {
		t.Skip("Skipping integration test: OVH_TEST_PROJECT not set")
	}
	
	regionName := os.Getenv("OVH_TEST_REGION")
	if regionName == "" {
		regionName = "US-EAST-VA-1"
	}
	
	flavorName := os.Getenv("OVH_TEST_FLAVOR")
	if flavorName == "" {
		flavorName = "b3-8"
	}
	
	imageName := os.Getenv("OVH_TEST_IMAGE")
	if imageName == "" {
		imageName = "Ubuntu 24.04"
	}
	
	// Create temp directory for driver
	tmpDir, err := os.MkdirTemp("", "driver-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create driver
	machineName := fmt.Sprintf("test-vm-%d", time.Now().Unix())
	driver := &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineName,
			StorePath:   tmpDir,
			SSHUser:     "ubuntu",
		},
		ApplicationKey:    appKey,
		ApplicationSecret: appSecret,
		ConsumerKey:       consumerKey,
		Endpoint:          endpoint,
		ProjectName:       projectName,
		RegionName:        regionName,
		FlavorName:        flavorName,
		ImageName:         imageName,
		BillingPeriod:     "hourly",
	}
	
	t.Logf("Testing VM creation: %s", machineName)
	t.Logf("  Project: %s", projectName)
	t.Logf("  Region: %s", regionName)
	t.Logf("  Flavor: %s", flavorName)
	t.Logf("  Image: %s", imageName)
	
	// Run PreCreateCheck
	t.Log("Running PreCreateCheck...")
	if err := driver.PreCreateCheck(); err != nil {
		t.Fatalf("PreCreateCheck failed: %v", err)
	}
	
	t.Logf("  Project ID: %s", driver.ProjectID)
	t.Logf("  Flavor ID: %s", driver.FlavorID)
	t.Logf("  Image ID: %s", driver.ImageID)
	
	// Create the instance
	t.Log("Creating instance...")
	createStart := time.Now()
	if err := driver.Create(); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	createDuration := time.Since(createStart)
	
	t.Logf("Instance created in %v", createDuration)
	t.Logf("  Instance ID: %s", driver.InstanceID)
	t.Logf("  IP Address: %s", driver.IPAddress)
	t.Logf("  SSH Key ID: %s", driver.KeyPairID)
	
	// Ensure cleanup happens
	defer func() {
		t.Log("Cleaning up instance...")
		if err := driver.Remove(); err != nil {
			t.Errorf("Cleanup failed: %v", err)
		} else {
			t.Log("Instance removed successfully")
		}
	}()
	
	// Verify instance state
	t.Log("Checking instance state...")
	state, err := driver.GetState()
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	
	t.Logf("Instance state: %v", state)
	
	// Get instance URL
	url, err := driver.GetURL()
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}
	
	t.Logf("Docker URL: %s", url)
	
	// Test Stop
	t.Log("Testing Stop...")
	if err := driver.Stop(); err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	
	// Wait a bit for state to settle
	time.Sleep(10 * time.Second)
	
	state, _ = driver.GetState()
	t.Logf("State after Stop: %v", state)
	
	// Test Start
	t.Log("Testing Start...")
	if err := driver.Start(); err != nil {
		t.Errorf("Start failed: %v", err)
	}
	
	time.Sleep(10 * time.Second)
	
	state, _ = driver.GetState()
	t.Logf("State after Start: %v", state)
}

// TestIntegrationAPIConnectivity tests basic API connectivity
func TestIntegrationAPIConnectivity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	appKey, appSecret, consumerKey, endpoint := getTestCredentials(t)
	
	// Create API client
	client, err := NewAPI(endpoint, appKey, appSecret, consumerKey)
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}
	
	ctx := context.Background()
	
	// Test GetProjects
	t.Log("Testing GetProjects...")
	projects, err := client.GetProjects(ctx)
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}
	
	t.Logf("Found %d project(s)", len(projects))
	
	if len(projects) == 0 {
		t.Skip("No projects available for testing")
	}
	
	// Test GetProject
	projectID := projects[0]
	t.Logf("Testing GetProject for: %s", projectID)
	project, err := client.GetProject(ctx, projectID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	
	t.Logf("  Name: %s", project.Name)
	t.Logf("  Status: %s", project.Status)
	
	// Test GetRegions
	t.Log("Testing GetRegions...")
	regions, err := client.GetRegions(ctx, projectID)
	if err != nil {
		t.Fatalf("GetRegions failed: %v", err)
	}
	
	t.Logf("Found %d region(s): %v", len(regions), regions)
	
	if len(regions) == 0 {
		t.Skip("No regions available for testing")
	}
	
	testRegion := regions[0]
	
	// Test ListFlavors
	t.Logf("Testing ListFlavors for region: %s", testRegion)
	flavors, err := client.ListFlavors(ctx, projectID, testRegion)
	if err != nil {
		t.Fatalf("ListFlavors failed: %v", err)
	}
	
	t.Logf("Found %d flavor(s) in %s", len(flavors), testRegion)
	if len(flavors) > 0 {
		t.Logf("  Example: %s (%d vCPUs, %dGB RAM)", flavors[0].Name, flavors[0].Vcpus, flavors[0].MemoryGB)
	}
	
	// Test ListImages
	t.Logf("Testing ListImages for region: %s", testRegion)
	images, err := client.ListImages(ctx, projectID, testRegion)
	if err != nil {
		t.Fatalf("ListImages failed: %v", err)
	}
	
	t.Logf("Found %d image(s) in %s", len(images), testRegion)
	if len(images) > 0 {
		t.Logf("  Example: %s (%s)", images[0].Name, images[0].OS)
	}
	
	// Test ListSSHKeys
	t.Log("Testing ListSSHKeys...")
	sshkeys, err := client.ListSSHKeys(ctx, projectID, testRegion)
	if err != nil {
		t.Fatalf("ListSSHKeys failed: %v", err)
	}
	
	t.Logf("Found %d SSH key(s)", len(sshkeys))
}

// TestIntegrationQuotas tests quota retrieval
func TestIntegrationQuotas(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	appKey, appSecret, consumerKey, endpoint := getTestCredentials(t)
	
	client, err := NewAPI(endpoint, appKey, appSecret, consumerKey)
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}
	
	ctx := context.Background()
	
	projects, err := client.GetProjects(ctx)
	if err != nil || len(projects) == 0 {
		t.Skip("No projects available")
	}
	
	projectID := projects[0]
	
	regions, err := client.GetRegions(ctx, projectID)
	if err != nil || len(regions) == 0 {
		t.Skip("No regions available")
	}
	
	testRegion := regions[0]
	
	t.Logf("Getting quotas for project %s in region %s", projectID, testRegion)
	quota, err := client.GetQuotas(ctx, projectID, testRegion)
	if err != nil {
		t.Fatalf("GetQuotas failed: %v", err)
	}
	
	t.Logf("Quotas for %s:", testRegion)
	t.Logf("  Instances: %d", quota.Instance)
	t.Logf("  Cores: %d", quota.Cores)
	t.Logf("  RAM: %dMB", quota.RAM)
	t.Logf("  Volumes: %d", quota.Volumes)
	t.Logf("  Volume GB: %d", quota.VolumeGigabytes)
}

// TestIntegrationMKSClusters tests MKS cluster operations (list only, no creation)
func TestIntegrationMKSClusters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	appKey, appSecret, consumerKey, endpoint := getTestCredentials(t)
	
	client, err := NewAPI(endpoint, appKey, appSecret, consumerKey)
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}
	
	ctx := context.Background()
	
	projects, err := client.GetProjects(ctx)
	if err != nil || len(projects) == 0 {
		t.Skip("No projects available")
	}
	
	projectID := projects[0]
	
	t.Logf("Listing MKS clusters for project: %s", projectID)
	clusters, err := client.ListMKSClusters(ctx, projectID)
	if err != nil {
		t.Fatalf("ListMKSClusters failed: %v", err)
	}
	
	t.Logf("Found %d MKS cluster(s)", len(clusters))
	
	for i, cluster := range clusters {
		t.Logf("  Cluster %d:", i+1)
		t.Logf("    ID: %s", cluster.ID)
		t.Logf("    Name: %s", cluster.Name)
		t.Logf("    Region: %s", cluster.Region)
		t.Logf("    Version: %s", cluster.Version)
		t.Logf("    Status: %s", cluster.Status)
	}
}
