resource "freeipa_group" "pg" {
  name = "acc-pwp-group"
}
resource "freeipa_password_policy" "test" {
  name      = freeipa_group.pg.name
  minlength = 10
  maxlife   = 60
  priority  = 5
}
