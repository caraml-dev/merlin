package models

type InferenceType string

const (
	BinaryClassification     = "BINARY_CLASSIFICATION"
	MulticlassClassification = "MULTICLASS_CLASSIFICATION"
	Regression               = "REGRESSION"
	Ranking                  = "RANKING"
)

type InferenceSchema struct {
	Type InferenceType `json:"type"`
}

type ClassMapping struct {
	ClassName string `json:"class_name"`
	Value     *int   `json:"value"`
}

type BinaryClassificationOutput struct {
	PredictionScoreColumn string         `json:"prediction_score_column,omitempty"`
	PredictionLabelColumn string         `json:"prediction_label_column,omitempty"`
	ActualLabelColumn     string         `json:"actual_label_column,omitempty"`
	ClassOrder            []ClassMapping `json:"class_order,omitempty"`
	Threshold             *float64       `json:"threshold,omitempty"`
}

type MulticlassClassificationOutput struct {
	PredictionScoreColumns []string
}

type RankingOutput struct {
	PredictionGroudIDColumn string         `json:"prediction_group_id_column,omitempty"`
	PredictionScoreColumn   string         `json:"prediction_score_column,omitempty"`
	ClassOrder              []ClassMapping `json:"class_order,omitempty"`
	RelevanceScore          *float64       `json:"relevance_score,omitempty"`
}
