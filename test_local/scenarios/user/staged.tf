resource "freeipa_user" "test" {
  username   = "acc_user_staged"
  first_name = "Staged"
  last_name  = "User"
  staged     = true
}
