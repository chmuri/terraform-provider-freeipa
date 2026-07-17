package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SudoCommandDataSource struct {
	client *client.Client
}

type sudoCommandDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Command     types.String `tfsdk:"command"`
	Description types.String `tfsdk:"description"`
}

func NewSudoCommandDataSource() datasource.DataSource {
	return &SudoCommandDataSource{}
}

func (d *SudoCommandDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sudo_command"
}

func (d *SudoCommandDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA Sudo Command.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique command identifier.",
			},
			"command": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The absolute path of the sudo command.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the sudo command.",
			},
		},
	}
}

func (d *SudoCommandDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SudoCommandDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sudoCommandDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Result struct {
			SudoCmd     interface{} `json:"sudocmd"`
			Description interface{} `json:"description"`
		} `json:"result"`
	}

	err := d.client.Call(ctx, "sudocmd_show", []string{state.Command.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
