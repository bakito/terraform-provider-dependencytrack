package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRepositoryDataSource(t *testing.T) {
	server, cfg := testServer()
	defer server.Close()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: cfg + `
resource "dependencytrack_repository" "test" {
  url                     = "https://foo.bar"
  identifier              = "foo"
  enabled                 = true
  type                    = "GO_MODULES"
  username                = "foo"
  password                = "bar"
  authentication_required = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "id", testUUID),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "type", "GO_MODULES"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "identifier", "foo"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "url", "https://foo.bar"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "resolution_order", "1"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "authentication_required", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
				// The last_updated attribute does not exist in the HashiCups
				// API, therefore there is no value for it during import.
				ImportStateVerifyIgnore: []string{"last_updated"},
			},
			// Update and Read testing
			{
				Config: cfg + `
resource "dependencytrack_repository" "test" {
  url                     = "https://foo.org"
  identifier              = "foo"
  enabled                 = true
  type                    = "GO_MODULES"
  username                = "foo"
  password                = "bar"
  authentication_required = false
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "id", testUUID),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "type", "GO_MODULES"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "identifier", "foo"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "url", "https://foo.org"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "resolution_order", "1"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "authentication_required", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
