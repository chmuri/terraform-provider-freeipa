resource "freeipa_hbacrule" "test" {
  name             = "acc-hbac-disabled"
  host_category    = "all"
  service_category = "all"
  enabled          = false
}
