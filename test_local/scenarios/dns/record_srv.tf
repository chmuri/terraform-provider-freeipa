resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-srv.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "_ldap._tcp"
  record_type  = "SRV"
  record_value = "0 100 389 ldap.acc-srv.test.local."
}
