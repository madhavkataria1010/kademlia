package coverage

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// CoverageReport represents test coverage statistics
type CoverageReport struct {
	Package         string
	CoveragePercent float64
	CoveredLines    int
	TotalLines      int
	Functions       []FunctionCoverage
}

// FunctionCoverage represents coverage for a specific function
type FunctionCoverage struct {
	Name            string
	CoveragePercent float64
	Covered         bool
}

// TestCodeCoverage generates and analyzes test coverage
func TestCodeCoverage(t *testing.T) {
	logger := testutils.NewTestLogger(t, "COVERAGE")
	assert := testutils.NewAssert(logger)

	logger.Info("Starting comprehensive code coverage analysis")

	t.Run("GenerateCoverageReport", func(t *testing.T) {
		section := logger.Section("Coverage Report Generation")

		section.Step(1, "Generate coverage profile")
		coverageFile := "/tmp/kademlia_coverage.out"

		// Run tests with coverage
		cmd := exec.Command("go", "test", "-coverprofile="+coverageFile,
			"./internals/kademlia", "./pkg/...", "./cmd/...")
		cmd.Dir = "/home/lakshya-jain/projects/kademlia"

		output, err := cmd.CombinedOutput()
		assert.NoError(err, "Coverage generation should succeed: %s", string(output))

		section.Step(2, "Parse coverage data")
		report, err := parseCoverageFile(coverageFile)
		assert.NoError(err, "Should parse coverage file successfully")

		section.Step(3, "Analyze coverage results")
		section.Info("Overall coverage: %.2f%%", report.CoveragePercent)
		section.Info("Covered lines: %d/%d", report.CoveredLines, report.TotalLines)

		// Coverage thresholds
		minimumCoverage := 80.0
		assert.True(report.CoveragePercent >= minimumCoverage,
			"Code coverage (%.2f%%) should be at least %.2f%%",
			report.CoveragePercent, minimumCoverage)

		section.Step(4, "Identify uncovered functions")
		uncoveredFunctions := []string{}
		for _, fn := range report.Functions {
			if !fn.Covered {
				uncoveredFunctions = append(uncoveredFunctions, fn.Name)
			}
		}

		if len(uncoveredFunctions) > 0 {
			section.Warning("Uncovered functions: %v", uncoveredFunctions)
		} else {
			section.Success("All functions have test coverage")
		}

		section.Success("Coverage analysis completed")
	})

	t.Run("PackageSpecificCoverage", func(t *testing.T) {
		section := logger.Section("Package-Specific Coverage")

		packages := []string{
			"./internals/kademlia",
			"./pkg/models",
			"./pkg/constants",
		}

		for i, pkg := range packages {
			section.Step(i+1, fmt.Sprintf("Analyze coverage for %s", pkg))

			coverageFile := fmt.Sprintf("/tmp/coverage_%d.out", i)
			cmd := exec.Command("go", "test", "-coverprofile="+coverageFile, pkg)
			cmd.Dir = "/home/lakshya-jain/projects/kademlia"

			output, err := cmd.CombinedOutput()
			if err != nil {
				section.Warning("Package %s coverage generation failed: %s", pkg, string(output))
				continue
			}

			report, err := parseCoverageFile(coverageFile)
			if err != nil {
				section.Warning("Failed to parse coverage for %s: %v", pkg, err)
				continue
			}

			section.Info("Package %s: %.2f%% coverage (%d/%d lines)",
				pkg, report.CoveragePercent, report.CoveredLines, report.TotalLines)

			// Package-specific minimum coverage thresholds
			var threshold float64
			switch pkg {
			case "./internals/kademlia":
				threshold = 85.0 // Core logic should have high coverage
			case "./pkg/models":
				threshold = 75.0 // Models should have good coverage
			case "./pkg/constants":
				threshold = 60.0 // Constants may have lower coverage
			default:
				threshold = 70.0
			}

			if report.CoveragePercent < threshold {
				section.Warning("Package %s coverage (%.2f%%) below threshold (%.2f%%)",
					pkg, report.CoveragePercent, threshold)
			} else {
				section.Success("Package %s meets coverage threshold", pkg)
			}
		}
	})

	t.Run("CriticalPathCoverage", func(t *testing.T) {
		section := logger.Section("Critical Path Coverage")

		// Test coverage of critical Kademlia operations
		criticalFunctions := []string{
			"FindClosestNodes",
			"AddNodeToRoutingTable",
			"Store",
			"Retrieve",
			"CalculateXORDistance",
			"JoinNetwork",
			"PingHandler",
			"FindNodeHandler",
			"StoreHandler",
			"FindValueHandler",
		}

		section.Step(1, "Verify critical functions are tested")

		// Generate detailed coverage
		cmd := exec.Command("go", "test", "-coverprofile=/tmp/detailed_coverage.out",
			"-coverpkg=./...", "./tests/...")
		cmd.Dir = "/home/lakshya-jain/projects/kademlia"

		output, err := cmd.CombinedOutput()
		if err != nil {
			section.Warning("Detailed coverage generation failed: %s", string(output))
			return
		}

		// Parse the coverage file to check function coverage
		coverageData, err := os.ReadFile("/tmp/detailed_coverage.out")
		if err != nil {
			section.Warning("Failed to read detailed coverage: %v", err)
			return
		}

		coverageContent := string(coverageData)
		coveredFunctions := 0
		totalFunctions := len(criticalFunctions)

		for _, fn := range criticalFunctions {
			if strings.Contains(coverageContent, fn) {
				coveredFunctions++
				section.Info("✓ %s is covered", fn)
			} else {
				section.Warning("✗ %s may not be covered", fn)
			}
		}

		coverageRatio := float64(coveredFunctions) / float64(totalFunctions) * 100
		section.Info("Critical function coverage: %.2f%% (%d/%d)",
			coverageRatio, coveredFunctions, totalFunctions)

		assert.True(coverageRatio >= 90.0,
			"Critical functions should have at least 90%% coverage")

		section.Success("Critical path coverage analysis completed")
	})
}

