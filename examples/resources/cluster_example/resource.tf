provider "scaffolding" {
  # example configuration here
}

resource "scaffolding_cluster" "example" {
  configurable_attribute = "some-value"
  k3d_config             = <<EOF
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
