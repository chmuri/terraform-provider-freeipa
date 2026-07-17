package provider

import (
	"context"
	"strconv"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SudoRuleDataSource struct {
	client *client.Client
}

type sudoRuleDataSourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	UserCategory       types.String `tfsdk:"user_category"`
	HostCategory       types.String `tfsdk:"host_category"`
	CommandCategory    types.String `tfsdk:"command_category"`
	RunAsUserCategory  types.String `tfsdk:"runas_user_category"`
	RunAsGroupCategory types.String `tfsdk:"runas_group_category"`
	Order              types.Int64  `tfsdk:"order"`
	Users              types.Set    `tfsdk:"users"`
	Groups             types.Set    `tfsdk:"groups"`
	Hosts              types.Set    `tfsdk:"hosts"`
	HostGroups         types.Set    `tfsdk:"host_groups"`
	AllowCommands      types.Set    `tfsdk:"allow_commands"`
	AllowCommandGroups types.Set    `tfsdk:"allow_command_groups"`
	DenyCommands       types.Set    `tfsdk:"deny_commands"`
	DenyCommandGroups  types.Set    `tfsdk:"deny_command_groups"`
	Options            types.Set    `tfsdk:"options"`
	RunAsUsers         types.Set    `tfsdk:"runas_users"`
	RunAsGroups        types.Set    `tfsdk:"runas_groups"`
}

func NewSudoRuleDataSource() datasource.DataSource {
	return &SudoRuleDataSource{}
}

func (d *SudoRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sudo_rule"
}

func (d *SudoRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Sudo Rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the Sudo rule.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Sudo rule.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the Sudo rule.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the Sudo rule is enabled.",
			},
			"user_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User category.",
			},
			"host_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Host category.",
			},
			"command_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Command category.",
			},
			"runas_user_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RunAs User category.",
			},
			"runas_group_category": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RunAs Group category.",
			},
			"order": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The evaluation order of the Sudo rule.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of usernames associated with the rule.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of user groups associated with the rule.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of host FQDNs associated with the rule.",
			},
			"host_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of host groups associated with the rule.",
			},
			"allow_commands": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of allowed Sudo commands associated with the rule.",
			},
			"allow_command_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of allowed Sudo command groups associated with the rule.",
			},
			"deny_commands": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of denied Sudo commands associated with the rule.",
			},
			"deny_command_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of denied Sudo command groups associated with the rule.",
			},
			"options": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of Sudo options associated with the rule.",
			},
			"runas_users": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of users the commands can be run as.",
			},
			"runas_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of groups the commands can be run as.",
			},
		},
	}
}

func (d *SudoRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SudoRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sudoRuleDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPASudoRuleResult
	err := d.client.Call(ctx, "sudorule_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA Sudo rule", err.Error())
		return
	}

	res := result.Result
	state.ID = state.Name
	state.Name = types.StringValue(parseStringVal(res.Cn))

	if len(res.Description) > 0 {
		state.Description = types.StringValue(res.Description[0])
	} else {
		state.Description = types.StringNull()
	}

	if len(res.Usercategory) > 0 {
		state.UserCategory = types.StringValue(res.Usercategory[0])
	} else {
		state.UserCategory = types.StringNull()
	}
	if len(res.Hostcategory) > 0 {
		state.HostCategory = types.StringValue(res.Hostcategory[0])
	} else {
		state.HostCategory = types.StringNull()
	}
	if len(res.Cmdcategory) > 0 {
		state.CommandCategory = types.StringValue(res.Cmdcategory[0])
	} else {
		state.CommandCategory = types.StringNull()
	}
	if len(res.Runasusercategory) > 0 {
		state.RunAsUserCategory = types.StringValue(res.Runasusercategory[0])
	} else {
		state.RunAsUserCategory = types.StringNull()
	}
	if len(res.Runasgroupcategory) > 0 {
		state.RunAsGroupCategory = types.StringValue(res.Runasgroupcategory[0])
	} else {
		state.RunAsGroupCategory = types.StringNull()
	}

	if len(res.Ipaenabledflag) > 0 {
		state.Enabled = types.BoolValue(res.Ipaenabledflag[0])
	} else {
		state.Enabled = types.BoolValue(false)
	}

	var orderVal int64
	var orderFound bool
	rawOrder := res.Sudoorder
	if len(rawOrder) == 0 {
		rawOrder = res.Ipasudoorder
	}
	if len(rawOrder) > 0 && rawOrder[0] != nil {
		switch v := rawOrder[0].(type) {
		case string:
			if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
				orderVal = parsed
				orderFound = true
			}
		case float64:
			orderVal = int64(v)
			orderFound = true
		case int64:
			orderVal = v
			orderFound = true
		case int:
			orderVal = int64(v)
			orderFound = true
		}
	}
	if orderFound {
		state.Order = types.Int64Value(orderVal)
	} else {
		state.Order = types.Int64Null()
	}

	if len(res.MemberUser) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberUser)
		resp.Diagnostics.Append(d...)
		state.Users = val
	} else {
		state.Users = types.SetNull(types.StringType)
	}

	if len(res.MemberGroup) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberGroup)
		resp.Diagnostics.Append(d...)
		state.Groups = val
	} else {
		state.Groups = types.SetNull(types.StringType)
	}

	if len(res.MemberHost) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberHost)
		resp.Diagnostics.Append(d...)
		state.Hosts = val
	} else {
		state.Hosts = types.SetNull(types.StringType)
	}

	if len(res.MemberHostGroup) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberHostGroup)
		resp.Diagnostics.Append(d...)
		state.HostGroups = val
	} else {
		state.HostGroups = types.SetNull(types.StringType)
	}

	if len(res.MemberAllowCmd) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberAllowCmd)
		resp.Diagnostics.Append(d...)
		state.AllowCommands = val
	} else {
		state.AllowCommands = types.SetNull(types.StringType)
	}

	if len(res.MemberAllowCmdGrp) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberAllowCmdGrp)
		resp.Diagnostics.Append(d...)
		state.AllowCommandGroups = val
	} else {
		state.AllowCommandGroups = types.SetNull(types.StringType)
	}

	if len(res.MemberDenyCmd) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberDenyCmd)
		resp.Diagnostics.Append(d...)
		state.DenyCommands = val
	} else {
		state.DenyCommands = types.SetNull(types.StringType)
	}

	if len(res.MemberDenyCmdGrp) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberDenyCmdGrp)
		resp.Diagnostics.Append(d...)
		state.DenyCommandGroups = val
	} else {
		state.DenyCommandGroups = types.SetNull(types.StringType)
	}

	if len(res.Ipasudoopt) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.Ipasudoopt)
		resp.Diagnostics.Append(d...)
		state.Options = val
	} else {
		state.Options = types.SetNull(types.StringType)
	}

	if len(res.MemberRunAsUser) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberRunAsUser)
		resp.Diagnostics.Append(d...)
		state.RunAsUsers = val
	} else {
		state.RunAsUsers = types.SetNull(types.StringType)
	}

	if len(res.MemberRunAsGrp) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, res.MemberRunAsGrp)
		resp.Diagnostics.Append(d...)
		state.RunAsGroups = val
	} else {
		state.RunAsGroups = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
