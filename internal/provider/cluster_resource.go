package provider

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/yaml.v3"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ClusterResource{}

func NewClusterResource() resource.Resource {
	return &ClusterResource{}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	client *http.Client
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	K3dConfig            types.String `tfsdk:"k3d_config"`
	Kubeconfig           types.String `tfsdk:"kubeconfig"`
	Host                 types.String `tfsdk:"host"`
	ClientCertificate    types.String `tfsdk:"client_certificate"`
	ClientKey            types.String `tfsdk:"client_key"`
	ClusterCACertificate types.String `tfsdk:"cluster_ca_certificate"`
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster resource",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Cluster resource unique identifier.",
				Type:                types.StringType,
				Computed:            true,
			},
			"name": {
				MarkdownDescription: "Cluster name.",
				Required:            true,
				Type:                types.StringType,
			},
			"k3d_config": {
				MarkdownDescription: "K3d config content.",
				Required:            true,
				Type:                types.StringType,
			},
			"kubeconfig": {
				MarkdownDescription: "Kubeconfig content. Dump in a file for use with kubectl and other tools.",
				Type:                types.StringType,
				Computed:            true,
				Sensitive:           true,
			},
			"host": {
				MarkdownDescription: "Cluster host. Use to authenticate other providers with the cluster.",
				Type:                types.StringType,
				Computed:            true,
			},
			"client_certificate": {
				MarkdownDescription: "Client certificate encoded in base 64.\n" +
					"Use to authenticate other providers with the cluster.",
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
			},
			"client_key": {
				MarkdownDescription: "Client key encoded in base 64.\n" +
					"Use to authenticate other providers with the cluster.",
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
			},
			"cluster_ca_certificate": {
				MarkdownDescription: "Cluster CA certificate encoded in base 64.\n" +
					"Use to authenticate other providers with the cluster.",
				Type:      types.StringType,
				Computed:  true,
				Sensitive: true,
			},
		},
	}, nil
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	checksum := md5.Sum([]byte(data.K3dConfig.String()))
	configPath := fmt.Sprintf(filepath.Join(os.TempDir(), "k3d-config-%x.yaml"), checksum)
	if err := os.WriteFile(configPath, []byte(data.K3dConfig.ValueString()), 0600); err != nil {
		resp.Diagnostics.AddError("Failed writing temporary k3d config", fmt.Sprint(err))
		return
	}

	cmd := exec.Command("k3d", "cluster", "create", data.Name.ValueString(), "--config", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputString := string(output)
		if strings.Contains(outputString, "already exists") {
			// TODO Handle name already exists.
		}
		if strings.Contains(outputString, "Schema Validation failed") {
			// TODO Handle schema validation failure.
		}
		if strings.Contains(outputString, "permission denied") {
			// TODO Handle running without permission for Docker.
		}
		if strings.Contains(outputString, "executable file not found") {
			// TODO Handle k3d is not installed.
		}
		resp.Diagnostics.AddError("Failed creating k3d cluster", outputString)
		return
	}
	configChecksum := fmt.Sprintf("%x", checksum)
	data.ID = types.StringValue(configChecksum)

	cmd = exec.Command("k3d", "kubeconfig", "get", data.Name.ValueString())
	output, err = cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError("Failed getting Kubeconfig from k3d", string(output))
		return
	}

	var kubeconfig Kubeconfig
	if err := yaml.Unmarshal(output, &kubeconfig); err != nil {
		resp.Diagnostics.AddError("Failed parsing Kubeconfig", fmt.Sprint(err))
		return
	}

	if len(kubeconfig.Clusters) != 1 || len(kubeconfig.Users) != 1 {
		resp.Diagnostics.AddError(
			"Kubeconfig parsed with more than 1 user or cluster.",
			"contact the provider's developer")
		return
	}
	data.Host = types.StringValue(kubeconfig.Clusters[0].Cluster.Server)
	data.ClusterCACertificate = types.StringValue(kubeconfig.Clusters[0].Cluster.CertificateAuthorityData)
	data.ClientCertificate = types.StringValue(kubeconfig.Users[0].User.ClientCertificateData)
	data.ClientKey = types.StringValue(kubeconfig.Users[0].User.ClientKeyData)
	data.Kubeconfig = types.StringValue(string(output))

	if err := os.Remove(configPath); err != nil {
		// TODO Continue as this is not a critical error?
		resp.Diagnostics.AddError("Failed removing temporary k3d config", fmt.Sprint(err))
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	cmd := exec.Command("k3d", "cluster", "list", "--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError("Failed listing k3d cluster", fmt.Sprint(err))
		return
	}
	var clusters []K3dClusterInfo
	if err := json.Unmarshal(output, &clusters); err != nil {
		fmt.Println(err)
		return
	}
	cluster, err := findCluster(clusters, data.Name.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if cluster.ServersRunning < cluster.ServersCount {
		// TODO handle needing to start the cluster?
	}

	cmd = exec.Command("k3d", "kubeconfig", "get", data.Name.ValueString())
	output, err = cmd.CombinedOutput()
	if err != nil {
		resp.Diagnostics.AddError("Failed getting Kubeconfig from k3d", string(output))
		return
	}

	var kubeconfig Kubeconfig
	if err := yaml.Unmarshal(output, &kubeconfig); err != nil {
		resp.Diagnostics.AddError("Failed parsing Kubeconfig", fmt.Sprint(err))
		return
	}

	if len(kubeconfig.Clusters) != 1 || len(kubeconfig.Users) != 1 {
		resp.Diagnostics.AddError(
			"Kubeconfig parsed with more than 1 user or cluster.",
			"contact the provider's developer")
		return
	}
	data.Host = types.StringValue(kubeconfig.Clusters[0].Cluster.Server)
	data.ClusterCACertificate = types.StringValue(kubeconfig.Clusters[0].Cluster.CertificateAuthorityData)
	data.ClientCertificate = types.StringValue(kubeconfig.Users[0].User.ClientCertificateData)
	data.ClientKey = types.StringValue(kubeconfig.Users[0].User.ClientKeyData)
	data.Kubeconfig = types.StringValue(string(output))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findCluster(clusters []K3dClusterInfo, name string) (K3dClusterInfo, error) {
	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return K3dClusterInfo{}, fmt.Errorf("clusters does not contain a cluster with matching name")
}

type K3dClusterInfo struct {
	Name           string `json:"name"`
	ServersCount   int    `json:"serversCount"`
	ServersRunning int    `json:"serversRunning"`
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }
	resp.Diagnostics.AddError(
		"Updating clusters is not supported by k3d",
		"Destroy the resource and apply again to recreate the cluster.")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ClusterResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
	cmd := exec.Command("k3d", "cluster", "delete", data.Name.ValueString())
	if err := cmd.Run(); err != nil {
		resp.Diagnostics.AddError("Failed deleting k3d cluster", fmt.Sprint(err))
		return
	}
}

type Kubeconfig struct {
	Users    []KubeconfigUser    `yaml:"users"`
	Clusters []KubeconfigCluster `yaml:"clusters"`
}

type KubeconfigUser struct {
	User KubeconfigUserData `yaml:"user"`
}

type KubeconfigUserData struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}

type KubeconfigCluster struct {
	Cluster KubeconfigClusterData `yaml:"cluster"`
}

type KubeconfigClusterData struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}
