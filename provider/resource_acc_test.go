package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const providerConfig = `
provider "freeipa" {
  host     = %[1]q
  username = %[2]q
  password = %[3]q
  insecure = %[4]s
}
`

func accProviderConfig() string {
	return fmt.Sprintf(providerConfig,
		os.Getenv("FREEIPA_HOST"),
		os.Getenv("FREEIPA_USERNAME"),
		os.Getenv("FREEIPA_PASSWORD"),
		os.Getenv("FREEIPA_INSECURE"),
	)
}

func skipIfNotAcc(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance test; set TF_ACC=1 to run")
	}
}

// ─────────────────────────────────────────────────────────────
// User resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_User_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	username := "acc_test_user"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + fmt.Sprintf(`
resource "freeipa_user" "test" {
  username   = %[1]q
  first_name = "Acc"
  last_name  = "Test"
  email      = "acc_user@test.local"
}`, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "username", username),
					resource.TestCheckResourceAttr("freeipa_user.test", "first_name", "Acc"),
					resource.TestCheckResourceAttr("freeipa_user.test", "last_name", "Test"),
					resource.TestCheckResourceAttr("freeipa_user.test", "email", "acc_user@test.local"),
					resource.TestCheckResourceAttr("freeipa_user.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("freeipa_user.test", "cn"),
					resource.TestCheckResourceAttrSet("freeipa_user.test", "uid"),
					resource.TestCheckResourceAttrSet("freeipa_user.test", "gid"),
				),
			},
			{
				Config: accProviderConfig() + fmt.Sprintf(`
