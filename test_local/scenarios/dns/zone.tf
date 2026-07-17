resource "freeipa_dns_zone" "test" {
  zone_name                = "acc.test.local"
  authoritative_nameserver = "ipa.test.local."
  admin_email              = "admin.test.local."
}
