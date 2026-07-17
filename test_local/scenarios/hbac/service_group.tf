resource "freeipa_hbac_service" "svc" {
  name = "acc-svc-for-grp"
}
resource "freeipa_hbac_service_group" "test" {
  name     = "acc-svc-grp"
  services = [freeipa_hbac_service.svc.name]
}
