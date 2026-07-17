resource "freeipa_user" "u" {
  username   = "acc_complex_do_user"
  first_name = "CDO"
  last_name  = "User"
}

resource "freeipa_group" "g" {
  name  = "acc-complex-do-group"
  users = [freeipa_user.u.username]
}

resource "freeipa_host" "h" {
  fqdn = "acc-complex-do.test.local"
}

resource "freeipa_hostgroup" "hg" {
  cn    = "acc-complex-do-hg"
  hosts = [freeipa_host.h.fqdn]
}

resource "freeipa_hbac_service" "svc" {
  name = "acc-complex-do-svc"
}

resource "freeipa_hbac_service_group" "svcg" {
  name     = "acc-complex-do-svcg"
  services = [freeipa_hbac_service.svc.name]
}

resource "freeipa_hbacrule" "hbac" {
  name             = "acc-complex-do-hbac"
  host_category    = "all"
  service_category = "all"
  users            = [freeipa_user.u.username]
  services         = [freeipa_hbac_service.svc.name]
}

resource "freeipa_sudo_command" "cmd" {
  command = "/usr/bin/acc-complex-do"
}

resource "freeipa_sudo_command_group" "cmg" {
  name     = "acc-complex-do-cmg"
  commands = [freeipa_sudo_command.cmd.command]
}

resource "freeipa_sudo_rule" "sr" {
  name                 = "acc-complex-do-sr"
  user_category        = "all"
  host_category        = "all"
  allow_command_groups = [freeipa_sudo_command_group.cmg.name]
}

resource "freeipa_dns_zone" "dz" {
  zone_name                = "acc-complex-do.test.local"
  authoritative_nameserver = "ipa.test.local."
}

resource "freeipa_dns_record" "dr" {
  zone_name    = freeipa_dns_zone.dz.zone_name
  name         = "svc"
  record_type  = "A"
  record_value = "10.10.10.50"
}

resource "freeipa_privilege" "priv" {
  name = "acc-complex-do-priv"
}

resource "freeipa_role" "role" {
  name       = "acc-complex-do-role"
  privileges = [freeipa_privilege.priv.name]
  groups     = [freeipa_group.g.name]
}

resource "freeipa_netgroup" "ng" {
  name  = "acc-complex-do-ng"
  hosts = [freeipa_host.h.fqdn]
}

resource "freeipa_password_policy" "pp" {
  name      = freeipa_group.g.name
  minlength = 10
  priority  = 5
}

resource "freeipa_vault" "v" {
  name = "acc-complex-do-vault"
}

resource "freeipa_vault_member" "vm" {
  name  = freeipa_vault.v.name
  users = [freeipa_user.u.username]
}
