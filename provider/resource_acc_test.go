package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("freeipa_user", &resource.Sweeper{
		Name: "freeipa_user",
		F:    sweepUsers,
	})
}

func sweepUsers(region string) error {
	cfg := &client.Config{
		Host: os.Getenv("FREEIPA_HOST"), Insecure: true,
		AuthMethod: client.AuthPassword,
		Username:   os.Getenv("FREEIPA_USERNAME"),
		Password:   os.Getenv("FREEIPA_PASSWORD"),
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		return err
	}
	if err := c.Login(); err != nil {
		log.Printf("[WARN] sweeper login failed: %v", err)
		return nil
	}
	ctx := context.Background()
	prefixes := []string{"acc_", "acc-test", "acc-ds"}
	for _, prefix := range prefixes {
		var result map[string]interface{}
		if err := c.Call(ctx, "user_find", []string{prefix}, nil, &result); err == nil {
			if res, ok := result["result"].(map[string]interface{}); ok {
				if members, ok := res["result"].([]interface{}); ok {
					for _, m := range members {
						if u, ok := m.(map[string]interface{}); ok {
							if uid, ok := u["uid"].([]interface{}); ok && len(uid) > 0 {
								username := fmt.Sprintf("%v", uid[0])
								c.Call(ctx, "user_del", []string{username}, nil, nil)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

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
			{ResourceName: "freeipa_netgroup.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"nisdomainname"}},
		},
	})
}

func TestAcc_Netgroup_WithMembers(t *testing.T) {
	skipIfNotAcc(t)
	// Clean up orphaned resources from previous test runs
	cleanupOrphan := func() {
		cfg := &client.Config{
			Host: os.Getenv("FREEIPA_HOST"), Insecure: true,
			AuthMethod: client.AuthPassword,
			Username:   os.Getenv("FREEIPA_USERNAME"),
			Password:   os.Getenv("FREEIPA_PASSWORD"),
		}
		c, _ := client.NewClient(cfg)
		if c != nil {
			c.Login()
			ctx := context.Background()
			c.Call(ctx, "netgroup_del", []string{"acc-ng-members"}, nil, nil)
			c.Call(ctx, "host_del", []string{"acc-ng-host.test.local"}, nil, nil)
			c.Call(ctx, "user_del", []string{"acc_ng_user"}, nil, nil)
		}
	}
	cleanupOrphan()
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

func TestAcc_DataSource_HostGroup(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hostgroup" "dshg" {
  cn = "acc-ds-hostgroup"
}
data "freeipa_hostgroup" "test" {
  name = freeipa_hostgroup.dshg.cn
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_hostgroup.test", "name", "acc-ds-hostgroup"),
				),
			},
		},
	})
}

func TestAcc_DataSource_HbacRule(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbacrule" "dshb" {
  name             = "acc-ds-hbacrule"
  host_category    = "all"
  service_category = "all"
}
data "freeipa_hbacrule" "test" {
  name = freeipa_hbacrule.dshb.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_hbacrule.test", "name", "acc-ds-hbacrule"),
				),
			},
		},
	})
}

func TestAcc_DataSource_DnsZone(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "dsdz" {
  zone_name                = "acc-ds.test.local"
  authoritative_nameserver = "ipa.test.local."
}
data "freeipa_dns_zone" "test" {
  zone_name = freeipa_dns_zone.dsdz.zone_name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_dns_zone.test", "zone_name", "acc-ds.test.local"),
				),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Extended coverage: Update + Import tests for resources missing them
// ─────────────────────────────────────────────────────────────

func TestAcc_HostGroup_Update(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hostgroup" "test" {
  cn          = "acc-hg-update"
  description = "Original"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hostgroup.test", "description", "Original"),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_hostgroup" "test" {
  cn          = "acc-hg-update"
  description = "Updated"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hostgroup.test", "description", "Updated"),
			},
		},
	})
}

func TestAcc_HbacSvc_Update(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbac_service" "test" {
  name        = "acc-svc-update"
  description = "Original"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hbac_service.test", "description", "Original"),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_hbac_service" "test" {
  name        = "acc-svc-update"
  description = "Updated"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hbac_service.test", "description", "Updated"),
			},
		},
	})
}

