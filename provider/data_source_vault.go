package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VaultDataSource struct {
	client *client.Client
}

type vaultDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
}

func NewVaultDataSource() datasource.DataSource {
	return &VaultDataSource{}
}

func (d *VaultDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault"
}

func (d *VaultDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Vault.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the vault.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the vault (cn).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the vault.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Type of the vault: standard, symmetric, asymmetric.",
			},
		},
	}
}

func (d *VaultDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *client.Client")
		return
	}

	d.client = c
}

func (d *VaultDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vaultDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAVaultResult
	err := d.client.Call(ctx, "vault_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA vault", err.Error())
		return
	}

	res := result.Result
	state.ID = state.Name
	state.Name = types.StringValue(parseStringVal(res.Cn))

	if res.Description != nil {
		state.Description = types.StringValue(parseStringVal(res.Description))
	} else {
		state.Description = types.StringNull()
	}

	if res.Type != nil {
		state.Type = types.StringValue(parseStringVal(res.Type))
	} else {
		state.Type = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
