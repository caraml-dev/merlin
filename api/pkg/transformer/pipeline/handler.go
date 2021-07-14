package pipeline

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer/types"
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

func (h *Handler) Preprocess(ctx context.Context, rawRequest []byte, rawRequestHeaders map[string]string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pipeline.Preprocess")
	defer span.Finish()

	env := getEnvironment(ctx)
	if !env.IsPreprocessOpExist() {
		return rawRequest, nil
	}

	var rawRequestObj types.JSONObject
	err := json.Unmarshal(rawRequest, &rawRequestObj)
	if err != nil {
		return nil, err
	}

	transformedRequest, err := env.Preprocess(ctx, rawRequestObj, rawRequestHeaders)
	if err != nil {
		return nil, err
	}

	jsonSpan, _ := opentracing.StartSpanFromContext(ctx, "pipeline.Preprocess.JsonMarshall")
	preprocessJson, err := json.Marshal(transformedRequest)
	jsonSpan.Finish()

	return preprocessJson, err
}

func (h *Handler) Postprocess(ctx context.Context, modelResponse []byte, modelResponseHeaders map[string]string) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pipeline.Postprocess")
	defer span.Finish()

	env := getEnvironment(ctx)
	if !env.IsPostProcessOpExist() {
		return modelResponse, nil
	}

	var modelResponseObj types.JSONObject
	err := json.Unmarshal(modelResponse, &modelResponseObj)
	if err != nil {
		return nil, err
	}

	transformedResponse, err := env.Postprocess(ctx, modelResponseObj, modelResponseHeaders)
	if err != nil {
		return nil, err
	}

	return json.Marshal(transformedResponse)
}

func (h *Handler) EmbedEnvironment(ctx context.Context) context.Context {
	env := NewEnvironment(h.compiledPipeline, h.logger)
	return context.WithValue(ctx, PipelineEnvironmentContext, env)
}

func getEnvironment(ctx context.Context) *Environment {
	return ctx.Value(PipelineEnvironmentContext).(*Environment)
}
