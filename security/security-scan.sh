#!/bin/bash

# ZohoSync Security Scanner
# Comprehensive SAST and dependency vulnerability scanning

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SECURITY_DIR="${PROJECT_ROOT}/security"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo "🔒 ZohoSync Security Analysis Report - ${TIMESTAMP}"
echo "=================================================="

# Create reports directory
mkdir -p "${SECURITY_DIR}/reports"
REPORT_DIR="${SECURITY_DIR}/reports/${TIMESTAMP}"
mkdir -p "${REPORT_DIR}"

cd "${PROJECT_ROOT}"

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Go security tools
install_security_tools() {
    echo "📦 Installing security tools..."
    
    if ! command_exists gosec; then
        echo "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    if ! command_exists staticcheck; then
        echo "Installing staticcheck..."
        go install honnef.co/go/tools/cmd/staticcheck@latest
    fi
    
    if ! command_exists golangci-lint; then
        echo "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    fi
    
    if ! command_exists nancy; then
        echo "Installing nancy..."
        go install github.com/sonatypecommunity/nancy@latest
    fi
}

# Install tools if needed
install_security_tools

echo ""
echo "🔍 Running SAST (Static Application Security Testing)..."
echo "======================================================="

# 1. Gosec Security Scanner
echo "Running Gosec security analysis..."
if gosec -fmt json -out "${REPORT_DIR}/gosec-report.json" -stdout -verbose=text ./...; then
    echo "✅ Gosec scan completed"
else
    echo "⚠️  Gosec found security issues - check report"
fi

# 2. StaticCheck
echo "Running StaticCheck analysis..."
if staticcheck -f json ./... > "${REPORT_DIR}/staticcheck-report.json"; then
    echo "✅ StaticCheck scan completed"
else
    echo "⚠️  StaticCheck found issues - check report"
fi

# 3. Go Vet
echo "Running Go Vet analysis..."
if go vet ./... 2> "${REPORT_DIR}/govet-report.txt"; then
    echo "✅ Go Vet scan completed"
else
    echo "⚠️  Go Vet found issues - check report"
fi

# 4. GolangCI-Lint comprehensive scan
echo "Running GolangCI-Lint comprehensive analysis..."
if golangci-lint run --out-format json > "${REPORT_DIR}/golangci-lint-report.json"; then
    echo "✅ GolangCI-Lint scan completed"
else
    echo "⚠️  GolangCI-Lint found issues - check report"
fi

echo ""
echo "🔍 Running Dependency Vulnerability Scanning..."
echo "=============================================="

# 5. Nancy - Dependency vulnerability scanner
echo "Running Nancy dependency vulnerability scan..."
if go list -json -deps ./... | nancy sleuth --loud --exclude-vulnerability-file "${SECURITY_DIR}/nancy-ignore.txt" > "${REPORT_DIR}/nancy-report.txt" 2>&1; then
    echo "✅ Nancy scan completed - no vulnerabilities found"
else
    echo "⚠️  Nancy found vulnerabilities - check report"
fi

# 6. Go mod vulnerability check (if available)
if command_exists govulncheck; then
    echo "Running govulncheck analysis..."
    if govulncheck -json ./... > "${REPORT_DIR}/govulncheck-report.json"; then
        echo "✅ Govulncheck scan completed"
    else
        echo "⚠️  Govulncheck found vulnerabilities - check report"
    fi
else
    echo "ℹ️  govulncheck not available (optional)"
fi

echo ""
echo "📊 Generating Security Summary Report..."
echo "======================================"

# Generate summary report
cat > "${REPORT_DIR}/security-summary.md" << EOF
# ZohoSync Security Analysis Report

**Generated**: ${TIMESTAMP}  
**Project**: ZohoSync  
**Scan Type**: SAST + Dependency Vulnerability Analysis

## Executive Summary

This report contains the results of automated security testing performed on the ZohoSync codebase.

## Tools Used

- **Gosec**: Go security analyzer for common security problems
- **StaticCheck**: Advanced Go static analysis
- **Go Vet**: Standard Go tool for suspicious constructs
- **GolangCI-Lint**: Comprehensive linting with security checks
- **Nancy**: Dependency vulnerability scanner
- **Govulncheck**: Go vulnerability database scanner (if available)

## Report Files

- \`gosec-report.json\` - Gosec security findings
- \`staticcheck-report.json\` - StaticCheck analysis results
- \`govet-report.txt\` - Go Vet findings
- \`golangci-lint-report.json\` - Comprehensive linting results
- \`nancy-report.txt\` - Dependency vulnerability scan
- \`govulncheck-report.json\` - Go vulnerability findings (if available)

## Security Recommendations

1. **Review all HIGH and MEDIUM severity findings**
2. **Address any hardcoded credentials or secrets**
3. **Validate all file path operations for traversal attacks**
4. **Ensure proper error handling for security-critical operations**
5. **Update dependencies with known vulnerabilities**
6. **Implement proper input validation and sanitization**

## Next Steps

1. Review detailed reports in this directory
2. Address critical and high-severity findings
3. Update dependencies as needed
4. Re-run security scan after fixes
5. Integrate security scanning into CI/CD pipeline

## Compliance Notes

- OAuth 2.0 implementation should be reviewed for proper token handling
- File operations should be validated for path traversal protection
- Network communications should use TLS/SSL
- Logging should not expose sensitive information

EOF

echo "✅ Security analysis complete!"
echo ""
echo "📋 Report Summary:"
echo "   Report Directory: ${REPORT_DIR}"
echo "   Summary Report: ${REPORT_DIR}/security-summary.md"
echo ""
echo "🔍 Next Steps:"
echo "   1. Review security-summary.md"
echo "   2. Examine individual tool reports"
echo "   3. Address high/critical findings"
echo "   4. Re-run scan after fixes"
echo ""
echo "📁 View reports: ls -la ${REPORT_DIR}"