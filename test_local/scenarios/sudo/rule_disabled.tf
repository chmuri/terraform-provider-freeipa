resource "freeipa_sudo_rule" "test" {
  name          = "acc-sudo-disabled"
  user_category = "all"
  host_category = "all"
  enabled       = false
}
