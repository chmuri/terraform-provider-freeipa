resource "freeipa_privilege" "priv" {
  name = "acc-priv-for-role"
}
resource "freeipa_role" "test" {
  name       = "acc-role-basic"
  privileges = [freeipa_privilege.priv.name]
}
