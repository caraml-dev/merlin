package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/caraml-dev/merlin/models"
	mErrors "github.com/caraml-dev/merlin/pkg/errors"
	"github.com/caraml-dev/merlin/storage"
	"github.com/caraml-dev/merlin/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func Test_modelSchemaService_List(t *testing.T) {
	type args struct {
		ctx     context.Context
		modelID models.ID
	}
	tests := []struct {
		name     string
		m        storage.ModelSchemaStorage
		args     args
		want     []*models.ModelSchema
		wantErr  bool
		expError error
	}{
		{
			name: "Success list model schemas",
			m: func() storage.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("FindAll", mock.Anything, models.ID(1)).Return([]*models.ModelSchema{
					{
						ModelID: models.ID(1),
						ID:      models.ID(1),
						Spec: &models.SchemaSpec{
							PredictionIDColumn: "prediction_id",
							SessionIDColumn:    "session_id",
							RowIDColumn:        "row_id",
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Boolean,
								"featureC": models.Int64,
							},
							ModelPredictionOutput: &models.ModelPredictionOutput{
								BinaryClassificationOutput: &models.BinaryClassificationOutput{
									ActualScoreColumn:     "actual_score",
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
							SessionIDColumn:    "session_id",
							RowIDColumn:        "row_id",
							FeatureTypes: map[string]models.ValueType{
								"featureA": models.Float64,
								"featureB": models.Boolean,
								"featureC": models.Int64,
							},
							ModelPredictionOutput: &models.ModelPredictionOutput{
								RankingOutput: &models.RankingOutput{
									RankScoreColumn:      "score",
									RelevanceScoreColumn: "relevance_score",
								},
							},
						},
					},
				}, nil)
				return storageMock
			}(),
			args: args{
				ctx:     context.TODO(),
				modelID: models.ID(1),
			},
			want: []*models.ModelSchema{
				{
					ModelID: models.ID(1),
					ID:      models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
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
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							RankingOutput: &models.RankingOutput{
								RankScoreColumn:      "score",
								RelevanceScoreColumn: "relevance_score",
							},
						},
					},
				},
			},
		},
		{
			name: "Not found",
			m: func() storage.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("FindAll", mock.Anything, models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return storageMock
			}(),
			args: args{
				ctx:     context.TODO(),
				modelID: models.ID(1),
			},
			want:     nil,
			wantErr:  true,
			expError: mErrors.NewNotFoundError("model schema with model id: 1 are not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &modelSchemaService{modelSchemaStorage: tt.m}
			got, err := svc.List(tt.args.ctx, tt.args.modelID)
			if tt.wantErr {
				assert.Equal(t, tt.expError, err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("modelSchemaService.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelSchemaService.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelSchemaService_Save(t *testing.T) {
	type args struct {
		ctx         context.Context
		modelSchema *models.ModelSchema
	}
	tests := []struct {
		name    string
		m       storage.ModelSchemaStorage
		args    args
		want    *models.ModelSchema
		wantErr bool
	}{
		{
			name: "Success create new",
			m: func() *mocks.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("Save", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}).Return(&models.ModelSchema{
					ID:      models.ID(1),
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}, nil)
				return storageMock
			}(),
			args: args{
				ctx: context.TODO(),
				modelSchema: &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				},
			},
			want: &models.ModelSchema{
				ID:      models.ID(1),
				ModelID: models.ID(1),
				Spec: &models.SchemaSpec{
					PredictionIDColumn: "prediction_id",
					SessionIDColumn:    "session_id",
					RowIDColumn:        "row_id",
					FeatureTypes: map[string]models.ValueType{
						"featureA": models.Float64,
						"featureB": models.Boolean,
						"featureC": models.Int64,
					},
					ModelPredictionOutput: &models.ModelPredictionOutput{
						BinaryClassificationOutput: &models.BinaryClassificationOutput{
							ActualScoreColumn:     "actual_score",
							NegativeClassLabel:    "negative",
							PositiveClassLabel:    "positive",
							PredictionScoreColumn: "prediction_score",
							PredictionLabelColumn: "prediction_label",
						},
					},
				},
			},
		},
		{
			name: "Failed save model schema",
			m: func() *mocks.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("Save", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}).Return(nil, errors.New("foreign key violation"))
				return storageMock
			}(),
			args: args{
				ctx: context.TODO(),
				modelSchema: &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &modelSchemaService{modelSchemaStorage: tt.m}
			got, err := svc.Save(tt.args.ctx, tt.args.modelSchema)
			if (err != nil) != tt.wantErr {
				t.Errorf("modelSchemaService.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelSchemaService.Save() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modelSchemaService_Delete(t *testing.T) {
	type args struct {
		ctx         context.Context
		modelSchema *models.ModelSchema
	}
	tests := []struct {
		name    string
		m       storage.ModelSchemaStorage
		args    args
		wantErr bool
	}{
		{
			name: "success",
			m: func() *mocks.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("Delete", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}).Return(nil)
				return storageMock
			}(),
			args: args{
				ctx: context.TODO(),
				modelSchema: &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
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
			name: "error",
			m: func() *mocks.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("Delete", mock.Anything, &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}).Return(errors.New("db is disconnected"))
				return storageMock
			}(),
			args: args{
				ctx: context.TODO(),
				modelSchema: &models.ModelSchema{
					ModelID: models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &modelSchemaService{modelSchemaStorage: tt.m}
			if err := svc.Delete(tt.args.ctx, tt.args.modelSchema); (err != nil) != tt.wantErr {
				t.Errorf("modelSchemaService.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_modelSchemaService_FindByID(t *testing.T) {
	type args struct {
		ctx           context.Context
		modelSchemaID models.ID
		modelID       models.ID
	}
	tests := []struct {
		name     string
		m        storage.ModelSchemaStorage
		args     args
		want     *models.ModelSchema
		wantErr  bool
		expError error
	}{
		{
			name: "Success get model schema",
			m: func() storage.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("FindByID", mock.Anything, models.ID(1), models.ID(1)).Return(&models.ModelSchema{
					ModelID: models.ID(1),
					ID:      models.ID(1),
					Spec: &models.SchemaSpec{
						PredictionIDColumn: "prediction_id",
						SessionIDColumn:    "session_id",
						RowIDColumn:        "row_id",
						FeatureTypes: map[string]models.ValueType{
							"featureA": models.Float64,
							"featureB": models.Boolean,
							"featureC": models.Int64,
						},
						ModelPredictionOutput: &models.ModelPredictionOutput{
							BinaryClassificationOutput: &models.BinaryClassificationOutput{
								ActualScoreColumn:     "actual_score",
								NegativeClassLabel:    "negative",
								PositiveClassLabel:    "positive",
								PredictionScoreColumn: "prediction_score",
								PredictionLabelColumn: "prediction_label",
							},
						},
					},
				}, nil)
				return storageMock
			}(),
			args: args{
				ctx:           context.TODO(),
				modelSchemaID: models.ID(1),
				modelID:       models.ID(1),
			},
			want: &models.ModelSchema{
				ModelID: models.ID(1),
				ID:      models.ID(1),
				Spec: &models.SchemaSpec{
					PredictionIDColumn: "prediction_id",
					SessionIDColumn:    "session_id",
					RowIDColumn:        "row_id",
					FeatureTypes: map[string]models.ValueType{
						"featureA": models.Float64,
						"featureB": models.Boolean,
						"featureC": models.Int64,
					},
					ModelPredictionOutput: &models.ModelPredictionOutput{
						BinaryClassificationOutput: &models.BinaryClassificationOutput{
							ActualScoreColumn:     "actual_score",
							NegativeClassLabel:    "negative",
							PositiveClassLabel:    "positive",
							PredictionScoreColumn: "prediction_score",
							PredictionLabelColumn: "prediction_label",
						},
					},
				},
			},
		},
		{
			name: "Schema not found",
			m: func() storage.ModelSchemaStorage {
				storageMock := &mocks.ModelSchemaStorage{}
				storageMock.On("FindByID", mock.Anything, models.ID(1), models.ID(1)).Return(nil, gorm.ErrRecordNotFound)
				return storageMock
			}(),
			args: args{
				ctx:           context.TODO(),
				modelSchemaID: models.ID(1),
				modelID:       models.ID(1),
			},
			want:     nil,
			wantErr:  true,
			expError: mErrors.NewNotFoundError("model schema with id: 1 are not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &modelSchemaService{
				modelSchemaStorage: tt.m,
			}
			got, err := svc.FindByID(tt.args.ctx, tt.args.modelSchemaID, tt.args.modelID)
			if tt.wantErr {
				assert.Equal(t, tt.expError, err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("modelSchemaService.FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("modelSchemaService.FindByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
