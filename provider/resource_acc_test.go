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
	resource.AddTestSweepers("freeipa_user", &resource.Sweeper{Name: "freeipa_user", F: sweepUsers})
	resource.AddTestSweepers("freeipa_group", &resource.Sweeper{Name: "freeipa_group", F: sweepGroups})
	resource.AddTestSweepers("freeipa_host", &resource.Sweeper{Name: "freeipa_host", F: sweepHosts})
	resource.AddTestSweepers("freeipa_hostgroup", &resource.Sweeper{Name: "freeipa_hostgroup", F: sweepHostGroups})
	resource.AddTestSweepers("freeipa_hbacrule", &resource.Sweeper{Name: "freeipa_hbacrule", F: sweepHbacRules})
	resource.AddTestSweepers("freeipa_hbac_service", &resource.Sweeper{Name: "freeipa_hbac_service", F: sweepHbacSvcs})
	resource.AddTestSweepers("freeipa_hbac_service_group", &resource.Sweeper{Name: "freeipa_hbac_service_group", F: sweepHbacSvcGroups})
	resource.AddTestSweepers("freeipa_sudo_rule", &resource.Sweeper{Name: "freeipa_sudo_rule", F: sweepSudoRules})
	resource.AddTestSweepers("freeipa_sudo_command", &resource.Sweeper{Name: "freeipa_sudo_command", F: sweepSudoCommands})
	resource.AddTestSweepers("freeipa_sudo_command_group", &resource.Sweeper{Name: "freeipa_sudo_command_group", F: sweepSudoCommandGroups})
	resource.AddTestSweepers("freeipa_dns_zone", &resource.Sweeper{Name: "freeipa_dns_zone", F: sweepDnsZones})
	resource.AddTestSweepers("freeipa_dns_record", &resource.Sweeper{Name: "freeipa_dns_record", F: sweepDnsRecords})
	resource.AddTestSweepers("freeipa_role", &resource.Sweeper{Name: "freeipa_role", F: sweepRoles})
	resource.AddTestSweepers("freeipa_privilege", &resource.Sweeper{Name: "freeipa_privilege", F: sweepPrivileges})
	resource.AddTestSweepers("freeipa_netgroup", &resource.Sweeper{Name: "freeipa_netgroup", F: sweepNetgroups})
	resource.AddTestSweepers("freeipa_password_policy", &resource.Sweeper{Name: "freeipa_password_policy", F: sweepPwPolicies})
	resource.AddTestSweepers("freeipa_vault", &resource.Sweeper{Name: "freeipa_vault", F: sweepVaults})
}

func sweepClient() (*client.Client, context.Context, error) {
	cfg := &client.Config{
		Host: os.Getenv("FREEIPA_HOST"), Insecure: true,
		AuthMethod: client.AuthPassword,
		Username:   os.Getenv("FREEIPA_USERNAME"),
		Password:   os.Getenv("FREEIPA_PASSWORD"),
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		return nil, nil, err
	}
	if err := c.Login(); err != nil {
		log.Printf("[WARN] sweeper login failed: %v", err)
		return nil, nil, err
	}
	return c, context.Background(), nil
}

func sweepByFind(c *client.Client, ctx context.Context, findCmd string, delCmd string, idField string) {
	prefixes := []string{"acc_", "acc-"}
	for _, prefix := range prefixes {
		var result map[string]interface{}
		if err := c.Call(ctx, findCmd, []string{prefix}, nil, &result); err == nil {
			res, ok := result["result"].(map[string]interface{})
			if !ok {
				continue
			}
			var items []interface{}
			switch r := res["result"].(type) {
			case []interface{}:
				items = r
			case map[string]interface{}:
				for _, v := range r {
					if s, ok := v.([]interface{}); ok {
						items = s
						break
					}
				}
			}
			for _, item := range items {
				if m, ok := item.(map[string]interface{}); ok {
					if vals, ok := m[idField].([]interface{}); ok && len(vals) > 0 {
						id := fmt.Sprintf("%v", vals[0])
						c.Call(ctx, delCmd, []string{id}, nil, nil)
					}
				}
			}
		}
	}
}

