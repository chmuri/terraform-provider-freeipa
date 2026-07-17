package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories are used to instantiate the provider during acceptance testing.
// The test framework will use this to inject the provider into the test execution.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"freeipa": func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6(New("test")())(), nil
	},
}

// TestAcc_User_BasicCRUD verifies the basic CRUD operations of the freeipa_user resource.
func TestAcc_User_BasicCRUD(t *testing.T) {
	// Only run acceptance tests if TF_ACC env variable is set
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping acceptance tests; TF_ACC is not set")
	}

	username := "acc_test_user"
	email := "acc_test_user@test.local"
	updatedEmail := "acc_test_user_updated@test.local"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the user resource
			{
				Config: testAccUserConfig(username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "username", username),
					resource.TestCheckResourceAttr("freeipa_user.test", "first_name", "Acc"),
					resource.TestCheckResourceAttr("freeipa_user.test", "last_name", "Test"),
					resource.TestCheckResourceAttr("freeipa_user.test", "email", email),
				),
			},
			// Step 2: Update the user resource (update email)
			{
				Config: testAccUserConfig(username, updatedEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("freeipa_user.test", "username", username),
					resource.TestCheckResourceAttr("freeipa_user.test", "email", updatedEmail),
				),
			},
			// Step 3: Import the user resource
			{
				ResourceName:            "freeipa_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // Ignore write-only/sensitive fields on verify
			},
		},
	})
}

func testAccUserConfig(username, email string) string {
	host := os.Getenv("FREEIPA_HOST")
	adminUser := os.Getenv("FREEIPA_USERNAME")
	adminPass := os.Getenv("FREEIPA_PASSWORD")
	insecure := os.Getenv("FREEIPA_INSECURE")

	return fmt.Sprintf(`
provider "freeipa" {
  host     = %q
  username = %q
  password = %q
  insecure = %s
}

resource "freeipa_user" "test" {
  username   = %q
  first_name = "Acc"
  last_name  = "Test"
  email      = %q
  password   = "SecretPassword123!"
}
`, host, adminUser, adminPass, insecure, username, email)
}
