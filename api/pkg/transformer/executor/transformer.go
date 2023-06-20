package executor

import (
	"context"
	"encoding/json"
	"fmt"

	prt "github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/pipeline"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type standardTransformer struct {
	compiledPipeline *pipeline.CompiledPipeline
	modelPredictor   ModelPredictor
	executorConfig   transformerExecutorConfig
	logger           *zap.Logger
}

// Transformer have predict function that process all the preprocess, model prediction and postproces
type Transformer interface {
	Execute(ctx context.Context, requestBody types.JSONObject, requestHeaders map[string]string) *types.PredictResponse
}

type transformerExecutorConfig struct {
	traceEnabled         bool
	transformerConfig    *spec.StandardTransformerConfig
	featureTableMetadata []*spec.FeatureTableMetadata
	feastOpts            feast.Options
	logger               *zap.Logger
	modelPredictor       ModelPredictor
	protocol             prt.Protocol
}

// NewStandardTransformerWithConfig initialize standard transformer executor object
func NewStandardTransformerWithConfig(ctx context.Context, transformerConfig *spec.StandardTransformerConfig, opts ...TransformerOptions) (Transformer, error) {
	defaultModelPredictor := newEchoMockPredictor()
	executorConfig := &transformerExecutorConfig{
		transformerConfig: transformerConfig,
		modelPredictor:    defaultModelPredictor,
		feastOpts: feast.Options{
			FeastGRPCConnCount: 1,
		},
	}
	for _, opt := range opts {
		opt(executorConfig)
	}

	feastServingClients, err := feast.InitFeastServingClients(executorConfig.feastOpts, executorConfig.featureTableMetadata, transformerConfig)
	if err != nil {
		return nil, err
	}

	compiler := pipeline.NewCompiler(
		symbol.NewRegistry(),
		feastServingClients,
		&executorConfig.feastOpts,
		pipeline.WithLogger(executorConfig.logger),
		pipeline.WithOperationTracingEnabled(executorConfig.traceEnabled),
		pipeline.WithProtocol(executorConfig.protocol),
	)
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
func (st *standardTransformer) Execute(ctx context.Context, requestBody types.JSONObject, requestHeaders map[string]string) *types.PredictResponse {
	span, ctx := opentracing.StartSpanFromContext(ctx, "st.Execute")
	defer span.Finish()

	requestPayload, err := createRequestPayload(st.executorConfig.protocol, requestBody)
	if err != nil {
		st.logger.Warn("request payload conversion", zap.Any("err", err.Error()))
		return generateErrorResponse(fmt.Errorf("request is not valid, user should specifies request with UPI PredictValuesRequest type"))
	}

	preprocessOut := requestPayload
	st.logger.Debug("raw request_body", zap.Any("request_body", requestBody))

	env := pipeline.NewEnvironment(st.compiledPipeline, st.executorConfig.logger)

	if env.IsPreprocessOpExist() {
		preprocessOut, err = env.Preprocess(ctx, requestPayload, requestHeaders)
		if err != nil {
			return generateErrorResponse(err)
		}
		st.logger.Debug("preprocess response", zap.Any("preprocess_response", preprocessOut))
	}

	reqBody, err := preprocessOut.AsOutput()
	if err != nil {
		return generateErrorResponse(err)
	}

	predictorRespBody, predictorRespHeaders, err := st.modelPredictor.ModelPrediction(ctx, reqBody, requestHeaders)
	if err != nil {
		return generateErrorResponse(err)
	}

	st.logger.Debug("predictor response", zap.Any("predict_response", predictorRespBody))

	predictionOut := predictorRespBody
	if env.IsPostProcessOpExist() {
		predictionOut, err = env.Postprocess(ctx, predictionOut, predictorRespHeaders)
		if err != nil {
			return generateErrorResponse(err)
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

	return resp
}

func createRequestPayload(protocol prt.Protocol, jsonRequestPayload types.JSONObject) (types.Payload, error) {
	if protocol == prt.HttpJson || protocol == "" {
		return jsonRequestPayload, nil
	}

	payloadData, err := json.Marshal(jsonRequestPayload)
	if err != nil {
		return nil, err
	}
	var requestPayload upiv1.PredictValuesRequest
	if err := protojson.Unmarshal(payloadData, &requestPayload); err != nil {
		return nil, err
	}

	return (*types.UPIPredictionRequest)(&requestPayload), nil
}

func generateErrorResponse(err error) *types.PredictResponse {
	return &types.PredictResponse{
		Response: types.JSONObject{
			"error": err.Error(),
		},
	}
}
