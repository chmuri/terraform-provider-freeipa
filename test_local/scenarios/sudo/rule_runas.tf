resource "freeipa_user" "ru" {
  username   = "acc_runas_user"
  first_name = "RunAs"
  last_name  = "User"
}
resource "freeipa_group" "rg" {
  name = "acc-runas-group"
}
resource "freeipa_sudo_rule" "test" {
  name                = "acc-sr-runas-full"
  user_category       = "all"
  host_category       = "all"
  runas_user_category = "all"
  runas_users         = [freeipa_user.ru.username]
  runas_groups        = [freeipa_group.rg.name]
}
