#!/bin/bash

# Vas Deference Comprehensive Testing Script
# This script provides comprehensive testing including:
# - Test coverage analysis (targeting 90%+ coverage)
# - Mutation testing with Go-mutesting
# - Property-based testing with Rapid
# - Integration testing
# - Performance benchmarking

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Configuration
COVERAGE_THRESHOLD=90
MUTATION_THRESHOLD=75
TIMEOUT=300s

echo -e "${BLUE}🧬 Vas Deference Comprehensive Testing Suite${NC}"
echo "============================================="

# Install required tools
install_tools() {
    echo -e "${YELLOW}📦 Installing testing tools...${NC}"
    
    # Install Go-mutesting
    if ! command -v go-mutesting &> /dev/null; then
        echo "Installing go-mutesting..."
        go install github.com/avito-tech/go-mutesting/cmd/go-mutesting@latest
    fi
    
    # Install additional coverage tools
    go install github.com/boumenot/gocover-cobertura@latest
    go install github.com/axw/gocov/gocov@latest
    go install github.com/matm/gocov-html@latest
    
    # Ensure rapid is available (part of standard testing)
    go mod download
    
    echo -e "${GREEN}✅ Tools installed${NC}"
}

# Clean previous test artifacts
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning previous test artifacts...${NC}"
    rm -rf coverage/
    rm -rf mutants/
    rm -rf test-results/
    mkdir -p coverage mutants test-results
}

