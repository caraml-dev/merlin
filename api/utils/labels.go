package utils

import (
	"fmt"
	"regexp"
)

var validLabelRegex = regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")

func IsValidLabel(name string) error {
	lengthOfName := len(name) < 64
	if !(lengthOfName) {
		return fmt.Errorf("length of name is greater than 63 characters")
	}

	if isValidName := validLabelRegex.MatchString(name); !isValidName {
		return fmt.Errorf("name violates kubernetes label constraint")
	}

	return nil
}
