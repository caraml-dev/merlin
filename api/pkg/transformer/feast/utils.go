package feast

import (
	"strings"

	"github.com/gojek/merlin/pkg/transformer/spec"
)

const defaultProjectName = "default"

// GetTableName get the table name of a given featureTableSpec
// If user doesn't specify featureTableSpec.TableName, then the table name will be constructed by combining all entities name within the spec.
func GetTableName(featureTableSpec *spec.FeatureTable) string {
	if featureTableSpec.TableName != "" {
		return featureTableSpec.TableName
	}

	entityNames := make([]string, 0)
	for _, n := range featureTableSpec.Entities {
		entityNames = append(entityNames, n.Name)
	}

	tableName := strings.Join(entityNames, "_")
	if featureTableSpec.Project != defaultProjectName {
		tableName = featureTableSpec.Project + "_" + tableName
	}

	return tableName
}

func getFeastServingSources(stdTransformerConfig *spec.StandardTransformerConfig) []spec.ServingSource {
	feastSources := make(map[spec.ServingSource]spec.ServingSource)
	if stdTransformerConfig.TransformerConfig.Feast != nil {
		for _, featureTableSpec := range stdTransformerConfig.TransformerConfig.Feast {
			feastSources[featureTableSpec.Source] = featureTableSpec.Source
		}
	} else {
		pipelines := []*spec.Pipeline{stdTransformerConfig.TransformerConfig.Preprocess, stdTransformerConfig.TransformerConfig.Postprocess}
		for _, pipeline := range pipelines {
			if pipeline == nil {
				continue
			}
			for _, input := range pipeline.Inputs {
				for _, featureTableSpec := range input.Feast {
					feastSources[featureTableSpec.Source] = featureTableSpec.Source
				}
			}
		}
	}
	sources := make([]spec.ServingSource, 0, len(feastSources))
	for source := range feastSources {
		sources = append(sources, source)
	}
	return sources
}

func getFeatureTableSpecs(stdTransformerConfig *spec.StandardTransformerConfig) []*spec.FeatureTable {
	if stdTransformerConfig.TransformerConfig.Feast != nil {
		return stdTransformerConfig.TransformerConfig.Feast
	}
	preprocessFeatureTableSpecs := getFeatureTableConfigsFromPipeline(stdTransformerConfig.TransformerConfig.Preprocess)
	postprocessFeatureTableSpecs := getFeatureTableConfigsFromPipeline(stdTransformerConfig.TransformerConfig.Postprocess)
	allFeatureTableSpecs := append(preprocessFeatureTableSpecs, postprocessFeatureTableSpecs...)
	return allFeatureTableSpecs
}

func getFeatureTableConfigsFromPipeline(pipeline *spec.Pipeline) []*spec.FeatureTable {
	if pipeline == nil {
		return []*spec.FeatureTable{}
	}

	if pipeline.Inputs == nil {
		return []*spec.FeatureTable{}
	}

	featureTableCfgs := make([]*spec.FeatureTable, 0)
	for _, input := range pipeline.Inputs {
		featureTableCfgs = append(featureTableCfgs, input.Feast...)
	}
	return featureTableCfgs
}
