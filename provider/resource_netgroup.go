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

type NetgroupResource struct {
	client *client.Client
}

type netgroupResourceModel struct {
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

func NewNetgroupResource() resource.Resource {
	return &NetgroupResource{}
}

func (r *NetgroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_netgroup"
}

func (r *NetgroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA Netgroups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the netgroup.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the netgroup (cn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the netgroup.",
			},
			"nisdomainname": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "NIS domain name (nisdomainname).",
			},
			"users": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of usernames associated with this netgroup.",
			},
			"groups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of group names associated with this netgroup.",
			},
			"hosts": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host FQDNs associated with this netgroup.",
			},
			"hostgroups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of hostgroups associated with this netgroup.",
			},
			"netgroups": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of netgroups associated with this netgroup.",
			},
		},
	}
}

func (r *NetgroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPANetgroupResult struct {
	Result struct {
		Cn              interface{} `json:"cn"`
		Description     interface{} `json:"description"`
		NisDomainName   interface{} `json:"nisdomainname"`
		MemberUser      interface{} `json:"memberuser_user"`
		MemberGroup     interface{} `json:"memberuser_group"`
		MemberHost      interface{} `json:"memberhost_host"`
		MemberHostGroup interface{} `json:"memberhost_hostgroup"`
		MemberNetgroup  interface{} `json:"member_netgroup"`
	} `json:"result"`
}

func (r *NetgroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan netgroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}
	if !plan.NisDomainName.IsNull() && !plan.NisDomainName.IsUnknown() {
		opts["nisdomainname"] = plan.NisDomainName.ValueString()
	}

	err := r.client.Call(ctx, "netgroup_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA netgroup", err.Error())
		return
	}

	// Add members
	memberOpts := map[string]interface{}{}
	var users, groups, hosts, hostgroups, netgroups []string
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
	if !plan.Netgroups.IsNull() && !plan.Netgroups.IsUnknown() {
		plan.Netgroups.ElementsAs(ctx, &netgroups, false)
		if len(netgroups) > 0 {
			memberOpts["netgroup"] = netgroups
		}
	}

	if len(memberOpts) > 0 {
		err = r.client.Call(ctx, "netgroup_add_member", []string{plan.Name.ValueString()}, memberOpts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add members to netgroup", err.Error())
			return
		}
	}

	plan.ID = plan.Name

	var result FreeIPANetgroupResult
	err = r.client.Call(ctx, "netgroup_show", []string{plan.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA netgroup after create", err.Error())
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
	if res.NisDomainName != nil {
		plan.NisDomainName = types.StringValue(parseStringVal(res.NisDomainName))
	} else {
		plan.NisDomainName = types.StringNull()
	}

	users = parseStringSlice(res.MemberUser)
	if len(users) > 0 {
		usersVal, d := types.SetValueFrom(ctx, types.StringType, users)
		resp.Diagnostics.Append(d...)
		plan.Users = usersVal
	} else {
		plan.Users = types.SetNull(types.StringType)
	}

	hosts = parseStringSlice(res.MemberHost)
	if len(hosts) > 0 {
		hostsVal, d := types.SetValueFrom(ctx, types.StringType, hosts)
		resp.Diagnostics.Append(d...)
		plan.Hosts = hostsVal
	} else {
		plan.Hosts = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetgroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state netgroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPANetgroupResult
	err := r.client.Call(ctx, "netgroup_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA netgroup", err.Error())
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

func (r *NetgroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state netgroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update description / nisdomainname if changed
	opts := map[string]interface{}{}
	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			opts["description"] = ""
		} else {
			opts["description"] = plan.Description.ValueString()
		}
	}
	if !plan.NisDomainName.Equal(state.NisDomainName) {
		if plan.NisDomainName.IsNull() {
			opts["nisdomainname"] = ""
		} else {
			opts["nisdomainname"] = plan.NisDomainName.ValueString()
		}
	}

	if len(opts) > 0 {
		err := r.client.Call(ctx, "netgroup_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil && !isEmptyModlistError(err) {
			resp.Diagnostics.AddError("Failed to update FreeIPA netgroup properties", err.Error())
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
		{plan.Netgroups, state.Netgroups, "netgroup"},
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
			err := r.client.Call(ctx, "netgroup_add_member", []string{plan.ID.ValueString()}, map[string]interface{}{
				mt.ParamName: added,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add members to FreeIPA netgroup", err.Error())
				return
			}
		}

		if len(removed) > 0 {
			err := r.client.Call(ctx, "netgroup_remove_member", []string{plan.ID.ValueString()}, map[string]interface{}{
				mt.ParamName: removed,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to remove members from FreeIPA netgroup", err.Error())
				return
			}
		}
	}

	if plan.NisDomainName.IsUnknown() {
		plan.NisDomainName = state.NisDomainName
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetgroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state netgroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "netgroup_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA netgroup", err.Error())
		return
	}
}

func (r *NetgroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
