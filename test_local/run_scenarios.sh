#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TF_DIR="$SCRIPT_DIR"
SCENARIOS_DIR="$TF_DIR/scenarios"
TERRAFORMRC="$TF_DIR/terraformrc"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0

log_pass()  { echo -e "  ${GREEN}PASS${NC} $1"; }
log_fail()  { echo -e "  ${RED}FAIL${NC} $1"; }
log_skip()  { echo -e "  ${YELLOW}SKIP${NC} $1"; }

setup() {
    if [ ! -f "$PROJECT_ROOT/terraform-provider-freeipa" ]; then
        echo "Building provider..."
        (cd "$PROJECT_ROOT" && go build -o terraform-provider-freeipa .)
    fi

    if [ ! -f "$PROJECT_ROOT/.ipa-clean-volume.tar.gz" ]; then
        echo "Creating clean IPA snapshot..."
        (cd "$PROJECT_ROOT" && make snapshot-create)
    fi

    echo "Restoring clean FreeIPA state..."
    (cd "$PROJECT_ROOT" && make docker-up)

    cp "$TERRAFORMRC" "$HOME/.terraformrc" 2>/dev/null || {
        echo "WARNING: Could not copy terraformrc. Trying CLI config instead..."
        export TF_CLI_CONFIG_FILE="$TERRAFORMRC"
    }
}

cleanup() {
    echo "Stopping FreeIPA..."
    (cd "$PROJECT_ROOT" && make docker-down) 2>/dev/null || true
}

run_scenario() {
    local name="$1"
    local tf_file="$2"
    local expect="${3:-pass}"

    echo ""
    echo "[$name] $(basename "$tf_file")"

    local workdir="$TF_DIR/.work"
    rm -rf "$workdir"
    mkdir -p "$workdir"

    # Copy provider config + scenario into workdir
    cat > "$workdir/provider.tf" <<'PROVIDER'
terraform {
  required_providers {
    freeipa = {
      source = "chmuri/freeipa"
    }
  }
}
provider "freeipa" {
  host     = "ipa.test.local"
  username = "admin"
  password = "SecretAdminPassword123!"
  insecure = true
}
PROVIDER
    cp "$tf_file" "$workdir/main.tf"

    # Init
    if ! (cd "$workdir" && terraform init -input=false > /dev/null 2>&1); then
        log_fail "$name: terraform init failed"
        ((FAIL++))
        return
    fi

    # Plan
    if ! (cd "$workdir" && terraform plan -input=false -detailed-exitcode > /dev/null 2>&1); then
        local plan_rc=$?
        if [ "$plan_rc" -eq 2 ]; then
            : # Changes expected, this is OK
        else
            log_fail "$name: terraform plan failed (exit $plan_rc)"
            ((FAIL++))
            return
        fi
    fi

    # Apply
    if ! (cd "$workdir" && echo "yes" | terraform apply -input=false -auto-approve > /tmp/tf_apply_$$.log 2>&1); then
        log_fail "$name: terraform apply failed"
        tail -20 /tmp/tf_apply_$$.log
        ((FAIL++))
        return
    fi

    if [ "$expect" = "plan_fails" ]; then
        log_fail "$name: apply succeeded but was expected to fail"
        ((FAIL++))
        return
    fi

    # Idempotency check: second plan should show no changes
    if (cd "$workdir" && terraform plan -input=false -detailed-exitcode > /dev/null 2>&1); then
        : # exit 0 means no changes -- good
    else
        local plan2_rc=$?
        if [ "$plan2_rc" -eq 2 ]; then
            log_fail "$name: NOT IDEMPOTENT -- second plan shows changes"
            ((FAIL++))
            return
        fi
    fi

    # Destroy
    if ! (cd "$workdir" && echo "yes" | terraform destroy -input=false -auto-approve > /tmp/tf_destroy_$$.log 2>&1); then
        log_fail "$name: terraform destroy failed"
        tail -20 /tmp/tf_destroy_$$.log
        ((FAIL++))
        return
    fi

    log_pass "$name: apply + destroy OK"
    ((PASS++))
    rm -f /tmp/tf_apply_$$.log /tmp/tf_destroy_$$.log
}

get_tf_files() {
    local dir="$1"
    find "$dir" -name "*.tf" -type f | sort
}

