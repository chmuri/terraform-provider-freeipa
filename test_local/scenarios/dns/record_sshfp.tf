resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-sshfp.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "sshhost"
  record_type  = "SSHFP"
  record_value = "1 1 ABCDEF0123456789ABCDEF0123456789ABCDEF01"
}
