package main

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

// Default values for docker-machine-driver-ovh
const (
	DefaultSecurityGroup = "default"
	DefaultProjectName   = "docker-machine"
	DefaultEndpoint      = "ovh-us"
	DefaultFlavorName    = "b3-8"
	DefaultRegionName    = "US-EAST-VA-1"
	DefaultImageName     = "Ubuntu 24.04"
	DefaultSSHUserName   = "ubuntu"
	DefaultBillingPeriod = "hourly"
)

func main() {
	plugin.RegisterDriver(&Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: DefaultSSHUserName,
			SSHPort: 22,
		}})
}
