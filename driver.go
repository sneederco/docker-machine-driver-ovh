package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

const (
	statusTimeout         = 300 // 5 minutes for instance operations
	statusCheckInterval   = 5 * time.Second
	defaultContextTimeout = 15 * time.Minute
)

var (
	fallbackRegions = []string{"US-EAST-VA-1", "US-WEST-OR-1", "CA-EAST-BHS-1", "GRA1", "GRA3", "SBG1", "SBG5", "BHS5", "DE1", "UK1", "WAW1"}
	fallbackFlavors = []string{"b3-8", "b2-7", "b2-15", "vps-ssd-1", "vps-ssd-2"}
	fallbackImages  = []string{"Ubuntu 24.04", "Ubuntu 22.04", "Ubuntu 20.04", "Debian 12"}
)

// Driver is a machine driver for OVH Public Cloud.
type Driver struct {
	*drivers.BaseDriver

	// API credentials (overridable)
	ApplicationKey    string
	ApplicationSecret string
	ConsumerKey       string
	Endpoint          string

	// Required parameters
	ProjectName string
	ProjectID   string
	RegionName  string
	FlavorName  string
	ImageName   string

	// Optional parameters
	PrivateNetworkName string
	KeyPairName        string
	BillingPeriod      string
	UserdataPath       string
	Tags               string
	SecurityGroup      string

	// OpenStack credentials for security group attachment
	OpenstackAuthUrl   string
	OpenstackUsername  string
	OpenstackPassword  string

	// Hosted MKS mode parameters
	HostedMKS              bool
	MKSClusterName         string
	MKSVersion             string
	MKSNodePoolName        string
	MKSNodePoolFlavor      string
	MKSNodePoolDesiredSize int
	MKSClusterID           string
	MKSNodePoolID          string

	// Internal state
	FlavorID    string
	ImageID     string
	InstanceID  string
	KeyPairID   string
	NetworkIDs  []string
	Userdata    string

	// API client
	client *API
}

// GetCreateFlags registers the "machine create" flags recognized by this driver.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "OVH_APPLICATION_KEY",
			Name:   "ovh-application-key",
			Usage:  "OVH API application key. May be stored in ovh.conf",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OVH_APPLICATION_SECRET",
			Name:   "ovh-application-secret",
			Usage:  "OVH API application secret. May be stored in ovh.conf",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "OVH_CONSUMER_KEY",
			Name:   "ovh-consumer-key",
			Usage:  "OVH API consumer key. May be stored in ovh.conf",
			Value:  "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-endpoint",
			Usage: "OVH Cloud API endpoint (ovh-us, ovh-eu, ovh-ca). Default: ovh-us",
			Value: DefaultEndpoint,
		},
		mcnflag.StringFlag{
			Name:  "ovh-project",
			Usage: "OVH Cloud project name or id",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-region",
			Usage: "OVH Cloud region name. Default: US-EAST-VA-1",
			Value: DefaultRegionName,
		},
		mcnflag.StringFlag{
			Name:  "ovh-flavor",
			Usage: "OVH Cloud flavor name or id. Default: b3-8",
			Value: DefaultFlavorName,
		},
		mcnflag.StringFlag{
			Name:  "ovh-image",
			Usage: "OVH Cloud image name or id. Default: Ubuntu 24.04",
			Value: DefaultImageName,
		},
		mcnflag.StringFlag{
			Name:  "ovh-private-network",
			Usage: "OVH Cloud private network name or vlan number. Default: public network only",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-ssh-key-name",
			Usage: "OVH Cloud ssh key name or id to use. Default: generate a random name",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-ssh-user",
			Usage: "OVH Cloud ssh username to use. Default: ubuntu",
			Value: DefaultSSHUserName,
		},
		mcnflag.StringFlag{
			Name:  "ovh-billing-period",
			Usage: "OVH Cloud billing period (hourly or monthly). Default: hourly",
			Value: DefaultBillingPeriod,
		},
		mcnflag.StringFlag{
			Name:  "ovh-userdata",
			Usage: "OVH Cloud custom cloud-init userdata script path",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-tags",
			Usage: "OVH Cloud instance metadata tags (comma-separated)",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-security-group",
			Usage: "OVH Cloud security group name or id. Default: default",
			Value: DefaultSecurityGroup,
		},
		mcnflag.StringFlag{
			Name:  "ovh-openstack-auth-url",
			Usage: "OpenStack auth URL for security group management (e.g., https://auth.cloud.ovh.us/v3)",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-openstack-username",
			Usage: "OpenStack username for security group management",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-openstack-password",
			Usage: "OpenStack password for security group management",
			Value: "",
		},
		mcnflag.BoolFlag{
			Name:  "ovh-hosted-mks",
			Usage: "Use OVH Managed Kubernetes Service (MKS) hosted mode instead of a single VM instance",
		},
		mcnflag.StringFlag{
			Name:  "ovh-mks-cluster-name",
			Usage: "OVH MKS cluster name (required when --ovh-hosted-mks is set)",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-mks-version",
			Usage: "OVH MKS kubernetes version override",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ovh-mks-nodepool-name",
			Usage: "OVH MKS nodepool name",
			Value: "default",
		},
		mcnflag.StringFlag{
			Name:  "ovh-mks-nodepool-flavor",
			Usage: "OVH MKS nodepool flavor name",
			Value: DefaultFlavorName,
		},
		mcnflag.IntFlag{
			Name:  "ovh-mks-nodepool-size",
			Usage: "OVH MKS nodepool desired node count",
			Value: 1,
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "ovh"
}

