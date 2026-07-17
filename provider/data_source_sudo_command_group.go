package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SudoCommandGroupDataSource struct {
	client *client.Client
}

type sudoCommandGroupDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Commands    types.Set    `tfsdk:"commands"`
}

func NewSudoCommandGroupDataSource() datasource.DataSource {
	return &SudoCommandGroupDataSource{}
}

func (d *SudoCommandGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sudo_command_group"
}

func (d *SudoCommandGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Sudo Command Group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier (group name).",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the sudo command group.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the sudo command group.",
			},
			"commands": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Set of sudo command paths that belong to this group.",
			},
		},
	}
}

func (d *SudoCommandGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SudoCommandGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sudoCommandGroupDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
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

	err := d.client.Call(ctx, "sudocmdgroup_show", []string{state.Name.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA sudo command group", err.Error())
		return
	}

	state.ID = state.Name
	state.Name = types.StringValue(parseStringVal(result.Result.CN))

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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
