resource "freeipa_hbacrule" "test" {
  name             = "acc-hbac-basic"
  host_category    = "all"
  service_category = "all"
}
