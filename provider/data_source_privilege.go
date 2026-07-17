package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivilegeDataSource struct {
	client *client.Client
}

type privilegeDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Set    `tfsdk:"permissions"`
}

func NewPrivilegeDataSource() datasource.DataSource {
	return &PrivilegeDataSource{}
}

func (d *PrivilegeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privilege"
}

func (d *PrivilegeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Privilege.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique name of the privilege.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the privilege (cn).",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the privilege.",
			},
			"permissions": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of permission names associated with this privilege.",
			},
		},
	}
}

func (d *PrivilegeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PrivilegeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state privilegeDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAPrivilegeResult
	err := d.client.Call(ctx, "privilege_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA privilege", err.Error())
		return
	}

	res := result.Result
	state.ID = state.Name
	state.Name = types.StringValue(parseStringVal(res.Cn))

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
