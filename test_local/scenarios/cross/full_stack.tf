resource "freeipa_user" "u" {
  username   = "acc_full_user"
  first_name = "Full"
  last_name  = "Stack"
  email      = "acc_full@test.local"
}

resource "freeipa_group" "g" {
  name        = "acc-full-group"
  description = "Full stack group"
  users       = [freeipa_user.u.username]
}

resource "freeipa_host" "h" {
  fqdn        = "acc-full.test.local"
  description = "Full stack host"
}

resource "freeipa_hostgroup" "hg" {
  cn    = "acc-full-hg"
  hosts = [freeipa_host.h.fqdn]
}

resource "freeipa_hbacrule" "hbac" {
  name             = "acc-full-hbac"
  host_category    = "all"
  service_category = "all"
  users            = [freeipa_user.u.username]
}

resource "freeipa_sudo_command" "cmd" {
  command = "/usr/bin/acc-full"
}

resource "freeipa_sudo_rule" "sr" {
  name           = "acc-full-sudo"
  user_category  = "all"
  host_category  = "all"
  allow_commands = [freeipa_sudo_command.cmd.command]
}

resource "freeipa_dns_zone" "dz" {
  zone_name                = "acc-full.test.local"
  authoritative_nameserver = "ipa.test.local."
}

resource "freeipa_dns_record" "dr" {
  zone_name    = freeipa_dns_zone.dz.zone_name
  name         = "app"
  record_type  = "A"
  record_value = "10.3.3.3"
}

resource "freeipa_privilege" "priv" {
  name = "acc-full-priv"
}

resource "freeipa_role" "role" {
  name       = "acc-full-role"
  privileges = [freeipa_privilege.priv.name]
  users      = [freeipa_user.u.username]
}

resource "freeipa_netgroup" "ng" {
  name  = "acc-full-ng"
  users = [freeipa_user.u.username]
  hosts = [freeipa_host.h.fqdn]
}

resource "freeipa_password_policy" "pp" {
  name      = freeipa_group.g.name
  minlength = 10
  priority  = 5
}

resource "freeipa_vault" "v" {
  name = "acc-full-vault"
}

resource "freeipa_vault_owner" "vo" {
  name        = freeipa_vault.v.name
  owner_users = [freeipa_user.u.username]
}
