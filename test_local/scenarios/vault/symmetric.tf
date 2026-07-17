resource "freeipa_vault" "test" {
  name        = "acc-vault-sym"
  description = "Symmetric vault"
  type        = "symmetric"
}