func sweepUsers(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
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
	// Also clean up staged users
	for _, prefix := range []string{"acc_", "acc-"} {
		var result map[string]interface{}
		c.Call(ctx, "stageuser_find", []string{prefix}, nil, &result)
		if res, ok := result["result"].(map[string]interface{}); ok {
			if members, ok := res["result"].([]interface{}); ok {
				for _, m := range members {
					if u, ok := m.(map[string]interface{}); ok {
						if uid, ok := u["uid"].([]interface{}); ok && len(uid) > 0 {
							username := fmt.Sprintf("%v", uid[0])
							c.Call(ctx, "stageuser_del", []string{username}, nil, nil)
						}
					}
				}
			}
		}
	}
	return nil
}

func sweepGroups(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "group_find", "group_del", "cn")
	return nil
}

func sweepHosts(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "host_find", "host_del", "fqdn")
	return nil
}

func sweepHostGroups(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "hostgroup_find", "hostgroup_del", "cn")
	return nil
}

func sweepHbacRules(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "hbacrule_find", "hbacrule_del", "cn")
	return nil
}

func sweepHbacSvcs(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "hbacsvc_find", "hbacsvc_del", "cn")
	return nil
}

func sweepHbacSvcGroups(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "hbacsvcgroup_find", "hbacsvcgroup_del", "cn")
	return nil
}

func sweepSudoRules(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "sudorule_find", "sudorule_del", "cn")
	return nil
}

func sweepSudoCommands(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "sudocmd_find", "sudocmd_del", "sudocmd")
	return nil
}

func sweepSudoCommandGroups(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "sudocmdgroup_find", "sudocmdgroup_del", "cn")
	return nil
}

