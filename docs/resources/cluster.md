---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "k3d_cluster Resource - terraform-provider-k3d"
subcategory: ""
description: |-
  The resource k3d_cluster manages k3d clusters for development.
  This resource can be used in conjunction with the Kubernetes and Helm providers to define an entire Kubernetes development environment as code.
  Updating cluster configuration or name is not supported by k3d. When changing the name or k3d_config attributes destroy the resource and apply again.
---

# k3d_cluster (Resource)

The resource `k3d_cluster` manages k3d clusters for development.

This resource can be used in conjunction with the Kubernetes and Helm providers to define an entire Kubernetes development environment as code.

Updating cluster configuration or name is not supported by k3d. When changing the `name` or `k3d_config` attributes destroy the resource and apply again.

## Example Usage

```terraform
resource "k3d_cluster" "example" {
  name       = "example-cluster"
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `k3d_config` (String) K3d config content. Use to set the amounts of servers, agents, container registries, ports, host aliases and more cluster related options. [See config options in k3d documentation](https://k3d.io/v5.4.6/usage/configfile/#config-options).
- `name` (String) Cluster name.

### Read-Only

- `client_certificate` (String, Sensitive) Client certificate encoded in base 64. Use to authenticate other providers with the cluster. Use `base64decode` and pass to `client_certificate` attribute when [configuring Kubernetes](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/guides/getting-started#provider-setup) or [Helm providers](https://registry.terraform.io/providers/hashicorp/helm/latest/docs#credentials-config).
- `client_key` (String, Sensitive) Client key encoded in base 64. Use to authenticate other providers with the cluster. Use `base64decode` and pass to `client_key` attribute when [configuring Kubernetes](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/guides/getting-started#provider-setup) or [Helm providers](https://registry.terraform.io/providers/hashicorp/helm/latest/docs#credentials-config).
- `cluster_ca_certificate` (String, Sensitive) Cluster CA certificate encoded in base 64. Use to authenticate other providers with the cluster. Use `base64decode` and pass to `cluster_ca_certificate` attribute when [configuring Kubernetes](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/guides/getting-started#provider-setup) or [Helm providers](https://registry.terraform.io/providers/hashicorp/helm/latest/docs#credentials-config).
- `host` (String) Cluster host. Use to authenticate other providers with the cluster. Pass to `host` attribute when [configuring Kubernetes](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/guides/getting-started#provider-setup) or [Helm providers](https://registry.terraform.io/providers/hashicorp/helm/latest/docs#credentials-config).
- `id` (String) Used internally by the provider.
- `kubeconfig` (String, Sensitive) Kubeconfig content. Dump in a file and point the `KUBECONFIG` environment variable or `--kubeconfig` flag at it to use kubectl or Helm with the cluster.


