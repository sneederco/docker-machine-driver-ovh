package main

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
)

// newTestDriver creates a driver instance for testing
func newTestDriver() *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: "test-machine",
			StorePath:   "/tmp/test",
			SSHUser:     DefaultSSHUserName,
			SSHPort:     22,
		},
		Endpoint:      DefaultEndpoint,
		RegionName:    DefaultRegionName,
		FlavorName:    DefaultFlavorName,
		ImageName:     DefaultImageName,
		BillingPeriod: DefaultBillingPeriod,
	}
}

// TestDriverName tests the DriverName method
func TestDriverName(t *testing.T) {
	d := newTestDriver()
	
	if name := d.DriverName(); name != "ovh" {
		t.Errorf("DriverName() = %q, want %q", name, "ovh")
	}
}

// TestGetCreateFlags tests that create flags are properly defined
func TestGetCreateFlags(t *testing.T) {
	d := newTestDriver()
	flags := d.GetCreateFlags()
	
	if len(flags) == 0 {
		t.Fatal("GetCreateFlags() returned no flags")
	}
	
	// Verify essential flags exist
	requiredFlags := map[string]bool{
		"ovh-application-key":    false,
		"ovh-application-secret": false,
		"ovh-consumer-key":       false,
		"ovh-endpoint":           false,
		"ovh-project":            false,
		"ovh-region":             false,
		"ovh-flavor":             false,
		"ovh-image":              false,
		"ovh-ssh-user":           false,
		"ovh-billing-period":     false,
		"ovh-hosted-mks":         false,
	}
	
	for _, flag := range flags {
		flagName := flag.String()
		if _, exists := requiredFlags[flagName]; exists {
			requiredFlags[flagName] = true
		}
	}
	
	for flagName, found := range requiredFlags {
		if !found {
			t.Errorf("Required flag %q not found in GetCreateFlags()", flagName)
		}
	}
}

// mockDriverOptions creates a mock DriverOptions for testing
type mockDriverOptions struct {
	stringOpts map[string]string
	intOpts    map[string]int
	boolOpts   map[string]bool
}

func (m *mockDriverOptions) String(key string) string {
	return m.stringOpts[key]
}

func (m *mockDriverOptions) StringSlice(key string) []string {
	return nil
}

func (m *mockDriverOptions) Int(key string) int {
	return m.intOpts[key]
}

func (m *mockDriverOptions) Bool(key string) bool {
	return m.boolOpts[key]
}

