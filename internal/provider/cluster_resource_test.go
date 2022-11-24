package provider

import (
	"os/exec"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccClusterResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "name", "k3d-provider-test"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "host"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "client_certificate"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "client_key"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "cluster_ca_certificate"),
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
				Config: testAccClusterResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "name", "k3d-provider-test"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "host"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "client_certificate"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "client_key"),
					resource.TestCheckResourceAttrSet("scaffolding_cluster.test", "cluster_ca_certificate"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccClusterResourceUsage(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"kubernetes": {
				Source: "hashicorp/kubernetes",
			},
		},
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterResourceUsageConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_cluster.test", "name", "k3d-provider-test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccClusterResourceConfig() string {
	return `
resource "scaffolding_cluster" "test" {
	name = "k3d-provider-test"
  k3d_config = <<EOF
apiVersion: k3d.io/v1alpha4
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
EOF
}
`
}

func testAccClusterResourceUsageConfig() string {
	return `
resource "scaffolding_cluster" "test" {
	name = "k3d-provider-test"
  k3d_config = <<EOF
apiVersion: k3d.io/v1alpha4
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
EOF
}

provider "kubernetes" {
	host = resource.scaffolding_cluster.test.host
	client_certificate = base64decode(resource.scaffolding_cluster.test.client_certificate)
	client_key = base64decode(resource.scaffolding_cluster.test.client_key)
	cluster_ca_certificate = base64decode(resource.scaffolding_cluster.test.cluster_ca_certificate)
}

resource "kubernetes_config_map" "test" {
	metadata {
		name = "test"
	}

	data = {
		test = "acceptance"
	}
}
`
}
