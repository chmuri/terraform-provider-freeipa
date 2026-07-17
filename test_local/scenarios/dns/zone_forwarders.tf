resource "freeipa_dns_zone" "test" {
  zone_name                = "acc-fwd.test.local"
  authoritative_nameserver = "ipa.test.local."
  forwarders               = ["8.8.8.8", "8.8.4.4"]
  forward_policy           = "first"
}
