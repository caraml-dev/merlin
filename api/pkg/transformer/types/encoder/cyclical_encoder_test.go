package encoder

import (
	"fmt"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewCyclicalEncoder(t *testing.T) {
	testCases := []struct {
		desc            string
		config          *spec.OrdinalEncoderConfig
		expectedEncoder *OrdinalEncoder
		expectedErr     error
	}{
		{
			desc: "Should success - target INT",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "-1",
				TargetValueType: spec.ValueType_INT,
				Mapping: map[string]string{
					"value1": "1",
					"value2": "2",
				},
			},
			expectedEncoder: &OrdinalEncoder{
				DefaultValue: -1,
				Mapping: map[string]interface{}{
					"value1": 1,
					"value2": 2,
				},
			},
		},
		{
			desc: "Should success - target STRING",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "abc",
				TargetValueType: spec.ValueType_STRING,
				Mapping: map[string]string{
					"value1": "--value1",
					"value2": "--value2",
				},
			},
			expectedEncoder: &OrdinalEncoder{
				DefaultValue: "abc",
				Mapping: map[string]interface{}{
					"value1": "--value1",
					"value2": "--value2",
				},
			},
		},
		{
			desc: "Should success - target FLOAT",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "0",
				TargetValueType: spec.ValueType_FLOAT,
				Mapping: map[string]string{
					"value1": "1.0",
					"value2": "2.0",
				},
			},
			expectedEncoder: &OrdinalEncoder{
				DefaultValue: float64(0),
				Mapping: map[string]interface{}{
					"value1": float64(1.0),
					"value2": float64(2.0),
				},
			},
		},
		{
			desc: "Should success - target BOOL",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "false",
				TargetValueType: spec.ValueType_BOOL,
				Mapping: map[string]string{
					"value1": "true",
					"value2": "false",
				},
			},
			expectedEncoder: &OrdinalEncoder{
				DefaultValue: false,
				Mapping: map[string]interface{}{
					"value1": true,
					"value2": false,
				},
			},
		},
		{
			desc: "Should success - target INT, with extra space",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "0 ",
				TargetValueType: spec.ValueType_INT,
				Mapping: map[string]string{
					"value1": "1 ",
					"value2": "2 ",
				},
			},
			expectedEncoder: &OrdinalEncoder{
				DefaultValue: 0,
				Mapping: map[string]interface{}{
					"value1": 1,
					"value2": 2,
				},
			},
		},
		{
			desc: "Should success - target INT, default value nil",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "",
				TargetValueType: spec.ValueType_INT,
				Mapping: map[string]string{
					"value1": "1",
					"value2": "2",
				},
			},
			expectedEncoder: &OrdinalEncoder{
				DefaultValue: nil,
				Mapping: map[string]interface{}{
					"value1": 1,
					"value2": 2,
				},
			},
		},
		{
			desc: "Should return error, when mapping target value is string but target type is INT",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "0",
				TargetValueType: spec.ValueType_INT,
				Mapping: map[string]string{
					"value1": "1one",
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf(`strconv.Atoi: parsing "1one": invalid syntax`),
		},
		{
			desc: "Should return error, when default value is string but target type is INT",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "zero",
				TargetValueType: spec.ValueType_INT,
				Mapping: map[string]string{
					"value1": "1",
					"value2": "2",
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf(`strconv.Atoi: parsing "zero": invalid syntax`),
		},
		{
			desc: "Should return error, when mapping target value is int but target type is boolean",
			config: &spec.OrdinalEncoderConfig{
				DefaultValue:    "false",
				TargetValueType: spec.ValueType_BOOL,
				Mapping: map[string]string{
					"value1": "2",
				},
			},
			expectedEncoder: nil,
			expectedErr:     fmt.Errorf(`strconv.ParseBool: parsing "2": invalid syntax`),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			gotEncoder, err := NewOrdinalEncoder(tC.config)
			if tC.expectedErr != nil {
				assert.EqualError(t, err, tC.expectedErr.Error())
			} else {
				assert.Equal(t, tC.expectedEncoder, gotEncoder)
			}
		})
	}
}

func TestCyclicalEncode(t *testing.T) {
	column := "col"
	testCases := []struct {
		desc           string
		ordinalEncoder *OrdinalEncoder
		reqValues      []interface{}
		expectedResult map[string]interface{}
		expectedError  error
	}{
		{
			desc: "Should success",
			ordinalEncoder: &OrdinalEncoder{
				DefaultValue: 0,
				Mapping: map[string]interface{}{
					"suv":   1,
					"sedan": 2,
					"mpv":   3,
				},
			},
			reqValues: []interface{}{
				"suv", "sedan", "mpv", nil, "sport car",
			},
			expectedResult: map[string]interface{}{
				column: []interface{}{
					1, 2, 3, 0, 0,
				},
			},
		},
		{
			desc: "Should success, orginal value is integer",
			ordinalEncoder: &OrdinalEncoder{
				DefaultValue: nil,
				Mapping: map[string]interface{}{
					"1": 1,
					"2": 2,
					"3": 3,
				},
			},
			reqValues: []interface{}{
				1, 2, 3, nil, 4,
			},
			expectedResult: map[string]interface{}{
				column: []interface{}{
					1, 2, 3, nil, nil,
				},
			},
		},
		{
			desc: "Should success, orginal value is integer",
			ordinalEncoder: &OrdinalEncoder{
				DefaultValue: float64(1),
				Mapping: map[string]interface{}{
					"true":  float64(100),
					"false": float64(0),
				},
			},
			reqValues: []interface{}{
				true, false, nil,
			},
			expectedResult: map[string]interface{}{
				column: []interface{}{
					float64(100),
					float64(0),
					float64(1),
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got, err := tC.ordinalEncoder.Encode(tC.reqValues, column)
			if tC.expectedError != nil {
				assert.EqualError(t, err, tC.expectedError.Error())
			} else {
				assert.Equal(t, tC.expectedResult, got)
			}
		})
	}
}
