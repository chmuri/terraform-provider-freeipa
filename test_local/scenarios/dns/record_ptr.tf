resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-ptr.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "1"
  record_type  = "PTR"
  record_value = "host.acc-ptr.test.local."
}
