# Kademlia Testing Makefile

# Generate timestamp for reports
TIMESTAMP := $(shell date +%Y-%m-%d_%H-%M-%S)
REPORTS_DIR := reports

.PHONY: test test-unit test-integration test-all test-coverage test-benchmark clean help setup-reports

# Default target
help: ## Show this help message
	@echo "Kademlia Test Suite Commands:"
	@echo "=============================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Setup reports directory
setup-reports:
	@mkdir -p $(REPORTS_DIR)/unit
	@mkdir -p $(REPORTS_DIR)/integration
	@mkdir -p $(REPORTS_DIR)/coverage
	@mkdir -p $(REPORTS_DIR)/benchmark


# Basic test commands
test: test-unit ## Run unit tests only

test-unit: setup-reports ## Run unit tests with verbose output and timestamped reports
	@echo "🧪 Running Unit Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/unit/unit_tests_$$TIMESTAMP.log" && \
	go test -v -timeout=30s ./tests/unit/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Unit tests PASSED" && \
	echo "📄 Detailed report: $$REPORT_FILE" && \
	echo "📊 Generating comprehensive summary..." && \
	./tests/enhanced_summary.sh "$$REPORT_FILE" && \
	echo "📋 Enhanced summary added to top of report file" || \
	(echo "❌ Unit tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	./tests/enhanced_summary.sh "$$REPORT_FILE" && \
	echo "📋 Enhanced summary with failure analysis added to top of report file" && \
	echo "🔍 Last 10 lines of errors:" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-integration: setup-reports ## Run integration tests with timestamped reports
	@echo "🔗 Running Integration Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/integration/integration_tests_$$TIMESTAMP.log" && \
	go test -v -timeout=2m ./tests/integration/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Integration tests PASSED" && \
	echo "📄 Detailed report: $$REPORT_FILE" && \
	echo "📊 Generating comprehensive summary..." && \
	./tests/enhanced_summary.sh "$$REPORT_FILE" && \
	echo "📋 Enhanced summary added to top of report file" || \
	(echo "❌ Integration tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	./tests/enhanced_summary.sh "$$REPORT_FILE" && \
	echo "📋 Enhanced summary with failure analysis added to top of report file" && \
	echo "🔍 Last 10 lines of errors:" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-all: setup-reports ## Run all tests (unit + integration) with timestamped reports
	@echo "🚀 Running All Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/all_tests_$$TIMESTAMP.log" && \
	go test -v ./tests/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ All tests PASSED" && \
	echo "📄 Detailed report: $$REPORT_FILE" || \
	(echo "❌ Some tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

# Coverage commands
test-coverage: setup-reports ## Run tests with coverage report and timestamped output
	@echo "📊 Generating Coverage Report..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	COVERAGE_FILE="$(REPORTS_DIR)/coverage/coverage_$$TIMESTAMP.out" && \
	COVERAGE_HTML="$(REPORTS_DIR)/coverage/coverage_$$TIMESTAMP.html" && \
	COVERAGE_SUMMARY="$(REPORTS_DIR)/coverage/coverage_summary_$$TIMESTAMP.txt" && \
	COVERAGE_DETAILED="$(REPORTS_DIR)/coverage/coverage_detailed_$$TIMESTAMP.txt" && \
	go test -v -coverprofile="$$COVERAGE_FILE" -covermode=atomic ./... > "$(REPORTS_DIR)/coverage/coverage_verbose_$$TIMESTAMP.log" 2>&1 && \
	go tool cover -html="$$COVERAGE_FILE" -o "$$COVERAGE_HTML" && \
	go tool cover -func="$$COVERAGE_FILE" > "$$COVERAGE_SUMMARY" && \
	COVERAGE_PERCENT=$$(go tool cover -func="$$COVERAGE_FILE" | tail -1 | awk '{print $$3}') && \
	echo "✅ Coverage analysis completed: $$COVERAGE_PERCENT" && \
	echo "📄 Coverage report (HTML): $$COVERAGE_HTML" && \
	echo "📄 Coverage summary: $$COVERAGE_SUMMARY" && \
	echo "📄 Detailed logs: $(REPORTS_DIR)/coverage/coverage_verbose_$$TIMESTAMP.log" && \
	echo "📊 Generating detailed coverage analysis..." && \
	cd tests && go run -tags summary ../tests/testutils/summary.go ../$(REPORTS_DIR)/coverage/coverage_verbose_$$TIMESTAMP.log > ../$$COVERAGE_DETAILED 2>/dev/null || echo "Coverage analysis completed" && \
	echo "📋 Detailed coverage analysis: $$COVERAGE_DETAILED" || \
	(echo "❌ Coverage generation FAILED" && \
	echo "📄 Error details: $(REPORTS_DIR)/coverage/coverage_verbose_$$TIMESTAMP.log" && \
	exit 1)

