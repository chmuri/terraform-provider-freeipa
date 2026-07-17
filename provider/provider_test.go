package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderSchema(t *testing.T) {
	p := New("test")()
	ctx := context.Background()
	resp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected provider schema to have attributes, got 0")
	}

	attrs := []string{"host", "username", "password", "insecure", "keytab_path", "realm", "krb5_conf"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected provider schema to contain attribute: %s", attr)
		}
	}
}

func TestUserResourceSchema(t *testing.T) {
	r := NewUserResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected user resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "username", "first_name", "last_name", "email", "password", "ssh_public_keys"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected user resource schema to contain attribute: %s", attr)
		}
	}
}

func TestGroupResourceSchema(t *testing.T) {
	r := NewGroupResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected group resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "name", "description", "users"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected group resource schema to contain attribute: %s", attr)
		}
	}
}

func TestHostResourceSchema(t *testing.T) {
	r := NewHostResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected host resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "fqdn", "description", "password"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected host resource schema to contain attribute: %s", attr)
		}
	}
}

func TestHbacRuleResourceSchema(t *testing.T) {
	r := NewHbacRuleResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected hbac rule resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "name", "description", "host_category", "user_category", "service_category", "users", "groups", "hosts", "services", "enabled"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected hbac rule resource schema to contain attribute: %s", attr)
		}
	}
}

func TestVaultResourceSchema(t *testing.T) {
	r := NewVaultResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected vault resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "name", "description", "type"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected vault resource schema to contain attribute: %s", attr)
		}
	}
}

func TestVaultOwnerResourceSchema(t *testing.T) {
	r := NewVaultOwnerResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected vault owner resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "name", "owner_users", "owner_groups", "owner_services"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected vault owner resource schema to contain attribute: %s", attr)
		}
	}
}

func TestVaultMemberResourceSchema(t *testing.T) {
	r := NewVaultMemberResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected vault member resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "name", "users", "groups", "services"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected vault member resource schema to contain attribute: %s", attr)
		}
	}
}

func TestSudoCommandResourceSchema(t *testing.T) {
	r := NewSudoCommandResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected sudo command resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "command", "description"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected sudo command resource schema to contain attribute: %s", attr)
		}
	}
}

func TestSudoCommandGroupResourceSchema(t *testing.T) {
	r := NewSudoCommandGroupResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)

	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected sudo command group resource schema to have attributes, got 0")
	}

	attrs := []string{"id", "name", "description", "commands"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected sudo command group resource schema to contain attribute: %s", attr)
		}
	}
}

func TestHostGroupResourceSchema(t *testing.T) {
	r := NewHostGroupResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	if len(resp.Schema.Attributes) == 0 {
		t.Error("expected hostgroup schema to have attributes, got 0")
	}
	for _, attr := range []string{"id", "cn", "description", "hosts", "host_groups", "member_managers"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected hostgroup schema to contain attribute: %s", attr)
		}
	}
}

func TestHbacSvcResourceSchema(t *testing.T) {
	r := NewHbacSvcResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected hbac service schema to contain attribute: %s", attr)
		}
	}
}

func TestHbacSvcGroupResourceSchema(t *testing.T) {
	r := NewHbacSvcGroupResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "services"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected hbac service group schema to contain attribute: %s", attr)
		}
	}
}

func TestSudoRuleResourceSchema(t *testing.T) {
	r := NewSudoRuleResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "enabled", "user_category", "host_category"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected sudo rule schema to contain attribute: %s", attr)
		}
	}
}

func TestDnsZoneResourceSchema(t *testing.T) {
	r := NewDnsZoneResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "zone_name", "authoritative_nameserver", "admin_email"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected dns zone schema to contain attribute: %s", attr)
		}
	}
}

func TestDnsRecordResourceSchema(t *testing.T) {
	r := NewDnsRecordResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "zone_name", "name", "record_type", "record_value"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected dns record schema to contain attribute: %s", attr)
		}
	}
}

func TestRoleResourceSchema(t *testing.T) {
	r := NewRoleResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "privileges", "users"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected role schema to contain attribute: %s", attr)
		}
	}
}

func TestPrivilegeResourceSchema(t *testing.T) {
	r := NewPrivilegeResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "permissions"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected privilege schema to contain attribute: %s", attr)
		}
	}
}

func TestNetgroupResourceSchema(t *testing.T) {
	r := NewNetgroupResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "nisdomainname", "users", "groups", "hosts"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected netgroup schema to contain attribute: %s", attr)
		}
	}
}

func TestPwPolicyResourceSchema(t *testing.T) {
	r := NewPwPolicyResource()
	ctx := context.Background()
	resp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "minlength", "maxlife", "minlife", "history", "priority"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected pwpolicy schema to contain attribute: %s", attr)
		}
	}
}

func TestUserDataSourceSchema(t *testing.T) {
	d := NewUserDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "username", "first_name", "last_name", "email"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected user data source schema to contain attribute: %s", attr)
		}
	}
}

func TestGroupDataSourceSchema(t *testing.T) {
	d := NewGroupDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected group data source schema to contain attribute: %s", attr)
		}
	}
}

func TestHostDataSourceSchema(t *testing.T) {
	d := NewHostDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "fqdn", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected host data source schema to contain attribute: %s", attr)
		}
	}
}

func TestHostGroupDataSourceSchema(t *testing.T) {
	d := NewHostGroupDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected hostgroup data source schema to contain attribute: %s", attr)
		}
	}
}

func TestHbacRuleDataSourceSchema(t *testing.T) {
	d := NewHbacRuleDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected hbacrule data source schema to contain attribute: %s", attr)
		}
	}
}

func TestDnsZoneDataSourceSchema(t *testing.T) {
	d := NewDnsZoneDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "zone_name"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected dns_zone data source schema to contain attribute: %s", attr)
		}
	}
}

func TestSudoRuleDataSourceSchema(t *testing.T) {
	d := NewSudoRuleDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "enabled"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected sudo_rule data source schema to contain attribute: %s", attr)
		}
	}
}

func TestSudoCommandDataSourceSchema(t *testing.T) {
	d := NewSudoCommandDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "command", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected sudo_command data source schema to contain attribute: %s", attr)
		}
	}
}

func TestSudoCommandGroupDataSourceSchema(t *testing.T) {
	d := NewSudoCommandGroupDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "commands"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected sudo_command_group data source schema to contain attribute: %s", attr)
		}
	}
}

func TestRoleDataSourceSchema(t *testing.T) {
	d := NewRoleDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "privileges"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected role data source schema to contain attribute: %s", attr)
		}
	}
}

func TestPrivilegeDataSourceSchema(t *testing.T) {
	d := NewPrivilegeDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "permissions"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected privilege data source schema to contain attribute: %s", attr)
		}
	}
}

func TestNetgroupDataSourceSchema(t *testing.T) {
	d := NewNetgroupDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected netgroup data source schema to contain attribute: %s", attr)
		}
	}
}

func TestPwPolicyDataSourceSchema(t *testing.T) {
	d := NewPwPolicyDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "minlength", "maxlife"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected pwpolicy data source schema to contain attribute: %s", attr)
		}
	}
}

func TestVaultDataSourceSchema(t *testing.T) {
	d := NewVaultDataSource()
	ctx := context.Background()
	resp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, resp)
	for _, attr := range []string{"id", "name", "description", "type"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected vault data source schema to contain attribute: %s", attr)
		}
	}
}