// getClient returns an OVH API client, creating it if needed
func (d *Driver) getClient() (*API, error) {
	if d.client == nil {
		client, err := NewAPI(d.Endpoint, d.ApplicationKey, d.ApplicationSecret, d.ConsumerKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create OVH API client. Visit https://github.com/sneederco/docker-machine-driver-ovh#example-usage for setup instructions. Error: %w", err)
		}
		d.client = client
	}
	return d.client, nil
}

// SetConfigFromFlags assigns and verifies the command-line arguments presented to the driver.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	// API credentials
	d.ApplicationKey = flags.String("ovh-application-key")
	d.ApplicationSecret = flags.String("ovh-application-secret")
	d.ConsumerKey = flags.String("ovh-consumer-key")
	d.Endpoint = flags.String("ovh-endpoint")

	// Required configuration
	d.ProjectName = flags.String("ovh-project")
	d.RegionName = flags.String("ovh-region")
	d.FlavorName = flags.String("ovh-flavor")
	d.ImageName = flags.String("ovh-image")

	// Optional configuration
	d.PrivateNetworkName = flags.String("ovh-private-network")
	d.KeyPairName = flags.String("ovh-ssh-key-name")
	d.BillingPeriod = flags.String("ovh-billing-period")
	d.UserdataPath = flags.String("ovh-userdata")
	d.Tags = flags.String("ovh-tags")
	d.SecurityGroup = flags.String("ovh-security-group")
	d.OpenstackAuthUrl = flags.String("ovh-openstack-auth-url")
	d.OpenstackUsername = flags.String("ovh-openstack-username")
	d.OpenstackPassword = flags.String("ovh-openstack-password")

	// MKS configuration
	d.HostedMKS = flags.Bool("ovh-hosted-mks")
	d.MKSClusterName = flags.String("ovh-mks-cluster-name")
	d.MKSVersion = flags.String("ovh-mks-version")
	d.MKSNodePoolName = flags.String("ovh-mks-nodepool-name")
	d.MKSNodePoolFlavor = flags.String("ovh-mks-nodepool-flavor")
	d.MKSNodePoolDesiredSize = flags.Int("ovh-mks-nodepool-size")

	// Swarm configuration
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	// SSH configuration
	d.SSHUser = flags.String("ovh-ssh-user")

	return nil
}

