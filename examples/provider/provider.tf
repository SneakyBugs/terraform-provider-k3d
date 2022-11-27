provider "k3d" {}

resource "k3d_cluster" "example" {
  name       = "example"
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

// Authentication with Kubernetes provider:
provider "kubernetes" {
  host                   = resource.k3d_cluster.example.host
  client_certificate     = base64decode(resource.k3d_cluster.example.client_certificate)
  client_key             = base64decode(resource.k3d_cluster.example.client_key)
  cluster_ca_certificate = base64decode(resource.k3d_cluster.example.cluster_ca_certificate)
}

// Authentication with Helm provider:
provider "helm" {
  kubernetes {
    host                   = resource.k3d_cluster.example.host
    client_certificate     = base64decode(resource.k3d_cluster.example.client_certificate)
    client_key             = base64decode(resource.k3d_cluster.example.client_key)
    cluster_ca_certificate = base64decode(resource.k3d_cluster.example.cluster_ca_certificate)
  }
}
