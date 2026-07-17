package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleDataSource struct {
	client *client.Client
}

type roleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Privileges  types.Set    `tfsdk:"privileges"`
	Users       types.Set    `tfsdk:"users"`
	Groups      types.Set    `tfsdk:"groups"`
	Hosts       types.Set    `tfsdk:"hosts"`
	HostGroups  types.Set    `tfsdk:"host_groups"`
}

func NewRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

func (d *RoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *RoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the role.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the role (cn).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the role.",
			},
			"privileges": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of privileges associated with this role.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of usernames associated with this role.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of group names associated with this role.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of host FQDNs associated with this role.",
			},
			"host_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of hostgroups associated with this role.",
			},
		},
	}
}

func (d *RoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state roleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPARoleResult
	err := d.client.Call(ctx, "role_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA role", err.Error())
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

	privs := parseStringSlice(res.MemberPrivilege)
	if len(privs) > 0 {
		privsVal, d := types.SetValueFrom(ctx, types.StringType, privs)
		resp.Diagnostics.Append(d...)
		state.Privileges = privsVal
	} else {
		state.Privileges = types.SetNull(types.StringType)
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
