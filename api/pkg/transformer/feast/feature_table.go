package feast

import (
	"context"
	"fmt"
	"strings"

	"github.com/feast-dev/feast/sdk/go/protos/feast/core"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

// GetAllFeatureTableMetadata retrieves all metadata for feature tables
// that specified in feast transformer or pipeline (preprocess or postprocess)
func GetAllFeatureTableMetadata(ctx context.Context, coreClient core.CoreServiceClient, standardTransformerConfig *spec.StandardTransformerConfig) ([]*spec.FeatureTableMetadata, error) {
	featureTableSpecs := getFeatureTableSpecs(standardTransformerConfig)
	if len(featureTableSpecs) == 0 {
		return nil, nil
	}
	return getFeatureTablesMetadata(ctx, coreClient, featureTableSpecs)
}

// UpdateFeatureTableSource will update feature table spec source in standard transformer
func UpdateFeatureTableSource(standardTransformerConfig *spec.StandardTransformerConfig, sourceByURLMap map[string]spec.ServingSource, defaultSource spec.ServingSource) {
	if featureTableCfgs := standardTransformerConfig.TransformerConfig.Feast; featureTableCfgs != nil {
		updateFeatureTableSource(featureTableCfgs, sourceByURLMap, defaultSource)
	} else {
		pipelines := []*spec.Pipeline{standardTransformerConfig.TransformerConfig.Preprocess, standardTransformerConfig.TransformerConfig.Postprocess}
		for _, pipeline := range pipelines {
			if pipeline == nil {
				continue
			}
			inputs := pipeline.Inputs
			for _, input := range inputs {
				updateFeatureTableSource(input.Feast, sourceByURLMap, defaultSource)
			}
		}
	}
}

func updateFeatureTableSource(featureTableSpecs []*spec.FeatureTable, sourceByURLMap map[string]spec.ServingSource, defaultSource spec.ServingSource) {
	for idx, featureTableCfg := range featureTableSpecs {
		if _, found := spec.ServingSource_name[int32(featureTableCfg.Source)]; !found {
			featureTableCfg.Source = defaultSource
			featureTableSpecs[idx] = featureTableCfg
			continue
		}

		if featureTableCfg.Source == spec.ServingSource_UNKNOWN {
			feastSource := defaultSource
			if featureTableCfg.ServingUrl != "" {
				source := sourceByURLMap[featureTableCfg.ServingUrl]
				if source != spec.ServingSource_UNKNOWN {
					feastSource = source
				}
			}
			featureTableCfg.Source = feastSource
			featureTableSpecs[idx] = featureTableCfg
		}
	}
}

func getFeatureTablesMetadata(ctx context.Context, coreClient core.CoreServiceClient, featureTableSpecs []*spec.FeatureTable) ([]*spec.FeatureTableMetadata, error) {
	featureTableMetadataMap := make(map[string]*spec.FeatureTableMetadata)
	for _, featureTable := range featureTableSpecs {
		project := featureTable.Project
		for _, featureRef := range featureTable.Features {
			// check whether feature table spec already fetched
			// skip if already there
			featureTableName := getFeatureTableFromFeatureRef(featureRef.Name)
			featureTableKeyName := fmt.Sprintf("%s-%s", project, featureTableName)
			if _, featureTableExist := featureTableMetadataMap[featureTableKeyName]; featureTableExist {
				continue
			}
			featureTableResp, err := coreClient.GetFeatureTable(ctx, &core.GetFeatureTableRequest{
				Project: featureTable.Project,
				Name:    featureTableName,
			})
			if err != nil {
				return nil, err
			}
			if featureTableResp.Table != nil {
				featureTableSpec := featureTableResp.Table.Spec
				featureTableMetadata := &spec.FeatureTableMetadata{
					Name:    featureTableSpec.Name,
					Project: project,
					MaxAge:  featureTableSpec.MaxAge,
				}
				featureTableMetadataMap[featureTableKeyName] = featureTableMetadata
			}
		}
	}
	featureTableMetadata := make([]*spec.FeatureTableMetadata, 0, len(featureTableMetadataMap))
	for _, spec := range featureTableMetadataMap {
		featureTableMetadata = append(featureTableMetadata, spec)
	}
	return featureTableMetadata, nil
}

func getFeatureTableFromFeatureRef(ref string) string {
	return strings.Split(ref, ":")[0]
}
