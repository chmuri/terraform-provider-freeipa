# Changelog

## v1.1.0 (2026-07-17)

**Compatibility:** Tested with FreeIPA **4.13.1** (API version **2.257**) on AlmaLinux 10.

### Added
- **8 new data sources**: `freeipa_sudo_rule`, `freeipa_sudo_command`, `freeipa_sudo_command_group`, `freeipa_role`, `freeipa_privilege`, `freeipa_netgroup`, `freeipa_password_policy`, `freeipa_vault` (14 total data sources)
- **14 data source documentation** pages in `docs/data-sources/`
- **37 new acceptance tests** (73 total, 16 skipped for environment reasons):
  - Import tests: DnsZone, DnsRecord, Vault, VaultOwner, VaultMember
  - Extended scenarios: User (RandomPassword, StagedToActive, UID/GID, Manager), Group (Nested, External, GidNumber), Host (ManagedBy), HostGroup (Nested, MemberManagers), HBAC Rule (Hosts, Services), Sudo Rule (DenyCommandGroups, CommandCategory), DNS (Zone Update, AAAA record, TXT record, TTL update), PwPolicy (Global, MinLife), Netgroup (NisDomain, GroupsHostgroups), Privilege (WithPermissions), Vault (WithType)
  - Data source acceptance: SudoRule, SudoCommand, SudoCommandGroup, Role, Privilege, Netgroup, PwPolicy, Vault
- **Extended sweeper**: 16 resource types (was 1 - user only). Covers: user, group, host, hostgroup, hbacrule, hbacsvc, hbacsvcgroup, sudorule, sudocmd, sudocmdgroup, dnszone, dnsrecord, role, privilege, netgroup, pwpolicy, vault
- **Integration test infrastructure**: 50 Terraform scenario files in `test_local/scenarios/`, 4 test scripts (`run_scenarios.sh`, `run_imports.sh`, `run_validate.sh`, `run_errors.sh`)

### Fixed
- **sudo_command**: Update method no longer makes unnecessary API calls; now compares plan with state before calling `sudocmd_mod`
- **dns_zone**: Fixed `parseIntVal(nil)` returning 0 instead of null in Read (affected all SOA integer fields)
- **dns_zone**: Update method now sends `nil` instead of empty string `""` for integer and boolean field clearing
- **dns_zone**: Strip trailing dot from `zone_name` returned by FreeIPA (prevents false replacement on refresh)
- **dns_zone**: Added `isEmptyModlistError` handling to `dnszone_mod`
- **dns_zone data source**: Same `parseIntVal(nil)` / `AllowSyncPtr` null fixes
- **hbacrule**: Added `isEmptyModlistError` handling to `hbacrule_mod`
- **sudorule**: Update method now sends `nil` instead of `""` for clearing `sudoorder` integer field
- **user**: `password` field now handled safely in Create (nulled when not random_password)
- **pwpolicy**: Global policy no longer sends `priority` parameter (not allowed on global policy)
- **pwpolicy**: Added `isEmptyModlistError` handling to `pwpolicy_mod`
- **netgroup**: Added `isEmptyModlistError` handling to `netgroup_mod`
- **role**: Added `isEmptyModlistError` handling to `role_mod`
- **privilege**: Added `isEmptyModlistError` handling to `privilege_mod`
- **vault**: Removed `type` parameter from `vault_add_internal` (not supported by API)
- **dns_record**: Set `ttl` to null in Create when not specified (prevents "unknown after apply" error)
- **hbacrule**: Fixed `hbacrule_add_service` API parameter name (`hbacservice` → `hbacsvc`)

### Changed
- **version**: 1.0.6 → 1.1.0
- Moved `difference()` function from `resource_group.go` to `helper.go` (shared utility)
- Renamed `hbacResourceModel` → `hbacRuleResourceModel` (consistent naming)
- Renamed `hostGroupResourceResult` → `FreeIPAHostGroupResult` (consistent naming)

## v1.0.6 (2026-07-17)

### Fixed
- **isNotFoundError**: Corrected error code from 4002 (DuplicateEntry) to 4001 (NotFound). This fixes stale-state detection for deleted resources, user staged fallback reads, and all delete/read error handling across the provider.
- **sudo_rule**: Fixed `runas_user_category` and `runas_group_category` API parameter names (`runasusercategory` → `ipasudorunasusercategory`, `runasgroupcategory` → `ipasudorunasgroupcategory`). Also fixed corresponding JSON response tags.

### Added
- **16 new acceptance tests** (46 total, 39 passing):
  - Update tests: HostGroup, HbacSvc, SudoCommand, SudoCommandGroup
  - Import tests: HbacRule, Role, PwPolicy
  - Option coverage: nonposix groups, SSH keys on hosts, sudo rule options, deny commands, run-as users/groups, user_category on HBAC rules, lockout policies, member managers
- **`parseStringVal`**: Added support for FreeIPA DNS name objects (`{"__dns_name__": "..."}` format)

### Changed
- **version**: 1.0.5 → 1.0.6

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
