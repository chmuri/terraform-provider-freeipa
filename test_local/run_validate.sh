#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

PASS=0
FAIL=0

log_pass()  { echo -e "  ${GREEN}PASS${NC} $1"; }
log_fail()  { echo -e "  ${RED}FAIL${NC} $1"; }

validate() {
    local name="$1"
    local tf_file="$2"

    local workdir="$SCRIPT_DIR/.work"
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
        log_fail "$name: init failed"
        ((FAIL++))
        return
    fi

    if ! (cd "$workdir" && terraform plan -input=false -detailed-exitcode > /dev/null 2>&1); then
        local rc=$?
        if [ "$rc" -eq 0 ] || [ "$rc" -eq 2 ]; then
            : # 0=no changes, 2=changes -- both mean valid
        else
            log_fail "$name: plan failed (exit $rc)"
            ((FAIL++))
            return
        fi
    fi

    log_pass "$name"
    ((PASS++))
}

main() {
    echo "=== Validating all Terraform scenarios ==="

    export TF_CLI_CONFIG_FILE="$SCRIPT_DIR/terraformrc"

    SCENARIOS_DIR="$SCRIPT_DIR/scenarios"

    while IFS= read -r tf; do
        [ -z "$tf" ] && continue
        local_name="${tf#$SCENARIOS_DIR/}"
        local_name="${local_name%.tf}"
        validate "$local_name" "$tf"
    done < <(find "$SCENARIOS_DIR" -name "*.tf" | sort)

    echo ""
    echo "=== Validation Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC} ($TOTAL total) ==="
    [ "$FAIL" -eq 0 ] || exit 1
}

main "$@"
