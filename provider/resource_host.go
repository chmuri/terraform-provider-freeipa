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

type HostResource struct {
	client *client.Client
}

type hostResourceModel struct {
	ID              types.String `tfsdk:"id"`
	FQDN            types.String `tfsdk:"fqdn"`
	Description     types.String `tfsdk:"description"`
	Password        types.String `tfsdk:"password"`
	IPAddress       types.String `tfsdk:"ip_address"`
	Locality        types.String `tfsdk:"locality"`
	Location        types.String `tfsdk:"location"`
	Platform        types.String `tfsdk:"platform"`
	OS              types.String `tfsdk:"os"`
	MacAddress      types.String `tfsdk:"mac_address"`
	SSHPublicKeys   types.List   `tfsdk:"ssh_public_keys"`
	UserCertificate types.String `tfsdk:"user_certificate"`
	Force           types.Bool   `tfsdk:"force"`
	ManagedBy       types.Set    `tfsdk:"managed_by"`
}

func NewHostResource() resource.Resource {
	return &HostResource{}
}

func (r *HostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *HostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA hosts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique FQDN identifying the host.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fqdn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Fully Qualified Domain Name of the host.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the host.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "One-Time Password (OTP) generated to join the realm, or user-defined password.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_address": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IP address to associate with the host in DNS.",
			},
			"locality": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Locality of the host.",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Location of the host.",
			},
			"platform": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Hardware platform of the host.",
			},
			"os": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Operating system and version of the host.",
			},
			"mac_address": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "MAC address of the host.",
			},
			"ssh_public_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of SSH public keys.",
			},
			"user_certificate": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Base-64 encoded host certificate.",
			},
			"force": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Force host addition even if DNS verification fails.",
			},
			"managed_by": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of host FQDNs that can manage this host.",
			},
		},
	}
}