resource "freeipa_user" "test" {
  username   = %[1]q
  first_name = "Acc"
  last_name  = "Test"
  email      = "acc_user_updated@test.local"
  city       = "TestCity"
  phone      = "123456789"
}`, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "email", "acc_user_updated@test.local"),
					resource.TestCheckResourceAttr("freeipa_user.test", "city", "TestCity"),
					resource.TestCheckResourceAttr("freeipa_user.test", "phone", "123456789"),
				),
			},
			{
				ResourceName:            "freeipa_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAcc_User_Options(t *testing.T) {
	skipIfNotAcc(t)
	tests := []struct {
		name   string
		config string
		checks []resource.TestCheckFunc
	}{
		{
			"disabled",
			accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc_disabled"
  first_name = "Acc"
  last_name  = "Disabled"
  enabled    = false
}`,
			[]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("freeipa_user.test", "enabled", "false"),
			},
		},
		{
			"with_address_fields",
			accProviderConfig() + `
resource "freeipa_user" "test" {
  username    = "acc_addr"
  first_name  = "Acc"
  last_name   = "Addr"
  street      = "123 Main St"
  city        = "Testville"
  state       = "TS"
  postalcode  = "12345"
}`,
			[]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("freeipa_user.test", "street", "123 Main St"),
				resource.TestCheckResourceAttr("freeipa_user.test", "city", "Testville"),
				resource.TestCheckResourceAttr("freeipa_user.test", "state", "TS"),
				resource.TestCheckResourceAttr("freeipa_user.test", "postalcode", "12345"),
			},
		},
		{
			"custom_shell_homedir",
			accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc_shell"
  first_name = "Acc"
  last_name  = "Shell"
  shell      = "/bin/bash"
  homedir    = "/custom/home"
  gecos      = "Custom GECOS"
}`,
			[]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("freeipa_user.test", "shell", "/bin/bash"),
				resource.TestCheckResourceAttr("freeipa_user.test", "homedir", "/custom/home"),
				resource.TestCheckResourceAttr("freeipa_user.test", "gecos", "Custom GECOS"),
			},
		},
		{
			"with_ssh_keys",
			accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc_ssh"
  first_name = "Acc"
  last_name  = "SSH"
  ssh_public_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPeM44/8ZkS2r6rU0Xj4W3g9t0C+7XU7k9j0N0Xj4W3g test@acc"
  ]
}`,
			[]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("freeipa_user.test", "ssh_public_keys.#", "1"),
			},
		},
		{
			"full_profile",
			accProviderConfig() + `
resource "freeipa_user" "test" {
  username    = "acc_full"
  first_name  = "Acc"
  last_name   = "Full"
  email       = "acc_full@test.local"
  cn          = "Acc Full"
  displayname = "Acc F."
  initials    = "AF"
  title       = "Engineer"
  orgunit     = "Engineering"
  mobile      = "5550001"
  pager       = "5550002"
  fax         = "5550003"
  manager     = ""
  carlicense  = "ABC123"
  preferredlanguage = "en"
}`,
			[]resource.TestCheckFunc{
				resource.TestCheckResourceAttr("freeipa_user.test", "cn", "Acc Full"),
				resource.TestCheckResourceAttr("freeipa_user.test", "displayname", "Acc F."),
				resource.TestCheckResourceAttr("freeipa_user.test", "title", "Engineer"),
				resource.TestCheckResourceAttr("freeipa_user.test", "mobile", "5550001"),
				resource.TestCheckResourceAttr("freeipa_user.test", "preferredlanguage", "en"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{Config: tt.config, Check: resource.ComposeAggregateTestCheckFunc(tt.checks...)},
				},
			})
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Group resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_Group_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "test" {
  name        = "acc-test-group"
  description = "Test group"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_group.test", "name", "acc-test-group"),
					resource.TestCheckResourceAttr("freeipa_group.test", "description", "Test group"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "test" {
  name        = "acc-test-group"
  description = "Updated test group"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_group.test", "description", "Updated test group"),
			},
			{ResourceName: "freeipa_group.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_Group_WithUsers(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "u1" {
  username   = "acc_group_u1"
  first_name = "G1"
  last_name  = "User"
}
resource "freeipa_user" "u2" {
  username   = "acc_group_u2"
  first_name = "G2"
  last_name  = "User"
}
resource "freeipa_group" "test" {
  name        = "acc-group-users"
  description = "Group with users"
  users       = [freeipa_user.u1.username, freeipa_user.u2.username]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_group.test", "name", "acc-group-users"),
					resource.TestCheckResourceAttr("freeipa_group.test", "users.#", "2"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Host resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_Host_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "test" {
  fqdn        = "acc-test-host.test.local"
  description = "Test host"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_host.test", "fqdn", "acc-test-host.test.local"),
					resource.TestCheckResourceAttr("freeipa_host.test", "description", "Test host"),
					resource.TestCheckResourceAttrSet("freeipa_host.test", "password"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "test" {
  fqdn        = "acc-test-host.test.local"
  description = "Updated host"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_host.test", "description", "Updated host"),
			},
			{
				ResourceName:            "freeipa_host.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "force"},
			},
		},
	})
}

func TestAcc_Host_Options(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "test" {
  fqdn        = "acc-host-opts.test.local"
  description = "Host with options"
  locality    = "Building A"
  location    = "Room 101"
  platform    = "x86_64"
  os          = "Linux"
  mac_address = "00:11:22:33:44:55"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_host.test", "locality", "Building A"),
					resource.TestCheckResourceAttr("freeipa_host.test", "location", "Room 101"),
					resource.TestCheckResourceAttr("freeipa_host.test", "platform", "x86_64"),
					resource.TestCheckResourceAttr("freeipa_host.test", "os", "Linux"),
					resource.TestCheckResourceAttr("freeipa_host.test", "mac_address", "00:11:22:33:44:55"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// HostGroup resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_HostGroup_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "hg_host" {
  fqdn        = "acc-hg-host.test.local"
  description = "Host for hostgroup"
}
resource "freeipa_hostgroup" "test" {
  cn          = "acc-test-hostgroup"
  description = "Test hostgroup"
  hosts       = [freeipa_host.hg_host.fqdn]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_hostgroup.test", "cn", "acc-test-hostgroup"),
					resource.TestCheckResourceAttr("freeipa_hostgroup.test", "description", "Test hostgroup"),
					resource.TestCheckResourceAttr("freeipa_hostgroup.test", "hosts.#", "1"),
				),
			},
			{ResourceName: "freeipa_hostgroup.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// HBAC Service tests
// ─────────────────────────────────────────────────────────────

func TestAcc_HbacSvc_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbac_service" "test" {
  name        = "acc-test-svc"
  description = "Test HBAC service"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_hbac_service.test", "name", "acc-test-svc"),
					resource.TestCheckResourceAttr("freeipa_hbac_service.test", "description", "Test HBAC service"),
				),
			},
			{ResourceName: "freeipa_hbac_service.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_HbacSvcGroup_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbac_service" "svc" {
  name = "acc-svc-for-group"
}
resource "freeipa_hbac_service_group" "test" {
  name        = "acc-test-svcgrp"
  description = "Test HBAC svc group"
  services    = [freeipa_hbac_service.svc.name]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_hbac_service_group.test", "name", "acc-test-svcgrp"),
					resource.TestCheckResourceAttr("freeipa_hbac_service_group.test", "services.#", "1"),
				),
			},
			{ResourceName: "freeipa_hbac_service_group.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_HbacRule_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbacrule" "test" {
  name             = "acc-test-hbac"
  description      = "Test HBAC rule"
  host_category    = "all"
  service_category = "all"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "name", "acc-test-hbac"),
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "host_category", "all"),
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "service_category", "all"),
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "enabled", "true"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_hbacrule" "test" {
  name             = "acc-test-hbac"
  description      = "Updated HBAC rule"
  host_category    = "all"
  service_category = "all"
  enabled          = false
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hbacrule.test", "description", "Updated HBAC rule"),
			},
		},
	})
}

