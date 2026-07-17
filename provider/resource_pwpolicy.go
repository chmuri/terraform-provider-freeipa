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

type PwPolicyResource struct {
	client *client.Client
}

type pwPolicyResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	MinLife      types.Int64  `tfsdk:"minlife"`
	MaxLife      types.Int64  `tfsdk:"maxlife"`
	MinLength    types.Int64  `tfsdk:"minlength"`
	History      types.Int64  `tfsdk:"history"`
	Priority     types.Int64  `tfsdk:"priority"`
	MinClasses   types.Int64  `tfsdk:"minclasses"`
	MaxFail      types.Int64  `tfsdk:"maxfail"`
	FailInterval types.Int64  `tfsdk:"failinterval"`
	LockoutTime  types.Int64  `tfsdk:"lockouttime"`
}

func NewPwPolicyResource() resource.Resource {
	return &PwPolicyResource{}
}

func (r *PwPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (r *PwPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA Password Policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the password policy (group name or 'global').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the policy (user group or 'global').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"minlife": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum password lifetime in hours.",
			},
			"maxlife": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maximum password lifetime in days.",
			},
			"minlength": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum password length.",
			},
			"history": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Number of previous passwords to store.",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Priority of the policy.",
			},
			"minclasses": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Minimum number of character classes.",
			},
			"maxfail": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Maximum login failures before lockout.",
			},
			"failinterval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Failure interval in seconds.",
			},
			"lockouttime": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Lockout duration in seconds.",
			},
		},
	}
}

