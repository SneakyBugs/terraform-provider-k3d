package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const k3dConfig = `apiVersion: k3d.io/v1alpha4
kind: Simple

# Expose ports 80 via 8080 and 443 via 8443.
ports:
  - port: 3080:80
    nodeFilters:
      - loadbalancer
  - port: 3443:443
    nodeFilters:
      - loadbalancer

registries:
  create:
    name: dev
    hostPort: "5000"
`

func TestAccClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterResourceConfig(k3dConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "k3d_config", k3dConfig),
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "name", "k3d-provider-test"),
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "id", "42f6391d7823ace7eb7d2d71ea5fb771"),
				),
			},
			// ImportState testing
			// {
			// 	ResourceName:      "scaffolding_cluster.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// This is not normally necessary, but is here because this
			// 	// example code does not have an actual upstream service.
			// 	// Once the Read method is able to refresh information from
			// 	// the upstream service, this can be removed.
			// 	ImportStateVerifyIgnore: []string{"configurable_attribute"},
			// },
			// Update and Read testing
			// {
			// 	Config: testAccClusterResourceConfig("two"),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("scaffolding_cluster.test", "configurable_attribute", "two"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccClusterResourceConfig(configurableAttribute string) string {
	return fmt.Sprintf(`
resource "scaffolding_cluster" "test" {
	name = "k3d-provider-test"
  k3d_config = %[1]q
}
`, configurableAttribute)
}
