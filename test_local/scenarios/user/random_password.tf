resource "freeipa_user" "test" {
  username        = "acc_user_random"
  first_name      = "Random"
  last_name       = "User"
  random_password = true
}
