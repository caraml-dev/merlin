package service

import (
	"context"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/config"
	"github.com/gojek/merlin/models"
	"github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/executor"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/pkg/errors"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

const (
	defaultFeastBatchSize = 50
)

// TransformerService handles the standard transformer simulation
type TransformerService interface {
	// SimulateTransformer simulate transformer execution including preprocessing and postprocessing
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
		return nil, fmt.Errorf("failed creating transformer executor: %w", err)
	}

	if simulationPayload.Protocol == protocol.HttpJson {
		return ts.simulate(ctx, transformerExecutor, simulationPayload.Payload, simulationPayload.Headers)
	}

	var requestPayload upiv1.PredictValuesRequest
	if err := mapstructure.Decode(simulationPayload.Payload, &requestPayload); err != nil {
		return nil, errors.Wrap(err, "request is not valid, user should specifies request with UPI PredictValuesRequest type")
	}

	payload := (*types.UPIPredictionRequest)(&requestPayload)
	return ts.simulate(ctx, transformerExecutor, payload, simulationPayload.Headers)
}

func (ts *transformerService) createTransformerExecutor(ctx context.Context, simulationPayload *models.TransformerSimulation) (executor.Transformer, error) {
	var mockModelResponseBody types.JSONObject
	var mockModelRequestHeaders map[string]string

	if simulationPayload.PredictionConfig != nil && simulationPayload.PredictionConfig.Mock != nil {
		if simulationPayload.PredictionConfig.Mock.Body != nil {
			mockModelResponseBody = simulationPayload.PredictionConfig.Mock.Body
		}

		if simulationPayload.PredictionConfig.Mock.Headers != nil {
			mockModelRequestHeaders = simulationPayload.PredictionConfig.Mock.Headers
		}
	}

	// logger := log.GetLogger()
	logger, _ := zap.NewDevelopment()

	transformerExecutor, err := executor.NewStandardTransformerWithConfig(
		ctx,
		simulationPayload.Config,
		executor.WithLogger(logger),
		executor.WithTraceEnabled(true),
		executor.WithModelPredictor(executor.NewMockModelPredictor(mockModelResponseBody, mockModelRequestHeaders, simulationPayload.Protocol)),
		executor.WithFeastOptions(feast.Options{
			StorageConfigs:     ts.cfg.ToFeastStorageConfigsForSimulation(),
			DefaultFeastSource: ts.cfg.DefaultFeastSource,
			BatchSize:          defaultFeastBatchSize,
		}),
		executor.WithProtocol(simulationPayload.Protocol),
	)
	if err != nil {
		return nil, fmt.Errorf("failed initializing standard transformer: %w", err)
	}

	return transformerExecutor, nil
}

func (ts *transformerService) simulate(ctx context.Context, transformer executor.Transformer, requestPayload types.Payload, requestHeaders map[string]string) (*types.PredictResponse, error) {
	return transformer.Execute(ctx, requestPayload, requestHeaders)
}