main() {
    echo "=== FreeIPA Terraform Provider Integration Tests ==="

    trap cleanup EXIT
    setup

    SCENARIOS=(
        "user_basic:$SCENARIOS_DIR/user/basic.tf"
        "user_random_password:$SCENARIOS_DIR/user/random_password.tf"
        "user_staged:$SCENARIOS_DIR/user/staged.tf"
        "user_disabled:$SCENARIOS_DIR/user/disabled.tf"
        "group_basic:$SCENARIOS_DIR/group/basic.tf"
        "group_with_users:$SCENARIOS_DIR/group/with_users.tf"
        "group_nested:$SCENARIOS_DIR/group/nested.tf"
        "host_basic:$SCENARIOS_DIR/host/basic.tf"
        "hostgroup_basic:$SCENARIOS_DIR/hostgroup/basic.tf"
        "hbac_service:$SCENARIOS_DIR/hbac/service.tf"
        "hbac_service_group:$SCENARIOS_DIR/hbac/service_group.tf"
        "hbac_rule:$SCENARIOS_DIR/hbac/rule.tf"
        "hbac_rule_disabled:$SCENARIOS_DIR/hbac/rule_disabled.tf"
        "sudo_command:$SCENARIOS_DIR/sudo/command.tf"
        "sudo_command_group:$SCENARIOS_DIR/sudo/command_group.tf"
        "sudo_rule:$SCENARIOS_DIR/sudo/rule.tf"
        "sudo_rule_disabled:$SCENARIOS_DIR/sudo/rule_disabled.tf"
        "dns_zone:$SCENARIOS_DIR/dns/zone.tf"
        "dns_zone_disabled:$SCENARIOS_DIR/dns/zone_disabled.tf"
        "dns_record:$SCENARIOS_DIR/dns/record.tf"
        "dns_record_mx:$SCENARIOS_DIR/dns/record_mx.tf"
        "dns_record_srv:$SCENARIOS_DIR/dns/record_srv.tf"
        "dns_record_ptr:$SCENARIOS_DIR/dns/record_ptr.tf"
        "dns_record_ns:$SCENARIOS_DIR/dns/record_ns.tf"
        "dns_record_cname:$SCENARIOS_DIR/dns/record_cname.tf"
        "dns_record_sshfp:$SCENARIOS_DIR/dns/record_sshfp.tf"
        "privilege:$SCENARIOS_DIR/role/privilege.tf"
        "role:$SCENARIOS_DIR/role/role.tf"
        "netgroup:$SCENARIOS_DIR/netgroup/basic.tf"
        "pwpolicy:$SCENARIOS_DIR/pwpolicy/basic.tf"
        "vault:$SCENARIOS_DIR/vault/basic.tf"
        "vault_symmetric:$SCENARIOS_DIR/vault/symmetric.tf"
        "ds_user:$SCENARIOS_DIR/ds/user.tf"
        "ds_group:$SCENARIOS_DIR/ds/group.tf"
        "ds_host:$SCENARIOS_DIR/ds/host.tf"
        "ds_sudo_rule:$SCENARIOS_DIR/ds/sudo_rule.tf"
        "ds_role:$SCENARIOS_DIR/ds/role.tf"
        "ds_privilege:$SCENARIOS_DIR/ds/privilege.tf"
        "ds_netgroup:$SCENARIOS_DIR/ds/netgroup.tf"
        "ds_pwpolicy:$SCENARIOS_DIR/ds/pwpolicy.tf"
        "ds_vault:$SCENARIOS_DIR/ds/vault.tf"
        "dns_zone_forwarders:$SCENARIOS_DIR/dns/zone_forwarders.tf"
        "sudo_rule_options:$SCENARIOS_DIR/sudo/rule_options.tf"
        "sudo_rule_runas:$SCENARIOS_DIR/sudo/rule_runas.tf"
        "hostgroup_nested:$SCENARIOS_DIR/hostgroup/nested_hosts.tf"
        "vault_owners_members_groups:$SCENARIOS_DIR/vault/owners_members_groups.tf"
        "cross_full_stack:$SCENARIOS_DIR/cross/full_stack.tf"
        "destroy_order:$SCENARIOS_DIR/destroy_order/chained.tf"
        "destroy_order_complex:$SCENARIOS_DIR/destroy_order/complex.tf"
    )

    for scenario in "${SCENARIOS[@]}"; do
        name="${scenario%%:*}"
        file="${scenario#*:}"
        run_scenario "$name" "$file"
    done

    echo ""
    echo "=== Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}, ${YELLOW}$SKIP skipped${NC} ==="

    if [ "$FAIL" -gt 0 ]; then
        exit 1
    fi
}

main "$@"
