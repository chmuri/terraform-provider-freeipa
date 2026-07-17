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

type SudoCommandResource struct {
	client *client.Client
}

type sudoCommandResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Command     types.String `tfsdk:"command"`
	Description types.String `tfsdk:"description"`
}

func NewSudoCommandResource() resource.Resource {
	return &SudoCommandResource{}
}

func (r *SudoCommandResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sudo_command"
}

func (r *SudoCommandResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA sudo commands.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique command identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The absolute path of the sudo command (e.g. /usr/bin/ls).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the sudo command.",
			},
		},
	}
}

func (r *SudoCommandResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SudoCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sudoCommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}

	err := r.client.Call(ctx, "sudocmd_add", []string{plan.Command.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA sudo command", err.Error())
		return
	}

	plan.ID = plan.Command
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SudoCommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sudoCommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Result struct {
			SudoCmd     interface{} `json:"sudocmd"`
			Description interface{} `json:"description"`
		} `json:"result"`
	}

	err := r.client.Call(ctx, "sudocmd_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA sudo command", err.Error())
		return
	}

	state.Command = types.StringValue(parseStringVal(result.Result.SudoCmd))
	state.ID = state.Command
	if result.Result.Description != nil {
		state.Description = types.StringValue(parseStringVal(result.Result.Description))
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SudoCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state sudoCommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		opts := map[string]interface{}{}
		if plan.Description.IsNull() {
			opts["description"] = ""
		} else {
			opts["description"] = plan.Description.ValueString()
		}

		err := r.client.Call(ctx, "sudocmd_mod", []string{plan.Command.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA sudo command", err.Error())
			return
		}
	}

	plan.ID = plan.Command
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SudoCommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sudoCommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "sudocmd_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA sudo command", err.Error())
	}
}

func (r *SudoCommandResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
