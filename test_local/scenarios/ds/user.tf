resource "freeipa_user" "dsu" {
  username   = "acc_ds_user"
  first_name = "DS"
  last_name  = "User"
  email      = "ds_user@test.local"
}

data "freeipa_user" "test" {
  username = freeipa_user.dsu.username
}

output "user_full_name" { value = data.freeipa_user.test.full_name }
output "user_enabled" { value = data.freeipa_user.test.enabled }
