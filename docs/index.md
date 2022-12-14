---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "k3d Provider"
subcategory: ""
description: |-
  This provider manages development Kubernetes clusters in Docker with k3d. Managing k3d clusters in Terraform allows you to provision development clusters and deploy additional software (such as a database for your app) in a single action.
  The idea behind this provider is to automate everything before tilt up with Terraform.
  Quick Start
  Make sure to install k3d, see the installation guide https://k3d.io/v5.4.6/#installation.
  You may need to run Terraform with sudo because k3d uses Docker.
  The example below creates a cluster and deploys a Postgres instance on it. It can be adapted to deploy any services your app needs for development with minimal effort.
---

# k3d Provider

This provider manages development Kubernetes clusters in Docker with k3d. Managing k3d clusters in Terraform allows you to provision development clusters and deploy additional software (such as a database for your app) in a single action.

The idea behind this provider is to automate everything before `tilt up` with Terraform.

## Quick Start

Make sure to install k3d, [see the installation guide](https://k3d.io/v5.4.6/#installation).

You may need to run Terraform with `sudo` because k3d uses Docker.

The example below creates a cluster and deploys a Postgres instance on it. It can be adapted to deploy any services your app needs for development with minimal effort.

## Example Usage

```terraform
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

<!-- schema generated by tfplugindocs -->
## Schema
