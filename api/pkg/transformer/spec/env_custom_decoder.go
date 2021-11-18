package spec

import "fmt"

func (source *ServingSource) Decode(value string) error {
	val, ok := ServingSource_value[value]
	if !ok {
		return fmt.Errorf("invalid serving source value %s", value)
	}
	*source = ServingSource(val)
	return nil
}
