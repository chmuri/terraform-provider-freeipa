resource "freeipa_host" "dsh" {
  fqdn = "acc-ds-host.test.local"
}

data "freeipa_host" "test" {
  fqdn = freeipa_host.dsh.fqdn
}

output "host_fqdn" { value = data.freeipa_host.test.fqdn }
