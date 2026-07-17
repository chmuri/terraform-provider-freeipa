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

run_error_test() {
    local name="$1"
    local tf_file="$2"

    echo ""
    echo "[$name] should fail"

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
    cp "$tf_file" "$workdir/main.tf"

    if ! (cd "$workdir" && terraform init -input=false > /dev/null 2>&1); then
        log_fail "$name: terraform init should not fail"
        ((FAIL++))
        return
    fi

    if (cd "$workdir" && echo "yes" | terraform apply -input=false -auto-approve > /tmp/tf_err_$$.log 2>&1); then
        log_fail "$name: apply succeeded but was expected to fail"
        ((FAIL++))
        return
    fi

    log_pass "$name: correctly failed"
    ((PASS++))
    rm -f /tmp/tf_err_$$.log
}

main() {
    echo "=== FreeIPA Terraform Error Scenario Tests ==="

    trap cleanup EXIT
    setup

    SCENARIOS_DIR="$TF_DIR/scenarios"

    declare -A TESTS=(
        ["duplicate_user"]="$SCENARIOS_DIR/errors/duplicate.tf"
    )

    for name in "${!TESTS[@]}"; do
        run_error_test "$name" "${TESTS[$name]}"
    done

    echo ""
    echo "=== Error Test Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC} ==="
    [ "$FAIL" -eq 0 ] || exit 1
}

main "$@"
