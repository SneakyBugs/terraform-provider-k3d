# k3d Terraform Provider

> Manage development environments with Terraform

This provider manages development Kubernetes clusters in Docker with k3d.
Managing k3d clusters in Terraform allows you to provision development clusters
and deploy additional software (such as a database for your app) in a single action.

The idea is to automate everything needed before `tilt up` in a Kubernetes
development environment.

## Usage Example

This example creates a k3d cluster and deploys a Postgres instance.
It can be adapted to deploy any service your application requires for development.

Usage requires:
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0 ([installation guide](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli#install-terraform))
- [k3d](https://k3d.io/v5.4.6/) >= 5.4 ([installation guide](https://k3d.io/v5.4.6/#installation))

Create a `main.tf` file with the following content:

```hcl
terraform {
  required_providers {
    k3d = {
      source  = "sneakybugs/k3d"
      version = "1.0.1"
    }
  }
}

resource "k3d_cluster" "example_cluster" {
  name = "example"
  # See https://k3d.io/v5.4.6/usage/configfile/#config-options
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
  host                   = resource.k3d_cluster.example_cluster.host
  client_certificate     = base64decode(resource.k3d_cluster.example_cluster.client_certificate)
  client_key             = base64decode(resource.k3d_cluster.example_cluster.client_key)
  cluster_ca_certificate = base64decode(resource.k3d_cluster.example_cluster.cluster_ca_certificate)
}

resource "kubernetes_secret" "postgres_credentials" {
  metadata {
    name = "postgres-credentials"
  }

  data = {
    "postgres-password"    = "development"
    "password"             = "development"
    "replication-password" = "development"
  }
}

provider "helm" {
  kubernetes {
    host                   = resource.k3d_cluster.example_cluster.host
    client_certificate     = base64decode(resource.k3d_cluster.example_cluster.client_certificate)
    client_key             = base64decode(resource.k3d_cluster.example_cluster.client_key)
    cluster_ca_certificate = base64decode(resource.k3d_cluster.example_cluster.cluster_ca_certificate)
  }
}

resource "helm_release" "database" {
  name       = "postgres"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "postgresql"
  set {
    name  = "auth.existingSecret"
    value = "postgres-credentials"
  }
}
```

Run `terraform init` and `terraform apply`, you should have a k3d cluster with a Postgres instance.
Run `terraform destroy` to tear everything down when you are done.

*Note:* You may need to run Terraform with `sudo` because k3d uses Docker.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18
- [k3d](https://k3d.io/v5.4.6/) >= 5.4

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run:

```
TF_ACC=1 sudo -E /usr/local/go/bin/go test ./... -v -timeout 120m
```

*Note:* `sudo` is required because of Docker. Running tests only creates local resources.
