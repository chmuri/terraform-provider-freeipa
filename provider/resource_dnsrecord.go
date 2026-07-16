package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/beeripa/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DnsRecordResource struct {
	client *client.Client
}

type dnsRecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ZoneName    types.String `tfsdk:"zone_name"`
	Name        types.String `tfsdk:"name"`
	TTL         types.Int64  `tfsdk:"ttl"`
	RecordType  types.String `tfsdk:"record_type"`
	RecordValue types.String `tfsdk:"record_value"`
}

func NewDnsRecordResource() resource.Resource {
	return &DnsRecordResource{}
}

func (r *DnsRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DnsRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA DNS Records.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Format: zone_name:name:record_type:record_value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The DNS zone name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the DNS record.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time to live.",
			},
			"record_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "DNS record type: A, AAAA, CNAME, MX, SRV, TXT, PTR, NS, SSHFP.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"record_value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "DNS record value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DnsRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func getRecordField(recordType string) string {
	switch strings.ToUpper(recordType) {
	case "A":
		return "arecord"
	case "AAAA":
		return "aaaarecord"
	case "CNAME":
		return "cnamerecord"
	case "MX":
		return "mxrecord"
	case "SRV":
		return "srvrecord"
	case "TXT":
		return "txtrecord"
	case "PTR":
		return "ptrrecord"
	case "NS":
		return "nsrecord"
	case "SSHFP":
		return "sshfprecord"
	default:
		return ""
	}
}

func parseDnsRecordID(id string) (zone, name, rtype, val string, err error) {
	parts := strings.SplitN(id, ":", 4)
	if len(parts) < 4 {
		return "", "", "", "", fmt.Errorf("invalid dns record id: %s", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func (r *DnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	field := getRecordField(plan.RecordType.ValueString())
	if field == "" {
		resp.Diagnostics.AddError("Unsupported record type", plan.RecordType.ValueString())
		return
	}

	opts := map[string]interface{}{
		field: plan.RecordValue.ValueString(),
	}
	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		opts["dnsttl"] = plan.TTL.ValueInt64()
	}

	err := r.client.Call(ctx, "dnsrecord_add", []string{plan.ZoneName.ValueString(), plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA DNS record", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s:%s:%s",
		plan.ZoneName.ValueString(),
		plan.Name.ValueString(),
		plan.RecordType.ValueString(),
		plan.RecordValue.ValueString(),
	))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, name, rtype, val, err := parseDnsRecordID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid state ID", err.Error())
		return
	}

	field := getRecordField(rtype)
	if field == "" {
		resp.Diagnostics.AddError("Unsupported record type", rtype)
		return
	}

	var rawResult map[string]interface{}
	err = r.client.Call(ctx, "dnsrecord_show", []string{zone, name}, map[string]interface{}{"all": true}, &rawResult)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA DNS record", err.Error())
		return
	}

	// Extract result object
	resultMap, ok := rawResult["result"].(map[string]interface{})
	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	recordValues := parseStringSlice(resultMap[field])
	if !contains(recordValues, val) {
		// Record value no longer exists
		resp.State.RemoveResource(ctx)
		return
	}

	state.ZoneName = types.StringValue(zone)
	state.Name = types.StringValue(name)
	state.RecordType = types.StringValue(rtype)
	state.RecordValue = types.StringValue(val)

	if ttlVal, ok := resultMap["dnsttl"]; ok && ttlVal != nil {
		state.TTL = types.Int64Value(parseIntVal(ttlVal))
	} else {
		state.TTL = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *DnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dnsRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.TTL.Equal(state.TTL) {
		field := getRecordField(plan.RecordType.ValueString())
		if field == "" {
			resp.Diagnostics.AddError("Unsupported record type", plan.RecordType.ValueString())
			return
		}

		opts := map[string]interface{}{}
		if plan.TTL.IsNull() {
			opts["dnsttl"] = ""
		} else {
			opts["dnsttl"] = plan.TTL.ValueInt64()
		}

		err := r.client.Call(ctx, "dnsrecord_mod", []string{plan.ZoneName.ValueString(), plan.Name.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update DNS record TTL", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone, name, rtype, val, err := parseDnsRecordID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid state ID", err.Error())
		return
	}

	field := getRecordField(rtype)
	if field == "" {
		resp.Diagnostics.AddError("Unsupported record type", rtype)
		return
	}

	opts := map[string]interface{}{
		field: val,
	}

	err = r.client.Call(ctx, "dnsrecord_del", []string{zone, name}, opts, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA DNS record", err.Error())
		return
	}
}

func (r *DnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import via ID: zone_name:name:record_type:record_value
	parts := strings.SplitN(req.ID, ":", 4)
	if len(parts) < 4 {
		resp.Diagnostics.AddError("Invalid Import ID", "Must be formatted as zone_name:name:record_type:record_value")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("record_type"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("record_value"), parts[3])...)
}
