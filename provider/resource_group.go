package provider

import (
	"context"
	"strconv"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupResource struct {
	client *client.Client
}

type groupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Users          types.Set    `tfsdk:"users"`
	GidNumber      types.Int64  `tfsdk:"gid_number"`
	Nonposix       types.Bool   `tfsdk:"nonposix"`
	External       types.Bool   `tfsdk:"external"`
	Groups         types.Set    `tfsdk:"groups"`
	MemberManagers types.Set    `tfsdk:"member_managers"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA user groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique CN of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the group (cn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the group.",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of usernames belonging to this group.",
			},
			"gid_number": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "GID (group ID) number.",
			},
			"nonposix": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Create as a non-POSIX group.",
			},
			"external": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow adding external non-IPA members from trusted domains.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Nested groups belonging to this group.",
			},
			"member_managers": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Users that can manage members of this group.",
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPAGroupResult struct {
	Result struct {
		Cn                []string `json:"cn"`
		Description       []string `json:"description"`
		MemberUser        []string `json:"member_user"`
		MemberGroup       []string `json:"member_group"`
		GidNumber         []string `json:"gidnumber"`
		ObjectClass       []string `json:"objectclass"`
		MemberManagerUser []string `json:"membermanager_user"`
	} `json:"result"`
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}
	if !plan.GidNumber.IsNull() && !plan.GidNumber.IsUnknown() {
		opts["gidnumber"] = plan.GidNumber.ValueInt64()
	}
	if !plan.Nonposix.IsNull() && !plan.Nonposix.IsUnknown() && plan.Nonposix.ValueBool() {
		opts["nonposix"] = true
	}
	if !plan.External.IsNull() && !plan.External.IsUnknown() && plan.External.ValueBool() {
		opts["external"] = true
	}

	err := r.client.Call(ctx, "group_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA group", err.Error())
		return
	}

	// Add group members (users & nested groups)
	var users []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &users, false)
	}
	var groups []string
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		plan.Groups.ElementsAs(ctx, &groups, false)
	}

	if len(users) > 0 || len(groups) > 0 {
		mopts := map[string]interface{}{}
		if len(users) > 0 {
			mopts["user"] = users
		}
		if len(groups) > 0 {
			mopts["group"] = groups
		}
		err = r.client.Call(ctx, "group_add_member", []string{plan.Name.ValueString()}, mopts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add members to group", err.Error())
			return
		}
	}

	// Add member managers
	var managers []string
	if !plan.MemberManagers.IsNull() && !plan.MemberManagers.IsUnknown() {
		plan.MemberManagers.ElementsAs(ctx, &managers, false)
	}
	if len(managers) > 0 {
		err = r.client.Call(ctx, "group_add_member_manager", []string{plan.Name.ValueString()}, map[string]interface{}{
			"user": managers,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add member managers to group", err.Error())
			return
		}
	}

	plan.ID = plan.Name

	if plan.GidNumber.IsUnknown() {
		plan.GidNumber = types.Int64Null()
	}
	if plan.Nonposix.IsUnknown() {
		plan.Nonposix = types.BoolValue(false)
	}
	if plan.External.IsUnknown() {
		plan.External = types.BoolValue(false)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAGroupResult
	err := r.client.Call(ctx, "group_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA group", err.Error())
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

	if len(res.GidNumber) > 0 {
		val, err := strconv.ParseInt(res.GidNumber[0], 10, 64)
		if err == nil {
			state.GidNumber = types.Int64Value(val)
		}
	} else {
		state.GidNumber = types.Int64Null()
	}

	state.Nonposix = types.BoolValue(!contains(res.ObjectClass, "posixgroup"))
	state.External = types.BoolValue(contains(res.ObjectClass, "ipaexternalgroup"))

	if len(res.MemberUser) > 0 {
		usersVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberUser)
		resp.Diagnostics.Append(d...)
		state.Users = usersVal
	} else {
		state.Users = types.SetNull(types.StringType)
	}

	if len(res.MemberGroup) > 0 {
		groupsVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberGroup)
		resp.Diagnostics.Append(d...)
		state.Groups = groupsVal
	} else {
		state.Groups = types.SetNull(types.StringType)
	}

	if len(res.MemberManagerUser) > 0 {
		managersVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberManagerUser)
		resp.Diagnostics.Append(d...)
		state.MemberManagers = managersVal
	} else {
		state.MemberManagers = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			opts["description"] = ""
		} else {
			opts["description"] = plan.Description.ValueString()
		}
	}
	if !plan.GidNumber.Equal(state.GidNumber) && !plan.GidNumber.IsNull() && !plan.GidNumber.IsUnknown() {
		opts["gidnumber"] = plan.GidNumber.ValueInt64()
	}
	if !plan.Nonposix.Equal(state.Nonposix) && !plan.Nonposix.IsNull() && !plan.Nonposix.IsUnknown() {
		if !plan.Nonposix.ValueBool() && state.Nonposix.ValueBool() {
			opts["posix"] = true
		}
	}
	if !plan.External.Equal(state.External) && !plan.External.IsNull() && !plan.External.IsUnknown() {
		if plan.External.ValueBool() && !state.External.ValueBool() {
			opts["external"] = true
		}
	}

	if len(opts) > 0 {
		err := r.client.Call(ctx, "group_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil && !isEmptyModlistError(err) {
			resp.Diagnostics.AddError("Failed to update FreeIPA group attributes", err.Error())
			return
		}
	}

	// Calculate membership delta (users)
	var planUsers, stateUsers []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &planUsers, false)
	}
	if !state.Users.IsNull() && !state.Users.IsUnknown() {
		state.Users.ElementsAs(ctx, &stateUsers, false)
	}
	addedUsers := difference(planUsers, stateUsers)
	removedUsers := difference(stateUsers, planUsers)

	// Calculate membership delta (groups)
	var planGroups, stateGroups []string
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		plan.Groups.ElementsAs(ctx, &planGroups, false)
	}
	if !state.Groups.IsNull() && !state.Groups.IsUnknown() {
		state.Groups.ElementsAs(ctx, &stateGroups, false)
	}
	addedGroups := difference(planGroups, stateGroups)
	removedGroups := difference(stateGroups, planGroups)

	if len(addedUsers) > 0 || len(addedGroups) > 0 {
		mopts := map[string]interface{}{}
		if len(addedUsers) > 0 {
			mopts["user"] = addedUsers
		}
		if len(addedGroups) > 0 {
			mopts["group"] = addedGroups
		}
		err := r.client.Call(ctx, "group_add_member", []string{plan.ID.ValueString()}, mopts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add new members to FreeIPA group", err.Error())
			return
		}
	}

	if len(removedUsers) > 0 || len(removedGroups) > 0 {
		mopts := map[string]interface{}{}
		if len(removedUsers) > 0 {
			mopts["user"] = removedUsers
		}
		if len(removedGroups) > 0 {
			mopts["group"] = removedGroups
		}
		err := r.client.Call(ctx, "group_remove_member", []string{plan.ID.ValueString()}, mopts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove members from FreeIPA group", err.Error())
			return
		}
	}

	// Calculate member managers delta
	var planManagers, stateManagers []string
	if !plan.MemberManagers.IsNull() && !plan.MemberManagers.IsUnknown() {
		plan.MemberManagers.ElementsAs(ctx, &planManagers, false)
	}
	if !state.MemberManagers.IsNull() && !state.MemberManagers.IsUnknown() {
		state.MemberManagers.ElementsAs(ctx, &stateManagers, false)
	}
	addedManagers := difference(planManagers, stateManagers)
	removedManagers := difference(stateManagers, planManagers)

	if len(addedManagers) > 0 {
		err := r.client.Call(ctx, "group_add_member_manager", []string{plan.ID.ValueString()}, map[string]interface{}{
			"user": addedManagers,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add member managers to FreeIPA group", err.Error())
			return
		}
	}

	if len(removedManagers) > 0 {
		err := r.client.Call(ctx, "group_remove_member_manager", []string{plan.ID.ValueString()}, map[string]interface{}{
			"user": removedManagers,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove member managers from FreeIPA group", err.Error())
			return
		}
	}

	if plan.GidNumber.IsUnknown() {
		plan.GidNumber = state.GidNumber
	}
	if plan.Nonposix.IsUnknown() {
		plan.Nonposix = state.Nonposix
	}
	if plan.External.IsUnknown() {
		plan.External = state.External
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "group_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA group", err.Error())
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string
	m := make(map[string]bool)

	for _, val := range slice2 {
		m[val] = true
	}

	for _, val := range slice1 {
		if !m[val] {
			diff = append(diff, val)
		}
	}
	return diff
}
