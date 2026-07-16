package provider

import (
	"context"
	"strconv"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SudoRuleResource struct {
	client *client.Client
}

type sudoRuleResourceModel struct {
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

func NewSudoRuleResource() resource.Resource {
	return &SudoRuleResource{}
}

func (r *SudoRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sudo_rule"
}

func (r *SudoRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA Sudo Rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the Sudo rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the Sudo rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the Sudo rule.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the Sudo rule is enabled.",
			},
			"user_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User category (e.g. 'all' to apply to all users, or null).",
			},
			"host_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Host category (e.g. 'all' to apply to all hosts, or null).",
			},
			"command_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Command category (e.g. 'all' to apply to all commands, or null).",
			},
			"runas_user_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "RunAs User category (e.g. 'all' to apply to all RunAs users, or null).",
			},
			"runas_group_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "RunAs Group category (e.g. 'all' to apply to all RunAs groups, or null).",
			},
			"order": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "The evaluation order of the Sudo rule.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of usernames associated with the rule.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of user groups associated with the rule.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host FQDNs associated with the rule.",
			},
			"host_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host groups associated with the rule.",
			},
			"allow_commands": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of allowed Sudo commands associated with the rule.",
			},
			"allow_command_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of allowed Sudo command groups associated with the rule.",
			},
			"deny_commands": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of denied Sudo commands associated with the rule.",
			},
			"deny_command_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of denied Sudo command groups associated with the rule.",
			},
			"options": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of Sudo options associated with the rule (e.g. '!authenticate').",
			},
			"runas_users": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of users the commands can be run as.",
			},
			"runas_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of groups the commands can be run as.",
			},
		},
	}
}

func (r *SudoRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", "Expected *client.Client")
		return
	}

	r.client = c
}

type FreeIPASudoRuleResult struct {
	Result struct {
		Cn                 []string      `json:"cn"`
		Description        []string      `json:"description"`
		Usercat            []string      `json:"usercat"`
		Hostcat            []string      `json:"hostcat"`
		Cmdcat             []string      `json:"cmdcat"`
		Runasusercat       []string      `json:"runasusercat"`
		Runasgroupcat      []string      `json:"runasgroupcat"`
		Sudoorder          []interface{} `json:"sudoorder"`
		Ipasudoorder       []interface{} `json:"ipasudoorder"`
		Ipaenabledflag     []bool        `json:"ipaenabledflag"`
		Ipasudoopt         []string      `json:"ipasudoopt"`
		MemberUser         []string      `json:"memberuser_user"`
		MemberGroup        []string      `json:"memberuser_group"`
		MemberHost         []string      `json:"memberhost_host"`
		MemberHostGroup    []string      `json:"memberhost_hostgroup"`
		MemberAllowCmd     []string      `json:"memberallowcmd_sudocmd"`
		MemberAllowCmdGrp  []string      `json:"memberallowcmd_sudocmdgroup"`
		MemberDenyCmd      []string      `json:"memberdenycmd_sudocmd"`
		MemberDenyCmdGrp   []string      `json:"memberdenycmd_sudocmdgroup"`
		MemberRunAsUser    []string      `json:"memberrunasuser_user"`
		MemberRunAsGrp     []string      `json:"memberrunasgroup_group"`
	} `json:"result"`
}

