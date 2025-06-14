# Kademlia Testing - Quick Reference

## Most Common Commands

```bash
# Quick test run (30 seconds)
make test-unit

# Full integration testing (2 minutes)  
make test-integration

# Performance benchmarks (3 minutes)
make test-benchmark

# Code quality checks (10 seconds)
make test-fmt && make test-vet

# View all available commands
make help

# Clean up test artifacts
make test-clean
```

## Report Locations

```bash
# Latest unit test report
ls -t reports/unit/ | head -1

# Latest integration test report  
ls -t reports/integration/ | head -1

# Latest benchmark report
ls -t reports/benchmark/ | head -1

# View latest unit test summary
head -50 reports/unit/$(ls -t reports/unit/ | head -1)
```

## Troubleshooting

```bash
# If tests hang, use timeout
timeout 60 make test-unit

# Check for test failures
grep -r "FAIL\|ERROR" reports/ | tail -10

# Clean and retry
make test-clean
make test-unit
```

## Current Test Status

- **Unit Tests**: 1,446/1,448 passing (99.86%)
- **Integration Tests**: All passing ✅
- **Benchmarks**: All working ✅  
- **Code Quality**: All checks passing ✅
- **Reports**: Enhanced summaries with failure analysis ✅

## Known Issues

1. **TestPingHandler**: Port validation error (failing test)
2. **TestKademliaIntegration**: Node discovery issue (failing test)
3. **Some commands hang**: Use timeout as workaround

For complete documentation, see `TESTING_GUIDE.md`.
