# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please **do not** open a public issue.  
Instead, report it privately via GitHub Security Advisories:

1. Go to [Security Advisories](https://github.com/chmuri/terraform-provider-freeipa/security/advisories/new)
2. Click **"Report a vulnerability"**
3. Describe the issue with as much detail as possible

You can also email the maintainer directly at `k.chmurzynski@gmail.com`.

### What to include

- Description of the vulnerability
- Steps to reproduce
- Affected versions
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: within 48 hours
- **Status update**: within 5 business days
- **Fix release**: depends on severity (critical: ASAP, low: next release)

## Security Considerations

### Credential Handling

This provider manages FreeIPA credentials. Never commit credentials to version control:

- Use environment variables (`FREEIPA_HOST`, `FREEIPA_USERNAME`, `FREEIPA_PASSWORD`)
- Use Terraform variables with `sensitive = true`
- Never hardcode passwords in HCL configurations

### TLS / SSL

FreeIPA communication uses HTTPS with TLS. In production:

- Set `insecure = false` (default)
- Ensure the FreeIPA server uses a valid TLS certificate
- Do not bypass certificate verification in production

### Kerberos Authentication

When using Kerberos keytab authentication:

- Protect the keytab file with appropriate file permissions (`0600`)
- Use `krb5_conf` to specify a custom Kerberos configuration
- Rotate keytab credentials regularly

### Dependency Management

All Go dependencies are managed via `go.mod` and `go.sum`.  
Run `go mod verify` to ensure module integrity before building.

Dependencies:
- `hashicorp/terraform-plugin-framework` — Terraform provider SDK
- `hashicorp/terraform-plugin-testing` — Acceptance test framework
- `jcmturner/gokrb5` — Kerberos authentication (SPNEGO)

### Security Updates

Security patches are released as patch versions (e.g., 1.0.4 → 1.0.5).  
Subscribe to [GitHub releases](https://github.com/chmuri/terraform-provider-freeipa/releases) for notifications.
