package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/caraml-dev/merlin/models"
	"github.com/caraml-dev/merlin/pkg/errors"
	internalValidator "github.com/caraml-dev/merlin/pkg/validator"
	"github.com/caraml-dev/merlin/service/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestModelSchemaController_GetAllSchemas(t *testing.T) {
	tests := []struct {
		desc               string
		vars               map[string]string
		modelSchemaService func() *mocks.ModelSchemaService
		expected           *Response
	}{
		{
			desc: "Should success get all schemas",
			vars: map[string]string{
				"model_id": "1",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("List", mock.Anything, models.ID(1)).Return([]*models.ModelSchema{
					{
						ModelID: models.ID(1),
						ID:      models.ID(1),
						Spec: &models.SchemaSpec{
							PredictionIDColumn: "prediction_id",
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Boolean,
								"featureC": models.Int64,
							},
							ModelPredictionOutput: &models.ModelPredictionOutput{
								BinaryClassificationOutput: &models.BinaryClassificationOutput{
									ActualLabelColumn:     "actual_label",
									NegativeClassLabel:    "negative",
									PositiveClassLabel:    "positive",
									PredictionScoreColumn: "prediction_score",
									PredictionLabelColumn: "prediction_label",
								},
							},
						},
					},
					{
						ModelID: models.ID(1),
						ID:      models.ID(2),
						Spec: &models.SchemaSpec{
							PredictionIDColumn: "prediction_id",
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Boolean,
								"featureC": models.Int64,
							},
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									PredictionGroudIDColumn: "session_id",
									RankScoreColumn:         "score",
									RelevanceScoreColumn:    "relevance_score",
								},
							},
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: []*models.ModelSchema{
					{
						ModelID: models.ID(1),
						ID:      models.ID(1),
						Spec: &models.SchemaSpec{
							PredictionIDColumn: "prediction_id",
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Boolean,
								"featureC": models.Int64,
							},
							ModelPredictionOutput: &models.ModelPredictionOutput{
								BinaryClassificationOutput: &models.BinaryClassificationOutput{
									ActualLabelColumn:     "actual_label",
									NegativeClassLabel:    "negative",
									PositiveClassLabel:    "positive",
									PredictionScoreColumn: "prediction_score",
									PredictionLabelColumn: "prediction_label",
								},
							},
						},
					},
					{
						ModelID: models.ID(1),
						ID:      models.ID(2),
						Spec: &models.SchemaSpec{
							PredictionIDColumn: "prediction_id",
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Boolean,
								"featureC": models.Int64,
							},
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									PredictionGroudIDColumn: "session_id",
									RankScoreColumn:         "score",
									RelevanceScoreColumn:    "relevance_score",
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "No schemas found",
			vars: map[string]string{
				"model_id": "1",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("List", mock.Anything, models.ID(1)).Return(nil, errors.NewNotFoundError("model schema with model id 1 is not found"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model schemas not found: not found: model schema with model id 1 is not found"},
			},
		},
		{
			desc: "Error fetching the schemas",
			vars: map[string]string{
				"model_id": "1",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("List", mock.Anything, models.ID(1)).Return(nil, fmt.Errorf("peer connection reset"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error get All schemas with model id: 1 with error: peer connection reset"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctrl := &ModelSchemaController{
				AppContext: &AppContext{
					ModelSchemaService: tt.modelSchemaService(),
				},
			}
			resp := ctrl.GetAllSchemas(&http.Request{}, tt.vars, nil)
			assertEqualResponses(t, tt.expected, resp)
		})
	}
}

func TestModelSchemaController_GetSchema(t *testing.T) {
	tests := []struct {
		desc               string
		vars               map[string]string
		modelSchemaService func() *mocks.ModelSchemaService
		expected           *Response
	}{
		{
			desc: "Should success get schema",
			vars: map[string]string{
				"model_id":  "1",
				"schema_id": "2",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(2), models.ID(1)).Return(&models.ModelSchema{
					ModelID: models.ID(1),
					ID:      models.ID(2),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualLabelColumn:     "actual_label",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.ModelSchema{
					ModelID: models.ID(1),
					ID:      models.ID(2),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualLabelColumn:     "actual_label",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				},
			},
		},
		{
			desc: "No schemas found",
			vars: map[string]string{
				"model_id":  "1",
				"schema_id": "2",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(2), models.ID(1)).Return(nil, errors.NewNotFoundError("model schema with id 2 is not found"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusNotFound,
				data: Error{Message: "Model schema with id: 2 not found: not found: model schema with id 2 is not found"},
			},
		},
		{
			desc: "Error fetching the schemas",
			vars: map[string]string{
				"model_id":  "1",
				"schema_id": "2",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("FindByID", mock.Anything, models.ID(2), models.ID(1)).Return(nil, fmt.Errorf("peer connection reset"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error get schema with id: 2, model id: 1 and error: peer connection reset"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctrl := &ModelSchemaController{
				AppContext: &AppContext{
					ModelSchemaService: tt.modelSchemaService(),
				},
			}
			resp := ctrl.GetSchema(&http.Request{}, tt.vars, nil)
			assertEqualResponses(t, tt.expected, resp)
		})
	}
}

func TestModelSchemaController_CreateOrUpdateSchema(t *testing.T) {
	tests := []struct {
		desc               string
		vars               map[string]string
		body               []byte
		modelSchemaService func() *mocks.ModelSchemaService
		expected           *Response
	}{
		{
			desc: "success create ranking schema",
			vars: map[string]string{
				"model_id": "1",
			},
			body: []byte(`{
				"spec": {
					"prediction_id_column":"prediction_id",
					"tag_columns": ["tags"],
					"feature_types": {
						"featureA": "float64",
						"featureB": "int64",
						"featureC": "boolean"
					},
					"model_prediction_output": {
						"prediction_group_id_column": "session_id",
						"rank_score_column": "score",
						"relevance_score": "relevance_score",
						"output_class": "RankingOutput"
					}
				}
			}`),
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("Save", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								PredictionGroudIDColumn: "session_id",
								RankScoreColumn:         "score",
								RelevanceScoreColumn:    "relevance_score",
								OutputClass:             models.Ranking,
							},
						},
					},
				}).Return(&models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								PredictionGroudIDColumn: "session_id",
								RankScoreColumn:         "score",
								RelevanceScoreColumn:    "relevance_score",
								OutputClass:             models.Ranking,
							},
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								PredictionGroudIDColumn: "session_id",
								RankScoreColumn:         "score",
								RelevanceScoreColumn:    "relevance_score",
								OutputClass:             models.Ranking,
							},
						},
					},
				},
			},
		},
		{
			desc: "success create binary classification schema",
			vars: map[string]string{
				"model_id": "1",
			},
			body: []byte(`{
				"spec": {
					"prediction_id_column":"prediction_id",
					"tag_columns": ["tags"],
					"feature_types": {
						"featureA": "float64",
						"featureB": "int64",
						"featureC": "boolean"
					},
					"model_prediction_output": {
						"actual_label_column": "actual_label",
						"negative_class_label": "negative",
						"prediction_score_column": "prediction_score",
						"prediction_label_column": "prediction_label",
						"positive_class_label": "positive",
						"output_class": "BinaryClassificationOutput"
					}
				}
			}`),
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("Save", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualLabelColumn:     "actual_label",
								NegativeClassLabel:    "negative",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
								PositiveClassLabel:    "positive",
								OutputClass:           models.BinaryClassification,
							},
						},
					},
				}).Return(&models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualLabelColumn:     "actual_label",
								NegativeClassLabel:    "negative",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
								PositiveClassLabel:    "positive",
								OutputClass:           models.BinaryClassification,
							},
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualLabelColumn:     "actual_label",
								NegativeClassLabel:    "negative",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
								PositiveClassLabel:    "positive",
								OutputClass:           models.BinaryClassification,
							},
						},
					},
				},
			},
		},
		{
			desc: "success create regression schema",
			vars: map[string]string{
				"model_id": "1",
			},
			body: []byte(`{
				"spec": {
					"prediction_id_column":"prediction_id",
					"tag_columns": ["tags"],
					"feature_types": {
						"featureA": "float64",
						"featureB": "int64",
						"featureC": "boolean"
					},
					"model_prediction_output": {
						"prediction_score_column": "prediction_score",
						"actual_score_column": "actual_score",
						"output_class": "RegressionOutput"
					}
				}
			}`),
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("Save", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RegressionOutput: &models.RegressionOutput{
								PredictionScoreColumn: "prediction_score",
								ActualScoreColumn:     "actual_score",
								OutputClass:           models.Regression,
							},
						},
					},
				}).Return(&models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RegressionOutput: &models.RegressionOutput{
								PredictionScoreColumn: "prediction_score",
								ActualScoreColumn:     "actual_score",
								OutputClass:           models.Regression,
							},
						},
					},
				}, nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusOK,
				data: &models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RegressionOutput: &models.RegressionOutput{
								PredictionScoreColumn: "prediction_score",
								ActualScoreColumn:     "actual_score",
								OutputClass:           models.Regression,
							},
						},
					},
				},
			},
		},
		{
			desc: "fail to save schema",
			vars: map[string]string{
				"model_id": "1",
			},
			body: []byte(`{
				"spec": {
					"prediction_id_column":"prediction_id",
					"tag_columns": ["tags"],
					"feature_types": {
						"featureA": "float64",
						"featureB": "int64",
						"featureC": "boolean"
					},
					"model_prediction_output": {
						"prediction_group_id_column": "session_id",
						"rank_score_column": "score",
						"relevance_score": "relevance_score",
						"output_class": "RankingOutput"
					}
				}
			}`),
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("Save", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						TagColumns:         []string{"tags"},
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Int64,
							"featureC": models.Boolean,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								PredictionGroudIDColumn: "session_id",
								RankScoreColumn:         "score",
								RelevanceScoreColumn:    "relevance_score",
								OutputClass:             models.Ranking,
							},
						},
					},
				}).Return(nil, fmt.Errorf("peer connection is reset"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error save model schema: peer connection is reset"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctrl := &ModelSchemaController{
				AppContext: &AppContext{
					ModelSchemaService: tt.modelSchemaService(),
				},
			}
			var modelSchema *models.ModelSchema
			err := json.Unmarshal(tt.body, &modelSchema)
			require.NoError(t, err)

			validate, _ := internalValidator.NewValidator()
			err = validate.Struct(modelSchema)
			require.NoError(t, err)

			resp := ctrl.CreateOrUpdateSchema(&http.Request{}, tt.vars, modelSchema)
			assertEqualResponses(t, tt.expected, resp)
		})
	}
}

