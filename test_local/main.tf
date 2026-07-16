terraform {
  required_providers {
    freeipa = {
      source  = "chmuri/freeipa"
      version = "1.0.0"
    }
  }
}

provider "freeipa" {
  host     = "ipa.test.local"
  insecure = true
  username = "admin"
  password = "SecretAdminPassword123!"
}

resource "freeipa_user" "test_user" {
  username   = "jdoe"
  first_name = "John"
  last_name  = "Doe"
  email      = "jdoe@test.local"
  ssh_public_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPeM44/8ZkS2r6rU0Xj4W3g9t0C+7XU7k9j0N0Xj4W3g test@local"
  ]
}

resource "freeipa_group" "test_group" {
  name        = "test-group"
  description = "A test group created by Terraform"
  users       = [freeipa_user.test_user.username]
}

resource "freeipa_host" "test_host" {
  fqdn        = "webserver.test.local"
  description = "Test web server host"
}

resource "freeipa_hbacrule" "test_hbac" {
  name             = "allow_test_group"
  description      = "Allow test group access to all hosts"
  host_category    = "all"
  service_category = "all"
  groups           = [freeipa_group.test_group.name]
}

output "host_otp" {
  value     = freeipa_host.test_host.password
  sensitive = true
}
