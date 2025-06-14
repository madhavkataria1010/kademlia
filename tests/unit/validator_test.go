package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Aradhya2708/kademlia/internals/kademlia"
	validators "github.com/Aradhya2708/kademlia/internals/validator"
	"github.com/Aradhya2708/kademlia/tests/testutils"
)

// TestValidators tests the validation functionality
func TestValidators(t *testing.T) {
	logger := testutils.NewTestLogger(t, "VALIDATORS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting validator tests")

	t.Run("ValidHexadecimalIDs", func(t *testing.T) {
		section := logger.Section("Valid Hexadecimal IDs")

		section.Step(1, "Test valid 40-character hex IDs")
		validIDs := []string{
			fixtures.GenerateValidHexID("test1"),
			fixtures.GenerateValidHexID("test2"),
			"1234567890abcdef1234567890abcdef12345678",
			"ABCDEF1234567890ABCDEF1234567890ABCDEF12",
			"0000000000000000000000000000000000000000",
			"ffffffffffffffffffffffffffffffffffffffff",
		}

		for i, id := range validIDs {
			section.Step(i+2, "Validating ID: "+id[:8]+"...")
			err := validators.ValidateID(id, validators.HexadecimalValidator)
			assert.NoError(err, "Valid ID should pass validation: %s...", id[:8])
		}

		section.Success("All valid IDs passed validation")
	})

	t.Run("InvalidHexadecimalIDs", func(t *testing.T) {
		section := logger.Section("Invalid Hexadecimal IDs")

		section.Step(1, "Test invalid IDs")
		invalidIDs := fixtures.GenerateInvalidIDs()

		for desc, invalidID := range invalidIDs {
			section.Step(2, "Testing invalid ID: "+desc)
			err := validators.ValidateID(invalidID, validators.HexadecimalValidator)
			assert.HasError(err, "Invalid ID should fail validation: %s", desc)
		}

		section.Success("All invalid IDs properly rejected")
	})

	t.Run("ValidatorConfig", func(t *testing.T) {
		section := logger.Section("Validator Config")

		section.Step(1, "Verify default hexadecimal validator config")
		config := validators.HexadecimalValidator

		assert.Equal(40, config.Length, "Should require 40 characters")
		assert.NotNil(config.Pattern, "Should have regex pattern")

		section.Step(2, "Test pattern matching")
		// Test pattern directly
		validHex := "1234567890abcdef1234567890abcdef12345678"
		invalidHex := "1234567890abcdef1234567890abcdef1234567g"

		assert.True(config.Pattern.MatchString(validHex), "Pattern should match valid hex")
		assert.False(config.Pattern.MatchString(invalidHex), "Pattern should not match invalid hex")

		section.Success("Validator config working correctly")
	})

	t.Run("EdgeCases", func(t *testing.T) {
		section := logger.Section("Edge Cases")

		section.Step(1, "Test boundary lengths")
		// Test length exactly one less and one more than required
		shortID := "1234567890abcdef1234567890abcdef1234567"  // 39 chars
		longID := "1234567890abcdef1234567890abcdef123456789" // 41 chars

		err := validators.ValidateID(shortID, validators.HexadecimalValidator)
		assert.HasError(err, "Should reject ID with 39 characters")

		err = validators.ValidateID(longID, validators.HexadecimalValidator)
		assert.HasError(err, "Should reject ID with 41 characters")

		section.Step(2, "Test mixed case")
		mixedCaseID := "1234567890ABCdef1234567890abcDEF12345678"
		err = validators.ValidateID(mixedCaseID, validators.HexadecimalValidator)
		assert.NoError(err, "Should accept mixed case hex")

		section.Step(3, "Test special characters")
		specialCharIDs := []string{
			"1234567890abcdef1234567890abcdef1234567!",
			"1234567890abcdef1234567890abcdef1234567@",
			"1234567890abcdef1234567890abcdef1234567#",
			"1234567890abcdef1234567890abcdef1234567$",
			"1234567890abcdef 234567890abcdef12345678", // space
		}

		for _, id := range specialCharIDs {
			err = validators.ValidateID(id, validators.HexadecimalValidator)
			assert.HasError(err, "Should reject ID with special characters")
		}

		section.Success("Edge cases handled correctly")
	})

	t.Run("ValidatorPerformance", func(t *testing.T) {
		section := logger.Section("Validator Performance")

		section.Step(1, "Test validation performance")
		validID := fixtures.GenerateValidHexID("perf")
		invalidID := "invalid-id"

		// Warm up
		for i := 0; i < 100; i++ {
			validators.ValidateID(validID, validators.HexadecimalValidator)
			validators.ValidateID(invalidID, validators.HexadecimalValidator)
		}

		section.Step(2, "Benchmark validation speed")
		numValidations := 10000

		// Time valid ID validations
		start := time.Now()
		for i := 0; i < numValidations; i++ {
			validators.ValidateID(validID, validators.HexadecimalValidator)
		}
		validDuration := time.Since(start)

		// Time invalid ID validations
		start = time.Now()
		for i := 0; i < numValidations; i++ {
			validators.ValidateID(invalidID, validators.HexadecimalValidator)
		}
		invalidDuration := time.Since(start)

		section.Step(3, "Verify performance metrics")
		validRate := float64(numValidations) / validDuration.Seconds()
		invalidRate := float64(numValidations) / invalidDuration.Seconds()

		section.Info("Valid ID validation rate: %.0f validations/sec", validRate)
		section.Info("Invalid ID validation rate: %.0f validations/sec", invalidRate)

		assert.True(validRate > 10000, "Should validate valid IDs quickly")
		assert.True(invalidRate > 10000, "Should validate invalid IDs quickly")

		section.Success("Validator performance acceptable")
	})
}

