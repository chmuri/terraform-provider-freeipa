package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type HbacRuleResource struct {
	client *client.Client
}

type hbacResourceModel struct {
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

func NewHbacRuleResource() resource.Resource {
	return &HbacRuleResource{}
}

func (r *HbacRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hbacrule"
}

func (r *HbacRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA HBAC (Host-Based Access Control) rules.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the HBAC rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the HBAC rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the HBAC rule.",
			},
			"host_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Host category (e.g. 'all' to apply to all hosts, or null).",
			},
			"user_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User category (e.g. 'all' to apply to all users, or null).",
			},
			"service_category": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Service category (e.g. 'all' to apply to all services, or null).",
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
			"services": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of HBAC service names associated with the rule.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the HBAC rule is enabled.",
			},
		},
	}
}

func (r *HbacRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPAHbacRuleResult struct {
	Result struct {
		Ipauniqueid     []string `json:"ipauniqueid"`
		Cn              []string `json:"cn"`
		Description     []string `json:"description"`
		Hostcategory    []string `json:"hostcategory"`
		Usercategory    []string `json:"usercategory"`
		Servicecategory []string `json:"servicecategory"`
		MemberUser      []string `json:"memberuser_user"`
		MemberGroup     []string `json:"memberuser_group"`
		MemberHost      []string `json:"memberhost_host"`
		MemberService   []string `json:"memberservice_hbacservice"`
		Ipaenabledflag  []bool   `json:"ipaenabledflag"`
	} `json:"result"`
}

func (r *HbacRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hbacResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}
	if !plan.HostCategory.IsNull() && !plan.HostCategory.IsUnknown() {
		opts["hostcategory"] = plan.HostCategory.ValueString()
	}
	if !plan.UserCategory.IsNull() && !plan.UserCategory.IsUnknown() {
		opts["usercategory"] = plan.UserCategory.ValueString()
	}
	if !plan.ServiceCategory.IsNull() && !plan.ServiceCategory.IsUnknown() {
		opts["servicecategory"] = plan.ServiceCategory.ValueString()
	}

	err := r.client.Call(ctx, "hbacrule_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA HBAC rule", err.Error())
		return
	}

	// Enable or disable rule based on state
	if !plan.Enabled.ValueBool() {
		err = r.client.Call(ctx, "hbacrule_disable", []string{plan.Name.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to disable FreeIPA HBAC rule", err.Error())
			return
		}
	}

	// Add associations (users, groups, hosts, services)
	if plan.UserCategory.IsNull() || plan.UserCategory.ValueString() != "all" {
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
			err = r.client.Call(ctx, "hbacrule_add_user", []string{plan.Name.ValueString()}, assocOpts, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to associate users/groups with HBAC rule", err.Error())
				return
			}
		}
	}

	if plan.HostCategory.IsNull() || plan.HostCategory.ValueString() != "all" {
		var hosts []string
		if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
			plan.Hosts.ElementsAs(ctx, &hosts, false)
		}
		if len(hosts) > 0 {
			err = r.client.Call(ctx, "hbacrule_add_host", []string{plan.Name.ValueString()}, map[string]interface{}{
				"host": hosts,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to associate hosts with HBAC rule", err.Error())
				return
			}
		}
	}

	if plan.ServiceCategory.IsNull() || plan.ServiceCategory.ValueString() != "all" {
		var services []string
		if !plan.Services.IsNull() && !plan.Services.IsUnknown() {
			plan.Services.ElementsAs(ctx, &services, false)
		}
		if len(services) > 0 {
			err = r.client.Call(ctx, "hbacrule_add_service", []string{plan.Name.ValueString()}, map[string]interface{}{
				"hbacservice": services,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to associate services with HBAC rule", err.Error())
				return
			}
		}
	}

	plan.ID = plan.Name

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HbacRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hbacResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAHbacRuleResult
	err := r.client.Call(ctx, "hbacrule_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA HBAC rule", err.Error())
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

	if len(res.Hostcategory) > 0 {
		state.HostCategory = types.StringValue(res.Hostcategory[0])
	} else {
		state.HostCategory = types.StringNull()
	}
	if len(res.Usercategory) > 0 {
		state.UserCategory = types.StringValue(res.Usercategory[0])
	} else {
		state.UserCategory = types.StringNull()
	}
	if len(res.Servicecategory) > 0 {
		state.ServiceCategory = types.StringValue(res.Servicecategory[0])
	} else {
		state.ServiceCategory = types.StringNull()
	}

	if len(res.Ipaenabledflag) > 0 {
		state.Enabled = types.BoolValue(res.Ipaenabledflag[0])
	} else {
		state.Enabled = types.BoolValue(false)
	}

	// Populate memberships
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

	if len(res.MemberHost) > 0 {
		hostsVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberHost)
		resp.Diagnostics.Append(d...)
		state.Hosts = hostsVal
	} else {
		state.Hosts = types.SetNull(types.StringType)
	}

	if len(res.MemberService) > 0 {
		servicesVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberService)
		resp.Diagnostics.Append(d...)
		state.Services = servicesVal
	} else {
		state.Services = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *HbacRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state hbacResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Mod attributes if changed
	modOpts := map[string]interface{}{}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			modOpts["description"] = ""
		} else {
			modOpts["description"] = plan.Description.ValueString()
		}
	}
	if !plan.HostCategory.Equal(state.HostCategory) {
		if plan.HostCategory.IsNull() {
			modOpts["hostcategory"] = ""
		} else {
			modOpts["hostcategory"] = plan.HostCategory.ValueString()
		}
	}
	if !plan.UserCategory.Equal(state.UserCategory) {
		if plan.UserCategory.IsNull() {
			modOpts["usercategory"] = ""
		} else {
			modOpts["usercategory"] = plan.UserCategory.ValueString()
		}
	}
	if !plan.ServiceCategory.Equal(state.ServiceCategory) {
		if plan.ServiceCategory.IsNull() {
			modOpts["servicecategory"] = ""
		} else {
			modOpts["servicecategory"] = plan.ServiceCategory.ValueString()
		}
	}

	if len(modOpts) > 0 {
		err := r.client.Call(ctx, "hbacrule_mod", []string{plan.ID.ValueString()}, modOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA HBAC rule attributes", err.Error())
			return
		}
	}

	// Update Enabled status
	if !plan.Enabled.Equal(state.Enabled) {
		method := "hbacrule_enable"
		if !plan.Enabled.ValueBool() {
			method = "hbacrule_disable"
		}
		err := r.client.Call(ctx, method, []string{plan.ID.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to toggle FreeIPA HBAC rule enabled flag", err.Error())
			return
		}
	}

	// Update associations (Users/Groups, Hosts, Services)
	// 1. Users
	var planUsers, stateUsers []string
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() {
		plan.Users.ElementsAs(ctx, &planUsers, false)
	}
	if !state.Users.IsNull() && !state.Users.IsUnknown() {
		state.Users.ElementsAs(ctx, &stateUsers, false)
	}
	usersAdded := difference(planUsers, stateUsers)
	usersRemoved := difference(stateUsers, planUsers)

	// 2. Groups
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
		err := r.client.Call(ctx, "hbacrule_add_user", []string{plan.ID.ValueString()}, addOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add user/group associations to HBAC rule", err.Error())
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
		err := r.client.Call(ctx, "hbacrule_remove_user", []string{plan.ID.ValueString()}, remOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove user/group associations from HBAC rule", err.Error())
			return
		}
	}

	// 3. Hosts
	var planHosts, stateHosts []string
	if !plan.Hosts.IsNull() && !plan.Hosts.IsUnknown() {
		plan.Hosts.ElementsAs(ctx, &planHosts, false)
	}
	if !state.Hosts.IsNull() && !state.Hosts.IsUnknown() {
		state.Hosts.ElementsAs(ctx, &stateHosts, false)
	}
	hostsAdded := difference(planHosts, stateHosts)
	hostsRemoved := difference(stateHosts, planHosts)

	if len(hostsAdded) > 0 {
		err := r.client.Call(ctx, "hbacrule_add_host", []string{plan.ID.ValueString()}, map[string]interface{}{
			"host": hostsAdded,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add host associations to HBAC rule", err.Error())
			return
		}
	}
	if len(hostsRemoved) > 0 {
		err := r.client.Call(ctx, "hbacrule_remove_host", []string{plan.ID.ValueString()}, map[string]interface{}{
			"host": hostsRemoved,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove host associations from HBAC rule", err.Error())
			return
		}
	}

	// 4. Services
	var planServices, stateServices []string
	if !plan.Services.IsNull() && !plan.Services.IsUnknown() {
		plan.Services.ElementsAs(ctx, &planServices, false)
	}
	if !state.Services.IsNull() && !state.Services.IsUnknown() {
		state.Services.ElementsAs(ctx, &stateServices, false)
	}
	servicesAdded := difference(planServices, stateServices)
	servicesRemoved := difference(stateServices, planServices)

	if len(servicesAdded) > 0 {
		err := r.client.Call(ctx, "hbacrule_add_service", []string{plan.ID.ValueString()}, map[string]interface{}{
			"hbacservice": servicesAdded,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add service associations to HBAC rule", err.Error())
			return
		}
	}
	if len(servicesRemoved) > 0 {
		err := r.client.Call(ctx, "hbacrule_remove_service", []string{plan.ID.ValueString()}, map[string]interface{}{
			"hbacservice": servicesRemoved,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove service associations from HBAC rule", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HbacRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hbacResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "hbacrule_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA HBAC rule", err.Error())
		return
	}
}

func (r *HbacRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
