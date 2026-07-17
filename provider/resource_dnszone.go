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

type DnsZoneResource struct {
	client *client.Client
}

type dnsZoneResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	ZoneName                types.String `tfsdk:"zone_name"`
	SkipOverlapCheck        types.Bool   `tfsdk:"skip_overlap_check"`
	SkipNameserverCheck     types.Bool   `tfsdk:"skip_nameserver_check"`
	Force                   types.Bool   `tfsdk:"force"`
	AuthoritativeNameserver types.String `tfsdk:"authoritative_nameserver"`
	AdminEmail              types.String `tfsdk:"admin_email"`
	Refresh                 types.Int64  `tfsdk:"refresh"`
	Retry                   types.Int64  `tfsdk:"retry"`
	Expire                  types.Int64  `tfsdk:"expire"`
	Minimum                 types.Int64  `tfsdk:"minimum"`
	TTL                     types.Int64  `tfsdk:"ttl"`
	DefaultTTL              types.Int64  `tfsdk:"default_ttl"`
	DynamicUpdate           types.Bool   `tfsdk:"dynamic_update"`
	AllowQuery              types.String `tfsdk:"allow_query"`
	AllowTransfer           types.String `tfsdk:"allow_transfer"`
	AllowSyncPtr            types.Bool   `tfsdk:"allow_sync_ptr"`
	Forwarders              types.Set    `tfsdk:"forwarders"`
	ForwardPolicy           types.String `tfsdk:"forward_policy"`
	Enabled                 types.Bool   `tfsdk:"enabled"`
}

func NewDnsZoneResource() resource.Resource {
	return &DnsZoneResource{}
}

func (r *DnsZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (r *DnsZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA DNS Zones.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique zone name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the DNS zone.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"skip_overlap_check": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Force DNS zone creation even if it overlaps with an existing zone.",
			},
			"skip_nameserver_check": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Force DNS zone creation even if the nameserver is not resolvable.",
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Force zone creation.",
			},
			"authoritative_nameserver": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Authoritative nameserver (idnssoamname).",
			},
			"admin_email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Administrator e-mail address (idnssoarname).",
			},
			"refresh": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "SOA record refresh time (idnssoarefresh).",
			},
			"retry": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "SOA record retry time (idnssoaretry).",
			},
			"expire": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "SOA record expire time (idnssoaexpire).",
			},
			"minimum": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "How long negative responses should be cached (idnssoaminttl).",
			},
			"ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time to live for records at the zone apex (dnsttl).",
			},
			"default_ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default time to live for records without an explicit TTL (idnsdefaultttl).",
			},
			"dynamic_update": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow dynamic updates (idnsdynamicupdate).",
			},
			"allow_query": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Semicolon-separated list of IP addresses or networks allowed to issue queries (idnsallowquery).",
			},
			"allow_transfer": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Semicolon-separated list of IP addresses or networks allowed to transfer the zone (idnsallowtransfer).",
			},
			"allow_sync_ptr": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow synchronization of forward and reverse records (idnsallowsyncptr).",
			},
			"forwarders": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "List of DNS forwarders (idnsforwarders).",
			},
			"forward_policy": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Forwarding policy: only, first, none (idnsforwardpolicy).",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the DNS zone is enabled/active.",
			},
		},
	}
}

