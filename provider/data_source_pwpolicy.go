package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PwPolicyDataSource struct {
	client *client.Client
}

type pwPolicyDataSourceModel struct {
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

func NewPwPolicyDataSource() datasource.DataSource {
	return &PwPolicyDataSource{}
}

func (d *PwPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (d *PwPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Password Policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the password policy.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the policy (user group or 'global').",
			},
			"minlife": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Minimum password lifetime in hours.",
			},
			"maxlife": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum password lifetime in days.",
			},
			"minlength": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Minimum password length.",
			},
			"history": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of previous passwords to store.",
			},
			"priority": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Priority of the policy.",
			},
			"minclasses": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Minimum number of character classes.",
			},
			"maxfail": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Maximum login failures before lockout.",
			},
			"failinterval": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Failure interval in seconds.",
			},
			"lockouttime": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Lockout duration in seconds.",
			},
		},
	}
}

func (d *PwPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PwPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pwPolicyDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := getPwPolicyArgs(state.Name.ValueString())

	var rawResult map[string]interface{}
	err := d.client.Call(ctx, "pwpolicy_show", args, map[string]interface{}{"all": true}, &rawResult)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA password policy", err.Error())
		return
	}

	res, ok := rawResult["result"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("Failed to parse password policy result", "Unexpected response format")
		return
	}

	state.ID = state.Name
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
