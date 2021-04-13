package pipeline

import (
	"context"
	"encoding/json"

	"github.com/gojek/merlin/pkg/transformer/types"
)

type Handler struct {
	compiledPipeline *CompiledPipeline
}

const PipelineEnvironmentContext = "merlin-transfomer-environment"

func NewHandler(compiledPipeline *CompiledPipeline) *Handler {
	return &Handler{compiledPipeline: compiledPipeline}
}

func (h *Handler) Preprocess(ctx context.Context, rawRequest []byte, rawRequestHeaders map[string]string) ([]byte, error) {
	env := getEnvironment(ctx)
	var rawRequestObj types.JSONObject
	err := json.Unmarshal(rawRequest, &rawRequestObj)
	if err != nil {
		return nil, err
	}

	transformedRequest, err := env.Preprocess(ctx, rawRequestObj, rawRequestHeaders)
	return json.Marshal(transformedRequest)
}

func (h *Handler) Postprocess(ctx context.Context, modelResponse []byte, modelResponseHeaders map[string]string) ([]byte, error) {
	env := getEnvironment(ctx)

	var modelResponseObj types.JSONObject
	err := json.Unmarshal(modelResponse, &modelResponseObj)
	if err != nil {
		return nil, err
	}

	transformedResponse, err := env.Preprocess(ctx, modelResponseObj, modelResponseHeaders)
	return json.Marshal(transformedResponse)
}

func (h *Handler) EmbedEnvironment(ctx context.Context) context.Context {
	env := NewEnvironment(h.compiledPipeline)
	return context.WithValue(ctx, PipelineEnvironmentContext, env)
}

func getEnvironment(ctx context.Context) *Environment {
	return ctx.Value(PipelineEnvironmentContext).(*Environment)
}
