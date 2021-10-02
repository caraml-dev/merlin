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
