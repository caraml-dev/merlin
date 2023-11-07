package feast

import (
	"errors"

	transTypes "github.com/caraml-dev/merlin/pkg/transformer/types"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
)

// internalFeatureTable helper type for internal table processing
type internalFeatureTable struct {
	entities    []feast.Row
	columnNames []string
	columnTypes []types.ValueType_Enum
	indexRows   []int
	valueRows   transTypes.ValueRows
}

type orderedFeastRow struct {
	Index int
	Row   feast.Row
}

func (it *internalFeatureTable) toFeatureTable(tableName string) *transTypes.FeatureTable {
	sortedValueRows := make(transTypes.ValueRows, len(it.valueRows))
	for i, index := range it.indexRows {
		sortedValueRows[index] = it.valueRows[i]
	}
	return &transTypes.FeatureTable{
		Name:        tableName,
		Columns:     it.columnNames,
		ColumnTypes: it.columnTypes,
		Data:        sortedValueRows,
	}
}

func (it *internalFeatureTable) mergeFeatureTable(right *internalFeatureTable) error {
	if right == nil {
		return nil
	}

	// It is assumed that
	// both table have same number of columns
	if len(it.columnTypes) != len(right.columnTypes) ||
		len(it.columnNames) != len(right.columnNames) {
		return errors.New("unable to merge tables: different number of columns")
	}

	// merge column types to ensure that if in previous batch we don't know a column type (i.e. ValueType_INVALID)
	// it will be overwritten by the new batch which knows the column type
	it.columnTypes = mergeColumnTypes(it.columnTypes, right.columnTypes)
	it.entities = append(it.entities, right.entities...)
	it.valueRows = append(it.valueRows, right.valueRows...)
	it.indexRows = append(it.indexRows, right.indexRows...)

	return nil
}

// callResult result returned from one batch call to Feast
type callResult struct {
	tableName    string
	featureTable *internalFeatureTable
	err          error
}

func mergeColumnTypes(dst []types.ValueType_Enum, src []types.ValueType_Enum) []types.ValueType_Enum {
	if len(dst) == 0 {
		dst = src
		return dst
	}

	for i, t := range dst {
		if t == types.ValueType_INVALID {
			dst[i] = src[i]
		}
	}
	return dst
}
