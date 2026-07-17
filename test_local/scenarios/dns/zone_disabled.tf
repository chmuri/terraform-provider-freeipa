resource "freeipa_dns_zone" "test" {
  zone_name                = "acc-disable.test.local"
  authoritative_nameserver = "ipa.test.local."
  enabled                  = false
}
