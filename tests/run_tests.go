package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TestRunner manages and executes the complete test suite
type TestRunner struct {
	projectRoot string
	verbose     bool
	coverage    bool
	benchmark   bool
	parallel    bool
	pattern     string
	timestamp   string
	reportsDir  string
}

// TestSuite represents a group of related tests
type TestSuite struct {
	Name        string
	Path        string
	Description string
	Timeout     time.Duration
}

func main() {
	runner := &TestRunner{
		timestamp:  time.Now().Format("2006-01-02_15-04-05"),
		reportsDir: "reports", // Changed from "../reports" to "reports"
	}

	// Parse command line flags
	flag.StringVar(&runner.projectRoot, "root", "..", "Project root directory")
	flag.BoolVar(&runner.verbose, "v", false, "Verbose output")
	flag.BoolVar(&runner.coverage, "cover", false, "Generate coverage report")
	flag.BoolVar(&runner.benchmark, "bench", false, "Run benchmark tests")
	flag.BoolVar(&runner.parallel, "parallel", true, "Run tests in parallel")
	flag.StringVar(&runner.pattern, "run", "", "Run only tests matching pattern")
	flag.Parse()

	// Ensure reports directory exists
	reportsPath := filepath.Join(runner.projectRoot, runner.reportsDir)
	if err := os.MkdirAll(reportsPath, 0755); err != nil {
		log.Fatalf("Failed to create reports directory: %v", err)
	}

	fmt.Printf("🧪 Kademlia Test Suite Runner\n")
	fmt.Printf("📅 Timestamp: %s\n", runner.timestamp)
	fmt.Printf("📁 Reports Directory: %s\n", reportsPath)
	fmt.Println(strings.Repeat("=", 50))

	if err := runner.Run(); err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}
}

// Run executes the complete test suite
func (tr *TestRunner) Run() error {
	fmt.Println("🚀 Starting comprehensive test suite execution...")

	// Define test suites
	suites := []TestSuite{
		{
			Name:        "Unit Tests",
			Path:        "./tests/unit/...",
			Description: "Individual component tests",
			Timeout:     30 * time.Second,
		},
		{
			Name:        "Integration Tests",
			Path:        "./tests/integration/...",
			Description: "Multi-component workflow tests",
			Timeout:     2 * time.Minute,
		},
		{
			Name:        "Coverage Analysis",
			Path:        "./tests/coverage/...",
			Description: "Code coverage analysis",
			Timeout:     1 * time.Minute,
		},
	}

	totalResults := make([]*TestResult, 0)

	// Run each test suite
	for _, suite := range suites {
		fmt.Printf("\n📋 Running %s...\n", suite.Name)
		fmt.Printf("   %s\n", suite.Description)

		result, err := tr.runTestSuite(suite)
		if err != nil {
			fmt.Printf("❌ Suite %s failed: %v\n", suite.Name, err)
			continue
		}

		totalResults = append(totalResults, result)

		if result.Success {
			fmt.Printf("✅ %s completed successfully (%.2fs)\n", suite.Name, result.Duration.Seconds())
		} else {
			fmt.Printf("❌ %s failed (%.2fs)\n", suite.Name, result.Duration.Seconds())
		}

		if result.Coverage > 0 {
			fmt.Printf("📊 Coverage: %.1f%%\n", result.Coverage)
		}
	}

	// Run benchmarks if requested
	if tr.benchmark {
		fmt.Println("\n⚡ Running performance benchmarks...")
		if err := tr.runBenchmarks(); err != nil {
			fmt.Printf("❌ Benchmarks failed: %v\n", err)
		} else {
			fmt.Println("✅ Benchmarks completed successfully")
		}
	}

	// Generate summary report
	tr.generateSummaryReport(totalResults)

	fmt.Println("\n🎉 Test suite execution completed!")
	fmt.Printf("📊 Reports saved to: %s\n", tr.reportsDir)

	return nil
}

// TestResult contains the results of a test suite execution
type TestResult struct {
	Success  bool
	Duration time.Duration
	Coverage float64
	Output   string
	Suite    string
}

