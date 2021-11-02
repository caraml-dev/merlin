package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidLabel(t *testing.T) {
	testCases := []struct {
		desc        string
		name        string
		expectedErr error
	}{
		{
			desc:        "valid",
			name:        "value",
			expectedErr: nil,
		},
		{
			desc:        "invalid, name greater than 63 characters",
			name:        "orchestrator-orchestrator-orchestrator-orchestrator-orchestrator",
			expectedErr: fmt.Errorf("length of name is greater than 63 characters"),
		},
		{
			desc:        "invalid, using non numerical character",
			name:        "value!",
			expectedErr: fmt.Errorf("name violates kubernetes label constraint"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotErr := IsValidLabel(tC.name)
			assert.Equal(t, tC.expectedErr, gotErr)
		})
	}
}
