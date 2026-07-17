package provider

import (
	"context"
	"strings"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DnsZoneDataSource struct {
	client *client.Client
}

type dnsZoneDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	ZoneName                types.String `tfsdk:"zone_name"`
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

func NewDnsZoneDataSource() datasource.DataSource {
	return &DnsZoneDataSource{}
}

func (d *DnsZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (d *DnsZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA DNS Zone.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique zone name.",
			},
			"zone_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the DNS zone.",
			},
			"authoritative_nameserver": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Authoritative nameserver.",
			},
			"admin_email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Administrator e-mail address.",
			},
			"refresh": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "SOA record refresh time.",
			},
			"retry": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "SOA record retry time.",
			},
			"expire": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "SOA record expire time.",
			},
			"minimum": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "How long negative responses should be cached.",
			},
			"ttl": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Time to live for records at the zone apex.",
			},
			"default_ttl": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Default time to live for records without an explicit TTL.",
			},
			"dynamic_update": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Allow dynamic updates.",
			},
			"allow_query": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Semicolon-separated list of IP addresses or networks allowed to issue queries.",
			},
			"allow_transfer": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Semicolon-separated list of IP addresses or networks allowed to transfer the zone.",
			},
			"allow_sync_ptr": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Allow synchronization of forward and reverse records.",
			},
			"forwarders": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of DNS forwarders.",
			},
			"forward_policy": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Forwarding policy: only, first, none.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the DNS zone is enabled/active.",
			},
		},
	}
}

func (d *DnsZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DnsZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dnsZoneDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPADnsZoneResult
	err := d.client.Call(ctx, "dnszone_show", []string{state.ZoneName.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA DNS zone", err.Error())
		return
	}

	res := result.Result
	zoneName := strings.TrimSuffix(parseStringVal(res.IdnsName), ".")
	state.ZoneName = types.StringValue(zoneName)
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

	if res.IdnssoaRefresh != nil {
		state.Refresh = types.Int64Value(parseIntVal(res.IdnssoaRefresh))
	} else {
		state.Refresh = types.Int64Null()
	}
	if res.IdnssoaRetry != nil {
		state.Retry = types.Int64Value(parseIntVal(res.IdnssoaRetry))
	} else {
		state.Retry = types.Int64Null()
	}
	if res.IdnssoaExpire != nil {
		state.Expire = types.Int64Value(parseIntVal(res.IdnssoaExpire))
	} else {
		state.Expire = types.Int64Null()
	}
	if res.IdnssoaMinTTL != nil {
		state.Minimum = types.Int64Value(parseIntVal(res.IdnssoaMinTTL))
	} else {
		state.Minimum = types.Int64Null()
	}
	if res.DnsTTL != nil {
		state.TTL = types.Int64Value(parseIntVal(res.DnsTTL))
	} else {
		state.TTL = types.Int64Null()
	}
	if res.IdnsDefaultTTL != nil {
		state.DefaultTTL = types.Int64Value(parseIntVal(res.IdnsDefaultTTL))
	} else {
		state.DefaultTTL = types.Int64Null()
	}
	if res.IdnsDynamicUpdate != nil {
		state.DynamicUpdate = types.BoolValue(parseBoolVal(res.IdnsDynamicUpdate))
	} else {
		state.DynamicUpdate = types.BoolNull()
	}

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

	if res.IdnsAllowSyncPtr != nil {
		state.AllowSyncPtr = types.BoolValue(parseBoolVal(res.IdnsAllowSyncPtr))
	} else {
		state.AllowSyncPtr = types.BoolNull()
	}

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
