package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/feast/mocks"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	transTypes "github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/caraml-dev/merlin/pkg/transformer/types/table"
)

func TestFeastOp_Execute(t *testing.T) {
	sr := symbol.NewRegistry()
	logger, _ := zap.NewDevelopment()

	type mockFeastRetriever struct {
		result []*transTypes.FeatureTable
		error  error
	}
	type args struct {
		context     context.Context
		environment *Environment
	}
	tests := []struct {
		name               string
		mockFeastRetriever mockFeastRetriever
		args               args
		expVariables       map[string]interface{}
		wantErr            bool
	}{
		{
			name: "produce one table",
			mockFeastRetriever: mockFeastRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name: "driver_table",
						Columns: []string{
							"driver_id",
							"feature_a",
							"feature_b",
						},
						Data: transTypes.ValueRows{
							{
								"1111", 11.11, 12,
							},
							{
								"2222", 22.22, 22,
							},
						},
						ColumnTypes: []types.ValueType_Enum{types.ValueType_STRING, types.ValueType_DOUBLE, types.ValueType_INT64},
					},
				},
				error: nil,
			},
			args: args{
				context: context.Background(),
				environment: &Environment{
					symbolRegistry: sr,
					logger:         logger,
				},
			},
			expVariables: map[string]interface{}{
				"driver_table": table.New(
					series.New([]string{"1111", "2222"}, series.String, "driver_id"),
					series.New([]float64{11.11, 22.22}, series.Float, "feature_a"),
					series.New([]float64{12, 22}, series.Int, "feature_b")),
			},
			wantErr: false,
		},
		{
			name: "produce two table",
			mockFeastRetriever: mockFeastRetriever{
				result: []*transTypes.FeatureTable{
					{
						Name: "driver_table",
						Columns: []string{
							"driver_id",
							"feature_a",
							"feature_b",
						},
						Data: transTypes.ValueRows{
							{
								"1111", 11.11, 12,
							},
							{
								"2222", 22.22, 22,
							},
						},
						ColumnTypes: []types.ValueType_Enum{types.ValueType_STRING, types.ValueType_DOUBLE, types.ValueType_INT64},
					},
					{
						Name: "customer_table",
						Columns: []string{
							"driver_id",
							"feature_a",
							"feature_b",
						},
						Data: transTypes.ValueRows{
							{
								"1111", 11.11, 12,
							},
							{
								"2222", 22.22, 22,
							},
						},
						ColumnTypes: []types.ValueType_Enum{types.ValueType_STRING, types.ValueType_DOUBLE, types.ValueType_INT64},
					},
				},
				error: nil,
			},
			args: args{
				context: context.Background(),
				environment: &Environment{
					symbolRegistry: sr,
					logger:         logger,
				},
			},
			expVariables: map[string]interface{}{
				"driver_table": table.New(
					series.New([]string{"1111", "2222"}, series.String, "driver_id"),
					series.New([]float64{11.11, 22.22}, series.Float, "feature_a"),
					series.New([]float64{12, 22}, series.Int, "feature_b")),
				"customer_table": table.New(
					series.New([]string{"1111", "2222"}, series.String, "driver_id"),
					series.New([]float64{11.11, 22.22}, series.Float, "feature_a"),
					series.New([]float64{12, 22}, series.Int, "feature_b")),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := zap.NewDevelopment()
			if err != nil {
				panic(err)
			}

			mockFeastRetriever := &mocks.FeatureRetriever{}
			mockFeastRetriever.On("RetrieveFeatureOfEntityInSymbolRegistry", mock.Anything, tt.args.environment.symbolRegistry).
				Return(tt.mockFeastRetriever.result, tt.mockFeastRetriever.error)

			op := &FeastOp{
				feastRetriever: mockFeastRetriever,
				logger:         logger,
			}
			if err := op.Execute(tt.args.context, tt.args.environment); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			for varName, varValue := range tt.expVariables {
				switch v := varValue.(type) {
				case time.Time:
					assert.True(t, v.Sub(sr[varName].(time.Time)) < time.Second)
				default:
					assert.Equal(t, v, sr[varName])
				}
			}
		})
	}
}
