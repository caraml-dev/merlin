package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidInput     error = errors.New("invalid input")
	ErrDeadlineExceeded error = errors.New("deadline exceeded")
	ErrNotFound         error = errors.New("not found")
)

// NewInvalidInputError create new InvalidInputError with specified reason
// errors.Is(error, InvalidInputError) will return true if a new error is created using this function
func NewInvalidInputError(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, reason)
}

// NewInvalidInputErrorf create new InvalidInputError with specified reason string format
// errors.Is(error, InvalidInputError) will return true if a new error is created using this function
func NewInvalidInputErrorf(reasonFormat string, a ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, fmt.Sprintf(reasonFormat, a...))
}

// NewDeadlineExceededError create new DeadlineExceededError with specified reason
func NewDeadlineExceededError(reason string) error {
	return fmt.Errorf("%w: %s", ErrDeadlineExceeded, reason)
}

// NewDeadlineExceededErrorf create new DeadlineExceededError with specified reason string format
func NewDeadlineExceededErrorf(reasonFormat string, a ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrDeadlineExceeded, fmt.Sprintf(reasonFormat, a...))
}

// NotFoundError create new NotFoundError with specified reason
func NewNotFoundError(reason string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, reason)
}

// NewNotFoundErrorf create new NotFoundError with specified reason string format
func NewNotfoundErrorf(reasonFormat string, a ...interface{}) error {
	return fmt.Errorf("%w: %s", ErrNotFound, fmt.Sprintf(reasonFormat, a...))
}