// PreCreateCheck validates the driver configuration before creating an instance.
func (d *Driver) PreCreateCheck() error {
	ctx := context.Background()

	client, err := d.getClient()
	if err != nil {
		return err
	}

	// Validate billing period
	log.Debug("Validating billing period")
	if d.BillingPeriod != "monthly" && d.BillingPeriod != "hourly" {
		return fmt.Errorf("invalid billing period '%s'. Must be 'hourly' or 'monthly'", d.BillingPeriod)
	}
	log.Debugf("Selected billing period: %s", d.BillingPeriod)

	// Load userdata if provided
	if d.UserdataPath != "" {
		log.Debugf("Loading userdata from: %s", d.UserdataPath)
		userdataBytes, err := ioutil.ReadFile(d.UserdataPath)
		if err != nil {
			return fmt.Errorf("failed to read userdata file '%s': %w", d.UserdataPath, err)
		}
		d.Userdata = string(userdataBytes)
		log.Debug("Userdata loaded successfully")
	}

	// Resolve project ID
	if err := d.resolveProject(ctx, client); err != nil {
		return err
	}

	// Validate required parameters based on mode
	if d.HostedMKS {
		if strings.TrimSpace(d.MKSClusterName) == "" {
			return fmt.Errorf("missing required value for '--ovh-mks-cluster-name' when '--ovh-hosted-mks' is enabled")
		}
		if strings.TrimSpace(d.MKSNodePoolFlavor) == "" {
			return fmt.Errorf("missing required value for '--ovh-mks-nodepool-flavor' when '--ovh-hosted-mks' is enabled")
		}
		if d.MKSNodePoolDesiredSize < 1 {
			return fmt.Errorf("invalid value %d for '--ovh-mks-nodepool-size'. Must be >= 1", d.MKSNodePoolDesiredSize)
		}
		log.Debug("Hosted MKS mode validated successfully")
		return nil
	}

	// Standard VM mode validation
	if strings.TrimSpace(d.RegionName) == "" {
		return fmt.Errorf("missing required value for '--ovh-region'")
	}
	if strings.TrimSpace(d.FlavorName) == "" {
		return fmt.Errorf("missing required value for '--ovh-flavor'")
	}
	if strings.TrimSpace(d.ImageName) == "" {
		return fmt.Errorf("missing required value for '--ovh-image'")
	}

	// Validate region
	if err := d.validateRegion(ctx, client); err != nil {
		return err
	}

	// Validate flavor
	if err := d.validateFlavor(ctx, client); err != nil {
		return err
	}

	// Validate image
	if err := d.validateImage(ctx, client); err != nil {
		return err
	}

	// Validate private network if specified
	if err := d.validateNetwork(ctx, client); err != nil {
		return err
	}

	// Prepare SSH key
	if err := d.prepareSSHKey(); err != nil {
		return err
	}

	return nil
}

// resolveProject resolves the project name to a project ID.
func (d *Driver) resolveProject(ctx context.Context, client *API) error {
	log.Debug("Resolving project")

	if d.ProjectName != "" {
		project, err := client.GetProjectByName(ctx, d.ProjectName)
		if err != nil {
			return fmt.Errorf("failed to find project '%s': %w", d.ProjectName, err)
		}
		d.ProjectID = project.ID
		log.Debugf("Found project '%s' with ID: %s", project.Name, d.ProjectID)
		return nil
	}

	// No project specified, try to auto-select
	projects, err := client.GetProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("no Cloud projects found. Create one at: %s", CustomerInterface)
	}

	if len(projects) == 1 {
		d.ProjectID = projects[0]
		project, _ := client.GetProject(ctx, d.ProjectID)
		if project != nil {
			log.Debugf("Auto-selected project '%s' (ID: %s)", project.Name, d.ProjectID)
		} else {
			log.Debugf("Auto-selected project ID: %s", d.ProjectID)
		}
		return nil
	}

	// Multiple projects exist, build a helpful error message
	var projectNames []string
	for _, projectID := range projects {
		project, err := client.GetProject(ctx, projectID)
		if err != nil {
			projectNames = append(projectNames, projectID)
		} else {
			projectNames = append(projectNames, fmt.Sprintf("%s (%s)", project.Name, projectID))
		}
	}

	return fmt.Errorf("multiple Cloud projects found: %s. Please specify one using '--ovh-project'", strings.Join(projectNames, ", "))
}

// validateRegion validates the specified region exists.
func (d *Driver) validateRegion(ctx context.Context, client *API) error {
	log.Debug("Validating region")

	regions, err := client.GetRegions(ctx, d.ProjectID)
	if err != nil {
		if containsIgnoreCase(fallbackRegions, d.RegionName) {
			log.Warnf("Could not fetch regions from OVH API, accepting fallback region: %s", d.RegionName)
			return nil
		}
		return fmt.Errorf("failed to validate region: %w", err)
	}

	if !containsIgnoreCase(regions, d.RegionName) {
		return fmt.Errorf("invalid region '%s'. Available regions: %s. Visit: %s", d.RegionName, strings.Join(regions, ", "), CustomerInterface)
	}

	log.Debugf("Region validated: %s", d.RegionName)
	return nil
}

// validateFlavor validates the specified flavor exists and resolves its ID.
func (d *Driver) validateFlavor(ctx context.Context, client *API) error {
	log.Debug("Validating flavor")

	flavor, err := client.GetFlavorByName(ctx, d.ProjectID, d.RegionName, d.FlavorName)
	if err != nil {
		if containsIgnoreCase(fallbackFlavors, d.FlavorName) {
			log.Warnf("Could not resolve flavor via OVH API, using fallback value: %s", d.FlavorName)
			d.FlavorID = d.FlavorName
			return nil
		}
		return fmt.Errorf("failed to validate flavor: %w", err)
	}

	d.FlavorID = flavor.ID
	log.Debugf("Flavor validated: %s (ID: %s, %d vCPUs, %dGB RAM)", flavor.Name, d.FlavorID, flavor.Vcpus, flavor.MemoryGB)
	return nil
}

