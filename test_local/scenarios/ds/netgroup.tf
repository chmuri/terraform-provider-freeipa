resource "freeipa_netgroup" "dsn" {
  name        = "acc-ds-ng"
  description = "DS netgroup"
}

data "freeipa_netgroup" "test" {
  name = freeipa_netgroup.dsn.name
}

output "ng_desc" { value = data.freeipa_netgroup.test.description }
