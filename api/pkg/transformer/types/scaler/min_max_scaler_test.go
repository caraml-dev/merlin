package scaler

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/stretchr/testify/assert"
)

func TestMinMaxScaler_Validate(t *testing.T) {
	testCases := []struct {
		desc          string
		scaler        *MinMaxScaler
		expectedError error
	}{
		{
			desc: "Valid",
			scaler: &MinMaxScaler{
				config: &spec.MinMaxScalerConfig{
					Min: 1,
					Max: 3,
				},
			},
			expectedError: nil,
		},
		{
			desc: "Not Valid, min and max equal value",
			scaler: &MinMaxScaler{
				config: &spec.MinMaxScalerConfig{
					Min: 1,
					Max: 1,
				},
			},
			expectedError: fmt.Errorf("minmax scaler require different value between min and max"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := tC.scaler.Validate()
			assert.Equal(t, tC.expectedError, got)
		})
	}
}

func TestMinMaxScaler_Scale(t *testing.T) {
	testCases := []struct {
		desc           string
		scaler         *MinMaxScaler
		values         []interface{}
		expectedResult interface{}
		expectedError  error
	}{
		{
			desc: "Should scaled correctly",
			scaler: &MinMaxScaler{
				config: &spec.MinMaxScalerConfig{
					Min: 1,
					Max: 3,
				},
			},
			values: []interface{}{
				1, 2, 3,
			},
			expectedResult: []interface{}{
				float64(0), float64(0.5), float64(1),
			},
			expectedError: nil,
		},
		{
			desc: "Should failed if values specfied are not integer",
			scaler: &MinMaxScaler{
				config: &spec.MinMaxScalerConfig{
					Min: 1,
					Max: 3,
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