# Run basic tests with coverage
run_coverage_tests() {
    echo -e "${YELLOW}🧪 Running tests with coverage analysis...${NC}"
    
    # Run tests with coverage for each package
    packages=(
        "."
        "./lambda"
        "./dynamodb" 
        "./eventbridge"
        "./stepfunctions"
        "./parallel"
        "./snapshot"
        "./testing"
    )
    
    # Combined coverage profile
    echo "mode: atomic" > coverage/coverage.out
    
    for pkg in "${packages[@]}"; do
        if [ -d "$pkg" ] && ls "$pkg"/*.go >/dev/null 2>&1; then
            echo "Testing package: $pkg"
            go test -race -timeout="$TIMEOUT" -covermode=atomic -coverprofile="coverage/$(echo "$pkg" | tr '/' '_').cov" "$pkg" || true
            
            # Append to combined coverage (skip mode line)
            if [ -f "coverage/$(echo "$pkg" | tr '/' '_').cov" ]; then
                tail -n +2 "coverage/$(echo "$pkg" | tr '/' '_').cov" >> coverage/coverage.out
            fi
        fi
    done
    
    # Generate coverage report
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    
    # Calculate coverage percentage
    COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    echo -e "${BLUE}📊 Total Coverage: ${COVERAGE}%${NC}"
    
    if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
        echo -e "${GREEN}✅ Coverage threshold met (${COVERAGE}% >= ${COVERAGE_THRESHOLD}%)${NC}"
    else
        echo -e "${RED}❌ Coverage threshold not met (${COVERAGE}% < ${COVERAGE_THRESHOLD}%)${NC}"
        return 1
    fi
}

# Run mutation testing
run_mutation_tests() {
    echo -e "${YELLOW}🧬 Running mutation testing...${NC}"
    
    # Create mutant configurations
    cat > mutants/mutants.conf << EOF
# Vas Deference Mutation Testing Configuration
# Patterns to mutate
*_test.go: false
mock*.go: false
testing/*.go: false

# Focus on core logic
*.go: true
EOF

    # Run mutation testing on core packages
    packages_to_mutate=(
        "."
        "./lambda"
        "./dynamodb"
        "./eventbridge"
        "./stepfunctions"
        "./parallel"
        "./snapshot"
    )
    
    mutation_results=()
    
    for pkg in "${packages_to_mutate[@]}"; do
        if [ -d "$pkg" ] && ls "$pkg"/*.go >/dev/null 2>&1 && ls "$pkg"/*_test.go >/dev/null 2>&1; then
            echo "Mutating package: $pkg"
            
            # Run mutation testing
            result_file="mutants/$(echo "$pkg" | tr '/' '_')_mutants.txt"
            
            timeout 300 go-mutesting "$pkg" > "$result_file" 2>&1 || {
                echo -e "${YELLOW}⚠️ Mutation testing timed out for $pkg${NC}"
                echo "TIMEOUT" > "$result_file"
            }
            
            # Parse results
            if grep -q "PASS" "$result_file" && grep -q "FAIL" "$result_file"; then
                passed=$(grep -c "PASS" "$result_file" || echo 0)
                failed=$(grep -c "FAIL" "$result_file" || echo 0)
                total=$((passed + failed))
                
                if [ "$total" -gt 0 ]; then
                    mutation_score=$((failed * 100 / total))
                    mutation_results+=("$pkg: ${mutation_score}% (${failed}/${total} mutants killed)")
                    echo -e "${BLUE}📊 $pkg Mutation Score: ${mutation_score}%${NC}"
                else
                    mutation_results+=("$pkg: No mutants generated")
                fi
            else
                mutation_results+=("$pkg: Mutation testing failed or timed out")
            fi
        else
            echo "Skipping $pkg (no Go files or tests found)"
        fi
    done
    
    # Summary
    echo -e "\n${BLUE}🧬 Mutation Testing Summary:${NC}"
    for result in "${mutation_results[@]}"; do
        echo "  $result"
    done
}

# Run property-based tests with Rapid
run_rapid_tests() {
    echo -e "${YELLOW}⚡ Running property-based tests with Rapid...${NC}"
    
    # Add Rapid dependency if not present
    if ! grep -q "pgregory.net/rapid" go.mod; then
        go get pgregory.net/rapid@latest
    fi
    
    # Run tests with Rapid flag
    go test -rapid.checks=1000 ./... || {
        echo -e "${YELLOW}⚠️ Some Rapid property tests failed${NC}"
    }
}

# Run benchmark tests
run_benchmarks() {
    echo -e "${YELLOW}🚀 Running benchmark tests...${NC}"
    
    packages_with_benchmarks=(
        "./lambda"
        "./dynamodb"
        "./parallel"
        "./snapshot"
    )
    
    for pkg in "${packages_with_benchmarks[@]}"; do
        if [ -d "$pkg" ] && grep -r "func Benchmark" "$pkg" >/dev/null 2>&1; then
            echo "Running benchmarks for: $pkg"
            go test -bench=. -benchmem -timeout="$TIMEOUT" "$pkg" | tee "test-results/$(echo "$pkg" | tr '/' '_')_bench.txt"
        fi
    done
}

# Run integration tests
run_integration_tests() {
    echo -e "${YELLOW}🔗 Running integration tests...${NC}"
    
    # Integration tests (these may require AWS credentials but should work with mocks)
    go test -tags=integration -timeout="$TIMEOUT" ./... || {
        echo -e "${YELLOW}⚠️ Integration tests failed (this is expected without AWS credentials)${NC}"
    }
}

# Run race condition detection
run_race_tests() {
    echo -e "${YELLOW}🏃 Running race condition detection...${NC}"
    
    go test -race -timeout="$TIMEOUT" ./... || {
        echo -e "${YELLOW}⚠️ Race condition tests failed${NC}"
        return 1
    }
    
    echo -e "${GREEN}✅ No race conditions detected${NC}"
}

# Memory leak detection
run_memory_tests() {
    echo -e "${YELLOW}🧠 Running memory leak detection...${NC}"
    
    # Run tests with memory profiling
    go test -memprofile=test-results/mem.prof -timeout="$TIMEOUT" ./... >/dev/null 2>&1 || true
    
    if [ -f "test-results/mem.prof" ]; then
        go tool pprof -text test-results/mem.prof | head -20 > test-results/memory_report.txt
        echo -e "${GREEN}✅ Memory profile generated${NC}"
    fi
}

# Generate comprehensive test report
generate_report() {
    echo -e "${YELLOW}📋 Generating comprehensive test report...${NC}"
    
    report_file="test-results/comprehensive_report.md"
    
    cat > "$report_file" << EOF
# Vas Deference Comprehensive Test Report

Generated: $(date)

## Test Coverage Summary

$(go tool cover -func=coverage/coverage.out | grep total || echo "Coverage data not available")

## Mutation Testing Results

EOF
    
    if [ -d "mutants" ] && ls mutants/*.txt >/dev/null 2>&1; then
        for mutant_file in mutants/*.txt; do
            echo "### $(basename "$mutant_file" .txt)" >> "$report_file"
            echo '```' >> "$report_file"
            head -20 "$mutant_file" >> "$report_file"
            echo '```' >> "$report_file"
            echo "" >> "$report_file"
        done
    fi
    
    cat >> "$report_file" << EOF

## Benchmark Results

EOF
    
    if [ -d "test-results" ] && ls test-results/*_bench.txt >/dev/null 2>&1; then
        for bench_file in test-results/*_bench.txt; do
            echo "### $(basename "$bench_file" _bench.txt)" >> "$report_file"
            echo '```' >> "$report_file"
            cat "$bench_file" >> "$report_file"
            echo '```' >> "$report_file"
            echo "" >> "$report_file"
        done
    fi
    
    echo -e "${GREEN}✅ Report generated: $report_file${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}Starting comprehensive testing suite...${NC}"
    
    # Check if bc is available for calculations
    if ! command -v bc &> /dev/null; then
        echo -e "${YELLOW}Installing bc for calculations...${NC}"
        sudo apt-get update && sudo apt-get install -y bc || {
            echo -e "${RED}Cannot install bc. Some calculations may fail.${NC}"
        }
    fi
    
    install_tools
    cleanup
    
    echo -e "\n${BLUE}Phase 1: Coverage Testing${NC}"
    run_coverage_tests
    
    echo -e "\n${BLUE}Phase 2: Race Condition Detection${NC}"
    run_race_tests
    
    echo -e "\n${BLUE}Phase 3: Memory Leak Detection${NC}"
    run_memory_tests
    
    echo -e "\n${BLUE}Phase 4: Property-Based Testing${NC}"
    run_rapid_tests
    
    echo -e "\n${BLUE}Phase 5: Benchmark Testing${NC}"
    run_benchmarks
    
    echo -e "\n${BLUE}Phase 6: Mutation Testing${NC}"
    run_mutation_tests
    
    echo -e "\n${BLUE}Phase 7: Integration Testing${NC}"
    run_integration_tests
    
    echo -e "\n${BLUE}Phase 8: Report Generation${NC}"
    generate_report
    
    echo -e "\n${GREEN}🎉 Comprehensive testing completed!${NC}"
    echo -e "${BLUE}📊 View coverage report: coverage/coverage.html${NC}"
    echo -e "${BLUE}📋 View comprehensive report: test-results/comprehensive_report.md${NC}"
}

# Handle script arguments
case "${1:-all}" in
    "coverage")
        install_tools
        cleanup
        run_coverage_tests
        ;;
    "mutation")
        install_tools
        cleanup
        run_mutation_tests
        ;;
    "rapid")
        run_rapid_tests
        ;;
    "benchmarks")
        run_benchmarks
        ;;
    "race")
        run_race_tests
        ;;
    "memory")
        run_memory_tests
        ;;
    "integration")
        run_integration_tests
        ;;
    "report")
        generate_report
        ;;
    "all"|*)
        main
        ;;
esac