func TestAcc_HbacRule_WithUsersGroups(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "hu" {
  username   = "acc_hbac_user"
  first_name = "Hbac"
  last_name  = "User"
}
resource "freeipa_group" "hg" {
  name = "acc-hbac-group"
}
resource "freeipa_hbacrule" "test" {
  name             = "acc-hbac-members"
  description      = "HBAC with members"
  users            = [freeipa_user.hu.username]
  groups           = [freeipa_group.hg.name]
  host_category    = "all"
  service_category = "all"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "name", "acc-hbac-members"),
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "users.#", "1"),
					resource.TestCheckResourceAttr("freeipa_hbacrule.test", "groups.#", "1"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Sudo resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_SudoCommand_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "test" {
  command     = "/usr/bin/acc_test_cmd"
  description = "Test sudo command"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_command.test", "command", "/usr/bin/acc_test_cmd"),
					resource.TestCheckResourceAttr("freeipa_sudo_command.test", "description", "Test sudo command"),
				),
			},
			{ResourceName: "freeipa_sudo_command.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_SudoCommandGroup_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "sc" {
  command = "/usr/bin/acc_cmd_for_group"
}
resource "freeipa_sudo_command_group" "test" {
  name        = "acc-test-cmdgrp"
  description = "Test sudo cmd group"
  commands    = [freeipa_sudo_command.sc.command]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_command_group.test", "name", "acc-test-cmdgrp"),
					resource.TestCheckResourceAttr("freeipa_sudo_command_group.test", "commands.#", "1"),
				),
			},
			{ResourceName: "freeipa_sudo_command_group.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_SudoRule_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_rule" "test" {
  name          = "acc-test-sudorule"
  description   = "Test sudo rule"
  user_category = "all"
  host_category = "all"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "name", "acc-test-sudorule"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "user_category", "all"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "host_category", "all"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_rule" "test" {
  name          = "acc-test-sudorule"
  description   = "Updated sudo rule"
  user_category = "all"
  host_category = "all"
  enabled       = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "description", "Updated sudo rule"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "enabled", "false"),
				),
			},
			{ResourceName: "freeipa_sudo_rule.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_SudoRule_WithCommands(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "sc1" {
  command = "/usr/bin/acc_sr_cmd1"
}
resource "freeipa_sudo_command_group" "scg" {
  name     = "acc-sr-cmdgrp"
  commands = [freeipa_sudo_command.sc1.command]
}
resource "freeipa_sudo_rule" "test" {
  name                 = "acc-sr-with-cmds"
  description          = "Sudo rule with commands"
  allow_commands       = [freeipa_sudo_command.sc1.command]
  allow_command_groups = [freeipa_sudo_command_group.scg.name]
  user_category        = "all"
  host_category        = "all"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "name", "acc-sr-with-cmds"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "allow_commands.#", "1"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "allow_command_groups.#", "1"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// DNS resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_DnsZone_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "test" {
  zone_name                = "acc.test.local"
  authoritative_nameserver = "ipa.test.local."
  admin_email              = "admin@test.local."
  skip_overlap_check       = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_dns_zone.test", "zone_name", "acc.test.local"),
				),
			},
		},
	})
}

func TestAcc_DnsRecord_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-rec.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "www"
  record_type  = "A"
  record_value = "10.10.10.10"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "name", "www"),
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "record_type", "A"),
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "record_value", "10.10.10.10"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Role and Privilege tests
// ─────────────────────────────────────────────────────────────

