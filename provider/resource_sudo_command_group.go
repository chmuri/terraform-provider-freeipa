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

type SudoCommandGroupResource struct {
	client *client.Client
}

type sudoCommandGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Commands    types.Set    `tfsdk:"commands"`
}

func NewSudoCommandGroupResource() resource.Resource {
	return &SudoCommandGroupResource{}
}

func (r *SudoCommandGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sudo_command_group"
}

func (r *SudoCommandGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA sudo command groups (sudocmdgroups).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier (group name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the sudo command group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the sudo command group.",
			},
			"commands": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of sudo command paths that belong to this group.",
			},
		},
	}
}

func (r *SudoCommandGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SudoCommandGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sudoCommandGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}

	err := r.client.Call(ctx, "sudocmdgroup_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA sudo command group", err.Error())
		return
	}

	// Add commands
	if !plan.Commands.IsNull() && !plan.Commands.IsUnknown() {
		var cmds []string
		plan.Commands.ElementsAs(ctx, &cmds, false)
		if len(cmds) > 0 {
			err = r.client.Call(ctx, "sudocmdgroup_add_member", []string{plan.Name.ValueString()},
				map[string]interface{}{"sudocmd": cmds}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add commands to sudo command group", err.Error())
				return
			}
		}
	}

	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SudoCommandGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sudoCommandGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Result struct {
			CN          interface{} `json:"cn"`
			Description interface{} `json:"description"`
			MemberCmd   interface{} `json:"member_sudocmd"`
		} `json:"result"`
	}

	err := r.client.Call(ctx, "sudocmdgroup_show", []string{state.ID.ValueString()},
		map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA sudo command group", err.Error())
		return
	}

	state.Name = types.StringValue(parseStringVal(result.Result.CN))
	state.ID = state.Name

	if result.Result.Description != nil {
		state.Description = types.StringValue(parseStringVal(result.Result.Description))
	} else {
		state.Description = types.StringNull()
	}

	cmds := parseStringSlice(result.Result.MemberCmd)
	if len(cmds) > 0 {
		val, d := types.SetValueFrom(ctx, types.StringType, cmds)
		resp.Diagnostics.Append(d...)
		state.Commands = val
	} else {
		state.Commands = types.SetNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SudoCommandGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state sudoCommandGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update description
	if !plan.Description.Equal(state.Description) {
		opts := map[string]interface{}{}
		if plan.Description.IsNull() {
			opts["description"] = ""
		} else {
			opts["description"] = plan.Description.ValueString()
		}
		err := r.client.Call(ctx, "sudocmdgroup_mod", []string{plan.Name.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update sudo command group", err.Error())
			return
		}
	}

	// Update commands membership
	var planCmds, stateCmds []string
	if !plan.Commands.IsNull() && !plan.Commands.IsUnknown() {
		plan.Commands.ElementsAs(ctx, &planCmds, false)
	}
	if !state.Commands.IsNull() && !state.Commands.IsUnknown() {
		state.Commands.ElementsAs(ctx, &stateCmds, false)
	}

	added := difference(planCmds, stateCmds)
	removed := difference(stateCmds, planCmds)

	if len(added) > 0 {
		err := r.client.Call(ctx, "sudocmdgroup_add_member", []string{plan.Name.ValueString()},
			map[string]interface{}{"sudocmd": added}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add commands to sudo command group", err.Error())
			return
		}
	}
	if len(removed) > 0 {
		err := r.client.Call(ctx, "sudocmdgroup_remove_member", []string{plan.Name.ValueString()},
			map[string]interface{}{"sudocmd": removed}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove commands from sudo command group", err.Error())
			return
		}
	}

	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *SudoCommandGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state sudoCommandGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "sudocmdgroup_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA sudo command group", err.Error())
	}
}

func (r *SudoCommandGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
