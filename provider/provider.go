package provider

import (
	"context"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type freeipaProvider struct {
	version string
}

type freeipaProviderModel struct {
	Host       types.String `tfsdk:"host"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	Insecure   types.Bool   `tfsdk:"insecure"`
	KeytabPath types.String `tfsdk:"keytab_path"`
	Realm      types.String `tfsdk:"realm"`
	Krb5Conf   types.String `tfsdk:"krb5_conf"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &freeipaProvider{
			version: version,
		}
	}
}

func (p *freeipaProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "freeipa"
	resp.Version = p.version
}

func (p *freeipaProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "FreeIPA host address (e.g. ipa.example.com).",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "FreeIPA administrator username.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "FreeIPA administrator password.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Bypass SSL verification.",
				Optional:            true,
			},
			"keytab_path": schema.StringAttribute{
				MarkdownDescription: "Path to Kerberos Keytab file.",
				Optional:            true,
			},
			"realm": schema.StringAttribute{
				MarkdownDescription: "Kerberos Realm.",
				Optional:            true,
			},
			"krb5_conf": schema.StringAttribute{
				MarkdownDescription: "Path to Kerberos configuration file krb5.conf.",
				Optional:            true,
			},
		},
	}
}

func (p *freeipaProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data freeipaProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	authMethod := client.AuthPassword
	if !data.KeytabPath.IsNull() && !data.KeytabPath.IsUnknown() {
		authMethod = client.AuthKerberosKeytab
	}

	cfg := &client.Config{
		Host:       data.Host.ValueString(),
		Insecure:   data.Insecure.ValueBool(),
		AuthMethod: authMethod,
		Username:   data.Username.ValueString(),
		Password:   data.Password.ValueString(),
		KeytabPath: data.KeytabPath.ValueString(),
		Realm:      data.Realm.ValueString(),
		Krb5Conf:   data.Krb5Conf.ValueString(),
	}

	c, err := client.NewClient(cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create FreeIPA client",
			"An unexpected error occurred when creating the FreeIPA client: "+err.Error(),
		)
		return
	}

	// Perform login to check connection (if hosts and credentials are provided)
	if cfg.Host != "" && (cfg.Password != "" || cfg.KeytabPath != "") {
		err = c.Login()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to authenticate to FreeIPA",
				"An unexpected error occurred when authenticating: "+err.Error(),
			)
			return
		}
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *freeipaProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
		NewGroupResource,
		NewHostResource,
		NewHostGroupResource,
		NewHbacRuleResource,
		NewHbacSvcResource,
		NewHbacSvcGroupResource,
		NewSudoRuleResource,
		NewSudoCommandResource,
		NewSudoCommandGroupResource,
		NewDnsZoneResource,
		NewDnsRecordResource,
		NewRoleResource,
		NewPrivilegeResource,
		NewNetgroupResource,
		NewPwPolicyResource,
		NewVaultResource,
		NewVaultOwnerResource,
		NewVaultMemberResource,
	}
}

func (p *freeipaProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
		NewGroupDataSource,
		NewHostDataSource,
		NewHostGroupDataSource,
		NewHbacRuleDataSource,
		NewDnsZoneDataSource,
	}
}