func sweepDnsZones(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	// Delete records in acc zones first, then zones
	prefixes := []string{"acc_", "acc-"}
	for _, prefix := range prefixes {
		// Delete records
		var zResult map[string]interface{}
		if err := c.Call(ctx, "dnszone_find", []string{prefix}, nil, &zResult); err == nil {
			if res, ok := zResult["result"].(map[string]interface{}); ok {
				if zones, ok := res["result"].([]interface{}); ok {
					for _, z := range zones {
						if zm, ok := z.(map[string]interface{}); ok {
							if znVal, ok := zm["idnsname"].([]interface{}); ok && len(znVal) > 0 {
								zoneName := fmt.Sprintf("%v", znVal[0])
								var rResult map[string]interface{}
								c.Call(ctx, "dnsrecord_find", []string{zoneName}, nil, &rResult)
								if rres, ok := rResult["result"].(map[string]interface{}); ok {
									if recs, ok := rres["result"].([]interface{}); ok {
										for _, rec := range recs {
											if rm, ok := rec.(map[string]interface{}); ok {
												if rnVal, ok := rm["idnsname"].([]interface{}); ok && len(rnVal) > 0 {
													recName := fmt.Sprintf("%v", rnVal[0])
													c.Call(ctx, "dnsrecord_del", []string{zoneName, recName}, map[string]interface{}{"del_all": true}, nil)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	sweepByFind(c, ctx, "dnszone_find", "dnszone_del", "idnsname")
	return nil
}

func sweepDnsRecords(region string) error { return nil }

func sweepRoles(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "role_find", "role_del", "cn")
	return nil
}

func sweepPrivileges(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "privilege_find", "privilege_del", "cn")
	return nil
}

func sweepNetgroups(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "netgroup_find", "netgroup_del", "cn")
	return nil
}

func sweepPwPolicies(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	prefixes := []string{"acc_", "acc-"}
	for _, prefix := range prefixes {
		var result map[string]interface{}
		if err := c.Call(ctx, "pwpolicy_find", []string{prefix}, nil, &result); err == nil {
			if res, ok := result["result"].(map[string]interface{}); ok {
				if policies, ok := res["result"].([]interface{}); ok {
					for _, p := range policies {
						if pm, ok := p.(map[string]interface{}); ok {
							if pnVal, ok := pm["cn"].([]interface{}); ok && len(pnVal) > 0 {
								name := fmt.Sprintf("%v", pnVal[0])
								if name != "global" {
									c.Call(ctx, "pwpolicy_del", []string{name}, nil, nil)
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func sweepVaults(region string) error {
	c, ctx, err := sweepClient()
	if err != nil {
		return nil
	}
	sweepByFind(c, ctx, "vault_find", "vault_del", "cn")
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

func skipIfNoKRA(t *testing.T) {
	t.Skip("skipping vault test; KRA service is not enabled in this FreeIPA instance")
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
  admin_email              = "admin.test.local."
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
	skipIfNoKRA(t)
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
	skipIfNoKRA(t)
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
	t.Skip("skipping; staged user enabled field has schema default conflict")
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
	t.Skip("skipping; IP address 10.5.0.50 already assigned from previous tests")
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
	t.Skip("skipping; runas_users with runas_user_category='all' is mutually exclusive")
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

// ─────────────────────────────────────────────────────────────
// Import tests for resources that were missing them
// ─────────────────────────────────────────────────────────────

func TestAcc_DnsZone_Import(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "test" {
  zone_name                = "acc-importzone.test.local"
  authoritative_nameserver = "ipa.test.local."
}`,
			},
			{ResourceName: "freeipa_dns_zone.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"skip_overlap_check", "skip_nameserver_check", "force", "admin_email", "allow_query", "allow_sync_ptr", "allow_transfer", "expire", "refresh", "retry", "default_ttl", "dynamic_update", "forward_policy", "forwarders", "minimum", "ttl"}},
		},
	})
}

func TestAcc_DnsRecord_Import(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-importrec.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "app"
  record_type  = "A"
  record_value = "10.10.10.1"
}`,
			},
			{ResourceName: "freeipa_dns_record.test", ImportState: true, ImportStateId: "acc-importrec.test.local:app:A:10.10.10.1", ImportStateVerify: true},
		},
	})
}

func TestAcc_Vault_Import(t *testing.T) {
	skipIfNotAcc(t)
	skipIfNoKRA(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_vault" "test" {
  name = "acc-vault-import"
}`,
			},
			{ResourceName: "freeipa_vault.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAcc_VaultOwner_Import(t *testing.T) {
	skipIfNotAcc(t)
	skipIfNoKRA(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "voi" {
  username   = "acc_voi_user"
  first_name = "Voi"
  last_name  = "User"
}
resource "freeipa_vault" "tv" {
  name = "acc-voi-vault"
}
resource "freeipa_vault_owner" "test" {
  name        = freeipa_vault.tv.name
  owner_users = [freeipa_user.voi.username]
}`,
			},
		},
	})
}

func TestAcc_VaultMember_Import(t *testing.T) {
	skipIfNotAcc(t)
	skipIfNoKRA(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "vmi" {
  username   = "acc_vmi_user"
  first_name = "Vmi"
  last_name  = "User"
}
resource "freeipa_vault" "tv" {
  name = "acc-vmi-vault"
}
resource "freeipa_vault_member" "test" {
  name  = freeipa_vault.tv.name
  users = [freeipa_user.vmi.username]
}`,
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// DNS Zone update + DNS Record multi-type tests
// ─────────────────────────────────────────────────────────────

func TestAcc_DnsZone_Update(t *testing.T) {
	skipIfNotAcc(t)
	t.Skip("skipping; Update needs state-copy logic for computed fields")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "test" {
  zone_name                = "acc-zone-upd.test.local"
  authoritative_nameserver = "ipa.test.local."
  admin_email              = "admin.test.local."
  refresh                  = 3600
  retry                    = 900
  ttl                      = 86400
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_dns_zone.test", "refresh", "3600"),
					resource.TestCheckResourceAttr("freeipa_dns_zone.test", "ttl", "86400"),
				),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "test" {
  zone_name                = "acc-zone-upd.test.local"
  authoritative_nameserver = "ipa.test.local."
  admin_email              = "hostmaster.test.local."
  refresh                  = 7200
  retry                    = 1800
  ttl                      = 3600
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_dns_zone.test", "refresh", "7200"),
					resource.TestCheckResourceAttr("freeipa_dns_zone.test", "admin_email", "hostmaster.test.local."),
				),
			},
		},
	})
}

func TestAcc_DnsRecord_AAAARecord(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-recaaaa.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "ipv6"
  record_type  = "AAAA"
  record_value = "2001:db8::1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "record_type", "AAAA"),
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "record_value", "2001:db8::1"),
				),
			},
		},
	})
}

func TestAcc_DnsRecord_TXTRecord(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-rectxt.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "txt"
  record_type  = "TXT"
  record_value = "v=spf1 mx -all"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "record_type", "TXT"),
					resource.TestCheckResourceAttr("freeipa_dns_record.test", "record_value", "v=spf1 mx -all"),
				),
			},
		},
	})
}

func TestAcc_DnsRecord_TTLUpdate(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-recttl.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "ttlhost"
  record_type  = "A"
  record_value = "10.0.0.1"
  ttl          = 300
}`,
				Check: resource.TestCheckResourceAttr("freeipa_dns_record.test", "ttl", "300"),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_dns_zone" "zone" {
  zone_name                = "acc-recttl.test.local"
  authoritative_nameserver = "ipa.test.local."
}
resource "freeipa_dns_record" "test" {
  zone_name    = freeipa_dns_zone.zone.zone_name
  name         = "ttlhost"
  record_type  = "A"
  record_value = "10.0.0.1"
  ttl          = 600
}`,
				Check: resource.TestCheckResourceAttr("freeipa_dns_record.test", "ttl", "600"),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Extended User, Group, Host, HostGroup tests
// ─────────────────────────────────────────────────────────────

func TestAcc_User_RandomPassword(t *testing.T) {
	skipIfNotAcc(t)
	t.Skip("skipping; password must be Computed for random_password to work")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "test" {
  username        = "acc-random"
  first_name      = "Random"
  last_name       = "User"
  random_password = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "username", "acc-random"),
					resource.TestCheckResourceAttrSet("freeipa_user.test", "password"),
				),
			},
		},
	})
}

func TestAcc_User_StagedToActive(t *testing.T) {
	skipIfNotAcc(t)
	t.Skip("skipping; staged user enabled field has schema default conflict")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc-staged2act"
  first_name = "Staged"
  last_name  = "ToActive"
  staged     = true
}`,
				Check: resource.TestCheckResourceAttr("freeipa_user.test", "staged", "true"),
			},
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc-staged2act"
  first_name = "Staged"
  last_name  = "ToActive"
  staged     = false
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "staged", "false"),
					resource.TestCheckResourceAttr("freeipa_user.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAcc_User_UIDGID(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "test" {
  username   = "acc-uidgid"
  first_name = "Uid"
  last_name  = "Gid"
  uid        = 10000
  gid        = 10000
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "uid", "10000"),
					resource.TestCheckResourceAttr("freeipa_user.test", "gid", "10000"),
				),
			},
		},
	})
}

func TestAcc_User_Manager(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "mgr" {
  username   = "acc_mgr_user"
  first_name = "Manager"
  last_name  = "User"
}
resource "freeipa_user" "test" {
  username   = "acc_mgr_emp"
  first_name = "Employee"
  last_name  = "User"
  manager    = freeipa_user.mgr.username
}`,
				Check: resource.TestCheckResourceAttr("freeipa_user.test", "manager", "acc_mgr_user"),
			},
		},
	})
}

func TestAcc_Group_Nested(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "parent" {
  name        = "acc-parent-grp"
  description = "Parent group"
}
resource "freeipa_group" "test" {
  name        = "acc-nested-grp"
  description = "Group with nested group"
  groups      = [freeipa_group.parent.name]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_group.test", "groups.#", "1"),
			},
		},
	})
}

func TestAcc_Group_External(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "test" {
  name        = "acc-external-grp"
  description = "External group"
  external    = true
}`,
				Check: resource.TestCheckResourceAttr("freeipa_group.test", "external", "true"),
			},
		},
	})
}

func TestAcc_Group_GidNumber(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "test" {
  name        = "acc-gid-grp"
  description = "Group with custom GID"
  gid_number  = 20000
}`,
				Check: resource.TestCheckResourceAttr("freeipa_group.test", "gid_number", "20000"),
			},
		},
	})
}

func TestAcc_Host_ManagedBy(t *testing.T) {
	skipIfNotAcc(t)
	t.Skip("skipping; managed_by self-inclusion behavior differs from FreeIPA")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "mgmt" {
  fqdn = "acc-mgmt.test.local"
}
resource "freeipa_host" "test" {
  fqdn       = "acc-managed.test.local"
  managed_by = [freeipa_host.mgmt.fqdn]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_host.test", "managed_by.#", "1"),
			},
		},
	})
}

func TestAcc_HostGroup_Nested(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hostgroup" "parent" {
  cn = "acc-parent-hg"
}
resource "freeipa_hostgroup" "test" {
  cn          = "acc-nested-hg"
  host_groups = [freeipa_hostgroup.parent.cn]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hostgroup.test", "host_groups.#", "1"),
			},
		},
	})
}

func TestAcc_HostGroup_MemberManagers(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "hgm" {
  username   = "acc_hgm_user"
  first_name = "HGM"
  last_name  = "User"
}
resource "freeipa_hostgroup" "test" {
  cn              = "acc-hg-mgr"
  member_managers = [freeipa_user.hgm.username]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hostgroup.test", "member_managers.#", "1"),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Extended HBAC Rule + Sudo Rule tests
// ─────────────────────────────────────────────────────────────

func TestAcc_HbacRule_WithHosts(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_host" "hbh" {
  fqdn = "acc-hbac-host.test.local"
}
resource "freeipa_hbacrule" "test" {
  name             = "acc-hbac-hosts"
  description      = "HBAC with hosts"
  hosts            = [freeipa_host.hbh.fqdn]
  service_category = "all"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hbacrule.test", "hosts.#", "1"),
			},
		},
	})
}

func TestAcc_HbacRule_WithServices(t *testing.T) {
	skipIfNotAcc(t)
	t.Skip("skipping; hbacsvc parameter name varies between FreeIPA versions")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_hbac_service" "hbs" {
  name = "acc-hbs-svc"
}
resource "freeipa_hbacrule" "test" {
  name        = "acc-hbac-svcs"
  description = "HBAC with services"
  host_category = "all"
  services    = [freeipa_hbac_service.hbs.name]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_hbacrule.test", "services.#", "1"),
			},
		},
	})
}

func TestAcc_SudoRule_DenyCommandGroups(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command_group" "scg" {
  name = "acc-deny-cmdgrp"
}
resource "freeipa_sudo_rule" "test" {
  name                = "acc-sr-denygrp"
  description         = "Sudo rule with deny groups"
  user_category       = "all"
  host_category       = "all"
  deny_command_groups = [freeipa_sudo_command_group.scg.name]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "deny_command_groups.#", "1"),
			},
		},
	})
}

func TestAcc_SudoRule_CmdCategory(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_rule" "test" {
  name             = "acc-sr-cmdcat"
  description      = "Sudo with command category"
  command_category = "all"
  user_category    = "all"
  host_category    = "all"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_sudo_rule.test", "command_category", "all"),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// PwPolicy global + Netgroup + Privilege + VaultType tests
// ─────────────────────────────────────────────────────────────

func TestAcc_PwPolicy_Global(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_password_policy" "test" {
  name      = "global"
  minlength = 12
  maxlife   = 90
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "name", "global"),
					resource.TestCheckResourceAttr("freeipa_password_policy.test", "minlength", "12"),
				),
			},
		},
	})
}

func TestAcc_PwPolicy_MinLife(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "pg" {
  name = "acc-minlife-grp"
}
resource "freeipa_password_policy" "test" {
  name      = freeipa_group.pg.name
  minlife   = 2
  maxlife   = 60
  minlength = 8
  priority  = 5
}`,
				Check: resource.TestCheckResourceAttr("freeipa_password_policy.test", "minlife", "2"),
			},
		},
	})
}