// validateImage validates the specified image exists and resolves its ID.
func (d *Driver) validateImage(ctx context.Context, client *API) error {
	log.Debug("Validating image")

	image, err := client.GetImageByName(ctx, d.ProjectID, d.RegionName, d.ImageName)
	if err != nil {
		if containsIgnoreCase(fallbackImages, d.ImageName) {
			log.Warnf("Could not resolve image via OVH API, using fallback value: %s", d.ImageName)
			d.ImageID = d.ImageName
			return nil
		}
		return fmt.Errorf("failed to validate image: %w", err)
	}

	d.ImageID = image.ID
	log.Debugf("Image validated: %s (ID: %s, OS: %s)", image.Name, d.ImageID, image.OS)
	return nil
}

// validateNetwork validates the private network if specified and configures network IDs.
func (d *Driver) validateNetwork(ctx context.Context, client *API) error {
	log.Debug("Validating network configuration")

	if d.PrivateNetworkName == "" {
		log.Debug("No private network specified, using public network only")
		return nil
	}

	privateNetwork, err := client.GetPrivateNetworkByName(ctx, d.ProjectID, d.PrivateNetworkName)
	if err != nil {
		return fmt.Errorf("failed to find private network '%s': %w", d.PrivateNetworkName, err)
	}

	d.NetworkIDs = append(d.NetworkIDs, privateNetwork.ID)
	log.Debugf("Private network validated: %s (ID: %s)", privateNetwork.Name, privateNetwork.ID)

	// Add public network as well
	publicNetworkID, err := client.GetPublicNetworkID(ctx, d.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get public network ID: %w", err)
	}

	d.NetworkIDs = append(d.NetworkIDs, publicNetworkID)
	log.Debugf("Public network ID: %s", publicNetworkID)

	return nil
}

// prepareSSHKey prepares the SSH key configuration.
func (d *Driver) prepareSSHKey() error {
	log.Debug("Preparing SSH key configuration")

	// Use a common key or create a machine-specific one
	if d.KeyPairName != "" {
		keyPath := filepath.Join(d.StorePath, "sshkeys", d.KeyPairName)
		if _, err := os.Stat(keyPath); err == nil {
			d.SSHKeyPath = keyPath
			log.Debugf("Using existing SSH key: %s", d.KeyPairName)
		} else {
			log.Debugf("SSH key '%s' not found locally, assuming it exists in OVH or ~/.ssh/", d.KeyPairName)
		}
	} else {
		// Generate a unique key name for this machine
		d.KeyPairName = fmt.Sprintf("%s-%s", d.MachineName, mcnutils.GenerateRandomID())
		sanitizeKeyPairName(&d.KeyPairName)
		d.SSHKeyPath = d.ResolveStorePath(d.KeyPairName)
		log.Debugf("Will create SSH key: %s", d.KeyPairName)
	}

	return nil
}

// ensureSSHKey ensures an SSH key exists in OVH for the machine.
func (d *Driver) ensureSSHKey(ctx context.Context) error {
	client, err := d.getClient()
	if err != nil {
		return err
	}

	log.Debugf("Checking for existing SSH key: %s", d.KeyPairName)

	// Try to find existing key
	sshKey, _ := client.GetSshkeyByName(ctx, d.ProjectID, d.RegionName, d.KeyPairName)
	if sshKey != nil {
		d.KeyPairID = sshKey.ID
		log.Debugf("Found existing SSH key (ID: %s)", d.KeyPairID)
		return nil
	}

	// Key doesn't exist, generate and upload it
	log.Debugf("Creating new SSH key: %s", d.KeyPairName)

	// Ensure key directory exists
	keyfile := d.GetSSHKeyPath()
	keypath := filepath.Dir(keyfile)
	if err := os.MkdirAll(keypath, 0700); err != nil {
		return fmt.Errorf("failed to create SSH key directory: %w", err)
	}

	// Generate SSH key pair
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	// Read public key
	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return fmt.Errorf("failed to read public SSH key: %w", err)
	}

	// Upload to OVH
	sshKey, err = client.CreateSshkey(ctx, d.ProjectID, d.KeyPairName, string(publicKey))
	if err != nil {
		return fmt.Errorf("failed to upload SSH key to OVH: %w", err)
	}

	d.KeyPairID = sshKey.ID
	log.Debugf("Created SSH key (ID: %s)", d.KeyPairID)

	return nil
}

