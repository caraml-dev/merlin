package pipeline

import (
	"context"
	"fmt"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"

	prt "github.com/caraml-dev/merlin/pkg/protocol"
	"github.com/caraml-dev/merlin/pkg/transformer/feast"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/symbol"
)

func ValidateTransformerConfig(ctx context.Context, coreClient core.CoreServiceClient, transformerConfig *spec.StandardTransformerConfig, feastOptions *feast.Options, protocol prt.Protocol) error {
	if transformerConfig.TransformerConfig.Feast != nil {
		return feast.ValidateTransformerConfig(ctx, coreClient, transformerConfig.TransformerConfig.Feast, symbol.NewRegistryWithCompiledJSONPath(nil), feastOptions)
	}

	// compile pipeline
	compiler := NewCompiler(
		symbol.NewRegistry(),
		nil,
		feastOptions,
		WithProtocol(protocol),
	)
	_, err := compiler.Compile(transformerConfig)
	if err != nil {
		return err
	}

	// validate all feast features in preprocess input
	err = validateFeastFeaturesInPipeline(ctx, coreClient, transformerConfig.TransformerConfig.Preprocess, compiler.sr, feastOptions)
	if err != nil {
		return err
	}

	// validate all feast features in post process input
	return validateFeastFeaturesInPipeline(ctx, coreClient, transformerConfig.TransformerConfig.Postprocess, compiler.sr, feastOptions)
}

func validateFeastFeaturesInPipeline(ctx context.Context, coreClient core.CoreServiceClient, pipeline *spec.Pipeline, symbolRegistry symbol.Registry, feastOptions *feast.Options) error {
	if pipeline == nil {
		return nil
	}

	if pipeline.Inputs == nil {
		return nil
	}

	for _, input := range pipeline.Inputs {
		if input.Feast != nil {
			err := feast.ValidateTransformerConfig(ctx, coreClient, input.Feast, symbolRegistry, feastOptions)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func httpTransformerValidation(config *spec.StandardTransformerConfig) error {
	if config.TransformerConfig == nil {
		return nil
	}

	if config.PredictionLogConfig != nil && config.PredictionLogConfig.Enable {
		return fmt.Errorf("prediction log config only available for UPI_V1 protocol")
	}

	validationFn := func(step *spec.Pipeline) error {
		for _, input := range step.Inputs {
			if input.Autoload != nil {
				return fmt.Errorf("autoload is only supported for upi_v1 protocol")
			}
		}
		for _, output := range step.Outputs {
			if output.UpiPreprocessOutput != nil || output.UpiPostprocessOutput != nil {
				return fmt.Errorf("jsonOutput is only supported for http protocol")
			}
		}
		return nil
	}

	if preprocess := config.TransformerConfig.Preprocess; preprocess != nil {
		if err := validationFn(preprocess); err != nil {
			return err
		}
	}
	if postprocess := config.TransformerConfig.Postprocess; postprocess != nil {
		if err := validationFn(postprocess); err != nil {
			return err
		}
	}
	return nil
}

func upiTransformerValidation(spec *spec.StandardTransformerConfig) error {
	if spec.TransformerConfig == nil {
		return nil
	}

	if preprocess := spec.TransformerConfig.Preprocess; preprocess != nil {
		for _, output := range preprocess.Outputs {
			if output.JsonOutput != nil {
				return fmt.Errorf("json output is not supported")
			}
			if output.UpiPostprocessOutput != nil {
				return fmt.Errorf("UPIPostprocessOutput is not supported in preprocess step")
			}
		}
	}
	if postprocess := spec.TransformerConfig.Postprocess; postprocess != nil {
		for _, output := range postprocess.Outputs {
			if output.JsonOutput != nil {
				return fmt.Errorf("json output is not supported")
			}
			if output.UpiPreprocessOutput != nil {
				return fmt.Errorf("UPIPreprocessOutput is not supported in postprocess step")
			}
		}
	}
	return nil
}
