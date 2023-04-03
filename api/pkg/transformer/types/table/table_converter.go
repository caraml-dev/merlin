package table

import (
	"errors"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
)

const (
	columnsJsonKey = "columns"
	dataJsonKey    = "data"
)

func TableToJson(tbl *Table, format spec.FromTable_JsonFormat) (interface{}, error) {
	switch format {
	case spec.FromTable_RECORD:
		return tableToJsonRecordFormat(tbl)
	case spec.FromTable_VALUES:
		return tableToJsonValuesFormat(tbl)
	case spec.FromTable_SPLIT:
		return tableToJsonSplitFormat(tbl)
	default:
		return nil, errors.New("unsupported format")
	}
}

func tableToJsonRecordFormat(tbl *Table) (interface{}, error) {
	columnsValues, err := getTableRecordsPerColumns(tbl)
	if err != nil {
		return nil, err
	}

	dataFrame := tbl.DataFrame()
	columns := dataFrame.Names()

	records := make([]interface{}, 0, dataFrame.Nrow())
	for i := 0; i < dataFrame.Nrow(); i++ {
		rowJson := make(map[string]interface{}, len(columns))
		for _, col := range columns {
			values := columnsValues[col]
			rowJson[col] = values[i]
		}
		records = append(records, rowJson)
	}
	return records, nil
}

func tableToJsonValuesFormat(tbl *Table) (interface{}, error) {
	columnsValues, err := getTableRecordsPerColumns(tbl)
	if err != nil {
		return nil, err
	}

	dataFrame := tbl.DataFrame()
	columns := dataFrame.Names()

	var records []interface{}
	for i := 0; i < dataFrame.Nrow(); i++ {
		var rowRecord []interface{}
		for _, col := range columns {
			values := columnsValues[col]
			rowRecord = append(rowRecord, values[i])
		}
		records = append(records, rowRecord)
	}
	return records, nil
}

func tableToJsonSplitFormat(tbl *Table) (interface{}, error) {
	dataFrame := tbl.DataFrame()
	columns := dataFrame.Names()
	records := make(map[string]interface{})
	records[columnsJsonKey] = columns

	values, err := tableToJsonValuesFormat(tbl)
	if err != nil {
		return nil, err
	}
	records[dataJsonKey] = values

	return records, nil
}

func getTableRecordsPerColumns(tbl *Table) (map[string][]interface{}, error) {
	columns := tbl.DataFrame().Names()

	columnsValues := make(map[string][]interface{})
	for _, col := range columns {
		valueSeries := tbl.Col(col)
		columnsValues[col] = valueSeries.GetRecords()
	}
	return columnsValues, nil
}
