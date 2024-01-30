package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// ModelPredictionOutputClass is type for kinds of model type
type ModelPredictionOutputClass string

const (
	BinaryClassification ModelPredictionOutputClass = "BinaryClassificationOutput"
	Regression           ModelPredictionOutputClass = "RegressionOutput"
	Ranking              ModelPredictionOutputClass = "RankingOutput"
)

// Value type is type that represent type of the value
type ValueType string

const (
	Float64 ValueType = "float64"
	Int64   ValueType = "int64"
	Boolean ValueType = "boolean"
	String  ValueType = "string"
)

// ModelSchema
type ModelSchema struct {
	ID      ID          `json:"id"`
	Spec    *SchemaSpec `json:"spec,omitempty"`
	ModelID ID          `json:"model_id"`
}

// SchemaSpec
type SchemaSpec struct {
	PredictionIDColumn    string                 `json:"prediction_id_column"`
	ModelPredictionOutput *ModelPredictionOutput `json:"model_prediction_output"`
	TagColumns            []string               `json:"tag_columns"`
	FeatureTypes          map[string]ValueType   `json:"feature_types"`
}

// Value returning a value for `SchemaSpec` instance
// This is required to be implemented when this instance is treated as JSONB column
func (s SchemaSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan returning error when assigning value from db driver is failing
// This is required to be implemented when this instance is treated as JSONB column
func (s *SchemaSpec) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

type ModelPredictionOutput struct {
	BinaryClassificationOutput *BinaryClassificationOutput
	RankingOutput              *RankingOutput
	RegressionOutput           *RegressionOutput
}

func newStrictDecoder(data []byte) *json.Decoder {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.DisallowUnknownFields()
	return dec
}

// UnmarshalJSON custom deserialization of bytes into `ModelPredictionOutput`
func (m *ModelPredictionOutput) UnmarshalJSON(data []byte) error {
	var err error
	outputClassStruct := struct {
		OutputClass ModelPredictionOutputClass `json:"output_class"`
	}{}
	err = json.Unmarshal(data, &outputClassStruct)
	if err != nil {
		return err
	}

	strictDecoder := newStrictDecoder(data)
	switch outputClassStruct.OutputClass {
	case BinaryClassification:
		err := strictDecoder.Decode(&m.BinaryClassificationOutput)
		if err != nil {
			return err
		}
	case Regression:
		err := strictDecoder.Decode(&m.RegressionOutput)
		if err != nil {
			return err
		}
	case Ranking:
		err := strictDecoder.Decode(&m.RankingOutput)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("output class %v it not supported", outputClassStruct.OutputClass)
	}

	return nil
}

// MarshalJSON custom serialization of `ModelPredictionOutput` into json byte
func (m ModelPredictionOutput) MarshalJSON() ([]byte, error) {
	if m.BinaryClassificationOutput != nil {
		return json.Marshal(&m.BinaryClassificationOutput)
	}

	if m.RankingOutput != nil {
		return json.Marshal(&m.RankingOutput)
	}

	if m.RegressionOutput != nil {
		return json.Marshal(&m.RegressionOutput)
	}

	return nil, nil
}

// BinaryClassificationOutput is specification for prediction of binary classification model
type BinaryClassificationOutput struct {
	ActualLabelColumn     string                     `json:"actual_label_column"`
	NegativeClassLabel    string                     `json:"negative_class_label"`
	PredictionScoreColumn string                     `json:"prediction_score_column"`
	PredictionLabelColumn string                     `json:"prediction_label_column"`
	PositiveClassLabel    string                     `json:"positive_class_label"`
	ScoreThreshold        *float64                   `json:"score_threshold,omitempty"`
	OutputClass           ModelPredictionOutputClass `json:"output_class" validate:"required"`
}

// RankingOutput is specification for prediction of ranking model
type RankingOutput struct {
	PredictionGroudIDColumn string                     `json:"prediction_group_id_column"`
	RankScoreColumn         string                     `json:"rank_score_column"`
	RelevanceScoreColumn    string                     `json:"relevance_score"`
	OutputClass             ModelPredictionOutputClass `json:"output_class" validate:"required"`
}

// Regression is specification for prediction of regression model
type RegressionOutput struct {
	PredictionScoreColumn string                     `json:"prediction_score_column"`
	ActualScoreColumn     string                     `json:"actual_score_column"`
	OutputClass           ModelPredictionOutputClass `json:"output_class" validate:"required"`
}
