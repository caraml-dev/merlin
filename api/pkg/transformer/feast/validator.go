package feast

import (
	"context"
	"fmt"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer"
)

// ValidateTransformerConfig validate transformer config by checking the presence of entity and features in feast core
func ValidateTransformerConfig(ctx context.Context, coreClient core.CoreServiceClient, trfCfg *transformer.StandardTransformerConfig) error {
	if trfCfg.TransformerConfig == nil {
		return NewValidationError("transformerConfig is empty")
	}

	if len(trfCfg.TransformerConfig.Feast) == 0 {
		return NewValidationError("feature retrieval config is empty")
	}

	req := &core.ListEntitiesRequest{}
	res, err := coreClient.ListEntities(ctx, req)
	if err != nil {
		return errors.Wrap(err, "error retrieving list of entity")
	}

	// allEntities contains all entities registerd in feast
	allEntities := make(map[string]*core.EntitySpecV2)
	for _, entity := range res.GetEntities() {
		allEntities[entity.GetSpec().GetName()] = entity.GetSpec()
	}

	// for each feature retrieval table
	for _, config := range trfCfg.TransformerConfig.Feast {
		if len(config.Entities) == 0 {
			return NewValidationError("no entity")
		}

		if len(config.Features) == 0 {
			return NewValidationError("no feature")
		}
		// check that entities is non empty
		// check that all entity has json path
		// check that all entity has type
		// check all entity given in config are all registered ones
		entities := make([]string, 0)
		for _, entity := range config.Entities {
			spec, found := allEntities[entity.Name]
			if !found {
				return NewValidationError("entity not found: " + entity.Name)
			}

			if len(entity.JsonPath) == 0 {
				return NewValidationError(fmt.Sprintf("json path for %s is not specified", entity.Name))
			}

			if spec.ValueType.String() != entity.ValueType {
				return NewValidationError(fmt.Sprintf("mismatched value type for %s, expect: %s, got: %s", entity.Name, spec.ValueType.String(), entity.ValueType))
			}

			entities = append(entities, entity.Name)
		}

		// get all features that are referenced by all entities defined in config
		req := &core.ListFeaturesRequest{Filter: &core.ListFeaturesRequest_Filter{
			Entities: entities,
			Project:  config.Project,
		}}
		res, err := coreClient.ListFeatures(ctx, req)
		if err != nil {
			return errors.Wrap(err, "error retrieving list of features")
		}

		featureShortNames := make(map[string]*core.FeatureSpecV2)
		for _, feature := range res.Features {
			featureShortNames[feature.Name] = feature
		}

		for _, feature := range config.Features {
			// check against feature short name or fully qualified name
			fs, fqNameFound := res.Features[feature.Name]
			fs2, shortNameFound := featureShortNames[feature.Name]
			if !fqNameFound && !shortNameFound {
				return NewValidationError(fmt.Sprintf("feature not found for entities %s in project %s: %s", entities, config.Project, feature.Name))
			}

			featureSpec := fs
			if fs2 != nil {
				featureSpec = fs2
			}

			if featureSpec.ValueType.String() != feature.ValueType {
				return NewValidationError(fmt.Sprintf("mismatched value type for %s, expect: %s, got: %s", feature.Name, featureSpec.ValueType.String(), feature.ValueType))
			}
		}
	}

	return nil
}
