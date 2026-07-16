package provider

import (
	"context"

	"github.com/beeripa/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HostGroupResource struct {
	client *client.Client
}

type hostGroupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	CN             types.String `tfsdk:"cn"`
	Description    types.String `tfsdk:"description"`
	Hosts          types.Set    `tfsdk:"hosts"`
	HostGroups     types.Set    `tfsdk:"host_groups"`
	MemberManagers types.Set    `tfsdk:"member_managers"`
}

func NewHostGroupResource() resource.Resource {
	return &HostGroupResource{}
}

func (r *HostGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hostgroup"
}

func (r *HostGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA host groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique CN of the host group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the host group (cn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the host group.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host FQDNs belonging to this host group.",
			},
			"host_groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host group names belonging to this host group.",
			},
			"member_managers": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of usernames or group names allowed to manage this host group.",
			},
		},
	}
}

func (r *HostGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type hostGroupResourceResult struct {
	Result struct {
		Cn                 []string `json:"cn"`
		Description        []string `json:"description"`
		MemberHost         []string `json:"member_host"`
		MemberHostGroup    []string `json:"member_hostgroup"`
		MemberManagerUser  []string `json:"membermanager_user"`
		MemberManagerGroup []string `json:"membermanager_group"`
	} `json:"result"`
}

func (r *HostGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}

	err := r.client.Call(ctx, "hostgroup_add", []string{plan.CN.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA host group", err.Error())
		return
	}

	// Add members (hosts, hostgroups)
	var hosts []string
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		plan.Hosts.ElementsAs(ctx, &hosts, false)
	}
	var hostGroups []string
	if !plan.HostGroups.IsNull() && !plan.HostGroups.IsUnknown() {
		plan.HostGroups.ElementsAs(ctx, &hostGroups, false)
	}

	if len(hosts) > 0 || len(hostGroups) > 0 {
		memberOpts := map[string]interface{}{}
		if len(hosts) > 0 {
			memberOpts["host"] = hosts
		}
		if len(hostGroups) > 0 {
			memberOpts["hostgroup"] = hostGroups
		}
		err = r.client.Call(ctx, "hostgroup_add_member", []string{plan.CN.ValueString()}, memberOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add members to host group", err.Error())
			return
		}
	}

	// Add member managers
	var managers []string
	if !plan.MemberManagers.IsNull() && !plan.MemberManagers.IsUnknown() {
		plan.MemberManagers.ElementsAs(ctx, &managers, false)
	}
	for _, mgr := range managers {
		// Try to add as user manager
		err = r.client.Call(ctx, "hostgroup_add_member_manager", []string{plan.CN.ValueString()}, map[string]interface{}{
			"user": []string{mgr},
		}, nil)
		if err != nil {
			// Try as group manager
			errGroup := r.client.Call(ctx, "hostgroup_add_member_manager", []string{plan.CN.ValueString()}, map[string]interface{}{
				"group": []string{mgr},
			}, nil)
			if errGroup != nil {
				resp.Diagnostics.AddError("Failed to add member manager (tried user and group)", err.Error()+" / "+errGroup.Error())
				return
			}
		}
	}

	plan.ID = plan.CN

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HostGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result hostGroupResourceResult
	err := r.client.Call(ctx, "hostgroup_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA host group", err.Error())
		return
	}

	res := result.Result
	if len(res.Cn) > 0 {
		state.CN = types.StringValue(res.Cn[0])
		state.ID = types.StringValue(res.Cn[0])
	}
	if len(res.Description) > 0 {
		state.Description = types.StringValue(res.Description[0])
	} else {
		state.Description = types.StringNull()
	}

	if len(res.MemberHost) > 0 {
		hostsVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberHost)
		resp.Diagnostics.Append(d...)
		state.Hosts = hostsVal
	} else {
		state.Hosts = types.SetNull(types.StringType)
	}

	if len(res.MemberHostGroup) > 0 {
		hgVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberHostGroup)
		resp.Diagnostics.Append(d...)
		state.HostGroups = hgVal
	} else {
		state.HostGroups = types.SetNull(types.StringType)
	}

	// Merge membermanager_user and membermanager_group
	var combinedManagers []string
	combinedManagers = append(combinedManagers, res.MemberManagerUser...)
	combinedManagers = append(combinedManagers, res.MemberManagerGroup...)

	if len(combinedManagers) > 0 {
		mgrVal, d := types.SetValueFrom(ctx, types.StringType, combinedManagers)
		resp.Diagnostics.Append(d...)
		state.MemberManagers = mgrVal
	} else {
		state.MemberManagers = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *HostGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state hostGroupResourceModel
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
		err := r.client.Call(ctx, "hostgroup_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA host group description", err.Error())
			return
		}
	}

	// Hosts
	var planHosts, stateHosts []string
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		plan.Hosts.ElementsAs(ctx, &planHosts, false)
	}
	if !state.Hosts.IsNull() && !state.Hosts.IsUnknown() {
		state.Hosts.ElementsAs(ctx, &stateHosts, false)
	}
	hostsAdded := difference(planHosts, stateHosts)
	hostsRemoved := difference(stateHosts, planHosts)

	// HostGroups
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
		err := r.client.Call(ctx, "hostgroup_add_member", []string{plan.ID.ValueString()}, addOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add members to host group", err.Error())
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
		err := r.client.Call(ctx, "hostgroup_remove_member", []string{plan.ID.ValueString()}, remOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove members from host group", err.Error())
			return
		}
	}

	// Member managers
	var planManagers, stateManagers []string
	if !plan.MemberManagers.IsNull() && !plan.MemberManagers.IsUnknown() {
		plan.MemberManagers.ElementsAs(ctx, &planManagers, false)
	}
	if !state.MemberManagers.IsNull() && !state.MemberManagers.IsUnknown() {
		state.MemberManagers.ElementsAs(ctx, &stateManagers, false)
	}
	mgrsAdded := difference(planManagers, stateManagers)
	mgrsRemoved := difference(stateManagers, planManagers)

	if len(mgrsAdded) > 0 || len(mgrsRemoved) > 0 {
		// Fetch current manager groups/users to know how to remove them
		var current hostGroupResourceResult
		err := r.client.Call(ctx, "hostgroup_show", []string{plan.ID.ValueString()}, map[string]interface{}{"all": true}, &current)
		if err != nil {
			resp.Diagnostics.AddError("Failed to show host group managers for update", err.Error())
			return
		}

		// Process Removals
		for _, rem := range mgrsRemoved {
			isUser := false
			isGroup := false
			for _, u := range current.Result.MemberManagerUser {
				if u == rem {
					isUser = true
					break
				}
			}
			for _, g := range current.Result.MemberManagerGroup {
				if g == rem {
					isGroup = true
					break
				}
			}

			if isUser {
				err = r.client.Call(ctx, "hostgroup_remove_member_manager", []string{plan.ID.ValueString()}, map[string]interface{}{
					"user": []string{rem},
				}, nil)
				if err != nil {
					resp.Diagnostics.AddError("Failed to remove user member manager", err.Error())
					return
				}
			}
			if isGroup {
				err = r.client.Call(ctx, "hostgroup_remove_member_manager", []string{plan.ID.ValueString()}, map[string]interface{}{
					"group": []string{rem},
				}, nil)
				if err != nil {
					resp.Diagnostics.AddError("Failed to remove group member manager", err.Error())
					return
				}
			}
		}

		// Process Additions
		for _, add := range mgrsAdded {
			// Try user first
			err = r.client.Call(ctx, "hostgroup_add_member_manager", []string{plan.ID.ValueString()}, map[string]interface{}{
				"user": []string{add},
			}, nil)
			if err != nil {
				// Try group
				errGroup := r.client.Call(ctx, "hostgroup_add_member_manager", []string{plan.ID.ValueString()}, map[string]interface{}{
					"group": []string{add},
				}, nil)
				if errGroup != nil {
					resp.Diagnostics.AddError("Failed to add member manager (tried user and group)", err.Error()+" / "+errGroup.Error())
					return
				}
			}
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HostGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "hostgroup_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA host group", err.Error())
		return
	}
}

func (r *HostGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
