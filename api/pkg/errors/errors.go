package errors

import (
	"errors"
	"fmt"
)

var (
	InvalidInputError     error = errors.New("invalid input")
	DeadlineExceededError error = errors.New("deadline exceeded")
)

// NewInvalidInputError create new InvalidInputError with specified reason
// errors.Is(error, InvalidInputError) will return true if a new error is created using this function
func NewInvalidInputError(reason string) error {
	return fmt.Errorf("%w: %s", InvalidInputError, reason)
}

// NewInvalidInputErrorf create new InvalidInputError with specified reason string format
// errors.Is(error, InvalidInputError) will return true if a new error is created using this function
func NewInvalidInputErrorf(reasonFormat string, a ...interface{}) error {
	return fmt.Errorf("%w: %s", InvalidInputError, fmt.Sprintf(reasonFormat, a...))
}

// NewDeadlineExceededError create new DeadlineExceededError with specified reason
func NewDeadlineExceededError(reason string) error {
	return fmt.Errorf("%w: %s", DeadlineExceededError, reason)
}

// NewDeadlineExceededErrorf reate new DeadlineExceededError with specified reason string format
func NewDeadlineExceededErrorf(reasonFormat string, a ...interface{}) error {
	return fmt.Errorf("%w: %s", DeadlineExceededError, fmt.Sprintf(reasonFormat, a...))
}
