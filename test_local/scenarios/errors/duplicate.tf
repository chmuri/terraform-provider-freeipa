resource "freeipa_user" "dup" {
  username   = "acc_dup_user"
  first_name = "Dup1"
  last_name  = "User"
}

resource "freeipa_user" "test" {
  username   = "acc_dup_user"
  first_name = "Dup2"
  last_name  = "User"
}
