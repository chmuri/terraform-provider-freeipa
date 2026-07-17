resource "freeipa_role" "dsr" {
  name = "acc-ds-role"
}

data "freeipa_role" "test" {
  name = freeipa_role.dsr.name
}

output "role_name" { value = data.freeipa_role.test.name }
