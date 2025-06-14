package testutils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TestSummary represents comprehensive test execution statistics
type TestSummary struct {
	TestType         string
	ExecutionTime    time.Time
	Duration         time.Duration
	TotalTests       int
	PassedTests      int
	FailedTests      int
	SkippedTests     int
	SuccessRate      float64
	CoveragePercent  float64
	BenchmarkResults []BenchmarkResult
	ErrorSummary     []string
	PackageCoverage  map[string]float64
}

// BenchmarkResult represents benchmark test results
type BenchmarkResult struct {
	Name          string
	Iterations    int
	NanosPerOp    float64
	BytesPerOp    int
	AllocsPerOp   int
	MemBytesPerOp int
}

// parseTestOutput parses Go test output and extracts comprehensive statistics
func parseTestOutput(filePath string, testType string) (*TestSummary, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	summary := &TestSummary{
		TestType:        testType,
		ExecutionTime:   time.Now(),
		PackageCoverage: make(map[string]float64),
	}

	scanner := bufio.NewScanner(file)
	var startTime, endTime time.Time
	var totalTests, passedTests, failedTests, skippedTests int
	var benchmarkResults []BenchmarkResult
	var errorSummary []string
	var coveragePercent float64

	// Regular expressions for parsing test output
	testResultRegex := regexp.MustCompile(`^(PASS|FAIL|SKIP):\s+(\S+)`)
	benchmarkRegex := regexp.MustCompile(`^Benchmark(\w+)\s+(\d+)\s+(\d+\.?\d*)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)
	coverageRegex := regexp.MustCompile(`coverage:\s+(\d+\.?\d*)%`)
	packageCoverageRegex := regexp.MustCompile(`(\S+)\s+(\d+\.?\d*)%`)
	errorRegex := regexp.MustCompile(`^\s*--- FAIL:|panic:|Error:|FAIL`)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse test results
		if matches := testResultRegex.FindStringSubmatch(line); matches != nil {
			totalTests++
			switch matches[1] {
			case "PASS":
				passedTests++
			case "FAIL":
				failedTests++
			case "SKIP":
				skippedTests++
			}
		}

		// Parse benchmark results
		if matches := benchmarkRegex.FindStringSubmatch(line); matches != nil {
			iterations, _ := strconv.Atoi(matches[2])
			nanosPerOp, _ := strconv.ParseFloat(matches[3], 64)

			result := BenchmarkResult{
				Name:       matches[1],
				Iterations: iterations,
				NanosPerOp: nanosPerOp,
			}

			if len(matches) > 4 && matches[4] != "" {
				result.BytesPerOp, _ = strconv.Atoi(matches[4])
			}
			if len(matches) > 5 && matches[5] != "" {
				result.AllocsPerOp, _ = strconv.Atoi(matches[5])
			}

			benchmarkResults = append(benchmarkResults, result)
		}

		// Parse coverage information
		if matches := coverageRegex.FindStringSubmatch(line); matches != nil {
			coveragePercent, _ = strconv.ParseFloat(matches[1], 64)
		}

		// Parse package-specific coverage
		if matches := packageCoverageRegex.FindStringSubmatch(line); matches != nil && strings.Contains(line, "%") {
			packageName := matches[1]
			coverage, _ := strconv.ParseFloat(matches[2], 64)
			summary.PackageCoverage[packageName] = coverage
		}

		// Collect error information
		if errorRegex.MatchString(line) {
			errorSummary = append(errorSummary, strings.TrimSpace(line))
		}

		// Try to parse timestamps (this is approximate)
		if strings.Contains(line, "=== RUN") && startTime.IsZero() {
			startTime = time.Now()
		}
		if strings.Contains(line, "PASS:") || strings.Contains(line, "FAIL:") {
			endTime = time.Now()
		}
	}

	// Calculate success rate
	if totalTests > 0 {
		summary.SuccessRate = float64(passedTests) / float64(totalTests) * 100
	}

	// Set duration
	if !endTime.IsZero() && !startTime.IsZero() {
		summary.Duration = endTime.Sub(startTime)
	}

	// Populate summary
	summary.TotalTests = totalTests
	summary.PassedTests = passedTests
	summary.FailedTests = failedTests
	summary.SkippedTests = skippedTests
	summary.CoveragePercent = coveragePercent
	summary.BenchmarkResults = benchmarkResults
	summary.ErrorSummary = errorSummary

	return summary, nil
}

// generateSummaryReport creates a formatted summary report
func (s *TestSummary) generateSummaryReport() string {
	var report strings.Builder

	report.WriteString("\n")
	report.WriteString(strings.Repeat("=", 80) + "\n")
	report.WriteString(fmt.Sprintf("COMPREHENSIVE TEST SUMMARY - %s TESTS\n", strings.ToUpper(s.TestType)))
	report.WriteString(strings.Repeat("=", 80) + "\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n", s.ExecutionTime.Format("2006-01-02 15:04:05 MST")))

	if s.Duration != 0 {
		report.WriteString(fmt.Sprintf("Duration: %v\n", s.Duration))
	}
	report.WriteString("\n")

	// Test Results Section
	report.WriteString("📊 TEST RESULTS OVERVIEW\n")
	report.WriteString(strings.Repeat("-", 40) + "\n")
	report.WriteString(fmt.Sprintf("Total Tests:     %d\n", s.TotalTests))
	report.WriteString(fmt.Sprintf("✅ Passed:       %d\n", s.PassedTests))
	report.WriteString(fmt.Sprintf("❌ Failed:       %d\n", s.FailedTests))
	report.WriteString(fmt.Sprintf("⏭️  Skipped:      %d\n", s.SkippedTests))
	report.WriteString(fmt.Sprintf("📈 Success Rate: %.2f%%\n", s.SuccessRate))
	if s.Duration != 0 {
		report.WriteString(fmt.Sprintf("⏱️ Duration:     %v\n", s.Duration))
	}
	report.WriteString("\n")

	// Coverage Section (if available)
	if s.CoveragePercent > 0 {
		report.WriteString("📊 CODE COVERAGE ANALYSIS\n")
		report.WriteString(strings.Repeat("-", 40) + "\n")
		report.WriteString(fmt.Sprintf("Overall Coverage: %.2f%%\n", s.CoveragePercent))

		if len(s.PackageCoverage) > 0 {
			report.WriteString("\nPackage Coverage Breakdown:\n")
			for pkg, coverage := range s.PackageCoverage {
				status := "🟢"
				if coverage < 70 {
					status = "🟡"
				}
				if coverage < 50 {
					status = "🔴"
				}
				report.WriteString(fmt.Sprintf("  %s %s: %.2f%%\n", status, pkg, coverage))
			}
		}
		report.WriteString("\n")
	}

	// Benchmark Results Section (if available)
	if len(s.BenchmarkResults) > 0 {
		report.WriteString("🏁 BENCHMARK RESULTS\n")
		report.WriteString(strings.Repeat("-", 40) + "\n")
		report.WriteString("Test Name                    Iterations    ns/op       B/op    allocs/op\n")
		report.WriteString(strings.Repeat("-", 70) + "\n")

		for _, bench := range s.BenchmarkResults {
			report.WriteString(fmt.Sprintf("%-28s %10d %10.2f %8d %10d\n",
				bench.Name, bench.Iterations, bench.NanosPerOp,
				bench.BytesPerOp, bench.AllocsPerOp))
		}
		report.WriteString("\n")
	}

	// Error Summary Section (if there are errors)
	if len(s.ErrorSummary) > 0 {
		report.WriteString("❌ ERROR SUMMARY\n")
		report.WriteString(strings.Repeat("-", 40) + "\n")
		for i, err := range s.ErrorSummary {
			if i >= 10 { // Limit to first 10 errors
				report.WriteString(fmt.Sprintf("... and %d more errors\n", len(s.ErrorSummary)-10))
				break
			}
			report.WriteString(fmt.Sprintf("%d. %s\n", i+1, err))
		}
		report.WriteString("\n")
	}

	// Final Assessment
	report.WriteString("🎯 FINAL ASSESSMENT\n")
	report.WriteString(strings.Repeat("-", 40) + "\n")

	if s.FailedTests == 0 {
		report.WriteString("✅ ALL TESTS PASSED - EXCELLENT!\n")
		if s.CoveragePercent >= 80 {
			report.WriteString("🎉 High code coverage achieved!\n")
		} else if s.CoveragePercent >= 60 {
			report.WriteString("📊 Good code coverage, consider improving further\n")
		} else if s.CoveragePercent > 0 {
			report.WriteString("⚠️  Low code coverage, consider adding more tests\n")
		}
	} else {
		report.WriteString("❌ SOME TESTS FAILED - ATTENTION REQUIRED\n")
		report.WriteString("🔧 Review failed tests and fix issues before deployment\n")
	}

	// Recommendations
	report.WriteString("\n📋 RECOMMENDATIONS\n")
	report.WriteString(strings.Repeat("-", 40) + "\n")

	if s.FailedTests > 0 {
		report.WriteString("• Fix failing tests immediately\n")
	}
	if s.CoveragePercent > 0 && s.CoveragePercent < 80 {
		report.WriteString("• Increase test coverage to 80%+ for better reliability\n")
	}
	if len(s.BenchmarkResults) > 0 {
		report.WriteString("• Monitor benchmark performance for regressions\n")
	}
	if s.SkippedTests > 0 {
		report.WriteString("• Review and enable skipped tests if possible\n")
	}

	report.WriteString("\n")
	report.WriteString(strings.Repeat("=", 80) + "\n")
	report.WriteString("END OF SUMMARY REPORT\n")
	report.WriteString(strings.Repeat("=", 80) + "\n\n")

	return report.String()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <log-file> <test-type>\n", os.Args[0])
		os.Exit(1)
	}

	logFile := os.Args[1]
	testType := os.Args[2]

	summary, err := parseTestOutput(logFile, testType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing test output: %v\n", err)
		os.Exit(1)
	}

	summaryReport := summary.generateSummaryReport()
	fmt.Print(summaryReport)
}
