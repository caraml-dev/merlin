package feast

import (
	"fmt"

	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
)

type defaultValues map[string]*feastTypes.Value

func (d defaultValues) GetDefaultValue(project string, feature string) (*feastTypes.Value, bool) {
	val, ok := d[fmt.Sprintf("%s-%s", project, feature)]
	return val, ok
}

func (d defaultValues) SetDefaultValue(project string, feature string, value *feastTypes.Value) {
	d[fmt.Sprintf("%s-%s", project, feature)] = value
}

func compileDefaultValues(featureTableSpecs []*spec.FeatureTable) defaultValues {
	defaultValues := defaultValues{}
	// populate default values
	for _, ft := range featureTableSpecs {
		for _, f := range ft.Features {
			if len(f.DefaultValue) != 0 {
				feastValType := feastTypes.ValueType_Enum(feastTypes.ValueType_Enum_value[f.ValueType])
				defVal, err := converter.ToFeastValue(f.DefaultValue, feastValType)
				if err != nil {
					continue
				}

				defaultValues.SetDefaultValue(ft.Project, f.Name, defVal)
			}
		}
	}

	return defaultValues
}
