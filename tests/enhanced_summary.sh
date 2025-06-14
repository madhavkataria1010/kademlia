#!/bin/bash
# Enhanced Test Summary Generator with Top-Level Summary and Error Analysis

LOG_FILE="$1"

if [ $# -ne 1 ]; then
    echo "Usage: $0 <test-log-file>"
    exit 1
fi

if [ ! -f "$LOG_FILE" ]; then
    echo "Test log file not found: $LOG_FILE"
    exit 1
fi

# Create temporary files
TEMP_SUMMARY=$(mktemp)
TEMP_ORIGINAL=$(mktemp)

# Copy original log to temp file
cp "$LOG_FILE" "$TEMP_ORIGINAL"

# Extract only the detailed test logs section for analysis (ignore any existing summary)
DETAILED_LOGS_SECTION=$(grep -A 999999 "DETAILED TEST LOGS BEGIN BELOW" "$LOG_FILE" 2>/dev/null || cat "$LOG_FILE")
TEMP_LOG_FOR_ANALYSIS=$(mktemp)
echo "$DETAILED_LOGS_SECTION" > "$TEMP_LOG_FOR_ANALYSIS"

# If no detailed logs section found, use the whole file
if [ ! -s "$TEMP_LOG_FOR_ANALYSIS" ]; then
    cp "$LOG_FILE" "$TEMP_LOG_FOR_ANALYSIS"
fi

# Generate comprehensive summary
{
    echo "================================================================================"
    echo "🎯 COMPREHENSIVE TEST EXECUTION SUMMARY"
    echo "================================================================================"
    echo "Generated: $(date '+%Y-%m-%d %H:%M:%S %Z')"
    echo "Report File: $(basename "$LOG_FILE")"
    echo ""
    
    # Parse test results from different formats
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    SKIPPED_TESTS=0
    WARNING_COUNT=0
    COMPILE_ERRORS=0
    
    # Check for warnings first (analyze only the test logs, not summary)
    WARNING_COUNT=$(grep -ci "warning\|warn:\|\[⚠ WARNING\]" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
    
    # Check for compilation/setup failures
    if grep -q "setup failed\|build failed\|cannot find package" "$TEMP_LOG_FOR_ANALYSIS"; then
        COMPILE_ERRORS=1
        echo "🚨 COMPILATION/SETUP FAILURE DETECTED"
        echo "----------------------------------------"
        
        # Extract specific error messages
        echo "💥 Error Details:"
        grep -E "expected.*found|cannot find package|undefined:|setup failed|build failed" "$TEMP_LOG_FOR_ANALYSIS" | head -5 | sed 's/^/  • /'
        echo ""
        
        # Show failed packages
        if grep -q "FAIL.*\[setup failed\]" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "📦 Failed Packages/Modules:"
            grep "FAIL.*\[setup failed\]" "$TEMP_LOG_FOR_ANALYSIS" | sed 's/FAIL[[:space:]]*/  ❌ /' | sed 's/\[setup failed\]//'
            echo ""
        fi
        
        TOTAL_TESTS=0
        PASSED_TESTS=0
        FAILED_TESTS=1
        SKIPPED_TESTS=0
    else
        # Check for different test output formats (use clean log for analysis)
        if grep -q "✓\|✗\|⚠" "$TEMP_LOG_FOR_ANALYSIS"; then
            # gotestsum format
            PASSED_TESTS=$(grep -c "✓" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            FAILED_TESTS=$(grep "✗" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null | wc -l)
            SKIPPED_TESTS=$(grep "⚠" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null | wc -l)
            
            # Also check for custom framework error logging 
            FRAMEWORK_ERRORS=$(grep -c "\[✗ ERROR\]" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            if [ "$FRAMEWORK_ERRORS" -gt 0 ]; then
                FAILED_TESTS=$FRAMEWORK_ERRORS
            fi            # Also check for DONE line
            if grep -q "DONE" "$TEMP_LOG_FOR_ANALYSIS"; then
                DONE_LINE=$(grep "DONE" "$TEMP_LOG_FOR_ANALYSIS" | tail -1)
                if echo "$DONE_LINE" | grep -q "tests in"; then
                    TOTAL_FROM_DONE=$(echo "$DONE_LINE" | grep -o '[0-9]\+' | head -1)
                    if [ -n "$TOTAL_FROM_DONE" ] && [ "$TOTAL_FROM_DONE" -gt 0 ]; then
                        TOTAL_TESTS=$TOTAL_FROM_DONE
                    fi
                fi
            fi
        
        if [ "$TOTAL_TESTS" -eq 0 ]; then
            TOTAL_TESTS=$((PASSED_TESTS + FAILED_TESTS + SKIPPED_TESTS))
        fi
        
        elif grep -q "=== RUN\|--- PASS\|--- FAIL" "$TEMP_LOG_FOR_ANALYSIS"; then
            # Standard Go test format
            TOTAL_TESTS=$(grep -c "=== RUN" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            PASSED_TESTS=$(grep -c -- "--- PASS:" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            FAILED_TESTS=$(grep -- "--- FAIL:" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null | wc -l)
            SKIPPED_TESTS=$(grep -- "--- SKIP:" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null | wc -l)
            
            # Also check for custom framework error logging 
            FRAMEWORK_ERRORS=$(grep -c "\[✗ ERROR\]" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            if [ "$FRAMEWORK_ERRORS" -gt 0 ]; then
                FAILED_TESTS=$FRAMEWORK_ERRORS
            fi
        else
            # Fallback: try to find any pass/fail indicators
            PASSED_TESTS=$(grep -ci "pass\|success\|\[✓ SUCCESS\]" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            FAILED_TESTS=$(grep -ci "fail\|error\|\[✗ ERROR\]" "$TEMP_LOG_FOR_ANALYSIS" 2>/dev/null || echo 0)
            TOTAL_TESTS=$((PASSED_TESTS + FAILED_TESTS))
        fi
    fi
    
    echo "📊 TEST RESULTS OVERVIEW"
    echo "----------------------------------------"
    echo "Total Tests:     $TOTAL_TESTS"
    echo "✅ Passed:       $PASSED_TESTS"
    echo "❌ Failed:       $FAILED_TESTS"
    echo "⏭️  Skipped:      $SKIPPED_TESTS"
    echo "⚠️  Warnings:     $WARNING_COUNT"
    
    # Calculate success rate
    if [ "$TOTAL_TESTS" -gt 0 ]; then
        SUCCESS_RATE=$(echo "scale=2; $PASSED_TESTS * 100 / $TOTAL_TESTS" | bc -l 2>/dev/null || echo "0")
        echo "📈 Success Rate: ${SUCCESS_RATE}%"
    else
        if [ "$COMPILE_ERRORS" -eq 1 ]; then
            echo "📈 Success Rate: 0% (Compilation Failed)"
        else
            echo "📈 Success Rate: N/A (No Tests Found)"
        fi
    fi
    
    # Extract and display exact execution time (use original log for timing info)
    EXECUTION_TIME="N/A"
    if grep -q "ok.*github.com.*[0-9]\+\.[0-9]*s" "$TEMP_LOG_FOR_ANALYSIS"; then
        EXECUTION_TIME=$(grep "ok.*github.com.*[0-9]\+\.[0-9]*s" "$TEMP_LOG_FOR_ANALYSIS" | tail -1 | grep -o "[0-9]\+\.[0-9]*s")
    elif grep -q "ok.*github.com.*(cached)" "$TEMP_LOG_FOR_ANALYSIS"; then
        EXECUTION_TIME="cached (previous run)"
    elif grep -q "DONE.*in" "$TEMP_LOG_FOR_ANALYSIS"; then
        EXECUTION_TIME=$(grep "DONE.*in" "$TEMP_LOG_FOR_ANALYSIS" | tail -1 | grep -o "in [0-9]*\.[0-9]*s" | sed 's/in //')
    elif grep -q "ELAPSED.*[0-9]\+\.[0-9]*s" "$TEMP_LOG_FOR_ANALYSIS"; then
        EXECUTION_TIME=$(grep "ELAPSED.*[0-9]\+\.[0-9]*s" "$TEMP_LOG_FOR_ANALYSIS" | tail -1 | grep -o "[0-9]\+\.[0-9]*s")
    fi
    echo "⏱️  Execution Time: $EXECUTION_TIME"
    echo ""
    
    # Show warnings if any detected (analyze clean log)
    if [ "$WARNING_COUNT" -gt 0 ]; then
        echo "⚠️  WARNINGS DETECTED"
        echo "----------------------------------------"
        echo "Number of Warnings: $WARNING_COUNT"
        echo ""
        echo "📋 Warning Details:"
        grep -i "warning\|warn:\|\[⚠ WARNING\]" "$TEMP_LOG_FOR_ANALYSIS" | head -10 | sed 's/^/  ⚠️  /' | sed 's/[[:space:]]*$//'
        echo ""
    fi
    
    # Detailed error analysis for failed tests
    if [ "$FAILED_TESTS" -gt 0 ] || [ "$COMPILE_ERRORS" -eq 1 ]; then
        echo "🔍 DETAILED FAILURE ANALYSIS"
        echo "----------------------------------------"
        echo "Number of Failed Tests: $FAILED_TESTS"
        if [ "$COMPILE_ERRORS" -eq 1 ]; then
            echo "Compilation/Setup Errors: Yes"
        fi
        echo ""
        
        # Extract failed test names and details with exact error messages (analyze clean log)
        echo "📋 Failed Test Cases with Exact Errors:"
        
        # Check for custom framework format first
        if grep -q "\[✗ ERROR\]" "$TEMP_LOG_FOR_ANALYSIS"; then
            # Custom framework format: logger.go:74: [timestamp] MODULE [✗ ERROR] TestName: error message
            echo "  Custom framework errors detected:"
            while IFS= read -r line; do
                test_name=$(echo "$line" | sed 's/.*\[✗ ERROR\] //' | sed 's/: .*//')
                error_msg=$(echo "$line" | sed 's/.*\[✗ ERROR\] [^:]*: //')
                echo "  ❌ Test: $test_name"
                echo "    💥 $error_msg"
                echo ""
            done < <(grep "\[✗ ERROR\]" "$TEMP_LOG_FOR_ANALYSIS" | head -10)
            
        elif grep -q -- "--- FAIL:" "$TEMP_LOG_FOR_ANALYSIS"; then
            # Standard Go test format failures - get exact test names and errors
            while IFS= read -r fail_line; do
                test_name=$(echo "$fail_line" | sed 's/--- FAIL: //' | awk '{print $1}' | sed 's/([^)]*)$//')
                echo "  ❌ Test: $test_name"
                
                # Get the error details for this specific test by looking ahead
                test_section=$(grep -A 15 "^$fail_line" "$TEMP_LOG_FOR_ANALYSIS" | head -15)
                error_details=$(echo "$test_section" | grep -E "Error|panic|assertion|expected|actual|got|FAIL:" | head -3)
                if [ -n "$error_details" ]; then
                    echo "$error_details" | sed 's/^[[:space:]]*/    💥 /'
                else
                    echo "    💥 (Error details not captured in standard format)"
                fi
                echo ""
            done < <(grep -- "--- FAIL:" "$TEMP_LOG_FOR_ANALYSIS" | head -10)
            
        elif grep -q "✗" "$TEMP_LOG_FOR_ANALYSIS"; then
            # gotestsum format failures
            while IFS= read -r fail_line; do
                test_name=$(echo "$fail_line" | sed 's/.*✗ //' | awk '{print $1}')
                echo "  ❌ Test: $test_name"
                
                # Look for error context around this failure
                line_num=$(grep -n "$fail_line" "$TEMP_LOG_FOR_ANALYSIS" | head -1 | cut -d: -f1)
                if [ -n "$line_num" ]; then
                    error_context=$(sed -n "$((line_num-2)),$((line_num+5))p" "$TEMP_LOG_FOR_ANALYSIS" | grep -E "Error|panic|assertion|expected|actual|FAIL" | head -3)
                    if [ -n "$error_context" ]; then
                        echo "$error_context" | sed 's/^[[:space:]]*/    💥 /'
                    else
                        echo "    💥 (Error details not captured in gotestsum format)"
                    fi
                fi
                echo ""
            done < <(grep "✗" "$TEMP_LOG_FOR_ANALYSIS" | head -10)
            
        else
            # Try to find any failed tests by looking for FAIL lines and framework errors
            failed_indicators=$(grep -E "FAIL|ERROR.*Test|Test.*ERROR" "$TEMP_LOG_FOR_ANALYSIS" | head -10)
            if [ -n "$failed_indicators" ]; then
                echo "$failed_indicators" | while IFS= read -r line; do
                    echo "  ❌ $line"
                done
            else
                echo "  ❌ Failed tests detected but specific names not parsed"
            fi
            echo ""
        fi
        
        # Common failure patterns analysis (analyze clean log)
        echo "🔬 Failure Pattern Analysis:"
        
        # Check for compilation errors
        if grep -q "build failed\|cannot find package\|undefined:" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "  🚫 COMPILATION ERRORS detected"
            echo "     - Likely cause: Missing dependencies or syntax errors"
            echo "     - Action: Run 'go mod tidy' and check import statements"
        fi
        
        # Check for import/module issues
        if grep -q "no required module provides package\|module.*not found" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "  📦 MODULE DEPENDENCY ISSUES detected"
            echo "     - Likely cause: Incorrect import paths or missing modules"
            echo "     - Action: Verify go.mod file and import paths"
        fi
        
        # Check for test setup issues
        if grep -q "nil pointer\|panic:" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "  ⚡ RUNTIME PANIC detected"
            echo "     - Likely cause: Uninitialized variables or nil pointer access"
            echo "     - Action: Check test setup and variable initialization"
        fi
        
        # Check for assertion failures
        if grep -q "assertion failed\|expected.*got" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "  🎯 ASSERTION FAILURES detected"
            echo "     - Likely cause: Logic errors or incorrect test expectations"
            echo "     - Action: Review test logic and expected vs actual values"
        fi
        
        # Check for timeout issues
        if grep -q "timeout\|deadline exceeded" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "  ⏰ TIMEOUT ISSUES detected"
            echo "     - Likely cause: Slow operations or infinite loops"
            echo "     - Action: Optimize code or increase test timeout"
        fi
        
        echo ""
    fi
    
    # Coverage analysis if available (analyze clean log)
    if grep -q "coverage:" "$TEMP_LOG_FOR_ANALYSIS"; then
        echo "📊 CODE COVERAGE ANALYSIS"
        echo "----------------------------------------"
        COVERAGE=$(grep "coverage:" "$TEMP_LOG_FOR_ANALYSIS" | tail -1 | grep -o '[0-9]*\.[0-9]*%' | head -1)
        if [ -n "$COVERAGE" ]; then
            echo "Overall Coverage: $COVERAGE"
            
            # Coverage assessment
            COVERAGE_NUM=$(echo "$COVERAGE" | grep -o '[0-9]*\.[0-9]*' | head -1)
            if [ -n "$COVERAGE_NUM" ]; then
                if awk "BEGIN {exit !($COVERAGE_NUM >= 80)}" 2>/dev/null; then
                    echo "✅ Excellent coverage (≥80%)"
                elif awk "BEGIN {exit !($COVERAGE_NUM >= 60)}" 2>/dev/null; then
                    echo "⚠️  Good coverage but room for improvement (60-79%)"
                else
                    echo "❌ Low coverage - needs attention (<60%)"
                fi
            fi
        else
            echo "Coverage data found but percentage not extractable"
        fi
        echo ""
    fi
    
    # Benchmark results if available (analyze clean log)
    if grep -q "Benchmark.*ns/op\|allocs/op" "$TEMP_LOG_FOR_ANALYSIS"; then
        echo "🏁 BENCHMARK RESULTS SUMMARY"
        echo "----------------------------------------"
        echo "Benchmark tests detected in log:"
        grep -E "Benchmark.*ns/op" "$TEMP_LOG_FOR_ANALYSIS" | head -5 | while read -r bench_line; do
            bench_name=$(echo "$bench_line" | awk '{print $1}')
            ns_per_op=$(echo "$bench_line" | awk '{print $3}')
            echo "  🏃 $bench_name: $ns_per_op ns/op"
        done
        echo ""
    fi
    
    # Execution time analysis (analyze clean log)
    if grep -q "DONE.*in\|ok.*[0-9]s" "$TEMP_LOG_FOR_ANALYSIS"; then
        echo "⏱️  EXECUTION TIME ANALYSIS"
        echo "----------------------------------------"
        if grep -q "DONE.*in" "$TEMP_LOG_FOR_ANALYSIS"; then
            TIME_INFO=$(grep "DONE.*in" "$TEMP_LOG_FOR_ANALYSIS" | tail -1)
            echo "Total execution: $TIME_INFO"
        fi
        
        # Check for slow tests
        if grep -q "SLOW" "$TEMP_LOG_FOR_ANALYSIS" || grep -E "\([1-9][0-9]\.[0-9]+s\)" "$TEMP_LOG_FOR_ANALYSIS"; then
            echo "⚠️  Slow tests detected - consider optimization"
        fi
        echo ""
    fi
    
    # Overall assessment and recommendations
    echo "🎯 FINAL ASSESSMENT & RECOMMENDATIONS"
    echo "----------------------------------------"
    
    if [ "$FAILED_TESTS" -eq 0 ] && [ "$TOTAL_TESTS" -gt 0 ]; then
        echo "✅ EXCELLENT: All tests passed successfully!"
        if [ "$WARNING_COUNT" -gt 0 ]; then
            echo "   ⚠️  Note: $WARNING_COUNT warning(s) detected (see details above)"
        fi
        echo ""
        echo "📋 Recommendations:"
        echo "  • Maintain current code quality"
        echo "  • Consider adding more edge case tests"
        if [ "$WARNING_COUNT" -gt 0 ]; then
            echo "  • Address the $WARNING_COUNT warning(s) to improve code quality"
        fi
        if [ -n "$COVERAGE" ]; then
            COVERAGE_NUM=$(echo "$COVERAGE" | grep -o '[0-9]*\.[0-9]*' | head -1)
            if [ -n "$COVERAGE_NUM" ]; then
                if ! awk "BEGIN {exit !($COVERAGE_NUM >= 80)}" 2>/dev/null; then
                    echo "  • Improve test coverage to 80%+"
                fi
            fi
        fi
    elif [ "$FAILED_TESTS" -gt 0 ] || [ "$COMPILE_ERRORS" -eq 1 ]; then
        echo "❌ ATTENTION REQUIRED: $FAILED_TESTS test(s) failed"
        if [ "$COMPILE_ERRORS" -eq 1 ]; then
            echo "   🚨 Additional Issue: Compilation/Setup errors detected"
        fi
        echo ""
        echo "🚨 IMMEDIATE ACTIONS NEEDED:"
        if [ "$COMPILE_ERRORS" -eq 1 ]; then
            echo "  1. Fix compilation errors (see error details above)"
            echo "  2. Run 'go mod tidy' to resolve dependencies"
            echo "  3. Check import paths and package declarations"
            echo "  4. Verify all required files exist and are properly formatted"
        fi
        echo "  1. Review failed test details above"
        echo "  2. Fix identified issues (compilation, logic, setup)"
        echo "  3. Re-run tests to verify fixes"
        echo "  4. Consider adding regression tests"
        echo ""
        echo "📋 Prevention Strategies:"
        echo "  • Implement pre-commit hooks"
        echo "  • Add more comprehensive error handling"
        echo "  • Increase test coverage for edge cases"
        if [ "$WARNING_COUNT" -gt 0 ]; then
            echo "  • Address the $WARNING_COUNT warning(s) detected"
        fi
    elif [ "$TOTAL_TESTS" -eq 0 ]; then
        echo "⚠️  WARNING: No tests detected or executed"
        echo ""
        echo "🔧 Possible Issues:"
        echo "  • Test files not properly named (*_test.go)"
        echo "  • Compilation errors preventing test execution"
        echo "  • Incorrect test directory structure"
        echo "  • Missing test functions (must start with 'Test')"
    fi
    
    echo ""
    echo "================================================================================"
    echo "📝 DETAILED TEST LOGS BEGIN BELOW"
    echo "================================================================================"
    echo ""
    
} > "$TEMP_SUMMARY"

# Combine summary with original logs
cat "$TEMP_SUMMARY" > "$LOG_FILE"
cat "$TEMP_ORIGINAL" >> "$LOG_FILE"

# Cleanup
rm -f "$TEMP_SUMMARY" "$TEMP_ORIGINAL" "$TEMP_LOG_FOR_ANALYSIS"

echo "Enhanced summary with error analysis added to top of $LOG_FILE"
