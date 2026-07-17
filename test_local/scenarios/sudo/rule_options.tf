resource "freeipa_sudo_command" "cmd" {
  command = "/usr/bin/acc-opt-cmd"
}
resource "freeipa_sudo_rule" "test" {
  name           = "acc-sr-options"
  user_category  = "all"
  host_category  = "all"
  allow_commands = [freeipa_sudo_command.cmd.command]
  options        = ["!authenticate", "!requiretty"]
}
