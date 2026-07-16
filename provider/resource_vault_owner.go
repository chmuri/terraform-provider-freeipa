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

type VaultOwnerResource struct {
    client *client.Client
}

type vaultOwnerResourceModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    OwnerUsers  types.Set    `tfsdk:"owner_users"`
    OwnerGroups types.Set    `tfsdk:"owner_groups"`
    OwnerServices types.Set  `tfsdk:"owner_services"`
}

func NewVaultOwnerResource() resource.Resource {
    return &VaultOwnerResource{}
}

func (r *VaultOwnerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_vault_owner"
}

func (r *VaultOwnerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages owners of a FreeIPA vault. This resource does **not** create the vault itself; it only adds or removes owners.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Internal identifier (vault name).",
                PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Name of the vault to which owners are attached.",
                PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
            },
            "owner_users": schema.SetAttribute{
                ElementType:         types.StringType,
                Optional:            true,
                MarkdownDescription: "Set of usernames that are owners of the vault.",
            },
            "owner_groups": schema.SetAttribute{
                ElementType:         types.StringType,
                Optional:            true,
                MarkdownDescription: "Set of group names that are owners of the vault.",
            },
            "owner_services": schema.SetAttribute{
                ElementType:         types.StringType,
                Optional:            true,
                MarkdownDescription: "Set of service principals that are owners of the vault.",
            },
        },
    }
}

func (r *VaultOwnerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VaultOwnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan vaultOwnerResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    ownerOpts := map[string]interface{}{}
    var u, g, s []string
    if !plan.OwnerUsers.IsNull() && !plan.OwnerUsers.IsUnknown() {
        plan.OwnerUsers.ElementsAs(ctx, &u, false)
        if len(u) > 0 { ownerOpts["users"] = u }
    }
    if !plan.OwnerGroups.IsNull() && !plan.OwnerGroups.IsUnknown() {
        plan.OwnerGroups.ElementsAs(ctx, &g, false)
        if len(g) > 0 { ownerOpts["groups"] = g }
    }
    if !plan.OwnerServices.IsNull() && !plan.OwnerServices.IsUnknown() {
        plan.OwnerServices.ElementsAs(ctx, &s, false)
        if len(s) > 0 { ownerOpts["services"] = s }
    }
    if len(ownerOpts) == 0 {
        resp.Diagnostics.AddError("No owners supplied", "At least one owner (user, group or service) must be defined.")
        return
    }
    err := r.client.Call(ctx, "vault_add_owner", []string{plan.Name.ValueString()}, ownerOpts, nil)
    if err != nil {
        resp.Diagnostics.AddError("Failed to add owners to FreeIPA vault", err.Error())
        return
    }
    plan.ID = plan.Name
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *VaultOwnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state vaultOwnerResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    var result struct {
        Result struct {
            OwnerUser    interface{} `json:"owner_user"`
            OwnerGroup   interface{} `json:"owner_group"`
            OwnerService interface{} `json:"owner_service"`
        } `json:"result"`
    }
    err := r.client.Call(ctx, "vault_show", []string{state.ID.ValueString()}, map[string]interface{}{ "all": true }, &result)
    if err != nil {
        if isNotFoundError(err) { resp.State.RemoveResource(ctx); return }
        resp.Diagnostics.AddError("Failed to read vault owners", err.Error())
        return
    }
    // Populate sets
    ownersU := parseStringSlice(result.Result.OwnerUser)
    if len(ownersU) > 0 {
        val, d := types.SetValueFrom(ctx, types.StringType, ownersU)
        resp.Diagnostics.Append(d...)
        state.OwnerUsers = val
    } else {
        state.OwnerUsers = types.SetNull(types.StringType)
    }
    ownersG := parseStringSlice(result.Result.OwnerGroup)
    if len(ownersG) > 0 {
        val, d := types.SetValueFrom(ctx, types.StringType, ownersG)
        resp.Diagnostics.Append(d...)
        state.OwnerGroups = val
    } else {
        state.OwnerGroups = types.SetNull(types.StringType)
    }
    ownersS := parseStringSlice(result.Result.OwnerService)
    if len(ownersS) > 0 {
        val, d := types.SetValueFrom(ctx, types.StringType, ownersS)
        resp.Diagnostics.Append(d...)
        state.OwnerServices = val
    } else {
        state.OwnerServices = types.SetNull(types.StringType)
    }
    state.ID = state.Name
    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
}

func (r *VaultOwnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state vaultOwnerResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() { return }

    // Compute diffs for each attribute
    type diffInfo struct { Plan, State types.Set; Param string }
    diffs := []diffInfo{{plan.OwnerUsers, state.OwnerUsers, "users"}, {plan.OwnerGroups, state.OwnerGroups, "groups"}, {plan.OwnerServices, state.OwnerServices, "services"}}
    for _, d := range diffs {
        var planVals, stateVals []string
        if !d.Plan.IsNull() && !d.Plan.IsUnknown() { d.Plan.ElementsAs(ctx, &planVals, false) }
        if !d.State.IsNull() && !d.State.IsUnknown() { d.State.ElementsAs(ctx, &stateVals, false) }
        added := difference(planVals, stateVals)
        removed := difference(stateVals, planVals)
        if len(added) > 0 {
            err := r.client.Call(ctx, "vault_add_owner", []string{plan.Name.ValueString()}, map[string]interface{}{ d.Param: added }, nil)
            if err != nil { resp.Diagnostics.AddError("Failed to add owners", err.Error()); return }
        }
        if len(removed) > 0 {
            err := r.client.Call(ctx, "vault_remove_owner", []string{plan.Name.ValueString()}, map[string]interface{}{ d.Param: removed }, nil)
            if err != nil { resp.Diagnostics.AddError("Failed to remove owners", err.Error()); return }
        }
    }
    // Preserve immutable fields
    plan.ID = plan.Name
    diags := resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *VaultOwnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state vaultOwnerResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }
    // Remove all owners – we rely on current state to know what to remove
    var users, groups, services []string
    if !state.OwnerUsers.IsNull() && !state.OwnerUsers.IsUnknown() { state.OwnerUsers.ElementsAs(ctx, &users, false) }
    if !state.OwnerGroups.IsNull() && !state.OwnerGroups.IsUnknown() { state.OwnerGroups.ElementsAs(ctx, &groups, false) }
    if !state.OwnerServices.IsNull() && !state.OwnerServices.IsUnknown() { state.OwnerServices.ElementsAs(ctx, &services, false) }
    opts := map[string]interface{}{}
    if len(users) > 0 { opts["users"] = users }
    if len(groups) > 0 { opts["groups"] = groups }
    if len(services) > 0 { opts["services"] = services }
    if len(opts) > 0 {
        err := r.client.Call(ctx, "vault_remove_owner", []string{state.Name.ValueString()}, opts, nil)
        if err != nil && !isNotFoundError(err) {
            resp.Diagnostics.AddError("Failed to remove owners during delete", err.Error())
        }
    }
}

func (r *VaultOwnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
