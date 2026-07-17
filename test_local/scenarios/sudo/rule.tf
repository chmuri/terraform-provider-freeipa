resource "freeipa_sudo_rule" "test" {
  name          = "acc-sudo-basic"
  user_category = "all"
  host_category = "all"
}
