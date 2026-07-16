package provider

import (
	"context"
	"testing"

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