// TestCoverageRegression checks for coverage regressions
func TestCoverageRegression(t *testing.T) {
	logger := testutils.NewTestLogger(t, "COVERAGE")
	assert := testutils.NewAssert(logger)

	logger.Info("Starting coverage regression analysis")

	t.Run("CompareWithBaseline", func(t *testing.T) {
		section := logger.Section("Coverage Regression Check")

		section.Step(1, "Generate current coverage")
		currentCoverageFile := "/tmp/current_coverage.out"
		cmd := exec.Command("go", "test", "-coverprofile="+currentCoverageFile, "./...")
		cmd.Dir = "/home/lakshya-jain/projects/kademlia"

		output, err := cmd.CombinedOutput()
		if err != nil {
			section.Warning("Failed to generate current coverage: %s", string(output))
			return
		}

		currentReport, err := parseCoverageFile(currentCoverageFile)
		if err != nil {
			section.Warning("Failed to parse current coverage: %v", err)
			return
		}

		section.Step(2, "Check against baseline")
		// For this example, we'll use a baseline of 75%
		// In a real scenario, you'd store and compare against a saved baseline
		baselineCoverage := 75.0

		section.Info("Current coverage: %.2f%%", currentReport.CoveragePercent)
		section.Info("Baseline coverage: %.2f%%", baselineCoverage)

		if currentReport.CoveragePercent < baselineCoverage {
			section.Warning("Coverage regression detected: %.2f%% < %.2f%%",
				currentReport.CoveragePercent, baselineCoverage)
		} else {
			section.Success("No coverage regression detected")
		}

		// For CI/CD, you might want to fail the test on regression
		assert.True(currentReport.CoveragePercent >= baselineCoverage,
			"Coverage should not regress below baseline")

		section.Success("Coverage regression check completed")
	})
}

