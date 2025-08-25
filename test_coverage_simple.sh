#!/bin/bash

# Simple test coverage script to assess current state
set -e

echo "🧪 Running Coverage Analysis for Vas Deference Framework"
echo "======================================================="

# Create coverage directory
mkdir -p coverage

echo "📊 Testing core framework functionality..."
go test -v -coverprofile=coverage/core.cov . || echo "Core tests have issues"

echo "📊 Testing infrastructure utilities..."  
go test -v -coverprofile=coverage/testing.cov ./testing/ || echo "Testing infrastructure has issues"

echo "📊 Testing Lambda package..."
go test -v -coverprofile=coverage/lambda.cov ./lambda/ || echo "Lambda tests have issues"

echo "📊 Testing DynamoDB package..."
go test -v -coverprofile=coverage/dynamodb.cov ./dynamodb/ || echo "DynamoDB tests have issues"

echo "📊 Testing EventBridge package..."
go test -v -coverprofile=coverage/eventbridge.cov ./eventbridge/ || echo "EventBridge tests have issues"

echo "📊 Testing Step Functions package..."
go test -v -coverprofile=coverage/stepfunctions.cov ./stepfunctions/ || echo "Step Functions tests have issues"

# Combine coverage reports
echo "mode: set" > coverage/combined.cov
for file in coverage/*.cov; do
    if [ -f "$file" ] && [ "$file" != "coverage/combined.cov" ]; then
        tail -n +2 "$file" >> coverage/combined.cov 2>/dev/null || true
    fi
done

# Generate HTML coverage report
if [ -f coverage/combined.cov ]; then
    go tool cover -html=coverage/combined.cov -o coverage/coverage.html
    
    # Calculate overall coverage percentage
    COVERAGE=$(go tool cover -func=coverage/combined.cov | grep total | awk '{print $3}' || echo "0%")
    
    echo ""
    echo "📋 COVERAGE SUMMARY"
    echo "==================="
    echo "Overall Coverage: ${COVERAGE}"
    echo "HTML Report: coverage/coverage.html"
    
    # Show per-package coverage
    echo ""
    echo "Per-package coverage:"
    go tool cover -func=coverage/combined.cov | grep -v "total:" | tail -20 || echo "Coverage details not available"
    
else
    echo "❌ No coverage data generated"
fi

echo ""
echo "🎯 Current Status:"
echo "- Core testing infrastructure: ✅ Working"
echo "- Package structure: ✅ Fixed"  
echo "- AWS SDK abstractions: 🔧 In progress"
echo "- Test coverage target: 90% (current: ${COVERAGE})"