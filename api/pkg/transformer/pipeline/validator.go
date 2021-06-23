package pipeline

import (
	"context"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"

	"github.com/gojek/merlin/pkg/transformer/cache"
	"github.com/gojek/merlin/pkg/transformer/feast"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

func ValidateTransformerConfig(ctx context.Context, coreClient core.CoreServiceClient, transformerConfig *spec.StandardTransformerConfig) error {
	if transformerConfig.TransformerConfig.Feast != nil {
		return feast.ValidateTransformerConfig(ctx, coreClient, transformerConfig.TransformerConfig.Feast, symbol.NewRegistryWithCompiledJSONPath(nil))
	}

	// compile pipeline
	compiler := NewCompiler(symbol.NewRegistry(), nil, &feast.Options{}, &cache.Options{}, nil)
	_, err := compiler.Compile(transformerConfig)
	if err != nil {
		return err
	}

	// validate all feast features in preprocess input
	err = validateFeastFeaturesInPipeline(ctx, coreClient, transformerConfig.TransformerConfig.Preprocess, compiler.sr)
	if err != nil {
		return err
	}

	// validate all feast features in post process input
	return validateFeastFeaturesInPipeline(ctx, coreClient, transformerConfig.TransformerConfig.Postprocess, compiler.sr)
}

func validateFeastFeaturesInPipeline(ctx context.Context, coreClient core.CoreServiceClient, pipeline *spec.Pipeline, symbolRegistry symbol.Registry) error {
	if pipeline == nil {
		return nil
	}

	if pipeline.Inputs == nil {
		return nil
	}

	for _, input := range pipeline.Inputs {
		if input.Feast != nil {
			err := feast.ValidateTransformerConfig(ctx, coreClient, input.Feast, symbolRegistry)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
