package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HostDataSource struct {
	client *client.Client
}

type hostDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	FQDN        types.String `tfsdk:"fqdn"`
	Description types.String `tfsdk:"description"`
}

func NewHostDataSource() datasource.DataSource {
	return &HostDataSource{}
}

func (d *HostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (d *HostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA host.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique FQDN of the host.",
			},
			"fqdn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Fully Qualified Domain Name (FQDN) of the host.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the host.",
			},
		},
	}
}

func (d *HostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state hostDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAHostShowResult
	err := d.client.Call(ctx, "host_show", []string{state.FQDN.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA host", err.Error())
		return
	}

	res := result.Result
	state.ID = state.FQDN
	state.FQDN = types.StringValue(parseStringVal(res.Fqdn))
	state.Description = types.StringValue(parseStringVal(res.Description))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