// TestValidatorIntegration tests validator integration with other components
func TestValidatorIntegration(t *testing.T) {
	logger := testutils.NewTestLogger(t, "VALIDATORS")
	assert := testutils.NewAssert(logger)
	fixtures := testutils.NewTestFixtures(logger)

	logger.Info("Starting validator integration tests")

	t.Run("ValidatorWithHandlers", func(t *testing.T) {
		section := logger.Section("Validator with Handlers")

		section.Step(1, "Test validator integration in find_node handler")
		node := fixtures.CreateTestNode(8080, "valid")
		routingTable := kademlia.NewRoutingTable(node.ID)

		// Valid request
		validID := fixtures.GenerateValidHexID("handler")
		req, _ := http.NewRequest("GET", "/find_node?id="+validID, nil)
		rr := httptest.NewRecorder()

		kademlia.FindNodeHandler(rr, req, node, routingTable)
		assert.Equal(http.StatusOK, rr.Code, "Valid ID should be accepted by handler")

		// Invalid request
		invalidID := "invalid"
		req, _ = http.NewRequest("GET", "/find_node?id="+invalidID, nil)
		rr = httptest.NewRecorder()

		kademlia.FindNodeHandler(rr, req, node, routingTable)
		assert.Equal(http.StatusBadRequest, rr.Code, "Invalid ID should be rejected by handler")

		section.Success("Validator integration with handlers working")
	})

	t.Run("ValidatorErrorMessages", func(t *testing.T) {
		section := logger.Section("Validator Error Messages")

		section.Step(1, "Test specific error messages")

		// Test length error
		shortID := "short"
		err := validators.ValidateID(shortID, validators.HexadecimalValidator)
		assert.HasError(err, "Short ID should produce error")
		assert.Contains(err.Error(), "length", "Error should mention length")

		// Test format error
		longButInvalidID := "1234567890abcdef1234567890abcdef1234567g"
		err = validators.ValidateID(longButInvalidID, validators.HexadecimalValidator)
		assert.HasError(err, "Invalid format should produce error")
		assert.Contains(err.Error(), "format", "Error should mention format")

		section.Success("Validator error messages are descriptive")
	})
}