func Benchmark_Unmarshal(b *testing.B) {
	data := []byte(` {
			"prediction_id_column":"prediction_id",
			"tag_columns": ["tags"],
			"feature_types": {
				"featureA": "float64",
				"featureB": "int64",
				"featureC": "boolean"
			},
			"model_prediction_output": {
				"actual_label_column": "actual_label",
				"negative_class_label": "negative",
				"prediction_score_column": "prediction_score",
				"prediction_label_column": "prediction_label",
				"positive_class_label": "positive",
				"output_class": "BinaryClassificationOutput"
			}
	}`)
	for i := 0; i < b.N; i++ {
		var schemaSpec models.SchemaSpec
		_ = json.Unmarshal(data, &schemaSpec)
	}
}

func TestModelSchemaController_DeleteSchema(t *testing.T) {
	tests := []struct {
		desc               string
		vars               map[string]string
		modelSchemaService func() *mocks.ModelSchemaService
		expected           *Response
	}{
		{
			desc: "Should success get schema",
			vars: map[string]string{
				"model_id":  "1",
				"schema_id": "2",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("Delete", mock.Anything, &models.ModelSchema{ID: models.ID(2), ModelID: models.ID(1)}).Return(nil)
				return mockSvc
			},
			expected: &Response{
				code: http.StatusNoContent,
			},
		},
		{
			desc: "Error deleting the schema",
			vars: map[string]string{
				"model_id":  "1",
				"schema_id": "2",
			},
			modelSchemaService: func() *mocks.ModelSchemaService {
				mockSvc := &mocks.ModelSchemaService{}
				mockSvc.On("Delete", mock.Anything, &models.ModelSchema{ID: models.ID(2), ModelID: models.ID(1)}).Return(fmt.Errorf("peer connection reset"))
				return mockSvc
			},
			expected: &Response{
				code: http.StatusInternalServerError,
				data: Error{Message: "Error delete model schema: peer connection reset"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctrl := &ModelSchemaController{
				AppContext: &AppContext{
					ModelSchemaService: tt.modelSchemaService(),
				},
			}
			resp := ctrl.DeleteSchema(&http.Request{}, tt.vars, nil)
			assertEqualResponses(t, tt.expected, resp)
		})
	}
}
