resource "freeipa_sudo_rule" "dss" {
  name          = "acc-ds-sr"
  user_category = "all"
  host_category = "all"
}

data "freeipa_sudo_rule" "test" {
  name = freeipa_sudo_rule.dss.name
}

output "sr_category" { value = data.freeipa_sudo_rule.test.user_category }
