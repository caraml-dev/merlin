package model

import (
	"context"

	"github.com/dmitryikh/leaves"
	"github.com/dmitryikh/leaves/mat"
)

type XGBoostModel struct {
	model *leaves.Ensemble
}

func NewXGBoostModel(modelPath string) (*XGBoostModel, error) {
	xgbModel, err := leaves.XGEnsembleFromFile(modelPath, false)
	if err != nil {
		return nil, err
	}
	return &XGBoostModel{
		model: xgbModel,
	}, nil
}

func (xgb *XGBoostModel) Predict(ctx context.Context, payload Request) (*Response, error) {
	nRow := len(payload.Instances)
	nCol := len(payload.Instances[0])
	vals := make([]float64, 0, nRow*nCol)
	for i := 0; i < nRow; i++ {
		for j := 0; j < nCol; j++ {
			val := payload.Instances[i][j]
			vals = append(vals, val)
		}
	}
	predictions := make([]float64, nRow*xgb.model.NOutputGroups())
	dmat, err := mat.DenseMatFromArray(vals, nRow, nCol)
	if err != nil {
		return nil, err
	}
	if err := xgb.model.PredictDense(dmat.Values, dmat.Rows, dmat.Cols, predictions, xgb.model.NEstimators(), 4); err != nil {
		return nil, err
	}
	groupPredictionResults := make([]float64, 0, nRow)
	for i := 0; i < nRow; i++ {
		row := i * xgb.model.NOutputGroups()
		groupClassProbs := predictions[row : row+xgb.model.NOutputGroups()]
		maxProbIndex := -1
		maxProbVal := 0.0
		for j, proVal := range groupClassProbs {
			if maxProbVal < proVal {
				maxProbVal = proVal
				maxProbIndex = j
			}
		}
		groupPredictionResults = append(groupPredictionResults, float64(maxProbIndex))
	}
	result := &Response{
		Predictions: groupPredictionResults,
	}
	return result, nil
}