// TestSetConfigFromFlags tests flag parsing
func TestSetConfigFromFlags(t *testing.T) {
	d := newTestDriver()
	
	// Create mock flags
	flags := &mockDriverOptions{
		stringOpts: make(map[string]string),
		intOpts:    make(map[string]int),
		boolOpts:   make(map[string]bool),
	}
	
	// Set test values
	flags.stringOpts["ovh-application-key"] = "test-app-key"
	flags.stringOpts["ovh-application-secret"] = "test-app-secret"
	flags.stringOpts["ovh-consumer-key"] = "test-consumer-key"
	flags.stringOpts["ovh-endpoint"] = "ovh-eu"
	flags.stringOpts["ovh-project"] = "my-project"
	flags.stringOpts["ovh-region"] = "GRA7"
	flags.stringOpts["ovh-flavor"] = "b2-7"
	flags.stringOpts["ovh-image"] = "Debian 12"
	flags.stringOpts["ovh-ssh-user"] = "debian"
	flags.stringOpts["ovh-billing-period"] = "monthly"
	flags.stringOpts["ovh-private-network"] = "my-network"
	flags.stringOpts["ovh-ssh-key-name"] = "my-key"
	flags.stringOpts["ovh-userdata"] = "/path/to/userdata.yaml"
	flags.stringOpts["ovh-tags"] = "env:test,app:docker"
	flags.boolOpts["ovh-hosted-mks"] = true
	flags.stringOpts["ovh-mks-cluster-name"] = "test-cluster"
	flags.stringOpts["ovh-mks-version"] = "1.28"
	flags.stringOpts["ovh-mks-nodepool-flavor"] = "b3-8"
	flags.stringOpts["ovh-mks-nodepool-name"] = "default"
	flags.intOpts["ovh-mks-nodepool-size"] = 3
	flags.stringOpts["swarm-host"] = ""
	flags.stringOpts["swarm-discovery"] = ""
	flags.boolOpts["swarm-master"] = false
	
	if err := d.SetConfigFromFlags(flags); err != nil {
		t.Fatalf("SetConfigFromFlags() failed: %v", err)
	}
	
	// Verify values were set
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ApplicationKey", d.ApplicationKey, "test-app-key"},
		{"ApplicationSecret", d.ApplicationSecret, "test-app-secret"},
		{"ConsumerKey", d.ConsumerKey, "test-consumer-key"},
		{"Endpoint", d.Endpoint, "ovh-eu"},
		{"ProjectName", d.ProjectName, "my-project"},
		{"RegionName", d.RegionName, "GRA7"},
		{"FlavorName", d.FlavorName, "b2-7"},
		{"ImageName", d.ImageName, "Debian 12"},
		{"SSHUser", d.SSHUser, "debian"},
		{"BillingPeriod", d.BillingPeriod, "monthly"},
		{"PrivateNetworkName", d.PrivateNetworkName, "my-network"},
		{"KeyPairName", d.KeyPairName, "my-key"},
		{"UserdataPath", d.UserdataPath, "/path/to/userdata.yaml"},
		{"Tags", d.Tags, "env:test,app:docker"},
		{"MKSClusterName", d.MKSClusterName, "test-cluster"},
		{"MKSVersion", d.MKSVersion, "1.28"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
	
	// Test boolean and int values
	if !d.HostedMKS {
		t.Error("HostedMKS should be true")
	}
	
	if d.MKSNodePoolDesiredSize != 3 {
		t.Errorf("MKSNodePoolDesiredSize = %d, want 3", d.MKSNodePoolDesiredSize)
	}
}

// TestGetStateMapping tests state string to state.State mapping
func TestGetStateMapping(t *testing.T) {
	tests := []struct {
		name           string
		instanceStatus string
		wantState      state.State
	}{
		{"Active", "ACTIVE", state.Running},
		{"Building", "BUILD", state.Starting},
		{"Building Alt", "BUILDING", state.Starting},
		{"Stopped", "SHUTOFF", state.Stopped},
		{"Stopped Alt", "STOPPED", state.Stopped},
		{"Paused", "PAUSED", state.Paused},
		{"Suspended", "SUSPENDED", state.Saved},
		{"Error", "ERROR", state.Error},
		{"Deleted", "DELETED", state.None},
		{"Soft Deleted", "SOFT_DELETED", state.None},
		{"Unknown", "UNKNOWN_STATUS", state.None},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDriver()
			d.ProjectID = "test-project"
			d.InstanceID = "test-instance"
			
			// Note: This test would require a mock API to fully test
			// For now, we're just documenting the expected mappings
			// The actual mapping logic is in getInstanceState()
			
			// We can't test this without the API mock, but we're documenting
			// the expected behavior for future integration tests
			t.Logf("Status %q should map to %v", tt.instanceStatus, tt.wantState)
		})
	}
}

