package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HbacRuleDataSource struct {
	client *client.Client
}

type hbacRuleDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	HostCategory    types.String `tfsdk:"host_category"`
	UserCategory    types.String `tfsdk:"user_category"`
	ServiceCategory types.String `tfsdk:"service_category"`
	Users           types.Set    `tfsdk:"users"`
	Groups          types.Set    `tfsdk:"groups"`
	Hosts           types.Set    `tfsdk:"hosts"`
	Services        types.Set    `tfsdk:"services"`
	Enabled         types.Bool   `tfsdk:"enabled"`
}

func NewHbacRuleDataSource() datasource.DataSource {
	return &HbacRuleDataSource{}
}

func (d *HbacRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hbacrule"
}

func (d *HbacRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA HBAC rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the HBAC rule.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the HBAC rule (cn).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the HBAC rule.",
			},
			"host_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Host category.",
			},
			"user_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User category.",
			},
			"service_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service category.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Associated users.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Associated user groups.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Associated hosts.",
			},
			"services": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Associated HBAC services.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the HBAC rule is enabled.",
			},
		},
	}
}

func (d *HbacRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *HbacRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state hbacRuleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAHbacRuleResult
	err := d.client.Call(ctx, "hbacrule_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA HBAC rule", err.Error())
		return
	}

	res := result.Result
	state.ID = state.Name
	state.Name = types.StringValue(parseStringVal(res.Cn))
	state.Description = types.StringValue(parseStringVal(res.Description))
	state.HostCategory = types.StringValue(parseStringVal(res.Hostcategory))
	state.UserCategory = types.StringValue(parseStringVal(res.Usercategory))
	state.ServiceCategory = types.StringValue(parseStringVal(res.Servicecategory))
	state.Enabled = types.BoolValue(parseBoolVal(res.Ipaenabledflag))

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

	services := parseStringSlice(res.MemberService)
	if len(services) > 0 {
		servicesVal, d := types.SetValueFrom(ctx, types.StringType, services)
		resp.Diagnostics.Append(d...)
		state.Services = servicesVal
	} else {
		state.Services = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
