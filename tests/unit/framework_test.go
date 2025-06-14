package unit

import (
	"testing"

	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestFramework tests that our testing framework works correctly
func TestFramework(t *testing.T) {
	logger := testutils.NewTestLogger(t, "FRAMEWORK")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Testing the test framework itself")

	t.Run("LoggerFunctionality", func(t *testing.T) {
		section := logger.Section("Logger Test")

		section.Step(1, "Test basic logging")
		section.Info("This is an info message")
		section.Success("This is a success message")
		section.Warning("This is a warning message")

		section.Success("Logger working correctly")
	})

	t.Run("AssertFunctionality", func(t *testing.T) {
		section := logger.Section("Assert Test")

		section.Step(1, "Test basic assertions")
		assert.Equal(1, 1, "One should equal one")
		assert.NotEqual(1, 2, "One should not equal two")
		assert.True(true, "True should be true")
		assert.False(false, "False should be false")

		section.Success("Assertions working correctly")
	})

	t.Run("FixturesFunctionality", func(t *testing.T) {
		section := logger.Section("Fixtures Test")

		section.Step(1, "Test fixture generation")
		validID := fixtures.GenerateValidHexID("test")
		assert.Equal(40, len(validID), "Generated ID should be 40 characters")

		invalidIDs := fixtures.GenerateInvalidIDs()
		assert.True(len(invalidIDs) > 0, "Should generate invalid IDs")

		section.Success("Fixtures working correctly")
	})
}
