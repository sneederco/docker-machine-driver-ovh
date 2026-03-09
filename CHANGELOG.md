# Changelog

All notable changes to the OVHcloud Driver for Rancher will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation suite with Sneederco branding
- Detailed installation guide covering all platforms
- Complete configuration reference with examples
- Step-by-step Rancher integration guide
- Troubleshooting guide with common issues and solutions
- Architecture overview with technical diagrams
- Contributing guide with development setup and code standards

### Changed
- Updated README.md with modern, professional formatting
- Improved quick start guide (reduced to 5 clear steps)
- Enhanced configuration options table with better organization
- Updated all documentation links to new docs/ structure

### Documentation
- Added `docs/installation.md` - Detailed installation instructions
- Added `docs/configuration.md` - Complete flag reference
- Added `docs/rancher-integration.md` - Rancher setup guide
- Added `docs/troubleshooting.md` - Common issues and solutions
- Added `docs/architecture.md` - Technical architecture overview
- Added `CONTRIBUTING.md` - Contribution guidelines
- Added `CHANGELOG.md` - This file

## [Previous Versions]

### [2.0.0] - API Rewrite
- Complete rewrite of OVH API client
- Improved error handling and retry logic
- Better state management
- Enhanced logging and debugging

### [1.5.0] - MKS Support
- Experimental support for OVH Managed Kubernetes Service (MKS)
- Added MKS-specific configuration flags
- MKS cluster creation and deletion
- Basic nodepool management

### [1.4.0] - vRack Integration
- Full support for OVH vRack private networking
- VLAN configuration
- Private network interface creation
- Documentation for vRack setup

### [1.3.0] - Rancher Integration
- Improved compatibility with Rancher 2.x
- Better node provisioning flow
- Enhanced driver metadata
- Fixed scaling issues in Rancher

### [1.2.0] - Billing Options
- Added monthly billing support
- Billing period selection flag
- Cost optimization documentation

### [1.1.0] - Multi-Region Support
- Support for all OVHcloud regions
- Region fallback mechanism
- Improved region availability checking

### [1.0.0] - Initial Release
- Basic instance provisioning
- SSH key management
- Docker Machine integration
- OVH API authentication
- Support for common instance types
- Ubuntu, Debian, CoreOS support

## Version History Notes

### Upgrade Notes

#### Upgrading to 2.x
- API client completely rewritten - existing machines should continue to work
- New authentication flow - no action required if using ovh.conf
- Improved error messages - check logs if you have custom error handling

#### Upgrading to 1.5.x (MKS)
- MKS mode is experimental - use with caution in production
- MKS requires different workflow - see docs/configuration.md
- Existing single-instance machines unaffected

#### Upgrading to 1.4.x (vRack)
- vRack support requires additional setup - see docs/configuration.md
- Private networks need manual interface configuration post-creation
- No impact on existing public-network-only deployments

### Deprecation Notices

- None currently

### Security Updates

- None currently

## Release Planning

### Planned for Next Release

**Documentation & Branding:**
- ✅ Complete documentation rewrite
- ✅ Sneederco branding integration
- ✅ Comprehensive guides for all features

**Future Enhancements:**
- Custom UI component for Rancher
- Improved MKS integration
- Auto-scaling support
- Spot instance support (if OVH adds)
- Enhanced monitoring integration
- Backup and snapshot support
- Load balancer integration

### Long-term Roadmap

**Stability & Performance:**
- Reduce instance creation time
- Better handling of API rate limits
- Improved error recovery
- Enhanced state reconciliation

**Feature Additions:**
- Object storage integration
- Block storage management
- GPU instance support
- Multi-project management
- Cost tracking and reporting

**Developer Experience:**
- Better CLI output formatting
- Progress indicators for long operations
- Dry-run mode for testing
- Config file generation helper

**Enterprise Features:**
- LDAP/SSO integration for API keys
- Audit logging
- Compliance reporting
- Multi-tenancy support

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to contribute to this project.

## Release Process

1. Update CHANGELOG.md with new version
2. Update version in code
3. Create git tag: `git tag -a v1.x.x -m "Release v1.x.x"`
4. Push tag: `git push origin v1.x.x`
5. CI will build and publish release artifacts
6. Update release notes on GitHub

## Maintainers

**Current maintainer:** [Sneederco](https://github.com/sneederco)

**Original authors:**
- OVH Team
- Community contributors

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

[Back to README](README.md) | [Contributing](CONTRIBUTING.md)