// waitForInstanceStatus waits for the instance to reach the specified status.
func (d *Driver) waitForInstanceStatus(ctx context.Context, targetStatus string) (*Instance, error) {
	client, err := d.getClient()
	if err != nil {
		return nil, err
	}

	log.Debugf("Waiting for instance to reach status: %s", targetStatus)

	timeout := time.After(statusTimeout * time.Second)
	ticker := time.NewTicker(statusCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled while waiting for instance status: %w", ctx.Err())

		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for instance to reach status '%s'", targetStatus)

		case <-ticker.C:
			instance, err := client.GetInstance(ctx, d.ProjectID, d.InstanceID)
			if err != nil {
				return nil, fmt.Errorf("failed to get instance status: %w", err)
			}

			log.Debugf("Instance status: %s (target: %s)", instance.Status, targetStatus)

			if instance.Status == "ERROR" {
				return nil, fmt.Errorf("instance entered ERROR state")
			}

			if instance.Status == targetStatus {
				return instance, nil
			}
		}
	}
}

// Create creates a new OVH Cloud instance or MKS cluster.
func (d *Driver) Create() error {
	ctx := context.Background()

	client, err := d.getClient()
	if err != nil {
		return err
	}

	if d.HostedMKS {
		return d.createMKSCluster(ctx, client)
	}

	return d.createInstance(ctx, client)
}

// createMKSCluster creates an OVH Managed Kubernetes cluster.
func (d *Driver) createMKSCluster(ctx context.Context, client *API) error {
	log.Infof("Creating OVH Managed Kubernetes cluster: %s", d.MKSClusterName)

	clusterReq := MKSClusterCreateReq{
		Name:   d.MKSClusterName,
		Region: d.RegionName,
	}
	if d.MKSVersion != "" {
		clusterReq.Version = d.MKSVersion
	}

	cluster, err := client.CreateMKSCluster(ctx, d.ProjectID, clusterReq)
	if err != nil {
		return fmt.Errorf("failed to create MKS cluster: %w", err)
	}

	d.MKSClusterID = cluster.ID
	log.Infof("Created MKS cluster (ID: %s)", d.MKSClusterID)

	// Create node pool
	log.Infof("Creating node pool: %s", d.MKSNodePoolName)

	nodePoolReq := MKSNodePoolCreateReq{
		Name:         d.MKSNodePoolName,
		FlavorName:   d.MKSNodePoolFlavor,
		DesiredNodes: d.MKSNodePoolDesiredSize,
	}

	nodePool, err := client.CreateMKSNodePool(ctx, d.ProjectID, d.MKSClusterID, nodePoolReq)
	if err != nil {
		// Cleanup cluster on node pool creation failure
		log.Warnf("Failed to create node pool, cleaning up cluster...")
		_ = client.DeleteMKSCluster(ctx, d.ProjectID, d.MKSClusterID)
		return fmt.Errorf("failed to create MKS node pool: %w", err)
	}

	d.MKSNodePoolID = nodePool.ID
	log.Infof("Created MKS cluster '%s' (ID: %s) with node pool '%s' (ID: %s)",
		d.MKSClusterName, d.MKSClusterID, d.MKSNodePoolName, d.MKSNodePoolID)

	return nil
}