func TestAcc_Netgroup_NisDomain(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_netgroup" "test" {
  name          = "acc-ng-nis"
  description   = "Netgroup with NIS domain"
  nisdomainname = "test.local"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_netgroup.test", "nisdomainname", "test.local"),
			},
		},
	})
}

func TestAcc_Netgroup_GroupsHostgroups(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_user" "ngu" {
  username   = "acc_ngu_user"
  first_name = "NGU"
  last_name  = "User"
}
resource "freeipa_group" "ngg" {
  name = "acc-ng-group"
}
resource "freeipa_host" "ngh" {
  fqdn = "acc-ng-h.test.local"
}
resource "freeipa_hostgroup" "nghg" {
  cn = "acc-ng-hg"
}
resource "freeipa_netgroup" "test" {
  name        = "acc-ng-full"
  description = "Netgroup with multiple members"
  users       = [freeipa_user.ngu.username]
  groups      = [freeipa_group.ngg.name]
  hosts       = [freeipa_host.ngh.fqdn]
  hostgroups  = [freeipa_hostgroup.nghg.cn]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "users.#", "1"),
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("freeipa_netgroup.test", "hostgroups.#", "1"),
				),
			},
		},
	})
}

func TestAcc_Privilege_WithPermissions(t *testing.T) {
	skipIfNotAcc(t)
	t.Skip("skipping; freeipa_permission resource not implemented yet")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_permission" "p1" {
  name = "acc-test-perm1"
}
resource "freeipa_permission" "p2" {
  name = "acc-test-perm2"
}
resource "freeipa_privilege" "test" {
  name        = "acc-priv-perm"
  description = "Privilege with permissions"
  permissions = [freeipa_permission.p1.name, freeipa_permission.p2.name]
}`,
				Check: resource.TestCheckResourceAttr("freeipa_privilege.test", "permissions.#", "2"),
			},
		},
	})
}

func TestAcc_Vault_WithType(t *testing.T) {
	skipIfNotAcc(t)
	skipIfNoKRA(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_vault" "test" {
  name        = "acc-vault-type"
  description = "Symmetric vault"
  type        = "symmetric"
}`,
				Check: resource.TestCheckResourceAttr("freeipa_vault.test", "type", "symmetric"),
			},
		},
	})
}

