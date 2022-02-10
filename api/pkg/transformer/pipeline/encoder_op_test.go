package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types/encoder"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEncoderOp_Execute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	testCases := []struct {
		desc       string
		specs      []*spec.Encoder
		env        *Environment
		expEncoder map[string]interface{}
		wantErr    error
	}{
		{
			desc: "ordinal encoder config",
			specs: []*spec.Encoder{
				{
					Name: "ordinalEncoder",
					EncoderConfig: &spec.Encoder_OrdinalEncoderConfig{
						OrdinalEncoderConfig: &spec.OrdinalEncoderConfig{
							DefaultValue:    "1",
							TargetValueType: spec.ValueType_INT,
							Mapping: map[string]string{
								"suv":   "1",
								"sedan": "2",
								"mpv":   "3",
							},
						},
					},
				},
			},
			env: &Environment{
				symbolRegistry: symbol.NewRegistry(),
				logger:         logger,
			},
			expEncoder: map[string]interface{}{
				"ordinalEncoder": &encoder.OrdinalEncoder{
					DefaultValue: 1,
					Mapping: map[string]interface{}{
						"suv":   1,
						"sedan": 2,
						"mpv":   3,
					},
				},
			},
		},
		{
			desc: "cyclical encoder config",
			specs: []*spec.Encoder{
				{
					Name: "cyclicalEncoder",
					EncoderConfig: &spec.Encoder_CyclicalEncoderConfig{
						CyclicalEncoderConfig: &spec.CyclicalEncoderConfig{
							EncodeBy: &spec.CyclicalEncoderConfig_ByEpochTime{
								ByEpochTime: &spec.ByEpochTime{
									PeriodType: spec.PeriodType_MONTH,
								},
							},
						},
					},
				},
			},
			env: &Environment{
				symbolRegistry: symbol.NewRegistry(),
				logger:         logger,
			},
			expEncoder: map[string]interface{}{
				"cyclicalEncoder": &encoder.CyclicalEncoder{
					PeriodType: spec.PeriodType_MONTH,
					Min:        0,
					Max:        0,
				},
			},
		},
		{
			desc: "multiple ordinal encoder config",
			specs: []*spec.Encoder{
				{
					Name: "ordinalEncoder1",
					EncoderConfig: &spec.Encoder_OrdinalEncoderConfig{
						OrdinalEncoderConfig: &spec.OrdinalEncoderConfig{
							DefaultValue:    "1",
							TargetValueType: spec.ValueType_INT,
							Mapping: map[string]string{
								"suv":   "1",
								"sedan": "2",
								"mpv":   "3",
							},
						},
					},
				},
				{
					Name: "ordinalEncoder2",
					EncoderConfig: &spec.Encoder_OrdinalEncoderConfig{
						OrdinalEncoderConfig: &spec.OrdinalEncoderConfig{
							DefaultValue:    "unknown",
							TargetValueType: spec.ValueType_STRING,
							Mapping: map[string]string{
								"1": "suv",
								"2": "sedan",
								"3": "mpv",
							},
						},
					},
				},
			},
			env: &Environment{
				symbolRegistry: symbol.NewRegistry(),
				logger:         logger,
			},
			expEncoder: map[string]interface{}{
				"ordinalEncoder1": &encoder.OrdinalEncoder{
					DefaultValue: 1,
					Mapping: map[string]interface{}{
						"suv":   1,
						"sedan": 2,
						"mpv":   3,
					},
				},
				"ordinalEncoder2": &encoder.OrdinalEncoder{
					DefaultValue: "unknown",
					Mapping: map[string]interface{}{
						"1": "suv",
						"2": "sedan",
						"3": "mpv",
					},
				},
			},
		},
		{
			desc: "no encoder config",
			specs: []*spec.Encoder{
				{
					Name: "ordinalEncoder1",
				},
			},
			env: &Environment{
				symbolRegistry: symbol.NewRegistry(),
				logger:         logger,
			},
			wantErr: fmt.Errorf("encoder spec have unexpected type <nil>"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			op := EncoderOp{
				encoderSpecs: tC.specs,
			}
			gotErr := op.Execute(context.Background(), tC.env)
			assert.Equal(t, tC.wantErr, gotErr)
			for name, value := range tC.expEncoder {
				assert.Equal(t, value, tC.env.symbolRegistry[name])
			}
		})
	}
}
