package service

import (
	"context"

	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/log"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/transformer/executor"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/types"
)

const (
	defaultFeastBatchSize = 50
)

// TransformerService handles the standard transformer simulation
type TransformerService interface {
	SimulateTransformer(ctx context.Context, simulationPayload *models.TransformerSimulation) (*types.PredictResponse, error)
}

type transformerService struct {
	cfg config.StandardTransformerConfig
}

// NewTransformerService
func NewTransformerService(cfg config.StandardTransformerConfig) TransformerService {
	return &transformerService{
		cfg: cfg,
	}
}

var _ TransformerService = (*transformerService)(nil)

// SimulateTransformer function will call transformer executor that initialize feast serving and do the call to standard transformer
func (ts *transformerService) SimulateTransformer(ctx context.Context, simulationPayload *models.TransformerSimulation) (*types.PredictResponse, error) {
	transformerExecutor, err := ts.createTransformerExecutor(ctx, simulationPayload)
	if err != nil {
		return nil, err
	}
	return ts.simulate(ctx, transformerExecutor, simulationPayload.Payload, simulationPayload.Headers)
}

func (ts *transformerService) simulate(ctx context.Context, transformer executor.Transformer, requestPayload types.JSONObject, requestHeaders map[string]string) (*types.PredictResponse, error) {
	return transformer.Predict(ctx, requestPayload, requestHeaders)
}

func (ts *transformerService) createTransformerExecutor(ctx context.Context, simulationPayload *models.TransformerSimulation) (executor.Transformer, error) {
	var mockResponseBody types.JSONObject
	var mockRequestHeaders map[string]string

	if simulationPayload.PredictionConfig != nil && simulationPayload.PredictionConfig.Mock != nil {
		if simulationPayload.PredictionConfig.Mock.Body != nil {
			mockResponseBody = simulationPayload.PredictionConfig.Mock.Body
		}
		if simulationPayload.PredictionConfig.Mock.Headers != nil {
			mockRequestHeaders = simulationPayload.PredictionConfig.Mock.Headers
		}
	}

	transformerExecutor, err := executor.NewStandardTransformerWithConfig(
		ctx,
		simulationPayload.Config,
		executor.WithLogger(log.GetLogger()),
		executor.WithTraceEnabled(true),
		executor.WithModelPredictor(executor.NewMockPredictor(mockResponseBody, mockRequestHeaders)),
		executor.WithFeastOptions(feast.Options{
			StorageConfigs:     ts.cfg.ToFeastStorageConfigsForSimulation(),
			DefaultFeastSource: ts.cfg.DefaultFeastSource,
			BatchSize:          defaultFeastBatchSize,
		}),
	)
	if err != nil {
		return nil, err
	}
	return transformerExecutor, nil
}
