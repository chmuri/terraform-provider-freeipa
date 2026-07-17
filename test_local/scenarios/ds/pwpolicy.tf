resource "freeipa_group" "pg" {
  name = "acc-ds-pwpol-grp"
}

resource "freeipa_password_policy" "pp" {
  name      = freeipa_group.pg.name
  minlength = 10
  priority  = 5
}

data "freeipa_password_policy" "test" {
  name = freeipa_password_policy.pp.name
}

output "pp_minlength" { value = data.freeipa_password_policy.test.minlength }
