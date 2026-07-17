resource "freeipa_privilege" "dsp" {
  name = "acc-ds-priv"
}

data "freeipa_privilege" "test" {
  name = freeipa_privilege.dsp.name
}

output "priv_name" { value = data.freeipa_privilege.test.name }