func (r *HostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPAHostResult struct {
	Result struct {
		Fqdn           []string `json:"fqdn"`
		Description    []string `json:"description"`
		RandomPassword string   `json:"randompassword"`
	} `json:"result"`
}

type FreeIPAHostShowResult struct {
	Result struct {
		Fqdn               []string `json:"fqdn"`
		Description        []string `json:"description"`
		L                  []string `json:"l"`
		Nshostlocation     []string `json:"nshostlocation"`
		Nshardwareplatform []string `json:"nshardwareplatform"`
		Nsosversion        []string `json:"nsosversion"`
		Macaddress         []string `json:"macaddress"`
		Ipasshpubkey       []string `json:"ipasshpubkey"`
		Usercertificate    []string `json:"usercertificate"`
		ManagedbyHost      []string `json:"managedby_host"`
	} `json:"result"`
}

func (r *HostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}
	if !plan.Locality.IsNull() && !plan.Locality.IsUnknown() {
		opts["locality"] = plan.Locality.ValueString()
	}
	if !plan.Location.IsNull() && !plan.Location.IsUnknown() {
		opts["location"] = plan.Location.ValueString()
	}
	if !plan.Platform.IsNull() && !plan.Platform.IsUnknown() {
		opts["platform"] = plan.Platform.ValueString()
	}
	if !plan.OS.IsNull() && !plan.OS.IsUnknown() {
		opts["os"] = plan.OS.ValueString()
	}
	if !plan.MacAddress.IsNull() && !plan.MacAddress.IsUnknown() {
		opts["macaddress"] = plan.MacAddress.ValueString()
	}
	if !plan.SSHPublicKeys.IsNull() && !plan.SSHPublicKeys.IsUnknown() {
		var sshKeys []string
		plan.SSHPublicKeys.ElementsAs(ctx, &sshKeys, false)
		opts["ipasshpubkey"] = sshKeys
	}
	if !plan.UserCertificate.IsNull() && !plan.UserCertificate.IsUnknown() {
		opts["usercertificate"] = plan.UserCertificate.ValueString()
	}
	if !plan.IPAddress.IsNull() && !plan.IPAddress.IsUnknown() {
		opts["ip_address"] = plan.IPAddress.ValueString()
	}

	// If no password is specified, ask FreeIPA to generate a random OTP
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		opts["random"] = true
	} else {
		opts["userpassword"] = plan.Password.ValueString()
	}

	if !plan.Force.IsNull() && !plan.Force.IsUnknown() {
		opts["force"] = plan.Force.ValueBool()
	} else {
		opts["force"] = true
	}

	var result FreeIPAHostResult
	err := r.client.Call(ctx, "host_add", []string{plan.FQDN.ValueString()}, opts, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA host", err.Error())
		return
	}

	// Add managedby hosts
	var managedBy []string
	if !plan.ManagedBy.IsNull() && !plan.ManagedBy.IsUnknown() {
		plan.ManagedBy.ElementsAs(ctx, &managedBy, false)
	}
	if len(managedBy) > 0 {
		err = r.client.Call(ctx, "host_add_managedby", []string{plan.FQDN.ValueString()}, map[string]interface{}{
			"host": managedBy,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add managedby hosts to FreeIPA host", err.Error())
			return
		}
	}

	plan.ID = plan.FQDN

	// Capture generated OTP if random was used
	if plan.Password.IsNull() || plan.Password.IsUnknown() {
		if result.Result.RandomPassword != "" {
			plan.Password = types.StringValue(result.Result.RandomPassword)
		} else {
			plan.Password = types.StringNull()
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAHostShowResult
	err := r.client.Call(ctx, "host_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA host", err.Error())
		return
	}

	res := result.Result
	if len(res.Fqdn) > 0 {
		state.FQDN = types.StringValue(res.Fqdn[0])
		state.ID = types.StringValue(res.Fqdn[0])
	}
	if len(res.Description) > 0 {
		state.Description = types.StringValue(res.Description[0])
	} else {
		state.Description = types.StringNull()
	}

	if len(res.L) > 0 {
		state.Locality = types.StringValue(res.L[0])
	} else {
		state.Locality = types.StringNull()
	}

	if len(res.Nshostlocation) > 0 {
		state.Location = types.StringValue(res.Nshostlocation[0])
	} else {
		state.Location = types.StringNull()
	}

	if len(res.Nshardwareplatform) > 0 {
		state.Platform = types.StringValue(res.Nshardwareplatform[0])
	} else {
		state.Platform = types.StringNull()
	}

	if len(res.Nsosversion) > 0 {
		state.OS = types.StringValue(res.Nsosversion[0])
	} else {
		state.OS = types.StringNull()
	}

	if len(res.Macaddress) > 0 {
		state.MacAddress = types.StringValue(res.Macaddress[0])
	} else {
		state.MacAddress = types.StringNull()
	}

	if len(res.Ipasshpubkey) > 0 {
		sshKeysVal, d := types.ListValueFrom(ctx, types.StringType, res.Ipasshpubkey)
		resp.Diagnostics.Append(d...)
		state.SSHPublicKeys = sshKeysVal
	} else {
		state.SSHPublicKeys = types.ListNull(types.StringType)
	}

	if len(res.Usercertificate) > 0 {
		state.UserCertificate = types.StringValue(res.Usercertificate[0])
	} else {
		state.UserCertificate = types.StringNull()
	}

	if len(res.ManagedbyHost) > 0 {
		managedByVal, d := types.SetValueFrom(ctx, types.StringType, res.ManagedbyHost)
		resp.Diagnostics.Append(d...)
		state.ManagedBy = managedByVal
	} else {
		state.ManagedBy = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *HostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state hostResourceModel
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
	if !plan.Locality.Equal(state.Locality) {
		if plan.Locality.IsNull() {
			opts["locality"] = ""
		} else {
			opts["locality"] = plan.Locality.ValueString()
		}
	}
	if !plan.Location.Equal(state.Location) {
		if plan.Location.IsNull() {
			opts["location"] = ""
		} else {
			opts["location"] = plan.Location.ValueString()
		}
	}
	if !plan.Platform.Equal(state.Platform) {
		if plan.Platform.IsNull() {
			opts["platform"] = ""
		} else {
			opts["platform"] = plan.Platform.ValueString()
		}
	}
	if !plan.OS.Equal(state.OS) {
		if plan.OS.IsNull() {
			opts["os"] = ""
		} else {
			opts["os"] = plan.OS.ValueString()
		}
	}
	if !plan.MacAddress.Equal(state.MacAddress) {
		if plan.MacAddress.IsNull() {
			opts["macaddress"] = ""
		} else {
			opts["macaddress"] = plan.MacAddress.ValueString()
		}
	}
	if !plan.SSHPublicKeys.Equal(state.SSHPublicKeys) {
		if plan.SSHPublicKeys.IsNull() {
			opts["ipasshpubkey"] = []string{}
		} else {
			var sshKeys []string
			plan.SSHPublicKeys.ElementsAs(ctx, &sshKeys, false)
			opts["ipasshpubkey"] = sshKeys
		}
	}
	if !plan.UserCertificate.Equal(state.UserCertificate) {
		if plan.UserCertificate.IsNull() {
			opts["usercertificate"] = ""
		} else {
			opts["usercertificate"] = plan.UserCertificate.ValueString()
		}
	}

	if !plan.Force.IsNull() && !plan.Force.IsUnknown() {
		opts["force"] = plan.Force.ValueBool()
	}

	if len(opts) > 0 {
		err := r.client.Call(ctx, "host_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA host", err.Error())
			return
		}
	}

	// Calculate managedby delta
	var planManagedBy, stateManagedBy []string
	if !plan.ManagedBy.IsNull() && !plan.ManagedBy.IsUnknown() {
		plan.ManagedBy.ElementsAs(ctx, &planManagedBy, false)
	}
	if !state.ManagedBy.IsNull() && !state.ManagedBy.IsUnknown() {
		state.ManagedBy.ElementsAs(ctx, &stateManagedBy, false)
	}
	addedManagedBy := difference(planManagedBy, stateManagedBy)
	removedManagedBy := difference(stateManagedBy, planManagedBy)

	if len(addedManagedBy) > 0 {
		err := r.client.Call(ctx, "host_add_managedby", []string{plan.ID.ValueString()}, map[string]interface{}{
			"host": addedManagedBy,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add managedby hosts to FreeIPA host", err.Error())
			return
		}
	}

	if len(removedManagedBy) > 0 {
		err := r.client.Call(ctx, "host_remove_managedby", []string{plan.ID.ValueString()}, map[string]interface{}{
			"host": removedManagedBy,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove managedby hosts from FreeIPA host", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "host_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA host", err.Error())
		return
	}
}

func (r *HostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

