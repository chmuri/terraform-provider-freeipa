resource "freeipa_user" "u1" {
  username   = "acc_grp_usr1"
  first_name = "G1"
  last_name  = "User"
}
resource "freeipa_user" "u2" {
  username   = "acc_grp_usr2"
  first_name = "G2"
  last_name  = "User"
}
resource "freeipa_group" "test" {
  name        = "acc-group-users"
  description = "Group with user members"
  users       = [freeipa_user.u1.username, freeipa_user.u2.username]
}
