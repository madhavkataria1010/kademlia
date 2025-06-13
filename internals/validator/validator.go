package validators

import (
	"errors"
	"regexp"
)

// ValidatorConfig holds the configuration for validation
type ValidatorConfig struct {
	Length  int
	Pattern *regexp.Regexp
}

// HexadecimalValidator is a default validator for 160-bit IDs
var HexadecimalValidator = ValidatorConfig{
	Length:  40,
	Pattern: regexp.MustCompile("^[a-fA-F0-9]{40}$"),
}

// ValidateID checks if a given ID matches the required format
func ValidateID(id string, config ValidatorConfig) error {
	if len(id) != config.Length {
		return errors.New("invalid length")
	}
	if !config.Pattern.MatchString(id) {
		return errors.New("invalid format")
	}
	return nil
}
