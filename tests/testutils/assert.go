package testutils

import (
	"fmt"
	"strings"
)

// Assert provides assertion methods with logging
type Assert struct {
	logger *TestLogger
}

// NewAssert creates a new Assert instance
func NewAssert(logger *TestLogger) *Assert {
	return &Assert{logger: logger}
}

// Equal asserts two values are equal
func (a *Assert) Equal(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if expected != actual {
		msg := "values not equal"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: expected=%v, actual=%v", msg, expected, actual)
		return false
	}
	a.logger.Success("assertion passed: %v == %v", expected, actual)
	return true
}

// NotEqual asserts two values are not equal
func (a *Assert) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) bool {
	if expected == actual {
		msg := "values should not be equal"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: both values are %v", msg, expected)
		return false
	}
	a.logger.Success("assertion passed: %v != %v", expected, actual)
	return true
}

// NotNil asserts value is not nil
func (a *Assert) NotNil(value interface{}, msgAndArgs ...interface{}) bool {
	if value == nil {
		msg := "value should not be nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error(msg)
		return false
	}
	a.logger.Success("assertion passed: value is not nil")
	return true
}

// Nil asserts value is nil
func (a *Assert) Nil(value interface{}, msgAndArgs ...interface{}) bool {
	if value != nil {
		msg := "value should be nil"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: got %v", msg, value)
		return false
	}
	a.logger.Success("assertion passed: value is nil")
	return true
}

// True asserts condition is true
func (a *Assert) True(condition bool, msgAndArgs ...interface{}) bool {
	if !condition {
		msg := "condition should be true"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error(msg)
		return false
	}
	a.logger.Success("assertion passed: condition is true")
	return true
}

// False asserts condition is false
func (a *Assert) False(condition bool, msgAndArgs ...interface{}) bool {
	if condition {
		msg := "condition should be false"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error(msg)
		return false
	}
	a.logger.Success("assertion passed: condition is false")
	return true
}

// Contains asserts that a string contains a substring
func (a *Assert) Contains(str, substr string, msgAndArgs ...interface{}) bool {
	if !strings.Contains(str, substr) {
		msg := "string should contain substring"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: '%s' not found in '%s'", msg, substr, str)
		return false
	}
	a.logger.Success("assertion passed: string contains substring")
	return true
}

// Greater asserts that first value is greater than second
func (a *Assert) Greater(first, second int, msgAndArgs ...interface{}) bool {
	if first <= second {
		msg := "first value should be greater than second"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: %d <= %d", msg, first, second)
		return false
	}
	a.logger.Success("assertion passed: %d > %d", first, second)
	return true
}

// GreaterOrEqual asserts that first value is greater than or equal to second
func (a *Assert) GreaterOrEqual(first, second int, msgAndArgs ...interface{}) bool {
	if first < second {
		msg := "first value should be greater than or equal to second"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: %d < %d", msg, first, second)
		return false
	}
	a.logger.Success("assertion passed: %d >= %d", first, second)
	return true
}

// NoError asserts that error is nil
func (a *Assert) NoError(err error, msgAndArgs ...interface{}) bool {
	if err != nil {
		msg := "expected no error"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error("%s: got error %v", msg, err)
		return false
	}
	a.logger.Success("assertion passed: no error")
	return true
}

// HasError asserts that error is not nil
func (a *Assert) HasError(err error, msgAndArgs ...interface{}) bool {
	if err == nil {
		msg := "expected an error"
		if len(msgAndArgs) > 0 {
			msg = fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}
		a.logger.Error(msg)
		return false
	}
	a.logger.Success("assertion passed: error present: %v", err)
	return true
}
