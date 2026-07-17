resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-ns.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "@"
  record_type  = "NS"
  record_value = "ns2.acc-ns.test.local."
}
