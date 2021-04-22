package feast

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/oliveagle/jsonpath"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/pkg/errors"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/symbol"
)

// ValidateTransformerConfig validate transformer config by checking the presence of entity and features in feast core
func ValidateTransformerConfig(ctx context.Context, coreClient core.CoreServiceClient, featureTableConfigs []*spec.FeatureTable) error {
	// for each feature retrieval table
	for _, config := range featureTableConfigs {
		if len(config.Entities) == 0 {
			return NewValidationError("no entity")
		}

		if len(config.Features) == 0 {
			return NewValidationError("no feature")
		}

		entitiesReq := &core.ListEntitiesRequest{
			Filter: &core.ListEntitiesRequest_Filter{
				Project: config.Project,
			},
		}
		entitiesRes, err := coreClient.ListEntities(ctx, entitiesReq)
		if err != nil {
			return errors.Wrapf(err, "error retrieving list of entity for project %s", config.Project)
		}

		// allEntities contains all entities registerd in feast
		allEntities := make(map[string]*core.EntitySpecV2)
		for _, entity := range entitiesRes.GetEntities() {
			allEntities[entity.GetSpec().GetName()] = entity.GetSpec()
		}

		// check that entities is non empty
		// check that all entity has json path
		// check that all entity has type
		// check all entity given in config are all registered ones
		entities := make([]string, 0)
		for _, entity := range config.Entities {
			entitySpec, found := allEntities[entity.Name]
			if !found {
				return NewValidationError("entity not found: " + entity.Name)
			}

			switch entity.Extractor.(type) {
			case *spec.Entity_JsonPath:
				if len(entity.GetJsonPath()) == 0 {
					return NewValidationError(fmt.Sprintf("json path for %s is not specified", entity.Name))
				}
				_, err = jsonpath.Compile(entity.GetJsonPath())
				if err != nil {
					return NewValidationError(fmt.Sprintf("jsonpath compilation failed: %v", err))
				}
			case *spec.Entity_Udf, *spec.Entity_Expression:
				expressionExtractor := getExpressionExtractor(entity)
				_, err = expr.Compile(expressionExtractor, expr.Env(symbol.NewRegistryWithCompiledJSONPath(nil)))
				if err != nil {
					return NewValidationError(fmt.Sprintf("udf compilation failed: %v", err))
				}
			default:
				return NewValidationError(fmt.Sprintf("one of json_path, udf must be specified"))
			}

			if entitySpec.ValueType.String() != entity.ValueType {
				return NewValidationError(fmt.Sprintf("mismatched value type for %s, expect: %s, got: %s", entity.Name, entitySpec.ValueType.String(), entity.ValueType))
			}

			entities = append(entities, entity.Name)
		}

		// get all features that are referenced by all entities defined in config
		featuresReq := &core.ListFeaturesRequest{
			Filter: &core.ListFeaturesRequest_Filter{
				Entities: entities,
				Project:  config.Project,
			},
		}
		featuresRes, err := coreClient.ListFeatures(ctx, featuresReq)
		if err != nil {
			return errors.Wrap(err, "error retrieving list of features")
		}

		featureShortNames := make(map[string]*core.FeatureSpecV2)
		for _, feature := range featuresRes.Features {
			featureShortNames[feature.Name] = feature
		}

		for _, feature := range config.Features {
			// check against feature short name or fully qualified name
			fs, fqNameFound := featuresRes.Features[feature.Name]
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