// createInstance creates a standard OVH Cloud instance.
func (d *Driver) createInstance(ctx context.Context, client *API) error {
	// Ensure SSH key exists
	if err := d.ensureSSHKey(ctx); err != nil {
		return fmt.Errorf("failed to ensure SSH key: %w", err)
	}

	// Create instance
	log.Infof("Creating OVH Cloud instance: %s", d.MachineName)
	log.Debugf("  Region: %s", d.RegionName)
	log.Debugf("  Flavor: %s (ID: %s)", d.FlavorName, d.FlavorID)
	log.Debugf("  Image: %s (ID: %s)", d.ImageName, d.ImageID)
	log.Debugf("  Billing: %s", d.BillingPeriod)
	if d.Userdata != "" {
		log.Debugf("  Userdata: %d bytes", len(d.Userdata))
	}

	monthlyBilling := d.BillingPeriod == "monthly"

	instance, err := client.CreateInstance(
		ctx,
		d.ProjectID,
		d.MachineName,
		d.KeyPairID,
		d.FlavorID,
		d.ImageID,
		d.RegionName,
		d.Userdata,
		d.NetworkIDs,
		monthlyBilling,
	)
	if err != nil {
		// Cleanup SSH key if we created it
		d.cleanupSSHKey(ctx, client)
		return fmt.Errorf("failed to create instance: %w", err)
	}

	d.InstanceID = instance.ID
	log.Infof("Instance created (ID: %s)", d.InstanceID)

	// Wait for instance to become active
	log.Info("Waiting for instance to become active...")
	instance, err = d.waitForInstanceStatus(ctx, "ACTIVE")
	if err != nil {
		// Cleanup on failure
		d.cleanupInstance(ctx, client)
		return err
	}

	// Extract IP address
	d.IPAddress = ""
	for _, ip := range instance.IPAddresses {
		if ip.Type == "public" {
			d.IPAddress = ip.IP
			break
		}
	}

	if d.IPAddress == "" {
		d.cleanupInstance(ctx, client)
		return fmt.Errorf("no public IP address found for instance")
	}

	log.Infof("Instance is active. IP address: %s", d.IPAddress)

	// Attach security group if specified and OpenStack credentials provided
	log.Debugf("SG Check: SecurityGroup=%s, AuthUrl=%s, Username=%s, Password=%s",
		d.SecurityGroup, d.OpenstackAuthUrl, d.OpenstackUsername, d.OpenstackPassword != "")
	if d.SecurityGroup != "" && d.SecurityGroup != "default" &&
		d.OpenstackAuthUrl != "" && d.OpenstackUsername != "" && d.OpenstackPassword != "" {
		log.Warnf("Attaching security group %s to instance...", d.SecurityGroup)
		if err := AttachSecurityGroupToInstance(
			d.OpenstackAuthUrl,
			d.OpenstackUsername,
			d.OpenstackPassword,
			d.ProjectID,
			d.RegionName,
			d.InstanceID,
			d.SecurityGroup,
		); err != nil {
			log.Warnf("Failed to attach security group: %v", err)
			// Don't fail the whole creation, just warn
		} else {
			log.Warnf("Security group %s attached successfully", d.SecurityGroup)
		}
	}

	// Disable IPv6 on the instance to ensure RKE2 uses IPv4 for cluster communication
	// This is critical for multi-node clusters where nodes need to reach each other
	log.Info("Disabling IPv6 on instance to ensure IPv4 cluster communication...")
	if err := d.disableIPv6(); err != nil {
		log.Warnf("Failed to disable IPv6: %v (continuing anyway)", err)
	} else {
		log.Info("IPv6 disabled successfully")
	}

	return nil
}

// cleanupSSHKey removes the SSH key if we created it (for cleanup on failure).
func (d *Driver) cleanupSSHKey(ctx context.Context, client *API) {
	if d.KeyPairID != "" && strings.HasPrefix(d.KeyPairName, d.MachineName) {
		log.Debugf("Cleaning up SSH key (ID: %s)", d.KeyPairID)
		_ = client.DeleteSshkey(ctx, d.ProjectID, d.KeyPairID)
	}
}

// cleanupInstance removes the instance and SSH key (for cleanup on failure).
func (d *Driver) cleanupInstance(ctx context.Context, client *API) {
	if d.InstanceID != "" {
		log.Debugf("Cleaning up instance (ID: %s)", d.InstanceID)
		_ = client.DeleteInstance(ctx, d.ProjectID, d.InstanceID)
	}
	d.cleanupSSHKey(ctx, client)
}

// disableIPv6 disables IPv6 on the instance via SSH.
// This ensures RKE2 uses IPv4 addresses for cluster communication,
// which is critical for multi-node clusters on OVH where IPv6 may not be routable.
func (d *Driver) disableIPv6() error {
	// Wait a moment for SSH to be ready
	time.Sleep(5 * time.Second)

	// Commands to disable IPv6
	commands := []string{
		// Disable IPv6 via sysctl
		"sudo sysctl -w net.ipv6.conf.all.disable_ipv6=1",
		"sudo sysctl -w net.ipv6.conf.default.disable_ipv6=1",
		// Make it persistent
		"echo 'net.ipv6.conf.all.disable_ipv6=1' | sudo tee -a /etc/sysctl.conf",
		"echo 'net.ipv6.conf.default.disable_ipv6=1' | sudo tee -a /etc/sysctl.conf",
	}

	// Run each command via SSH
	for _, cmd := range commands {
		if _, err := drivers.RunSSHCommandFromDriver(d, cmd); err != nil {
			return fmt.Errorf("failed to run '%s': %w", cmd, err)
		}
	}

	return nil
}

