#!/bin/bash

# VasDeference Test Execution Script
# This script runs all tests in the correct order with proper error handling

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TIMEOUT=300s
VERBOSE=${VERBOSE:-false}
COVERAGE=${COVERAGE:-true}
PARALLEL=${PARALLEL:-true}

# Test packages in dependency order
TEST_PACKAGES=(
    "./testing/helpers"
    "./testing"
    "./snapshot"
    "./parallel"
    "./dynamodb"
    "./eventbridge" 
    "./stepfunctions"
    "./integration"
)

# Functions
print_header() {
    echo -e "${GREEN}===================================================="
    echo -e "VasDeference Test Suite"
    echo -e "====================================================${NC}"
}

print_separator() {
    echo -e "${YELLOW}----------------------------------------------------${NC}"
}

run_package_tests() {
    local package=$1
    local package_name=$(basename "$package")
    
    echo -e "${GREEN}Testing package: $package_name${NC}"
    
    # Check if package has tests
    if ! find "$package" -name "*_test.go" -type f | grep -q .; then
        echo -e "${YELLOW}No test files found in $package, skipping${NC}"
        return 0
    fi
    
    # Build test command
    local cmd="go test"
    
    if [[ "$VERBOSE" == "true" ]]; then
        cmd="$cmd -v"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        local coverage_file="coverage/${package_name}.cov"
        mkdir -p coverage
        cmd="$cmd -coverprofile=$coverage_file -covermode=atomic"
    fi
    
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd -parallel 4"
    fi
    
    cmd="$cmd -timeout $TIMEOUT $package"
    
    echo "Executing: $cmd"
    
    # Run the test
    if eval "$cmd"; then
        echo -e "${GREEN}✓ $package_name tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $package_name tests failed${NC}"
        return 1
    fi
}

run_specific_tests() {
    local pattern=$1
    echo -e "${GREEN}Running tests matching pattern: $pattern${NC}"
    
    # Only run on packages that have tests and compile properly
    local test_dirs=(
        "./testing/helpers"
        "./testing"
        "./dynamodb" 
        "./eventbridge"
        "./stepfunctions"
        "./snapshot"
        "./parallel"
        "./integration"
    )
    
    local cmd="go test"
    if [[ "$VERBOSE" == "true" ]]; then
        cmd="$cmd -v"
    fi
    cmd="$cmd -run $pattern -timeout $TIMEOUT ${test_dirs[*]}"
    
    if eval "$cmd"; then
        echo -e "${GREEN}✓ Pattern tests passed${NC}"
    else
        echo -e "${RED}✗ Pattern tests failed${NC}"
        exit 1
    fi
}

compile_all_packages() {
    echo -e "${GREEN}Compiling all packages...${NC}"
    
    # Main package
    if go build -v . >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Main package compiles${NC}"
    else
        echo -e "${RED}✗ Main package compilation failed${NC}"
        return 1
    fi
    
    # Test packages (excluding examples and lambda for now due to compatibility issues)
    local compile_packages=(
        "./testing/helpers"
        "./testing"
        "./dynamodb"
        "./eventbridge"
        "./stepfunctions"
        "./snapshot"
        "./parallel"
        "./integration"
    )
    
    for package in "${compile_packages[@]}"; do
        local package_name=$(basename "$package")
        if go build -v "$package" >/dev/null 2>&1; then
            echo -e "${GREEN}✓ $package_name compiles${NC}"
        else
            echo -e "${YELLOW}⚠ $package_name compilation issues${NC}"
        fi
    done
}

generate_coverage_report() {
    if [[ "$COVERAGE" != "true" ]]; then
        return 0
    fi
    
    echo -e "${GREEN}Generating coverage report...${NC}"
    
    # Merge coverage files
    if find coverage -name "*.cov" -type f | head -1 >/dev/null; then
        echo "mode: atomic" > coverage/merged.cov
        find coverage -name "*.cov" -type f -exec tail -n +2 {} \; >> coverage/merged.cov
        
        # Generate HTML report
        go tool cover -html=coverage/merged.cov -o coverage/report.html
        
        # Show coverage percentage
        local coverage_percent=$(go tool cover -func=coverage/merged.cov | tail -1 | awk '{print $3}')
        echo -e "${GREEN}Total coverage: $coverage_percent${NC}"
        echo -e "${GREEN}Coverage report: coverage/report.html${NC}"
    fi
}

clean_previous_runs() {
    echo -e "${YELLOW}Cleaning previous test artifacts...${NC}"
    rm -rf coverage/*.cov coverage/report.html 2>/dev/null || true
    go clean -testcache
}

show_help() {
    cat << EOF
VasDeference Test Runner

Usage: $0 [OPTIONS] [PATTERN]

OPTIONS:
    -h, --help       Show this help
    -v, --verbose    Enable verbose output
    -c, --coverage   Generate coverage report (default: true)
    -p, --parallel   Run tests in parallel (default: true)
    --no-coverage    Disable coverage reporting
    --no-parallel    Disable parallel execution
    --compile-only   Only compile, don't run tests
    --clean          Clean and exit

PATTERN:
    If provided, runs only tests matching the pattern

Examples:
    $0                          # Run all tests
    $0 -v                       # Run with verbose output
    $0 TestDynamoDB             # Run only DynamoDB tests
    $0 --compile-only           # Just check compilation
    $0 --clean                  # Clean artifacts

EOF
}

# Main execution
main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -p|--parallel)
                PARALLEL=true
                shift
                ;;
            --no-coverage)
                COVERAGE=false
                shift
                ;;
            --no-parallel)
                PARALLEL=false
                shift
                ;;
            --compile-only)
                print_header
                compile_all_packages
                exit $?
                ;;
            --clean)
                clean_previous_runs
                echo -e "${GREEN}Cleanup complete${NC}"
                exit 0
                ;;
            *)
                # Assume it's a test pattern
                print_header
                run_specific_tests "$1"
                exit $?
                ;;
        esac
    done
    
    print_header
    clean_previous_runs
    
    # Compile first
    echo -e "${GREEN}Step 1: Compilation Check${NC}"
    compile_all_packages
    print_separator
    
    # Run tests
    echo -e "${GREEN}Step 2: Running Test Suites${NC}"
    local failed_packages=()
    
    for package in "${TEST_PACKAGES[@]}"; do
        if ! run_package_tests "$package"; then
            failed_packages+=("$package")
        fi
        print_separator
    done
    
    # Generate coverage report
    if [[ "$COVERAGE" == "true" ]]; then
        echo -e "${GREEN}Step 3: Coverage Report${NC}"
        generate_coverage_report
        print_separator
    fi
    
    # Summary
    echo -e "${GREEN}Step 4: Test Summary${NC}"
    if [[ ${#failed_packages[@]} -eq 0 ]]; then
        echo -e "${GREEN}✅ All test packages passed!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Failed packages:${NC}"
        for package in "${failed_packages[@]}"; do
            echo -e "${RED}  - $package${NC}"
        done
        exit 1
    fi
}

# Execute main function
main "$@"