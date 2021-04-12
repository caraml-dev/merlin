package types

import (
	"fmt"

	"github.com/feast-dev/feast/sdk/go/protos/feast/types"

	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type ValueRow []interface{}

type ValueRows []ValueRow

type FeatureTable struct {
	Name        string                 `json:"-"`
	Columns     []string               `json:"columns"`
	ColumnTypes []types.ValueType_Enum `json:"-"`
	Data        ValueRows              `json:"data"`
}
