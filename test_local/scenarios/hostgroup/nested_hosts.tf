resource "freeipa_host" "h1" {
  fqdn = "acc-hg-host1.test.local"
}
resource "freeipa_host" "h2" {
  fqdn = "acc-hg-host2.test.local"
}
resource "freeipa_hostgroup" "parent" {
  cn          = "acc-hg-parent"
  description = "Parent hostgroup"
  hosts       = [freeipa_host.h1.fqdn]
}
resource "freeipa_hostgroup" "test" {
  cn          = "acc-hg-nested"
  description = "Nested hostgroup"
  hosts       = [freeipa_host.h2.fqdn]
  host_groups = [freeipa_hostgroup.parent.cn]
}