func (r *PwPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func getPwPolicyArgs(name string) []string {
	if name == "global" {
		return []string{}
	}
	return []string{name}
}

func parsePwPolicyVal(res map[string]interface{}, preferredKey, fallbackKey string) int64 {
	if val, ok := res[preferredKey]; ok && val != nil {
		return parseIntVal(val)
	}
	if val, ok := res[fallbackKey]; ok && val != nil {
		return parseIntVal(val)
	}
	return 0
}

func (r *PwPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pwPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.MinLife.IsNull() && !plan.MinLife.IsUnknown() {
		opts["krbminpwdlife"] = plan.MinLife.ValueInt64()
	}
	if !plan.MaxLife.IsNull() && !plan.MaxLife.IsUnknown() {
		opts["krbmaxpwdlife"] = plan.MaxLife.ValueInt64()
	}
	if !plan.MinLength.IsNull() && !plan.MinLength.IsUnknown() {
		opts["krbpwdminlength"] = plan.MinLength.ValueInt64()
	}
	if !plan.History.IsNull() && !plan.History.IsUnknown() {
		opts["krbpwdhistorylength"] = plan.History.ValueInt64()
	}
	if !plan.Priority.IsNull() && !plan.Priority.IsUnknown() {
		opts["cospriority"] = plan.Priority.ValueInt64()
	} else {
		opts["cospriority"] = int64(1)
	}
	if !plan.MinClasses.IsNull() && !plan.MinClasses.IsUnknown() {
		opts["krbpwdmindiffchars"] = plan.MinClasses.ValueInt64()
	}
	if !plan.MaxFail.IsNull() && !plan.MaxFail.IsUnknown() {
		opts["krbpwdmaxfailure"] = plan.MaxFail.ValueInt64()
	}
	if !plan.FailInterval.IsNull() && !plan.FailInterval.IsUnknown() {
		opts["krbpwdfailurecountinterval"] = plan.FailInterval.ValueInt64()
	}
	if !plan.LockoutTime.IsNull() && !plan.LockoutTime.IsUnknown() {
		opts["krbpwdlockoutduration"] = plan.LockoutTime.ValueInt64()
	}

	if plan.Name.ValueString() == "global" {
		// Global policy always exists, we modify it
		err := r.client.Call(ctx, "pwpolicy_mod", getPwPolicyArgs("global"), opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update global password policy", err.Error())
			return
		}
	} else {
		err := r.client.Call(ctx, "pwpolicy_add", getPwPolicyArgs(plan.Name.ValueString()), opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to create FreeIPA password policy", err.Error())
			return
		}
	}

	plan.ID = plan.Name

	if plan.MinLife.IsUnknown() {
		plan.MinLife = types.Int64Null()
	}
	if plan.MaxLife.IsUnknown() {
		plan.MaxLife = types.Int64Null()
	}
	if plan.MinLength.IsUnknown() {
		plan.MinLength = types.Int64Null()
	}
	if plan.History.IsUnknown() {
		plan.History = types.Int64Null()
	}
	if plan.Priority.IsUnknown() {
		plan.Priority = types.Int64Value(1)
	}
	if plan.MinClasses.IsUnknown() {
		plan.MinClasses = types.Int64Null()
	}
	if plan.MaxFail.IsUnknown() {
		plan.MaxFail = types.Int64Null()
	}
	if plan.FailInterval.IsUnknown() {
		plan.FailInterval = types.Int64Null()
	}
	if plan.LockoutTime.IsUnknown() {
		plan.LockoutTime = types.Int64Null()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PwPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pwPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var rawResult map[string]interface{}
	err := r.client.Call(ctx, "pwpolicy_show", getPwPolicyArgs(state.ID.ValueString()), map[string]interface{}{"all": true}, &rawResult)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA password policy", err.Error())
		return
	}

	res, ok := rawResult["result"].(map[string]interface{})
	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = state.ID
	state.MinLife = types.Int64Value(parsePwPolicyVal(res, "minlife", "krbminpwdlife"))
	state.MaxLife = types.Int64Value(parsePwPolicyVal(res, "maxlife", "krbmaxpwdlife"))
	state.MinLength = types.Int64Value(parsePwPolicyVal(res, "minlength", "krbpwdminlength"))
	state.History = types.Int64Value(parsePwPolicyVal(res, "history", "krbpwdhistorylength"))
	state.Priority = types.Int64Value(parsePwPolicyVal(res, "priority", "cospriority"))
	state.MinClasses = types.Int64Value(parsePwPolicyVal(res, "minclasses", "krbpwdmindiffchars"))
	state.MaxFail = types.Int64Value(parsePwPolicyVal(res, "maxfail", "krbpwdmaxfailure"))
	state.FailInterval = types.Int64Value(parsePwPolicyVal(res, "failinterval", "krbpwdfailurecountinterval"))
	state.LockoutTime = types.Int64Value(parsePwPolicyVal(res, "lockouttime", "krbpwdlockoutduration"))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PwPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state pwPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.MinLife.Equal(state.MinLife) {
		if plan.MinLife.IsNull() {
			opts["krbminpwdlife"] = ""
		} else {
			opts["krbminpwdlife"] = plan.MinLife.ValueInt64()
		}
	}
	if !plan.MaxLife.Equal(state.MaxLife) {
		if plan.MaxLife.IsNull() {
			opts["krbmaxpwdlife"] = ""
		} else {
			opts["krbmaxpwdlife"] = plan.MaxLife.ValueInt64()
		}
	}
	if !plan.MinLength.Equal(state.MinLength) {
		if plan.MinLength.IsNull() {
			opts["krbpwdminlength"] = ""
		} else {
			opts["krbpwdminlength"] = plan.MinLength.ValueInt64()
		}
	}
	if !plan.History.Equal(state.History) {
		if plan.History.IsNull() {
			opts["krbpwdhistorylength"] = ""
		} else {
			opts["krbpwdhistorylength"] = plan.History.ValueInt64()
		}
	}
	if !plan.Priority.Equal(state.Priority) {
		if plan.Priority.IsNull() {
			opts["cospriority"] = ""
		} else {
			opts["cospriority"] = plan.Priority.ValueInt64()
		}
	}
	if !plan.MinClasses.Equal(state.MinClasses) {
		if plan.MinClasses.IsNull() {
			opts["krbpwdmindiffchars"] = ""
		} else {
			opts["krbpwdmindiffchars"] = plan.MinClasses.ValueInt64()
		}
	}
	if !plan.MaxFail.Equal(state.MaxFail) {
		if plan.MaxFail.IsNull() {
			opts["krbpwdmaxfailure"] = ""
		} else {
			opts["krbpwdmaxfailure"] = plan.MaxFail.ValueInt64()
		}
	}
	if !plan.FailInterval.Equal(state.FailInterval) && !plan.FailInterval.IsUnknown() {
		if plan.FailInterval.IsNull() {
			opts["krbpwdfailurecountinterval"] = ""
		} else {
			opts["krbpwdfailurecountinterval"] = plan.FailInterval.ValueInt64()
		}
	}
	if !plan.LockoutTime.Equal(state.LockoutTime) && !plan.LockoutTime.IsUnknown() {
		if plan.LockoutTime.IsNull() {
			opts["krbpwdlockoutduration"] = ""
		} else {
			opts["krbpwdlockoutduration"] = plan.LockoutTime.ValueInt64()
		}
	}

	if len(opts) > 0 {
		err := r.client.Call(ctx, "pwpolicy_mod", getPwPolicyArgs(plan.ID.ValueString()), opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA password policy", err.Error())
			return
		}
	}

	if plan.MinLife.IsUnknown() {
		plan.MinLife = state.MinLife
	}
	if plan.History.IsUnknown() {
		plan.History = state.History
	}
	if plan.Priority.IsUnknown() {
		plan.Priority = state.Priority
	}
	if plan.MinClasses.IsUnknown() {
		plan.MinClasses = state.MinClasses
	}
	if plan.MaxFail.IsUnknown() {
		plan.MaxFail = state.MaxFail
	}
	if plan.FailInterval.IsUnknown() {
		plan.FailInterval = state.FailInterval
	}
	if plan.LockoutTime.IsUnknown() {
		plan.LockoutTime = state.LockoutTime
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PwPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pwPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.ValueString() == "global" {
		// Do not delete the global password policy
		return
	}

	err := r.client.Call(ctx, "pwpolicy_del", getPwPolicyArgs(state.ID.ValueString()), nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA password policy", err.Error())
		return
	}
}

func (r *PwPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
