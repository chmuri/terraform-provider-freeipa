# Changelog

## v1.0.5 (2026-07-17)

### Added
- **SECURITY.md**: Vulnerability reporting process and security considerations
- **CONTRIBUTING.md**: Development setup, testing guide, release process
- **CODE_OF_CONDUCT.md**: Contributor Covenant v2.0
- **CODEOWNERS**: GitHub code ownership configuration
- **README badges**: Go version, license, test coverage, FreeIPA version

### Changed
- **docker-compose**: Enabled DNS plugin (`--setup-dns --no-forwarders`) for full test coverage
- **version**: 1.0.4 → 1.0.5

### Fixed
- **Makefile**: Fixed docker healthcheck to use proper Kerberos authentication

## v1.0.4 (2026-07-17)

### Fixed
- **sudo_rule**: Corrected API parameter names for FreeIPA 4.13+ (`usercat` → `usercategory`, `hostcat` → `hostcategory`, etc.)
- **sudo_rule**: JSON response field names updated to match FreeIPA 4.13 API
- **sudo_rule**: Handle `EmptyModlist` error (4202) in Update method
- **pwpolicy**: Corrected API parameter names (`maxlife` → `krbmaxpwdlife`, etc.)
- **pwpolicy**: Fixed `failinterval` → `krbpwdfailurecountinterval` (matching FreeIPA 4.13 internals)
- **pwpolicy**: Added default `cospriority=1` (required by FreeIPA API)
- **dns_zone**: Removed `idnssoamailaddr` from `dnszone_add` (not supported), set via `dnszone_mod`
- **vault**: Changed `vault_add` → `vault_add_internal` (correct FreeIPA 4.13 command)
- **vault_member / vault_owner**: Fixed API parameters (`users` → `user`, `groups` → `group`, etc.)
- **hbacsvcgroup**: Fixed JSON response field (`member_hbacservice` → `member_hbacsvc`)
- **role**: Fixed JSON response field (`member_privilege` → `memberof_privilege`)
- **netgroup**: Fixed JSON response fields (`member_user` → `memberuser_user`, etc.)
- **host / user / group / pwpolicy**: Handle unknown computed fields after Create/Update
- **user**: Added `UseStateForUnknown` plan modifier for all computed fields
- **user**: Skip unknown computed field comparisons in Update (prevents false validation errors)
- **hostgroup**: Fixed test configuration attribute (`name` → `cn`)

### Added
- **28 unit tests**: Schema validation for all 19 resources, 6 data sources, and provider
- **31 acceptance tests**: CRUD, option variants, membership scenarios, data sources
- **Read-after-Create**: For hbacsvcgroup, role, netgroup (ensures state consistency)
- **Data source acceptance tests**: user, group, host, hostgroup, hbacrule, dns_zone
- **Makefile**: Added `TESTARGS` support for `test-acc` target
- **Makefile**: Fixed docker healthcheck to use proper Kerberos authentication

### Changed
- **version**: 1.0.3 → 1.0.4
- **docker-compose**: Removed `--no-forwarders` (requires `--setup-dns` on newer FreeIPA)
- **test_local/main.tf**: Updated required provider version to 1.0.4

## v1.0.3

Initial release with support for 19 resources and 6 data sources.