func (r *DnsZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPADnsZoneResult struct {
	Result struct {
		IdnsName          interface{} `json:"idnsname"`
		IdnssoaMName      interface{} `json:"idnssoamname"`
		IdnssoaRname    interface{} `json:"idnssoarname"`
		IdnssoaRefresh    interface{} `json:"idnssoarefresh"`
		IdnssoaRetry      interface{} `json:"idnssoaretry"`
		IdnssoaExpire     interface{} `json:"idnssoaexpire"`
		IdnssoaMinTTL     interface{} `json:"idnssoaminttl"`
		DnsTTL            interface{} `json:"dnsttl"`
		IdnsDefaultTTL    interface{} `json:"idnsdefaultttl"`
		IdnsDynamicUpdate interface{} `json:"idnsdynamicupdate"`
		IdnsAllowQuery    interface{} `json:"idnsallowquery"`
		IdnsAllowTransfer interface{} `json:"idnsallowtransfer"`
		IdnsAllowSyncPtr  interface{} `json:"idnsallowsyncptr"`
		IdnsForwarders    interface{} `json:"idnsforwarders"`
		IdnsForwardPolicy interface{} `json:"idnsforwardpolicy"`
		IdnsZoneActive    interface{} `json:"idnszoneactive"`
	} `json:"result"`
}

func (r *DnsZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsZoneResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.SkipOverlapCheck.IsNull() && !plan.SkipOverlapCheck.IsUnknown() {
		opts["skip_overlap_check"] = plan.SkipOverlapCheck.ValueBool()
	}
	if !plan.SkipNameserverCheck.IsNull() && !plan.SkipNameserverCheck.IsUnknown() {
		opts["skip_nameserver_check"] = plan.SkipNameserverCheck.ValueBool()
	}
	if !plan.Force.IsNull() && !plan.Force.IsUnknown() {
		opts["force"] = plan.Force.ValueBool()
	}
	if !plan.AuthoritativeNameserver.IsNull() && !plan.AuthoritativeNameserver.IsUnknown() {
		opts["idnssoamname"] = plan.AuthoritativeNameserver.ValueString()
	}
	adminEmail := ""
	if !plan.AdminEmail.IsNull() && !plan.AdminEmail.IsUnknown() {
		adminEmail = plan.AdminEmail.ValueString()
	}
	if !plan.Refresh.IsNull() && !plan.Refresh.IsUnknown() {
		opts["idnssoarefresh"] = plan.Refresh.ValueInt64()
	}
	if !plan.Retry.IsNull() && !plan.Retry.IsUnknown() {
		opts["idnssoaretry"] = plan.Retry.ValueInt64()
	}
	if !plan.Expire.IsNull() && !plan.Expire.IsUnknown() {
		opts["idnssoaexpire"] = plan.Expire.ValueInt64()
	}
	if !plan.Minimum.IsNull() && !plan.Minimum.IsUnknown() {
		opts["idnssoaminttl"] = plan.Minimum.ValueInt64()
	}
	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		opts["dnsttl"] = plan.TTL.ValueInt64()
	}
	if !plan.DefaultTTL.IsNull() && !plan.DefaultTTL.IsUnknown() {
		opts["idnsdefaultttl"] = plan.DefaultTTL.ValueInt64()
	}
	if !plan.DynamicUpdate.IsNull() && !plan.DynamicUpdate.IsUnknown() {
		opts["idnsdynamicupdate"] = plan.DynamicUpdate.ValueBool()
	}
	if !plan.AllowQuery.IsNull() && !plan.AllowQuery.IsUnknown() {
		opts["idnsallowquery"] = plan.AllowQuery.ValueString()
	}
	if !plan.AllowTransfer.IsNull() && !plan.AllowTransfer.IsUnknown() {
		opts["idnsallowtransfer"] = plan.AllowTransfer.ValueString()
	}
	if !plan.AllowSyncPtr.IsNull() && !plan.AllowSyncPtr.IsUnknown() {
		opts["idnsallowsyncptr"] = plan.AllowSyncPtr.ValueBool()
	}
	if !plan.Forwarders.IsNull() && !plan.Forwarders.IsUnknown() {
		var fw []string
		plan.Forwarders.ElementsAs(ctx, &fw, false)
		if len(fw) > 0 {
			opts["idnsforwarders"] = fw
		}
	}
	if !plan.ForwardPolicy.IsNull() && !plan.ForwardPolicy.IsUnknown() {
		opts["idnsforwardpolicy"] = plan.ForwardPolicy.ValueString()
	}

	err := r.client.Call(ctx, "dnszone_add", []string{plan.ZoneName.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA DNS zone", err.Error())
		return
	}

	if adminEmail != "" {
		err = r.client.Call(ctx, "dnszone_mod", []string{plan.ZoneName.ValueString()}, map[string]interface{}{
			"idnssoarname": adminEmail,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to set admin email on FreeIPA DNS zone", err.Error())
			return
		}
	}

	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && !plan.Enabled.ValueBool() {
		err = r.client.Call(ctx, "dnszone_disable", []string{plan.ZoneName.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to disable FreeIPA DNS zone", err.Error())
			return
		}
	}

	plan.ID = plan.ZoneName

	if plan.Refresh.IsUnknown() {
		plan.Refresh = types.Int64Null()
	}
	if plan.Retry.IsUnknown() {
		plan.Retry = types.Int64Null()
	}
	if plan.Expire.IsUnknown() {
		plan.Expire = types.Int64Null()
	}
	if plan.Minimum.IsUnknown() {
		plan.Minimum = types.Int64Null()
	}
	if plan.TTL.IsUnknown() {
		plan.TTL = types.Int64Null()
	}
	if plan.DefaultTTL.IsUnknown() {
		plan.DefaultTTL = types.Int64Null()
	}
	if plan.DynamicUpdate.IsUnknown() {
		plan.DynamicUpdate = types.BoolNull()
	}
	if plan.AllowQuery.IsUnknown() {
		plan.AllowQuery = types.StringNull()
	}
	if plan.AllowTransfer.IsUnknown() {
		plan.AllowTransfer = types.StringNull()
	}
	if plan.AllowSyncPtr.IsUnknown() {
		plan.AllowSyncPtr = types.BoolNull()
	}
	if plan.Forwarders.IsUnknown() {
		plan.Forwarders = types.SetNull(types.StringType)
	}
	if plan.ForwardPolicy.IsUnknown() {
		plan.ForwardPolicy = types.StringNull()
	}
	if plan.Enabled.IsUnknown() {
		plan.Enabled = types.BoolValue(true)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DnsZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsZoneResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPADnsZoneResult
	err := r.client.Call(ctx, "dnszone_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA DNS zone", err.Error())
		return
	}

	res := result.Result
	state.ZoneName = types.StringValue(parseStringVal(res.IdnsName))
	state.ID = state.ZoneName

	if res.IdnssoaMName != nil {
		state.AuthoritativeNameserver = types.StringValue(parseStringVal(res.IdnssoaMName))
	} else {
		state.AuthoritativeNameserver = types.StringNull()
	}

	if res.IdnssoaRname != nil {
		state.AdminEmail = types.StringValue(parseStringVal(res.IdnssoaRname))
	} else {
		state.AdminEmail = types.StringNull()
	}

	state.Refresh = types.Int64Value(parseIntVal(res.IdnssoaRefresh))
	state.Retry = types.Int64Value(parseIntVal(res.IdnssoaRetry))
	state.Expire = types.Int64Value(parseIntVal(res.IdnssoaExpire))
	state.Minimum = types.Int64Value(parseIntVal(res.IdnssoaMinTTL))
	state.TTL = types.Int64Value(parseIntVal(res.DnsTTL))
	state.DefaultTTL = types.Int64Value(parseIntVal(res.IdnsDefaultTTL))
	state.DynamicUpdate = types.BoolValue(parseBoolVal(res.IdnsDynamicUpdate))

	if res.IdnsAllowQuery != nil {
		state.AllowQuery = types.StringValue(parseStringVal(res.IdnsAllowQuery))
	} else {
		state.AllowQuery = types.StringNull()
	}

	if res.IdnsAllowTransfer != nil {
		state.AllowTransfer = types.StringValue(parseStringVal(res.IdnsAllowTransfer))
	} else {
		state.AllowTransfer = types.StringNull()
	}

	state.AllowSyncPtr = types.BoolValue(parseBoolVal(res.IdnsAllowSyncPtr))

	fwSlice := parseStringSlice(res.IdnsForwarders)
	if len(fwSlice) > 0 {
		fwVal, d := types.SetValueFrom(ctx, types.StringType, fwSlice)
		resp.Diagnostics.Append(d...)
		state.Forwarders = fwVal
	} else {
		state.Forwarders = types.SetNull(types.StringType)
	}

	if res.IdnsForwardPolicy != nil {
		state.ForwardPolicy = types.StringValue(parseStringVal(res.IdnsForwardPolicy))
	} else {
		state.ForwardPolicy = types.StringNull()
	}

	state.Enabled = types.BoolValue(parseBoolVal(res.IdnsZoneActive))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *DnsZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dnsZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.AuthoritativeNameserver.Equal(state.AuthoritativeNameserver) {
		if plan.AuthoritativeNameserver.IsNull() {
			opts["idnssoamname"] = ""
		} else {
			opts["idnssoamname"] = plan.AuthoritativeNameserver.ValueString()
		}
	}
	if !plan.AdminEmail.Equal(state.AdminEmail) {
		if plan.AdminEmail.IsNull() {
			opts["idnssoarname"] = ""
		} else {
			opts["idnssoarname"] = plan.AdminEmail.ValueString()
		}
	}
	if !plan.Refresh.Equal(state.Refresh) {
		if plan.Refresh.IsNull() {
			opts["idnssoarefresh"] = ""
		} else {
			opts["idnssoarefresh"] = plan.Refresh.ValueInt64()
		}
	}
	if !plan.Retry.Equal(state.Retry) {
		if plan.Retry.IsNull() {
			opts["idnssoaretry"] = ""
		} else {
			opts["idnssoaretry"] = plan.Retry.ValueInt64()
		}
	}
	if !plan.Expire.Equal(state.Expire) {
		if plan.Expire.IsNull() {
			opts["idnssoaexpire"] = ""
		} else {
			opts["idnssoaexpire"] = plan.Expire.ValueInt64()
		}
	}
	if !plan.Minimum.Equal(state.Minimum) {
		if plan.Minimum.IsNull() {
			opts["idnssoaminttl"] = ""
		} else {
			opts["idnssoaminttl"] = plan.Minimum.ValueInt64()
		}
	}
	if !plan.TTL.Equal(state.TTL) {
		if plan.TTL.IsNull() {
			opts["dnsttl"] = ""
		} else {
			opts["dnsttl"] = plan.TTL.ValueInt64()
		}
	}
	if !plan.DefaultTTL.Equal(state.DefaultTTL) {
		if plan.DefaultTTL.IsNull() {
			opts["idnsdefaultttl"] = ""
		} else {
			opts["idnsdefaultttl"] = plan.DefaultTTL.ValueInt64()
		}
	}
	if !plan.DynamicUpdate.Equal(state.DynamicUpdate) {
		if plan.DynamicUpdate.IsNull() {
			opts["idnsdynamicupdate"] = ""
		} else {
			opts["idnsdynamicupdate"] = plan.DynamicUpdate.ValueBool()
		}
	}
	if !plan.AllowQuery.Equal(state.AllowQuery) {
		if plan.AllowQuery.IsNull() {
			opts["idnsallowquery"] = ""
		} else {
			opts["idnsallowquery"] = plan.AllowQuery.ValueString()
		}
	}
	if !plan.AllowTransfer.Equal(state.AllowTransfer) {
		if plan.AllowTransfer.IsNull() {
			opts["idnsallowtransfer"] = ""
		} else {
			opts["idnsallowtransfer"] = plan.AllowTransfer.ValueString()
		}
	}
	if !plan.AllowSyncPtr.Equal(state.AllowSyncPtr) {
		if plan.AllowSyncPtr.IsNull() {
			opts["idnsallowsyncptr"] = ""
		} else {
			opts["idnsallowsyncptr"] = plan.AllowSyncPtr.ValueBool()
		}
	}
	if !plan.Forwarders.Equal(state.Forwarders) {
		if plan.Forwarders.IsNull() {
			opts["idnsforwarders"] = []string{}
		} else {
			var fw []string
			plan.Forwarders.ElementsAs(ctx, &fw, false)
			opts["idnsforwarders"] = fw
		}
	}
	if !plan.ForwardPolicy.Equal(state.ForwardPolicy) {
		if plan.ForwardPolicy.IsNull() {
			opts["idnsforwardpolicy"] = ""
		} else {
			opts["idnsforwardpolicy"] = plan.ForwardPolicy.ValueString()
		}
	}

	if len(opts) > 0 {
		err := r.client.Call(ctx, "dnszone_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA DNS zone", err.Error())
			return
		}
	}

	if !plan.Enabled.Equal(state.Enabled) {
		var method string
		if plan.Enabled.ValueBool() {
			method = "dnszone_enable"
		} else {
			method = "dnszone_disable"
		}
		err := r.client.Call(ctx, method, []string{plan.ID.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update DNS zone active status", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DnsZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "dnszone_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA DNS zone", err.Error())
		return
	}
}

func (r *DnsZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
