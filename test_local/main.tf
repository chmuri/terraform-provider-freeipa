terraform {
  required_providers {
    freeipa = {
      source  = "chmuri/freeipa"
      version = "~> 1.1.1"
    }
  }
}

provider "freeipa" {
  host     = "ipa.test.local"
  insecure = true
  username = "admin"
  password = "SecretAdminPassword123!"
}

# 1. Identity Resources (User, Group, Host, HostGroup)
resource "freeipa_user" "test_user" {
  username   = "jdoe"
  first_name = "John"
  last_name  = "Doe"
  email      = "jdoe@test.local"
  ssh_public_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPeM44/8ZkS2r6rU0Xj4W3g9t0C+7XU7k9j0N0Xj4W3g test@local"
  ]
}

resource "freeipa_group" "test_group" {
  name        = "test-group"
  description = "A test group created by Terraform"
  users       = [freeipa_user.test_user.username]
}

resource "freeipa_host" "test_host" {
  fqdn        = "webserver.test.local"
  description = "Test web server host"
}

resource "freeipa_hostgroup" "test_hostgroup" {
  cn          = "test-hostgroup"
  description = "Test group of web servers"
  hosts       = [freeipa_host.test_host.fqdn]
}

# 2. Host-Based Access Control (HBAC Service, Group, and Rule)
resource "freeipa_hbac_service" "test_hbac_service" {
  name        = "test-hbac-service"
  description = "Test HBAC Service for Custom Login"
}

resource "freeipa_hbac_service_group" "test_hbac_service_group" {
  name        = "test-hbac-svc-group"
  description = "Test HBAC Service Group"
  services    = [freeipa_hbac_service.test_hbac_service.name]
}

resource "freeipa_hbacrule" "test_hbac" {
  name             = "allow_test_group"
  description      = "Allow test group access to all hosts"
  host_category    = "all"
  service_category = "all"
  groups           = [freeipa_group.test_group.name]
}

# 3. Sudo Authorization (Sudo Command, Group, and Rule)
resource "freeipa_sudo_command" "test_sudo_command" {
  command     = "/usr/bin/systemctl"
  description = "Allow systemctl command"
}

resource "freeipa_sudo_command_group" "test_sudo_command_group" {
  name        = "test-sudo-command-group"
  description = "Test Sudo Command Group"
  commands    = [freeipa_sudo_command.test_sudo_command.command]
}

resource "freeipa_sudo_rule" "test_sudo_rule" {
  name                 = "test-sudo-rule"
  description          = "Test Sudo Rule"
  user_category        = "all"
  host_category        = "all"
  allow_command_groups = [freeipa_sudo_command_group.test_sudo_command_group.name]
}

# 4. Domain Name System (DNS Zone and Record) — requires DNS plugin in FreeIPA
resource "freeipa_dns_zone" "test_dns_zone" {
  zone_name                = "corp.test.local"
  authoritative_nameserver = "ipa.test.local."
  admin_email              = "hostmaster.test.local."
}

resource "freeipa_dns_record" "test_dns_record" {
  zone_name    = freeipa_dns_zone.test_dns_zone.zone_name
  name         = "www"
  record_type  = "A"
  record_value = "10.5.0.20"
}

# 5. Role-Based Access Control (Privilege and Role)
resource "freeipa_privilege" "test_privilege" {
  name        = "test-privilege"
  description = "Test Privilege"
}

resource "freeipa_role" "test_role" {
  name        = "test-role"
  description = "Test Role"
  privileges  = [freeipa_privilege.test_privilege.name]
  users       = [freeipa_user.test_user.username]
}

# 6. NIS Netgroup
resource "freeipa_netgroup" "test_netgroup" {
  name        = "test-netgroup"
  description = "Test Netgroup"
}

# 7. Password Policy
resource "freeipa_password_policy" "test_password_policy" {
  name      = freeipa_group.test_group.name
  minlength = 8
  maxlife   = 90
}

# 8. Secure Skiff Storage (Vault, Owner, and Member) — requires KRA in FreeIPA
resource "freeipa_vault" "test_vault" {
  name        = "test-vault"
  description = "Test Vault"
  type        = "standard"
}

resource "freeipa_vault_owner" "test_vault_owner" {
  name        = freeipa_vault.test_vault.name
  owner_users = [freeipa_user.test_user.username]
}

resource "freeipa_vault_member" "test_vault_member" {
  name  = freeipa_vault.test_vault.name
  users = [freeipa_user.test_user.username]
}

# Outputs
output "host_otp" {
  value     = freeipa_host.test_host.password
  sensitive = true
}
