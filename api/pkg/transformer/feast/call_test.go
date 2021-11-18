package feast

import (
	"context"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	feastmocks "github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/gojek/merlin/pkg/transformer/spec"
	transTypes "github.com/gojek/merlin/pkg/transformer/types"
)

func TestCall_do(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	type mockFeastCall struct {
		request  *feast.OnlineFeaturesRequest
		response *feast.OnlineFeaturesResponse
	}

	type fields struct {
		featureTableSpec        *spec.FeatureTable
		columns                 []string
		entitySet               map[string]bool
		defaultValues           defaultValues
		feastClient             feast.Client
		feastURL                string
		logger                  *zap.Logger
		statusMonitoringEnabled bool
		valueMonitoringEnabled  bool
		servingSource           spec.ServingSource
	}

	type args struct {
		ctx        context.Context
		entityList []feast.Row
		features   []string
	}

	featureTableSpec := &spec.FeatureTable{
		Project: "default",
		Entities: []*spec.Entity{
			{
				Name:      "entity_1",
				ValueType: "STRING",
			},
			{
				Name:      "entity_2",
				ValueType: "STRING",
			},
		},
		Features: []*spec.Feature{
			{
				Name:      "feature_1",
				ValueType: "INT64",
			},
			{
				Name:      "feature_2",
				ValueType: "INT64",
			},
			{
				Name:      "feature_3",
				ValueType: "INT64",
			},
			{
				Name:      "feature_4",
				ValueType: "INT64",
			},
		},
		TableName:  "my-table",
		ServingUrl: "localhost:6565",
		Source:     spec.ServingSource_BIGTABLE,
	}

	columns := []string{
		"entity_1",
		"entity_2",
		"feature_1",
		"feature_2",
		"feature_3",
		"feature_4",
	}

	entitySet := map[string]bool{
		"entity_1": true,
		"entity_2": true,
	}

	project := "default"
	defValues := defaultValues{}
	defValues.SetDefaultValue(project, "feature_1", feast.Int64Val(1))
	defValues.SetDefaultValue(project, "feature_2", feast.Int64Val(2))
	defValues.SetDefaultValue(project, "feature_3", feast.Int64Val(3))
	defValues.SetDefaultValue(project, "feature_4", feast.Int64Val(4))

	tests := []struct {
		name          string
		fields        fields
		args          args
		mockFeastCall mockFeastCall
		want          callResult
	}{
		{
			name: "all feature values are present",
			fields: fields{
				featureTableSpec:        featureTableSpec,
				columns:                 columns,
				entitySet:               entitySet,
				defaultValues:           defValues,
				feastClient:             &feastmocks.Client{},
				feastURL:                "localhost:6565",
				logger:                  logger,
				statusMonitoringEnabled: true,
				valueMonitoringEnabled:  true,
			},
			args: args{
				ctx: context.Background(),
				entityList: []feast.Row{
					{
						"entity_1": feast.StrVal("1001"),
						"entity_2": feast.StrVal("1002"),
					},
					{
						"entity_1": feast.StrVal("2001"),
						"entity_2": feast.StrVal("2002"),
					},
				},
				features: []string{
					"feature_1",
					"feature_2",
					"feature_3",
					"feature_4",
				},
			},
			mockFeastCall: mockFeastCall{
				request: &feast.OnlineFeaturesRequest{
					Project: project, // used as identifier for mocking. must match config
				},
				response: &feast.OnlineFeaturesResponse{
					RawResponse: &serving.GetOnlineFeaturesResponse{
						FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
							{
								Fields: map[string]*types.Value{
									"entity_1":  feast.StrVal("1001"),
									"entity_2":  feast.StrVal("1002"),
									"feature_1": feast.Int64Val(1111),
									"feature_2": feast.Int64Val(2222),
									"feature_3": feast.Int64Val(3333),
									"feature_4": feast.Int64Val(4444),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"entity_1":  serving.GetOnlineFeaturesResponse_PRESENT,
									"entity_2":  serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_3": serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_4": serving.GetOnlineFeaturesResponse_PRESENT,
								},
							},
							{
								Fields: map[string]*types.Value{
									"entity_1":  feast.StrVal("2001"),
									"entity_2":  feast.StrVal("2002"),
									"feature_1": feast.Int64Val(5555),
									"feature_2": feast.Int64Val(6666),
									"feature_3": feast.Int64Val(7777),
									"feature_4": feast.Int64Val(8888),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"entity_1":  serving.GetOnlineFeaturesResponse_PRESENT,
									"entity_2":  serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_1": serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_2": serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_3": serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_4": serving.GetOnlineFeaturesResponse_PRESENT,
								},
							},
						},
					},
				},
			},
			want: callResult{
				tableName: "my-table",
				featureTable: &internalFeatureTable{
					entities: []feast.Row{
						{
							"entity_1": feast.StrVal("1001"),
							"entity_2": feast.StrVal("1002"),
						},
						{
							"entity_1": feast.StrVal("2001"),
							"entity_2": feast.StrVal("2002"),
						},
					},
					columnNames: []string{
						"entity_1",
						"entity_2",
						"feature_1",
						"feature_2",
						"feature_3",
						"feature_4",
					},
					columnTypes: []types.ValueType_Enum{
						types.ValueType_STRING,
						types.ValueType_STRING,
						types.ValueType_INT64,
						types.ValueType_INT64,
						types.ValueType_INT64,
						types.ValueType_INT64,
					},
					valueRows: transTypes.ValueRows{
						transTypes.ValueRow{
							"1001", "1002", int64(1111), int64(2222), int64(3333), int64(4444),
						},
						transTypes.ValueRow{
							"2001", "2002", int64(5555), int64(6666), int64(7777), int64(8888),
						},
					},
				},
				err: nil,
			},
		},
		{
			name: "all feature values are not present and no default values",
			fields: fields{
				featureTableSpec:        featureTableSpec,
				columns:                 columns,
				entitySet:               entitySet,
				defaultValues:           defaultValues{},
				feastClient:             &feastmocks.Client{},
				feastURL:                "localhost:6565",
				logger:                  logger,
				statusMonitoringEnabled: true,
				valueMonitoringEnabled:  true,
			},
			args: args{
				ctx: context.Background(),
				entityList: []feast.Row{
					{
						"entity_1": feast.StrVal("1001"),
						"entity_2": feast.StrVal("1002"),
					},
					{
						"entity_1": feast.StrVal("2001"),
						"entity_2": feast.StrVal("2002"),
					},
				},
				features: []string{
					"feature_1",
					"feature_2",
					"feature_3",
					"feature_4",
				},
			},
			mockFeastCall: mockFeastCall{
				request: &feast.OnlineFeaturesRequest{
					Project: "default", // used as identifier for mocking. must match config
				},
				response: &feast.OnlineFeaturesResponse{
					RawResponse: &serving.GetOnlineFeaturesResponse{
						FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
							{
								Fields: map[string]*types.Value{
									"entity_1":  feast.StrVal("1001"),
									"entity_2":  feast.StrVal("1002"),
									"feature_1": feast.Int64Val(1111),
									"feature_2": feast.Int64Val(2222),
									"feature_3": feast.Int64Val(3333),
									"feature_4": feast.Int64Val(4444),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"entity_1":  serving.GetOnlineFeaturesResponse_PRESENT,
									"entity_2":  serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_1": serving.GetOnlineFeaturesResponse_NOT_FOUND,
									"feature_2": serving.GetOnlineFeaturesResponse_NULL_VALUE,
									"feature_3": serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE,
									"feature_4": serving.GetOnlineFeaturesResponse_NOT_FOUND,
								},
							},
							{
								Fields: map[string]*types.Value{
									"entity_1":  feast.StrVal("2001"),
									"entity_2":  feast.StrVal("2002"),
									"feature_1": feast.Int64Val(5555),
									"feature_2": feast.Int64Val(6666),
									"feature_3": feast.Int64Val(7777),
									"feature_4": feast.Int64Val(8888),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"entity_1":  serving.GetOnlineFeaturesResponse_PRESENT,
									"entity_2":  serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_1": serving.GetOnlineFeaturesResponse_NOT_FOUND,
									"feature_2": serving.GetOnlineFeaturesResponse_NULL_VALUE,
									"feature_3": serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE,
									"feature_4": serving.GetOnlineFeaturesResponse_NOT_FOUND,
								},
							},
						},
					},
				},
			},
			want: callResult{
				tableName: "my-table",
				featureTable: &internalFeatureTable{
					entities: []feast.Row{
						{
							"entity_1": feast.StrVal("1001"),
							"entity_2": feast.StrVal("1002"),
						},
						{
							"entity_1": feast.StrVal("2001"),
							"entity_2": feast.StrVal("2002"),
						},
					},
					columnNames: []string{
						"entity_1",
						"entity_2",
						"feature_1",
						"feature_2",
						"feature_3",
						"feature_4",
					},
					columnTypes: []types.ValueType_Enum{
						types.ValueType_STRING,
						types.ValueType_STRING,
						types.ValueType_INVALID,
						types.ValueType_INVALID,
						types.ValueType_INVALID,
						types.ValueType_INVALID,
					},
					valueRows: transTypes.ValueRows{
						transTypes.ValueRow{
							"1001", "1002", nil, nil, nil, nil,
						},
						transTypes.ValueRow{
							"2001", "2002", nil, nil, nil, nil,
						},
					},
				},
				err: nil,
			},
		},
		{
			name: "all feature values are not present but have default values",
			fields: fields{
				featureTableSpec:        featureTableSpec,
				columns:                 columns,
				entitySet:               entitySet,
				defaultValues:           defValues,
				feastClient:             &feastmocks.Client{},
				feastURL:                "localhost:6565",
				logger:                  logger,
				statusMonitoringEnabled: true,
				valueMonitoringEnabled:  true,
			},
			args: args{
				ctx: context.Background(),
				entityList: []feast.Row{
					{
						"entity_1": feast.StrVal("1001"),
						"entity_2": feast.StrVal("1002"),
					},
					{
						"entity_1": feast.StrVal("2001"),
						"entity_2": feast.StrVal("2002"),
					},
				},
				features: []string{
					"feature_1",
					"feature_2",
					"feature_3",
					"feature_4",
				},
			},
			mockFeastCall: mockFeastCall{
				request: &feast.OnlineFeaturesRequest{
					Project: "default", // used as identifier for mocking. must match config
				},
				response: &feast.OnlineFeaturesResponse{
					RawResponse: &serving.GetOnlineFeaturesResponse{
						FieldValues: []*serving.GetOnlineFeaturesResponse_FieldValues{
							{
								Fields: map[string]*types.Value{
									"entity_1":  feast.StrVal("1001"),
									"entity_2":  feast.StrVal("1002"),
									"feature_1": feast.Int64Val(1111),
									"feature_2": feast.Int64Val(2222),
									"feature_3": feast.Int64Val(3333),
									"feature_4": feast.Int64Val(4444),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"entity_1":  serving.GetOnlineFeaturesResponse_PRESENT,
									"entity_2":  serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_1": serving.GetOnlineFeaturesResponse_NOT_FOUND,
									"feature_2": serving.GetOnlineFeaturesResponse_NULL_VALUE,
									"feature_3": serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE,
									"feature_4": serving.GetOnlineFeaturesResponse_NOT_FOUND,
								},
							},
							{
								Fields: map[string]*types.Value{
									"entity_1":  feast.StrVal("2001"),
									"entity_2":  feast.StrVal("2002"),
									"feature_1": feast.Int64Val(5555),
									"feature_2": feast.Int64Val(6666),
									"feature_3": feast.Int64Val(7777),
									"feature_4": feast.Int64Val(8888),
								},
								Statuses: map[string]serving.GetOnlineFeaturesResponse_FieldStatus{
									"entity_1":  serving.GetOnlineFeaturesResponse_PRESENT,
									"entity_2":  serving.GetOnlineFeaturesResponse_PRESENT,
									"feature_1": serving.GetOnlineFeaturesResponse_NOT_FOUND,
									"feature_2": serving.GetOnlineFeaturesResponse_NULL_VALUE,
									"feature_3": serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE,
									"feature_4": serving.GetOnlineFeaturesResponse_NOT_FOUND,
								},
							},
						},
					},
				},
			},
			want: callResult{
				tableName: "my-table",
				featureTable: &internalFeatureTable{
					entities: []feast.Row{
						{
							"entity_1": feast.StrVal("1001"),
							"entity_2": feast.StrVal("1002"),
						},
						{
							"entity_1": feast.StrVal("2001"),
							"entity_2": feast.StrVal("2002"),
						},
					},
					columnNames: []string{
						"entity_1",
						"entity_2",
						"feature_1",
						"feature_2",
						"feature_3",
						"feature_4",
					},
					columnTypes: []types.ValueType_Enum{
						types.ValueType_STRING,
						types.ValueType_STRING,
						types.ValueType_INT64,
						types.ValueType_INT64,
						types.ValueType_INT64,
						types.ValueType_INT64,
					},
					valueRows: transTypes.ValueRows{
						transTypes.ValueRow{
							"1001", "1002", int64(1), int64(2), int64(3), int64(4),
						},
						transTypes.ValueRow{
							"2001", "2002", int64(1), int64(2), int64(3), int64(4),
						},
					},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &call{
				featureTableSpec:        tt.fields.featureTableSpec,
				columns:                 tt.fields.columns,
				entitySet:               tt.fields.entitySet,
				defaultValues:           tt.fields.defaultValues,
				feastClient:             tt.fields.feastClient,
				servingSource:           tt.fields.servingSource,
				logger:                  tt.fields.logger,
				statusMonitoringEnabled: tt.fields.statusMonitoringEnabled,
				valueMonitoringEnabled:  tt.fields.valueMonitoringEnabled,
			}

			fc.feastClient.(*feastmocks.Client).
				On("GetOnlineFeatures", mock.Anything, mock.MatchedBy(func(req *feast.OnlineFeaturesRequest) bool {
					return req.Project == tt.mockFeastCall.request.Project
				})).Return(tt.mockFeastCall.response, nil)

			got := fc.do(tt.args.ctx, tt.args.entityList, tt.args.features)
			assert.Equal(t, tt.want, got)
		})
	}
}