func TestAcc_HbacRule_Import(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbacrule" "test" {
  name             = "acc-hbac-import"
  host_category    = "all"
  service_category = "all"
}`,
			},
			{ResourceName: "freeipa_hbacrule.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_SudoCommand_Update(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "test" {
  command     = "/usr/bin/acc-upd-cmd"
  description = "Original"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_command.test", "description", "Original"),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "test" {
  command     = "/usr/bin/acc-upd-cmd"
  description = "Updated"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_command.test", "description", "Updated"),
			},
		},
	})
}

func TestAcc_SudoCommandGroup_Update(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command_group" "test" {
  name        = "acc-cmdgrp-update"
  description = "Original"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_command_group.test", "description", "Original"),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command_group" "test" {
  name        = "acc-cmdgrp-update"
  description = "Updated"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_command_group.test", "description", "Updated"),
			},
		},
	})
}

func TestAcc_Role_Import(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_role" "test" {
  name = "acc-role-import"
}`,
			},
			{ResourceName: "freeipa_role.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_PwPolicy_Import(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "ppi" {
  name = "acc-pwpolicy-import"
}
resource "freeipa_password_policy" "test" {
  name      = freeipa_group.ppi.name
  minlength = 10
  priority  = 5
}`,
			},
			{
				ResourceName:      "freeipa_password_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"minlife", "maxlife", "history", "minclasses", "maxfail", "failinterval", "lockouttime"},
			},
		},
	})
}

func TestAcc_Group_Nonposix_External(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "test" {
  name        = "acc-group-nonposix"
  description = "Non-POSIX group"
  nonposix    = true
}`,
				Check: resource.TestCheckResourceAttr("freeipa_group.test", "nonposix", "true"),
			},
		},
	})
}

func TestAcc_User_Staged(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc-staged"
  first_name = "Staged"
  last_name  = "User"
  staged     = true
}`,
				Check: resource.TestCheckResourceAttr("freeipa_user.test", "staged", "true"),
			},
		},
	})
}

func TestAcc_Host_SSHKeys(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "test" {
  fqdn            = "acc-host-ssh.test.local"
  description     = "Host with SSH keys"
  ssh_public_keys = [
    "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPeM44/8ZkS2r6rU0Xj4W3g9t0C+7XU7k9j0N0Xj4W3g host@test"
  ]
  ip_address      = "10.5.0.50"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_host.test", "ssh_public_keys.#", "1"),
					resource.TestCheckResourceAttr("freeipa_host.test", "ip_address", "10.5.0.50"),
				),
			},
		},
	})
}

func TestAcc_SudoRule_Options(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_rule" "test" {
  name          = "acc-sr-options"
  description   = "Sudo rule with options"
  user_category = "all"
  host_category = "all"
  order         = 50
  options       = ["!authenticate"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "order", "50"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "options.#", "1"),
				),
			},
		},
	})
}

func TestAcc_SudoRule_DenyCommands(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "dsc" {
  command = "/usr/bin/acc-deny-cmd"
}
resource "freeipa_sudo_rule" "test" {
  name           = "acc-sr-deny"
  description    = "Sudo rule with deny"
  user_category  = "all"
  host_category  = "all"
  deny_commands  = [freeipa_sudo_command.dsc.command]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "deny_commands.#", "1"),
			},
		},
	})
}

func TestAcc_SudoRule_Runas(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "sru" {
  username   = "acc_sru_user"
  first_name = "SRU"
  last_name  = "User"
}
resource "freeipa_group" "srg" {
  name = "acc-sru-group"
}
resource "freeipa_sudo_rule" "test" {
  name                = "acc-sr-runas"
  description         = "Sudo rule with run-as"
  user_category       = "all"
  host_category       = "all"
  runas_user_category = "all"
  runas_users         = [freeipa_user.sru.username]
  runas_groups        = [freeipa_group.srg.name]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "runas_user_category", "all"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "runas_users.#", "1"),
					resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "runas_groups.#", "1"),
				),
			},
		},
	})
}

func TestAcc_HbacRule_UserCategory(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbacrule" "test" {
  name             = "acc-hbac-usercat"
  description      = "HBAC with user category"
  user_category    = "all"
  host_category    = "all"
  service_category = "all"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hbacrule.test", "user_category", "all"),
			},
		},
	})
}

func TestAcc_PwPolicy_Lockout(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "plg" {
  name = "acc-plock-group"
}
resource "freeipa_password_policy" "test" {
  name         = freeipa_group.plg.name
  maxlife      = 60
  minlength    = 8
  priority     = 5
  maxfail      = 3
  lockouttime  = 600
  failinterval = 60
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "maxfail", "3"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "lockouttime", "600"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "failinterval", "60"),
				),
			},
		},
	})
}

func TestAcc_Group_MemberManagers(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "gmm" {
  username   = "acc_gmm_user"
  first_name = "GMM"
  last_name  = "User"
}
resource "freeipa_group" "test" {
  name            = "acc-group-mgr"
  description     = "Group with member managers"
  member_managers = [freeipa_user.gmm.username]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_group.test", "member_managers.#", "1"),
			},
		},
	})
}
