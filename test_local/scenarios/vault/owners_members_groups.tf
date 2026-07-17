resource "freeipa_user" "vu" {
  username   = "acc_vault_user"
  first_name = "Vault"
  last_name  = "User"
}
resource "freeipa_group" "vg" {
  name = "acc-vault-group"
}
resource "freeipa_vault" "tv" {
  name = "acc-vault-groups"
}
resource "freeipa_vault_owner" "vo" {
  name         = freeipa_vault.tv.name
  owner_users  = [freeipa_user.vu.username]
  owner_groups = [freeipa_group.vg.name]
}
resource "freeipa_vault_member" "vm" {
  name   = freeipa_vault.tv.name
  users  = [freeipa_user.vu.username]
  groups = [freeipa_group.vg.name]
}
