resource "freeipa_host" "test" {
  fqdn        = "acc-basic.test.local"
  description = "Basic test host"
}
