#!/bin/bash

# Containr Security Audit Script
# Comprehensive security testing and vulnerability assessment

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

RESULTS_DIR="security-audit-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$RESULTS_DIR"

TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

echo -e "${BLUE}╔══════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Containr Security Audit & Testing     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
echo
echo "Results directory: $RESULTS_DIR"
echo

# Helper functions
log_test() {
    echo -e "\n${YELLOW}[TEST]${NC} $1"
    ((TOTAL_TESTS++))
}

log_pass() {
    echo -e "${GREEN}  ✓ PASS:${NC} $1"
    ((PASSED_TESTS++))
}

log_fail() {
    echo -e "${RED}  ✗ FAIL:${NC} $1"
    ((FAILED_TESTS++))
}

log_warn() {
    echo -e "${YELLOW}  ⚠ WARN:${NC} $1"
    ((WARNINGS++))
}

log_info() {
    echo -e "${BLUE}  ℹ INFO:${NC} $1"
}

# 1. Static Code Analysis
echo -e "\n${BLUE}=== Phase 1: Static Code Analysis ===${NC}"

log_test "Running gosec (security scanner)"
if command -v gosec &> /dev/null; then
    gosec -fmt=json -out="$RESULTS_DIR/gosec.json" ./... 2>&1 | tee "$RESULTS_DIR/gosec.log"
    ISSUES=$(jq '.Issues | length' "$RESULTS_DIR/gosec.json" 2>/dev/null || echo "0")
    if [ "$ISSUES" -eq 0 ]; then
        log_pass "No security issues found by gosec"
    else
        log_warn "Found $ISSUES potential security issues"
    fi
else
    log_warn "gosec not installed (run: go install github.com/securego/gosec/v2/cmd/gosec@latest)"
fi

log_test "Running staticcheck (static analysis)"
if command -v staticcheck &> /dev/null; then
    staticcheck ./... > "$RESULTS_DIR/staticcheck.log" 2>&1 || true
    ISSUES=$(wc -l < "$RESULTS_DIR/staticcheck.log")
    if [ "$ISSUES" -eq 0 ]; then
        log_pass "No issues found by staticcheck"
    else
        log_warn "Found $ISSUES static analysis issues"
    fi
else
    log_warn "staticcheck not installed"
fi

log_test "Checking for hardcoded secrets"
if command -v gitleaks &> /dev/null; then
    gitleaks detect --no-git --report-path="$RESULTS_DIR/gitleaks.json" 2>&1 || true
    if [ $? -eq 0 ]; then
        log_pass "No hardcoded secrets found"
    else
        log_fail "Potential hardcoded secrets detected - review gitleaks report"
    fi
else
    # Manual check for common patterns
    grep -r -n "password\|secret\|api_key\|token" . --include="*.go" \
        --exclude-dir="vendor" --exclude-dir=".git" > "$RESULTS_DIR/secret-patterns.txt" 2>/dev/null || true
    if [ -s "$RESULTS_DIR/secret-patterns.txt" ]; then
        log_warn "Found potential secret patterns - manual review required"
    else
        log_pass "No obvious secret patterns found"
    fi
fi

# 2. Dependency Security
echo -e "\n${BLUE}=== Phase 2: Dependency Security ===${NC}"

log_test "Checking for vulnerable dependencies"
if command -v nancy &> /dev/null; then
    go list -json -m all | nancy sleuth > "$RESULTS_DIR/nancy.log" 2>&1 || true
    log_info "Dependency audit complete - check nancy.log"
else
    go list -m all > "$RESULTS_DIR/dependencies.txt"
    log_warn "nancy not installed - dependency list saved"
fi

log_test "Checking Go version"
GO_VERSION=$(go version | awk '{print $3}')
log_info "Go version: $GO_VERSION"

#  3. Capability & Privilege Tests
echo -e "\n${BLUE}=== Phase 3: Capability & Privilege Tests ===${NC}"