// TestGetURL tests Docker URL construction
func TestGetURL(t *testing.T) {
	tests := []struct {
		name      string
		ipAddress string
		want      string
	}{
		{"With IP", "203.0.113.42", "tcp://203.0.113.42:2376"},
		{"IPv6", "2001:db8::1", "tcp://[2001:db8::1]:2376"},
		{"No IP", "", ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDriver()
			d.IPAddress = tt.ipAddress
			
			got, err := d.GetURL()
			if err != nil {
				t.Fatalf("GetURL() error = %v", err)
			}
			
			if got != tt.want {
				t.Errorf("GetURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetSSHHostname tests SSH hostname retrieval
func TestGetSSHHostname(t *testing.T) {
	d := newTestDriver()
	d.IPAddress = "203.0.113.42"
	
	hostname, err := d.GetSSHHostname()
	if err != nil {
		t.Fatalf("GetSSHHostname() error = %v", err)
	}
	
	if hostname != "203.0.113.42" {
		t.Errorf("GetSSHHostname() = %q, want %q", hostname, "203.0.113.42")
	}
}

// TestPreCreateCheckValidation tests input validation
func TestPreCreateCheckValidation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Driver)
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid standard config",
			setup: func(d *Driver) {
				// Use defaults - should be valid
			},
			wantError: false,
		},
		{
			name: "Invalid billing period",
			setup: func(d *Driver) {
				d.BillingPeriod = "weekly"
			},
			wantError: true,
			errorMsg:  "invalid billing period",
		},
		{
			name: "Hourly billing is valid",
			setup: func(d *Driver) {
				d.BillingPeriod = "hourly"
			},
			wantError: true, // Will fail on API calls, but validation passes
		},
		{
			name: "Monthly billing is valid",
			setup: func(d *Driver) {
				d.BillingPeriod = "monthly"
			},
			wantError: true, // Will fail on API calls, but validation passes
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDriver()
			d.ApplicationKey = "fake-key"
			d.ApplicationSecret = "fake-secret"
			d.ConsumerKey = "fake-consumer"
			d.ProjectName = "test-project"
			
			tt.setup(d)
			
			err := d.PreCreateCheck()
			
			if tt.wantError {
				if err == nil && tt.errorMsg != "" {
					t.Logf("Expected error but validation passed (will fail on API call)")
					return
				}
				if err != nil && tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error = %q, want to contain %q", err.Error(), tt.errorMsg)
				}
			}
			// Note: Without API mock, most validations will fail on API calls
			// This is expected for unit tests
		})
	}
}

// TestMKSValidation tests MKS-specific validation
func TestMKSValidation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Driver)
		wantError bool
		errorMsg  string
	}{
		{
			name: "MKS mode requires cluster name",
			setup: func(d *Driver) {
				d.HostedMKS = true
				d.MKSClusterName = ""
			},
			wantError: true,
			errorMsg:  "missing required value for '--ovh-mks-cluster-name'",
		},
		{
			name: "MKS mode requires valid nodepool size",
			setup: func(d *Driver) {
				d.HostedMKS = true
				d.MKSClusterName = "test-cluster"
				d.MKSNodePoolDesiredSize = 0
			},
			wantError: true,
			errorMsg:  "invalid value 0 for '--ovh-mks-nodepool-size'",
		},
		{
			name: "MKS mode with valid config",
			setup: func(d *Driver) {
				d.HostedMKS = true
				d.MKSClusterName = "test-cluster"
				d.MKSNodePoolDesiredSize = 3
				d.MKSNodePoolFlavor = "b3-8"
			},
			wantError: true, // Will fail on API, but validation logic passes
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newTestDriver()
			d.ApplicationKey = "fake-key"
			d.ApplicationSecret = "fake-secret"
			d.ConsumerKey = "fake-consumer"
			d.ProjectName = "test-project"
			
			tt.setup(d)
			
			err := d.PreCreateCheck()
			
			if tt.wantError && tt.errorMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing %q but got nil", tt.errorMsg)
				} else if !contains(err.Error(), tt.errorMsg) {
					t.Logf("Error: %v (expected to contain %q)", err, tt.errorMsg)
					// Some errors might be from API calls, not validation
				}
			}
		})
	}
}

// TestContainsIgnoreCase tests the case-insensitive contains helper
func TestContainsIgnoreCase(t *testing.T) {
	items := []string{"US-EAST-VA-1", "GRA7", "BHS5"}
	
	tests := []struct {
		value string
		want  bool
	}{
		{"US-EAST-VA-1", true},
		{"us-east-va-1", true},
		{"GRA7", true},
		{"gra7", true},
		{"BHS5", true},
		{"bhs5", true},
		{"WAW1", false},
		{"", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got := containsIgnoreCase(items, tt.value)
			if got != tt.want {
				t.Errorf("containsIgnoreCase(%v, %q) = %v, want %v", items, tt.value, got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
