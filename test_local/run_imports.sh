#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TF_DIR="$SCRIPT_DIR"
TERRAFORMRC="$TF_DIR/terraformrc"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

PASS=0
FAIL=0

log_pass()  { echo -e "  ${GREEN}PASS${NC} $1"; }
log_fail()  { echo -e "  ${RED}FAIL${NC} $1"; }

setup() {
    if [ ! -f "$PROJECT_ROOT/terraform-provider-freeipa" ]; then
        (cd "$PROJECT_ROOT" && go build -o terraform-provider-freeipa .)
    fi
    echo "Restoring clean FreeIPA state..."
    (cd "$PROJECT_ROOT" && make docker-up)
    export TF_CLI_CONFIG_FILE="$TERRAFORMRC"
}

cleanup() {
    (cd "$PROJECT_ROOT" && make docker-down) 2>/dev/null || true
}

run_import() {
    local name="$1"
    local tf_config="$2"
    local resource_addr="$3"
    local import_id="$4"

    echo ""
    echo "[$name] import $resource_addr"

    local workdir="$TF_DIR/.work"
    rm -rf "$workdir"
    mkdir -p "$workdir"

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
    cp "$tf_config" "$workdir/main.tf"

    if ! (cd "$workdir" && terraform init -input=false > /dev/null 2>&1); then
        log_fail "$name: terraform init failed"
        ((FAIL++))
        return
    fi

    if ! (cd "$workdir" && echo "yes" | terraform apply -input=false -auto-approve > /dev/null 2>&1); then
        log_fail "$name: terraform apply failed"
        ((FAIL++))
        return
    fi

    rm -f "$workdir/terraform.tfstate" "$workdir/terraform.tfstate.backup"

    if ! (cd "$workdir" && terraform import -input=false "$resource_addr" "$import_id" > /dev/null 2>&1); then
        log_fail "$name: terraform import failed"
        ((FAIL++))
        return
    fi

    if ! (cd "$workdir" && terraform plan -input=false -detailed-exitcode > /dev/null 2>&1); then
        local plan_rc=$?
        if [ "$plan_rc" -eq 0 ]; then
            log_pass "$name: import OK, plan shows no diff"
        else
            log_fail "$name: import succeeded but plan showed diff (exit $plan_rc)"
            ((FAIL++))
            return
        fi
    fi

    (cd "$workdir" && echo "yes" | terraform destroy -input=false -auto-approve > /dev/null 2>&1) || true

    log_pass "$name: import + verify OK"
    ((PASS++))
}

main() {
    echo "=== FreeIPA Terraform Import Tests ==="

    trap cleanup EXIT
    setup

    SCENARIOS_DIR="$TF_DIR/scenarios"

    run_import "user" "$SCENARIOS_DIR/user/basic.tf" "freeipa_user.test" "acc_user_basic"
    run_import "group" "$SCENARIOS_DIR/group/basic.tf" "freeipa_group.test" "acc-group-basic"
    run_import "host" "$SCENARIOS_DIR/host/basic.tf" "freeipa_host.test" "acc-basic.test.local"
    run_import "hostgroup" "$SCENARIOS_DIR/hostgroup/basic.tf" "freeipa_hostgroup.test" "acc-hg-basic"
    run_import "hbac_service" "$SCENARIOS_DIR/hbac/service.tf" "freeipa_hbac_service.test" "acc-hbacsvc"
    run_import "hbac_rule" "$SCENARIOS_DIR/hbac/rule.tf" "freeipa_hbacrule.test" "acc-hbac-basic"
    run_import "sudo_command" "$SCENARIOS_DIR/sudo/command.tf" "freeipa_sudo_command.test" "/usr/bin/acc-test"
    run_import "sudo_rule" "$SCENARIOS_DIR/sudo/rule.tf" "freeipa_sudo_rule.test" "acc-sudo-basic"
    run_import "dns_zone" "$SCENARIOS_DIR/dns/zone.tf" "freeipa_dns_zone.test" "acc.test.local"
    run_import "privilege" "$SCENARIOS_DIR/role/privilege.tf" "freeipa_privilege.test" "acc-priv-basic"
    run_import "role" "$SCENARIOS_DIR/role/role.tf" "freeipa_role.test" "acc-role-basic"
    run_import "netgroup" "$SCENARIOS_DIR/netgroup/basic.tf" "freeipa_netgroup.test" "acc-ng-basic"
    run_import "vault" "$SCENARIOS_DIR/vault/basic.tf" "freeipa_vault.test" "acc-vault-basic"

    echo ""
    echo "=== Import Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC} ==="
    [ "$FAIL" -eq 0 ] || exit 1
}

main "$@"
