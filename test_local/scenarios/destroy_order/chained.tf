resource "freeipa_user" "u" {
  username   = "acc_do_user"
  first_name = "DO"
  last_name  = "User"
}

resource "freeipa_group" "g" {
  name  = "acc-do-group"
  users = [freeipa_user.u.username]
}

resource "freeipa_password_policy" "pp" {
  name      = freeipa_group.g.name
  minlength = 10
  priority  = 5
}

resource "freeipa_netgroup" "ng" {
  name  = "acc-do-ng"
  users = [freeipa_user.u.username]
}

resource "freeipa_privilege" "priv" {
  name = "acc-do-priv"
}

resource "freeipa_role" "role" {
  name       = "acc-do-role"
  privileges = [freeipa_privilege.priv.name]
  users      = [freeipa_user.u.username]
}

resource "freeipa_hbacrule" "hbac" {
  name             = "acc-do-hbac"
  host_category    = "all"
  service_category = "all"
  groups           = [freeipa_group.g.name]
}
