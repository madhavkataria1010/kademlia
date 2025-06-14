package testutils

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestLogger provides structured logging for tests
type TestLogger struct {
	t      *testing.T
	prefix string
	indent int
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T, prefix string) *TestLogger {
	return &TestLogger{t: t, prefix: prefix}
}

// Info logs an info message
func (l *TestLogger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Success logs a success message
func (l *TestLogger) Success(format string, args ...interface{}) {
	l.log("✓ SUCCESS", format, args...)
}

// Error logs an error message
func (l *TestLogger) Error(format string, args ...interface{}) {
	l.log("✗ ERROR", format, args...)
}

// Warning logs a warning message
func (l *TestLogger) Warning(format string, args ...interface{}) {
	l.log("⚠ WARNING", format, args...)
}

// Step logs a test step
func (l *TestLogger) Step(step int, description string) {
	l.log("STEP", "[%d] %s", step, description)
}

// Section creates a new section with increased indentation
func (l *TestLogger) Section(name string) *TestLogger {
	l.log("SECTION", "=== %s ===", name)
	return &TestLogger{
		t:      l.t,
		prefix: l.prefix,
		indent: l.indent + 1,
	}
}

// log is the internal logging method
func (l *TestLogger) log(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05.000")
	indentation := strings.Repeat("  ", l.indent)
	message := fmt.Sprintf(format, args...)

	var testName string
	if l.t != nil {
		testName = l.t.Name()
	} else {
		testName = "BENCHMARK"
	}

	logLine := fmt.Sprintf("[%s] %s%s [%s] %s: %s",
		timestamp, indentation, l.prefix, level, testName, message)

	if l.t != nil {
		l.t.Log(logLine)
	} else {
		// For benchmarks or when testing.T is nil, just print to stdout
		fmt.Println(logLine)
	}
}
