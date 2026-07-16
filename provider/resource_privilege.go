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

type PrivilegeResource struct {
	client *client.Client
}

type privilegeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func NewPrivilegeResource() resource.Resource {
	return &PrivilegeResource{}
}

func (r *PrivilegeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privilege"
}

func (r *PrivilegeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA Privilege resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the privilege.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the privilege (cn).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the privilege.",
			},
			"permissions": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of permission names associated with this privilege.",
			},
		},
	}
}

func (r *PrivilegeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPAPrivilegeResult struct {
	Result struct {
		Cn               interface{} `json:"cn"`
		Description      interface{} `json:"description"`
		MemberPermission interface{} `json:"member_permission"`
	} `json:"result"`
}

func (r *PrivilegeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan privilegeResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}

	err := r.client.Call(ctx, "privilege_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA privilege", err.Error())
		return
	}

	// Add permissions
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		var permissions []string
		plan.Permissions.ElementsAs(ctx, &permissions, false)
		for _, perm := range permissions {
			err = r.client.Call(ctx, "privilege_add_permission", []string{plan.Name.ValueString()}, map[string]interface{}{
				"permission": perm,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add permission to FreeIPA privilege", err.Error())
				return
			}
		}
	}

	plan.ID = plan.Name

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PrivilegeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state privilegeResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAPrivilegeResult
	err := r.client.Call(ctx, "privilege_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA privilege", err.Error())
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

	perms := parseStringSlice(res.MemberPermission)
	if len(perms) > 0 {
		permsVal, d := types.SetValueFrom(ctx, types.StringType, perms)
		resp.Diagnostics.Append(d...)
		state.Permissions = permsVal
	} else {
		state.Permissions = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PrivilegeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state privilegeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update description if changed
	if !plan.Description.Equal(state.Description) {
		opts := map[string]interface{}{}
		if plan.Description.IsNull() {
			opts["description"] = ""
		} else {
			opts["description"] = plan.Description.ValueString()
		}
		err := r.client.Call(ctx, "privilege_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA privilege description", err.Error())
			return
		}
	}

	// Permissions update
	var planPerms, statePerms []string
	if !plan.Permissions.IsNull() && !plan.Permissions.IsUnknown() {
		plan.Permissions.ElementsAs(ctx, &planPerms, false)
	}
	if !state.Permissions.IsNull() && !state.Permissions.IsUnknown() {
		state.Permissions.ElementsAs(ctx, &statePerms, false)
	}

	permsAdded := difference(planPerms, statePerms)
	permsRemoved := difference(statePerms, planPerms)

	for _, p := range permsAdded {
		err := r.client.Call(ctx, "privilege_add_permission", []string{plan.ID.ValueString()}, map[string]interface{}{
			"permission": p,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add permission to FreeIPA privilege", err.Error())
			return
		}
	}

	for _, p := range permsRemoved {
		err := r.client.Call(ctx, "privilege_remove_permission", []string{plan.ID.ValueString()}, map[string]interface{}{
			"permission": p,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove permission from FreeIPA privilege", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PrivilegeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state privilegeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "privilege_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA privilege", err.Error())
		return
	}
}

func (r *PrivilegeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
