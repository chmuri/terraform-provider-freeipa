package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleResource struct {
	client *client.Client
}

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Privileges  types.Set    `tfsdk:"privileges"`
	Users       types.Set    `tfsdk:"users"`
	Groups      types.Set    `tfsdk:"groups"`
	Hosts       types.Set    `tfsdk:"hosts"`
	HostGroups  types.Set    `tfsdk:"hostgroups"`
}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

func (r *RoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA RBAC Roles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the role.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the role (cn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the role.",
			},
			"privileges": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of privileges associated with this role.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of usernames associated with this role.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of group names associated with this role.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host FQDNs associated with this role.",
			},
			"hostgroups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of hostgroups associated with this role.",
			},
		},
	}
}

func (r *RoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPARoleResult struct {
	Result struct {
		Cn              interface{} `json:"cn"`
		Description     interface{} `json:"description"`
		MemberPrivilege interface{} `json:"memberof_privilege"`
		MemberUser      interface{} `json:"member_user"`
		MemberGroup     interface{} `json:"member_group"`
		MemberHost      interface{} `json:"member_host"`
		MemberHostGroup interface{} `json:"member_hostgroup"`
	} `json:"result"`
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}

	err := r.client.Call(ctx, "role_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA role", err.Error())
		return
	}

	// Add privileges
	if !plan.Privileges.IsNull() && !plan.Privileges.IsUnknown() {
		var privileges []string
		plan.Privileges.ElementsAs(ctx, &privileges, false)
		for _, priv := range privileges {
			err = r.client.Call(ctx, "role_add_privilege", []string{plan.Name.ValueString()}, map[string]interface{}{
				"privilege": priv,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add privilege to role", err.Error())
				return
			}
		}
	}

	// Add members
	memberOpts := map[string]interface{}{}
	var users, groups, hosts, hostgroups []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &users, false)
		if len(users) > 0 {
			memberOpts["user"] = users
		}
	}
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		plan.Groups.ElementsAs(ctx, &groups, false)
		if len(groups) > 0 {
			memberOpts["group"] = groups
		}
	}
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		plan.Hosts.ElementsAs(ctx, &hosts, false)
		if len(hosts) > 0 {
			memberOpts["host"] = hosts
		}
	}
	if !plan.HostGroups.IsNull() && !plan.HostGroups.IsUnknown() {
		plan.HostGroups.ElementsAs(ctx, &hostgroups, false)
		if len(hostgroups) > 0 {
			memberOpts["hostgroup"] = hostgroups
		}
	}

	if len(memberOpts) > 0 {
		err = r.client.Call(ctx, "role_add_member", []string{plan.Name.ValueString()}, memberOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add members to role", err.Error())
			return
		}
	}

	plan.ID = plan.Name

	var result FreeIPARoleResult
	err = r.client.Call(ctx, "role_show", []string{plan.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA role after create", err.Error())
		return
	}

	res := result.Result
	plan.Name = types.StringValue(parseStringVal(res.Cn))
	plan.ID = plan.Name
	if res.Description != nil {
		plan.Description = types.StringValue(parseStringVal(res.Description))
	} else {
		plan.Description = types.StringNull()
	}
	privs := parseStringSlice(res.MemberPrivilege)
	if len(privs) > 0 {
		privsVal, d := types.SetValueFrom(ctx, types.StringType, privs)
		resp.Diagnostics.Append(d...)
		plan.Privileges = privsVal
	} else {
		plan.Privileges = types.SetNull(types.StringType)
	}
	users = parseStringSlice(res.MemberUser)
	if len(users) > 0 {
		usersVal, d := types.SetValueFrom(ctx, types.StringType, users)
		resp.Diagnostics.Append(d...)
		plan.Users = usersVal
	} else {
		plan.Users = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPARoleResult
	err := r.client.Call(ctx, "role_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA role", err.Error())
		return
	}

	res := result.Result
	state.Name = types.StringValue(parseStringVal(res.Cn))
	state.ID = state.Name

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

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state roleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update description if changed
	if !plan.Description.Equal(state.Description) {
		opts := map[string]interface{}{}
		if plan.Description.IsNull() {
			opts["description"] = ""
		} else {
			opts["description"] = plan.Description.ValueString()
		}
		err := r.client.Call(ctx, "role_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil && !isEmptyModlistError(err) {
			resp.Diagnostics.AddError("Failed to update FreeIPA role description", err.Error())
			return
		}
	}

	// Privilege update
	var planPrivs, statePrivs []string
	if !plan.Privileges.IsNull() && !plan.Privileges.IsUnknown() {
		plan.Privileges.ElementsAs(ctx, &planPrivs, false)
	}
	if !state.Privileges.IsNull() && !state.Privileges.IsUnknown() {
		state.Privileges.ElementsAs(ctx, &statePrivs, false)
	}

	privsAdded := difference(planPrivs, statePrivs)
	privsRemoved := difference(statePrivs, planPrivs)

	for _, p := range privsAdded {
		err := r.client.Call(ctx, "role_add_privilege", []string{plan.ID.ValueString()}, map[string]interface{}{
			"privilege": p,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add privilege to FreeIPA role", err.Error())
			return
		}
	}

	for _, p := range privsRemoved {
		err := r.client.Call(ctx, "role_remove_privilege", []string{plan.ID.ValueString()}, map[string]interface{}{
			"privilege": p,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove privilege from FreeIPA role", err.Error())
			return
		}
	}

	// Member updates
	memberTypes := []struct {
		PlanVal   types.Set
		StateVal  types.Set
		ParamName string
	}{
		{plan.Users, state.Users, "user"},
		{plan.Groups, state.Groups, "group"},
		{plan.Hosts, state.Hosts, "host"},
		{plan.HostGroups, state.HostGroups, "hostgroup"},
	}

	for _, mt := range memberTypes {
		var planM, stateM []string
		if !mt.PlanVal.IsNull() && !mt.PlanVal.IsUnknown() {
			mt.PlanVal.ElementsAs(ctx, &planM, false)
		}
		if !mt.StateVal.IsNull() && !mt.StateVal.IsUnknown() {
			mt.StateVal.ElementsAs(ctx, &stateM, false)
		}

		added := difference(planM, stateM)
		removed := difference(stateM, planM)

		if len(added) > 0 {
			err := r.client.Call(ctx, "role_add_member", []string{plan.ID.ValueString()}, map[string]interface{}{
				mt.ParamName: added,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add members to FreeIPA role", err.Error())
				return
			}
		}

		if len(removed) > 0 {
			err := r.client.Call(ctx, "role_remove_member", []string{plan.ID.ValueString()}, map[string]interface{}{
				mt.ParamName: removed,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to remove members from FreeIPA role", err.Error())
				return
			}
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "role_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA role", err.Error())
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
