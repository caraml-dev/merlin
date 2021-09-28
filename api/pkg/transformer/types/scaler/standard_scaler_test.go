package scaler

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/stretchr/testify/assert"
)

func TestStandardScaler_Validate(t *testing.T) {
	testCases := []struct {
		desc          string
		scaler        *StandardScaler
		expectedError error
	}{
		{
			desc: "Valid",
			scaler: &StandardScaler{
				config: &spec.StandardScalerConfig{
					Mean: 3,
					Std:  1.22,
				},
			},
			expectedError: nil,
		},
		{
			desc: "Not Valid, standard deviation must not 0",
			scaler: &StandardScaler{
				config: &spec.StandardScalerConfig{
					Mean: 1,
					Std:  0,
				},
			},
			expectedError: fmt.Errorf("standard scaler require non zero standard deviation"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.scaler.Validate()
			assert.Equal(t, tC.expectedError, got)
		})
	}
}

func TestStandardScaler_Scale(t *testing.T) {
	testCases := []struct {
		desc           string
		scaler         *StandardScaler
		values         []interface{}
		expectedResult interface{}
		expectedError  error
	}{
		{
			desc: "Should scaled correctly",
			scaler: &StandardScaler{
				config: &spec.StandardScalerConfig{
					Mean: 3,
					Std:  2,
				},
			},
			values: []interface{}{
				1, 2, 3,
			},
			expectedResult: []interface{}{
				float64(-1), float64(-0.5), float64(0),
			},
			expectedError: nil,
		},
		{
			desc: "Should failed if values specfied are not integer",
			scaler: &StandardScaler{
				config: &spec.StandardScalerConfig{
					Mean: 3,
					Std:  1.22,
				},
			},
			values: []interface{}{
				"abc", "def", "ghi",
			},
			expectedError: &strconv.NumError{Num: "abc", Err: fmt.Errorf("invalid syntax"), Func: "ParseFloat"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := tC.scaler.Scale(tC.values)
			assert.Equal(t, tC.expectedError, err)
			assert.Equal(t, tC.expectedResult, got)
		})
	}
}
