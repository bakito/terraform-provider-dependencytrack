package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCoffeesDataSource(t *testing.T) {
	server, cfg := testServer()
	defer server.Close()
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: cfg + `data "dependencytrack_repositories" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of items
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.#", "1"),

					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.id", testExistingUUID),
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.type", "GO_MODULES"),
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.identifier", "proxy.golang.org"),
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.url", "https://proxy.golang.org"),
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.resolution_order", "1"),
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.enabled", "true"),
					resource.TestCheckResourceAttr("data.dependencytrack_repositories.test", "repositories.0.internal", "false"),
				),
			},
		},
	})
}