func (r *SudoRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sudoRuleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}
	if !plan.UserCategory.IsNull() && !plan.UserCategory.IsUnknown() {
		opts["usercategory"] = plan.UserCategory.ValueString()
	}
	if !plan.HostCategory.IsNull() && !plan.HostCategory.IsUnknown() {
		opts["hostcategory"] = plan.HostCategory.ValueString()
	}
	if !plan.CommandCategory.IsNull() && !plan.CommandCategory.IsUnknown() {
		opts["cmdcategory"] = plan.CommandCategory.ValueString()
	}
	if !plan.RunAsUserCategory.IsNull() && !plan.RunAsUserCategory.IsUnknown() {
		opts["runasusercategory"] = plan.RunAsUserCategory.ValueString()
	}
	if !plan.RunAsGroupCategory.IsNull() && !plan.RunAsGroupCategory.IsUnknown() {
		opts["runasgroupcategory"] = plan.RunAsGroupCategory.ValueString()
	}
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		opts["sudoorder"] = plan.Order.ValueInt64()
	}

	err := r.client.Call(ctx, "sudorule_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA Sudo rule", err.Error())
		return
	}

	// Toggle enable flag
	if !plan.Enabled.ValueBool() {
		err = r.client.Call(ctx, "sudorule_disable", []string{plan.Name.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to disable Sudo rule", err.Error())
			return
		}
	}

	// Add associations: Users/Groups
	var users, groups []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &users, false)
	}
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		plan.Groups.ElementsAs(ctx, &groups, false)
	}
	if len(users) > 0 || len(groups) > 0 {
		assocOpts := map[string]interface{}{}
		if len(users) > 0 {
			assocOpts["user"] = users
		}
		if len(groups) > 0 {
			assocOpts["group"] = groups
		}
		err = r.client.Call(ctx, "sudorule_add_user", []string{plan.Name.ValueString()}, assocOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate users/groups with Sudo rule", err.Error())
			return
		}
	}

	// Add associations: Hosts/HostGroups
	var hosts, hostGroups []string
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		plan.Hosts.ElementsAs(ctx, &hosts, false)
	}
	if !plan.HostGroups.IsNull() && !plan.HostGroups.IsUnknown() {
		plan.HostGroups.ElementsAs(ctx, &hostGroups, false)
	}
	if len(hosts) > 0 || len(hostGroups) > 0 {
		assocOpts := map[string]interface{}{}
		if len(hosts) > 0 {
			assocOpts["host"] = hosts
		}
		if len(hostGroups) > 0 {
			assocOpts["hostgroup"] = hostGroups
		}
		err = r.client.Call(ctx, "sudorule_add_host", []string{plan.Name.ValueString()}, assocOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate hosts/host groups with Sudo rule", err.Error())
			return
		}
	}

	// Add associations: Allow Commands / Command Groups
	var allowCmds, allowCmdGrps []string
	if !plan.AllowCommands.IsNull() && !plan.AllowCommands.IsUnknown() {
		plan.AllowCommands.ElementsAs(ctx, &allowCmds, false)
	}
	if !plan.AllowCommandGroups.IsNull() && !plan.AllowCommandGroups.IsUnknown() {
		plan.AllowCommandGroups.ElementsAs(ctx, &allowCmdGrps, false)
	}
	if len(allowCmds) > 0 || len(allowCmdGrps) > 0 {
		assocOpts := map[string]interface{}{}
		if len(allowCmds) > 0 {
			assocOpts["sudocmd"] = allowCmds
		}
		if len(allowCmdGrps) > 0 {
			assocOpts["sudocmdgroup"] = allowCmdGrps
		}
		err = r.client.Call(ctx, "sudorule_add_allow_command", []string{plan.Name.ValueString()}, assocOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate allowed commands/groups with Sudo rule", err.Error())
			return
		}
	}

	// Add associations: Deny Commands / Command Groups
	var denyCmds, denyCmdGrps []string
	if !plan.DenyCommands.IsNull() && !plan.DenyCommands.IsUnknown() {
		plan.DenyCommands.ElementsAs(ctx, &denyCmds, false)
	}
	if !plan.DenyCommandGroups.IsNull() && !plan.DenyCommandGroups.IsUnknown() {
		plan.DenyCommandGroups.ElementsAs(ctx, &denyCmdGrps, false)
	}
	if len(denyCmds) > 0 || len(denyCmdGrps) > 0 {
		assocOpts := map[string]interface{}{}
		if len(denyCmds) > 0 {
			assocOpts["sudocmd"] = denyCmds
		}
		if len(denyCmdGrps) > 0 {
			assocOpts["sudocmdgroup"] = denyCmdGrps
		}
		err = r.client.Call(ctx, "sudorule_add_deny_command", []string{plan.Name.ValueString()}, assocOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to associate denied commands/groups with Sudo rule", err.Error())
			return
		}
	}

	// Add options
	var options []string
	if !plan.Options.IsNull() && !plan.Options.IsUnknown() {
		plan.Options.ElementsAs(ctx, &options, false)
	}
	for _, opt := range options {
		err = r.client.Call(ctx, "sudorule_add_option", []string{plan.Name.ValueString()}, map[string]interface{}{
			"ipasudoopt": opt,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add option to Sudo rule", err.Error())
			return
		}
	}

	// Add RunAsUsers
	var runasUsers []string
	if !plan.RunAsUsers.IsNull() && !plan.RunAsUsers.IsUnknown() {
		plan.RunAsUsers.ElementsAs(ctx, &runasUsers, false)
	}
	if len(runasUsers) > 0 {
		err = r.client.Call(ctx, "sudorule_add_runasuser", []string{plan.Name.ValueString()}, map[string]interface{}{
			"user": runasUsers,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add RunAs Users to Sudo rule", err.Error())
			return
		}
	}

	// Add RunAsGroups
	var runasGroups []string
	if !plan.RunAsGroups.IsNull() && !plan.RunAsGroups.IsUnknown() {
		plan.RunAsGroups.ElementsAs(ctx, &runasGroups, false)
	}
	if len(runasGroups) > 0 {
		err = r.client.Call(ctx, "sudorule_add_runasgroup", []string{plan.Name.ValueString()}, map[string]interface{}{
			"group": runasGroups,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add RunAs Groups to Sudo rule", err.Error())
			return
		}
	}

	plan.ID = plan.Name

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SudoRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sudoRuleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPASudoRuleResult
	err := r.client.Call(ctx, "sudorule_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA Sudo rule", err.Error())
		return
	}

	res := result.Result
	if len(res.Cn) > 0 {
		state.Name = types.StringValue(res.Cn[0])
		state.ID = types.StringValue(res.Cn[0])
	}
	if len(res.Description) > 0 {
		state.Description = types.StringValue(res.Description[0])
	} else {
		state.Description = types.StringNull()
	}
	if len(res.Usercat) > 0 {
		state.UserCategory = types.StringValue(res.Usercat[0])
	} else {
		state.UserCategory = types.StringNull()
	}
	if len(res.Hostcat) > 0 {
		state.HostCategory = types.StringValue(res.Hostcat[0])
	} else {
		state.HostCategory = types.StringNull()
	}
	if len(res.Cmdcat) > 0 {
		state.CommandCategory = types.StringValue(res.Cmdcat[0])
	} else {
		state.CommandCategory = types.StringNull()
	}
	if len(res.Runasusercat) > 0 {
		state.RunAsUserCategory = types.StringValue(res.Runasusercat[0])
	} else {
		state.RunAsUserCategory = types.StringNull()
	}
	if len(res.Runasgroupcat) > 0 {
		state.RunAsGroupCategory = types.StringValue(res.Runasgroupcat[0])
	} else {
		state.RunAsGroupCategory = types.StringNull()
	}

	if len(res.Ipaenabledflag) > 0 {
		state.Enabled = types.BoolValue(res.Ipaenabledflag[0])
	} else {
		state.Enabled = types.BoolValue(false)
	}

	// Parse order
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

	// Parse lists
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

func (r *SudoRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state sudoRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Modify options
	modOpts := map[string]interface{}{}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			modOpts["description"] = ""
		} else {
			modOpts["description"] = plan.Description.ValueString()
		}
	}
	if !plan.UserCategory.Equal(state.UserCategory) {
		if plan.UserCategory.IsNull() {
			modOpts["usercategory"] = ""
		} else {
			modOpts["usercategory"] = plan.UserCategory.ValueString()
		}
	}
	if !plan.HostCategory.Equal(state.HostCategory) {
		if plan.HostCategory.IsNull() {
			modOpts["hostcategory"] = ""
		} else {
			modOpts["hostcategory"] = plan.HostCategory.ValueString()
		}
	}
	if !plan.CommandCategory.Equal(state.CommandCategory) {
		if plan.CommandCategory.IsNull() {
			modOpts["cmdcategory"] = ""
		} else {
			modOpts["cmdcategory"] = plan.CommandCategory.ValueString()
		}
	}
	if !plan.RunAsUserCategory.Equal(state.RunAsUserCategory) {
		if plan.RunAsUserCategory.IsNull() {
			modOpts["runasusercategory"] = ""
		} else {
			modOpts["runasusercategory"] = plan.RunAsUserCategory.ValueString()
		}
	}
	if !plan.RunAsGroupCategory.Equal(state.RunAsGroupCategory) {
		if plan.RunAsGroupCategory.IsNull() {
			modOpts["runasgroupcategory"] = ""
		} else {
			modOpts["runasgroupcategory"] = plan.RunAsGroupCategory.ValueString()
		}
	}
	if !plan.Order.Equal(state.Order) {
		if plan.Order.IsNull() {
			modOpts["sudoorder"] = ""
		} else {
			modOpts["sudoorder"] = plan.Order.ValueInt64()
		}
	}

	if len(modOpts) > 0 {
		err := r.client.Call(ctx, "sudorule_mod", []string{plan.ID.ValueString()}, modOpts, nil)
		if err != nil && !isEmptyModlistError(err) {
			resp.Diagnostics.AddError("Failed to update Sudo rule attributes", err.Error())
			return
		}
	}

	// Toggle enabled
	if !plan.Enabled.Equal(state.Enabled) {
		method := "sudorule_enable"
		if !plan.Enabled.ValueBool() {
			method = "sudorule_disable"
		}
		err := r.client.Call(ctx, method, []string{plan.ID.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to change Sudo rule enabled flag", err.Error())
			return
		}
	}

	// Users/Groups Membership
	var planUsers, stateUsers []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &planUsers, false)
	}
	if !state.Users.IsNull() && !state.Users.IsUnknown() {
		state.Users.ElementsAs(ctx, &stateUsers, false)
	}
	usersAdded := difference(planUsers, stateUsers)
	usersRemoved := difference(stateUsers, planUsers)

	var planGroups, stateGroups []string
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		plan.Groups.ElementsAs(ctx, &planGroups, false)
	}
	if !state.Groups.IsNull() && !state.Groups.IsUnknown() {
		state.Groups.ElementsAs(ctx, &stateGroups, false)
	}
	groupsAdded := difference(planGroups, stateGroups)
	groupsRemoved := difference(stateGroups, planGroups)

	if len(usersAdded) > 0 || len(groupsAdded) > 0 {
		addOpts := map[string]interface{}{}
		if len(usersAdded) > 0 {
			addOpts["user"] = usersAdded
		}
		if len(groupsAdded) > 0 {
			addOpts["group"] = groupsAdded
		}
		err := r.client.Call(ctx, "sudorule_add_user", []string{plan.ID.ValueString()}, addOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add users/groups to Sudo rule", err.Error())
			return
		}
	}
	if len(usersRemoved) > 0 || len(groupsRemoved) > 0 {
		remOpts := map[string]interface{}{}
		if len(usersRemoved) > 0 {
			remOpts["user"] = usersRemoved
		}
		if len(groupsRemoved) > 0 {
			remOpts["group"] = groupsRemoved
		}
		err := r.client.Call(ctx, "sudorule_remove_user", []string{plan.ID.ValueString()}, remOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove users/groups from Sudo rule", err.Error())
			return
		}
	}

	// Hosts/HostGroups Membership
	var planHosts, stateHosts []string
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		plan.Hosts.ElementsAs(ctx, &planHosts, false)
	}
	if !state.Hosts.IsNull() && !state.Hosts.IsUnknown() {
		state.Hosts.ElementsAs(ctx, &stateHosts, false)
	}
	hostsAdded := difference(planHosts, stateHosts)
	hostsRemoved := difference(stateHosts, planHosts)

	var planHostGroups, stateHostGroups []string
	if !plan.HostGroups.IsNull() && !plan.HostGroups.IsUnknown() {
		plan.HostGroups.ElementsAs(ctx, &planHostGroups, false)
	}
	if !state.HostGroups.IsNull() && !state.HostGroups.IsUnknown() {
		state.HostGroups.ElementsAs(ctx, &stateHostGroups, false)
	}
	hgAdded := difference(planHostGroups, stateHostGroups)
	hgRemoved := difference(stateHostGroups, planHostGroups)

	if len(hostsAdded) > 0 || len(hgAdded) > 0 {
		addOpts := map[string]interface{}{}
		if len(hostsAdded) > 0 {
			addOpts["host"] = hostsAdded
		}
		if len(hgAdded) > 0 {
			addOpts["hostgroup"] = hgAdded
		}
		err := r.client.Call(ctx, "sudorule_add_host", []string{plan.ID.ValueString()}, addOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add hosts/host groups to Sudo rule", err.Error())
			return
		}
	}
	if len(hostsRemoved) > 0 || len(hgRemoved) > 0 {
		remOpts := map[string]interface{}{}
		if len(hostsRemoved) > 0 {
			remOpts["host"] = hostsRemoved
		}
		if len(hgRemoved) > 0 {
			remOpts["hostgroup"] = hgRemoved
		}
		err := r.client.Call(ctx, "sudorule_remove_host", []string{plan.ID.ValueString()}, remOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove hosts/host groups from Sudo rule", err.Error())
			return
		}
	}

	// Allow Commands / Command Groups Membership
	var planAllowCmds, stateAllowCmds []string
	if !plan.AllowCommands.IsNull() && !plan.AllowCommands.IsUnknown() {
		plan.AllowCommands.ElementsAs(ctx, &planAllowCmds, false)
	}
	if !state.AllowCommands.IsNull() && !state.AllowCommands.IsUnknown() {
		state.AllowCommands.ElementsAs(ctx, &stateAllowCmds, false)
	}
	allowCmdsAdded := difference(planAllowCmds, stateAllowCmds)
	allowCmdsRemoved := difference(stateAllowCmds, planAllowCmds)

	var planAllowCmdGrps, stateAllowCmdGrps []string
	if !plan.AllowCommandGroups.IsNull() && !plan.AllowCommandGroups.IsUnknown() {
		plan.AllowCommandGroups.ElementsAs(ctx, &planAllowCmdGrps, false)
	}
	if !state.AllowCommandGroups.IsNull() && !state.AllowCommandGroups.IsUnknown() {
		state.AllowCommandGroups.ElementsAs(ctx, &stateAllowCmdGrps, false)
	}
	allowCmdGrpsAdded := difference(planAllowCmdGrps, stateAllowCmdGrps)
	allowCmdGrpsRemoved := difference(stateAllowCmdGrps, planAllowCmdGrps)

	if len(allowCmdsAdded) > 0 || len(allowCmdGrpsAdded) > 0 {
		addOpts := map[string]interface{}{}
		if len(allowCmdsAdded) > 0 {
			addOpts["sudocmd"] = allowCmdsAdded
		}
		if len(allowCmdGrpsAdded) > 0 {
			addOpts["sudocmdgroup"] = allowCmdGrpsAdded
		}
		err := r.client.Call(ctx, "sudorule_add_allow_command", []string{plan.ID.ValueString()}, addOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add allowed commands/groups to Sudo rule", err.Error())
			return
		}
	}
	if len(allowCmdsRemoved) > 0 || len(allowCmdGrpsRemoved) > 0 {
		remOpts := map[string]interface{}{}
		if len(allowCmdsRemoved) > 0 {
			remOpts["sudocmd"] = allowCmdsRemoved
		}
		if len(allowCmdGrpsRemoved) > 0 {
			remOpts["sudocmdgroup"] = allowCmdGrpsRemoved
		}
		err := r.client.Call(ctx, "sudorule_remove_allow_command", []string{plan.ID.ValueString()}, remOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove allowed commands/groups from Sudo rule", err.Error())
			return
		}
	}

	// Deny Commands / Command Groups Membership
	var planDenyCmds, stateDenyCmds []string
	if !plan.DenyCommands.IsNull() && !plan.DenyCommands.IsUnknown() {
		plan.DenyCommands.ElementsAs(ctx, &planDenyCmds, false)
	}
	if !state.DenyCommands.IsNull() && !state.DenyCommands.IsUnknown() {
		state.DenyCommands.ElementsAs(ctx, &stateDenyCmds, false)
	}
	denyCmdsAdded := difference(planDenyCmds, stateDenyCmds)
	denyCmdsRemoved := difference(stateDenyCmds, planDenyCmds)

	var planDenyCmdGrps, stateDenyCmdGrps []string
	if !plan.DenyCommandGroups.IsNull() && !plan.DenyCommandGroups.IsUnknown() {
		plan.DenyCommandGroups.ElementsAs(ctx, &planDenyCmdGrps, false)
	}
	if !state.DenyCommandGroups.IsNull() && !state.DenyCommandGroups.IsUnknown() {
		state.DenyCommandGroups.ElementsAs(ctx, &stateDenyCmdGrps, false)
	}
	denyCmdGrpsAdded := difference(planDenyCmdGrps, stateDenyCmdGrps)
	denyCmdGrpsRemoved := difference(stateDenyCmdGrps, planDenyCmdGrps)

	if len(denyCmdsAdded) > 0 || len(denyCmdGrpsAdded) > 0 {
		addOpts := map[string]interface{}{}
		if len(denyCmdsAdded) > 0 {
			addOpts["sudocmd"] = denyCmdsAdded
		}
		if len(denyCmdGrpsAdded) > 0 {
			addOpts["sudocmdgroup"] = denyCmdGrpsAdded
		}
		err := r.client.Call(ctx, "sudorule_add_deny_command", []string{plan.ID.ValueString()}, addOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add denied commands/groups to Sudo rule", err.Error())
			return
		}
	}
	if len(denyCmdsRemoved) > 0 || len(denyCmdGrpsRemoved) > 0 {
		remOpts := map[string]interface{}{}
		if len(denyCmdsRemoved) > 0 {
			remOpts["sudocmd"] = denyCmdsRemoved
		}
		if len(denyCmdGrpsRemoved) > 0 {
			remOpts["sudocmdgroup"] = denyCmdGrpsRemoved
		}
		err := r.client.Call(ctx, "sudorule_remove_deny_command", []string{plan.ID.ValueString()}, remOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove denied commands/groups from Sudo rule", err.Error())
			return
		}
	}

	// Options
	var planOptions, stateOptions []string
	if !plan.Options.IsNull() && !plan.Options.IsUnknown() {
		plan.Options.ElementsAs(ctx, &planOptions, false)
	}
	if !state.Options.IsNull() && !state.Options.IsUnknown() {
		state.Options.ElementsAs(ctx, &stateOptions, false)
	}
	optsAdded := difference(planOptions, stateOptions)
	optsRemoved := difference(stateOptions, planOptions)

	for _, add := range optsAdded {
		err := r.client.Call(ctx, "sudorule_add_option", []string{plan.ID.ValueString()}, map[string]interface{}{
			"ipasudoopt": add,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add option to Sudo rule", err.Error())
			return
		}
	}
	for _, rem := range optsRemoved {
		err := r.client.Call(ctx, "sudorule_remove_option", []string{plan.ID.ValueString()}, map[string]interface{}{
			"ipasudoopt": rem,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove option from Sudo rule", err.Error())
			return
		}
	}

	// RunAs Users
	var planRunasUsers, stateRunasUsers []string
	if !plan.RunAsUsers.IsNull() && !plan.RunAsUsers.IsUnknown() {
		plan.RunAsUsers.ElementsAs(ctx, &planRunasUsers, false)
	}
	if !state.RunAsUsers.IsNull() && !state.RunAsUsers.IsUnknown() {
		state.RunAsUsers.ElementsAs(ctx, &stateRunasUsers, false)
	}
	runasUsersAdded := difference(planRunasUsers, stateRunasUsers)
	runasUsersRemoved := difference(stateRunasUsers, planRunasUsers)

	if len(runasUsersAdded) > 0 {
		err := r.client.Call(ctx, "sudorule_add_runasuser", []string{plan.ID.ValueString()}, map[string]interface{}{
			"user": runasUsersAdded,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add RunAs Users to Sudo rule", err.Error())
			return
		}
	}
	if len(runasUsersRemoved) > 0 {
		err := r.client.Call(ctx, "sudorule_remove_runasuser", []string{plan.ID.ValueString()}, map[string]interface{}{
			"user": runasUsersRemoved,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove RunAs Users from Sudo rule", err.Error())
			return
		}
	}

	// RunAs Groups
	var planRunasGroups, stateRunasGroups []string
	if !plan.RunAsGroups.IsNull() && !plan.RunAsGroups.IsUnknown() {
		plan.RunAsGroups.ElementsAs(ctx, &planRunasGroups, false)
	}
	if !state.RunAsGroups.IsNull() && !state.RunAsGroups.IsUnknown() {
		state.RunAsGroups.ElementsAs(ctx, &stateRunasGroups, false)
	}
	runasGroupsAdded := difference(planRunasGroups, stateRunasGroups)
	runasGroupsRemoved := difference(stateRunasGroups, planRunasGroups)

	if len(runasGroupsAdded) > 0 {
		err := r.client.Call(ctx, "sudorule_add_runasgroup", []string{plan.ID.ValueString()}, map[string]interface{}{
			"group": runasGroupsAdded,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add RunAs Groups to Sudo rule", err.Error())
			return
		}
	}
	if len(runasGroupsRemoved) > 0 {
		err := r.client.Call(ctx, "sudorule_remove_runasgroup", []string{plan.ID.ValueString()}, map[string]interface{}{
			"group": runasGroupsRemoved,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove RunAs Groups from Sudo rule", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SudoRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sudoRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "sudorule_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA Sudo rule", err.Error())
		return
	}
}

func (r *SudoRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
