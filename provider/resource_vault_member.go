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

type VaultMemberResource struct {
	client *client.Client
}

type vaultMemberResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Users    types.Set    `tfsdk:"users"`
	Groups   types.Set    `tfsdk:"groups"`
	Services types.Set    `tfsdk:"services"`
}

func NewVaultMemberResource() resource.Resource {
	return &VaultMemberResource{}
}

func (r *VaultMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vault_member"
}

func (r *VaultMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages members of a FreeIPA vault. This resource does **not** create the vault itself; it only adds or removes members.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier (vault name).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the vault to which members are attached.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of usernames that are members of the vault.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of group names that are members of the vault.",
			},
			"services": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of service principals that are members of the vault.",
			},
		},
	}
}

func (r *VaultMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VaultMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vaultMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	memberOpts := map[string]interface{}{}
	var u, g, s []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &u, false)
		if len(u) > 0 {
			memberOpts["users"] = u
		}
	}
	if !plan.Groups.IsNull() && !plan.Groups.IsUnknown() {
		plan.Groups.ElementsAs(ctx, &g, false)
		if len(g) > 0 {
			memberOpts["groups"] = g
		}
	}
	if !plan.Services.IsNull() && !plan.Services.IsUnknown() {
		plan.Services.ElementsAs(ctx, &s, false)
		if len(s) > 0 {
			memberOpts["services"] = s
		}
	}
	if len(memberOpts) == 0 {
		resp.Diagnostics.AddError("No members supplied", "At least one member (user, group or service) must be defined.")
		return
	}
	err := r.client.Call(ctx, "vault_add_member", []string{plan.Name.ValueString()}, memberOpts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add members to FreeIPA vault", err.Error())
		return
	}
	plan.ID = plan.Name
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VaultMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vaultMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Result struct {
			MemberUser    interface{} `json:"member_user"`
			MemberGroup   interface{} `json:"member_group"`
			MemberService interface{} `json:"member_service"`
		} `json:"result"`
	}
	err := r.client.Call(ctx, "vault_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read vault members", err.Error())
		return
	}

	// Populate sets
	membersU := parseStringSlice(result.Result.MemberUser)
	if len(membersU) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, membersU)
		resp.Diagnostics.Append(d...)
		state.Users = val
	} else {
		state.Users = types.SetNull(types.StringType)
	}

	membersG := parseStringSlice(result.Result.MemberGroup)
	if len(membersG) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, membersG)
		resp.Diagnostics.Append(d...)
		state.Groups = val
	} else {
		state.Groups = types.SetNull(types.StringType)
	}

	membersS := parseStringSlice(result.Result.MemberService)
	if len(membersS) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, membersS)
		resp.Diagnostics.Append(d...)
		state.Services = val
	} else {
		state.Services = types.SetNull(types.StringType)
	}

	state.ID = state.Name
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *VaultMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vaultMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compute diffs for each attribute
	type diffInfo struct {
		Plan, State types.Set
		Param       string
	}
	diffs := []diffInfo{
		{plan.Users, state.Users, "users"},
		{plan.Groups, state.Groups, "groups"},
		{plan.Services, state.Services, "services"},
	}
	for _, d := range diffs {
		var planVals, stateVals []string
		if !d.Plan.IsNull() && !d.Plan.IsUnknown() {
			d.Plan.ElementsAs(ctx, &planVals, false)
		}
		if !d.State.IsNull() && !d.State.IsUnknown() {
			d.State.ElementsAs(ctx, &stateVals, false)
		}
		added := difference(planVals, stateVals)
		removed := difference(stateVals, planVals)
		if len(added) > 0 {
			err := r.client.Call(ctx, "vault_add_member", []string{plan.Name.ValueString()}, map[string]interface{}{d.Param: added}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add members", err.Error())
				return
			}
		}
		if len(removed) > 0 {
			err := r.client.Call(ctx, "vault_remove_member", []string{plan.Name.ValueString()}, map[string]interface{}{d.Param: removed}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to remove members", err.Error())
				return
			}
		}
	}
	// Preserve immutable fields
	plan.ID = plan.Name
	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VaultMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vaultMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Remove all members – we rely on current state to know what to remove
	var users, groups, services []string
	if !state.Users.IsNull() && !state.Users.IsUnknown() {
		state.Users.ElementsAs(ctx, &users, false)
	}
	if !state.Groups.IsNull() && !state.Groups.IsUnknown() {
		state.Groups.ElementsAs(ctx, &groups, false)
	}
	if !state.Services.IsNull() && !state.Services.IsUnknown() {
		state.Services.ElementsAs(ctx, &services, false)
	}
	opts := map[string]interface{}{}
	if len(users) > 0 {
		opts["users"] = users
	}
	if len(groups) > 0 {
		opts["groups"] = groups
	}
	if len(services) > 0 {
		opts["services"] = services
	}
	if len(opts) > 0 {
		err := r.client.Call(ctx, "vault_remove_member", []string{state.Name.ValueString()}, opts, nil)
		if err != nil && !isNotFoundError(err) {
			resp.Diagnostics.AddError("Failed to remove members during delete", err.Error())
		}
	}
}

func (r *VaultMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
