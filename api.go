package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/sirupsen/logrus"
)

const (
	// CustomerInterface is the URL of the customer interface, for error messages
	CustomerInterface = "https://www.ovh.com/manager/cloud/index.html"

	// Retry configuration
	defaultMaxRetries    = 3
	defaultRetryDelay    = 1 * time.Second
	defaultRetryMaxDelay = 30 * time.Second

	// Rate limiting
	defaultRateLimitDelay = 100 * time.Millisecond
)

// APIError wraps OVH API errors with additional context
type APIError struct {
	Operation  string // The operation that failed (e.g., "GetProject")
	Resource   string // The resource being accessed (e.g., "project/abc123")
	StatusCode int    // HTTP status code
	OVHCode    int    // OVH-specific error code
	Message    string // Error message
	Err        error  // Original error
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s failed for %s (HTTP %d): %s", e.Operation, e.Resource, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%s failed for %s: %s", e.Operation, e.Resource, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error is a 404 Not Found
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == 404
}

// IsRateLimited returns true if the error is a 429 Too Many Requests
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == 429
}

// IsRetryable returns true if the error can be retried
func (e *APIError) IsRetryable() bool {
	// Retry on rate limits and 5xx errors
	return e.IsRateLimited() || (e.StatusCode >= 500 && e.StatusCode < 600)
}

// API is a handle to an OVH API client with retry and rate limiting support
type API struct {
	client        *ovh.Client
	logger        *logrus.Logger
	maxRetries    int
	retryDelay    time.Duration
	rateLimitWait time.Duration
}

// APIConfig holds configuration for the API client
type APIConfig struct {
	MaxRetries    int
	RetryDelay    time.Duration
	RateLimitWait time.Duration
	Logger        *logrus.Logger
}

// Project is a go representation of a Cloud project
type Project struct {
	Name         string `json:"description"`
	ID           string `json:"project_id"`
	Unleash      bool   `json:"unleash"`
	CreationDate string `json:"creationDate"`
	OrderID      int    `json:"orderID"`
	Status       string `json:"status"`
}

// Projects is a list of project IDs
type Projects []string

// Region represents an OVH cloud region
type Region struct {
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Type      string   `json:"type"`
	Services  []string `json:"services,omitempty"`
	Continent string   `json:"continent,omitempty"`
}

// Regions is a list of regions
type Regions []Region

// Flavor is a go representation of Cloud Flavor
type Flavor struct {
	Region      string `json:"region"`
	Name        string `json:"name"`
	ID          string `json:"id"`
	OS          string `json:"osType"`
	Vcpus       int    `json:"vcpus"`
	MemoryGB    int    `json:"ram"`
	DiskSpaceGB int    `json:"disk"`
	Type        string `json:"type"`
	Available   bool   `json:"available"`
	Quota       int    `json:"quota,omitempty"`
}

// Flavors is a list flavors
type Flavors []Flavor

// Image is a go representation of a Cloud Image (VM template)
type Image struct {
	Region       string `json:"region"`
	Name         string `json:"name"`
	ID           string `json:"id"`
	OS           string `json:"type"`
	CreationDate string `json:"creationDate"`
	Status       string `json:"status"`
	MinDisk      int    `json:"minDisk"`
	Visibility   string `json:"visibility"`
	Size         float64  `json:"size,omitempty"`
	PlanCode     string `json:"planCode,omitempty"`
}

// Images is a list of Images
type Images []Image

