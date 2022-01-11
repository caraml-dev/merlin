package bigtablestore

import (
	"fmt"
	"strings"
)

type FeatureRef struct {
	FeatureTable string
	Feature      string
}

// ParseFeatureRef parse feature reference string in the format of <feature table>:<feature name>
func ParseFeatureRef(ref string) (FeatureRef, error) {
	split := strings.Split(ref, ":")
	if len(split) != 2 {
		return FeatureRef{}, fmt.Errorf("malformed feature ref: %s", ref)
	}
	return FeatureRef{
		FeatureTable: split[0],
		Feature:      split[1],
	}, nil
}

// UniqueFeatureTablesFromFeatureRef returns a list of unique feature tables associated with
// the given feature references, sorted by the order of appearance in the feature references slice
func UniqueFeatureTablesFromFeatureRef(refs []string) ([]string, error) {
	featureTables := make([]string, 0)
	existingFeatureTables := make(map[string]bool)
	for _, r := range refs {
		featureRef, err := ParseFeatureRef(r)
		if err != nil {
			return nil, err
		}
		if _, exist := existingFeatureTables[featureRef.FeatureTable]; !exist {
			existingFeatureTables[featureRef.FeatureTable] = true
			featureTables = append(featureTables, featureRef.FeatureTable)
		}
	}
	return featureTables, nil
}
