resource "freeipa_group" "dsg" {
  name        = "acc-ds-group"
  description = "DS group"
}

data "freeipa_group" "test" {
  name = freeipa_group.dsg.name
}

output "group_desc" { value = data.freeipa_group.test.description }
