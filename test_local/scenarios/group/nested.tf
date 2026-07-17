resource "freeipa_group" "parent" {
  name = "acc-parent"
}
resource "freeipa_group" "test" {
  name   = "acc-nested"
  groups = [freeipa_group.parent.name]
}
