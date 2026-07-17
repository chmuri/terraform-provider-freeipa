resource "freeipa_sudo_command" "cmd" {
  command = "/usr/bin/acc-grp-cmd"
}
resource "freeipa_sudo_command_group" "test" {
  name     = "acc-cmd-grp"
  commands = [freeipa_sudo_command.cmd.command]
}
