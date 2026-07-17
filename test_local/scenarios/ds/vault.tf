resource "freeipa_vault" "dsv" {
  name        = "acc-ds-vault"
  description = "DS vault"
}

data "freeipa_vault" "test" {
  name = freeipa_vault.dsv.name
}

output "vault_desc" { value = data.freeipa_vault.test.description }