// GetState returns the current state of the instance or MKS cluster.
func (d *Driver) GetState() (state.State, error) {
	ctx := context.Background()

	client, err := d.getClient()
	if err != nil {
		return state.None, err
	}

	if d.HostedMKS {
		return d.getMKSState(ctx, client)
	}

	return d.getInstanceState(ctx, client)
}

// getMKSState returns the state of an MKS cluster.
func (d *Driver) getMKSState(ctx context.Context, client *API) (state.State, error) {
	if d.MKSClusterID == "" {
		return state.None, nil
	}

	// List all clusters and find ours
	clusters, err := client.ListMKSClusters(ctx, d.ProjectID)
	if err != nil {
		return state.None, fmt.Errorf("failed to list MKS clusters: %w", err)
	}

	// Find our cluster by ID
	for _, cluster := range clusters {
		if cluster.ID == d.MKSClusterID {
			log.Debugf("MKS cluster status: %s", cluster.Status)

			switch cluster.Status {
			case "READY":
				return state.Running, nil
			case "CREATING", "UPDATING":
				return state.Starting, nil
			case "DELETING":
				return state.Stopping, nil
			case "ERROR":
				return state.Error, nil
			default:
				return state.None, nil
			}
		}
	}

	// Cluster not found
	return state.None, nil
}

// getInstanceState returns the state of a standard instance.
func (d *Driver) getInstanceState(ctx context.Context, client *API) (state.State, error) {
	if d.InstanceID == "" {
		return state.None, nil
	}

	instance, err := client.GetInstance(ctx, d.ProjectID, d.InstanceID)
	if err != nil {
		return state.None, fmt.Errorf("failed to get instance state: %w", err)
	}

	log.Debugf("Instance status: %s", instance.Status)

	switch instance.Status {
	case "ACTIVE":
		return state.Running, nil
	case "BUILD", "BUILDING":
		return state.Starting, nil
	case "SHUTOFF", "STOPPED":
		return state.Stopped, nil
	case "PAUSED":
		return state.Paused, nil
	case "SUSPENDED":
		return state.Saved, nil
	case "ERROR":
		return state.Error, nil
	case "DELETED", "SOFT_DELETED":
		return state.None, nil
	default:
		return state.None, nil
	}
}

