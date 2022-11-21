package provider

import (
	"fmt"
	"os/exec"
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
			// Read and recreate if missing testing
			{
				PreConfig: func() {
					// Delete the cluster created in the previous step.
					cmd := exec.Command("k3d", "cluster", "delete", "k3d-provider-test")
					err := cmd.Run()
					if err != nil {
						t.Error(err)
					}
				},
				Config: testAccClusterResourceConfig(k3dConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "k3d_config", k3dConfig),
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "name", "k3d-provider-test"),
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "id", "42f6391d7823ace7eb7d2d71ea5fb771"),
				),
			},
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

func testAccClusterResourceConfig(k3dConfig string) string {
	return fmt.Sprintf(`
resource "scaffolding_cluster" "test" {
	name = "k3d-provider-test"
  k3d_config = %[1]q
}
`, k3dConfig)
}
