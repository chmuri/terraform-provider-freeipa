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

type HbacSvcGroupResource struct {
	client *client.Client
}

type hbacSvcGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Services    types.Set    `tfsdk:"services"`
}

func NewHbacSvcGroupResource() resource.Resource {
	return &HbacSvcGroupResource{}
}

func (r *HbacSvcGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hbac_service_group"
}

func (r *HbacSvcGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA HBAC Service Groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the HBAC service group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the HBAC service group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the HBAC service group.",
			},
			"services": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Set of HBAC service names associated with this group.",
			},
		},
	}
}

func (r *HbacSvcGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type FreeIPAHbacSvcGroupResult struct {
	Result struct {
		Cn                []string `json:"cn"`
		Description       []string `json:"description"`
		MemberHbacService []string `json:"member_hbacservice"`
	} `json:"result"`
}

func (r *HbacSvcGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hbacSvcGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		opts["description"] = plan.Description.ValueString()
	}

	err := r.client.Call(ctx, "hbacsvcgroup_add", []string{plan.Name.ValueString()}, opts, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA HBAC service group", err.Error())
		return
	}

	// Add service members
	if !plan.Services.IsNull() && !plan.Services.IsUnknown() {
		var services []string
		plan.Services.ElementsAs(ctx, &services, false)
		if len(services) > 0 {
			err = r.client.Call(ctx, "hbacsvcgroup_add_member", []string{plan.Name.ValueString()}, map[string]interface{}{
				"hbacsvc": services,
			}, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to add services to HBAC service group", err.Error())
				return
			}
		}
	}

	plan.ID = plan.Name

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HbacSvcGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hbacSvcGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAHbacSvcGroupResult
	err := r.client.Call(ctx, "hbacsvcgroup_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FreeIPA HBAC service group", err.Error())
		return
	}

	res := result.Result
	if len(res.Cn) > 0 {
		state.Name = types.StringValue(res.Cn[0])
		state.ID = types.StringValue(res.Cn[0])
	}
	if len(res.Description) > 0 {
		state.Description = types.StringValue(res.Description[0])
	} else {
		state.Description = types.StringNull()
	}

	if len(res.MemberHbacService) > 0 {
		svcVal, d := types.SetValueFrom(ctx, types.StringType, res.MemberHbacService)
		resp.Diagnostics.Append(d...)
		state.Services = svcVal
	} else {
		state.Services = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *HbacSvcGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state hbacSvcGroupResourceModel
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
		err := r.client.Call(ctx, "hbacsvcgroup_mod", []string{plan.ID.ValueString()}, opts, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update FreeIPA HBAC service group description", err.Error())
			return
		}
	}

	// Update services membership
	var planServices, stateServices []string
	if !plan.Services.IsNull() && !plan.Services.IsUnknown() {
		plan.Services.ElementsAs(ctx, &planServices, false)
	}
	if !state.Services.IsNull() && !state.Services.IsUnknown() {
		state.Services.ElementsAs(ctx, &stateServices, false)
	}

	added := difference(planServices, stateServices)
	removed := difference(stateServices, planServices)

	if len(added) > 0 {
		err := r.client.Call(ctx, "hbacsvcgroup_add_member", []string{plan.ID.ValueString()}, map[string]interface{}{
			"hbacsvc": added,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to add services to FreeIPA HBAC service group", err.Error())
			return
		}
	}

	if len(removed) > 0 {
		err := r.client.Call(ctx, "hbacsvcgroup_remove_member", []string{plan.ID.ValueString()}, map[string]interface{}{
			"hbacsvc": removed,
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to remove services from FreeIPA HBAC service group", err.Error())
			return
		}
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HbacSvcGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hbacSvcGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Call(ctx, "hbacsvcgroup_del", []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA HBAC service group", err.Error())
		return
	}
}

func (r *HbacSvcGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
