resource "freeipa_user" "test" {
  username   = "acc_user_basic"
  first_name = "Basic"
  last_name  = "User"
  email      = "acc_user_basic@test.local"
}