// Start starts a stopped instance.
func (d *Driver) Start() error {
	ctx := context.Background()

	if d.HostedMKS {
		return fmt.Errorf("start is not supported for MKS clusters")
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	log.Infof("Starting instance (ID: %s)", d.InstanceID)

	if err := client.StartInstance(ctx, d.ProjectID, d.InstanceID); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	// Wait for instance to become active
	if _, err := d.waitForInstanceStatus(ctx, "ACTIVE"); err != nil {
		return err
	}

	log.Info("Instance started successfully")
	return nil
}

// Stop stops a running instance.
func (d *Driver) Stop() error {
	ctx := context.Background()

	if d.HostedMKS {
		return fmt.Errorf("stop is not supported for MKS clusters")
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	log.Infof("Stopping instance (ID: %s)", d.InstanceID)

	if err := client.StopInstance(ctx, d.ProjectID, d.InstanceID); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// Wait for instance to shut off
	if _, err := d.waitForInstanceStatus(ctx, "SHUTOFF"); err != nil {
		return err
	}

	log.Info("Instance stopped successfully")
	return nil
}

// Restart reboots a running instance.
func (d *Driver) Restart() error {
	ctx := context.Background()

	if d.HostedMKS {
		return fmt.Errorf("restart is not supported for MKS clusters")
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	log.Infof("Restarting instance (ID: %s)", d.InstanceID)

	if err := client.RebootInstance(ctx, d.ProjectID, d.InstanceID, false); err != nil {
		return fmt.Errorf("failed to restart instance: %w", err)
	}

	// Wait for instance to become active again
	if _, err := d.waitForInstanceStatus(ctx, "ACTIVE"); err != nil {
		return err
	}

	log.Info("Instance restarted successfully")
	return nil
}

// Kill force-stops an instance (hard reboot).
func (d *Driver) Kill() error {
	if d.HostedMKS {
		return fmt.Errorf("kill is not supported for MKS clusters")
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	log.Infof("Force-stopping instance (ID: %s)", d.InstanceID)

	if err := client.RebootInstance(context.Background(), d.ProjectID, d.InstanceID, true); err != nil {
		return fmt.Errorf("failed to force-stop instance: %w", err)
	}

	log.Info("Instance force-stopped")
	return nil
}

// Remove deletes the instance or MKS cluster and associated resources.
func (d *Driver) Remove() error {
	ctx := context.Background()

	client, err := d.getClient()
	if err != nil {
		return err
	}

	if d.HostedMKS {
		return d.removeMKSCluster(ctx, client)
	}

	return d.removeInstance(ctx, client)
}

// removeMKSCluster deletes an MKS cluster.
func (d *Driver) removeMKSCluster(ctx context.Context, client *API) error {
	if d.MKSClusterID == "" {
		log.Info("No MKS cluster to remove")
		return nil
	}

	log.Infof("Deleting MKS cluster (ID: %s)", d.MKSClusterID)

	if err := client.DeleteMKSCluster(ctx, d.ProjectID, d.MKSClusterID); err != nil {
		return fmt.Errorf("failed to delete MKS cluster: %w", err)
	}

	log.Info("MKS cluster deleted successfully")
	return nil
}

// removeInstance deletes a standard instance and its SSH key.
func (d *Driver) removeInstance(ctx context.Context, client *API) error {
	// Delete instance
	if d.InstanceID != "" {
		log.Infof("Deleting instance (ID: %s)", d.InstanceID)
		if err := client.DeleteInstance(ctx, d.ProjectID, d.InstanceID); err != nil {
			return fmt.Errorf("failed to delete instance: %w", err)
		}
		log.Info("Instance deleted successfully")
	}

	// Delete SSH key if we created it
	if d.KeyPairID != "" && strings.HasPrefix(d.KeyPairName, d.MachineName) {
		log.Debugf("Deleting SSH key (ID: %s)", d.KeyPairID)
		if err := client.DeleteSshkey(ctx, d.ProjectID, d.KeyPairID); err != nil {
			log.Warnf("Failed to delete SSH key: %v", err)
		} else {
			log.Debug("SSH key deleted successfully")
		}
	} else {
		log.Debugf("Keeping SSH key '%s' (not machine-specific)", d.KeyPairName)
	}

	return nil
}

// GetURL returns the Docker daemon URL for this machine.
func (d *Driver) GetURL() (string, error) {
	if d.IPAddress == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(d.IPAddress, "2376")), nil
}

// GetSSHHostname returns the hostname for SSH connections.
func (d *Driver) GetSSHHostname() (string, error) {
	// If IP is not set, try to fetch it dynamically from OVH API
	// This fixes an issue where the IP is lost after security group attachment
	if d.IPAddress == "" && d.InstanceID != "" && !d.HostedMKS {
		ctx := context.Background()
		client, err := d.getClient()
		if err == nil {
			instance, err := client.GetInstance(ctx, d.ProjectID, d.InstanceID)
			if err == nil {
				for _, ip := range instance.IPAddresses {
					if ip.Type == "public" {
						d.IPAddress = ip.IP
						break
					}
				}
			}
		}
	}
	return d.IPAddress, nil
}

// GetSSHKeyPath returns the path to the SSH private key.
func (d *Driver) GetSSHKeyPath() string {
	return d.SSHKeyPath
}

// publicSSHKeyPath returns the path to the SSH public key.
func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// ListHostedMKSClusters is a helper for hosted MKS mode.
func (d *Driver) ListHostedMKSClusters() (MKSClusters, error) {
	ctx := context.Background()
	client, err := d.getClient()
	if err != nil {
		return nil, err
	}
	return client.ListMKSClusters(ctx, d.ProjectID)
}

// ScaleHostedMKSNodePool is a helper for hosted MKS mode.
func (d *Driver) ScaleHostedMKSNodePool(desiredNodes int) error {
	if d.MKSClusterID == "" || d.MKSNodePoolID == "" {
		return fmt.Errorf("MKS cluster and node pool IDs are required for scaling")
	}

	ctx := context.Background()
	client, err := d.getClient()
	if err != nil {
		return err
	}

	log.Infof("Scaling MKS node pool to %d nodes", desiredNodes)
	return client.ScaleMKSNodePool(ctx, d.ProjectID, d.MKSClusterID, d.MKSNodePoolID, desiredNodes)
}

// Helper functions

// containsIgnoreCase checks if a slice contains a string (case-insensitive).
func containsIgnoreCase(items []string, value string) bool {
	for _, item := range items {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

// sanitizeKeyPairName sanitizes a key pair name for OVH.
func sanitizeKeyPairName(s *string) {
	*s = strings.Replace(*s, ".", "_", -1)
}