test-coverage-func: setup-reports ## Show coverage by function with timestamped report
	@echo "📋 Function Coverage Analysis..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	COVERAGE_FILE="$(REPORTS_DIR)/coverage/coverage_func_$$TIMESTAMP.out" && \
	FUNC_REPORT="$(REPORTS_DIR)/coverage/func_coverage_$$TIMESTAMP.txt" && \
	go test -coverprofile="$$COVERAGE_FILE" ./... > /dev/null 2>&1 && \
	go tool cover -func="$$COVERAGE_FILE" | tee "$$FUNC_REPORT" && \
	echo "✅ Function coverage analysis completed" && \
	echo "📄 Function coverage report: $$FUNC_REPORT" || \
	(echo "❌ Function coverage FAILED" && exit 1)

# Benchmark commands with timestamped reports
test-benchmark: setup-reports ## Run benchmark tests with timestamped reports
	@echo "🏁 Running Benchmark Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/benchmark/benchmark_$$TIMESTAMP.log" && \
	go test -v -bench=. -benchmem -timeout=5m ./tests/benchmark/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Benchmark tests COMPLETED" && \
	echo "📄 Benchmark report: $$REPORT_FILE" && \
	echo "📊 Generating comprehensive summary..." && \
	./tests/enhanced_summary.sh "$$REPORT_FILE" && \
	echo "📋 Enhanced summary added to top of report file" || \
	(echo "❌ Benchmark tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	./tests/enhanced_summary.sh "$$REPORT_FILE" && \
	echo "📋 Enhanced summary with failure analysis added to top of report file" && \
	echo "🔍 Last 10 lines of errors:" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-benchmark-cpu: setup-reports ## Run CPU profiling benchmarks with timestamped reports
	@echo "🔥 Running CPU Profiling Benchmarks..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	PROFILE_FILE="$(REPORTS_DIR)/benchmark/cpu_profile_$$TIMESTAMP.prof" && \
	REPORT_FILE="$(REPORTS_DIR)/benchmark/cpu_bench_$$TIMESTAMP.log" && \
	go test -v -bench=. -cpuprofile="$$PROFILE_FILE" ./tests/benchmark/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ CPU profiling benchmarks COMPLETED" && \
	echo "📄 CPU profile: $$PROFILE_FILE" && \
	echo "📄 Benchmark report: $$REPORT_FILE" || \
	(echo "❌ CPU profiling benchmarks FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	exit 1)

test-benchmark-mem: setup-reports ## Run memory profiling benchmarks with timestamped reports
	@echo "💾 Running Memory Profiling Benchmarks..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	PROFILE_FILE="$(REPORTS_DIR)/benchmark/mem_profile_$$TIMESTAMP.prof" && \
	REPORT_FILE="$(REPORTS_DIR)/benchmark/mem_bench_$$TIMESTAMP.log" && \
	go test -v -bench=. -memprofile="$$PROFILE_FILE" ./tests/benchmark/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Memory profiling benchmarks COMPLETED" && \
	echo "📄 Memory profile: $$PROFILE_FILE" && \
	echo "📄 Benchmark report: $$REPORT_FILE" || \
	(echo "❌ Memory profiling benchmarks FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	exit 1)

# Specific test categories with timestamped reports
test-models: setup-reports ## Run model tests only with timestamped reports
	@echo "🏗️  Running Model Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/unit/model_tests_$$TIMESTAMP.log" && \
	go test -v -run="TestNode|TestBucket|TestRoutingTable|TestKeyValueStore|TestMessage" ./tests/unit/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Model tests PASSED" && \
	echo "📄 Model test report: $$REPORT_FILE" || \
	(echo "❌ Model tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-handlers: setup-reports ## Run handler tests only with timestamped reports
	@echo "🌐 Running Handler Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/unit/handler_tests_$$TIMESTAMP.log" && \
	go test -v -run="Test.*Handler" ./tests/unit/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Handler tests PASSED" && \
	echo "📄 Handler test report: $$REPORT_FILE" || \
	(echo "❌ Handler tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-kademlia: setup-reports ## Run core Kademlia algorithm tests with timestamped reports
	@echo "🔍 Running Kademlia Core Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/unit/kademlia_core_tests_$$TIMESTAMP.log" && \
	go test -v -run="TestKademlia" ./tests/unit/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Kademlia core tests PASSED" && \
	echo "📄 Kademlia core test report: $$REPORT_FILE" || \
	(echo "❌ Kademlia core tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-validators: setup-reports ## Run validator tests only with timestamped reports
	@echo "✅ Running Validator Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/unit/validator_tests_$$TIMESTAMP.log" && \
	go test -v -run="TestValidator" ./tests/unit/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Validator tests PASSED" && \
	echo "📄 Validator test report: $$REPORT_FILE" || \
	(echo "❌ Validator tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-integration-workflow: setup-reports ## Run workflow integration tests with timestamped reports
	@echo "🔄 Running Workflow Integration Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/integration/workflow_tests_$$TIMESTAMP.log" && \
	go test -v -run="TestFullKademliaWorkflow" ./tests/integration/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Workflow integration tests PASSED" && \
	echo "📄 Workflow test report: $$REPORT_FILE" || \
	(echo "❌ Workflow integration tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-integration-resilience: setup-reports ## Run resilience integration tests with timestamped reports
	@echo "🛡️  Running Resilience Integration Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/integration/resilience_tests_$$TIMESTAMP.log" && \
	go test -v -run="TestNetworkResilience" ./tests/integration/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Resilience integration tests PASSED" && \
	echo "📄 Resilience test report: $$REPORT_FILE" || \
	(echo "❌ Resilience integration tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-integration-scalability: setup-reports ## Run scalability integration tests with timestamped reports
	@echo "📈 Running Scalability Integration Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/integration/scalability_tests_$$TIMESTAMP.log" && \
	go test -v -run="TestScalability" ./tests/integration/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Scalability integration tests PASSED" && \
	echo "📄 Scalability test report: $$REPORT_FILE" || \
	(echo "❌ Scalability integration tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

# Continuous testing with timestamped reports
test-watch: ## Run tests in watch mode (requires entr)
	@echo "👀 Watching for changes..."
	find . -name "*.go" | entr -c make test-unit

test-race: setup-reports ## Run tests with race detection and timestamped reports
	@echo "🏃 Running Tests with Race Detection..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/race_test_$$TIMESTAMP.log" && \
	go test -race -timeout=1m ./tests/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Race detection tests PASSED" && \
	echo "📄 Race test report: $$REPORT_FILE" || \
	(echo "❌ Race detection tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-short: setup-reports ## Run tests with -short flag and timestamped reports
	@echo "⚡ Running Short Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/short_test_$$TIMESTAMP.log" && \
	go test -short ./tests/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Short tests PASSED" && \
	echo "📄 Short test report: $$REPORT_FILE" || \
	(echo "❌ Short tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

# Quality assurance with timestamped reports
test-vet: setup-reports ## Run go vet with timestamped reports
	@echo "🔍 Running Go Vet Analysis..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/vet_$$TIMESTAMP.log" && \
	go vet ./... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Go vet analysis PASSED" && \
	echo "📄 Vet report: $$REPORT_FILE" || \
	(echo "❌ Go vet analysis FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	cat "$$REPORT_FILE" && \
	exit 1)

test-fmt: setup-reports ## Check code formatting with timestamped reports
	@echo "📝 Checking Code Formatting..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/fmt_check_$$TIMESTAMP.log" && \
	go fmt ./... > "$$REPORT_FILE" 2>&1 && \
	if [ -s "$$REPORT_FILE" ]; then \
		echo "❌ Code formatting issues found" && \
		echo "📄 Formatting report: $$REPORT_FILE" && \
		cat "$$REPORT_FILE" && \
		exit 1; \
	else \
		echo "✅ Code formatting OK" && \
		echo "📄 No formatting issues found: $$REPORT_FILE"; \
	fi

test-lint: setup-reports ## Run golint with timestamped reports (requires golint)
	@echo "🧹 Running GoLint Analysis..."
	@command -v golint >/dev/null 2>&1 || (echo "❌ golint not installed. Run: go install golang.org/x/lint/golint@latest" && exit 1)
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/lint_$$TIMESTAMP.log" && \
	golint ./... > "$$REPORT_FILE" 2>&1 && \
	if [ -s "$$REPORT_FILE" ]; then \
		echo "⚠️  Linting suggestions found" && \
		echo "📄 Lint report: $$REPORT_FILE" && \
		cat "$$REPORT_FILE"; \
	else \
		echo "✅ No linting issues found" && \
		echo "📄 Clean lint report: $$REPORT_FILE"; \
	fi

# Test utilities
test-generate-mocks: ## Generate test mocks (requires mockgen)
	@echo "🎭 Generating test mocks..."
	@command -v mockgen >/dev/null 2>&1 || (echo "❌ mockgen not installed. Run: go install github.com/golang/mock/mockgen@latest" && exit 1)
	# Add mockgen commands here when needed

# Custom test runner
test-runner: setup-reports ## Build and run custom test runner
	@echo "🚀 Building Custom Test Runner"
	@cd tests && go build -o run_tests run_tests.go
	@echo "🏃 Executing Test Suite with Custom Runner ($(TIMESTAMP))"
	@cd tests && ./run_tests -cover -bench -v | tee ../$(REPORTS_DIR)/custom_runner_$(TIMESTAMP).log

# Complete test suite with reports
test-complete: setup-reports ## Run complete test suite with all reports
	@echo "🎯 Complete Test Suite Execution ($(TIMESTAMP))"
	@echo "================================================"
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	COMPLETE_LOG="$(REPORTS_DIR)/complete_suite_$$TIMESTAMP.log" && \
	echo "Starting complete test suite at $$(date)" > "$$COMPLETE_LOG" && \
	echo "Coverage Tests:" >> "$$COMPLETE_LOG" && \
	make test-coverage >> "$$COMPLETE_LOG" 2>&1 && \
	echo "Benchmark Tests:" >> "$$COMPLETE_LOG" && \
	make test-benchmark >> "$$COMPLETE_LOG" 2>&1 && \
	echo "Integration Tests:" >> "$$COMPLETE_LOG" && \
	make test-integration >> "$$COMPLETE_LOG" 2>&1 && \
	echo "Code Quality Checks:" >> "$$COMPLETE_LOG" && \
	make test-vet >> "$$COMPLETE_LOG" 2>&1 && \
	echo "✅ Complete test suite finished successfully" && \
	echo "📄 Complete suite log: $$COMPLETE_LOG" && \
	echo "📁 All reports saved in $(REPORTS_DIR)/" || \
	(echo "❌ Complete test suite FAILED" && \
	echo "📄 Error details in: $$COMPLETE_LOG" && \
	exit 1)

# Performance testing with timestamped reports
test-stress: setup-reports ## Run stress tests with timestamped reports
	@echo "💪 Running Stress Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/integration/stress_test_$$TIMESTAMP.log" && \
	go test -v -run="TestScalability" -timeout=10m ./tests/integration/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Stress tests PASSED" && \
	echo "📄 Stress test report: $$REPORT_FILE" || \
	(echo "❌ Stress tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

test-memory: setup-reports ## Run memory leak tests with timestamped reports
	@echo "🔍 Running Memory Leak Tests..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	PROFILE_FILE="$(REPORTS_DIR)/benchmark/memory_leak_$$TIMESTAMP.prof" && \
	REPORT_FILE="$(REPORTS_DIR)/integration/memory_test_$$TIMESTAMP.log" && \
	go test -v -run="TestScalability/HighVolumeOperations" -memprofile="$$PROFILE_FILE" ./tests/integration/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Memory leak tests COMPLETED" && \
	echo "📄 Memory profile: $$PROFILE_FILE" && \
	echo "📄 Memory test report: $$REPORT_FILE" && \
	echo "🔍 Analyzing memory profile..." && \
	go tool pprof -top "$$PROFILE_FILE" || \
	(echo "❌ Memory leak tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	exit 1)

# CI/CD targets with timestamped reports
test-ci: setup-reports ## Run all CI tests with timestamped reports
	@echo "🤖 Running CI Test Suite..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	CI_LOG="$(REPORTS_DIR)/ci_suite_$$TIMESTAMP.log" && \
	echo "CI Test Suite - $$(date)" > "$$CI_LOG" && \
	echo "======================" >> "$$CI_LOG" && \
	make test-fmt >> "$$CI_LOG" 2>&1 && \
	make test-vet >> "$$CI_LOG" 2>&1 && \
	make test-coverage >> "$$CI_LOG" 2>&1 && \
	echo "✅ CI test suite completed successfully" && \
	echo "📄 CI suite log: $$CI_LOG" || \
	(echo "❌ CI test suite FAILED" && \
	echo "📄 Error details: $$CI_LOG" && \
	tail -10 "$$CI_LOG" && \
	exit 1)

test-ci-fast: setup-reports ## Run fast CI tests with timestamped reports
	@echo "⚡ Running Fast CI Test Suite..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	CI_FAST_LOG="$(REPORTS_DIR)/ci_fast_$$TIMESTAMP.log" && \
	echo "Fast CI Test Suite - $$(date)" > "$$CI_FAST_LOG" && \
	echo "=========================" >> "$$CI_FAST_LOG" && \
	make test-short >> "$$CI_FAST_LOG" 2>&1 && \
	make test-fmt >> "$$CI_FAST_LOG" 2>&1 && \
	make test-vet >> "$$CI_FAST_LOG" 2>&1 && \
	echo "✅ Fast CI test suite completed successfully" && \
	echo "📄 Fast CI suite log: $$CI_FAST_LOG" || \
	(echo "❌ Fast CI test suite FAILED" && \
	echo "📄 Error details: $$CI_FAST_LOG" && \
	tail -10 "$$CI_FAST_LOG" && \
	exit 1)

# Comprehensive reporting
test-report: setup-reports ## Generate comprehensive test report with timestamps
	@echo "📊 Generating Comprehensive Test Report..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/comprehensive_report_$$TIMESTAMP.txt" && \
	echo "Kademlia Test Report - $$(date)" > "$$REPORT_FILE" && \
	echo "===============================" >> "$$REPORT_FILE" && \
	echo "" >> "$$REPORT_FILE" && \
	echo "Coverage Summary:" >> "$$REPORT_FILE" && \
	make test-coverage >> "$$REPORT_FILE" 2>&1 && \
	echo "" >> "$$REPORT_FILE" && \
	echo "Test Results:" >> "$$REPORT_FILE" && \
	go test -v ./tests/... >> "$$REPORT_FILE" 2>&1 || true && \
	echo "✅ Comprehensive test report generated" && \
	echo "📄 Test report: $$REPORT_FILE"

# Development helpers with timestamped reports
test-debug: setup-reports ## Run tests with debug output and timestamped reports
	@echo "🐛 Running Tests with Debug Output..."
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/debug_test_$$TIMESTAMP.log" && \
	go test -v -tags debug ./tests/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Debug tests COMPLETED" && \
	echo "📄 Debug test report: $$REPORT_FILE" || \
	(echo "❌ Debug tests FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	exit 1)

test-single: setup-reports ## Run a single test with timestamped report (usage: make test-single TEST=TestName)
	@if [ -z "$(TEST)" ]; then \
		echo "❌ Usage: make test-single TEST=TestName"; \
		exit 1; \
	fi
	@echo "🎯 Running Single Test: $(TEST)"
	@TIMESTAMP=$(shell date +%Y-%m-%d_%H-%M-%S) && \
	REPORT_FILE="$(REPORTS_DIR)/single_test_$(TEST)_$$TIMESTAMP.log" && \
	go test -v -run="$(TEST)" ./tests/... > "$$REPORT_FILE" 2>&1 && \
	echo "✅ Single test $(TEST) PASSED" && \
	echo "📄 Single test report: $$REPORT_FILE" || \
	(echo "❌ Single test $(TEST) FAILED" && \
	echo "📄 Error details: $$REPORT_FILE" && \
	tail -10 "$$REPORT_FILE" && \
	exit 1)

# Test cleanup
test-clean: ## Clean test artifacts and reports
	@echo "🧹 Cleaning Test Artifacts..."
	@rm -rf $(REPORTS_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.prof *.trace
	@rm -f tests/run_tests
	@echo "✅ Test artifacts cleaned"

# Documentation
test-godoc: ## Generate and serve test documentation
	@echo "📚 Starting documentation server..."
	@echo "Open http://localhost:6060/pkg/github.com/Aradhya2708/kademlia/"
	godoc -http=:6060

# Default make target
.DEFAULT_GOAL := help
