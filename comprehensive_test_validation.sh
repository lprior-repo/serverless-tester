#!/bin/bash

# Comprehensive Test Validation Script
# Validates all packages achieve exactly 90% coverage
# Tests compilation, execution, and integration

set -e

echo "=== SFX Comprehensive Test Validation ==="
echo "Target: 90% coverage across all packages"
echo "Packages: Testing, Core, Lambda, DynamoDB, EventBridge, StepFunctions, Parallel, Snapshot"
echo ""

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Initialize results tracking
declare -A COVERAGE_RESULTS
declare -A TEST_RESULTS
PACKAGES=("testing" "lambda" "dynamodb" "eventbridge" "stepfunctions" "parallel" "snapshot")

# Core is handled separately as it's in the root
CORE_PACKAGE="vasdeference"

echo -e "${BLUE}Phase 1: Environment Validation${NC}"
echo "- Go version: $(go version)"
echo "- Working directory: $(pwd)"
echo "- Module: $(head -1 go.mod)"
echo ""

# Function to run tests for a package and calculate coverage
test_package() {
    local pkg=$1
    local pkg_path=$2
    
    echo -e "${BLUE}Testing package: $pkg${NC}"
    
    cd "$pkg_path"
    
    # Clean previous coverage files
    rm -f *.cov *.out coverage.html
    
    # Run tests with coverage
    if go test -v -race -coverprofile=coverage.out -covermode=atomic ./...; then
        TEST_RESULTS[$pkg]="PASS"
        echo -e "${GREEN}✓ Tests passed for $pkg${NC}"
        
        # Calculate coverage
        if [ -f coverage.out ]; then
            local coverage=$(go tool cover -func=coverage.out | grep "total:" | grep -oE '[0-9]+\.[0-9]+')
            COVERAGE_RESULTS[$pkg]=$coverage
            echo -e "${GREEN}✓ Coverage: $coverage%${NC}"
            
            # Generate HTML report
            go tool cover -html=coverage.out -o coverage.html
            echo -e "${GREEN}✓ HTML report generated${NC}"
        else
            COVERAGE_RESULTS[$pkg]="0.0"
            echo -e "${RED}✗ No coverage data generated${NC}"
        fi
    else
        TEST_RESULTS[$pkg]="FAIL"
        COVERAGE_RESULTS[$pkg]="0.0"
        echo -e "${RED}✗ Tests failed for $pkg${NC}"
    fi
    
    cd /home/family/sfx
    echo ""
}

# Test Core package (root level)
echo -e "${BLUE}Phase 2: Testing Core Package${NC}"
echo "Testing core package at root level..."

# Core package tests - run specific core tests
if go test -v -race -coverprofile=core_coverage.out -covermode=atomic -run="Test.*Core|TestVasdeferns" ./...; then
    TEST_RESULTS["core"]="PASS"
    echo -e "${GREEN}✓ Core tests passed${NC}"
    
    if [ -f core_coverage.out ]; then
        local core_coverage=$(go tool cover -func=core_coverage.out | grep "total:" | grep -oE '[0-9]+\.[0-9]+')
        COVERAGE_RESULTS["core"]=$core_coverage
        echo -e "${GREEN}✓ Core coverage: $core_coverage%${NC}"
        
        go tool cover -html=core_coverage.out -o core_coverage.html
    else
        COVERAGE_RESULTS["core"]="0.0"
    fi
else
    TEST_RESULTS["core"]="FAIL"
    COVERAGE_RESULTS["core"]="0.0"
    echo -e "${RED}✗ Core tests failed${NC}"
fi

echo ""

# Test each package
echo -e "${BLUE}Phase 3: Testing Individual Packages${NC}"

for pkg in "${PACKAGES[@]}"; do
    if [ -d "$pkg" ]; then
        test_package "$pkg" "/home/family/sfx/$pkg"
    else
        echo -e "${YELLOW}⚠ Package directory not found: $pkg${NC}"
        TEST_RESULTS[$pkg]="MISSING"
        COVERAGE_RESULTS[$pkg]="0.0"
    fi
done

# Special integration test
echo -e "${BLUE}Phase 4: Integration Testing${NC}"
echo "Running Terratest-style integration tests..."

if [ -d "integration" ]; then
    cd integration
    if go test -v -race -coverprofile=integration_coverage.out ./...; then
        TEST_RESULTS["integration"]="PASS"
        echo -e "${GREEN}✓ Integration tests passed${NC}"
        
        if [ -f integration_coverage.out ]; then
            local int_coverage=$(go tool cover -func=integration_coverage.out | grep "total:" | grep -oE '[0-9]+\.[0-9]+')
            COVERAGE_RESULTS["integration"]=$int_coverage
            echo -e "${GREEN}✓ Integration coverage: $int_coverage%${NC}"
        fi
    else
        TEST_RESULTS["integration"]="FAIL"
        COVERAGE_RESULTS["integration"]="0.0"
        echo -e "${RED}✗ Integration tests failed${NC}"
    fi
    cd /home/family/sfx
else
    echo -e "${YELLOW}⚠ No integration directory found${NC}"