func TestAcc_Privilege_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_privilege" "test" {
  name        = "acc-test-priv"
  description = "Test privilege"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_privilege.test", "name", "acc-test-priv"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_privilege" "test" {
  name        = "acc-test-priv"
  description = "Updated privilege"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_privilege.test", "description", "Updated privilege"),
			},
			{ResourceName: "freeipa_privilege.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_Role_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_privilege" "priv" {
  name = "acc-priv-for-role"
}
resource "freeipa_role" "test" {
  name        = "acc-test-role"
  description = "Test role"
  privileges  = [freeipa_privilege.priv.name]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_role.test", "name", "acc-test-role"),
					resource.TestCheckResourceAttr("freeipa_role.test", "privileges.#", "1"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "ru" {
  username   = "acc_role_user"
  first_name = "Role"
  last_name  = "User"
}
resource "freeipa_privilege" "priv" {
  name = "acc-priv-for-role"
}
resource "freeipa_role" "test" {
  name        = "acc-test-role"
  description = "Updated role"
  privileges  = [freeipa_privilege.priv.name]
  users       = [freeipa_user.ru.username]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_role.test", "users.#", "1"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Netgroup tests
// ─────────────────────────────────────────────────────────────

func TestAcc_Netgroup_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_netgroup" "test" {
  name        = "acc-test-netgroup"
  description = "Test netgroup"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "name", "acc-test-netgroup"),
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "description", "Test netgroup"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_netgroup" "test" {
  name        = "acc-test-netgroup"
  description = "Updated netgroup"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_netgroup.test", "description", "Updated netgroup"),
			},
			{ResourceName: "freeipa_netgroup.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_Netgroup_WithMembers(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "nu" {
  username   = "acc_ng_user"
  first_name = "Net"
  last_name  = "User"
}
resource "freeipa_host" "nh" {
  fqdn = "acc-ng-host.test.local"
}
resource "freeipa_netgroup" "test" {
  name        = "acc-ng-members"
  description = "Netgroup with members"
  users       = [freeipa_user.nu.username]
  hosts       = [freeipa_host.nh.fqdn]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "name", "acc-ng-members"),
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "users.#", "1"),
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "hosts.#", "1"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Password policy tests
// ─────────────────────────────────────────────────────────────

func TestAcc_PwPolicy_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "pg" {
  name = "acc-pwpolicy-group"
}
resource "freeipa_password_policy" "test" {
  name      = freeipa_group.pg.name
  minlength = 10
  maxlife   = 60
  history   = 5
  priority  = 10
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "name", "acc-pwpolicy-group"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "minlength", "10"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "maxlife", "60"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "pg" {
  name = "acc-pwpolicy-group"
}
resource "freeipa_password_policy" "test" {
  name      = freeipa_group.pg.name
  minlength = 12
  maxlife   = 90
  minclasses = 3
  maxfail    = 5
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "minlength", "12"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "maxlife", "90"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "minclasses", "3"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "maxfail", "5"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Vault resource tests
// ─────────────────────────────────────────────────────────────

func TestAcc_Vault_CRUD(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_vault" "test" {
  name        = "acc-test-vault"
  description = "Test vault"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_vault.test", "name", "acc-test-vault"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_vault" "test" {
  name        = "acc-test-vault"
  description = "Updated vault"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_vault.test", "description", "Updated vault"),
			},
		},
	})
}

func TestAcc_Vault_WithOwnersMembers(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "vu" {
  username   = "acc_vault_user"
  first_name = "Vault"
  last_name  = "User"
}
resource "freeipa_vault" "tv" {
  name        = "acc-vault-with-owner"
  description = "Vault with owner"
}
resource "freeipa_vault_owner" "vo" {
  name        = freeipa_vault.tv.name
  owner_users = [freeipa_user.vu.username]
}
resource "freeipa_vault_member" "vm" {
  name  = freeipa_vault.tv.name
  users = [freeipa_user.vu.username]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_vault_owner.vo", "owner_users.#", "1"),
					resource.TestCheckResourceAttr("freeipa_vault_member.vm", "users.#", "1"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Data source tests
// ─────────────────────────────────────────────────────────────

func TestAcc_DataSource_User(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "dsu" {
  username   = "acc_ds_user"
  first_name = "DS"
  last_name  = "User"
  email      = "ds_user@test.local"
}
data "freeipa_user" "test" {
  username = freeipa_user.dsu.username
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_user.test", "username", "acc_ds_user"),
					resource.TestCheckResourceAttr("data.freeipa_user.test", "email", "ds_user@test.local"),
				),
			},
		},
	})
}

func TestAcc_DataSource_Group(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "dsg" {
  name        = "acc-ds-group"
  description = "DS group"
}
data "freeipa_group" "test" {
  name = freeipa_group.dsg.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_group.test", "name", "acc-ds-group"),
				),
			},
		},
	})
}

func TestAcc_DataSource_Host(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "dsh" {
  fqdn = "acc-ds-host.test.local"
}
data "freeipa_host" "test" {
  fqdn = freeipa_host.dsh.fqdn
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_host.test", "fqdn", "acc-ds-host.test.local"),
				),
			},
		},
	})
}