log_test "Verifying default capability dropping"
if grep -q "CAP_NET_RAW" pkg/capabilities/*.go; then
    log_pass "CAP_NET_RAW is being managed"
else
    log_warn "CAP_NET_RAW management not found"
fi

log_test "Checking for unnecessary privileged operations"
PRIV_OPS=$(grep -r "os.Setuid\|syscall.Setuid\|os.Setgid" pkg/ --include="*.go" | wc -l)
if [ "$PRIV_OPS" -gt 0 ]; then
    log_warn "Found $PRIV_OPS potential privileged operations - review required"
else
    log_pass "No obvious privileged operations in user-facing code"
fi

# 4. Seccomp Profile Tests
echo -e "\n${BLUE}=== Phase 4: Seccomp Profile Security ===${NC}"

log_test "Checking seccomp profile implementation"
if [ -f "pkg/seccomp/seccomp.go" ]; then
    if grep -q "DefaultProfile" pkg/seccomp/seccomp.go; then
        log_pass "Default seccomp profile implemented"
    else
        log_fail "No default seccomp profile found"
    fi
else
    log_fail "Seccomp package not found"
fi

# 5. Network Security Tests
echo -e "\n${BLUE}=== Phase 5: Network Security ===${NC}"

log_test "Checking for insecure network operations"
INSECURE_NET=$(grep -r "http://" pkg/ --include="*.go" | grep -v "localhost\|127.0.0.1\|comment" | wc -l)
if [ "$INSECURE_NET" -eq 0 ]; then
    log_pass "No insecure HTTP connections found"
else
    log_warn "Found $INSECURE_NET potential insecure HTTP connections"
fi

log_test "Checking TLS configuration"
if grep -q "tls.Config" pkg/ -r --include="*.go"; then
    if grep -q "MinVersion.*tls.VersionTLS12" pkg/ -r --include="*.go"; then
        log_pass "TLS 1.2+ enforced"
    else
        log_warn "TLS version enforcement not found"
    fi
else
    log_info "No TLS configuration found (may be intended)"
fi

# 6. Input Validation
echo -e "\n${BLUE}=== Phase 6: Input Validation ===${NC}"

log_test "Checking for path traversal vulnerabilities"
if grep -q "filepath.Clean\|filepath.Abs" pkg/rootfs/*.go pkg/volume/*.go; then
    log_pass "Path sanitization found"
else
    log_warn "Path sanitization not evident - manual review recommended"
fi

log_test "Checking command injection prevention"
if grep -q "exec.Command" pkg/ -r --include="*.go"; then
    # Check if commands are properly sanitized
    if grep -B5 "exec.Command" pkg/ -r --include="*.go" | grep -q "strings.Contains\|validation"; then
        log_pass "Command validation found"
    else
        log_warn "exec.Command found - ensure proper input validation"
    fi
fi

# 7. Container Escape Tests
echo -e "\n${BLUE}=== Phase 7: Container Escape Prevention ===${NC}"

log_test "Checking namespace isolation"
for ns in pid mount uts ipc net user; do
    if grep -q "CLONE_NEW$(echo $ns | tr '[:lower:]' '[:upper:]')" pkg/namespace/*.go; then
        log_pass "$(echo $ns | tr '[:lower:]' '[:upper:]') namespace isolation implemented"
    else
        log_fail "$(echo $ns | tr '[:lower:]' '[:upper:]') namespace isolation not found"
    fi
done

log_test "Checking pivot_root implementation"
if grep -q "pivot_root\|PivotRoot" pkg/rootfs/*.go; then
    log_pass "pivot_root implemented for filesystem isolation"
else
    log_warn "pivot_root implementation not found"
fi

# 8. Resource Limit Enforcement
echo -e "\n${BLUE}=== Phase 8: Resource Limit Enforcement ===${NC}"

log_test "Checking cgroup limits"
for limit in memory cpu pids; do
    if grep -q "$limit" pkg/cgroup/*.go; then
        log_pass "$(echo $limit | tr '[:lower:]' '[:upper:]') cgroup limits implemented"
    else
        log_warn "$(echo $limit | tr '[:lower:]' '[:upper:]') cgroup limits not found"
    fi
done

# 9. File Permission Checks
echo -e "\n${BLUE}=== Phase 9: File Permissions & Ownership ===${NC}"

log_test "Checking sensitive file permissions"
SENSITIVE_FILES=$(find pkg/ -name "*.go" -exec grep -l "0600\|0400\|0700" {} \; | wc -l)
log_info "Files with restrictive permissions: $SENSITIVE_FILES"

# 10. Error Handling & Information Disclosure
echo -e "\n${BLUE}=== Phase 10: Error Handling ===${NC}"

log_test "Checking for verbose error messages"
VERBOSE_ERRORS=$(grep -r "fmt.Printf.*err\|log.Fatal.*%v" pkg/ --include="*.go" | wc -l)
if [ "$VERBOSE_ERRORS" -gt 20 ]; then
    log_warn "Many verbose error messages found - may leak information"
else
    log_pass "Error handling appears reasonable"
fi

# Generate Report
echo -e "\n${BLUE}=== Generating Security Report ===${NC}"

cat > "$RESULTS_DIR/security-report.md" <<EOF
# Containr Security Audit Report

**Date:** $(date)
**Auditor:** Automated Security Scanner

## Summary

- **Total Tests:** $TOTAL_TESTS
- **Passed:** $PASSED_TESTS
- **Failed:** $FAILED_TESTS
- **Warnings:** $WARNINGS

## Test Results

### Phase 1: Static Code Analysis
- gosec security scanner results available in \`gosec.json\`
- staticcheck results in \`staticcheck.log\`

### Phase 2: Dependency Security
- Dependency list in \`dependencies.txt\`
- Go version: $GO_VERSION

### Phase 3-10: Security Controls
See detailed logs above for individual test results.

## Recommendations

EOF

if [ $FAILED_TESTS -gt 0 ]; then
    echo "1. Address all failed security tests immediately" >> "$RESULTS_DIR/security-report.md"
fi

if [ $WARNINGS -gt 5 ]; then
    echo "2. Review and address warning items" >> "$RESULTS_DIR/security-report.md"
fi

echo "3. Conduct manual penetration testing" >> "$RESULTS_DIR/security-report.md"
echo "4. Review all exec.Command usages for injection vulnerabilities" >> "$RESULTS_DIR/security-report.md"
echo "5. Implement automated security scanning in CI/CD" >> "$RESULTS_DIR/security-report.md"

# Summary
echo -e "\n${BLUE}╔══════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Security Audit Complete         ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════╝${NC}"
echo
echo "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
echo
echo "Detailed report: $RESULTS_DIR/security-report.md"

# Exit code
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "\n${RED}Security audit FAILED${NC}"
    exit 1
else
    echo -e "\n${GREEN}Security audit PASSED${NC}"
    exit 0
fi