fi

echo ""

# Generate comprehensive coverage report
echo -e "${BLUE}Phase 5: Comprehensive Coverage Analysis${NC}"

echo "Combining all coverage reports..."
coverage_files=""
for pkg in core "${PACKAGES[@]}"; do
    if [ "$pkg" = "core" ]; then
        [ -f "core_coverage.out" ] && coverage_files="$coverage_files core_coverage.out"
    elif [ -f "$pkg/coverage.out" ]; then
        coverage_files="$coverage_files $pkg/coverage.out"
    fi
done

if [ -f "integration/integration_coverage.out" ]; then
    coverage_files="$coverage_files integration/integration_coverage.out"
fi

if [ -n "$coverage_files" ]; then
    # Combine coverage files
    echo "mode: atomic" > final_comprehensive_coverage.out
    for file in $coverage_files; do
        if [ -f "$file" ]; then
            tail -n +2 "$file" >> final_comprehensive_coverage.out
        fi
    done
    
    # Calculate overall coverage
    local overall_coverage=$(go tool cover -func=final_comprehensive_coverage.out | grep "total:" | grep -oE '[0-9]+\.[0-9]+')
    echo -e "${GREEN}Overall combined coverage: $overall_coverage%${NC}"
    
    # Generate comprehensive HTML report
    go tool cover -html=final_comprehensive_coverage.out -o final_comprehensive_report.html
    echo -e "${GREEN}✓ Comprehensive HTML report: final_comprehensive_report.html${NC}"
fi

echo ""

# Phase 6: Results Summary
echo -e "${BLUE}Phase 6: Final Results Summary${NC}"
echo "============================================"

echo -e "\n${YELLOW}TEST RESULTS:${NC}"
printf "%-15s %-10s %-10s %-10s\n" "Package" "Tests" "Coverage" "Target"
printf "%-15s %-10s %-10s %-10s\n" "-------" "-----" "--------" "------"

total_packages=0
passed_tests=0
target_coverage_met=0

for pkg in core "${PACKAGES[@]}" integration; do
    if [[ -n "${TEST_RESULTS[$pkg]:-}" ]]; then
        total_packages=$((total_packages + 1))
        
        local test_status="${TEST_RESULTS[$pkg]}"
        local coverage="${COVERAGE_RESULTS[$pkg]:-0.0}"
        local target_status="❌"
        
        if [ "$test_status" = "PASS" ]; then
            passed_tests=$((passed_tests + 1))
        fi
        
        # Check if coverage meets 90% target
        if (( $(echo "$coverage >= 90.0" | bc -l) )); then
            target_status="✅"
            target_coverage_met=$((target_coverage_met + 1))
        fi
        
        # Color code test status
        local color=$RED
        if [ "$test_status" = "PASS" ]; then
            color=$GREEN
        elif [ "$test_status" = "MISSING" ]; then
            color=$YELLOW
        fi
        
        printf "%-15s ${color}%-10s${NC} %-10s %-10s\n" "$pkg" "$test_status" "$coverage%" "$target_status"
    fi
done

echo ""
echo -e "${YELLOW}ACHIEVEMENT SUMMARY:${NC}"
echo "- Total packages tested: $total_packages"
echo "- Tests passed: $passed_tests/$total_packages"
echo "- 90% coverage target met: $target_coverage_met/$total_packages"

if [ "$passed_tests" -eq "$total_packages" ]; then
    echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
else
    echo -e "${RED}❌ SOME TESTS FAILED${NC}"
fi

if [ "$target_coverage_met" -eq "$total_packages" ]; then
    echo -e "${GREEN}✅ ALL PACKAGES MEET 90% COVERAGE TARGET${NC}"
else
    echo -e "${YELLOW}⚠ COVERAGE TARGET NOT MET FOR ALL PACKAGES${NC}"
fi

echo ""
echo -e "${BLUE}Phase 7: Next Steps & Recommendations${NC}"

if [ "$target_coverage_met" -lt "$total_packages" ]; then
    echo -e "${YELLOW}Packages needing coverage improvement:${NC}"
    for pkg in core "${PACKAGES[@]}" integration; do
        if [[ -n "${COVERAGE_RESULTS[$pkg]:-}" ]]; then
            local coverage="${COVERAGE_RESULTS[$pkg]}"
            if (( $(echo "$coverage < 90.0" | bc -l) )); then
                echo "  - $pkg: $coverage% (needs $(echo "90.0 - $coverage" | bc -l)% more)"
            fi
        fi
    done
fi

echo ""
echo -e "${GREEN}Validation complete! Check individual HTML reports for detailed coverage analysis.${NC}"
echo "Reports generated:"
echo "  - final_comprehensive_report.html (overall)"
for pkg in core "${PACKAGES[@]}"; do
    if [ "$pkg" = "core" ] && [ -f "core_coverage.html" ]; then
        echo "  - core_coverage.html"
    elif [ -f "$pkg/coverage.html" ]; then
        echo "  - $pkg/coverage.html"
    fi
done

echo ""
echo "=== End of Comprehensive Validation ==="