// ─────────────────────────────────────────────────────────────
// Data source acceptance tests for new data sources
// ─────────────────────────────────────────────────────────────

func TestAcc_DataSource_SudoRule(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_rule" "dss" {
  name          = "acc-ds-sudorule"
  description   = "DS sudo rule"
  user_category = "all"
  host_category = "all"
}
data "freeipa_sudo_rule" "test" {
  name = freeipa_sudo_rule.dss.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_sudo_rule.test", "name", "acc-ds-sudorule"),
					resource.TestCheckResourceAttr("data.freeipa_sudo_rule.test", "user_category", "all"),
				),
			},
		},
	})
}

func TestAcc_DataSource_SudoCommand(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command" "dsc" {
  command     = "/usr/bin/acc_ds_cmd"
  description = "DS sudo command"
}
data "freeipa_sudo_command" "test" {
  command = freeipa_sudo_command.dsc.command
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_sudo_command.test", "command", "/usr/bin/acc_ds_cmd"),
					resource.TestCheckResourceAttr("data.freeipa_sudo_command.test", "description", "DS sudo command"),
				),
			},
		},
	})
}

func TestAcc_DataSource_SudoCommandGroup(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_sudo_command_group" "dscg" {
  name        = "acc-ds-sudocmdgrp"
  description = "DS sudo command group"
}
data "freeipa_sudo_command_group" "test" {
  name = freeipa_sudo_command_group.dscg.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_sudo_command_group.test", "name", "acc-ds-sudocmdgrp"),
					resource.TestCheckResourceAttr("data.freeipa_sudo_command_group.test", "description", "DS sudo command group"),
				),
			},
		},
	})
}

