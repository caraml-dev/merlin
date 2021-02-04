package feast

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	err error
}

func NewValidationError(reason string) error {
	return &ValidationError{err: errors.New(reason)}
}

func (va *ValidationError) Error() string {
	return fmt.Sprintf("invalid transformer config: %s", va.err.Error())
}

func (va *ValidationError) Unwrap() error {
	return va.err
}
