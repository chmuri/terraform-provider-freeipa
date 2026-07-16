package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HostGroupDataSource struct {
	client *client.Client
}

type hostGroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Hosts       types.Set    `tfsdk:"hosts"`
}

func NewHostGroupDataSource() datasource.DataSource {
	return &HostGroupDataSource{}
}

func (d *HostGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hostgroup"
}

func (d *HostGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA hostgroup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the hostgroup.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the hostgroup (cn).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the hostgroup.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of host FQDNs belonging to this hostgroup.",
			},
		},
	}
}

func (d *HostGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type hostGroupDSResult struct {
	Result struct {
		Cn          interface{} `json:"cn"`
		Description interface{} `json:"description"`
		MemberHost  interface{} `json:"member_host"`
	} `json:"result"`
}

func (d *HostGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state hostGroupDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result hostGroupDSResult
	err := d.client.Call(ctx, "hostgroup_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA hostgroup", err.Error())
		return
	}

	res := result.Result
	state.ID = state.Name
	state.Name = types.StringValue(parseStringVal(res.Cn))
	state.Description = types.StringValue(parseStringVal(res.Description))

	hosts := parseStringSlice(res.MemberHost)
	if len(hosts) > 0 {
		hostsVal, d := types.SetValueFrom(ctx, types.StringType, hosts)
		resp.Diagnostics.Append(d...)
		state.Hosts = hostsVal
	} else {
		state.Hosts = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
