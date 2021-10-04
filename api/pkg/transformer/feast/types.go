package feast

import (
	"errors"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"

	transTypes "github.com/gojek/merlin/pkg/transformer/types"
)

// internalFeatureTable helper type for internal table processing
type internalFeatureTable struct {
	entities    []feast.Row
	columnNames []string
	columnTypes []types.ValueType_Enum
	valueRows   transTypes.ValueRows
}

func (it *internalFeatureTable) toFeatureTable(tableName string) *transTypes.FeatureTable {
	return &transTypes.FeatureTable{
		Name:        tableName,
		Columns:     it.columnNames,
		ColumnTypes: it.columnTypes,
		Data:        it.valueRows,
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

	return nil
}

// callResult result returned from one batch call to Feast
type callResult struct {
	tableName    string
	featureTable *internalFeatureTable
	err          error
}
