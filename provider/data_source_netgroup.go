package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NetgroupDataSource struct {
	client *client.Client
}

type netgroupDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	NisDomainName types.String `tfsdk:"nisdomainname"`
	Users         types.Set    `tfsdk:"users"`
	Groups        types.Set    `tfsdk:"groups"`
	Hosts         types.Set    `tfsdk:"hosts"`
	HostGroups    types.Set    `tfsdk:"hostgroups"`
	Netgroups     types.Set    `tfsdk:"netgroups"`
}

func NewNetgroupDataSource() datasource.DataSource {
	return &NetgroupDataSource{}
}

func (d *NetgroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_netgroup"
}

func (d *NetgroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Netgroup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the netgroup.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the netgroup (cn).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the netgroup.",
			},
			"nisdomainname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "NIS domain name.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of usernames associated with this netgroup.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of group names associated with this netgroup.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of host FQDNs associated with this netgroup.",
			},
			"hostgroups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of hostgroups associated with this netgroup.",
			},
			"netgroups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of netgroups associated with this netgroup.",
			},
		},
	}
}

func (d *NetgroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetgroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state netgroupDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPANetgroupResult
	err := d.client.Call(ctx, "netgroup_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA netgroup", err.Error())
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

	if res.NisDomainName != nil {
		state.NisDomainName = types.StringValue(parseStringVal(res.NisDomainName))
	} else {
		state.NisDomainName = types.StringNull()
	}

	users := parseStringSlice(res.MemberUser)
	if len(users) > 0 {
		usersVal, d := types.SetValueFrom(ctx, types.StringType, users)
		resp.Diagnostics.Append(d...)
		state.Users = usersVal
	} else {
		state.Users = types.SetNull(types.StringType)
	}

	groups := parseStringSlice(res.MemberGroup)
	if len(groups) > 0 {
		groupsVal, d := types.SetValueFrom(ctx, types.StringType, groups)
		resp.Diagnostics.Append(d...)
		state.Groups = groupsVal
	} else {
		state.Groups = types.SetNull(types.StringType)
	}

	hosts := parseStringSlice(res.MemberHost)
	if len(hosts) > 0 {
		hostsVal, d := types.SetValueFrom(ctx, types.StringType, hosts)
		resp.Diagnostics.Append(d...)
		state.Hosts = hostsVal
	} else {
		state.Hosts = types.SetNull(types.StringType)
	}

	hostgroups := parseStringSlice(res.MemberHostGroup)
	if len(hostgroups) > 0 {
		hgVal, d := types.SetValueFrom(ctx, types.StringType, hostgroups)
		resp.Diagnostics.Append(d...)
		state.HostGroups = hgVal
	} else {
		state.HostGroups = types.SetNull(types.StringType)
	}

	netgroups := parseStringSlice(res.MemberNetgroup)
	if len(netgroups) > 0 {
		ngVal, d := types.SetValueFrom(ctx, types.StringType, netgroups)
		resp.Diagnostics.Append(d...)
		state.Netgroups = ngVal
	} else {
		state.Netgroups = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
