package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/chmuri/terraform-provider-freeipa/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserResource struct {
	client *client.Client
}

type userResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Username          types.String `tfsdk:"username"`
	FirstName         types.String `tfsdk:"first_name"`
	LastName          types.String `tfsdk:"last_name"`
	Email             types.String `tfsdk:"email"`
	Password          types.String `tfsdk:"password"`
	SSHPublicKeys     types.List   `tfsdk:"ssh_public_keys"`
	Cn                types.String `tfsdk:"cn"`
	FullName          types.String `tfsdk:"full_name"`
	Displayname       types.String `tfsdk:"displayname"`
	Initials          types.String `tfsdk:"initials"`
	Homedir           types.String `tfsdk:"homedir"`
	Gecos             types.String `tfsdk:"gecos"`
	Shell             types.String `tfsdk:"shell"`
	Uid               types.Int64  `tfsdk:"uid"`
	Gid               types.Int64  `tfsdk:"gid"`
	Street            types.String `tfsdk:"street"`
	City              types.String `tfsdk:"city"`
	State             types.String `tfsdk:"state"`
	Postalcode        types.String `tfsdk:"postalcode"`
	Phone             types.String `tfsdk:"phone"`
	Mobile            types.String `tfsdk:"mobile"`
	Pager             types.String `tfsdk:"pager"`
	Fax               types.String `tfsdk:"fax"`
	Orgunit           types.String `tfsdk:"orgunit"`
	Title             types.String `tfsdk:"title"`
	Manager           types.String `tfsdk:"manager"`
	Carlicense        types.String `tfsdk:"carlicense"`
	UserAuthType      types.List   `tfsdk:"user_auth_type"`
	Class             types.String `tfsdk:"class"`
	Departmentnumber  types.String `tfsdk:"departmentnumber"`
	Employeenumber    types.String `tfsdk:"employeenumber"`
	Employeetype      types.String `tfsdk:"employeetype"`
	Preferredlanguage types.String `tfsdk:"preferredlanguage"`
	Certificate       types.String `tfsdk:"certificate"`
	RandomPassword    types.Bool   `tfsdk:"random_password"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	Staged            types.Bool   `tfsdk:"staged"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages FreeIPA users.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique username identifying the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The username (login).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"first_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "First name.",
			},
			"last_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Last name.",
			},
			"email": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Email address.",
			},
			"password": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Initial password for the user.",
			},
			"ssh_public_keys": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "List of SSH public keys.",
			},
			"cn": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Common name.",
			},
			"full_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Full name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"displayname": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Display name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"initials": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Initials.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"homedir": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Home directory.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gecos": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "GECOS.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"shell": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Login shell.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uid": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "User ID Number.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"gid": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Group ID Number.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"street": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Street address.",
			},
			"city": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "City.",
			},
			"state": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "State/Province.",
			},
			"postalcode": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ZIP/Postal code.",
			},
			"phone": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Telephone Number.",
			},
			"mobile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Mobile Telephone Number.",
			},
			"pager": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Pager Number.",
			},
			"fax": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Fax Number.",
			},
			"orgunit": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Org. Unit.",
			},
			"title": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Job Title.",
			},
			"manager": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Manager.",
			},
			"carlicense": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Car License.",
			},
			"user_auth_type": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Types of supported user authentication.",
			},
			"class": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User category.",
			},
			"departmentnumber": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Department Number.",
			},
			"employeenumber": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Employee Number.",
			},
			"employeetype": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Employee Type.",
			},
			"preferredlanguage": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Preferred Language.",
			},
			"certificate": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Base-64 encoded user certificate.",
			},
			"random_password": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Generate a random user password.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the user account is enabled.",
			},
			"staged": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the user is in staging area.",
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if rpcErr, ok := err.(*client.RPCError); ok {
		return rpcErr.Code == 4001 // NotFound
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

type FreeIPAUserResult struct {
	Result struct {
		Uid                      []string `json:"uid"`
		Givenname                []string `json:"givenname"`
		Sn                       []string `json:"sn"`
		Mail                     []string `json:"mail"`
		SSHPublicKeys            []string `json:"ipasshpubkey"`
		Cn                       []string `json:"cn"`
		Displayname              []string `json:"displayname"`
		Initials                 []string `json:"initials"`
		Homedirectory            []string `json:"homedirectory"`
		Gecos                    []string `json:"gecos"`
		Loginshell               []string `json:"loginshell"`
		Uidnumber                []string `json:"uidnumber"`
		Gidnumber                []string `json:"gidnumber"`
		Street                   []string `json:"street"`
		L                        []string `json:"l"`
		St                       []string `json:"st"`
		Postalcode               []string `json:"postalcode"`
		Telephonenumber          []string `json:"telephonenumber"`
		Mobile                   []string `json:"mobile"`
		Pager                    []string `json:"pager"`
		Facsimiletelephonenumber []string `json:"facsimiletelephonenumber"`
		Ou                       []string `json:"ou"`
		Title                    []string `json:"title"`
		Manager                  []string `json:"manager"`
		Carlicense               []string `json:"carlicense"`
		Ipauserauthtype          []string `json:"ipauserauthtype"`
		Userclass                []string `json:"userclass"`
		Departmentnumber         []string `json:"departmentnumber"`
		Employeenumber           []string `json:"employeenumber"`
		Employeetype             []string `json:"employeetype"`
		Preferredlanguage        []string `json:"preferredlanguage"`
		Usercertificate          []string `json:"usercertificate"`
		Nsaccountlock            *bool    `json:"nsaccountlock"`
		Randompassword           string   `json:"randompassword"`
	} `json:"result"`
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := map[string]interface{}{
		"givenname": plan.FirstName.ValueString(),
		"sn":        plan.LastName.ValueString(),
	}

	if !plan.Email.IsNull() && !plan.Email.IsUnknown() {
		opts["mail"] = plan.Email.ValueString()
	}

	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		opts["userpassword"] = plan.Password.ValueString()
	}

	if !plan.SSHPublicKeys.IsNull() && !plan.SSHPublicKeys.IsUnknown() {
		var sshKeys []string
		plan.SSHPublicKeys.ElementsAs(ctx, &sshKeys, false)
		opts["ipasshpubkey"] = sshKeys
	}

	if !plan.Cn.IsNull() && !plan.Cn.IsUnknown() {
		opts["cn"] = plan.Cn.ValueString()
	} else if !plan.FullName.IsNull() && !plan.FullName.IsUnknown() {
		opts["cn"] = plan.FullName.ValueString()
	}

	if !plan.Displayname.IsNull() && !plan.Displayname.IsUnknown() {
		opts["displayname"] = plan.Displayname.ValueString()
	}
	if !plan.Initials.IsNull() && !plan.Initials.IsUnknown() {
		opts["initials"] = plan.Initials.ValueString()
	}
	if !plan.Homedir.IsNull() && !plan.Homedir.IsUnknown() {
		opts["homedirectory"] = plan.Homedir.ValueString()
	}
	if !plan.Gecos.IsNull() && !plan.Gecos.IsUnknown() {
		opts["gecos"] = plan.Gecos.ValueString()
	}
	if !plan.Shell.IsNull() && !plan.Shell.IsUnknown() {
		opts["loginshell"] = plan.Shell.ValueString()
	}
	if !plan.Uid.IsNull() && !plan.Uid.IsUnknown() {
		opts["uidnumber"] = plan.Uid.ValueInt64()
	}
	if !plan.Gid.IsNull() && !plan.Gid.IsUnknown() {
		opts["gidnumber"] = plan.Gid.ValueInt64()
	}
	if !plan.Street.IsNull() && !plan.Street.IsUnknown() {
		opts["street"] = plan.Street.ValueString()
	}
	if !plan.City.IsNull() && !plan.City.IsUnknown() {
		opts["l"] = plan.City.ValueString()
	}
	if !plan.State.IsNull() && !plan.State.IsUnknown() {
		opts["st"] = plan.State.ValueString()
	}
	if !plan.Postalcode.IsNull() && !plan.Postalcode.IsUnknown() {
		opts["postalcode"] = plan.Postalcode.ValueString()
	}
	if !plan.Phone.IsNull() && !plan.Phone.IsUnknown() {
		opts["telephonenumber"] = plan.Phone.ValueString()
	}
	if !plan.Mobile.IsNull() && !plan.Mobile.IsUnknown() {
		opts["mobile"] = plan.Mobile.ValueString()
	}
	if !plan.Pager.IsNull() && !plan.Pager.IsUnknown() {
		opts["pager"] = plan.Pager.ValueString()
	}
	if !plan.Fax.IsNull() && !plan.Fax.IsUnknown() {
		opts["facsimiletelephonenumber"] = plan.Fax.ValueString()
	}
	if !plan.Orgunit.IsNull() && !plan.Orgunit.IsUnknown() {
		opts["ou"] = plan.Orgunit.ValueString()
	}
	if !plan.Title.IsNull() && !plan.Title.IsUnknown() {
		opts["title"] = plan.Title.ValueString()
	}
	if !plan.Manager.IsNull() && !plan.Manager.IsUnknown() {
		opts["manager"] = plan.Manager.ValueString()
	}
	if !plan.Carlicense.IsNull() && !plan.Carlicense.IsUnknown() {
		opts["carlicense"] = plan.Carlicense.ValueString()
	}
	if !plan.UserAuthType.IsNull() && !plan.UserAuthType.IsUnknown() {
		var authTypes []string
		plan.UserAuthType.ElementsAs(ctx, &authTypes, false)
		opts["ipauserauthtype"] = authTypes
	}
	if !plan.Class.IsNull() && !plan.Class.IsUnknown() {
		opts["userclass"] = plan.Class.ValueString()
	}
	if !plan.Departmentnumber.IsNull() && !plan.Departmentnumber.IsUnknown() {
		opts["departmentnumber"] = plan.Departmentnumber.ValueString()
	}
	if !plan.Employeenumber.IsNull() && !plan.Employeenumber.IsUnknown() {
		opts["employeenumber"] = plan.Employeenumber.ValueString()
	}
	if !plan.Employeetype.IsNull() && !plan.Employeetype.IsUnknown() {
		opts["employeetype"] = plan.Employeetype.ValueString()
	}
	if !plan.Preferredlanguage.IsNull() && !plan.Preferredlanguage.IsUnknown() {
		opts["preferredlanguage"] = plan.Preferredlanguage.ValueString()
	}
	if !plan.Certificate.IsNull() && !plan.Certificate.IsUnknown() {
		opts["usercertificate"] = plan.Certificate.ValueString()
	}
	if !plan.RandomPassword.IsNull() && !plan.RandomPassword.IsUnknown() && plan.RandomPassword.ValueBool() {
		opts["random"] = true
	}

	method := "user_add"
	isStaged := false
	if !plan.Staged.IsNull() && !plan.Staged.IsUnknown() && plan.Staged.ValueBool() {
		method = "stageuser_add"
		isStaged = true
	}

	var result FreeIPAUserResult
	err := r.client.Call(ctx, method, []string{plan.Username.ValueString()}, opts, &result)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FreeIPA user", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Username.ValueString())

	if isStaged {
		plan.Staged = types.BoolValue(true)
	} else {
		plan.Staged = types.BoolValue(false)
		// Handle enabled/disabled for active users
		if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() && !plan.Enabled.ValueBool() {
			err = r.client.Call(ctx, "user_disable", []string{plan.Username.ValueString()}, nil, nil)
			if err != nil {
				resp.Diagnostics.AddError("Failed to disable FreeIPA user", err.Error())
				return
			}
			plan.Enabled = types.BoolValue(false)
		} else {
			plan.Enabled = types.BoolValue(true)
		}
	}

	if !plan.RandomPassword.IsNull() && !plan.RandomPassword.IsUnknown() && plan.RandomPassword.ValueBool() {
		if result.Result.Randompassword != "" {
			plan.Password = types.StringValue(result.Result.Randompassword)
		}
	}

	if plan.Cn.IsNull() || plan.Cn.IsUnknown() {
		if len(result.Result.Cn) > 0 {
			plan.Cn = types.StringValue(result.Result.Cn[0])
		} else {
			plan.Cn = types.StringNull()
		}
	}
	if plan.Email.IsNull() || plan.Email.IsUnknown() {
		if len(result.Result.Mail) > 0 {
			plan.Email = types.StringValue(result.Result.Mail[0])
		} else {
			plan.Email = types.StringNull()
		}
	}
	if plan.FullName.IsNull() || plan.FullName.IsUnknown() {
		plan.FullName = plan.Cn
	}
	if plan.Displayname.IsNull() || plan.Displayname.IsUnknown() {
		if len(result.Result.Displayname) > 0 {
			plan.Displayname = types.StringValue(result.Result.Displayname[0])
		} else {
			plan.Displayname = types.StringNull()
		}
	}
	if plan.Initials.IsNull() || plan.Initials.IsUnknown() {
		if len(result.Result.Initials) > 0 {
			plan.Initials = types.StringValue(result.Result.Initials[0])
		} else {
			plan.Initials = types.StringNull()
		}
	}
	if plan.Homedir.IsNull() || plan.Homedir.IsUnknown() {
		if len(result.Result.Homedirectory) > 0 {
			plan.Homedir = types.StringValue(result.Result.Homedirectory[0])
		} else {
			plan.Homedir = types.StringNull()
		}
	}
	if plan.Gecos.IsNull() || plan.Gecos.IsUnknown() {
		if len(result.Result.Gecos) > 0 {
			plan.Gecos = types.StringValue(result.Result.Gecos[0])
		} else {
			plan.Gecos = types.StringNull()
		}
	}
	if plan.Shell.IsNull() || plan.Shell.IsUnknown() {
		if len(result.Result.Loginshell) > 0 {
			plan.Shell = types.StringValue(result.Result.Loginshell[0])
		} else {
			plan.Shell = types.StringNull()
		}
	}
	if plan.Uid.IsNull() || plan.Uid.IsUnknown() {
		if len(result.Result.Uidnumber) > 0 {
			uidVal, err := strconv.ParseInt(result.Result.Uidnumber[0], 10, 64)
			if err == nil {
				plan.Uid = types.Int64Value(uidVal)
			} else {
				plan.Uid = types.Int64Null()
			}
		} else {
			plan.Uid = types.Int64Null()
		}
	}
	if plan.Gid.IsNull() || plan.Gid.IsUnknown() {
		if len(result.Result.Gidnumber) > 0 {
			gidVal, err := strconv.ParseInt(result.Result.Gidnumber[0], 10, 64)
			if err == nil {
				plan.Gid = types.Int64Value(gidVal)
			} else {
				plan.Gid = types.Int64Null()
			}
		} else {
			plan.Gid = types.Int64Null()
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result FreeIPAUserResult
	showMethod := "user_show"
	err := r.client.Call(ctx, "user_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
	if err != nil {
		if isNotFoundError(err) {
			err = r.client.Call(ctx, "stageuser_show", []string{state.ID.ValueString()}, map[string]interface{}{"all": true}, &result)
			if err != nil {
				if isNotFoundError(err) {
					resp.State.RemoveResource(ctx)
					return
				}
				resp.Diagnostics.AddError("Failed to read stage FreeIPA user", err.Error())
				return
			}
			showMethod = "stageuser_show"
		} else {
			resp.Diagnostics.AddError("Failed to read FreeIPA user", err.Error())
			return
		}
	}

	res := result.Result
	if len(res.Uid) > 0 {
		state.Username = types.StringValue(res.Uid[0])
		state.ID = types.StringValue(res.Uid[0])
	}
	if len(res.Givenname) > 0 {
		state.FirstName = types.StringValue(res.Givenname[0])
	}
	if len(res.Sn) > 0 {
		state.LastName = types.StringValue(res.Sn[0])
	}
	if len(res.Mail) > 0 {
		state.Email = types.StringValue(res.Mail[0])
	} else {
		state.Email = types.StringNull()
	}

	if len(res.SSHPublicKeys) > 0 {
		sshKeysVal, d := types.ListValueFrom(ctx, types.StringType, res.SSHPublicKeys)
		resp.Diagnostics.Append(d...)
		state.SSHPublicKeys = sshKeysVal
	} else {
		state.SSHPublicKeys = types.ListNull(types.StringType)
	}

	if len(res.Cn) > 0 {
		state.Cn = types.StringValue(res.Cn[0])
		state.FullName = types.StringValue(res.Cn[0])
	} else {
		state.Cn = types.StringNull()
		state.FullName = types.StringNull()
	}

	if len(res.Displayname) > 0 {
		state.Displayname = types.StringValue(res.Displayname[0])
	} else {
		state.Displayname = types.StringNull()
	}

	if len(res.Initials) > 0 {
		state.Initials = types.StringValue(res.Initials[0])
	} else {
		state.Initials = types.StringNull()
	}

	if len(res.Homedirectory) > 0 {
		state.Homedir = types.StringValue(res.Homedirectory[0])
	} else {
		state.Homedir = types.StringNull()
	}

	if len(res.Gecos) > 0 {
		state.Gecos = types.StringValue(res.Gecos[0])
	} else {
		state.Gecos = types.StringNull()
	}

	if len(res.Loginshell) > 0 {
		state.Shell = types.StringValue(res.Loginshell[0])
	} else {
		state.Shell = types.StringNull()
	}

	if len(res.Uidnumber) > 0 {
		val, err := strconv.ParseInt(res.Uidnumber[0], 10, 64)
		if err == nil {
			state.Uid = types.Int64Value(val)
		}
	} else {
		state.Uid = types.Int64Null()
	}

	if len(res.Gidnumber) > 0 {
		val, err := strconv.ParseInt(res.Gidnumber[0], 10, 64)
		if err == nil {
			state.Gid = types.Int64Value(val)
		}
	} else {
		state.Gid = types.Int64Null()
	}

	if len(res.Street) > 0 {
		state.Street = types.StringValue(res.Street[0])
	} else {
		state.Street = types.StringNull()
	}

	if len(res.L) > 0 {
		state.City = types.StringValue(res.L[0])
	} else {
		state.City = types.StringNull()
	}

	if len(res.St) > 0 {
		state.State = types.StringValue(res.St[0])
	} else {
		state.State = types.StringNull()
	}

	if len(res.Postalcode) > 0 {
		state.Postalcode = types.StringValue(res.Postalcode[0])
	} else {
		state.Postalcode = types.StringNull()
	}

	if len(res.Telephonenumber) > 0 {
		state.Phone = types.StringValue(res.Telephonenumber[0])
	} else {
		state.Phone = types.StringNull()
	}

	if len(res.Mobile) > 0 {
		state.Mobile = types.StringValue(res.Mobile[0])
	} else {
		state.Mobile = types.StringNull()
	}

	if len(res.Pager) > 0 {
		state.Pager = types.StringValue(res.Pager[0])
	} else {
		state.Pager = types.StringNull()
	}

	if len(res.Facsimiletelephonenumber) > 0 {
		state.Fax = types.StringValue(res.Facsimiletelephonenumber[0])
	} else {
		state.Fax = types.StringNull()
	}

	if len(res.Ou) > 0 {
		state.Orgunit = types.StringValue(res.Ou[0])
	} else {
		state.Orgunit = types.StringNull()
	}

	if len(res.Title) > 0 {
		state.Title = types.StringValue(res.Title[0])
	} else {
		state.Title = types.StringNull()
	}

	if len(res.Manager) > 0 {
		state.Manager = types.StringValue(res.Manager[0])
	} else {
		state.Manager = types.StringNull()
	}

	if len(res.Carlicense) > 0 {
		state.Carlicense = types.StringValue(res.Carlicense[0])
	} else {
		state.Carlicense = types.StringNull()
	}

	if len(res.Ipauserauthtype) > 0 {
		authTypesVal, d := types.ListValueFrom(ctx, types.StringType, res.Ipauserauthtype)
		resp.Diagnostics.Append(d...)
		state.UserAuthType = authTypesVal
	} else {
		state.UserAuthType = types.ListNull(types.StringType)
	}

	if len(res.Userclass) > 0 {
		state.Class = types.StringValue(res.Userclass[0])
	} else {
		state.Class = types.StringNull()
	}

	if len(res.Departmentnumber) > 0 {
		state.Departmentnumber = types.StringValue(res.Departmentnumber[0])
	} else {
		state.Departmentnumber = types.StringNull()
	}

	if len(res.Employeenumber) > 0 {
		state.Employeenumber = types.StringValue(res.Employeenumber[0])
	} else {
		state.Employeenumber = types.StringNull()
	}

	if len(res.Employeetype) > 0 {
		state.Employeetype = types.StringValue(res.Employeetype[0])
	} else {
		state.Employeetype = types.StringNull()
	}

	if len(res.Preferredlanguage) > 0 {
		state.Preferredlanguage = types.StringValue(res.Preferredlanguage[0])
	} else {
		state.Preferredlanguage = types.StringNull()
	}

	if len(res.Usercertificate) > 0 {
		state.Certificate = types.StringValue(res.Usercertificate[0])
	} else {
		state.Certificate = types.StringNull()
	}

	if showMethod == "stageuser_show" {
		state.Staged = types.BoolValue(true)
		state.Enabled = types.BoolNull()
	} else {
		state.Staged = types.BoolValue(false)
		if res.Nsaccountlock != nil {
			state.Enabled = types.BoolValue(!*res.Nsaccountlock)
		} else {
			state.Enabled = types.BoolValue(true)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := "user_mod"
	if state.Staged.ValueBool() {
		method = "stageuser_mod"
	}

	// Transition from staged to active
	if state.Staged.ValueBool() && !plan.Staged.ValueBool() {
		err := r.client.Call(ctx, "stageuser_activate", []string{plan.Username.ValueString()}, nil, nil)
		if err != nil {
			resp.Diagnostics.AddError("Failed to activate FreeIPA user", err.Error())
			return
		}
		method = "user_mod"
		state.Staged = types.BoolValue(false)
		state.Enabled = types.BoolValue(true) // FreeIPA defaults active users to enabled
	}

	opts := map[string]interface{}{}

	if !plan.Cn.IsNull() && !plan.Cn.IsUnknown() {
		opts["cn"] = plan.Cn.ValueString()
	} else if !state.Cn.IsNull() && !state.Cn.IsUnknown() {
		opts["cn"] = state.Cn.ValueString()
	}

	if !plan.FirstName.Equal(state.FirstName) {
		opts["givenname"] = plan.FirstName.ValueString()
	}
	if !plan.LastName.Equal(state.LastName) {
		opts["sn"] = plan.LastName.ValueString()
	}
	if !plan.Email.Equal(state.Email) {
		if plan.Email.IsNull() {
			opts["mail"] = ""
		} else {
			opts["mail"] = plan.Email.ValueString()
		}
	}
	if !plan.SSHPublicKeys.Equal(state.SSHPublicKeys) {
		if plan.SSHPublicKeys.IsNull() {
			opts["ipasshpubkey"] = []string{}
		} else {
			var sshKeys []string
			plan.SSHPublicKeys.ElementsAs(ctx, &sshKeys, false)
			opts["ipasshpubkey"] = sshKeys
		}
	}
	if !plan.Cn.Equal(state.Cn) && !plan.Cn.IsUnknown() {
		if plan.Cn.IsNull() {
			opts["cn"] = ""
		} else {
			opts["cn"] = plan.Cn.ValueString()
		}
	}
	if !plan.FullName.Equal(state.FullName) && !plan.FullName.IsUnknown() {
		if plan.FullName.IsNull() {
			if plan.Cn.IsNull() {
				opts["cn"] = ""
			}
		} else {
			opts["cn"] = plan.FullName.ValueString()
		}
	}
	if !plan.Displayname.Equal(state.Displayname) && !plan.Displayname.IsUnknown() {
		if plan.Displayname.IsNull() {
			opts["displayname"] = ""
		} else {
			opts["displayname"] = plan.Displayname.ValueString()
		}
	}
	if !plan.Initials.Equal(state.Initials) && !plan.Initials.IsUnknown() {
		if plan.Initials.IsNull() {
			opts["initials"] = ""
		} else {
			opts["initials"] = plan.Initials.ValueString()
		}
	}
	if !plan.Homedir.Equal(state.Homedir) && !plan.Homedir.IsUnknown() {
		if plan.Homedir.IsNull() {
			opts["homedirectory"] = ""
		} else {
			opts["homedirectory"] = plan.Homedir.ValueString()
		}
	}
	if !plan.Gecos.Equal(state.Gecos) && !plan.Gecos.IsUnknown() {
		if plan.Gecos.IsNull() {
			opts["gecos"] = ""
		} else {
			opts["gecos"] = plan.Gecos.ValueString()
		}
	}
	if !plan.Shell.Equal(state.Shell) && !plan.Shell.IsUnknown() {
		if plan.Shell.IsNull() {
			opts["loginshell"] = ""
		} else {
			opts["loginshell"] = plan.Shell.ValueString()
		}
	}
	if !plan.Uid.Equal(state.Uid) && !plan.Uid.IsUnknown() {
		if plan.Uid.IsNull() {
			opts["uidnumber"] = nil
		} else {
			opts["uidnumber"] = plan.Uid.ValueInt64()
		}
	}
	if !plan.Gid.Equal(state.Gid) && !plan.Gid.IsUnknown() {
		if plan.Gid.IsNull() {
			opts["gidnumber"] = nil
		} else {
			opts["gidnumber"] = plan.Gid.ValueInt64()
		}
	}
	if !plan.Street.Equal(state.Street) {
		if plan.Street.IsNull() {
			opts["street"] = ""
		} else {
			opts["street"] = plan.Street.ValueString()
		}
	}
	if !plan.City.Equal(state.City) {
		if plan.City.IsNull() {
			opts["l"] = ""
		} else {
			opts["l"] = plan.City.ValueString()
		}
	}
	if !plan.State.Equal(state.State) {
		if plan.State.IsNull() {
			opts["st"] = ""
		} else {
			opts["st"] = plan.State.ValueString()
		}
	}
	if !plan.Postalcode.Equal(state.Postalcode) {
		if plan.Postalcode.IsNull() {
			opts["postalcode"] = ""
		} else {
			opts["postalcode"] = plan.Postalcode.ValueString()
		}
	}
	if !plan.Phone.Equal(state.Phone) {
		if plan.Phone.IsNull() {
			opts["telephonenumber"] = ""
		} else {
			opts["telephonenumber"] = plan.Phone.ValueString()
		}
	}
	if !plan.Mobile.Equal(state.Mobile) {
		if plan.Mobile.IsNull() {
			opts["mobile"] = ""
		} else {
			opts["mobile"] = plan.Mobile.ValueString()
		}
	}
	if !plan.Pager.Equal(state.Pager) {
		if plan.Pager.IsNull() {
			opts["pager"] = ""
		} else {
			opts["pager"] = plan.Pager.ValueString()
		}
	}
	if !plan.Fax.Equal(state.Fax) {
		if plan.Fax.IsNull() {
			opts["facsimiletelephonenumber"] = ""
		} else {
			opts["facsimiletelephonenumber"] = plan.Fax.ValueString()
		}
	}
	if !plan.Orgunit.Equal(state.Orgunit) {
		if plan.Orgunit.IsNull() {
			opts["ou"] = ""
		} else {
			opts["ou"] = plan.Orgunit.ValueString()
		}
	}
	if !plan.Title.Equal(state.Title) {
		if plan.Title.IsNull() {
			opts["title"] = ""
		} else {
			opts["title"] = plan.Title.ValueString()
		}
	}
	if !plan.Manager.Equal(state.Manager) {
		if plan.Manager.IsNull() {
			opts["manager"] = ""
		} else {
			opts["manager"] = plan.Manager.ValueString()
		}
	}
	if !plan.Carlicense.Equal(state.Carlicense) {
		if plan.Carlicense.IsNull() {
			opts["carlicense"] = ""
		} else {
			opts["carlicense"] = plan.Carlicense.ValueString()
		}
	}
	if !plan.UserAuthType.Equal(state.UserAuthType) {
		if plan.UserAuthType.IsNull() {
			opts["ipauserauthtype"] = []string{}
		} else {
			var authTypes []string
			plan.UserAuthType.ElementsAs(ctx, &authTypes, false)
			opts["ipauserauthtype"] = authTypes
		}
	}
	if !plan.Class.Equal(state.Class) {
		if plan.Class.IsNull() {
			opts["userclass"] = ""
		} else {
			opts["userclass"] = plan.Class.ValueString()
		}
	}
	if !plan.Departmentnumber.Equal(state.Departmentnumber) {
		if plan.Departmentnumber.IsNull() {
			opts["departmentnumber"] = ""
		} else {
			opts["departmentnumber"] = plan.Departmentnumber.ValueString()
		}
	}
	if !plan.Employeenumber.Equal(state.Employeenumber) {
		if plan.Employeenumber.IsNull() {
			opts["employeenumber"] = ""
		} else {
			opts["employeenumber"] = plan.Employeenumber.ValueString()
		}
	}
	if !plan.Employeetype.Equal(state.Employeetype) {
		if plan.Employeetype.IsNull() {
			opts["employeetype"] = ""
		} else {
			opts["employeetype"] = plan.Employeetype.ValueString()
		}
	}
	if !plan.Preferredlanguage.Equal(state.Preferredlanguage) {
		if plan.Preferredlanguage.IsNull() {
			opts["preferredlanguage"] = ""
		} else {
			opts["preferredlanguage"] = plan.Preferredlanguage.ValueString()
		}
	}
	if !plan.Certificate.Equal(state.Certificate) {
		if plan.Certificate.IsNull() {
			opts["usercertificate"] = ""
		} else {
			opts["usercertificate"] = plan.Certificate.ValueString()
		}
	}

	if len(opts) > 0 {
		var result FreeIPAUserResult
		err := r.client.Call(ctx, method, []string{plan.ID.ValueString()}, opts, &result)
		if err != nil && !isEmptyModlistError(err) {
			resp.Diagnostics.AddError("Failed to update FreeIPA user", err.Error())
			return
		}
	}

	// Handle enabled/disabled status for active users
	if !plan.Staged.ValueBool() {
		if !plan.Enabled.Equal(state.Enabled) {
			var err error
			if plan.Enabled.ValueBool() {
				err = r.client.Call(ctx, "user_enable", []string{plan.Username.ValueString()}, nil, nil)
			} else {
				err = r.client.Call(ctx, "user_disable", []string{plan.Username.ValueString()}, nil, nil)
			}
			if err != nil {
				resp.Diagnostics.AddError("Failed to change user enable/disable state", err.Error())
				return
			}
		}
	}

	if plan.Cn.IsUnknown() {
		plan.Cn = state.Cn
	}
	if plan.FullName.IsUnknown() {
		plan.FullName = state.FullName
	}
	if plan.Displayname.IsUnknown() {
		plan.Displayname = state.Displayname
	}
	if plan.Initials.IsUnknown() {
		plan.Initials = state.Initials
	}
	if plan.Homedir.IsUnknown() {
		plan.Homedir = state.Homedir
	}
	if plan.Gecos.IsUnknown() {
		plan.Gecos = state.Gecos
	}
	if plan.Shell.IsUnknown() {
		plan.Shell = state.Shell
	}
	if plan.Uid.IsUnknown() {
		plan.Uid = state.Uid
	}
	if plan.Gid.IsUnknown() {
		plan.Gid = state.Gid
	}

	diags := resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := "user_del"
	if state.Staged.ValueBool() {
		method = "stageuser_del"
	}

	err := r.client.Call(ctx, method, []string{state.ID.ValueString()}, nil, nil)
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete FreeIPA user", err.Error())
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
