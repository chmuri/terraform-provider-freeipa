package provider

import (
	"context"

	"github.com/beeripa/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	FirstName types.String `tfsdk:"first_name"`
	LastName  types.String `tfsdk:"last_name"`
	Email     types.String `tfsdk:"email"`
	FullName  types.String `tfsdk:"full_name"`
	Cn        types.String `tfsdk:"cn"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads details of a FreeIPA user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique username identifying the user.",
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The username (login).",
			},
			"first_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "First name.",
			},
			"last_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last name.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Email address.",
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Full name.",
			},
			"cn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Common name.",
			},
			"enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the user account is enabled.",
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAUserResult
	err := d.client.Call(ctx, "user_show", []string{state.Username.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read FreeIPA user", err.Error())
		return
	}

	res := result.Result
	state.ID = state.Username
	state.FirstName = types.StringValue(parseStringVal(res.Givenname))
	state.LastName = types.StringValue(parseStringVal(res.Sn))
	state.Email = types.StringValue(parseStringVal(res.Mail))
	state.Cn = types.StringValue(parseStringVal(res.Cn))
	state.FullName = types.StringValue(parseStringVal(res.Displayname))
	// nsaccountlock=true means disabled, so enabled = !nsaccountlock
	enabled := true
	if res.Nsaccountlock != nil {
		enabled = !*res.Nsaccountlock
	}
	state.Enabled = types.BoolValue(enabled)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
