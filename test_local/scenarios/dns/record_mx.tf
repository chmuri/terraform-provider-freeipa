resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-mx.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "@"
  record_type  = "MX"
  record_value = "10 mail.acc-mx.test.local."
}
