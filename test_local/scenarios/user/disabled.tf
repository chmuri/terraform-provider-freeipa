resource "freeipa_user" "test" {
  username   = "acc_u_disabled"
  first_name = "Disabled"
  last_name  = "User"
  enabled    = false
}
