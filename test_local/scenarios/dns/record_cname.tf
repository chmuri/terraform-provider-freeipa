resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-cname.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "arec" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "target"
  record_type  = "A"
  record_value = "10.5.5.5"
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "alias"
  record_type  = "CNAME"
  record_value = "target.acc-cname.test.local."
}
