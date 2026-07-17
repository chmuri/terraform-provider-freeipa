resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-rec.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "www"
  record_type  = "A"
  record_value = "10.10.10.10"
}