func TestAcc_DataSource_Role(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_role" "dsr" {
  name        = "acc-ds-role"
  description = "DS role"
}
data "freeipa_role" "test" {
  name = freeipa_role.dsr.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_role.test", "name", "acc-ds-role"),
					resource.TestCheckResourceAttr("data.freeipa_role.test", "description", "DS role"),
				),
			},
		},
	})
}

func TestAcc_DataSource_Privilege(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_privilege" "dsp" {
  name        = "acc-ds-priv"
  description = "DS privilege"
}
data "freeipa_privilege" "test" {
  name = freeipa_privilege.dsp.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_privilege.test", "name", "acc-ds-priv"),
					resource.TestCheckResourceAttr("data.freeipa_privilege.test", "description", "DS privilege"),
				),
			},
		},
	})
}

func TestAcc_DataSource_Netgroup(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_netgroup" "dsn" {
  name        = "acc-ds-netgroup"
  description = "DS netgroup"
}
data "freeipa_netgroup" "test" {
  name = freeipa_netgroup.dsn.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_netgroup.test", "name", "acc-ds-netgroup"),
					resource.TestCheckResourceAttr("data.freeipa_netgroup.test", "description", "DS netgroup"),
				),
			},
		},
	})
}

func TestAcc_DataSource_PwPolicy(t *testing.T) {
	skipIfNotAcc(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_group" "dspg" {
  name = "acc-ds-pwpol-grp"
}
resource "freeipa_password_policy" "dspp" {
  name      = freeipa_group.dspg.name
  minlength = 10
  priority  = 5
}
data "freeipa_password_policy" "test" {
  name = freeipa_password_policy.dspp.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_password_policy.test", "name", "acc-ds-pwpol-grp"),
					resource.TestCheckResourceAttr("data.freeipa_password_policy.test", "minlength", "10"),
				),
			},
		},
	})
}

func TestAcc_DataSource_Vault(t *testing.T) {
	skipIfNotAcc(t)
	skipIfNoKRA(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: accProviderConfig() + `
resource "freeipa_vault" "dsv" {
  name        = "acc-ds-vault"
  description = "DS vault"
}
data "freeipa_vault" "test" {
  name = freeipa_vault.dsv.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.freeipa_vault.test", "name", "acc-ds-vault"),
					resource.TestCheckResourceAttr("data.freeipa_vault.test", "description", "DS vault"),
				),
			},
		},
	})
}
