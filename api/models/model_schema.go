package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type InferenceType string

const (
	BinaryClassification     = "BINARY_CLASSIFICATION"
	MulticlassClassification = "MULTICLASS_CLASSIFICATION"
	Regression               = "REGRESSION"
	Ranking                  = "RANKING"
)

type ValueType string

const (
	Float64 ValueType = "float64"
	Int64   ValueType = "int64"
	Boolean ValueType = "boolean"
	String  ValueType = "string"
)

type ModelSchema struct {
	ID      ID          `json:"id"`
	Spec    *SchemaSpec `json:"spec,omitempty"`
	ModelID ID          `json:"-"`
}

type SchemaSpec struct {
	PredictionIDColumn    string                 `json:"prediction_id_column"`
	ModelPredictionOutput *ModelPredictionOutput `json:"model_prediction_output"`
	TagColumns            []string               `json:"tag_columns"`
	FeatureTypes          map[string]ValueType   `json:"feature_types"`
}

func (s SchemaSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

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

func (m *ModelPredictionOutput) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into BinaryClassificationOutput
	err = newStrictDecoder(data).Decode(&m.BinaryClassificationOutput)
	if err == nil {
		jsonBinaryClassificationOutput, _ := json.Marshal(m.BinaryClassificationOutput)
		if string(jsonBinaryClassificationOutput) == "{}" { // empty struct
			m.BinaryClassificationOutput = nil
		} else {
			match++
		}
	} else {
		m.BinaryClassificationOutput = nil
	}

	// try to unmarshal data into RankingOutput
	err = newStrictDecoder(data).Decode(&m.RankingOutput)
	if err == nil {
		jsonRankingOutput, _ := json.Marshal(m.RankingOutput)
		if string(jsonRankingOutput) == "{}" { // empty struct
			m.RankingOutput = nil
		} else {
			match++
		}
	} else {
		m.RankingOutput = nil
	}

	// try to unmarshal data into RegresionOutput
	err = newStrictDecoder(data).Decode(&m.RegressionOutput)
	if err == nil {
		jsonRegressionOutput, _ := json.Marshal(m.RegressionOutput)
		if string(jsonRegressionOutput) == "{}" { // empty struct
			m.RegressionOutput = nil
		} else {
			match++
		}
	} else {
		m.RegressionOutput = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		m.BinaryClassificationOutput = nil
		m.RankingOutput = nil
		m.RegressionOutput = nil
		return fmt.Errorf("data matches more than one schema in oneOf(ModelPredictionOutput)")
	} else if match == 1 {
		return nil // exactly one match
	}

	return fmt.Errorf("data failed to match schemas in oneOf(ModelPredictionOutput)")
}

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

type BinaryClassificationOutput struct {
	ActualLabelColumn     string   `json:"actual_label_column"`
	NegativeClassLabel    string   `json:"negative_class_label"`
	PredictionScoreColumn string   `json:"prediction_score_column"`
	PredictionLabelColumn string   `json:"prediction_label_column"`
	PositiveClassLabel    string   `json:"positive_class_label"`
	ScoreThreshold        *float64 `json:"score_threshold,omitempty"`
}

type RankingOutput struct {
	PredictionGroudIDColumn string `json:"prediction_group_id_column"`
	RankScoreColumn         string `json:"rank_score_column"`
	RelevanceScoreColumn    string `json:"relevance_score"`
}

type RegressionOutput struct {
	PredictionScoreColumn string `json:"prediction_score_column"`
	ActualScoreColumn     string `json:"actual_score_column"`
}