// Network defines the private network names
type Network struct {
	Status string `json:"status"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	ID     string `json:"id"`
	VlanID int    `json:"vlanid"`
	Region string `json:"region,omitempty"`
}

// Networks is a list of Network
type Networks []Network

// SSHKeyReq defines the fields for an SSH Key upload
type SSHKeyReq struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
	Region    string `json:"region,omitempty"`
}

// SSHKey is a go representation of Cloud SSH Key
type SSHKey struct {
	Name        string   `json:"name"`
	ID          string   `json:"id"`
	PublicKey   string   `json:"publicKey"`
	Fingerprint string   `json:"fingerPrint"`
	Regions     []string `json:"region"`
}

// SSHKeys is a list of SSHKey
type SSHKeys []SSHKey

// Legacy type aliases for backward compatibility
type Sshkey = SSHKey
type Sshkeys = SSHKeys
type SshkeyReq = SSHKeyReq

// Quota represents resource quotas for a region
type Quota struct {
	Region          string `json:"region"`
	Instance        int    `json:"instance"`
	Cores           int    `json:"cores"`
	RAM             int    `json:"ram"`
	KeyPairs        int    `json:"keypair"`
	Volumes         int    `json:"volume"`
	VolumeGigabytes int    `json:"volumeGigabytes"`
}

// IP is a go representation of a Cloud IP address
type IP struct {
	IP   string `json:"ip"`
	Type string `json:"type"`
}

// IPs is a list of IPs
type IPs []IP

// NetworkParam for Cloud instance
type NetworkParam struct {
	ID string `json:"networkId"`
}

type NetworkParams []NetworkParam

// InstanceReq defines the fields for a VM creation
type InstanceReq struct {
	UserData       string        `json:"userData,omitempty"`
	Name           string        `json:"name"`
	FlavorID       string        `json:"flavorId"`
	ImageID        string        `json:"imageId"`
	Region         string        `json:"region"`
	NetworkParams  NetworkParams `json:"networks"`
	SshkeyID       string        `json:"sshKeyId"`
	MonthlyBilling bool          `json:"monthlyBilling"`
	SecurityGroup  string        `json:"securityGroup,omitempty"`
}

// Instance is a go representation of Cloud instance
type Instance struct {
	Name           string        `json:"name"`
	ID             string        `json:"id"`
	Status         string        `json:"status"`
	Created        string        `json:"created"`
	Region         string        `json:"region"`
	NetworkParams  NetworkParams `json:"networks"`
	Image          Image         `json:"image"`
	Flavor         Flavor        `json:"flavor"`
	Sshkey         SSHKey        `json:"sshKey"`
	IPAddresses    IPs           `json:"ipAddresses"`
	MonthlyBilling bool          `json:"monthlyBilling"`
	SecurityGroup  string        `json:"securityGroup,omitempty"`
}

// MKSClusterCreateReq defines the fields for OVH Managed Kubernetes cluster creation.
type MKSClusterCreateReq struct {
	Name             string                 `json:"name"`
	Region           string                 `json:"region"`
	Version          string                 `json:"version,omitempty"`
	PrivateNetworkID string                 `json:"privateNetworkId,omitempty"`
	UpdatePolicy     string                 `json:"updatePolicy,omitempty"`
	Customization    map[string]interface{} `json:"customization,omitempty"`
}

// MKSCluster is a go representation of OVH Managed Kubernetes Service cluster.
type MKSCluster struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Region           string `json:"region"`
	Version          string `json:"version"`
	Status           string `json:"status"`
	PrivateNetworkID string `json:"privateNetworkId,omitempty"`
}

// MKSClusters is a list of managed kubernetes clusters.
type MKSClusters []MKSCluster

// MKSNodePoolCreateReq defines fields for creating a nodepool on an MKS cluster.
type MKSNodePoolCreateReq struct {
	Name           string   `json:"name"`
	FlavorName     string   `json:"flavorName"`
	DesiredNodes   int      `json:"desiredNodes"`
	MinNodes       int      `json:"minNodes,omitempty"`
	MaxNodes       int      `json:"maxNodes,omitempty"`
	MonthlyBilling bool     `json:"monthlyBilling,omitempty"`
	AntiAffinity   bool     `json:"antiAffinity,omitempty"`
	Autoscale      bool     `json:"autoscale,omitempty"`
	Template       string   `json:"template,omitempty"`
	Availability   []string `json:"availabilityZones,omitempty"`
}

// MKSNodePoolScaleReq defines fields for nodepool scaling.
type MKSNodePoolScaleReq struct {
	DesiredNodes int `json:"desiredNodes"`
}

// MKSNodePool is a go representation of an OVH MKS nodepool.
type MKSNodePool struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	FlavorName   string `json:"flavorName"`
	DesiredNodes int    `json:"desiredNodes"`
	Status       string `json:"status"`
}

// RebootReq defines the fields for a VM reboot
type RebootReq struct {
	Type string `json:"type"`
}

// NewAPI instantiates a Cloud API driver from credentials, for a given endpoint
func NewAPI(endpoint, applicationKey, applicationSecret, consumerKey string) (*API, error) {
	return NewAPIWithConfig(endpoint, applicationKey, applicationSecret, consumerKey, nil)
}

// NewAPIWithConfig instantiates a Cloud API driver with custom configuration
func NewAPIWithConfig(endpoint, applicationKey, applicationSecret, consumerKey string, config *APIConfig) (*API, error) {
	client, err := ovh.NewClient(endpoint, applicationKey, applicationSecret, consumerKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create OVH client: %w", err)
	}

	api := &API{
		client:        client,
		maxRetries:    defaultMaxRetries,
		retryDelay:    defaultRetryDelay,
		rateLimitWait: defaultRateLimitDelay,
		logger:        logrus.New(),
	}

	if config != nil {
		if config.MaxRetries > 0 {
			api.maxRetries = config.MaxRetries
		}
		if config.RetryDelay > 0 {
			api.retryDelay = config.RetryDelay
		}
		if config.RateLimitWait > 0 {
			api.rateLimitWait = config.RateLimitWait
		}
		if config.Logger != nil {
			api.logger = config.Logger
		}
	}

	return api, nil
}

// doWithRetry executes an API call with retry logic and exponential backoff
func (a *API) doWithRetry(ctx context.Context, operation string, resource string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return &APIError{
				Operation: operation,
				Resource:  resource,
				Message:   "context cancelled",
				Err:       ctx.Err(),
			}
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			if attempt > 0 {
				a.logger.WithFields(logrus.Fields{
					"operation": operation,
					"resource":  resource,
					"attempts":  attempt + 1,
				}).Debug("API call succeeded after retry")
			}
			return nil
		}

		// Wrap error if not already wrapped
		apiErr, ok := err.(*APIError)
		if !ok {
			// Try to extract info from OVH API error
			if ovhErr, ok := err.(*ovh.APIError); ok {
				apiErr = &APIError{
					Operation:  operation,
					Resource:   resource,
					StatusCode: ovhErr.Code,
					OVHCode:    ovhErr.Code,
					Message:    ovhErr.Message,
					Err:        err,
				}
			} else {
				apiErr = &APIError{
					Operation: operation,
					Resource:  resource,
					Message:   err.Error(),
					Err:       err,
				}
			}
		}

		lastErr = apiErr

		// Don't retry on non-retryable errors
		if !apiErr.IsRetryable() {
			a.logger.WithFields(logrus.Fields{
				"operation":   operation,
				"resource":    resource,
				"status_code": apiErr.StatusCode,
				"error":       apiErr.Message,
			}).Debug("Non-retryable error")
			return apiErr
		}

		// Don't retry if we've exhausted attempts
		if attempt >= a.maxRetries {
			a.logger.WithFields(logrus.Fields{
				"operation": operation,
				"resource":  resource,
				"attempts":  attempt + 1,
				"error":     apiErr.Message,
			}).Warn("Max retries exhausted")
			return apiErr
		}

		// Calculate backoff delay (exponential with jitter)
		backoff := time.Duration(math.Pow(2, float64(attempt))) * a.retryDelay
		if backoff > defaultRetryMaxDelay {
			backoff = defaultRetryMaxDelay
		}

		// Add rate limit wait if this was a rate limit error
		if apiErr.IsRateLimited() {
			backoff += a.rateLimitWait
		}

		a.logger.WithFields(logrus.Fields{
			"operation": operation,
			"resource":  resource,
			"attempt":   attempt + 1,
			"backoff":   backoff,
			"error":     apiErr.Message,
		}).Debug("Retrying after backoff")

		// Wait for backoff period or context cancellation
		select {
		case <-ctx.Done():
			return &APIError{
				Operation: operation,
				Resource:  resource,
				Message:   "context cancelled during retry",
				Err:       ctx.Err(),
			}
		case <-time.After(backoff):
		}
	}

	return lastErr
}

// GetProjects returns a list of string project IDs
func (a *API) GetProjects(ctx context.Context) (Projects, error) {
	var projects Projects
	err := a.doWithRetry(ctx, "GetProjects", "/cloud/project", func() error {
		return a.client.GetWithContext(ctx, "/cloud/project", &projects)
	})
	return projects, err
}

// GetProject returns the details of a project given a project id
func (a *API) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var project Project
	resource := fmt.Sprintf("/cloud/project/%s", projectID)
	err := a.doWithRetry(ctx, "GetProject", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &project)
	})
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProjectByName returns the details of a project given its name
func (a *API) GetProjectByName(ctx context.Context, projectName string) (*Project, error) {
	projects, err := a.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	// If projectName is a valid projectID return it
	for _, projectID := range projects {
		if projectID == projectName {
			return a.GetProject(ctx, projectID)
		}
	}

	// Attempt to find a project matching projectName
	for _, projectID := range projects {
		project, err := a.GetProject(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if project.Name == projectName {
			return project, nil
		}
	}

	return nil, &APIError{
		Operation: "GetProjectByName",
		Resource:  projectName,
		Message:   fmt.Sprintf("Project '%s' does not exist. Visit %s to create or rename a project", projectName, CustomerInterface),
	}
}

// ListRegions returns the list of available regions for a given project
func (a *API) ListRegions(ctx context.Context, projectID string) (Regions, error) {
	var regions Regions
	resource := fmt.Sprintf("/cloud/project/%s/region", projectID)
	err := a.doWithRetry(ctx, "ListRegions", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &regions)
	})
	return regions, err
}

// GetRegions returns the list of valid region names for a given project (legacy method)
func (a *API) GetRegions(ctx context.Context, projectID string) ([]string, error) {
	regions, err := a.ListRegions(ctx, projectID)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(regions))
	for i, r := range regions {
		names[i] = r.Name
	}
	return names, nil
}

// ListFlavors returns the list of available flavors for a given project in a given region
func (a *API) ListFlavors(ctx context.Context, projectID, region string) (Flavors, error) {
	var flavors Flavors
	resource := fmt.Sprintf("/cloud/project/%s/flavor?region=%s", projectID, region)
	err := a.doWithRetry(ctx, "ListFlavors", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &flavors)
	})
	return flavors, err
}

// GetFlavors returns the list of available flavors for a given project in a given region (legacy method)
func (a *API) GetFlavors(ctx context.Context, projectID, region string) (Flavors, error) {
	return a.ListFlavors(ctx, projectID, region)
}

// GetFlavorByName returns the details of a flavor given its name
func (a *API) GetFlavorByName(ctx context.Context, projectID, region, flavorName string) (*Flavor, error) {
	flavors, err := a.ListFlavors(ctx, projectID, region)
	if err != nil {
		return nil, err
	}

	// Find first matching Linux flavor
	for _, flavor := range flavors {
		if flavor.OS != "linux" {
			continue
		}
		if flavor.ID == flavorName || flavor.Name == flavorName {
			return &flavor, nil
		}
	}

	return nil, &APIError{
		Operation: "GetFlavorByName",
		Resource:  fmt.Sprintf("%s/%s", region, flavorName),
		Message:   fmt.Sprintf("Flavor '%s' does not exist in region %s. Visit %s for available flavors", flavorName, region, CustomerInterface),
	}
}

// ListImages returns a list of images for a given project in a given region
func (a *API) ListImages(ctx context.Context, projectID, region string) (Images, error) {
	var images Images
	resource := fmt.Sprintf("/cloud/project/%s/image?osType=linux&region=%s", projectID, region)
	err := a.doWithRetry(ctx, "ListImages", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &images)
	})
	return images, err
}

// GetImages returns a list of images for a given project in a given region (legacy method)
func (a *API) GetImages(ctx context.Context, projectID, region string) (Images, error) {
	return a.ListImages(ctx, projectID, region)
}

// GetImageByName returns the details of an image given its name, a project and a region
func (a *API) GetImageByName(ctx context.Context, projectID, region, imageName string) (*Image, error) {
	images, err := a.ListImages(ctx, projectID, region)
	if err != nil {
		return nil, err
	}

	// Find first matching image
	for _, image := range images {
		if image.OS != "linux" {
			continue
		}
		if image.ID == imageName || image.Name == imageName {
			return &image, nil
		}
	}

	return nil, &APIError{
		Operation: "GetImageByName",
		Resource:  fmt.Sprintf("%s/%s", region, imageName),
		Message:   fmt.Sprintf("Image '%s' does not exist in region %s. Visit %s for available images", imageName, region, CustomerInterface),
	}
}

// ListSSHKeys returns a list of SSH keys for a given project in a given region
func (a *API) ListSSHKeys(ctx context.Context, projectID, region string) (SSHKeys, error) {
	var sshkeys SSHKeys
	resource := fmt.Sprintf("/cloud/project/%s/sshkey?region=%s", projectID, region)
	err := a.doWithRetry(ctx, "ListSSHKeys", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &sshkeys)
	})
	return sshkeys, err
}

// GetSshkeys returns a list of SSH keys for a given project in a given region (legacy method)
func (a *API) GetSshkeys(ctx context.Context, projectID, region string) (SSHKeys, error) {
	return a.ListSSHKeys(ctx, projectID, region)
}

// GetSshkeyByName returns the details of an SSH key given its name in a given region
func (a *API) GetSshkeyByName(ctx context.Context, projectID, region, sshKeyName string) (*SSHKey, error) {
	sshkeys, err := a.ListSSHKeys(ctx, projectID, region)
	if err != nil {
		return nil, err
	}

	// Find first matching SSH key
	for _, sshkey := range sshkeys {
		if sshkey.ID == sshKeyName || sshkey.Name == sshKeyName {
			return &sshkey, nil
		}
	}

	return nil, &APIError{
		Operation: "GetSshkeyByName",
		Resource:  fmt.Sprintf("%s/%s", region, sshKeyName),
		Message:   fmt.Sprintf("SSH key '%s' does not exist in region %s. Visit %s for available SSH keys", sshKeyName, region, CustomerInterface),
	}
}

// CreateSSHKey uploads a new public key with name and returns resulting object
func (a *API) CreateSSHKey(ctx context.Context, projectID, name, publicKey string) (*SSHKey, error) {
	var sshkey SSHKey
	req := SSHKeyReq{
		Name:      name,
		PublicKey: publicKey,
	}
	resource := fmt.Sprintf("/cloud/project/%s/sshkey", projectID)
	err := a.doWithRetry(ctx, "CreateSSHKey", resource, func() error {
		return a.client.PostWithContext(ctx, resource, req, &sshkey)
	})
	if err != nil {
		return nil, err
	}
	return &sshkey, nil
}

// CreateSshkey uploads a new public key with name and returns resulting object (legacy method)
func (a *API) CreateSshkey(ctx context.Context, projectID, name, pubkey string) (*SSHKey, error) {
	return a.CreateSSHKey(ctx, projectID, name, pubkey)
}

// DeleteSshkey deletes an existing SSH key
func (a *API) DeleteSshkey(ctx context.Context, projectID, keyID string) error {
	resource := fmt.Sprintf("/cloud/project/%s/sshkey/%s", projectID, keyID)
	err := a.doWithRetry(ctx, "DeleteSshkey", resource, func() error {
		return a.client.DeleteWithContext(ctx, resource, nil)
	})
	// Ignore 404 errors
	if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
		return nil
	}
	return err
}

// GetQuotas returns resource quotas for a given project in a region
func (a *API) GetQuotas(ctx context.Context, projectID, region string) (*Quota, error) {
	var quota Quota
	resource := fmt.Sprintf("/cloud/project/%s/quota?region=%s", projectID, region)
	err := a.doWithRetry(ctx, "GetQuotas", resource, func() error {
		var quotas []Quota
		if err := a.client.GetWithContext(ctx, resource, &quotas); err != nil {
			return err
		}
		if len(quotas) == 0 {
			return fmt.Errorf("no quota information available for region %s", region)
		}
		quota = quotas[0]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

// GetNetworks returns public & private networks for a given project
func (a *API) GetNetworks(ctx context.Context, projectID string, privateNet bool) (Networks, error) {
	var networks Networks
	var url string
	if privateNet {
		url = fmt.Sprintf("/cloud/project/%s/network/private", projectID)
	} else {
		url = fmt.Sprintf("/cloud/project/%s/network/public", projectID)
	}
	err := a.doWithRetry(ctx, "GetNetworks", url, func() error {
		return a.client.GetWithContext(ctx, url, &networks)
	})
	return networks, err
}

// GetPublicNetworkID returns the public network id for a given project
func (a *API) GetPublicNetworkID(ctx context.Context, projectID string) (string, error) {
	networks, err := a.GetNetworks(ctx, projectID, false)
	if err != nil {
		return "", err
	}
	if len(networks) == 0 {
		return "", &APIError{
			Operation: "GetPublicNetworkID",
			Resource:  projectID,
			Message:   "no public network found",
		}
	}
	return networks[0].ID, nil
}

// GetPrivateNetworkByName returns the details of a network given its name & project
func (a *API) GetPrivateNetworkByName(ctx context.Context, projectID, networkName string) (*Network, error) {
	networks, err := a.GetNetworks(ctx, projectID, true)
	if err != nil {
		return nil, err
	}

	// Find first matching network
	for _, network := range networks {
		if fmt.Sprintf("%d", network.VlanID) == networkName || network.Name == networkName {
			return &network, nil
		}
	}

	var networkNames []string
	for _, network := range networks {
		networkNames = append(networkNames, network.Name)
	}

	return nil, &APIError{
		Operation: "GetPrivateNetworkByName",
		Resource:  networkName,
		Message:   fmt.Sprintf("Invalid private network %s. Available networks: %s", networkName, strings.Join(networkNames, ", ")),
	}
}

// CreateInstance starts a new public cloud instance and returns resulting object
func (a *API) CreateInstance(ctx context.Context, projectID, name, pubkeyID, flavorID, imageID, region, userData string, networkIDs []string, monthlyBilling bool, securityGroup string) (*Instance, error) {
	var instance Instance
	req := InstanceReq{
		UserData:       userData,
		Name:           name,
		SshkeyID:       pubkeyID,
		FlavorID:       flavorID,
		ImageID:        imageID,
		Region:         region,
		MonthlyBilling: monthlyBilling,
		SecurityGroup:  securityGroup,
	}

	for _, v := range networkIDs {
		req.NetworkParams = append(req.NetworkParams, NetworkParam{ID: v})
	}

	resource := fmt.Sprintf("/cloud/project/%s/instance", projectID)
	err := a.doWithRetry(ctx, "CreateInstance", resource, func() error {
		return a.client.PostWithContext(ctx, resource, req, &instance)
	})
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetInstance finds a VM instance given an ID
func (a *API) GetInstance(ctx context.Context, projectID, instanceID string) (*Instance, error) {
	var instance Instance
	resource := fmt.Sprintf("/cloud/project/%s/instance/%s", projectID, instanceID)
	err := a.doWithRetry(ctx, "GetInstance", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &instance)
	})
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// StartInstance powers on a stopped instance
func (a *API) StartInstance(ctx context.Context, projectID, instanceID string) error {
	resource := fmt.Sprintf("/cloud/project/%s/instance/%s/start", projectID, instanceID)
	return a.doWithRetry(ctx, "StartInstance", resource, func() error {
		return a.client.PostWithContext(ctx, resource, nil, nil)
	})
}

// StopInstance powers off a running instance
func (a *API) StopInstance(ctx context.Context, projectID, instanceID string) error {
	resource := fmt.Sprintf("/cloud/project/%s/instance/%s/stop", projectID, instanceID)
	return a.doWithRetry(ctx, "StopInstance", resource, func() error {
		return a.client.PostWithContext(ctx, resource, nil, nil)
	})
}

// RebootInstance reboots an instance
func (a *API) RebootInstance(ctx context.Context, projectID, instanceID string, hard bool) error {
	rebootType := "soft"
	if hard {
		rebootType = "hard"
	}
	req := RebootReq{Type: rebootType}
	resource := fmt.Sprintf("/cloud/project/%s/instance/%s/reboot", projectID, instanceID)
	return a.doWithRetry(ctx, "RebootInstance", resource, func() error {
		return a.client.PostWithContext(ctx, resource, req, nil)
	})
}

// DeleteInstance stops and destroys a public cloud instance
func (a *API) DeleteInstance(ctx context.Context, projectID, instanceID string) error {
	resource := fmt.Sprintf("/cloud/project/%s/instance/%s", projectID, instanceID)
	err := a.doWithRetry(ctx, "DeleteInstance", resource, func() error {
		return a.client.DeleteWithContext(ctx, resource, nil)
	})
	// Ignore 404 errors
	if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
		return nil
	}
	return err
}

// ListMKSClusters returns managed kubernetes clusters in a given project
func (a *API) ListMKSClusters(ctx context.Context, projectID string) (MKSClusters, error) {
	var clusters MKSClusters
	resource := fmt.Sprintf("/cloud/project/%s/kube", projectID)
	err := a.doWithRetry(ctx, "ListMKSClusters", resource, func() error {
		return a.client.GetWithContext(ctx, resource, &clusters)
	})
	return clusters, err
}

// CreateMKSCluster creates a managed kubernetes cluster
func (a *API) CreateMKSCluster(ctx context.Context, projectID string, req MKSClusterCreateReq) (*MKSCluster, error) {
	var cluster MKSCluster
	resource := fmt.Sprintf("/cloud/project/%s/kube", projectID)
	err := a.doWithRetry(ctx, "CreateMKSCluster", resource, func() error {
		return a.client.PostWithContext(ctx, resource, req, &cluster)
	})
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

// DeleteMKSCluster deletes a managed kubernetes cluster
func (a *API) DeleteMKSCluster(ctx context.Context, projectID, clusterID string) error {
	resource := fmt.Sprintf("/cloud/project/%s/kube/%s", projectID, clusterID)
	err := a.doWithRetry(ctx, "DeleteMKSCluster", resource, func() error {
		return a.client.DeleteWithContext(ctx, resource, nil)
	})
	// Ignore 404 errors
	if apiErr, ok := err.(*APIError); ok && apiErr.IsNotFound() {
		return nil
	}
	return err
}

// CreateMKSNodePool creates a nodepool in a managed kubernetes cluster
func (a *API) CreateMKSNodePool(ctx context.Context, projectID, clusterID string, req MKSNodePoolCreateReq) (*MKSNodePool, error) {
	var nodePool MKSNodePool
	resource := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", projectID, clusterID)
	err := a.doWithRetry(ctx, "CreateMKSNodePool", resource, func() error {
		return a.client.PostWithContext(ctx, resource, req, &nodePool)
	})
	if err != nil {
		return nil, err
	}
	return &nodePool, nil
}

// ScaleMKSNodePool updates desired node count in a managed kubernetes nodepool
func (a *API) ScaleMKSNodePool(ctx context.Context, projectID, clusterID, nodePoolID string, desiredNodes int) error {
	req := MKSNodePoolScaleReq{DesiredNodes: desiredNodes}
	resource := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s", projectID, clusterID, nodePoolID)
	return a.doWithRetry(ctx, "ScaleMKSNodePool", resource, func() error {
		return a.client.PutWithContext(ctx, resource, req, nil)
	})
}
