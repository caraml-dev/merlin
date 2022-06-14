package executor

import (
	"context"
	"encoding/json"

	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/pipeline"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type standardTransformer struct {
	compiledPipeline *pipeline.CompiledPipeline
	modelPredictor   ModelPredictor
	executorConfig   transformerExecutorConfig
	logger           *zap.Logger
}

// Transformer have predict function that process all the preprocess, model prediction and postproces
type Transformer interface {
	Execute(ctx context.Context, requestBody types.JSONObject, requestHeaders map[string]string) (*types.PredictResponse, error)
}

type transformerExecutorConfig struct {
	traceEnabled         bool
	transformerConfig    *spec.StandardTransformerConfig
	featureTableMetadata []*spec.FeatureTableMetadata
	feastOpts            feast.Options
	logger               *zap.Logger
	modelPredictor       ModelPredictor
}

// NewStandardTransformerWithConfig initialize standard transformer executor object
func NewStandardTransformerWithConfig(ctx context.Context, transformerConfig *spec.StandardTransformerConfig, opts ...TransformerOptions) (Transformer, error) {
	defaultModelPredictor := newEchoMockPredictor()
	executorConfig := &transformerExecutorConfig{
		transformerConfig: transformerConfig,
		modelPredictor:    defaultModelPredictor,
		feastOpts:         feast.Options{},
	}
	for _, opt := range opts {
		opt(executorConfig)
	}

	feastServingClients, err := feast.InitFeastServingClients(executorConfig.feastOpts, executorConfig.featureTableMetadata, transformerConfig)
	if err != nil {
		return nil, err
	}

	compiler := pipeline.NewCompiler(symbol.NewRegistry(), feastServingClients, &executorConfig.feastOpts, executorConfig.logger, executorConfig.traceEnabled)
	compiledPipeline, err := compiler.Compile(transformerConfig)
	if err != nil {
		return nil, err
	}

	return &standardTransformer{
		compiledPipeline: compiledPipeline,
		modelPredictor:   executorConfig.modelPredictor,
		executorConfig:   *executorConfig,
		logger:           executorConfig.logger,
	}, nil
}

// Predict will process all standard transformer request including preprocessing, model prediction and postprocess
func (st *standardTransformer) Execute(ctx context.Context, requestBody types.JSONObject, requestHeaders map[string]string) (*types.PredictResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "st.Execute")
	defer span.Finish()

	preprocessOut := requestBody
	st.logger.Debug("raw request_body", zap.Any("request_body", requestBody))

	env := pipeline.NewEnvironment(st.compiledPipeline, st.executorConfig.logger)

	var err error
	if env.IsPreprocessOpExist() {
		preprocessOut, err = env.Preprocess(ctx, requestBody, requestHeaders)
		if err != nil {
			return generateErrorResponse(err), err
		}
		st.logger.Debug("preprocess response", zap.Any("preprocess_response", preprocessOut))
	}

	reqBody, err := json.Marshal(preprocessOut)
	if err != nil {
		return generateErrorResponse(err), err
	}

	predictorRespBody, predictorRespHeaders, err := st.modelPredictor.ModelPrediction(ctx, reqBody, requestHeaders)
	if err != nil {
		return generateErrorResponse(err), err
	}

	st.logger.Debug("predictor response", zap.Any("predict_response", predictorRespBody))

	predictionOut := predictorRespBody
	if env.IsPostProcessOpExist() {
		predictionOut, err = env.Postprocess(ctx, predictionOut, predictorRespHeaders)
		if err != nil {
			return generateErrorResponse(err), err
		}
		st.logger.Debug("postprocess response", zap.Any("postprocess_response", predictionOut))
	}

	resp := &types.PredictResponse{
		Response: predictionOut,
	}

	if st.executorConfig.traceEnabled {
		resp.Tracing = &types.OperationTracing{
			PreprocessTracing:  make([]types.TracingDetail, 0),
			PostprocessTracing: make([]types.TracingDetail, 0),
		}

		if env.IsPreprocessOpExist() {
			resp.Tracing.PreprocessTracing = env.PreprocessTracingDetail()
		}
		if env.IsPostProcessOpExist() {
			resp.Tracing.PostprocessTracing = env.PostprocessTracingDetail()
		}
		st.logger.Debug("executor tracing", zap.Any("tracing_details", resp.Tracing))
	}

	return resp, nil
}

func generateErrorResponse(err error) *types.PredictResponse {
	return &types.PredictResponse{
		Response: types.JSONObject{
			"error": err.Error(),
		},
	}
}
