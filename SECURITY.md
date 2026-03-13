# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly:

1. **Do NOT open a public GitHub issue** for security vulnerabilities
2. Email the maintainers at: security@sneederco.com
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Any suggested fixes (optional)

We will acknowledge receipt within 48 hours and aim to provide a fix within 7 days for critical issues.

## Security Best Practices

When using this driver:

- **Never commit OVH API credentials** to version control
- Use environment variables or secure secret management for credentials
- Rotate your OVH API tokens periodically
- Use the minimum required API permissions (see README for scope requirements)
- Review cloud-init scripts before deployment

## Scope

This policy covers:
- `sneederco/docker-machine-driver-ovh` (this repo)
- `sneederco/ui-driver-ovh` (companion Rancher UI extension)
