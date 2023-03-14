package pipeline

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/types"
)

type Handler struct {
	compiledPipeline *CompiledPipeline
	logger           *zap.Logger
}

const PipelineEnvironmentContext = "merlin-transfomer-environment"

func NewHandler(compiledPipeline *CompiledPipeline, logger *zap.Logger) *Handler {
	return &Handler{
		compiledPipeline: compiledPipeline,
		logger:           logger,
	}
}

func (h *Handler) Preprocess(ctx context.Context, rawRequest types.Payload, rawRequestHeaders map[string]string) (types.Payload, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pipeline.Preprocess")
	defer span.Finish()

	env := getEnvironment(ctx)
	if !env.IsPreprocessOpExist() {
		return rawRequest, nil
	}

	rawRequestObj, err := rawRequest.AsInput()
	if err != nil {
		return nil, err
	}

	result, err := env.Preprocess(ctx, rawRequestObj, rawRequestHeaders)
	if err != nil {
		return nil, err
	}

	jsonSpan, _ := opentracing.StartSpanFromContext(ctx, "pipeline.Preprocess.JsonMarshall")
	jsonSpan.Finish()

	return result.AsOutput()
}

func (h *Handler) Postprocess(ctx context.Context, modelResponse types.Payload, modelResponseHeaders map[string]string) (types.Payload, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pipeline.Postprocess")
	defer span.Finish()

	env := getEnvironment(ctx)
	if !env.IsPostProcessOpExist() {
		return modelResponse, nil
	}

	modelResponseObj, err := modelResponse.AsInput()
	if err != nil {
		return nil, err
	}

	transformedResponse, err := env.Postprocess(ctx, modelResponseObj, modelResponseHeaders)
	if err != nil {
		return nil, err
	}

	return transformedResponse.AsOutput()
}

func (h *Handler) PredictionLogHandler(ctx context.Context, result *types.PredictionResult) {
	env := getEnvironment(ctx)
	env.PublishPredictionLog(context.Background(), result)
}

func (h *Handler) EmbedEnvironment(ctx context.Context) context.Context {
	env := NewEnvironment(h.compiledPipeline, h.logger)
	return context.WithValue(ctx, PipelineEnvironmentContext, env) //nolint: staticcheck
}

func getEnvironment(ctx context.Context) *Environment {
	return ctx.Value(PipelineEnvironmentContext).(*Environment)
}