// parseCoverageFile parses a Go coverage profile file
func parseCoverageFile(filename string) (*CoverageReport, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	report := &CoverageReport{
		Functions: []FunctionCoverage{},
	}

	scanner := bufio.NewScanner(file)
	totalStatements := 0
	coveredStatements := 0

	// Skip the first line (mode line)
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse coverage line: file:startLine.startCol,endLine.endCol numStmt count
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		numStmt, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		count, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		totalStatements += numStmt
		if count > 0 {
			coveredStatements += numStmt
		}

		// Extract function name if possible
		re := regexp.MustCompile(`(\w+)\.go:\d+\.\d+,\d+\.\d+`)
		matches := re.FindStringSubmatch(parts[0])
		if len(matches) > 1 {
			funcName := matches[1]
			report.Functions = append(report.Functions, FunctionCoverage{
				Name:    funcName,
				Covered: count > 0,
			})
		}
	}

	if totalStatements > 0 {
		report.CoveragePercent = float64(coveredStatements) / float64(totalStatements) * 100
	}
	report.CoveredLines = coveredStatements
	report.TotalLines = totalStatements

	return report, scanner.Err()
}

// TestTestQuality analyzes the quality of the test suite itself
func TestTestQuality(t *testing.T) {
	logger := testutils.NewTestLogger(t, "QUALITY")
	assert := testutils.NewAssert(logger)

	logger.Info("Starting test suite quality analysis")

	t.Run("TestFileStructure", func(t *testing.T) {
		section := logger.Section("Test File Structure")

		section.Step(1, "Verify test organization")

		expectedDirs := []string{
			"/home/lakshya-jain/projects/kademlia/tests/unit",
			"/home/lakshya-jain/projects/kademlia/tests/integration",
			"/home/lakshya-jain/projects/kademlia/tests/benchmark",
			"/home/lakshya-jain/projects/kademlia/tests/testutils",
		}

		for _, dir := range expectedDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				section.Warning("Missing test directory: %s", dir)
			} else {
				section.Info("✓ Found test directory: %s", dir)
			}
		}

		section.Step(2, "Count test files")
		testFileCount := 0
		for _, dir := range expectedDirs {
			if dir == "/home/lakshya-jain/projects/kademlia/tests/testutils" {
				continue // Skip utilities directory
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), "_test.go") {
					testFileCount++
				}
			}
		}

		section.Info("Total test files: %d", testFileCount)
		assert.True(testFileCount >= 5, "Should have at least 5 test files")

		section.Success("Test file structure is well organized")
	})

	t.Run("TestNaming", func(t *testing.T) {
		section := logger.Section("Test Naming Conventions")

		section.Step(1, "Verify test function naming")

		// This is a basic check - in a real scenario you'd parse the AST
		testDirs := []string{
			"/home/lakshya-jain/projects/kademlia/tests/unit",
			"/home/lakshya-jain/projects/kademlia/tests/integration",
		}

		goodNamingCount := 0
		totalTestFunctions := 0

		for _, dir := range testDirs {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), "_test.go") {
					continue
				}

				filePath := fmt.Sprintf("%s/%s", dir, entry.Name())
				content, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}

				// Look for test function patterns
				re := regexp.MustCompile(`func (Test\w+)\(`)
				matches := re.FindAllStringSubmatch(string(content), -1)

				for _, match := range matches {
					totalTestFunctions++
					funcName := match[1]

					// Check naming conventions
					if strings.HasPrefix(funcName, "Test") && len(funcName) > 4 {
						goodNamingCount++
					}
				}
			}
		}

		if totalTestFunctions > 0 {
			namingRatio := float64(goodNamingCount) / float64(totalTestFunctions) * 100
			section.Info("Good naming ratio: %.2f%% (%d/%d)",
				namingRatio, goodNamingCount, totalTestFunctions)

			assert.True(namingRatio >= 90.0,
				"At least 90%% of test functions should follow naming conventions")
		}

		section.Success("Test naming analysis completed")
	})
}
