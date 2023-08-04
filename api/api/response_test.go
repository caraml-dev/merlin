package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Utility function to compare responses without the stacktrace
func assertEqualResponses(t *testing.T, want interface{}, got interface{}) {
	options := []cmp.Option{
		cmp.AllowUnexported(Response{}),
		cmpopts.IgnoreFields(Response{}, "stacktrace"),
	}
	if !cmp.Equal(want, got, options...) {
		t.Errorf("Responses mismatched")
		t.Log(cmp.Diff(want, got, options...))
	}
}