// runTestSuite executes a single test suite
func (tr *TestRunner) runTestSuite(suite TestSuite) (*TestResult, error) {
	start := time.Now()

	// Build go test command
	args := []string{"test"}

	if tr.verbose {
		args = append(args, "-v")
	}

	if tr.coverage {
		coverageFile := filepath.Join(tr.projectRoot, tr.reportsDir, "coverage", fmt.Sprintf("coverage_%s_%s.out",
			strings.ToLower(strings.ReplaceAll(suite.Name, " ", "_")), tr.timestamp))
		// Ensure the directory exists
		os.MkdirAll(filepath.Dir(coverageFile), 0755)
		args = append(args, "-coverprofile="+coverageFile)
		args = append(args, "-covermode=atomic")
	}

	// Add timeout
	args = append(args, "-timeout="+suite.Timeout.String())

	if tr.parallel {
		args = append(args, "-parallel=4")
	}

	if tr.pattern != "" {
		args = append(args, "-run="+tr.pattern)
	}

	// Add package path
	args = append(args, suite.Path)

	cmd := exec.Command("go", args...)
	cmd.Dir = tr.projectRoot

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Save test output to file
	outputFile := filepath.Join(tr.projectRoot, tr.reportsDir, strings.ToLower(strings.ReplaceAll(suite.Name, " ", "_")),
		fmt.Sprintf("output_%s.txt", tr.timestamp))
	os.MkdirAll(filepath.Dir(outputFile), 0755)
	os.WriteFile(outputFile, output, 0644)

	// Parse coverage if available
	coverage := tr.parseCoverage(outputStr)

	// Print output if verbose or if failed
	if tr.verbose || err != nil {
		fmt.Println(outputStr)
	}

	return &TestResult{
		Success:  err == nil,
		Duration: time.Since(start),
		Coverage: coverage,
		Output:   outputStr,
		Suite:    suite.Name,
	}, nil
}

// runBenchmarks executes benchmark tests
func (tr *TestRunner) runBenchmarks() error {
	benchmarkFile := filepath.Join(tr.projectRoot, tr.reportsDir, "benchmark", fmt.Sprintf("benchmark_%s.txt", tr.timestamp))

	args := []string{"test", "-bench=.", "-benchmem", "-benchtime=5s", "./tests/benchmark/..."}

	cmd := exec.Command("go", args...)
	cmd.Dir = tr.projectRoot

	output, err := cmd.CombinedOutput()

	// Save benchmark results
	os.MkdirAll(filepath.Dir(benchmarkFile), 0755)
	os.WriteFile(benchmarkFile, output, 0644)

	if tr.verbose {
		fmt.Println(string(output))
	}

	return err
}

// parseCoverage extracts coverage percentage from test output
func (tr *TestRunner) parseCoverage(output string) float64 {
	re := regexp.MustCompile(`coverage:\s+(\d+\.?\d*)%`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		if coverage, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return coverage
		}
	}
	return 0
}

// generateSummaryReport creates a comprehensive test summary
func (tr *TestRunner) generateSummaryReport(results []*TestResult) {
	summaryFile := filepath.Join(tr.projectRoot, tr.reportsDir, fmt.Sprintf("summary_%s.txt", tr.timestamp))

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Kademlia Test Suite Summary\n"))
	summary.WriteString(fmt.Sprintf("Timestamp: %s\n", tr.timestamp))
	summary.WriteString(fmt.Sprintf("Total Suites: %d\n\n", len(results)))

	totalDuration := time.Duration(0)
	successCount := 0
	totalCoverage := 0.0
	coverageCount := 0

	for _, result := range results {
		totalDuration += result.Duration
		if result.Success {
			successCount++
		}
		if result.Coverage > 0 {
			totalCoverage += result.Coverage
			coverageCount++
		}

		status := "FAILED"
		if result.Success {
			status = "PASSED"
		}

		summary.WriteString(fmt.Sprintf("Suite: %s\n", result.Suite))
		summary.WriteString(fmt.Sprintf("Status: %s\n", status))
		summary.WriteString(fmt.Sprintf("Duration: %.2fs\n", result.Duration.Seconds()))
		if result.Coverage > 0 {
			summary.WriteString(fmt.Sprintf("Coverage: %.1f%%\n", result.Coverage))
		}
		summary.WriteString("\n")
	}

	summary.WriteString("Overall Statistics:\n")
	summary.WriteString(fmt.Sprintf("Success Rate: %d/%d (%.1f%%)\n", successCount, len(results),
		float64(successCount)/float64(len(results))*100))
	summary.WriteString(fmt.Sprintf("Total Duration: %.2fs\n", totalDuration.Seconds()))
	if coverageCount > 0 {
		summary.WriteString(fmt.Sprintf("Average Coverage: %.1f%%\n", totalCoverage/float64(coverageCount)))
	}

	os.WriteFile(summaryFile, []byte(summary.String()), 0644)

	// Print summary to console
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("📊 TEST SUITE SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Print(summary.String())
}
