package errors

import (
	"errors"
	"fmt"
)

var (
	InvalidInputError error = errors.New("invalid input")
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
